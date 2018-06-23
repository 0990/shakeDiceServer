package user

import "github.com/0990/simpleGameServer/net"

type User struct {
	id       int32  //唯一id
	nickname string //昵称
	*net.Client
}
