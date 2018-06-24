package game

import (
	"encoding/json"
	"fmt"
	"github.com/0990/simpleGameServer/user"
	"time"
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
func (r *Room) OnGameMessage(userid int32, subID int32) {
	userParam, ok := r.GetUser(userid)
	if !ok {
		return
	}
	switch subID {
	case 1:
		fmt.Println("user send ready")
		//ready
		userParam.isReady = true

		//game start
		if r.IsReadyGameStart() {
			sendMap := make(map[string]interface{})
			sendMap["gameStart"] = true
			sendBytes, _ := json.Marshal(sendMap)
			r.SendToAll(sendBytes)
			time.Sleep(10 * time.Second)

			sendMapEnd := make(map[string]interface{})
			sendMapEnd["gameOver"] = true
			sendBytesEnd, _ := json.Marshal(sendMapEnd)
			r.SendToAll(sendBytesEnd)
			r.Close()

		}

	case 2:
		//play card
	default:

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
