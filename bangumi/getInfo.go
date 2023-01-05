package bangumi

import (
	"Lealra/config"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"github.com/dlclark/regexp2"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
)

var searchQueue sync.Mutex

type Bangumi struct {
	Index  int
	Image  string
	Name   string
	Status string
	Link   string
}

func packMsg(bangumi []Bangumi, source string, msg []interface{}) []interface{} {
	if len(bangumi) == 0 {
		msg = append(msg, source+"上好像暂无片源呢~")
	} else {
		msg = append(msg, source+"上相关结果如下~")
		var resSet []returnStruct.Node
		for _, v := range bangumi {
			var a returnStruct.Node
			a = returnStruct.Node{}
			a.Type = "node"
			a.Data.Uid = config.Settings.BotName.Id
			a.Data.Name = config.Settings.BotName.FullName
			if v.Image != "" {
				v.Image = "[CQ:image,file=" + v.Image + ",subType=0]\n"
			}
			a.Data.Content = v.Image + v.Name + "\n【状态】 " + v.Status + "\n【Link】⬇"
			resSet = append(resSet, a)
			a.Data.Content = v.Link
			resSet = append(resSet, a)
		}
		msg = append(msg, resSet)
	}
	return msg
}
func WatchBangumi(mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	compile := regexp2.MustCompile("(?<=我要看|我现在就要看|番剧搜索).+", 0)
	matched, _ := compile.FindStringMatch(mjson.RawMessage)
	if matched == nil {
		return "", nil
	}
	searchQueue.Lock()
	if config.Settings.Bangumi.MaxPoolSize > 0 {
		config.Settings.Bangumi.MaxPoolSize--
		searchQueue.Unlock()
	} else {
		searchQueue.Unlock()
		return "当前搜索队列已满哦~请耐心等待~", nil
	}
	err := myUtil.SendNotice(mjson, "正在努力寻找中……")
	if err != nil {
		myUtil.ErrLog.Println("动漫搜索返回结果时出现错误,error:", err)
	}
	tag := strings.Trim(matched.String(), " ")
	var msg []interface{}
	wg := &sync.WaitGroup{}
	lock := sync.Mutex{}
	wg.Add(1)
	go func() {
		cycdm, err := getCycdmInfo(tag)
		if err == nil {
			lock.Lock()
			msg = packMsg(cycdm, "次元动漫城", msg)
			lock.Unlock()
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		nt, err := getNtInfo(tag)
		if err == nil {
			lock.Lock()
			msg = packMsg(nt, "NT动漫", msg)
			lock.Unlock()
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		dmhy, err := getDmhyInfo(tag)
		if err == nil {
			lock.Lock()
			msg = packMsg(dmhy, "动漫花园", msg)
			lock.Unlock()
		}
		wg.Done()
	}()
	wg.Wait()
	searchQueue.Lock()
	if config.Settings.Bangumi.MaxPoolSize < 3 {
		config.Settings.Bangumi.MaxPoolSize++
	}
	searchQueue.Unlock()
	if len(msg) == 0 {
		return "可恶！这被地球束缚的信号！", nil
	}
	err = myUtil.SendForwardMsg(msg, mjson)
	if err != nil {
		myUtil.ErrLog.Println("发送各大动漫网站搜索结果时出现错误,error:", err)
		return "可恶！这被地球束缚的信号！", err
	}
	return "合理安排时间，愉悦追番哦~(*^_^*)", nil
}
