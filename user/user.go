package user

import "github.com/0990/shakeDiceServer/net"

type User struct {
	id       int32  //唯一id
	nickname string //昵称
	*net.Client
}

func (u *User) ID() int32 {
	return u.id
}
