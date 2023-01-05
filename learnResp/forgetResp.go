package learnResp

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

func Forget(mjson returnStruct.Message) (string, error) {
	m, err := returnStruct.GetReplyMsg(mjson)
	if err != nil {
		return "忘不了♫~忘不了♪♩~忘不了你的泪♬~", err
	}
	text := string([]rune(mjson.RawMessage)[len([]rune(mjson.RawMessage[:strings.Index(mjson.RawMessage, "忘记")]))+3:])
	resp := m.RetData.Message
	col := data.Db.Collection("learnedRespAccurate")
	del, err := col.DeleteOne(context.TODO(), bson.D{{"text", text}, {"resp", resp}})
	if err != nil {
		return "啊咧，奇怪的感觉……", err
	}
	if del.DeletedCount != 0 {
		if config.Settings.LearnAndResponse.UseBase64 {
			myUtil.ForgetLocalPicInCQCode(resp)
		}
		return string([]rune(text)[:1]) + "……什么来着？", nil
	}
	col = data.Db.Collection("learnedResp")
	del, err = col.DeleteOne(context.TODO(), bson.D{{"resp", resp}, {"text", bson.D{{"$regex", text}}}})
	if err != nil {
		return "啊咧，奇怪的感觉……", err
	}
	if del.DeletedCount != 0 {
		if config.Settings.LearnAndResponse.UseBase64 {
			myUtil.ForgetLocalPicInCQCode(resp)
		}
		return string([]rune(text)[:1]) + "……什么来着？", nil
	}
	if del.DeletedCount == 0 {
		return "NullPointerException!\t\n压根儿就没有学过这个东东~", nil
	}
	return "", nil
}
