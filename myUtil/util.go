package myUtil

import (
	"Lealra/config"
	"Lealra/data"
	"Lealra/returnStruct"
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
)

var WsLock sync.Mutex
var proxyIndex atomic.Uint32
var PublicWs *websocket.Conn

type LocalPic struct {
	Md5           string `bson:"md5"`
	Base64        string `bson:"base64"`
	CompressedPic []byte `bson:"compressed_pic"`
}

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
func MakeForwardMsgNode(msg interface{}) []returnStruct.Node {
	var resSet []returnStruct.Node
	m := reflect.ValueOf(msg)
	for i := 0; i < m.Len(); i++ {
		a := returnStruct.Node{}
		a.Type = "node"
		a.Data.Uid = config.Settings.BotName.Id
		a.Data.Name = config.Settings.BotName.FullName
		a.Data.Content = m.Index(i).Interface()
		resSet = append(resSet, a)
	}
	return resSet
}
func SendForwardMsg(msg interface{}, mjson returnStruct.Message) error {
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
	err = PublicWs.WriteMessage(returnStruct.MsgType, marshal)
	WsLock.Unlock()
	if err != nil {
		return err
	}
	return nil
}
func SendNotice(mjson returnStruct.Message, msg string) error {
	v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: msg}}
	if mjson.GroupID != 0 {
		v.Param.GroupID = mjson.GroupID
	} else {
		v.Param.UserID = mjson.UserID
	}
	o, _ := json.Marshal(v)
	WsLock.Lock()
	err := PublicWs.WriteMessage(returnStruct.MsgType, o)
	WsLock.Unlock()
	return err
}
func SendPrivateMessage(uid int64, msg string) {
	v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: msg, UserID: uid}}
	o, _ := json.Marshal(v)
	WsLock.Lock()
	err := PublicWs.WriteMessage(returnStruct.MsgType, o)
	WsLock.Unlock()
	if err != nil {
		ErrLog.Println("在发送私聊消息时出现错误！\nerror:", err, "\nmessage:", msg)
	}
}
func SendGroupMessage(groupId int64, msg string) {
	v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: msg, GroupID: groupId}}
	o, _ := json.Marshal(v)
	WsLock.Lock()
	err := PublicWs.WriteMessage(returnStruct.MsgType, o)
	WsLock.Unlock()
	if err != nil {
		ErrLog.Println("在发送群聊消息时出现错误！\nerror:", err, "\nmessage:", msg)
	}
}
func GetBase64CQCode(url string) string {
	str := Pic2Base64ByUrl(url)
	if str == "" {
		return str
	} else {
		return "[CQ:image,file=base64://" + str + "]"
	}
}
func Pic2Base64ByUrl(url string) string {
	get, err := http.Get(url)
	if err != nil {
		ErrLog.Println("url转base64时出现错误，error:", err, "url:", url)
		return ""
	}
	defer get.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(get.Body)
	return base64.StdEncoding.EncodeToString(body)
}

func StoreCQCode2Base64(response string) map[string]string {
	ret := map[string]string{}
	compile := regexp.MustCompile("file=[^\\]]+")
	index := compile.FindAllStringIndex(response, -1)
	for _, v := range index {
		pic := response[v[0]:v[1]]
		link := pic[strings.Index(pic, "url=")+4:]
		if link == "" {
			ErrLog.Println("获取 CQ 码图片 url 时出现获取错误!\n原字段：", response)
			continue
		}
		md5 := pic[5:strings.Index(pic, ".")]
		if md5 == "" {
			ErrLog.Println("获取 CQ 码图片 md5 时出现获取错误!\n原字段：", response)
			continue
		} else if LocalPicStorageCheck(md5) {
			continue
		}
		str := Pic2Base64ByUrl(link)
		if str == "" {
			ErrLog.Println("图片消息本地备份时出现资源获取错误!\n请求地址：", link, "\n原字段：", response)
			continue
		} else if ret[md5] == "" {
			ret[md5] = str
		}
	}
	return ret
}
func CQCode2Base64(response string) string {
	compile := regexp.MustCompile("file=[^\\]]+")
	return string(compile.ReplaceAllFunc([]byte(response), func(match []byte) []byte {
		pic := string(match)
		md5 := pic[5:strings.Index(pic, ".")]
		res := GetLocalPicStorage(md5)
		if res != "" {
			return []byte("file=base64://" + res)
		}
		return match
	}))
}

func IsNumber(s string) bool {
	for _, v := range s {
		if !unicode.IsNumber(v) {
			return false
		}
	}
	return true
}

func LocalPicStorageCheck(md5 string) bool {
	var res LocalPic
	c := data.Db.Collection("localPicStorage")
	filter := bson.D{{"md5", md5}}
	err := c.FindOne(context.TODO(), filter).Decode(&res)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false
		} else {
			ErrLog.Println("检查图片是否存储时出现错误！error:", err)
			return false
		}
	} else {
		return true
	}
}

func GetMd5OfPic(response string) []string {
	var ret []string
	compile := regexp.MustCompile("file=[^\\]\\.]+\\.")
	index := compile.FindAllStringIndex(response, -1)
	for _, v := range index {
		ret = append(ret, response[v[0]+5:v[1]-1])
	}
	return ret
}

func DeleteLocalPicStorage(md5 string) bool {
	res, err := data.Db.Collection("localPicStorage").DeleteOne(context.TODO(), bson.M{"md5": md5})
	if err != nil {
		ErrLog.Println("本地图片删除时出现错误！error:", err, "\nmd5:", md5)
		return false
	}
	if res.DeletedCount != 0 {
		return true
	} else {
		ErrLog.Println("本地图片删除时未找到对应记录！md5:", md5)
		return false
	}
}

func ForgetLocalPicInCQCode(resp string) {
	count := 0
	for _, v := range GetMd5OfPic(resp) {
		if DeleteLocalPicStorage(v) {
			count++
		}
	}
	MsgLog.Println("本地图片删除", count, "条记录")
}

func GetLocalPicStorage(md5 string) string {
	var ret LocalPic
	c := data.Db.Collection("localPicStorage")
	filter := bson.D{{"md5", md5}}
	res := c.FindOne(context.TODO(), filter)
	if res == nil {
		return ""
	} else {
		err := res.Decode(&ret)
		if err != nil {
			ErrLog.Println("获取本地图片时出现问题！\nerror:", err, "\nmd5:", md5)
			return ""
		}
		if !config.Settings.LearnAndResponse.Compress || ret.CompressedPic == nil {
			return ret.Base64
		} else {
			return getUnCompressed(ret.CompressedPic)
		}
	}
}

func LocalPicStorageUpdate(pic map[string]string) int {
	count := 0
	var err error
	var update bson.D
	c := data.Db.Collection("localPicStorage")
	opt := options.Update().SetUpsert(true)
	for k, v := range pic {
		filter := bson.D{{"md5", k}}
		if config.Settings.LearnAndResponse.Compress {
			update = bson.D{{"$set", bson.D{{"compressed_pic", getCompressed(v)}}}}
		} else {
			update = bson.D{{"$set", bson.D{{"base64", v}}}}
		}
		_, err = c.UpdateOne(context.TODO(), filter, update, opt)
		if err != nil {
			ErrLog.Println("将图片存储至本地时出现异常，error:", err)
		}
		count++
	}
	MsgLog.Println("成功转储", count, "张图片")
	return count
}

func getCompressed(input string) (data []byte) {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write([]byte(input))
	w.Close()
	return in.Bytes()
}
func getUnCompressed(input []byte) (data string) {
	b := bytes.NewReader(input)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	io.Copy(&out, r)
	return out.String()
}

func GetProxyClient() *http.Client {
	index := proxyIndex.Add(1) % uint32(len(config.Settings.Proxy.HttpsProxy))
	uri, err := url.Parse(config.Settings.Proxy.HttpsProxy[index])
	if err != nil {
		ErrLog.Println("parse url error: ", err)
		return &http.Client{}
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(uri),
		},
	}
	MsgLog.Println("使用代理：" + uri.String())
	return client
}
