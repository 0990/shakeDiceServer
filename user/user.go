package user

import (
	"github.com/0990/shakeDiceServer/net"
	"encoding/json"
	"github.com/0990/shakeDiceServer/msg"
)

type User struct {
	logoned bool  //登录状态
	id       int32  //唯一id
	nickname string //昵称
	*net.Client
}

func (u *User) ID() int32 {
	return u.id
}

func(u *User)Nickname()string{
	return u.nickname
}

func(u *User)Login(msgMap map[string]interface{}){
	u.id = GetManager().generateID()
	u.nickname = msgMap["nickname"].(string)
	u.logoned = true

	sendMap := make(map[string]interface{})
	sendMap["userID"] = u.id
	sendMap["nickname"] = u.nickname
	u.Send(msg.MainID_Logon,msg.SUserInfo,sendMap)
}

func(u *User)Send(mainID,subID int32,sendMap map[string]interface{}){
	if sendMap == nil{
		sendMap = make(map[string]interface{})
	}
	sendMap["mainID"] = mainID
	sendMap["subID"] = subID
	sendBytes, _ := json.Marshal(sendMap)
	u.SendMsg(sendBytes)
}

func(u *User)IsLogined()bool{
	return u.logoned
}

