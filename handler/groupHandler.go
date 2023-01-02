package handler

import (
	"Lealra/JMComic"
	"Lealra/aiTalk"
	"Lealra/bangumi"
	"Lealra/bilibili"
	"Lealra/config"
	"Lealra/data"
	"Lealra/genshin"
	"Lealra/learnResp"
	"Lealra/myUtil"
	"Lealra/news"
	"Lealra/note"
	"Lealra/returnStruct"
	"Lealra/schoolTimeTable"
	"Lealra/selectedMsg"
	"Lealra/setu"
	"context"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type repeatInfo struct {
	Content string `bson:"content"`
	Count   int    `bson:"count"`
	GroupId int64  `bson:"group_id"`
}

func groupHandler(mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	ml := strings.Fields(mjson.RawMessage)
	mflen := len(ml)
	m := mjson.RawMessage
	var res string
	var err error
	if aiTalk.MsgHandler(ml, mjson, ws) {
		return "", nil
	}
	if mflen == 1 {
		if ml[0] == "读日记" {
			res, err = Note(mjson.UserID, "", "", 1)
			if res != "" {
				return res, err
			}
		}
		if m == "看看新闻" {
			return news.GetNews60s()
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
			if ml[0] == "[CQ:at,qq="+strconv.FormatInt(config.Settings.BotName.Id, 10)+"]" {
				if strings.Contains(ml[1], "课") {
					res, err := SchoolTimeTable(ml[1], mjson.UserID)
					if res != "" {
						return res, err
					}
				}
				if ml[1] == "查看已绑定学号" {
					res, err := schoolTimeTable.GetSid(mjson.UserID)
					if res != "" {
						return "您已绑定的学号为：" + res, err
					}
				}
			}
		}
		if mflen == 2 && ml[0] == "绑定学号" {
			res, err := schoolTimeTable.SetSid(ml[1], mjson.UserID)
			if res != "" {
				return res, err
			}
		}
		if len(m) > 10 && m[1:9] == "CQ:reply" {
			if ok, _ := myUtil.IsInArray(config.Settings.Auth.Admin, mjson.UserID); ok {
				if ok, index := myUtil.IsInArray(ml, "模糊学习"); ok && index < len(ml)-1 {
					res, err = learnResp.LearnResp(mjson, false)
				} else if ok, index = myUtil.IsInArray(ml, "忘记"); ok && index < len(ml)-1 {
					res, err = learnResp.Forget(mjson)
				}
				if res != "" {
					return res, err
				}
			}
			if ok, index := myUtil.IsInArray(ml, "学习"); ok && index < len(ml)-1 {
				res, err = learnResp.LearnResp(mjson, true)
			}
			if res != "" {
				return res, err
			}
			res, err = selectedMsg.SetSelected(mjson)
			if res != "" {
				return res, err
			}
		}
		res, err = bilibili.SubscribeHandler(ws, mjson, ml)
		if res != "" {
			return res, err
		}
		res, err = news.BangumiNewsHandler(ml, mjson, ws)
		if res != "" {
			return res, err
		}
		if ml[0] == config.Settings.BotName.LowerCasedName+"画画" {
			res, err := aiPaint(mjson, ws)
			if res != "" {
				return res, err
			}
		}
		res, err = JMComic.BenziBot(ml, mjson)
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
			res, err = Music(mjson, ws)
			if res != "" {
				return res, err
			}
		}
		if mflen <= 4 && ml[0] == "精华" {
			res, err = selectedMsg.GetSelectedMsg(mjson, m[7:], ws)
			return res, err
		}
	}
	if config.Settings.Bangumi.BangumiSearch {
		res, err = bangumi.WatchBangumi(mjson, ws)
		if res != "" {
			return res, err
		}
	}
	if config.Settings.Setu.SetuGroupSender {
		res, err = setu.GetSetu(ws, mjson)
		if res == "over" {
			return "", err
		}
		if res != "" {
			return res, err
		}
	}
	res, err = genshin.GachaHandler(mjson.UserID, mjson.Sender.Nickname, ml, mflen)
	if res != "" {
		return res, err
	}
	res, err = GetTag(mjson, ws, true)
	if res != "" {
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
	if config.Settings.Func.Repeat {
		res, err = repeat(mjson)
		if res != "" {
			return res, err
		}
	}
	if config.Settings.Func.MakeChoice {
		res, err = MakeChoice(mjson)
		if res != "" {
			return res, err
		}
	}
	return res, err
}
func repeat(mjson returnStruct.Message) (string, error) {
	//db.repeatCount.find({"group_id":876990194})
	repeatNum := 3
	var res repeatInfo
	filter := bson.D{{"group_id", mjson.GroupID}}
	c := data.Db.Collection("repeatCount")
	err := c.FindOne(context.TODO(), filter).Decode(&res)
	if len(res.Content) == 0 {
		_, err := c.InsertOne(context.TODO(), bson.D{{"group_id", mjson.GroupID}, {"content", "114514"}, {"count", 0}})
		if err != nil {
			myUtil.ErrLog.Println("创建新的群复读记录时出现异常:", err)
			return "", err
		}
	}
	if strings.Compare(mjson.RawMessage, res.Content) == 0 {
		//{"group_id":876990194},{$inc:{"count":1}}
		update := bson.D{{"$inc", bson.D{{"count", 1}}}}
		_, err := c.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			myUtil.ErrLog.Println("复读判定出现错误：", err)
			return "", err
		}
		if res.Count+1 == repeatNum {
			rand.Seed(time.Now().Unix())
			myrand := rand.Intn(6)
			switch myrand {
			case 0:
				return res.Content, err
			case 1:
				return "卧槽有复读机", err
			case 2:
				return res.Content, err
			}
		}
	} else {
		update := bson.D{{"$set", bson.D{{"count", 1}, {"content", mjson.RawMessage}}}}
		_, err := c.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			myUtil.ErrLog.Println("复读词重写出现错误：", err)
			return "", err
		}
	}
	return "", err
}
