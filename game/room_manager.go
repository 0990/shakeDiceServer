package game

import (
	"github.com/0990/shakeDiceServer/user"
	"sync"
	"github.com/0990/shakeDiceServer/msg"
	"github.com/0990/shakeDiceServer/util"
	"fmt"
)

var id int32 = 0
var m *RoomManager

type RoomManager struct {
	id2rooms     map[int32]*Room
	userid2rooms map[int32]*Room
	mutex        sync.RWMutex
}

func GetManager() *RoomManager {
	if m == nil {
		m = &RoomManager{
			id2rooms:     make(map[int32]*Room),
			userid2rooms: make(map[int32]*Room),
		}
	}
	return m
}

func (p *RoomManager) generateID() int32 {
	id++
	return id
}

func (p *RoomManager) CreateRoom(user *user.User) int32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	room:=newRoom(p.generateID(),user.ID())
	go room.Run()
	p.id2rooms[room.id] = room
	sendMap := make(map[string]interface{})
	sendMap["roomID"] = room.id
	sendMap["success"] = true
	user.Send(msg.MainID_Server,msg.SCreateRoom,sendMap)
	return room.id
}

func(p *RoomManager)OnMessage(user *user.User,msgMap map[string]interface{}){
	subID := util.GetInt32(msgMap,"subID")
	switch subID {
	case msg.CCreateRoom:
		p.CreateRoom(user)
	case msg.CEnterRoom:
		roomID := util.GetInt32(msgMap,"roomID")
		p.EnterRoom(user,roomID)
	}
}

func(p *RoomManager)OnGameMessage(user *user.User,msgMap map[string]interface{}){
	if room, ok := p.GetRoomByUserid(user.ID()); ok {
		room.Post(func() {
			room.OnGameMessage(user.ID(), msgMap)
		})
	}else{
		fmt.Println("onGameMessage,not in game!",msgMap)
	}
}

func (p *RoomManager) DestroyRoom(roomid int32) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if room, ok := p.id2rooms[roomid]; ok {
		for _, user := range room.users {
			delete(p.userid2rooms, user.ID())
		}
		delete(p.id2rooms, roomid)
	}
}

func (p *RoomManager) GetRoom(roomid int32) (*Room, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	user, ok := p.id2rooms[roomid]
	return user, ok
}

func (p *RoomManager) GetRoomByUserid(userid int32) (*Room, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	room, ok := p.userid2rooms[userid]
	return room, ok
}

func (p *RoomManager) AttachUserID2RoomID(userid int32, room *Room) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	p.userid2rooms[userid] = room
}

func (p *RoomManager) EnterRoom( user *user.User,roomid int32) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if room, ok := p.GetRoom(roomid); ok {
		room.Post(func() {
			ok:= room.EnterUser(user)
			sendMap := make(map[string]interface{})
			sendMap["roomID"] = room.id
			sendMap["success"] = ok
			user.Send(msg.MainID_Server,msg.SEnterRoom,sendMap)
		})
		return true
	}
	return false
}
