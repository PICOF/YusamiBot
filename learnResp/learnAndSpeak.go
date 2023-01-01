package learnResp

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
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
	resp := regexp.MustCompile("subType=[0-9]*").ReplaceAllString(msg.RetData.Message, "subType=0")
	if config.Settings.LearnAndResponse.UseBase64 {
		//CQCode2Base64(resp)
	}
	filter := bson.D{{"text", string([]rune(mjson.RawMessage)[len([]rune(mjson.RawMessage[:strings.Index(mjson.RawMessage, "学习")]))+3:])}, {"resp", resp}}
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

// IsZombieResp 纯纯拿来判断僵尸回复（指那些qq图床已经失效的回复）的函数
func IsZombieResp(response string) bool {
	var get *http.Response
	var err error
	compile := regexp.MustCompile("CQ:image[^\\]]+")
	index := compile.FindAllStringIndex(response, -1)
	for _, v := range index {
		pic := response[v[0]:v[1]]
		url := pic[strings.Index(pic, "url=")+4:]
		get, err = http.Get(url)
		if err != nil {
			myUtil.ErrLog.Println("检查僵尸回复时出现网络错误：", err)
			return false
		}
		if get.StatusCode == 404 {
			return true
		}
	}
	if get != nil {
		get.Body.Close()
	}
	return false
}
func ScanZombieResp(isAccurate bool) int {
	var cName string
	if isAccurate {
		cName = "learnedRespAccurate"
	} else {
		cName = "learnedResp"
	}
	ret := 0
	c := data.Db.Collection(cName)
	cur, err := c.Find(context.TODO(), bson.D{})
	if err != nil {
		myUtil.ErrLog.Println("Error when deleting zombie response:", err)
		return -1
	}
	defer cur.Close(context.TODO())
	var wg sync.WaitGroup
	for cur.Next(context.TODO()) {
		var elem LearnedResp
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err)
			continue
		}
		wg.Add(1)
		go func(response string, text string) {
			defer wg.Done()
			if IsZombieResp(response) {
				one, err := data.Db.Collection(cName).DeleteOne(context.TODO(), bson.D{{"text", text}, {"resp", response}})
				if err != nil {
					myUtil.ErrLog.Println("Error when deleting zombie response:", err, "\nresponse:", response, "\ntext:", text)
					return
				}
				if one.DeletedCount != 0 {
					ret++
				}
			}
		}(elem.Resp, elem.Text)
	}
	wg.Wait()
	return ret
}

//func CQCode2Base64(response string) (string, string) {
//	compile := regexp.MustCompile("CQ:image[^\\]]+")
//	index := compile.FindAllStringIndex(response, -1)
//	for _, v := range index {
//		pic := response[v[0]:v[1]]
//		url := pic[strings.Index(pic, "url=")+4:]
//		if url == "" {
//			continue
//		}
//		str := myUtil.GetBase64CQCode(url)
//		if str == "" {
//			continue
//		} else {
//			response = response[:v[0]-1] + str + response[v[1]+1:]
//		}
//	}
//	return response
//}

func StartExtendExpirationTimeLoop() {
	for myUtil.PublicWs == nil {
	}
	for {
		ExtendExpirationTime(myUtil.PublicWs, true)
		ExtendExpirationTime(myUtil.PublicWs, false)
		time.Sleep(time.Hour * 48)
	}
}

func ExtendExpirationTime(ws *websocket.Conn, isAccurate bool) {
	var cName string
	if isAccurate {
		cName = "learnedRespAccurate"
	} else {
		cName = "learnedResp"
	}
	c := data.Db.Collection(cName)
	cur, err := c.Find(context.TODO(), bson.D{})
	if err != nil {
		myUtil.ErrLog.Println("Error when extend expiration time of response:", err)
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var elem LearnedResp
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err, "\ntext:", elem.Text, " response:", elem.Resp, " isAccurate:", isAccurate)
		}
		myUtil.SendGroupMessage(ws, config.Settings.LearnAndResponse.GroupToRenew, elem.Resp)
		time.Sleep(time.Second * time.Duration(config.Settings.LearnAndResponse.MsgInterval))
	}
}

// AllCQCode2Base64 转换有风险，存储需谨慎。只能说，啊，都怪qq的图床要定期清理
// 暂时不进行相关调用，因为通过定期发送图片进行更新可以做到续期的效果（x
// 错误的，定期发送费时还吵
//func AllCQCode2Base64() int {
//	ret := 0
//	c := data.Db.Collection("base64Set")
//	cur, err := c.Find(context.TODO(), bson.D{})
//	if err != nil {
//		myUtil.ErrLog.Println("Error when turn CQ code into base64:", err)
//		return -1
//	}
//	defer cur.Close(context.TODO())
//	for cur.Next(context.TODO()) {
//		var elem LearnedResp
//		err := cur.Decode(&elem)
//		if err != nil {
//			myUtil.ErrLog.Println(err)
//			continue
//		}
//		_, err = c.UpdateOne(context.TODO(), bson.D{{"resp", elem.Resp}, {"text", elem.Text}}, bson.D{{"resp", CQCode2Base64(elem.Resp)}})
//		if err != nil {
//			return 0
//		}
//		ret++
//	}
//	return ret
//}
