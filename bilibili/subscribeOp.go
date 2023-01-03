package bilibili

import (
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var groupSwitch = make(map[int64]bool)
var groupLock sync.Mutex

type subscribeList struct {
	Uid       []int64 `bson:"uid"`
	Gid       int64   `bson:"gid"`
	Subscribe string  `bson:"subscribe"`
	Name      string  `bson:"name"`
}

func SubscribeHandler(ws *websocket.Conn, mjson returnStruct.Message, m []string) (string, error) {
	if m[0] != "b站" {
		return "", nil
	} else if !groupSwitch[mjson.GroupID] && m[1] != "开启" {
		return "还没有开启相关功能哦！", nil
	}
	m = append(m, "1")
	switch m[1] {
	case "关注":
		return AddSubscribe(mjson.UserID, mjson.GroupID, m[2])
	case "取关":
		fallthrough
	case "取消关注":
		return DeleteSubscribe(mjson.UserID, mjson.GroupID, m[2])
	case "动态":
		fallthrough
	case "查看动态":
		return LookDynamics(mjson, ws, m[2])
	case "直播":
		if len(m) == 3 {
			return getAllLiveByUid(mjson, ws), nil
		} else {
			return AskForLiveRoomStatus(m[2], mjson.GroupID, true)
		}
	case "关注列表":
		return GetSubscribedList(mjson.GroupID, mjson.UserID)
	case "查询用户":
		return GetUserList(m[2])
	case "开启":
		groupLock.Lock()
		if groupSwitch[mjson.GroupID] {
			groupLock.Unlock()
			return "已开启，请勿重复操作！", nil
		} else {
			groupSwitch[mjson.GroupID] = true
			groupLock.Unlock()
			go getAllLiveByGid(mjson, ws)
		}
		return "开启成功！", nil
	case "关闭":
		groupLock.Lock()
		groupSwitch[mjson.GroupID] = false
		groupLock.Unlock()
		myUtil.MsgLog.Println("群 ", mjson.GroupID, " 的b站辅助功能已关闭")
		return "关闭成功！", nil
	}
	return "", nil
}

func StartSubscribeScanner() {
	for myUtil.PublicWs == nil {
	}
	for {
		go ScanSubscribe(myUtil.PublicWs)
		time.Sleep(30 * time.Second)
	}
}
func LiveRoomInfoFormat(msg string, b BiliUp, showCover bool) string {
	msg += "▛" + b.LiveRoom.Title + "▟\n"
	if showCover {
		msg += myUtil.GetBase64CQCode(b.LiveRoom.Cover) + "\n"
	}
	msg += "◉ 开播时间\n" + b.LiveRoom.LiveTime + "\n"
	msg += "◉ 分区\n" + b.LiveRoom.ParentAreaName + " : " + b.LiveRoom.AreaName + "\n"
	msg += "◉ 标签\n" + b.LiveRoom.Tags + "\n"
	msg += "◉ 简介\n" + b.LiveRoom.Description + "\n"
	msg += "◉ 直通车\n" + b.LiveRoom.LiveUrl
	return msg
}
func ScanPersonalSubscribe(ws *websocket.Conn, sendMap map[int64][]int64, subscribe string) {
	v, _ := SubscribedMap.Load(subscribe)
	if v == nil {
		b := &BiliUp{}
		err := b.Init(subscribe)
		if err != nil {
			myUtil.ErrLog.Println("扫描用户关注列表时出错：", "出错关注：", subscribe, "error", err)
			return
		}
		v = b
	}
	go func(v *BiliUp) {
		dynamic, lastTime, err := v.GetDynamic()
		if err != nil {
			myUtil.ErrLog.Println("扫描用户关注列表时出错：", "出错关注：", subscribe, "error：", err)
			return
		}
		if dynamic {
			for i, d := range v.Dynamics {
				if d.Description.Timestamp <= lastTime {
					break
				}
				msg := "\n你关注的 up 主" + v.Info.Data.InfoCard.Name + "刚刚更新了:\n" + v.DynamicAnalysis(i)
				for gid, uList := range sendMap {
					if groupSwitch[gid] {
						go func(uList []int64, gid int64, msg string) {
							for _, v := range uList {
								msg = "[CQ:at,qq=" + strconv.FormatInt(v, 10) + "] " + msg
							}
							myUtil.SendGroupMessage(ws, gid, msg)
						}(uList, gid, msg)
					}
				}
			}
		}
	}(v.(*BiliUp))
	go func(v *BiliUp) {
		info, err := v.GetLiveRoomInfo()
		if err != nil {
			myUtil.ErrLog.Println("检查用户关注 up 直播间时出错：", err)
			return
		}
		if info {
			var msg string
			msg = LiveRoomInfoFormat(msg, *v, true)
			for gid, uList := range sendMap {
				if groupSwitch[gid] {
					go func(uList []int64, gid int64, msg string) {
						msg = "\n你关注的 up 主" + v.Info.Data.InfoCard.Name + "正在直播:\n" + msg
						for _, v := range uList {
							msg = "[CQ:at,qq=" + strconv.FormatInt(v, 10) + "] " + msg
						}
						myUtil.SendGroupMessage(ws, gid, msg)
					}(uList, gid, msg)
				}
			}
		}
	}(v.(*BiliUp))
}

func ScanSubscribe(ws *websocket.Conn) {
	c := data.Db.Collection("subscribe")
	distinct, err := c.Distinct(context.TODO(), "subscribe", bson.D{})
	if err != nil {
		myUtil.ErrLog.Println("获取去重b站关注列表时出现错误,error:", err)
		return
	}
	for _, v := range distinct {
		go func(subscribe string) {
			var sendMap = make(map[int64][]int64)
			filter := bson.D{{"subscribe", subscribe}}
			cur, err := c.Find(context.TODO(), filter)
			if err != nil {
				myUtil.ErrLog.Println("获取组群b站关注列表时出现错误,error:", err)
				return
			}
			for cur.Next(context.TODO()) {
				var elem subscribeList
				err := cur.Decode(&elem)
				if err != nil {
					myUtil.ErrLog.Println(err)
				}
				if len(elem.Uid) == 0 {
					go DeleteZombieSubscribe(subscribe, elem.Gid)
					continue
				}
				sendMap[elem.Gid] = elem.Uid
			}
			ScanPersonalSubscribe(ws, sendMap, subscribe)
		}(v.(string))
	}
}
func DeleteZombieSubscribe(subscribe string, gid int64) {
	filter := bson.D{{"subscribe", subscribe}, {"gid", gid}}
	c := data.Db.Collection("subscribe")
	res, err := c.DeleteOne(context.TODO(), filter)
	if err != nil {
		myUtil.ErrLog.Println("删除僵尸关注时出现错误：", err)
	}
	if res.DeletedCount > 0 {
		myUtil.MsgLog.Println("成功删除僵尸关注：", subscribe)
	}
}
func AddSubscribe(uid int64, gid int64, subscribe string) (string, error) {
	if !myUtil.IsNumber(subscribe) {
		return "请输入正确的upid！", errors.New("subscribe id must be a number")
	}
	v, _ := SubscribedMap.Load(subscribe)
	if v == nil {
		b := &BiliUp{}
		err := b.Init(subscribe)
		if err != nil {
			myUtil.ErrLog.Println("添加关注时出错：", "出错关注：", subscribe, "error", err)
			return "关注失败！可能是网络问题~", err
		}
		v = b
	}
	filter := bson.D{{"subscribe", subscribe}, {"gid", gid}, {"name", v.(*BiliUp).Info.Data.InfoCard.Name}}
	update := bson.D{{"$addToSet", bson.D{{"uid", uid}}}}
	opt := options.Update().SetUpsert(true)
	c := data.Db.Collection("subscribe")
	_, err := c.UpdateOne(context.TODO(), filter, update, opt)
	if err != nil {
		myUtil.ErrLog.Println("添加b站关注时出现异常:\nuid:", uid, "\ngid:", gid, "\nsubscribe:", subscribe, "\nerr:", err)
		return "被叔叔制裁力！", err
	}
	return "关注成功~", nil
}
func DeleteSubscribe(uid int64, gid int64, subscribe string) (string, error) {
	filter := bson.D{{"subscribe", subscribe}, {"gid", gid}}
	update := bson.D{{"$pull", bson.D{{"uid", uid}}}}
	c := data.Db.Collection("subscribe")
	res, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		myUtil.ErrLog.Println("删除b站关注时出现异常:\nuid:", uid, "\ngid:", gid, "\nsubscribe:", subscribe, "\nerr:", err)
		return "被叔叔制裁力！", err
	}
	if res.ModifiedCount == 0 {
		return "貌似没有过相关关注呢~", err
	}
	return "取关成功~", nil
}
func getSubscribeUidByName(name string) (string, error) {
	var elem subscribeList
	filter := bson.D{{"name", bson.D{{"$regex", name}, {"$options", "i"}}}}
	c := data.Db.Collection("subscribe")
	err := c.FindOne(context.TODO(), filter).Decode(&elem)
	if err != nil {
		return "", err
	}
	return elem.Subscribe, nil
}
func LookDynamics(mjson returnStruct.Message, ws *websocket.Conn, match string) (string, error) {
	if !myUtil.IsNumber(match) {
		uid, err := getSubscribeUidByName(match)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				return "未找到相关up，可能是名称错误或者未关注", nil
			} else {
				myUtil.ErrLog.Println("使用名称寻找订阅 upuid 时出现错误：error:", err)
				return "出现了外星人入侵级别的意外！", err
			}
		}
		match = uid
	}
	if v, _ := SubscribedMap.Load(match); v != nil {
		_, _, err := v.(*BiliUp).GetDynamic()
		if err != nil {
			myUtil.ErrLog.Println("查看动态时出错，error:", err)
			return "被叔叔狠狠的抓住了（", err
		}
		go func() {
			var msg []string
			for i := 0; i < len(v.(*BiliUp).Dynamics); i++ {
				msg = append(msg, v.(*BiliUp).DynamicAnalysis(i))
			}
			myUtil.SendForwardMsg(msg, mjson, ws)
		}()
		return "正在努力搬运 up 主" + v.(*BiliUp).Info.Data.InfoCard.Name + "的动态~", nil
	} else {
		return "未找到相关up，可能是名称错误或者未关注", nil
	}
}
func AskForLiveRoomStatus(match string, gid int64, display bool) (string, error) {
	if !myUtil.IsNumber(match) {
		uid, err := getSubscribeUidByName(match)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				return "未找到相关up，可能是名称错误或者未关注", nil
			} else {
				myUtil.ErrLog.Println("使用名称寻找订阅 upuid 时出现错误：error:", err)
				return "出现了外星人入侵级别的意外！", err
			}
		}
		match = uid
	}
	if v, _ := SubscribedMap.Load(match); v != nil {
		_, err := v.(*BiliUp).GetLiveRoomInfo()
		if err != nil {
			myUtil.ErrLog.Println("查看直播状态时出错，error:", err)
			return "被叔叔狠狠的抓住了（", err
		}
		if v.(*BiliUp).LiveRoom.Status {
			return LiveRoomInfoFormat("", *v.(*BiliUp), display), nil
		} else {
			if display {
				return v.(*BiliUp).Info.Data.InfoCard.Name + "当前没有直播哦~", nil
			} else {
				return "", nil
			}
		}
	} else {
		return "未找到相关up，可能是名称错误或者未关注", nil
	}
}
func GetSubscribedList(gid int64, uid int64) (string, error) {
	filter := bson.D{{"gid", gid}, {"uid", uid}}
	c := data.Db.Collection("subscribe")
	cur, err := c.Find(context.TODO(), filter)
	if err != nil {
		myUtil.ErrLog.Println("获取用户b站关注列表时出现错误,error:", err)
		return "查询失败，重试后仍然错误就可以暴打开发者了！", err
	}
	msg := "[CQ:at,qq=" + strconv.FormatInt(uid, 10) + "] 您在本群的关注列表如下：\n"
	for cur.Next(context.TODO()) {
		var elem subscribeList
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err)
			return "查询失败，重试后仍然错误就可以暴打开发者了！", err
		}
		msg += elem.Name + "      " + elem.Subscribe + "\n"
	}
	return msg, nil
}

type UserList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Result []struct {
			Uid    int64  `json:"mid"`
			Name   string `json:"uname"`
			Avatar string `json:"upic"`
		} `json:"result"`
	} `json:"data"`
}

func GetUserList(keyword string) (string, error) {
	get, err := http.Get("https://api.bilibili.com/x/web-interface/wbi/search/type?page_size=5&search_type=bili_user&keyword=" + url.QueryEscape(keyword))
	if err != nil {
		return "被叔叔制裁了（悲", err
	}
	defer get.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(get.Body)
	var res UserList
	err = json.Unmarshal(body, &res)
	if res.Code != 0 {
		return "被叔叔制裁了（悲", errors.New("getting user list error,message:" + res.Message)
	}
	msg := "相关用户搜索结果："
	for _, v := range res.Data.Result {
		msg += "\n\n【用户名】 " + v.Name + "\n【uid】 " + strconv.FormatInt(v.Uid, 10)
	}
	return msg, nil
}
func getAllLiveByGid(mjson returnStruct.Message, ws *websocket.Conn) string {
	list, ok := GetSubscribedListByGid(mjson.GroupID)
	if !ok {
		myUtil.ErrLog.Println("获取群内关注 up 直播信息时出错")
		return "查询失败，重试后仍然错误就可以暴打开发者了！"
	}
	go func() {
		var msg []string
		for _, v := range list {
			status, _ := AskForLiveRoomStatus(v, mjson.GroupID, false)
			if status != "" {
				msg = append(msg, status)
			}
		}
		if len(msg) == 0 {
			myUtil.SendGroupMessage(ws, mjson.GroupID, "没有正在直播中的 up！")
		} else {
			myUtil.SendForwardMsg(LiveInfoSplit(msg), mjson, ws)
		}
	}()
	return "正在获取相关直播信息~"
}
func getAllLiveByUid(mjson returnStruct.Message, ws *websocket.Conn) string {
	list, ok := GetSubscribedListByUid(mjson.GroupID, mjson.UserID)
	if !ok {
		myUtil.ErrLog.Println("获取群内关注 up 直播信息时出错")
		return "查询失败，重试后仍然错误就可以暴打开发者了！"
	}
	go func() {
		var msg []string
		for _, v := range list {
			status, _ := AskForLiveRoomStatus(v, mjson.GroupID, false)
			if status != "" {
				msg = append(msg, status)
			}
		}
		if len(msg) == 0 {
			myUtil.SendGroupMessage(ws, mjson.GroupID, "没有正在直播中的 up！")
		} else {
			myUtil.SendForwardMsg(LiveInfoSplit(msg), mjson, ws)
		}
	}()
	return "正在获取相关直播信息~"
}
func LiveInfoSplit(msg []string) [][]returnStruct.Node {
	var ret [][]returnStruct.Node
	for i := 0; ; i += 5 {
		if i+5 < len(msg) {
			ret = append(ret, myUtil.MakeForwardMsgNode(msg[i:i+5]))
		} else {
			ret = append(ret, myUtil.MakeForwardMsgNode(msg[i:]))
			break
		}
	}
	return ret
}
func GetSubscribedListByGid(gid int64) ([]string, bool) {
	filter := bson.D{{"gid", gid}}
	c := data.Db.Collection("subscribe")
	cur, err := c.Find(context.TODO(), filter)
	if err != nil {
		myUtil.ErrLog.Println("获取群组b站关注列表时出现错误,error:", err, " gid:", gid)
		return nil, false
	}
	var ret []string
	for cur.Next(context.TODO()) {
		var elem subscribeList
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err)
			return nil, false
		}
		ret = append(ret, elem.Subscribe)
	}
	return ret, true
}
func GetSubscribedListByUid(gid int64, uid int64) ([]string, bool) {
	filter := bson.D{{"gid", gid}, {"uid", uid}}
	c := data.Db.Collection("subscribe")
	cur, err := c.Find(context.TODO(), filter)
	if err != nil {
		myUtil.ErrLog.Println("获取用户b站关注列表时出现错误,error:", err, " gid:", gid, " uid:", uid)
		return nil, false
	}
	var ret []string
	for cur.Next(context.TODO()) {
		var elem subscribeList
		err := cur.Decode(&elem)
		if err != nil {
			myUtil.ErrLog.Println(err)
			return nil, false
		}
		ret = append(ret, elem.Subscribe)
	}
	return ret, true
}

type UpInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Mid  string `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"data"`
}

func GetUpInfoByUid(uid string) (*UpInfo, error) {
	get, err := http.Get("https://api.bilibili.com/x/space/acc/info?mid=" + uid)
	if err != nil {
		return nil, err
	}
	defer get.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(get.Body)
	var res UpInfo
	err = json.Unmarshal(body, &res)
	if res.Code != 0 {
		return nil, errors.New("getting up info by uid error,message:" + res.Message)
	}
	return &res, nil
}
