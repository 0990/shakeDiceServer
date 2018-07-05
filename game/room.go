package game

import (
	"encoding/json"
	"fmt"
	"github.com/0990/shakeDiceServer/user"
	"github.com/0990/shakeDiceServer/util"
	"github.com/0990/shakeDiceServer/msg"
	"github.com/looplab/fsm"
	"math/rand"
	"time"
)

const Char_Count int = 2

type Room struct {
	id         int32
	creatorid  int32
	users      [Char_Count]*userParam
	workerChan chan func()
	manager    *RoomManager
	FSM *fsm.FSM
	currOptSeat int
	callParam CallParam
	winSeat int
	autoDestoryTimer *time.Timer
}

type CallParam struct{
	init bool
	count int32
	diceNum int32
}

type userParam struct {
	*user.User
	isReady bool
	isClientReady bool
	seat int
	dice [4]int
}

func newRoom(id int32,creatorid int32)*Room{
	room:=&Room{
		id:         id,
		creatorid:  creatorid,
		workerChan: make(chan func(), 100),
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
		r.autoDestoryTimer.Stop()
		rand.Seed(time.Now().UnixNano())
		r.currOptSeat  = rand.Intn(Char_Count)
		for _,v:=range r.users{
			for i:=0;i<len(v.dice);i++{
				v.dice[i] = rand.Intn(6)+1
			}
			sendMap := make(map[string]interface{})
			sendMap["dice"] = v.dice
			sendMap["startSeat"] = r.currOptSeat
			r.SendGameMsg(v,msg.SGameStart,sendMap)
		}

	case "over":
		sendMap := make(map[string]interface{})
		sendMap["winSeat"] = r.winSeat
		r.SendGameMsg2All(msg.SGameEnd,sendMap)
		r.Close()
	}
}


func (r *Room) EnterUser(user *user.User) bool {
	userID := user.ID()
	if _, ok := r.GetUser(userID); ok {
		return false
	}
	emptySeat,ok:=r.GetEmptySeat()
	if !ok{
		return false
	}
	userParm:=&userParam{
		User:    user,
		isReady: false,
		seat:emptySeat,
	}
	r.users[emptySeat] = userParm
	GetManager().AttachUserID2RoomID(user.ID(), r)
	return true
}

func (r *Room) IsReadyGameStart() bool {
	readyCount:=0
	for _, v := range r.users {
		if v.isReady {
			readyCount++
		}
	}
	return readyCount>=Char_Count
}

//game message
func (r *Room) OnGameMessage(userid int32, msgMap map[string]interface{}) {
	user, ok := r.GetUser(userid)
	if !ok {
		return
	}
	subID := util.GetInt32(msgMap,"subID")
	switch subID {
	case msg.CClientReady:
		user.isClientReady = true
		var startUser []*userParam
		for i,v:=range r.users{
			if v.ID()==userid{
				startUser=append(r.users[i:],r.users[:i]...)
				break
			}
		}

		for _, v := range startUser{
			sendMap := make(map[string]interface{})
			sendMap["userID"] = v.ID()
			sendMap["nickname"] = v.Nickname()
			sendMap["seat"] = v.seat
			sendMap["ready"] = v.isReady
			v.Send(msg.MainID_Game,msg.SSyncUser,sendMap)
		}
	case msg.CReady:
		user.isReady = true
		sendMap := make(map[string]interface{})
		sendMap["seat"] = user.seat
		r.SendGameMsg2All(msg.SReady,sendMap)
		//game start
		if r.IsReadyGameStart() {
			r.FSM.Event("start")
		}
	case msg.CCallRoll:
		if user.seat!= r.currOptSeat{
			return
		}
		r.currOptSeat = (r.currOptSeat+1)%Char_Count
		//3个5 count个diceNum
		count := util.GetInt32(msgMap,"count")
		diceNum:= util.GetInt32(msgMap,"diceNum")
		if count<1||diceNum<1||diceNum>6{
			fmt.Println("callRoll param error")
			return
		}

		if count<= r.callParam.count&&diceNum<=r.callParam.diceNum{
			fmt.Println("callRoll param error")
			return
		}

		if !r.callParam.init{
			r.callParam.init  = true
		}

		r.callParam.count = count
		r.callParam.diceNum = diceNum
		sendMap := make(map[string]interface{})
		sendMap["optSeat"] = user.seat
		sendMap["nextOptSeat"] = r.currOptSeat
		sendMap["count"] = count
		sendMap["diceNum"] = diceNum
		r.SendGameMsg2All(msg.SCallRoll,sendMap)
	case msg.COpen:
		if user.seat!= r.currOptSeat{
			return
		}

		if !r.callParam.init{
			return
		}
		//计算diceNum的数量
		var diceNumCount int32 =0
		for _,user:=range r.users{
			for _,v:=range user.dice{
				if v==int(r.callParam.diceNum)||v==1{
					diceNumCount++
				}
			}
		}

		if diceNumCount < r.callParam.diceNum{
			r.winSeat = r.currOptSeat
		}else{
			r.winSeat = (r.currOptSeat+1)%Char_Count
		}

		sendMap := make(map[string]interface{})
		sendMap["optSeat"] = user.seat
		sendMap["count"] = r.callParam.count
		sendMap["diceNum"] = r.callParam.diceNum
		diceArr :=make([][4]int,0)
		for _,v:=range r.users{
			diceArr = append(diceArr,v.dice)
		}
		sendMap["diceArr"] = diceArr
		r.SendGameMsg2All(msg.SOpen,sendMap)
		time.Sleep(time.Second*3)
		r.FSM.Event("over")
	}
}

func(r *Room)SendGameMsg2All(subID int,sendMap map[string]interface{}){
	sendMap["mainID"] = msg.MainID_Game
	sendMap["subID"] = subID
	if sendBytes, err:= json.Marshal(sendMap);err!=nil{
		r.SendToAll(sendBytes)
	}
}

func(r *Room)GetSeatUser(seat int)*userParam{
	for _,v:=range r.users{
		if v.seat==seat{
			return v
		}
	}
	return nil
}

func(r *Room)GetEmptySeat()(int,bool){
	for i,v:=range r.users{
		if v==nil{
			return i,true
		}
	}
	return 0,false
}

func(r *Room)SendGameMsg(user *userParam,subID int,sendMap map[string]interface{}){
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
