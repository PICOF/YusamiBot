package myUtil

import (
	"Lealra/config"
	"Lealra/returnStruct"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
	"reflect"
	"sync"
)

var WsLock sync.Mutex

func IsInArray(array interface{}, target interface{}) (bool, int) {
	a := reflect.ValueOf(array)
	for i := 0; i < a.Len(); i++ {
		if reflect.DeepEqual(target, a.Index(i).Interface()) {
			return true, i
		}
	}
	return false, -1
}
func SeseQrcode(dir string, filename string) (string, error) {
	encode, err := qrcode.Encode("http://"+config.Settings.Server.Hostname+":"+config.Settings.Server.Port+"/"+dir+"/"+filename, qrcode.Medium, 256)
	if err != nil {
		ErrLog.Println("生成🐍图二维码时出现异常！err:", err)
		return "", err
	}
	ret := base64.StdEncoding.EncodeToString(encode)
	return ret, nil
}
func SendForwardMsg(msg interface{}, mjson returnStruct.Message, ws *websocket.Conn) error {
	var fmsg returnStruct.SendMsg
	if mjson.GroupID == 0 {
		fmsg.Param.UserID = mjson.UserID
	} else {
		fmsg.Param.GroupID = mjson.GroupID
	}
	fmsg.Action = "send_group_forward_msg"
	var a returnStruct.Node
	m := reflect.ValueOf(msg)
	for i := 0; i < m.Len(); i++ {
		a = returnStruct.Node{}
		a.Type = "node"
		a.Data.Uid = config.Settings.BotName.Id
		a.Data.Name = config.Settings.BotName.FullName
		a.Data.Content = m.Index(i).Interface()
		fmsg.Param.Messages = append(fmsg.Param.Messages, a)
	}
	marshal, err := json.Marshal(fmsg)
	if err != nil {
		return err
	}
	WsLock.Lock()
	err = ws.WriteMessage(returnStruct.MsgType, marshal)
	WsLock.Unlock()
	if err != nil {
		return err
	}
	return nil
}
func SendNotice(mjson returnStruct.Message, ws *websocket.Conn, msg string) error {
	v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: msg}}
	if mjson.GroupID != 0 {
		v.Param.GroupID = mjson.GroupID
	} else {
		v.Param.UserID = mjson.UserID
	}
	o, _ := json.Marshal(v)
	WsLock.Lock()
	err := ws.WriteMessage(returnStruct.MsgType, o)
	WsLock.Unlock()
	return err
}
