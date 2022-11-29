package learnResp

import (
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"regexp"
	"strings"
)

type LearnedResp struct {
	Resp string `bson:"resp"`
	Text string `bson:"text"`
}

func LearnResp(mjson returnStruct.Message, isAccurate bool) (string, error) {
	msg, err := returnStruct.GetReplyMsg(mjson)
	if err != nil {
		return "emo了,现在人家不想学习！", err
	}
	if len(msg.RetData.Message) == 0 {
		return "没有获取到要学习的信息呢……", nil
	}
	var col *mongo.Collection
	if isAccurate {
		col = data.Db.Collection("learnedRespAccurate")
	} else {
		col = data.Db.Collection("learnedResp")
	}
	compile := regexp.MustCompile("subType=[0-9]*")
	filter := bson.D{{"text", string([]rune(mjson.RawMessage)[len([]rune(mjson.RawMessage[:strings.Index(mjson.RawMessage, "学习")]))+3:])}, {"resp", compile.ReplaceAllString(msg.RetData.Message, "subType=0")}}
	_, err = col.InsertOne(context.TODO(), filter)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return "这是多余的学习的说！", err
		} else {
			myUtil.ErrLog.Println("学习新回复"+mjson.RawMessage+"-->"+msg.RawMessage+"时出错：", err)
		}
	}
	return "学习成功！", err
}
func Speak(mjson returnStruct.Message, isAccurate bool) (string, error) {
	var filter bson.D
	var col *mongo.Collection
	if isAccurate {
		col = data.Db.Collection("learnedRespAccurate")
	} else {
		filter = bson.D{{}}
		col = data.Db.Collection("learnedResp")
	}
	if isAccurate {
		aggregate, err := col.Aggregate(context.TODO(), mongo.Pipeline{bson.D{{"$match", bson.D{{"text", mjson.RawMessage}}}}, bson.D{{"$sample", bson.D{{"size", 1}}}}})
		if err != nil {
			myUtil.ErrLog.Println(err)
			return "记事本被楼下的🐱叼走了！w(ﾟДﾟ)w", err
		}
		aggregate.Next(context.TODO())
		var elem LearnedResp
		err = aggregate.Decode(&elem)
		if err != nil {
			if err.Error() == "EOF" {
				return "", nil
			}
			myUtil.ErrLog.Println(err)
			return "记事本里的某条记录闪瞎了我的眼睛w(ﾟДﾟ)w", err
		}
		return elem.Resp, nil
	} else {
		cur, err := col.Find(context.TODO(), filter)
		if err != nil {
			myUtil.ErrLog.Println("Error searching for learnedResp:", err)
			return "", err
		}
		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem LearnedResp
			err := cur.Decode(&elem)
			if err != nil {
				myUtil.ErrLog.Println(err)
				return "记事本里的某条记录闪瞎了我的眼睛w(ﾟДﾟ)w", err
			}
			if strings.Contains(mjson.RawMessage, elem.Text) {
				return elem.Resp, nil
			}
		}
	}
	return "", nil
}
