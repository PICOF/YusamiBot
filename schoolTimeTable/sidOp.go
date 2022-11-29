package schoolTimeTable

import (
	"Lealra/data"
	"Lealra/myUtil"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SidSet struct {
	Sid string `bson:"sid"`
	uid int64  `bson:"uid"`
}

func GetSid(uid int64) (string, error) {
	var res SidSet
	filter := bson.D{{"uid", uid}}
	c := data.Db.Collection("sidTable")
	err := c.FindOne(context.TODO(), filter).Decode(&res)
	return res.Sid, err
}

func SetSid(sid string, uid int64) (string, error) {
	filter := bson.D{{"uid", uid}}
	update := bson.D{{"$set", bson.D{{"sid", sid}}}}
	opt := options.Update().SetUpsert(true)
	c := data.Db.Collection("sidTable")
	_, err := c.UpdateOne(context.TODO(), filter, update, opt)
	if err != nil {
		myUtil.ErrLog.Println("绑定新的学号时出现异常:uid:", uid, "sid:", sid, "err:", err)
		return "没绑上，你真的是 uestc 学生吗", err
	}
	return "绑定成功~", nil
}

func UpdateSid(sid string, uid int64) (string, error) {
	filter := bson.D{{"uid", uid}}
	update := bson.D{{"$set", bson.D{{"sid", sid}}}}
	c := data.Db.Collection("sidTable")
	_, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		myUtil.ErrLog.Println("修改绑定学号时出现异常:uid:", uid, "sid:", sid, "err:", err)
		return "改绑失败，请稍后重试哟~", err
	}
	return "改绑成功~", nil
}
