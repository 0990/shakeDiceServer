package main

import (
	"github.com/0990/shakeDiceServer/net"
	"github.com/0990/shakeDiceServer/user"
	"encoding/json"
	"fmt"
	"github.com/0990/shakeDiceServer/msg"
	"github.com/0990/shakeDiceServer/game"
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
	net.RegisterMessageFun(onMessage)
	net.Run(workerChan)
}

func onMessage(client *net.Client, message []byte){
	sendUser, ok := user.GetManager().GetUserByClient(client)
	if ok {
		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			fmt.Println(err)
		}
		sendUser.SendMsg(message)
		mainID := int32(data["mainID"].(float64))
		switch mainID {
		case msg.MsgID_Logon:
		case msg.MsgID_Server:
			game.GetManager().OnMessage(sendUser,data)
		case msg.MsgID_Game:
			game.GetManager().OnGameMessage(sendUser,data)
		}
	}
}
