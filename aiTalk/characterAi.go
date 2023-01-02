package aiTalk

import (
	"Lealra/config"
	"Lealra/myUtil"
	"encoding/json"
	"errors"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"math/rand"
	"strings"
	"time"
)

type Reply struct {
	Replies []Text `json:"replies"`
	Index   int
}

type Text struct {
	Text string `json:"text"`
	ID   int64  `json:"id"`
}

type ChatCreate struct {
	ExternalId   string `json:"external_id"`
	Created      string `json:"created"`
	Participants []struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
		IsHuman bool `json:"is_human"`
	} `json:"participants"`
	Status   string `json:"status"`
	Messages []struct {
		Text string `json:"text"`
	} `json:"messages"`
}

type LazyUuid struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

type AiChat struct {
	BotId         string
	Uuid          string
	CreatedInfo   ChatCreate
	Count         int
	Client        cycletls.CycleTLS
	ClientOptions cycletls.Options
	LastReply     Reply
}

const letters = "123567890abcdef"

func (aiChat *AiChat) SetBot(BotId string) (bool, error) {
	//给什么是什么
	aiChat.BotId = BotId
	aiChat.getLazyUuid()
	aiChat.initClient()
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

func (aiChat *AiChat) initClient() {
	client := cycletls.Init()
	options := cycletls.Options{
		Timeout: config.Settings.AntiCf.Timeout,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Proxy:     config.Settings.AntiCf.Proxy,
		Ja3:       config.Settings.AntiCf.Ja3,
		UserAgent: config.Settings.AntiCf.UserAgent,
	}
	aiChat.Client = client
	aiChat.ClientOptions = options
}

func (aiChat *AiChat) getLazyUuid() {
	//8+4+4+4+12的格式
	rand.Seed(time.Now().UnixNano())
	num := len(letters)
	var ret string
	for i := 0; i < 32; i++ {
		ret += string(letters[rand.Intn(num)])
		if i == 7 || i == 11 || i == 15 || i == 19 {
			ret += "-"
		}
	}
	aiChat.Uuid = ret
}
func (aiChat *AiChat) getToken() (bool, error) {
	aiChat.ClientOptions.Body = "{\"lazy_uuid\":\"" + aiChat.Uuid + "\"}"
	post, err := aiChat.Client.Do("https://beta.character.ai/chat/auth/lazy/", aiChat.ClientOptions, "POST")
	if err != nil {
		return false, err
	}
	var msg LazyUuid
	err = json.Unmarshal([]byte(post.Body), &msg)
	if err != nil {
		myUtil.ErrLog.Println(post.Body, aiChat.Uuid)
		return false, err
	}
	if !msg.Success {
		return false, errors.New("bad request,get token failed")
	}
	aiChat.ClientOptions.Headers["authorization"] = "Token " + msg.Token
	return true, nil
}
func (aiChat *AiChat) createConversation() (bool, error) {
	if aiChat.ClientOptions.Headers["authorization"] == "" {
		return false, errors.New("no token,please create")
	}
	var msg ChatCreate
	var post cycletls.Response
	var err error
	var success bool
	aiChat.ClientOptions.Body = "{\"character_external_id\": \"" + aiChat.BotId + "\",\"override_history_set\": null}"
	for i := 0; i < 3; i++ {
		post, err = aiChat.Client.Do("https://beta.character.ai/chat/history/create/", aiChat.ClientOptions, "POST")
		if err != nil {
			return false, err
		}
		err = json.Unmarshal([]byte(post.Body), &msg)
		if err != nil {
			return false, err
		}
		if len(msg.Participants) > 1 {
			success = true
			break
		}
	}
	if !success || msg.Status != "OK" {
		return false, errors.New("创建聊天时出现错误：\n" + post.Body)
	}
	aiChat.CreatedInfo = msg
	if len(msg.Messages) > 0 {
		aiChat.LastReply = Reply{Replies: []Text{{Text: msg.Messages[0].Text}}, Index: 0}
	} else {
		aiChat.LastReply = Reply{Replies: []Text{{Text: "该人格暂无问候语，请直接开始对话"}}, Index: 0}
	}
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
	aiChat.Count++
	var tgt string
	for _, v := range aiChat.CreatedInfo.Participants {
		if !v.IsHuman {
			tgt = v.User.Username
			break
		}
	}
	if tgt == "" {
		myUtil.ErrLog.Println("aiChat CreatedInfo.Participants 未获取到人格相关信息：", aiChat.CreatedInfo.Participants)
		return false, "", errors.New("未找到 ai 对应的 internal_id")
	}
	aiChat.ClientOptions.Body = "{\"history_external_id\": \"" + aiChat.CreatedInfo.ExternalId + "\",\"character_external_id\": \"" + aiChat.BotId + "\",\"text\": \"" + msg + "\",\"tgt\": \"" + tgt + "\",\"ranking_method\": \"random\",\"staging\": false,\"model_server_address\": null,\"override_prefix\": null,\"override_rank\": null,\"rank_candidates\": null,\"filter_candidates\": null,\"prefix_limit\": null,\"prefix_token_limit\": null,\"livetune_coeff\": null,\"stream_params\": null,\"enable_tti\": true,\"initial_timeout\": null,\"insert_beginning\": null,\"translate_candidates\": null,\"stream_every_n_steps\": 16,\"chunks_to_pad\": 8,\"is_proactive\": false,\"image_rel_path\": \"\",\"image_description\": \"\",\"image_description_type\": \"\",\"image_origin_type\": \"\",\"voice_enabled\": false}"
	post, err := aiChat.Client.Do("https://beta.character.ai/chat/streaming/", aiChat.ClientOptions, "POST")
	if err != nil {
		return false, "", err
	}
	res := strings.Split(post.Body, "          ")
	text := res[len(res)-1]
	var reply Reply
	err = json.Unmarshal([]byte(text), &reply)
	if err != nil {
		return false, "", err
	}
	if reply.Replies == nil {
		return false, "", errors.New("消息发送时出现问题 " + post.Body)
	}
	aiChat.LastReply = reply
	return true, reply.Replies[0].Text, nil
}
func (aiChat *AiChat) GetAnotherMsg() string {
	len := len(aiChat.LastReply.Replies)
	if len < 2 {
		return "暂无可更换的对话！"
	}
	aiChat.LastReply.Index++
	return aiChat.LastReply.Replies[aiChat.LastReply.Index%len].Text
}
