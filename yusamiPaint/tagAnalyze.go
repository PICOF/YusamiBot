package yusamiPaint

import (
	"Lealra/myUtil"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Ret struct {
	Data []string `json:"data,omitempty"`
}
type Send struct {
	Data []interface{} `json:"data,omitempty"`
}

func TagAnalyze(base64Pic string, msgID int64, isGroup bool, userID int64) (string, error) {
	send, _ := json.Marshal(Send{Data: []interface{}{base64Pic, 0.5}})
	post, err := http.Post("https://picof-yusamialchemy.hf.space/api/predict", "application/json", bytes.NewBuffer(send))
	if err != nil {
		myUtil.ErrLog.Println("获取tag解析结果时出现问题,error:", err)
		return "", err
	}
	defer post.Body.Close()
	body, err := ioutil.ReadAll(post.Body)
	var bd Ret
	err = json.Unmarshal(body, &bd)
	if bd.Data == nil {
		myUtil.ErrLog.Println("json反序列化tag解析结果时出现问题,body:"+string(body)+"\nerror:", err)
		return "服务器出现了神秘波动！", err
	}
	front := ""
	if isGroup {
		front = "[CQ:reply,id=" + strconv.FormatInt(msgID, 10) + "]" + "[CQ:at,qq=" + strconv.FormatInt(userID, 10) + "]" + " [CQ:at,qq=" + strconv.FormatInt(userID, 10) + "] "
	}
	info := GetInfoofAiPic(bd.Data[0])
	return front + info + "炼金结果为：\n" + strings.ReplaceAll(bd.Data[1], "\\", ""), err
}

func GetInfoofAiPic(input string) string {
	ret := ""
	if strings.Contains(input, "AI generated image") {
		compile := regexp.MustCompile("Description:\\[[^?\\]]*")
		compiledMsg := compile.FindString(input)[13:]
		ret += "发现遗留的神秘魔导书！上面好像写着：\ntags:" + compiledMsg + "\n"
		compile = regexp.MustCompile("Comment:\\[[^}]*")
		compiledMsg = compile.FindString(input)[10:]
		res := strings.Split(compiledMsg, ", \"")
		for _, i := range res {
			ret += strings.ReplaceAll(i, "\"", "") + "\n"
		}
		return ret
	} else {
		return ret
	}
}
func TagToPic(base64Pic string, msgID int64, isGroup bool, userID int64) (string, error) {
	send, _ := json.Marshal(Send{Data: []interface{}{base64Pic, 0.5}})
	post, err := http.Post("https://picof-yusamialchemy.hf.space/api/predict", "application/json", bytes.NewBuffer(send))
	if err != nil {
		myUtil.ErrLog.Println("获取tag解析结果时出现问题,error:", err)
		return "", err
	}
	defer post.Body.Close()
	body, err := ioutil.ReadAll(post.Body)
	var bd Ret
	err = json.Unmarshal(body, &bd)
	if bd.Data == nil {
		myUtil.ErrLog.Println("json反序列化tag解析结果时出现问题,error:", err, "body:", string(body))
		return "服务器出现了神秘波动！", err
	}
	front := ""
	if isGroup {
		front = "[CQ:reply,id=" + strconv.FormatInt(msgID, 10) + "]" + "[CQ:at,qq=" + strconv.FormatInt(userID, 10) + "]" + " [CQ:at,qq=" + strconv.FormatInt(userID, 10) + "] "
	}
	tags := strings.ReplaceAll(bd.Data[1], "\\", "")
	pic, err := GetPic(tags, "", "Landscape", "0", "")
	if err != nil {
		return "", err
	}
	return front + "炼金结果为：\n" + pic, err
}
