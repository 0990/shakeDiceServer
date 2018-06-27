package msg

const(
	MsgID_Invalid = iota
	MsgID_Logon
	MsgID_Server
	MsgID_Game
)

const(
	CCreateRoom = iota
	CEnterRoom

	SCreateRoom = iota+100
	SEnterRoom
)

const(
	CReady = iota
	CCallRoll //报数

	SSyncUser = iota+100
	SReady
	SGameStart
	SCallRoll
	SGameEnd
)


type JsonMsg struct{
	MsgID int `json:"msgID"`
	Content []byte `json:"content"`
}

//msgid:1
type ReqLogon struct{
	Nickname string `json:"nickname"`
}

//msgid:2
type ReqCreateRoom struct{
}
//msgid 12
type RespCreateRoom struct{
	success bool `json:"success"`
	roomID int `json:"roomID"`
}

//msgid:3
type ReqEnterRoom struct{
}
//msgid:13
type RespEnterRoom struct{
	success bool `json:"success"`
	roomID int `json:"roomID"`
}

type SyncUserInfo struct{
	userID int `json:"userID"`
}

//msgid:5
type GameReady struct{
}

