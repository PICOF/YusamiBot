package main

import (
	"Lealra/JMComic"
	"Lealra/config"
	"Lealra/data"
	"Lealra/handler"
	"Lealra/myUtil"
	"Lealra/note"
	"Lealra/returnStruct"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	config.GetSetting()
	data.ConnInit()
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatalln(err)
		}
	}(data.Db.Client(), context.TODO())
	myUtil.SetLogger()
	go myUtil.RenewLoggers()
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	g := gin.Default()
	g.LoadHTMLGlob("diary/template/*")
	g.Static("/static", "diary/static")
	g.Static("/something/study", "nsfw")
	g.Static("/something/review", "pixiv/h")
	g.Static("/something/plan", "pixiv/safe")
	g.GET("/msg", func(c *gin.Context) {
		//进行一个websocket的升
		ws, err := upgrade.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatalln(err)
		}
		defer ws.Close()
		go func() {
			<-c.Done()
			fmt.Println("连接已断开", err.Error())
		}()
		for {
			// 读取客户端发送过来的消息，如果没发就会一直阻塞住
			_, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("消息读取异常:", err.Error())
				fmt.Println(err)
				break
			}
			go func() {
				res, err := handler.MsgHandler(ws, message, []int{config.Settings.Mode.GroupFilterMode, config.Settings.Mode.PrivateFilterMode})
				if res != nil {
					myUtil.WsLock.Lock()
					err := ws.WriteMessage(returnStruct.MsgType, res)
					myUtil.WsLock.Unlock()
					if err != nil {
						myUtil.ErrLog.Println("发送失败！: ", err.Error())
					}
				}
				if err != nil {
					myUtil.ErrLog.Println("发生了神奇的错误: ", err.Error())
				}
			}()
			//err = ws.WriteMessage(mt, []byte("{\"action\": \"send_private_msg\",\"params\": {\"user_id\": \"1730483316\",\"message\":"+"\"sd\""+"}}"))
			//if err != nil {
			//	fmt.Println(err)
			//	break
			//}
		}
	})
	g.GET("/diary/:uid/:date", func(c *gin.Context) {
		strUid := c.Param("uid")
		date := c.Param("date")
		if config.Settings.Func.PrivateDiary && strUid == "1730483316" {
			return
		}
		uid, err := strconv.ParseInt(strUid, 10, 64)
		if err != nil {
			myUtil.ErrLog.Println("访问每日实时记录时参数错误: ", err.Error())
			return
		}
		notes, err := note.GetNotes(uid, date)
		if err != nil {
			myUtil.ErrLog.Println("获取每日实时记录失败！")
			return
		}
		c.HTML(http.StatusOK, "diary.tmpl", struct {
			Content []template.HTML
			Date    string
		}{Content: note.Analyse(notes), Date: date})
	})
	g.GET("/MyDiary", func(c *gin.Context) {
		date := time.Now().Format("2006-01-02")
		notes, err := note.GetNotes(1730483316, date)
		if err != nil {
			myUtil.ErrLog.Println("获取每日实时记录失败！")
			return
		}
		c.HTML(http.StatusOK, "diary.tmpl", struct {
			Content []template.HTML
			Date    string
		}{Content: note.Analyse(notes), Date: date})
	})
	g.GET("/getComic/:id", func(c *gin.Context) {
		pageNum := JMComic.GetPageNum(c.Param("id"))
		println(pageNum)
		c.HTML(http.StatusOK, "comic.tmpl", gin.H{
			"id":      c.Param("id"),
			"pageNum": pageNum,
		})
	})
	g.GET("/getComic/:id/:page", func(c *gin.Context) {
		page, err := strconv.Atoi(c.Param("page"))
		var statusCode int
		var res JMComic.ComicPage
		if err != nil {
			statusCode = http.StatusBadRequest
			res = JMComic.ComicPage{StatusCode: 400}
		} else {
			statusCode = http.StatusOK
			res = JMComic.GetSexyComicPage(c.Param("id"), page+1)
		}
		c.JSON(statusCode, gin.H{
			"statusCode": res.StatusCode,
			"index":      res.Index,
			"comicImage": res.ComicImage,
		})
	})
	g.GET("/comicWs/:id", func(c *gin.Context) {
		var err error
		var ws *websocket.Conn
		var message []byte
		var req, step int
		ws, err = upgrade.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			myUtil.ErrLog.Println(c.Param("id") + "号本子传输通道建立失败" + err.Error())
			return
		}
		defer ws.Close()
		for {
			_, message, err = ws.ReadMessage()
			if err != nil {
				myUtil.ErrLog.Println("看本子交互消息读取异常:", err.Error())
				myUtil.ErrLog.Println(err)
				break
			}
			pack := strings.Split(string(message), " ")
			req, err = strconv.Atoi(pack[0])
			if err != nil {
				return
			}
			step, err = strconv.Atoi(pack[1])
			if err != nil {
				return
			}
			if JMComic.GetSexyComic(c.Param("id"), ws, req, req+step) {
				break
			}
		}
		myUtil.MsgLog.Println(c.Param("id") + "号本子传输已结束")
	})
	g.GET("/comic", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/comic/"+c.Query("query")+"/1")
	})
	g.GET("/comic/:query/:page", func(c *gin.Context) {
		comic, page, err := JMComic.FindComic(c.Param("query"), c.Param("page"))
		var content []template.HTML
		var unit string
		for _, v := range comic {
			unit = ""
			unit += "<div><img src=\"" + strings.Replace(v.Img, "cdn-msp2", "cdn-msp", 1) + "\" onerror=\"loadFailed(this)\"></div>"
			unit += "<div class=\"title\">" + v.Title + "</div>"
			unit += "<div class=\"author\">作者：" + strings.Join(v.Author, "、") + "</div>"
			unit += "<div class=\"tags\">标签：" + strings.Join(v.Tags, " · ") + "</div>"
			unit += "<div class=\"id\">车牌号：<a href=\"/getComic/" + v.Id + "\">" + v.Id + "</a></div>"
			content = append(content, template.HTML(unit))
		}
		c.HTML(http.StatusOK, "comicSearch.tmpl", struct {
			Topic   string
			Content []template.HTML
			Page    int
			Index   string
			Err     string
		}{Topic: c.Param("query"), Content: content, Page: page, Index: c.Param("page"), Err: err})
	})
	err := g.Run(":" + config.Settings.Server.Port)
	if err != nil {
		fmt.Println("Failed to start...,err:", err)
		return
	}
}
