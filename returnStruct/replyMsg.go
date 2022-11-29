package returnStruct

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func GetReplyMsg(mjson Message) (Message, error) {
	k, _ := strconv.ParseInt(mjson.RawMessage[13:strings.Index(mjson.RawMessage, "]")], 10, 32)
	ret, err := json.Marshal(SendMsg{Action: "get_msg", Param: Params{MessageId: int32(k)}})
	resp, err := http.Post("http://127.0.0.1:5700", "application/json", bytes.NewBuffer(ret))
	if err != nil {
		return Message{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var msg Message
	err = json.Unmarshal(body, &msg)
	return msg, err
}
