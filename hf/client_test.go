package hf

import (
	"encoding/json"
	"errors"
	"github.com/dlclark/regexp2"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"testing"
)

func TestGetClient(t *testing.T) {
	client, _ := GetClient()
	t.Log("client is created:", client)
}

func TestClient_SendMsg(t *testing.T) {
	client, err := GetClientWithUrl("https://huggingface.co/spaces/skytnt/moe-tts", false)
	if err != nil {
		println("SendMsg error:", err.Error())
		return
	}
	var res interface{}
	res, err = client.SendWsMsg(1, []interface{}{"こんにちは。", "綾地寧々", 1, false})
	if err != nil {
		println("SendMsg error:", err.Error())
		return
	}
	println(res)
}

func TestClient_InvokeFunc(t *testing.T) {
	client, err := GetClientWithUrl("https://huggingface.co/spaces/skytnt/moe-tts", true)
	if err != nil {
		println("SendMsg error:", err.Error())
		return
	}
	//var res interface{}
	//res, err = client.SendMsg(1, []interface{}{"こんにちは。", "綾地寧々", 1, false})
	//if err != nil {
	//	println("SendMsg error:", err.Error())
	//	return
	//}
	err = client.InvokeFunc(17, LinkModeWs)
	if err != nil {
		return
	}
	data, err := client.GetData(19)
	if err != nil {
		return
	}
	println(data.(string))
}

func TestGradioConfConvert(t *testing.T) {
	get, err := http.Get("https://sayashi-vits-models.hf.space/")
	if err != nil {
		logrus.Errorf("连接空间失败: %s", err.Error())
	}
	defer get.Body.Close()
	var b []byte
	b, err = io.ReadAll(get.Body)
	if err != nil {
		logrus.Errorf("读取返回消息时失败: %s", err.Error())
	}
	compile := regexp2.MustCompile("(?<=<script>window.gradio_config = ).*(?=;</script>)", 0)
	var match *regexp2.Match
	match, err = compile.FindStringMatch(string(b))
	if match == nil {
		if err == nil {
			err = errors.New("未找到对应字段")
		}
		logrus.Errorf("连接空间失败: %s", err.Error())
	}
	var conf SpaceConf
	err = json.Unmarshal([]byte(match.String()), &conf)
	if err != nil {
		logrus.Errorf("连接空间失败: %s", err.Error())
	}
	return
}
