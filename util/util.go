package util

func GetInt32(msgMap map[string]interface{},name string)int32{
	return int32(msgMap[name].(float64))
}
