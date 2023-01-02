package news

import (
	"Lealra/config"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Bangumi struct {
	BackgroundImage string `json:"backgroundImage"`
	TranslatedName  string `json:"translatedName"`
	Name            string `json:"name"`
}
type Weekday struct {
	Bangumi []Bangumi `json:"bangumi"`
}

type BangumiCalendar struct {
	Weekdays   []Weekday `json:"weekdays"`
	CreateDate int64     `json:"create"`
}

var bangumiCalendar BangumiCalendar
var weekdays = []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

func BangumiNewsHandler(msg []string, mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	if msg[0] != "番剧" || len(msg) < 2 {
		return "", nil
	}
	switch msg[1] {
	case "今日更新":
		if len(msg) == 2 {
			res, err := bangumiOfWeekday(time.Now().Weekday())
			if err != nil {
				return "呜哇哇~失去与二次元之间的连接了！", err
			}
			err = myUtil.SendForwardMsg(res, mjson, ws)
			if err != nil {
				myUtil.ErrLog.Println("发送bangumi今日时间表时出现错误,error:", err)
				return config.Settings.BotName.Name + " 的电波被邪恶的大魔王拦截了！", err
			}
		}
	case "搜索":
		res, err := findBangumi(strings.Join(msg[2:], " "), mjson, ws)
		if err != nil {
			return res, err
		}
	case "全部番剧":
		if len(msg) == 2 {
			res, err := bangumiOfOneWeek()
			err = myUtil.SendForwardMsg(res, mjson, ws)
			if err != nil {
				myUtil.ErrLog.Println("发送bangumi整周时间表时出现错误,error:", err)
				return config.Settings.BotName.Name + " 的电波被邪恶的大魔王拦截了！", err
			}
		}
	default:
		return "", nil
	}
	return "合理安排时间，愉悦追番哦~(*^_^*)", nil
}

func packMsg(bangumi Bangumi) string {
	var msg string
	if bangumi.BackgroundImage != "" {
		msg += "[CQ:image,file=" + bangumi.BackgroundImage + ",subType=0]"
	} else {
		msg += "【图片丢失】error:22"
	}
	if bangumi.TranslatedName != "" {
		msg += "\n" + bangumi.TranslatedName
	}
	if bangumi.Name != "" {
		msg += "\n" + bangumi.Name
	}
	return msg
}

func findBangumi(target string, mjson returnStruct.Message, ws *websocket.Conn) (string, error) {
	var table BangumiCalendar
	var err error
	table, err = renewTable()
	if err != nil {
		return "呜哇哇~失去与二次元之间的连接了！", err
	}
	var msgs []string
	for i, weekday := range table.Weekdays {
		for _, bangumi := range weekday.Bangumi {
			if strings.Contains(bangumi.TranslatedName, target) || strings.Contains(bangumi.Name, target) {
				msgs = append(msgs, "【"+weekdays[i]+"】\n"+packMsg(bangumi))
			}
		}
	}
	if len(msgs) == 0 {
		myUtil.MsgLog.Println("cannot find bangumi")
		return "未找到相关番剧呢，换个词再搜搜吧~", nil
	}
	err = myUtil.SendForwardMsg(msgs, mjson, ws)
	if err != nil {
		myUtil.ErrLog.Println("发送bangumi今日时间表时出现错误,error:", err)
		return config.Settings.BotName.Name + " 的电波被邪恶的大魔王拦截了！", err
	}
	return "", nil
}

func bangumiOfWeekday(weekday time.Weekday) ([]string, error) {
	var table BangumiCalendar
	var err error
	table, err = renewTable()
	if err != nil {
		return nil, err
	}
	today := table.Weekdays[weekday]
	var msg []string
	for _, bangumi := range today.Bangumi {
		msg = append(msg, packMsg(bangumi))
	}
	return msg, nil
}

func bangumiOfOneWeek() ([]interface{}, error) {
	var res []string
	var err error
	var ret []interface{}
	for i, v := range weekdays {
		res, err = bangumiOfWeekday(time.Weekday(i))
		if err != nil {
			return nil, err
		}
		ret = append(ret, "【"+v+"】⬇")
		ret = append(ret, myUtil.MakeForwardMsgNode(res))
	}
	return ret, nil
}

func renewTable() (BangumiCalendar, error) {
	if bangumiCalendar.CreateDate != 0 && time.Now().Unix()-bangumiCalendar.CreateDate < 86400 {
		return bangumiCalendar, nil
	}
	res, err := http.Get("https://bangumi.tv/calendar")
	if err != nil {
		myUtil.ErrLog.Println("连接bangumi时间表时出现错误,error:", err)
		return BangumiCalendar{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		myUtil.ErrLog.Println("status code error:\nStatusCode:", res.StatusCode, "\nStatus:", res.Status)
		return BangumiCalendar{}, errors.New("status code error:\nStatusCode:" + strconv.Itoa(res.StatusCode) + "\nStatus:" + res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		myUtil.ErrLog.Println("根据bangumi页面创建文档时出现错误,error:", err)
		return BangumiCalendar{}, err
	}
	var calendar BangumiCalendar
	calendar.CreateDate = time.Now().Unix()
	doc.Find("#colunmSingle").Find("li.week ").Each(func(i int, s *goquery.Selection) {
		var weekday Weekday
		s.Find(".coverList li").Each(func(j int, s *goquery.Selection) {
			var bangumi Bangumi
			backgroundImg, _ := s.Attr("style")
			bangumi.BackgroundImage = "https:" + strings.Split(backgroundImg, "'")[1]
			bangumi.TranslatedName = s.Find(".info p:first-child").Text()
			bangumi.Name = s.Find(".info small em").Text()
			weekday.Bangumi = append(weekday.Bangumi, bangumi)
		})
		calendar.Weekdays = append(calendar.Weekdays, weekday)
	})
	return calendar, nil
}
