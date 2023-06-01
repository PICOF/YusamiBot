package funnyReply

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/myUtil"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func CleanOldRecord() {
	c := data.Db.Collection("msgAnalysisResults")
	for {
		filter := bson.D{{"tag", ""}, {"last_time", bson.D{{"$lt", time.Now().Unix() - int64((time.Hour * time.Duration(config.Settings.Daoli.Expiration)).Seconds())}}}}
		res, err := c.DeleteMany(context.TODO(), filter)
		if err != nil {
			myUtil.ErrLog.Println("消息分析记录清理失败！time: ", time.Now().String())
		} else if res.DeletedCount > 0 {
			myUtil.MsgLog.Println("消息分析记录已清理：共删除 ", res.DeletedCount, " 条记录")
		}
		time.Sleep(time.Hour * time.Duration(config.Settings.Daoli.Expiration))
	}
}
