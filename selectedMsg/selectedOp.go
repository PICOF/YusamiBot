package selectedMsg

import (
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
	"time"
)

type SelectedMsg struct {
	Content  string `bson:"content"`
	Nickname string `bson:"nickname"`
	Time     int64  `bson:"time"`
	UserID   int64  `bson:"user_id"`
	ID       int32  `bson:"message_id"`
}

func GetQueryString(query []string) ([4]string, error) {
	var args [4]string
	var err error
	var index int
	for _, v := range query {
		c := strings.Index(v, "：")
		if c == -1 {
			index = strings.Index(v, ":")
			if index == -1 {
				return args, errors.New("invalid query string: SetSelectedMsgQuery")
			}
		} else {
			index = c
		}
		switch v[:index] {
		case "name":
			args[0] = string([]rune(v)[index+1:])
		case "uid":
			args[1] = string([]rune(v)[index+1:])
		case "word":
			args[2] = string([]rune(v)[index+1:])
		case "date":
			args[3] = string([]rune(v)[index+1:])
		default:
			err = errors.New("invalid query string: SetSelectedMsgQuery")
		}
	}
	return args, err
}

func GetSelectedMsg(mjson returnStruct.Message, name string, uid string, word string, date string, ws *websocket.Conn) (string, error) {
	c := data.Db.Collection("selectedMsg_" + strconv.FormatInt(mjson.GroupID, 10))
	queryFlag := 0
	var tfilter, nfilter, ufilter, wfilter bson.E
	if name != "" {
		nfilter = bson.E{Key: "nickname", Value: bson.D{{"$eq", name}}}
		queryFlag++
	} else {
		nfilter = bson.E{}
	}
	if uid != "" {
		u, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			return "", err
		}
		ufilter = bson.E{Key: "user_id", Value: bson.D{{"$eq", u}}}
		queryFlag++
	} else {
		ufilter = bson.E{}
	}
	if word != "" {
		wfilter = bson.E{Key: "content", Value: bson.D{{"$regex", word}}}
		queryFlag++
	} else {
		wfilter = bson.E{}
	}
	if date != "" {
		t, err := time.ParseInLocation("2006-1-2", date, time.Local)
		if err != nil {
			myUtil.ErrLog.Println("Error parsing time: ", err)
			return "不要乱填时间哦~", err
		}
		tfilter = bson.E{Key: "time", Value: bson.D{{"$gte", t.Unix()}, {"$lt", t.Unix() + 86400}}}
		queryFlag++
	} else {
		tfilter = bson.E{}
	}
	if queryFlag == 0 {
		return "你什么条件都不给我我怎么知道你要查什么呀~(￣﹃￣)", nil
	}
	filter := bson.D{tfilter, nfilter, ufilter, wfilter}
	opts := options.Find().SetSort(bson.D{{"time", 1}})
	cur, err := c.Find(context.TODO(), filter, opts)
	if err != nil {
		return "欸嘿~我不小心把记录本搞丢了~(。・ω・。)", err
	}
	var results []returnStruct.Node
	for cur.Next(context.TODO()) {
		var elem SelectedMsg
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err)
			return "记事本里的某条记录闪瞎了我的眼睛w(ﾟДﾟ)w", err
		}
		if strings.Contains(elem.Content, "forward") {
			var n []returnStruct.Node
			//id=Woqn0pPECCdfn/Cx/JKhUZtxWIjLOVmUSfN9qqEYVimRAYcDUeYRi7jOcc3Bpt1g
			index := strings.Index(elem.Content, "=")
			nodes, err := returnStruct.GetForwardMsgNodes(n, elem.Content[index+1:index+65])
			if err != nil {
				return "要记录什么来着？", err
			}
			results = append(results, returnStruct.Node{Type: "node", Data: returnStruct.NData{Name: elem.Nickname, Uid: elem.UserID, Content: nodes, Time: elem.Time}})
		} else {
			results = append(results, returnStruct.Node{Type: "node", Data: returnStruct.NData{Name: elem.Nickname, Uid: elem.UserID, Content: elem.Content, Time: elem.Time}})
		}
	}
	if results == nil {
		return "未找到相关精华消息~", err
	}
	ret := returnStruct.SendMsg{Action: "send_group_forward_msg", Param: returnStruct.Params{GroupID: mjson.GroupID, Messages: results}}
	res, err := json.Marshal(ret)
	//println(string(res))
	if err != nil {
		myUtil.ErrLog.Println("Error marshalling:", err)
	}
	err = ws.WriteMessage(returnStruct.MsgType, res)
	return "", err
}
func SetSelected(mjson returnStruct.Message) (string, error) {
	m := []rune(mjson.RawMessage)
	k := string(m[len(m)-2:])
	if k == "设精" || k == "加精" {
		msg, err := returnStruct.GetReplyMsg(mjson)
		if err != nil {
			myUtil.ErrLog.Println("添加精华消息时获取对象内容时出错：", err.Error())
			return "好像受到了奇怪的电波干扰~", err
		}
		if msg.RetCode != 0 {
			return "好像出现问题啦！" + msg.Wording, err
		}
		if msg.RetData.Message == "" {
			return "暂不支持这种消息的精华设置哦~", err
		}
		res, err := setSelectedMsg(mjson, msg, m)
		return res, err
	}
	return "", nil
}
func setSelectedMsg(mjson returnStruct.Message, msg returnStruct.Message, m []rune) (string, error) {
	res, err := StoreMsg(msg)
	if res != "" {
		return res, err
	}
	if err != nil {
		return "出现未知错误喵~请稍后重试", err
	}
	return strings.Fields(string(m[:len(m)-2]))[0] + "[CQ:at,qq=" + strconv.FormatInt(mjson.UserID, 10) + "]" + "设置好啦~", nil
}
func StoreMsg(message returnStruct.Message) (res string, err error) {
	col := data.Db.Collection("selectedMsg_" + strconv.FormatInt(message.RetData.GroupID, 10))
	_, err = col.Indexes().CreateOne(context.TODO(), mongo.IndexModel{Keys: bson.D{{"time", 1}}, Options: options.Index().SetUnique(true)})
	if err != nil {
		myUtil.ErrLog.Println("设置群组 ", message.RetData.GroupID, " 的精华消息索引时出错：", err)
	}
	filter := bson.D{{"content", message.RetData.Message}, {"nickname", message.RetData.Sender.Nickname}, {"time", message.RetData.Time}, {"user_id", message.RetData.Sender.UserID}, {"message_id", message.RetData.MessageID}}
	_, err = col.InsertOne(context.TODO(), filter)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			return "请不要重复设置精华消息的说！", err
		} else {
			myUtil.ErrLog.Println("插入群组 ", message.RetData.GroupID, " 的精华消息 ID", message.MessageID, "时出错：", err)
		}
	}
	return "", err
}
