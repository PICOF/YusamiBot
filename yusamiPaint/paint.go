package yusamiPaint

import (
	"Lealra/config"
	"Lealra/myUtil"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var filterMap = map[string]string{
	"，": ",",
	"（": "(",
	"）": ")",
	"｛": "{",
	"｝": "}",
}

var tokenLock sync.Mutex
var tokenIndex = 0

func GetPicFree(tags string, nTags string, shape string, nsfw string, seed string) (string, error) {
	for oldChr, newChr := range filterMap {
		tags = strings.ReplaceAll(tags, oldChr, newChr)
		nTags = strings.ReplaceAll(nTags, oldChr, newChr)
	}
	var get *http.Response
	var err error
	tokenLock.Lock()
	token := config.Settings.AiPaint.Free.Token[tokenIndex]
	tokenIndex = (tokenIndex + 1) % len(config.Settings.AiPaint.Free.Token)
	tokenLock.Unlock()
	myUtil.MsgLog.Println("使用Token：" + token + "进行绘图")
	get, err = http.Get(config.Settings.AiPaint.Free.Api + "/got_image?tags=" + url.QueryEscape(tags) + "&ntags=" + url.QueryEscape(nTags) + "&token=" + token + "&shape=" + shape + "&r18=" + nsfw + seed)
	if err != nil {
		myUtil.ErrLog.Println("请求网站绘图时出现异常！error:", err)
		return "没找到我的画笔QAQ", err
	}
	defer get.Body.Close()
	if get.Header.Get("Seed") == "" {
		myUtil.ErrLog.Println("请求网站绘图时出现异常！error:", get.Status, " ", get.StatusCode)
		return "网络塞车啦！过一会儿再试呢~", err
	}
	content, err := io.ReadAll(get.Body)
	if err != nil {
		myUtil.ErrLog.Println("读取ai绘图response时出现异常！error:", err)
		return "我画了个什么(((φ(◎ロ◎;)φ)))", err
	}
	if nsfw == "0" {
		err = os.WriteFile("pic/"+time.Now().Format("2006-01-02T3-04PM")+get.Header.Get("Seed")+".png", content, 0666)
		if err != nil {
			myUtil.ErrLog.Println("保存ai绘制图时出现异常！error:", err)
		}
		return "[CQ:image,file=base64://" + base64.StdEncoding.EncodeToString(content) + "]\n种子:" + get.Header.Get("Seed") + "\ntags:" + tags + "\nntags:" + nTags, nil
	} else {
		filename := time.Now().Format("2006-01-02T3-04PM") + get.Header.Get("Seed") + ".png"
		err = os.WriteFile("nsfw/"+filename, content, 0666)
		if err != nil {
			myUtil.ErrLog.Println("保存ai🐍图时出现异常！error:", err)
		}
		res, err := myUtil.SeseQrcode("something/study", filename)
		if err != nil {
			return "不给看！", err
		}
		return "[CQ:image,file=base64://" + res + "]\n种子:" + get.Header.Get("Seed"), nil
	}
}
