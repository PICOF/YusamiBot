package handler

import (
	"Lealra/HZYS"
	"Lealra/aiTalk"
	"Lealra/config"
	"Lealra/genshin"
	"Lealra/learnResp"
	"Lealra/myUtil"
	"Lealra/news"
	"Lealra/note"
	"Lealra/returnStruct"
	"Lealra/schoolTimeTable"
	"Lealra/setu"
	"github.com/gorilla/websocket"
	"strconv"
	"strings"
)

func privateHandler(mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	ml := strings.Fields(mjson.RawMessage)
	mflen := len(ml)
	var res string
	var err error
	if mflen == 1 {
		if ml[0] == "读日记" {
			res, err = Note(mjson.UserID, "", "", 1)
			if res != "" {
				return res, err
			}
		}
		if strings.Contains(ml[0], "课") {
			res, err := SchoolTimeTable(ml[0], mjson.UserID)
			if res != "" {
				return res, err
			}
		}
		if ml[0] == "查看已绑定学号" {
			res, err := schoolTimeTable.GetSid(mjson.UserID)
			if res != "" {
				return "您已绑定的学号为：" + res, err
			}
		}
		if mjson.RawMessage == "看看新闻" {
			return news.GetNews60s()
		}
		if ok, _ := myUtil.IsInArray(config.Settings.Auth.Admin, mjson.UserID); ok {
			switch mjson.RawMessage {
			case "刷新配置":
				config.GetSetting()
				return "刷新成功~", nil
			case "清理僵尸回复":
				myUtil.SendPrivateMessage(mjson.UserID, "嘟嘟嘟~开始清理啦！")
				num1 := learnResp.ScanZombieResp(true)
				num2 := learnResp.ScanZombieResp(false)
				return "共清理\n" + strconv.Itoa(num1) + "条僵尸化的精确回复\n" + strconv.Itoa(num2) + "条僵尸化的模糊回复", nil
			case "存储回复图片":
				myUtil.SendPrivateMessage(mjson.UserID, "嘟嘟嘟~开始转存啦！")
				num1 := learnResp.StorePicOfResponse(true)
				num2 := learnResp.StorePicOfResponse(false)
				return "共转储\n" + strconv.Itoa(num1) + "条来自精确回复的图片\n" + strconv.Itoa(num2) + "条来自模糊回复的图片", nil
			}
		}
	}
	if mflen >= 2 {
		if mflen == 2 {
			if ml[0] == "日记页面" {
				res, err = note.GetQrcode(mjson.UserID, ml[1])
				if res != "" {
					return res, err
				}
			}
			if ml[0] == "绑定学号" {
				res, err := schoolTimeTable.SetSid(ml[1], mjson.UserID)
				if res != "" {
					return res, err
				}
			}
		}
		res, err := news.BangumiNewsHandler(ml, mjson, ws)
		if res != "" {
			return res, err
		}
		if mflen >= 3 {
			if ml[0] == "修改日记" {
				res, err = Note(mjson.UserID, strings.Join(ml[2:], " "), ml[1], 2)
				if res != "" {
					return res, err
				}
			}
		}
		if ml[0] == "删除日记" {
			res, err = Note(mjson.UserID, "", strings.Join(ml[1:], " "), 3)
			if res != "" {
				return res, err
			}
		}
		if ml[0] == "写日记" {
			res, err = Note(mjson.UserID, mjson.RawMessage[len(ml[0])+1:], "", 0)
			if res != "" {
				return res, err
			}
		}
		if ml[0] == "网易云" {
			res, err = Music(mjson)
			if res != "" {
				return res, err
			}
		}
		if ml[0] == config.Settings.BotName.LowerCasedName+"画画" {
			res, err := aiPaint(mjson, ws)
			if res != "" {
				return res, err
			}
		}
	}
	res, err = setu.GetSetu(mjson)
	if res == "over" {
		return "", err
	} else if res != "" {
		return res, err
	}
	res, err = genshin.GachaHandler(mjson.UserID, mjson.Sender.Nickname, ml, mflen)
	if res != "" {
		return res, err
	}
	res, err = GetTag(mjson, false)
	if res != "" {
		return res, err
	}
	res, err = HZYS.Generate(mjson)
	if res != "" || err != nil {
		return res, err
	}
	res, err = learnResp.Speak(mjson, true)
	if res != "" {
		return res, err
	}
	res, err = learnResp.Speak(mjson, false)
	if res != "" {
		return res, err
	}
	res, err = MakeChoice(mjson)
	if res != "" {
		return res, err
	}
	res = aiTalk.SetCharacterAiToken(ml)
	if res != "" {
		return res, err
	}
	return "", err
}
