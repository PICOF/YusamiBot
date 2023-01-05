package JMComic

import (
	"Lealra/config"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

type queryRes struct {
	Status int        `json:"status"`
	Err    string     `json:"err,omitempty"`
	Res    []ComicSet `json:"res,omitempty"`
	Page   int        `json:"page"`
}

type ComicSet struct {
	Id     string   `json:"id"`
	Img    string   `json:"img"`
	Title  string   `json:"title"`
	Tags   []string `json:"tags"`
	Author []string `json:"author"`
}

var aphorism = []string{
	"小撸怡情，大撸伤身",
	"人类思考的最佳方式就是打个胶线",
	"我不敢与他碰面，因为他一拔枪就意味着他贤者模式的终结",
	"没有东西能在他的大炮下撑过60秒——除了他自己的左右手",
	"我能奖励自己一整天\n\t——史蒂夫·俊杰",
}

func BenziBot(msg []string, mjson returnStruct.Message) (string, error) {
	var res string
	if len(msg) < 2 || (msg[0] != "看本子" && msg[0] != "搜本子") {
		return "", nil
	} else {
		switch len(msg) {
		case 2:
			if msg[0] == "看本子" {
				for _, r := range msg[1] {
					if !unicode.IsNumber(r) {
						return "", nil
					}
				}
				encode, err := qrcode.Encode("http://"+config.Settings.Server.Hostname+":"+config.Settings.Server.Port+"/getComic/"+msg[1], qrcode.Medium, 256)
				if err != nil {
					myUtil.ErrLog.Println("生成本子搜索二维码时出现异常！err:", err)
					return "初始化失败啦！", err
				}
				res = "[CQ:image,file=base64://" + base64.StdEncoding.EncodeToString(encode) + "]"
				break
			}
			fallthrough
		default:
			query := strings.ReplaceAll(mjson.Message[len(msg[0])+1:], "&amp;", "&")
			query = strings.ReplaceAll(query, "&#91;", "[")
			query = strings.ReplaceAll(query, "&#93;", "]")
			encode, err := qrcode.Encode("http://"+config.Settings.Server.Hostname+":"+config.Settings.Server.Port+"/comic/"+query+"/1", qrcode.Medium, 256)
			if err != nil {
				myUtil.ErrLog.Println("生成本子搜索二维码时出现异常！err:", err)
				return "初始化失败啦！", err
			}
			res = "[CQ:image,file=base64://" + base64.StdEncoding.EncodeToString(encode) + "]"
		}
		return aphorism[int([]rune(msg[1])[0])%len(aphorism)] + res, nil
	}
}

func findComic(query string, ws *websocket.Conn, mjson returnStruct.Message) (string, error) {
	get, err := http.Get("http://127.0.0.1:5000/search?query=" + url.QueryEscape(query))
	if err != nil {
		myUtil.ErrLog.Println("连接本地禁漫搜索端口失败！error:", err)
		return "怎么会是呢，网断了，真的怪", err
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var res queryRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		myUtil.ErrLog.Println("解析禁漫搜索结果失败！error:", err)
		return "搜出来的是无字天书……", err
	}
	if res.Status != 0 {
		myUtil.ErrLog.Println("禁漫搜索端口可能已失效！错误码:", res.Status)
		return "怎么会是呢，网断了，真的怪", errors.New(res.Err)
	}
	if len(res.Res) == 0 {
		return "没有找到相关本子哦~", errors.New("未找到相应本子资源")
	}
	var comics []string
	for i, v := range res.Res {
		if i == 10 {
			break
		}
		//"[CQ:image,file="+v.Img+",subType=0]"+"\ntitle:"+v.Title+"\n车牌号"+v.Id+"\ntags:"+strings.Join(v.Tags, " ")
		comics = append(comics, "[CQ:image,file="+v.Img+",subType=0]"+"\ntitle:"+v.Title+"\n车牌号"+v.Id)
	}
	err = myUtil.SendForwardMsg(comics, mjson)
	if err != nil {
		return config.Settings.BotName.Name + "的消息被截胡了！", err
	}
	return "", nil
}

func FindComic(query string, page string) ([]ComicSet, int, string) {
	get, err := http.Get("http://127.0.0.1:5000/search?query=" + url.QueryEscape(query) + "&page=" + page)
	if err != nil {
		myUtil.ErrLog.Println("连接本地禁漫搜索端口失败！error:", err)
		return nil, 0, "相关服务未启动，请联系开发者"
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var res queryRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		myUtil.ErrLog.Println("解析禁漫搜索结果失败！error:", err)
		return nil, 0, "相关结果解析出现错误，请检查输入"
	}
	if res.Status != 0 {
		myUtil.ErrLog.Println("禁漫搜索端口可能已失效！错误码:", res.Status, "error:", res.Err)
		return nil, 0, "网络异常，请稍后重试！"
	}
	if len(res.Res) == 0 {
		return nil, 0, "未找到相应本子资源"
	}
	return res.Res, res.Page, ""
}

func getComic(id string, ws *websocket.Conn, mjson returnStruct.Message) (string, error) {
	img, _, _ := sexyComicHandler(id)
	if img == nil {
		return "我图图呢！", errors.New("failed to get image")
	}
	err := myUtil.SendForwardMsg(packMsg(img), mjson)
	if err != nil {
		return config.Settings.BotName.Name + "的图图被截胡了！", err
	}
	return "", nil
}

var noSex = []string{
	"原神648只能抽81次，我觉得有点坑",
	"兄弟们帮我看看这号是不是废了",
	"扣一不减功德，扣一送地狱火",
	"你说得对，但是科比永远是我心中最伟大的机车族战士",
	"关注永雏塔菲喵，关注永雏塔菲谢谢喵",
	"差不多得了，你给我整笑了",
	"叔叔我啊，真的要生气了",
	"好强的攻击性，入典！",
	"你说得对，但我还是感觉不如原神，至少这个游戏根本不能偷盘子抢椅子",
	"再玩雀魂我就是傻逼",
	"一边在桌面上看到美好的二次元",
	"我直接跳了给你们送保研吧，都过来一下蹭一下保研",
}

func packMsg(img []image.Image) []string {
	var msg []string
	msg = append(msg, "看看，这是可爱的猫猫")
	for i, v := range img {
		content := ""
		if v != nil {
			buffer := bytes.NewBuffer(nil)
			jpeg.Encode(buffer, v, nil)
			//png.Encode(buffer, v)
			content = "[CQ:image,file=base64://" + base64.StdEncoding.EncodeToString(buffer.Bytes()) + ",subType=0]" + noSex[i%len(noSex)]
		}
		msg = append(msg, content)
	}
	msg = append(msg, "太可爱了，动物是人类最好的朋友！")
	return msg
}
