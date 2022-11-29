package note

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/myUtil"
	"context"
	"encoding/base64"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"time"
)

type notes struct {
	Note []string `bson:"note"`
}

func ReadNotes(uid int64) (string, error) {
	var res notes
	filter := bson.D{{"uid", uid}, {"date", time.Now().Format("2006-01-02")}}
	c := data.Db.Collection("diary")
	err := c.FindOne(context.TODO(), filter).Decode(&res)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return "你今天还没有写任何日记哦~", nil
		} else {
			myUtil.ErrLog.Println("读取日记时出现异常:uid:", uid, "err:", err)
			return "我的日记本找不到了QAQ", err
		}
	}
	ret := config.Settings.BotName.Name + " 帮你读读今天的日记~"
	for _, v := range res.Note {
		ret += "\n"
		ret += v
	}
	return ret, err
}
func GetNotes(uid int64, date string) (notes, error) {
	var res notes
	filter := bson.D{{"uid", uid}, {"date", date}}
	c := data.Db.Collection("diary")
	err := c.FindOne(context.TODO(), filter).Decode(&res)
	return res, err
}
func GetQrcode(uid int64, date string) (string, error) {
	encode, err := qrcode.Encode("http://"+config.Settings.Server.Hostname+":"+config.Settings.Server.Port+"/diary/"+strconv.FormatInt(uid, 10)+"/"+date, qrcode.Medium, 256)
	if err != nil {
		myUtil.ErrLog.Println("生成日记网站二维码时出现异常:uid:", uid, "err:", err)
		return "", err
	}
	ret := base64.StdEncoding.EncodeToString(encode)
	return "[CQ:image,file=base64://" + ret + "]", err
}
