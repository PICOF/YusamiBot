package news

import (
	"Lealra/config"
	"Lealra/myUtil"
	"encoding/base64"
	"io/ioutil"
	"net/http"
)

func GetNews60s() (string, error) {
	get, err := http.Get("https://api.vvhan.com/api/60s")
	if err != nil {
		myUtil.ErrLog.Println("请求新闻60s时出现异常！error:", err)
		return "最好的消息就是没有消息\n——" + config.Settings.BotName.Name + " 如是说", err
	}
	defer get.Body.Close()
	content, err := ioutil.ReadAll(get.Body)
	if err != nil {
		myUtil.ErrLog.Println("解析新闻60s返回结果时出现异常！error:", err)
		return "最好的消息就是没有消息\n——" + config.Settings.BotName.Name + " 如是说", err
	}
	return "[CQ:image,file=base64://" + base64.StdEncoding.EncodeToString(content) + "]", nil
}
