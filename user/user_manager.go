package user

import "github.com/0990/shakeDiceServer/net"

var id int32 = 0
var m *UserManager

type UserManager struct {
	id2users    map[int32]*User
	client2user map[*net.Client]*User
}

func GetManager() *UserManager {
	if m == nil {
		m = &UserManager{
			id2users:    make(map[int32]*User),
			client2user: make(map[*net.Client]*User),
		}
	}
	return m
}

func (p *UserManager) generateID() int32 {
	id++
	return id
}

func (p *UserManager) CreateUser(client *net.Client) {
	user := User{
		id:     p.generateID(),
		Client: client,
	}
	p.id2users[user.id] = &user
	p.client2user[client] = &user
}

func (p *UserManager) DestroyUser(client *net.Client) {
	if user, ok := p.client2user[client]; ok {
		delete(p.id2users, user.id)
		delete(p.client2user, client)
	}
}

func (p *UserManager) GetUsers() map[int32]*User {
	return p.id2users
}

func (p *UserManager) GetUserByClient(client *net.Client) (*User, bool) {
	user, ok := p.client2user[client]
	return user, ok
}
