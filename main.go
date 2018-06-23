package main

import (
	"fmt"
	"github.com/0990/simpleGameServer/net"
	"github.com/0990/simpleGameServer/user"
)

func main() {
	workerChan := make(chan func(), 100)
	go func() {
		for workerFunc := range workerChan {
			workerFunc()
		}
	}()
	net.RegisterConnectFun(func(client *net.Client) {
		user.GetManager().CreateUser(client)
	})
	net.RegisterDisconnectFun(func(client *net.Client) {
		user.GetManager().DestroyUser(client)
	})
	net.RegisterMessageFun(func(client *net.Client, message []byte) {
		sendUser, ok := user.GetManager().GetUserByClient(client)
		if ok {
			fmt.Println("sendUser:", sendUser, ",message:", message)
		}
		users := user.GetManager().GetUsers()
		for _, user := range users {
			user.SendMsg(message)
		}
	})
	net.Run(workerChan)
}
