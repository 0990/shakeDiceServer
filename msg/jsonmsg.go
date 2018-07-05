package msg

const(
	MainID_Invalid = iota
	MainID_Logon
	MainID_Server
	MainID_Game
)
const (
	CLogon = iota
	SUserInfo = 100
)
const(
	CCreateRoom = iota
	CEnterRoom
	CGetMyRoom

	SCreateRoom      = 100
	SEnterRoom       = 101
	SSyncMyRoom      = 102
)

const(
	CClientReady = iota
	CReady
	CCallRoll //报数
	COpen //开

	SSyncUser = 100
	SReady = 101
	SGameStart = 102
	SCallRoll = 103
	SOpen = 104
	SGameEnd = 105
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

