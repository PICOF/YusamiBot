package returnStruct

const MsgType = 1

type Message struct {
	Anonymous     interface{} `json:"anonymous"`
	Font          int64       `json:"font"`
	GroupID       int64       `json:"group_id"`
	Message       string      `json:"message"`
	MessageID     int64       `json:"message_id"`
	MessageSeq    int64       `json:"message_seq"`
	MessageType   string      `json:"message_type"`
	PostType      string      `json:"post_type"`
	MetaEventType string      `json:"meta_event_type"`
	RawMessage    string      `json:"raw_message"`
	SelfID        int64       `json:"self_id"`
	Sender        Sender      `json:"sender"`
	SubType       string      `json:"sub_type"`
	Time          int64       `json:"time"`
	UserID        int64       `json:"user_id"`
	Echo          string      `json:"echo"`
	RetData       Data        `json:"data"`
	RetCode       int         `json:"retcode"`
	Wording       string      `json:"wording"`
	Content       string      `json:"content"`
}

type Data struct {
	MessageID int32     `json:"message_id"`
	Sender    Sender    `json:"sender"`
	Time      int32     `json:"time"`
	Message   string    `json:"message"`
	GroupID   int64     `json:"group_id"`
	Messages  []Message `json:"messages"`
}

type Sender struct {
	Age      int64  `json:"age"`
	Area     string `json:"area"`
	Card     string `json:"card"`
	Level    string `json:"level"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Sex      string `json:"sex"`
	Title    string `json:"title"`
	UserID   int64  `json:"user_id"`
}
