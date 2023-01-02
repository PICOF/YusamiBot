package aiTalk

import (
	"Lealra/config"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type OpenAiPersonal struct {
	Preset  string
	Context string
	Reply   string
	Memory  bool
}

type OpenAiReply struct {
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Text string `json:"text"`
}

func (openAi *OpenAiPersonal) EditMsg(text string, require string) (string, error) {
	encoded := new(bytes.Buffer)
	err := json.NewEncoder(encoded).Encode(text)
	if err != nil {
		return "哎呀，出错了！", err
	}
	body := "{\"model\": " + config.Settings.OpenAi.Setting.EditSetting.Model + ", \"input\": " + encoded.String() + ", \"instruction\": \"" + require + "\"}"
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/edits", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Settings.OpenAi.Token)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "哎呀，出错了！", err
	} else if resp.StatusCode != 200 {
		v, _ := ioutil.ReadAll(resp.Body)
		return "哎呀，出错了！", errors.New(resp.Status + " " + string(v) + " " + body)
	}
	defer resp.Body.Close()
	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	var reply OpenAiReply
	err = json.Unmarshal(content, &reply)
	if err != nil {
		return "哎呀，出错了！", err
	}
	return strings.Trim(strings.ReplaceAll(reply.Choices[0].Text, "\\n", "\n"), "\n "), nil
}

func (openAi *OpenAiPersonal) SendAndReceiveMsg(text string) (string, error) {
	switch text {
	case "撤回":
		openAi.Reply = ""
		text = ""
	case "刷新":
		openAi.Context = ""
		openAi.Reply = ""
		openAi.Preset = ""
		return "刷新成功！", nil
	case "配置":
		return "模型：" + config.Settings.OpenAi.Setting.ChatSetting.Model + "\n预设：" + openAi.Preset, nil
	default:
		text += "\n"
	}
	encoded := new(bytes.Buffer)
	err := json.NewEncoder(encoded).Encode(openAi.Preset + openAi.Context + openAi.Reply + text)
	if err != nil {
		return "哎呀，出错了！", err
	}
	body := "{\"model\": \"" + config.Settings.OpenAi.Setting.ChatSetting.Model + "\",\"prompt\": " + encoded.String() + ",\"max_tokens\": " + config.Settings.OpenAi.Setting.ChatSetting.MaxTokens + ",\"temperature\": " + config.Settings.OpenAi.Setting.ChatSetting.Temperature + ",\"top_p\": " + config.Settings.OpenAi.Setting.ChatSetting.TopP + ",\"frequency_penalty\": " + config.Settings.OpenAi.Setting.ChatSetting.FrequencyPenalty + ",\"presence_penalty\": " + config.Settings.OpenAi.Setting.ChatSetting.PresencePenalty + ",\"stream\": false}"
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Settings.OpenAi.Token)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "哎呀，出错了！", err
	} else if resp.StatusCode != 200 {
		v, _ := ioutil.ReadAll(resp.Body)
		return "哎呀，出错了！", errors.New(resp.Status + " " + string(v) + " " + body)
	}
	defer resp.Body.Close()
	var content []byte
	content, err = ioutil.ReadAll(resp.Body)
	var reply OpenAiReply
	err = json.Unmarshal(content, &reply)
	if err != nil {
		return "哎呀，出错了！", err
	}
	if openAi.Memory {
		openAi.Context += openAi.Reply
		openAi.Context += text
		openAi.Reply = strings.Trim(reply.Choices[0].Text, "\n ") + "\\n"
	}
	return strings.Trim(strings.ReplaceAll(reply.Choices[0].Text, "\\n", "\n"), "\n "), nil
}

func (openAi *OpenAiPersonal) SetPreset(preset string) {
	openAi.Preset = preset + "\\n"
}
