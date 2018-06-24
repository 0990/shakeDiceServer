package game

import (
	"github.com/0990/simpleGameServer/user"
	"sync"
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

func (p *RoomManager) CreateRoom(userid int32) int32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	room := Room{
		id:         p.generateID(),
		creatorid:  userid,
		workerChan: make(chan func(), 100),
		users:      make([]*userParam, 0),
	}
	go room.Run()
	p.id2rooms[room.id] = &room
	return room.id
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

func (p *RoomManager) EnterRoom(roomid int32, user *user.User) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if room, ok := p.GetRoom(roomid); ok {
		room.Post(func() {
			room.EnterUser(user)
		})
		return true
		//if success := room.EnterUser(user); success {
		//	p.userid2rooms[user.ID()] = room
		//	return true
		//}
	}
	return false
}
