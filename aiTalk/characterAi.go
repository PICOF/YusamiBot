package aiTalk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Reply struct {
	Replies []Text `json:"replies"`
}

type Text struct {
	Text string `json:"text"`
	ID   int64  `json:"id"`
}

type ChatCreate struct {
	ExternalId   string        `json:"external_id"`
	Created      string        `json:"created"`
	Participants []Participant `json:"participants"`
}

type Participant struct {
	User User `json:"user"`
}

type User struct {
	Username string `json:"username"`
}

type LazyUuid struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

type AiChat struct {
	BotId       string
	Cookies     []*http.Cookie
	Token       string
	Uuid        string
	CreatedInfo ChatCreate
	Count       int
}

const letters = "123567890abcdefghijklmnopqrstuvwxyz"

func (aiChat *AiChat) SetBot(BotId string) (bool, error) {
	//给什么是什么
	aiChat.BotId = BotId
	aiChat.getLazyUuid()
	token, err := aiChat.getToken()
	if err != nil {
		return token, err
	}
	conversation, err := aiChat.createConversation()
	if err != nil {
		return conversation, err
	}
	return true, nil
}

func (aiChat *AiChat) getLazyUuid() {
	//8+4+4+4+11的格式
	rand.Seed(time.Now().UnixNano())
	num := len(letters)
	var ret string
	for i := 0; i < 31; i++ {
		ret += string(letters[rand.Intn(num)])
		if i == 7 || i == 11 || i == 15 || i == 19 {
			ret += "-"
		}
	}
	aiChat.Uuid = ret
}
func (aiChat *AiChat) getToken() (bool, error) {
	if aiChat.Uuid == "" {
		return false, nil
	}
	post, err := http.Post("https://beta.character.ai/chat/auth/lazy/", "application/json", bytes.NewBuffer([]byte("{\"lazy_uuid\":\""+aiChat.Uuid+"\"}")))
	if err != nil {
		return false, err
	}
	defer post.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(post.Body)
	var msg LazyUuid
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return false, err
	}
	if !msg.Success {
		return false, errors.New("bad request,get token failed")
	}
	aiChat.Token = msg.Token
	aiChat.Cookies = post.Cookies()
	return true, nil
}
func (aiChat *AiChat) createConversation() (bool, error) {
	if aiChat.Token == "" {
		return false, errors.New("no token,please create")
	}
	body := "{\"character_external_id\": \"" + aiChat.BotId + "\",\"override_history_set\": null}"
	req, _ := http.NewRequest("POST", "https://beta.character.ai/chat/history/create/", bytes.NewBuffer([]byte(body)))
	for _, c := range aiChat.Cookies {
		req.AddCookie(c)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", "Token "+aiChat.Token)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	var msg ChatCreate
	err = json.Unmarshal(content, &msg)
	if err != nil {
		return false, err
	}
	aiChat.CreatedInfo = msg
	return true, nil
}
func (aiChat *AiChat) Renew() (bool, error) {
	aiChat.getLazyUuid()
	ok, err := aiChat.getToken()
	if err != nil {
		return false, err
	}
	return ok, nil
}
func (aiChat *AiChat) SendMag(msg string) (bool, string, error) {
	//貌似非注册用户只有七次聊天机会，但是嘛……好像history填对了也可以无限续杯
	if aiChat.Count == 7 {
		aiChat.Renew()
		aiChat.Count = 0
	}
	body := "{\"history_external_id\": \"" + aiChat.CreatedInfo.ExternalId + "\",\"character_external_id\": \"" + aiChat.BotId + "\",\"text\": \"" + msg + "\",\"tgt\": \"" + aiChat.CreatedInfo.Participants[1].User.Username + "\",\"ranking_method\": \"random\",\"staging\": false,\"model_server_address\": null,\"override_prefix\": null,\"override_rank\": null,\"rank_candidates\": null,\"filter_candidates\": null,\"prefix_limit\": null,\"prefix_token_limit\": null,\"livetune_coeff\": null,\"stream_params\": null,\"enable_tti\": true,\"initial_timeout\": null,\"insert_beginning\": null,\"translate_candidates\": null,\"stream_every_n_steps\": 16,\"chunks_to_pad\": 8,\"is_proactive\": false,\"image_rel_path\": \"\",\"image_description\": \"\",\"image_description_type\": \"\",\"image_origin_type\": \"\",\"voice_enabled\": false}"
	req, _ := http.NewRequest("POST", "https://beta.character.ai/chat/streaming/", bytes.NewBuffer([]byte(body)))
	for _, c := range aiChat.Cookies {
		req.AddCookie(c)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", "Token "+aiChat.Token)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return false, "", err
	} else {
		aiChat.Count++
	}
	defer resp.Body.Close()
	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	res := strings.Split(string(content), "          ")
	text := res[len(res)-1]
	var reply Reply
	err = json.Unmarshal([]byte(text), &reply)
	if err != nil {
		return false, "", err
	}
	aiChat.Cookies = resp.Cookies()
	if reply.Replies == nil {
		return false, "", errors.New("可能是身份验证问题")
	}
	return true, reply.Replies[0].Text, nil
}
