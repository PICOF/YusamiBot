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
	filter := bson.D{{"tag", ""}, {"last_time", bson.D{{"$lt", time.Now().Unix() - int64((time.Hour * time.Duration(config.Settings.Daoli.Expiration)).Seconds())}}}}
	for {
		_, err := c.DeleteMany(context.TODO(), filter)
		if err != nil {
			myUtil.ErrLog.Println("消息分析记录清理失败！time: ", time.Now().String())
		}
		time.Sleep(time.Hour * time.Duration(config.Settings.Daoli.Expiration))
	}
}
