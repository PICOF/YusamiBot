package bilibili

import (
	"Lealra/myUtil"
	"encoding/json"
	"errors"
	"github.com/dlclark/regexp2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var SubscribedMap sync.Map

type Dynamic struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data DynamicData `json:"data"`
}
type DynamicData struct {
	Cards []DynamicCard `json:"cards"`
}
type CardDesc struct {
	Type      int    `json:"type"`
	Repost    uint32 `json:"repost"`
	Like      uint32 `json:"like"`
	Timestamp uint32 `json:"timestamp"`
}

type EmojiDetail struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

type EmojiInfo struct {
	EmojiDetails []EmojiDetail `json:"emoji_details"`
}

type CardDisplay struct {
	EmojiInfo EmojiInfo `json:"emoji_info"`
}

type DynamicCard struct {
	Description CardDesc    `json:"desc"`
	Content     string      `json:"card"`
	Display     CardDisplay `json:"display"`
}

type LiveRoom struct {
	Status         bool
	RoomId         uint32 `json:"room_id"`
	LiveUrl        string `json:"live_url"`
	LiveTime       string `json:"live_time"`
	Title          string `json:"title"`
	ParentAreaName string `json:"parent_area_name"`
	AreaName       string `json:"area_name"`
	Description    string `json:"description"`
	Tags           string `json:"tags"`
	Cover          string `json:"cover"`
}

type InfoCard struct {
	Name string `json:"name"`
	Face string `json:"face"`
}

type InfoData struct {
	InfoCard InfoCard `json:"card"`
}

type BiliInfo struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    InfoData `json:"data"`
}

type BiliUp struct {
	Uid         string
	Info        BiliInfo
	LastDynamic uint32
	Dynamics    []DynamicCard
	LiveRoom    LiveRoom
}

func (b *BiliUp) Init(uid string) error {
	b.Uid = uid
	SubscribedMap.Store(uid, b)
	get, err := http.Get("https://api.bilibili.com/x/web-interface/card?mid=" + b.Uid)
	if err != nil {
		return err
	}
	defer get.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(get.Body)
	var info BiliInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return err
	}
	b.Info = info
	return nil
}

func (b *BiliUp) GetDynamic() (bool, uint32, error) {
	get, err := http.Get("https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/space_history?host_uid=" + b.Uid)
	if err != nil {
		return false, 0, err
	}
	defer get.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(get.Body)
	var d Dynamic
	err = json.Unmarshal(body, &d)
	if err != nil {
		return false, 0, err
	}
	if d.Data.Cards == nil {
		if d.Code == 0 {
			return false, 0, nil
		}
		return false, 0, errors.New("动态内容获取为空！")
	}
	b.Dynamics = d.Data.Cards
	if b.LastDynamic < d.Data.Cards[0].Description.Timestamp {
		lastTime := b.LastDynamic
		b.LastDynamic = d.Data.Cards[0].Description.Timestamp
		if lastTime == 0 {
			return false, 0, nil
		}
		return true, lastTime, nil
	} else {
		return false, 0, nil
	}
}

func (b *BiliUp) GetLiveRoomInfo() (bool, error) {
	get, err := http.Get("https://api.live.bilibili.com/xlive/web-room/v1/index/getRoomBaseInfo?uids=" + b.Uid + "&;req_biz=videoi")
	if err != nil {
		return false, err
	}
	defer get.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(get.Body)
	compile := regexp2.MustCompile("(?<=\""+b.Uid+"\":{)[^}]*", 0)
	tag, _ := compile.FindStringMatch(string(body))
	if tag == nil {
		b.LiveRoom.Status = false
		return false, nil
	}
	var l LiveRoom
	err = json.Unmarshal([]byte("{"+tag.String()+"}"), &l)
	if err != nil {
		return false, err
	}
	if l.LiveTime == b.LiveRoom.LiveTime {
		return false, nil
	}
	b.LiveRoom = l
	b.LiveRoom.Status = true
	return true, nil
}

type DynamicPic struct {
	ImgSrc string `json:"img_src"`
}

type DynamicItem struct {
	Content     string       `json:"content"`
	Description string       `json:"description"`
	Pictures    []DynamicPic `json:"pictures"`
	OrigType    int          `json:"orig_type"`
}

type NormalDynamic struct {
	Id         int64       `json:"id"`
	Item       DynamicItem `json:"item"`
	Origin     string      `json:"origin"`
	Duration   uint32      `json:"duration"`
	Aid        uint32      `json:"aid"`
	Desc       string      `json:"desc"`
	Dynamic    string      `json:"dynamic"`
	Pic        string      `json:"pic"`
	ShortLink  string      `json:"short_Link"`
	Title      string      `json:"title"`
	Categories []struct {
		Name string `json:"name"`
	} `json:"categories"`
	Summary   string   `json:"summary"`
	BannerUrl string   `json:"banner_url"`
	ImageUrls []string `json:"image_urls"`
	TypeInfo  string   `json:"typeInfo"`
	Cover     string   `json:"cover"`
}

func (b *BiliUp) DynamicAnalysis(index int) string {
	if index > len(b.Dynamics) || index < 0 {
		return ""
	}
	ret := DynamicTranslate(b.Dynamics[index].Description.Type, b.Dynamics[index].Content)
	for _, v := range b.Dynamics[index].Display.EmojiInfo.EmojiDetails {
		ret = strings.ReplaceAll(ret, v.Text, "[CQ:image,file=base64://"+myUtil.Pic2Base64ByUrl(v.Url+"@50w_50h.png")+"]")
	}
	return ret
}
func DynamicTranslate(DynamicType int, content string) string {
	if DynamicType == 1024 {
		return "【该动态已失效】"
	}
	var dynamic NormalDynamic
	err := json.Unmarshal([]byte(content), &dynamic)
	if err != nil {
		myUtil.ErrLog.Println("解析动态时出现错误！error:", err)
		return "【解析时出现错误！】"
	}
	var ret string
	switch DynamicType {
	case 1:
		ret = dynamic.Item.Content + "\n【转发动态】：\n" + DynamicTranslate(dynamic.Item.OrigType, dynamic.Origin)
	case 2:
		//图片动态
		for _, i := range dynamic.Item.Pictures {
			ret += "\n[CQ:image,file=base64://" + myUtil.Pic2Base64ByUrl(i.ImgSrc) + "]"
		}
		fallthrough
	case 4:
		//	文字动态
		ret = dynamic.Item.Content + dynamic.Item.Description + ret
	case 8:
		//	视频投稿
		ret = dynamic.Dynamic + "\n【视频】：" + dynamic.Title + "\n[CQ:image,file=" + dynamic.Pic + "]\n" + dynamic.Desc + "\n【链接】：" + dynamic.ShortLink
	case 64:
		//	专栏投稿
		var tags, pics, banner string
		for _, v := range dynamic.Categories {
			tags += v.Name + " "
		}
		for _, v := range dynamic.ImageUrls {
			pics += "[CQ:image,file=" + v + "]"
		}
		if dynamic.BannerUrl != "" {
			banner = "[CQ:image,file=" + dynamic.BannerUrl + "]"
		}
		ret = banner + "\n【专栏】：" + dynamic.Title + "\n【分类】：" + tags + "\n" + pics + "\n【链接】：" + "https://www.bilibili.com/read/cv" + strconv.FormatInt(dynamic.Id, 10)
	case 256:
		//音频投稿
		ret = "[CQ:image,file=" + dynamic.Cover + "]" + "\n【音频】：" + dynamic.Title + "\n【分类】：" + dynamic.TypeInfo + "\n【链接】：" + "https://www.bilibili.com/audio/au" + strconv.FormatInt(dynamic.Id, 10)
	}
	return ret
}
