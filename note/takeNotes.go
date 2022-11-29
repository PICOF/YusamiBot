package note

import (
	"Lealra/data"
	"Lealra/myUtil"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func TakeNotes(uid int64, msg string) (string, error) {
	//db.diary.updateOne({"user_id":1730483316},{$addToSet:{"note":"2022/11/12 \nwocao"}},{upsert:true})
	filter := bson.D{{"uid", uid}, {"date", time.Now().Format("2006-01-02")}}
	update := bson.D{{"$push", bson.D{{"note", time.Now().Format("2006-01-02(PM-03:04:05)") + "\n" + msg}}}}
	opt := options.Update().SetUpsert(true)
	c := data.Db.Collection("diary")
	_, err := c.UpdateOne(context.TODO(), filter, update, opt)
	if err != nil {
		myUtil.ErrLog.Println("记录日记时出现异常:uid:", uid, "msg:", msg, "err:", err)
		return "我的日记本找不到了QAQ", err
	}
	return "记录成功~", nil
}
func EditNotes(uid int64, msg string, match string) (string, error) {
	var res notes
	filter := bson.D{{"uid", uid}, {"date", time.Now().Format("2006-01-02")}, {"note", bson.D{{"$regex", match}}}}
	c := data.Db.Collection("diary")
	err := c.FindOne(context.TODO(), filter, options.FindOne().SetProjection(bson.D{{"note.$", 1}})).Decode(&res)
	if err != nil {
		myUtil.ErrLog.Println("修改日记时查询出现异常:uid:", uid, "msg:", msg, "err:", err)
		return "啊咧咧，你真的记过这样的东西吗~", err
	}
	update := bson.D{{"$set", bson.D{{"note.$", res.Note[0][:24] + msg}}}}
	_, err = c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		myUtil.ErrLog.Println("修改日记时覆写出现异常:uid:", uid, "msg:", msg, "match:", match, "err:", err)
		return "我的日记本找不到了QAQ", err
	}
	return "修改成功~", nil
}
func DeleteNotes(uid int64, match string) (string, error) {
	filter := bson.D{{"uid", uid}, {"date", time.Now().Format("2006-01-02")}}
	update := bson.D{{"$pull", bson.D{{"note", bson.D{{"$regex", match}}}}}}
	c := data.Db.Collection("diary")
	res, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		myUtil.ErrLog.Println("删除日记时出现异常:uid:", uid, "match:", match, "err:", err)
		return "我的日记本找不到了QAQ", err
	}
	if res.ModifiedCount == 0 {
		return "貌似没有过相关记录呢~", err
	}
	return "删除成功~", nil
}
