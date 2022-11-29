package JMComic

import (
	"Lealra/myUtil"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/image/webp"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"sync"
)

func drawFullImage(imgSet []image.Image, width int, height int) image.Image {
	res := image.NewRGBA(image.Rect(0, 0, width, height))
	y := 0
	for _, img := range imgSet {
		draw.Draw(res, res.Bounds(), img, image.Pt(0, -y), draw.Over)
		y += img.Bounds().Size().Y
	}
	return res
}

type ComicPage struct {
	StatusCode int    `json:"statusCode"`
	ComicImage string `json:"comicImage"`
	Index      int    `json:"index"`
}

func imageToBase64(img image.Image) string {
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	send := buf.Bytes()
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(send)
}
func GetSexyComicPage(id string, i int) ComicPage {
	var content []byte
	var img image.Image
	get, err := http.Get("https://cdn-msp.18comic.org/media/photos/" + id + "/" + fmt.Sprintf("%05d", i) + ".webp")
	if err != nil {
		myUtil.ErrLog.Println("获取"+id+"号本子图片时出现错误,error:", err)
		return ComicPage{1, "", i - 1}
	}
	defer get.Body.Close()
	if get.Header.Get("Content-Length") == "0" {
		return ComicPage{2, "", i - 1}
	}
	content, err = ioutil.ReadAll(get.Body)
	if err != nil {
		myUtil.ErrLog.Println("解析"+id+"号本子图片时出现错误,error:", err)
		return ComicPage{1, "", i - 1}
	}
	img, err = sexyImageHandler(content, id, fmt.Sprintf("%05d", i))
	if err != nil {
		return ComicPage{1, "", i - 1}
	}
	return ComicPage{0, imageToBase64(img), i - 1}
}
func GetPageNum(id string) int {
	var get *http.Response
	var err error
	left := 0
	right := 1000
	var mid int
	for left < right {
		mid = (left + right) / 2
		get, err = http.Get("https://cdn-msp.18comic.org/media/photos/" + id + "/" + fmt.Sprintf("%05d", mid) + ".webp")
		if err != nil {
			return 0
		}
		if get.Header.Get("content-length") == "0" {
			right = mid
		} else {
			left = mid + 1
		}
	}
	defer get.Body.Close()
	return left
}
func GetSexyComic(id string, ws *websocket.Conn, start int, end int) bool {
	var imgWait sync.WaitGroup
	var lock sync.Mutex
	isEnd := false
	for i := start; i < end; i++ {
		imgWait.Add(1)
		go func(index int) {
			page := GetSexyComicPage(id, index)
			if page.StatusCode == 2 {
				isEnd = true
			} else {
				lock.Lock()
				err := ws.WriteJSON(page)
				lock.Unlock()
				if err != nil {
					myUtil.ErrLog.Println("传输"+id+"号本子第"+strconv.Itoa(index)+"页图片时失败,error:", err)
				}
			}
			imgWait.Done()
		}(i)
	}
	imgWait.Wait()
	return isEnd
}
func sexyComicHandler(id string) (imgSet []image.Image, width int, height int) {
	var ret []image.Image
	height = 0
	width = 0
	var get *http.Response
	var err error
	var content []byte
	for i := 1; i < 40; i++ {
		get, err = http.Get("https://cdn-msp.18comic.org/media/photos/" + id + "/" + fmt.Sprintf("%05d", i) + ".webp")
		if err != nil {
			myUtil.ErrLog.Println("获取"+id+"号本子图片时出现错误,error:", err)
			return nil, 0, 0
		}
		content, err = ioutil.ReadAll(get.Body)
		if len(content) <= 10 {
			myUtil.MsgLog.Println(id + "号本子解析完成")
			break
		}
		if err != nil {
			myUtil.ErrLog.Println("解析"+id+"号本子图片时出现错误,error:", err)
			return nil, 0, 0
		}
		img, err := sexyImageHandler(content, id, fmt.Sprintf("%05d", i))
		if err != nil {
			return nil, 0, 0
		}
		width = int(math.Max(float64(width), float64(img.Bounds().Size().X)))
		height += img.Bounds().Size().Y
		ret = append(ret, img)
	}
	defer get.Body.Close()
	return ret, width, height
}
func sexyImageHandler(content []byte, id string, index string) (img image.Image, err error) {
	decode, err := webp.Decode(bytes.NewBuffer(content))
	if err != nil {
		myUtil.ErrLog.Println("解码"+id+"号本子的第"+index+"张图时出现异常！error:", err)
		return nil, err
	}
	res := image.NewRGBA(image.Rect(0, 0, decode.Bounds().Size().X, decode.Bounds().Size().Y))
	rgba := decode.(*image.YCbCr)
	num := getCutNum(id, index)
	height := decode.Bounds().Size().Y / num
	for i := 0; i < num; i++ {
		var h int
		if i == num-1 {
			h = height + decode.Bounds().Size().Y%num
		} else {
			h = height
		}
		subImage := rgba.SubImage(image.Rect(0, i*height, decode.Bounds().Size().X, i*height+h)).(*image.YCbCr)
		draw.Draw(res, res.Bounds(), subImage, image.Pt(0, 2*i*height+h-decode.Bounds().Size().Y), draw.Over)
	}
	return res, nil
}
func getCutNum(id string, index string) int {
	a := 1
	if len(id) >= len("220971") && id > "220971" {
		a = 10
	}
	if len(id) >= len("268850") && id > "268850" {
		str := fmt.Sprintf("%x", md5.Sum([]byte(id+index)))
		str = str[len(str)-1:]
		k := []rune(str)[0] % 10
		switch k {
		case 0:
			a = 2
		case 1:
			a = 4
		case 2:
			a = 6
		case 3:
			a = 8
		case 4:
			a = 10
		case 5:
			a = 12
		case 6:
			a = 14
		case 7:
			a = 16
		case 8:
			a = 18
		case 9:
			a = 20
		}
	}
	return a
}
