package returnStruct

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetForwardMsgNodes(ret []Node, id string) ([]Node, error) {
	get, err := http.Get("http://127.0.0.1:5700/get_forward_msg?message_id=" + id)
	if err != nil {
		return nil, err
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var msg Message
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return nil, err
	}
	for _, o := range msg.RetData.Messages {
		a := Node{}
		a.Data.Name = o.Sender.Nickname
		a.Data.Uid = o.Sender.UserID
		a.Type = "node"
		a.Data.Time = o.Time
		a.Data.Content = o.Content
		ret = append(ret, a)
	}
	return ret, err
}
