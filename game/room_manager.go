package game

import (
	"github.com/0990/shakeDiceServer/user"
	"sync"
	"github.com/0990/shakeDiceServer/msg"
	"github.com/0990/shakeDiceServer/util"
	"fmt"
	"time"
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

func (p *RoomManager) CreateRoom(user *user.User)  {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	//一个账号同时只有创建一个房间
	if room,ok:=p.GetRoomByCreatorID(user.ID());ok{
		sendMap := make(map[string]interface{})
		sendMap["success"] = false
		user.Send(msg.MainID_Server,msg.SCreateRoom,sendMap)
		syncMsg := make(map[string]interface{})
		syncMsg["roomID"] = room.id
		user.Send(msg.MainID_Server,msg.SSyncMyRoom,syncMsg)
		return
	}

	room:=newRoom(p.generateID(),user.ID())
	go room.Run()
	p.id2rooms[room.id] = room
	sendMap := make(map[string]interface{})
	sendMap["success"] = true
	//sendMap["roomID"] = room.id
	user.Send(msg.MainID_Server,msg.SCreateRoom,sendMap)

	syncMsg := make(map[string]interface{})
	syncMsg["roomID"] = room.id
	user.Send(msg.MainID_Server,msg.SSyncMyRoom,syncMsg)

	room.autoDestoryTimer = time.AfterFunc(600*time.Second,func(){
		sendMap := make(map[string]interface{})
		sendMap["roomID"] = room.id
		user.Send(msg.MainID_Server,msg.SDismissRoom,sendMap)
		room.Close()
	})
}

func(p *RoomManager)OnMessage(user *user.User,msgMap map[string]interface{}){
	subID := util.GetInt32(msgMap,"subID")
	switch subID {
	case msg.CCreateRoom:
		p.CreateRoom(user)
	case msg.CEnterRoom:
		roomID := util.GetInt32(msgMap,"roomID")
		p.EnterRoom(user,roomID)
	case msg.CGetMyRoom:
		var roomID int32 =0
		if room,ok:=p.GetRoomByCreatorID(user.ID());ok{
			roomID = room.id
		}else{
			roomID = -1
		}
		syncMsg := make(map[string]interface{})
		syncMsg["roomID"] = roomID
		user.Send(msg.MainID_Server,msg.SSyncMyRoom,syncMsg)
	}
}

func(p *RoomManager)GetRoomByCreatorID(userid int32)(*Room,bool){
	for _,v:=range p.id2rooms{
		if v.creatorid==userid{
			return v,true
		}
	}
	return nil,false
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
			if user==nil{
				continue
			}
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
