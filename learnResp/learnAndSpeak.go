package learnResp

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
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

//TODO: 学习和回复都需要修改一下相关逻辑：假如用户没有使用 MongoDB 或者连接未成功建立怎么办？不应该走常规流程，应当提醒用户使用 MongoDB

func LearnResp(mjson returnStruct.Message, isAccurate bool) {
	msg, err := returnStruct.GetReplyMsg(mjson)
	if err != nil {
		myUtil.SendGroupMessage(mjson.GroupID, "emo了,现在人家不想学习！")
		myUtil.ErrLog.Println("获取回复消息内容时出现问题!\nerror:", err, "\nmessage:", mjson.RawMessage)
		return
	}
	if len(msg.RetData.Message) == 0 {
		myUtil.SendGroupMessage(mjson.GroupID, "没有获取到要学习的信息呢……")
		myUtil.ErrLog.Println("获取消息为空，请检查 go-cqhttp 相关消息记录 time:", time.Now().Format("2006-01-02 15:04:05"))
		return
	}
	var col *mongo.Collection
	if isAccurate {
		col = data.Db.Collection("learnedRespAccurate")
	} else {
		col = data.Db.Collection("learnedResp")
	}
	resp := regexp.MustCompile("subType=[0-9]*").ReplaceAllString(msg.RetData.Message, "subType=0")
	filter := bson.D{{"text", string([]rune(mjson.RawMessage)[len([]rune(mjson.RawMessage[:strings.Index(mjson.RawMessage, "学习")]))+3:])}, {"resp", resp}}
	_, err = col.InsertOne(context.TODO(), filter)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			myUtil.SendGroupMessage(mjson.GroupID, "这是多余的学习的说！")
			myUtil.ErrLog.Println("重复学习相同内容，response:" + resp)
			return
		} else {
			myUtil.SendGroupMessage(mjson.GroupID, "学习失败！")
			myUtil.ErrLog.Println("学习新回复"+mjson.RawMessage+"-->"+msg.RetData.Message+"时出错：", err)
			return
		}
	}
	myUtil.SendGroupMessage(mjson.GroupID, "学习成功！")
	if config.Settings.LearnAndResponse.UseBase64 {
		myUtil.LocalPicStorageUpdate(myUtil.StoreCQCode2Base64(resp))
	}
	return
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
			myUtil.ErrLog.Println("查询获取精确学习回复时出现错误！\nerror:", err, "\ntext:", mjson.RawMessage)
			return "记事本被楼下的🐱叼走了！w(ﾟДﾟ)w", err
		}
		aggregate.Next(context.TODO())
		var elem LearnedResp
		err = aggregate.Decode(&elem)
		if err != nil {
			if err.Error() == "EOF" {
				return "", nil
			}
			myUtil.ErrLog.Println("解析精确学习回复时出现错误！\nerror:", err, "\ntext:", mjson.RawMessage)
			return "记事本里的某条记录闪瞎了我的眼睛w(ﾟДﾟ)w", err
		}
		if config.Settings.LearnAndResponse.UseBase64 {
			return myUtil.CQCode2Base64(elem.Resp), nil
		} else {
			return elem.Resp, nil
		}
	} else {
		cur, err := col.Find(context.TODO(), filter)
		if err != nil {
			myUtil.ErrLog.Println("查询获取模糊学习回复时出现错误！\nerror:", err, "\ntext:", mjson.RawMessage)
			return "记事本被楼下的🐱叼走了！w(ﾟДﾟ)w", err
		}
		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem LearnedResp
			err := cur.Decode(&elem)
			if err != nil {
				myUtil.ErrLog.Println("解析模糊学习回复时出现错误！\nerror:", err, "\ntext:", mjson.RawMessage)
				return "记事本里的某条记录闪瞎了我的眼睛w(ﾟДﾟ)w", err
			}
			if strings.Contains(mjson.RawMessage, elem.Text) {
				if config.Settings.LearnAndResponse.UseBase64 {
					return myUtil.CQCode2Base64(elem.Resp), nil
				} else {
					return elem.Resp, nil
				}
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

func StartExtendExpirationTimeLoop() {
	for myUtil.PublicWs == nil {
	}
	for {
		if config.Settings.LearnAndResponse.RenewSwitch {
			myUtil.MsgLog.Println("图片转发备份功能已开启！")
		}
		for config.Settings.LearnAndResponse.RenewSwitch {
			ExtendExpirationTime(true)
			ExtendExpirationTime(false)
			time.Sleep(time.Hour * 48)
		}
		myUtil.MsgLog.Println("图片转发备份功能已关闭！")
		<-config.PicRenewChan
	}
}

func ExtendExpirationTime(isAccurate bool) {
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
		myUtil.SendGroupMessage(config.Settings.LearnAndResponse.GroupToRenew, elem.Resp)
		time.Sleep(time.Second * time.Duration(config.Settings.LearnAndResponse.MsgInterval))
	}
}

// StorePicOfResponse 转换有风险，存储需谨慎。只能说，啊，都怪qq的图床要定期清理
// 暂时不进行相关调用，因为通过定期发送图片进行更新可以做到续期的效果（x
// 错误的，定期发送费时还吵
func StorePicOfResponse(isAccurate bool) int {
	var c *mongo.Collection
	ret := 0
	if isAccurate {
		c = data.Db.Collection("learnedRespAccurate")
	} else {
		c = data.Db.Collection("learnedResp")
	}
	cur, err := c.Find(context.TODO(), bson.D{})
	if err != nil {
		myUtil.ErrLog.Println("将回复全部转换为本地存储时出错:\n", err, "\nisAccurate", isAccurate)
		return -1
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var elem LearnedResp
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err)
			continue
		}
		ret += myUtil.LocalPicStorageUpdate(myUtil.StoreCQCode2Base64(elem.Resp))
	}
	return ret
}
