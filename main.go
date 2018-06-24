package main

import (
	"encoding/json"
	"fmt"
	"github.com/0990/simpleGameServer/game"
	"github.com/0990/simpleGameServer/net"
	"github.com/0990/simpleGameServer/user"
)

func main() {
	workerChan := make(chan func(), 100)
	go func() {
		for workerFunc := range workerChan {
			workerFunc()
		}
	}()
	net.RegisterConnectFun(func(client *net.Client) {
		user.GetManager().CreateUser(client)
	})
	net.RegisterDisconnectFun(func(client *net.Client) {
		user.GetManager().DestroyUser(client)
	})
	net.RegisterMessageFun(func(client *net.Client, message []byte) {
		sendUser, ok := user.GetManager().GetUserByClient(client)
		if ok {
			var data map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				fmt.Println(err)
			}
			sendUser.SendMsg(message)
			mainID := int32(data["mainID"].(float64))
			switch mainID {
			case 1:
				//create room
				roomid := game.GetManager().CreateRoom(sendUser.ID())
				sendMap := make(map[string]interface{})
				sendMap["roomID"] = roomid
				sendMap["createRoomsuccess"] = true
				sendBytes, _ := json.Marshal(sendMap)
				sendUser.SendMsg(sendBytes)
			case 2:
				//enter room
				roomID := int32(data["roomID"].(float64))
				game.GetManager().EnterRoom(roomID, sendUser)
				//if success {
				//	sendMap := make(map[string]interface{})
				//	sendMap["roomID"] = roomID
				//	sendMap["enterRoomsuccess"] = true
				//	sendBytes, _ := json.Marshal(sendMap)
				//	sendUser.SendMsg(sendBytes)
				//}
			case 3:
				//game message
				subID := int32(data["subID"].(float64))
				if room, ok := game.GetManager().GetRoomByUserid(sendUser.ID()); ok {
					room.Post(func() {
						room.OnGameMessage(sendUser.ID(), subID)
					})
				}
			default:

			}
		}
		users := user.GetManager().GetUsers()
		for _, user := range users {
			user.SendMsg(message)
		}
	})
	net.Run(workerChan)
}
