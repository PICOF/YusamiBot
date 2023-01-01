package handler

import (
	"Lealra/config"
	"Lealra/myUtil"
	"Lealra/note"
	"Lealra/returnStruct"
	"Lealra/schoolTimeTable"
	"Lealra/yusamiPaint"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NetEaseResp struct {
	Code   int64  `json:"code"`
	Result Result `json:"result"`
}

type Result struct {
	SongCount int64  `json:"songCount"`
	Songs     []Song `json:"songs"`
}

type Song struct {
	Album   Album           `json:"album"`
	Artists []ArtistElement `json:"artists"`
	Fee     int64           `json:"fee"`
	ID      int64           `json:"id"`
	Name    string          `json:"name"`
	Status  int64           `json:"status"`
}

type Album struct {
	Name string `json:"name"`
}

type ArtistElement struct {
	Name string `json:"name"`
}

func getNetEaseMusicId(song string, random bool) (string, string, error) {
	if random {
		return "", "功能尚在开发中，敬请期待", nil
	} else {
		resp, err := http.Get("http://music.163.com/api/search/get/web?csrf_token=hlpretag=&hlposttag=&s=" + url.QueryEscape(song) + "&type=1&offset=0&total=true&limit=5")
		defer resp.Body.Close() //在回复后必须关闭回复的主体
		body, _ := ioutil.ReadAll(resp.Body)
		var song NetEaseResp
		json.Unmarshal(body, &song)
		var id string
		var intro string
		if song.Code != 200 {
			println("获取音乐搜索结果错误！")
			return "网络好像出了点小问题~", "", nil
		} else {
			for _, i := range song.Result.Songs {
				if i.Fee == 1 {
					continue
				} else {
					id = strconv.FormatInt(i.ID, 10)
					for _, j := range i.Artists {
						intro += j.Name + ","
					}
					intro = intro[:len(intro)-1] + " 的 "
					intro += i.Name + "\r\n来自专辑：\r\n《" + i.Album.Name + "》"
					break
				}
			}
		}
		return "即将为您播放：\r\n" + intro, id, err
	}
}
func Music(mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	m := strings.Fields(mjson.RawMessage)
	if m[1] == "help" {
		return "使用[网易云]+空格+[歌名]进行搜索，部分收费曲目可能无法播放", nil
	} else {
		if string([]rune(mjson.RawMessage)[4:]) == "随便听听" {
			_, res, err := getNetEaseMusicId("", true)
			return res, err
		} else {
			notice, res, err := getNetEaseMusicId(string([]rune(mjson.RawMessage)[4:]), false)
			if res != "" {
				v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: notice}}
				if mjson.GroupID != 0 {
					v.Param.GroupID = mjson.GroupID
				} else {
					v.Param.UserID = mjson.UserID
				}
				o, _ := json.Marshal(v)
				myUtil.WsLock.Lock()
				ws.WriteMessage(returnStruct.MsgType, o)
				myUtil.WsLock.Unlock()
				return "[CQ:record,file=http://music.163.com/song/media/outer/url?id=" + res + ".mp3]", err
			} else {
				return "未找到相关曲目，换个关键词再搜搜吧~", err
			}
		}
	}
}
func MakeChoice(mjson returnStruct.Message) (string, error) {
	m := mjson.RawMessage
	if pos := strings.Index(m, "不"); pos > 0 { //&&m[0][:6]=="[CQ:at" 本来有个这个限制的，但是感觉at了就不好玩了
		l := len([]rune(m[:pos]))
		if l == len([]rune(m))-1 {
			return "", nil
		} else if strings.Compare(string([]rune(m)[l-1]), string([]rune(m)[l+1])) == 0 {
			rand.Seed(time.Now().Unix())
			myrand := rand.Intn(5)
			switch myrand {
			case 0:
				return "神说这是极好的~", nil
			case 1:
				return config.Settings.BotName.Name + " 帮你投硬币！（抛起）是正面！ヾ(≧▽≦*)o", nil
			case 2:
				return config.Settings.BotName.Name + " 帮你投硬币！（抛起）是反面！（；´д｀）ゞ", nil
			case 3:
				return string([]rune(m)[l-1]) + "!", nil
			case 4:
				return "还是不" + string([]rune(m)[l-1]) + "了吧……＞﹏＜", nil
			}
		}
	}
	return "", nil
}
func SchoolTimeTable(m string, uid int64) (string, error) {
	var class []schoolTimeTable.Class
	var err error
	d := "今天"
	if strings.Contains(m, "明天") || strings.Contains(m, "明日") {
		d = "明天"
		class, err = schoolTimeTable.GetClass(uid, true)
	} else {
		class, err = schoolTimeTable.GetClass(uid, false)
	}
	if err != nil {
		return "是不是还没有绑定学号呢~", err
	}
	if len(class) == 0 {
		return "好耶，没有课可以使劲摸鱼！", nil
	}
	ret := "你" + d + "的课程有：\n"
	for _, c := range class {
		ret += "\n第 " + strconv.Itoa(c.Start) + " 到 " + strconv.Itoa(c.Finish) + " 节的\n" + c.Name + "\n由 " + c.Teacher + " 老师主讲\n地点：" + c.Position + "\n"
	}
	return ret, nil
}
func Note(uid int64, msg string, match string, mod int) (string, error) {
	switch mod {
	case 0:
		notes, err := note.TakeNotes(uid, msg)
		return notes, err
	case 1:
		notes, err := note.ReadNotes(uid)
		return notes, err
	case 2:
		notes, err := note.EditNotes(uid, msg, match)
		return notes, err
	case 3:
		notes, err := note.DeleteNotes(uid, match)
		return notes, err
	}
	return "", nil
}
func aiPaint(mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	msg := mjson.RawMessage
	nsfw := "0"
	compile := regexp.MustCompile("[^n]tag=[^|｜、]*")
	tag := compile.FindString(msg)
	if tag == "" {
		return "请输入tag哦~", nil
	} else {
		tag = tag[5:]
	}
	v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: config.Settings.BotName.Name + " 马上为您绘制图片哦~"}}
	if mjson.GroupID != 0 {
		v.Param.GroupID = mjson.GroupID
	} else {
		v.Param.UserID = mjson.UserID
	}
	o, _ := json.Marshal(v)
	myUtil.WsLock.Lock()
	err := ws.WriteMessage(returnStruct.MsgType, o)
	myUtil.WsLock.Unlock()
	if err != nil {
		myUtil.ErrLog.Println("在绘制图片前返回预备信息时出现问题,error:", err)
		return config.Settings.BotName.Name + " 突然不会画画了！", err
	}
	compile = regexp.MustCompile("ntag=[^|｜、]*")
	nTag := compile.FindString(msg)
	if nTag != "" {
		nTag = nTag[5:]
	}
	compile = regexp.MustCompile("seed=[^|｜、]*")
	seed := compile.FindString(msg)
	if seed != "" {
		seed = "&seed=" + seed[5:]
	}
	compile = regexp.MustCompile("size=[^|｜、]*")
	size := compile.FindString(msg)
	if size != "" {
		size = size[5:]
	} else {
		size = "Portrait"
	}
	if strings.Contains(msg, "nsfw") {
		nsfw = "1"
	}
	return yusamiPaint.GetPic(tag, nTag, size, nsfw, seed)
	//if len(msg) == 3 {
	//	return yusamiPaint.GetPic(msg[2], "", "Portrait", "1")
	//} else {
	//	return yusamiPaint.GetPic(msg[2], "", msg[3], "1")
	//}
}

func GetTag(mjson returnStruct.Message, ws *websocket.Conn, isGroup bool) (string, error) {
	msg := mjson.RawMessage
	if len(msg) < 12 {
		return "", nil
	}
	if msg[:9] == "炼金术" || msg[len(msg)-9:] == "炼金术" {
		err := myUtil.SendNotice(mjson, ws, "正在努力炼金nya！")
		if err != nil {
			myUtil.ErrLog.Println("在提取tag前返回预备信息时出现问题,error:", err)
			return "咳咳咳！试管突然爆炸了！", err
		}
		res, err := getBase64Pic(msg, mjson)
		if err != nil {
			return res, err
		}
		return yusamiPaint.TagAnalyze(res, mjson.MessageID, isGroup, mjson.UserID)
	} else if msg[:12] == "图片炼成" || msg[len(msg)-12:] == "图片炼成" {
		err := myUtil.SendNotice(mjson, ws, "准备使用图片炼成术nya！")
		if err != nil {
			myUtil.ErrLog.Println("在提取tag前返回预备信息时出现问题,error:", err)
			return "咳咳咳！法阵突然爆炸了！", err
		}
		res, err := getBase64Pic(msg, mjson)
		if err != nil {
			return res, err
		}
		return yusamiPaint.TagToPic(res, mjson.MessageID, isGroup, mjson.UserID)
	} else {
		return "", nil
	}
}
func getBase64Pic(msg string, mjson returnStruct.Message) (string, error) {
	if msg[:9] == "[CQ:reply" {
		replyMsg, err := returnStruct.GetReplyMsg(mjson)
		if err != nil {
			myUtil.ErrLog.Println("在提取tag前获取目标回复信息时出现问题,error:", err)
			return "坩埚裂开了(◎﹏◎)", err
		}
		msg = replyMsg.RetData.Message
	}
	compile := regexp.MustCompile("url=[^\\]]*")
	compiledMsg := compile.FindString(msg)
	if compiledMsg == "" {
		return "啊咧，你好像没有给我原材料呢", errors.New("未找到对应回复消息")
	}
	res := compiledMsg[4:]
	get, err := http.Get(res)
	if err != nil {
		myUtil.ErrLog.Println("获取要提取tag的图片文件时出现问题,error: ", err)
		return "没、没有成功获取到原料nya！", err
	}
	defer get.Body.Close()
	pic, err := ioutil.ReadAll(get.Body)
	base64Pic := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pic)
	return base64Pic, nil
}

//func AddMemberToList(mjson returnStruct.Message) (string, error) {
//	if myUtil.IsInArray(config.Settings.Auth.Admin,mjson.UserID){
//
//	}
//}
