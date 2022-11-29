package returnStruct

type SendMsg struct {
	Action string `json:"action"`
	Param  Params `json:"params"`
}
type Params struct {
	UserID     int64  `json:"user_id"`
	GroupID    int64  `json:"group_id"`
	Message    string `json:"message"`
	Messages   []Node `json:"messages"`
	AutoEscape bool   `json:"auto_escape"`
	MessageId  int32  `json:"message_id"`
}

type Node struct {
	Type string `json:"type"`
	Data NData  `json:"data"`
}

type NData struct {
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Uid     int64       `json:"uin,omitempty"`
	Content interface{} `json:"content,omitempty"`
	Time    int64       `json:"time,omitempty"`
	Text    string      `json:"text,omitempty"`
}
