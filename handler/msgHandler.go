package handler

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
)

const groupFilterIndex = 0
const privateFilterIndex = 1

func MsgHandler(ws *websocket.Conn, msg []byte, filterMode []int) ([]byte, error) {
	var mjson returnStruct.Message
	err := json.Unmarshal(msg, &mjson)
	if err != nil {
		return nil, err
	}
	if mjson.PostType == "meta_event" {
		switch mjson.MetaEventType {
		case "heartbeat":
			//println("heartbeat")
			return nil, nil
		case "lifecycle":
			myUtil.MsgLog.Println("lifecycle")
			return nil, nil
		}
	}
	if mjson.PostType == "message" {
		if mjson.RawMessage == "" {
			return nil, errors.New("消息中出现零长段，user_id=" + strconv.FormatInt(mjson.UserID, 10) + "，group_id=" + strconv.FormatInt(mjson.GroupID, 10))
		}
		ret := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{}}
		if mjson.MessageType == "group" {
			if !hasPermission(filterMode[groupFilterIndex], mjson.GroupID, true) {
				myUtil.MsgLog.Println("Group " + strconv.FormatInt(mjson.GroupID, 10) + " do not have permission")
				return nil, err
			}
			myUtil.MsgLog.Print(string(msg))
			ret.Param.GroupID = mjson.GroupID
			res, err := groupHandler(mjson, ws)
			return resHandler(res, err, ret)
		} else if mjson.MessageType == "private" {
			if !hasPermission(filterMode[privateFilterIndex], mjson.UserID, false) {
				myUtil.MsgLog.Println("User " + strconv.FormatInt(mjson.UserID, 10) + " do not have permission")
				return nil, err
			}
			myUtil.MsgLog.Print(string(msg))
			ret.Param.UserID = mjson.UserID
			res, err := privateHandler(mjson, ws)
			return resHandler(res, err, ret)
		}
	}
	return nil, err
}
func resHandler(res string, err error, ret returnStruct.SendMsg) ([]byte, error) {
	if res != "" {
		ret.Param.Message = res
		k, _ := json.Marshal(ret)
		return k, err
	}
	if err != nil {
		ret.Param.Message = "哎呀，刚刚 " + config.Settings.BotName.Name + " 没听清，再说一遍好吗？"
		errBack, _ := json.Marshal(ret)
		return errBack, err
	}
	return nil, err
}
func hasPermission(filterMode int, id int64, isGroup bool) bool {
	var column string
	if isGroup {
		column = "group_id"
	} else {
		column = "user_id"
	}
	filter := bson.D{{"type", filterMode}, {column, bson.D{{"$in", bson.A{id}}}}}
	switch filterMode {
	case 1:
		res, _ := data.Db.Collection("PermissionList").FindOne(context.TODO(), filter).DecodeBytes()
		if res != nil {
			return false
		} else {
			return true
		}
	case 2:
		res, _ := data.Db.Collection("PermissionList").FindOne(context.TODO(), filter).DecodeBytes()
		if res != nil {
			return true
		} else {
			return false
		}
	}
	return true
}
