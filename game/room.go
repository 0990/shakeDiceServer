package game

import (
	"encoding/json"
	"fmt"
	"github.com/0990/shakeDiceServer/user"
	"github.com/0990/shakeDiceServer/util"
	"github.com/0990/shakeDiceServer/msg"
)

const Char_Count int = 2

type Room struct {
	id         int32
	creatorid  int32
	users      []*userParam
	workerChan chan func()
	manager    *RoomManager
}

type userParam struct {
	*user.User
	isReady bool
}

func (r *Room) Run() {
	go func() {
		fmt.Println("room start")
		for workerFunc := range r.workerChan {
			workerFunc()
		}
		fmt.Println("room end")
	}()
}

func (r *Room) Post(handler func()) {
	r.workerChan <- func() { handler() }
}

func (r *Room) Close() {
	close(r.workerChan)
	GetManager().DestroyRoom(r.id)
}

func (r *Room) EnterUser(user *user.User) bool {
	userID := user.ID()
	if _, ok := r.GetUser(userID); ok {
		return false
	}

	r.users = append(r.users, &userParam{
		User:    user,
		isReady: false,
	})

	GetManager().AttachUserID2RoomID(user.ID(), r)
	return true
}

func (r *Room) IsReadyGameStart() bool {
	if len(r.users) != Char_Count {
		return false
	}
	allReady := true
	for _, v := range r.users {
		if v.isReady == false {
			allReady = false
		}
	}
	return allReady
}

//game message
func (r *Room) OnGameMessage(userid int32, msgMap map[string]interface{}) {
	userParam, ok := r.GetUser(userid)
	if !ok {
		return
	}
	subID := util.GetInt32(msgMap,"subID")
	switch subID {
	case msg.CReady:
		fmt.Println("user send ready")
		//ready
		userParam.isReady = true
		//game start
		if r.IsReadyGameStart() {
			sendMap := make(map[string]interface{})
			r.SendGameMsg2All(msg.SGameStart,sendMap)
		}
	case msg.CCallRoll:
		//play card
	default:

	}
}

func(r *Room)onGameEnd(){
	sendMap := make(map[string]interface{})
	sendMap["winUser"] = 1
	r.SendGameMsg2All(msg.SGameEnd,sendMap)
	r.Close()
}

func(r *Room)SendGameMsg2All(subID int,sendMap map[string]interface{}){
	sendMap["mainID"] = msg.MsgID_Game
	sendMap["subID"] = subID
	if sendBytes, err:= json.Marshal(sendMap);err!=nil{
		r.SendToAll(sendBytes)
	}
}

func(r *Room)SendGameMsg(subID int,sendMap map[string]interface{},user *user.User){
	sendMap["mainID"] = msg.MsgID_Game
	sendMap["subID"] = subID
	if sendBytes, err:= json.Marshal(sendMap);err!=nil{
		user.SendMsg(sendBytes)
	}
}

func (r *Room) GetUser(userid int32) (*userParam, bool) {
	for _, v := range r.users {
		fmt.Println(v.ID())
		if v.ID() == userid {
			fmt.Println(v.ID(), userid)
			return v, true
		}
	}
	return nil, false
}

func (r *Room) SendToAll(message []byte) {
	for _, v := range r.users {
		v.SendMsg(message)
	}
}
