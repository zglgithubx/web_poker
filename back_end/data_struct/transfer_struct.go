package data_struct

type TransferMessage struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}
type ResponseMessage struct {
	//-2: 不该你出牌，-1: 格式不对重新出牌，0 : 不出，1：压不过，重新出牌，2：出牌
	SendCardStatus int   `json:"sendCardStatus"`
	CurSendCard    Cards `json:"curSendCard"`
	//下一个出牌的用户id
	NextUserId string `json:"nextUserId"`
	//当前响应消息的描述
	Describe string `json:"describe"`
}
