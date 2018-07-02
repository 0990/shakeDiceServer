package game

import (
	"encoding/json"
	"fmt"
	"github.com/0990/shakeDiceServer/user"
	"github.com/0990/shakeDiceServer/util"
	"github.com/0990/shakeDiceServer/msg"
	"github.com/looplab/fsm"
)

const Char_Count int = 2

type Room struct {
	id         int32
	creatorid  int32
	users      []*userParam
	workerChan chan func()
	manager    *RoomManager
	FSM *fsm.FSM
}

type userParam struct {
	*user.User
	isReady bool
	isClientReady bool
	seat int
}

func newRoom(id int32,creatorid int32)*Room{
	room:=&Room{
		id:         id,
		creatorid:  creatorid,
		workerChan: make(chan func(), 100),
		users:      make([]*userParam, 0),
	}
	room.FSM = fsm.NewFSM(
		"ready",
		fsm.Events{
			{Name:"start",Src:[]string{"ready"},Dst:"start"},
			{Name:"over",Src:[]string{"ready","start"},Dst:"over"},
		},
		fsm.Callbacks{
			"enter_state":func(e *fsm.Event) {room.enterState(e)},
		},
	)
	return room
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

func(r *Room)enterState(e *fsm.Event){
	switch e.Dst {
	case "start":
		sendMap := make(map[string]interface{})
		r.SendGameMsg2All(msg.SGameStart,sendMap)
	case "over":
		sendMap := make(map[string]interface{})
		sendMap["winUser"] = 1
		r.SendGameMsg2All(msg.SGameEnd,sendMap)
		r.Close()
	}
}


func (r *Room) EnterUser(user *user.User) bool {
	userID := user.ID()
	if _, ok := r.GetUser(userID); ok {
		return false
	}

	if len(r.users) >= Char_Count{
		return false
	}
	userParm:=&userParam{
		User:    user,
		isReady: false,
	}

	for i:=0;i<Char_Count;i++{
		existed:=false
		for _,v:=range r.users{
			if v.seat == i{
				existed = true
			}
		}
		if !existed{
			userParm.seat = i
		}
	}

	r.users = append(r.users, userParm)
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
	case msg.CClientReady:
		userParam.isClientReady = true
		for _, v := range r.users {
			sendMap := make(map[string]interface{})
			sendMap["userid"] = v.ID()
			sendMap["nickname"] = v.Nickname()
			sendMap["seat"] = v.seat
			v.Send(msg.MainID_Game,msg.SSyncUser,sendMap)
		}
	case msg.CReady:
		fmt.Println("user send ready")
		//ready
		userParam.isReady = true
		sendMap := make(map[string]interface{})
		sendMap["seat"] = userParam.seat
		userParam.Send(msg.MainID_Game,msg.SReady,sendMap)
		//game start
		if r.IsReadyGameStart() {
			r.FSM.Event("start")
		}
	case msg.CCallRoll:
		//play card
	default:

	}
}

func(r *Room)SendGameMsg2All(subID int,sendMap map[string]interface{}){
	sendMap["mainID"] = msg.MainID_Game
	sendMap["subID"] = subID
	if sendBytes, err:= json.Marshal(sendMap);err!=nil{
		r.SendToAll(sendBytes)
	}
}

func(r *Room)SendGameMsg(user *user.User,subID int,sendMap map[string]interface{}){
	sendMap["mainID"] = msg.MainID_Game
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
		if v.isClientReady{
			v.SendMsg(message)
		}
	}
}
