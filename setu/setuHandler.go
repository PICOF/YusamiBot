package setu

import (
	"Lealra/config"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"bytes"
	"encoding/json"
	"github.com/dlclark/regexp2"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var R18List = []string{"r18", "R18", "nsfw", "NSFW"}
var CNtoNum = map[string]string{
	"一": "1",
	"二": "2",
	"两": "2",
	"三": "3",
	"四": "4",
	"五": "5",
}

func translate(str string) string {
	for k, v := range CNtoNum {
		str = strings.ReplaceAll(str, k, v)
	}
	return str
}

func isR18(str string) (string, int) {
	res := 0
	for _, v := range R18List {
		if strings.Contains(str, v) {
			str = strings.ReplaceAll(str, v, "")
			res = 1
			break
		}
	}
	return str, res
}
func GetSetu(ws *websocket.Conn, mjson returnStruct.Message) (string, error) {
	var res string
	var err error
	var num int
	var tag []string
	str := translate(mjson.RawMessage)
	compile := regexp2.MustCompile("(?<=给[0-9]*张|来[0-9]*张|来点|给点|我要[0-9]*张).+(?=涩图|色图|🐍图)", 0)
	matched, _ := compile.FindStringMatch(str)
	if matched == nil {
		return "", err
	}
	pure, r18 := isR18(matched.String())
	if pure == "随机" {
		tag = []string{}
	} else {
		tag = strings.Split(pure, "、")
	}
	compile = regexp2.MustCompile("[0-9]+(?=张)", 0)
	matched, _ = compile.FindStringMatch(str)
	if matched == nil {
		num = 1
	} else {
		num, err = strconv.Atoi(matched.String())
		if matched == nil {
			myUtil.ErrLog.Println("解析请求色图张数时出现问题：", err)
			return "请仔细检查输入哦", err
		}
	}
	res, err = loliconApi(tag, num, r18, ws, mjson)
	if err != nil {
		myUtil.ErrLog.Println("使用loliconApi时出现问题：", err)
		return "wifi信号丢失力！", err
	}
	if res != "" {
		return res, nil
	} else {
		return "over", nil
	}
}

type Setu struct {
	Data  []Datum `json:"data"`
	Error string  `json:"error"`
}

type Datum struct {
	Author     string   `json:"author,omitempty"`
	PID        int64    `json:"pid,omitempty"`
	R18        bool     `json:"r18,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Title      string   `json:"title,omitempty"`
	Uid        int64    `json:"uid,omitempty"`
	UploadDate int64    `json:"uploadDate,omitempty"`
	Urls       Urls     `json:"urls,omitempty"`
}

type Urls struct {
	Original string `json:"original"`
}
type lolicon struct {
	R18   int      `json:"r18"`
	Level string   `json:"level,omitempty"`
	Tag   []string `json:"tag"`
	Num   int      `json:"num"`
}

func loliconApi(tag []string, num int, r18 int, ws *websocket.Conn, mjson returnStruct.Message) (string, error) {
	var level string
	if r18 == 1 {
		level = "5-6"
	} else {
		level = "0-3"
	}
	if num > 5 {
		num = 5
	} else if num < 0 {
		num = 1
	}
	ret, err := json.Marshal(lolicon{R18: r18, Num: num, Tag: tag, Level: level})
	if err != nil {
		myUtil.ErrLog.Println("序列化涩图请求时出现问题啦：json.Marshal error:", err)
		return "请检查参数输入是否有误哦~", err
	}
	v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{}}
	if mjson.GroupID == 0 {
		v.Param.UserID = mjson.UserID
	} else {
		v.Param.GroupID = mjson.GroupID
	}
	v.Param.Message = "正在努力搬运中~"
	marshal, err := json.Marshal(v)
	if err != nil {
		myUtil.ErrLog.Println("发送涩图通知时出现异常：", err)
		return "今天还是歇歇养生一下吧~", err
	}
	myUtil.WsLock.Lock()
	err = ws.WriteMessage(returnStruct.MsgType, marshal)
	myUtil.WsLock.Unlock()
	resp, err := http.Post(config.Settings.Setu.Api, "application/json", bytes.NewBuffer(ret))
	if err != nil {
		myUtil.ErrLog.Println("获取涩图地址时出现错误：Error:", err)
		return "被河蟹啦！", err
	}
	//之前有憨批在还没判断连接是否成功前就试图defer关闭，我不好说
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var se Setu
	err = json.Unmarshal(body, &se)
	if err != nil {
		myUtil.ErrLog.Println("解析涩图相关源信息时出现错误：Error:", err)
		return "被河蟹啦！", err
	}
	if len(se.Data) == 0 {
		return "啊咧，是不是你的xp太冷门了~", nil
	}
	myUtil.MsgLog.Println("即将发送", len(se.Data), "张涩图")
	if len(se.Data) == 1 {
		var pic string
		e := se.Data[0]
		pic, err = GetPic(e.Urls.Original, e.R18)
		v.Param.Message = pic + "\n作者：" + e.Author + "\n标题：" + e.Title + "\npid：" + strconv.FormatInt(e.PID, 10) + "\n是否NSFW：" + strconv.FormatBool(e.R18)
		marshal, err = json.Marshal(v)
		if err != nil {
			myUtil.ErrLog.Println("发送涩图时出现异常：", err)
			return "我图图呢？", err
		}
		myUtil.WsLock.Lock()
		err = ws.WriteMessage(returnStruct.MsgType, marshal)
		myUtil.WsLock.Unlock()
	} else {
		var data []string
		for _, e := range se.Data {
			//var t string
			//if e.R18{
			//	t=",type=flash"
			//}
			pic, err := GetPic(e.Urls.Original, e.R18)
			if err != nil {
				myUtil.ErrLog.Println("加载涩图图源时出现错误：Error:", err)
				return "我图图呢？", err
			}
			data = append(data, pic+"\n作者："+e.Author+"\n标题："+e.Title+"\npid："+strconv.FormatInt(e.PID, 10)+"\nNSFW："+strconv.FormatBool(e.R18))
		}
		err = myUtil.SendForwardMsg(data, mjson, ws)
	}
	if err != nil {
		myUtil.ErrLog.Println("传输涩图时出现错误：", err)
		return "刚刚一个不小心把图片搞丢了，欸嘿嘿……", err
	}
	return "", err
}

func GetPic(website string, nsfw bool) (string, error) {
	get, err := http.Get(website)
	if err != nil {
		myUtil.ErrLog.Println("请求涩图网站时出现异常！error:", err)
		return "我的小黄书呢QAQ", err
	}
	defer get.Body.Close()
	content, err := ioutil.ReadAll(get.Body)
	if nsfw {
		filename := website[strings.LastIndex(website, "/")+1:]
		err = ioutil.WriteFile("pixiv/h/"+filename, content, 0666)
		if err != nil {
			myUtil.ErrLog.Println("保存pixiv🐍图时出现异常！error:", err)
		}
		res, err := myUtil.SeseQrcode("something/review", filename)
		if err != nil {
			return "不给看！", err
		}
		return "[CQ:image,file=base64://" + res + "]", nil
	} else {
		filename := website[strings.LastIndex(website, "/")+1:]
		err = ioutil.WriteFile("pixiv/safe/"+filename, content, 0666)
		if err != nil {
			myUtil.ErrLog.Println("保存pixiv🐍图时出现异常！error:", err)
		}
		res, err := myUtil.SeseQrcode("something/plan", filename)
		if err != nil {
			return "不给看！", err
		}
		return "[CQ:image,file=base64://" + res + "]", nil
	}
}
