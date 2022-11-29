package bangumi

import (
	"Lealra/myUtil"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func getNtBangumiLink(url string) string {
	res, err := http.Get(url)
	if err != nil {
		myUtil.ErrLog.Println("进一步获取最新集播放地址时出现错误,error:", err)
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		myUtil.ErrLog.Println("status code error:\nStatusCode:", res.StatusCode, "\nStatus:", res.Status)
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		myUtil.ErrLog.Println("解析最新集播放地址时出现错误,error:", err)
		return ""
	}
	var ret string
	ret += "第一集\n"
	doc.Find("#main0 .movurl").Each(func(i int, s *goquery.Selection) {
		h, exist := s.Find("li:first-child a").Attr("href")
		if exist {
			ret += "http://www.ntyou.cc" + h + "\n"
		}
	})
	ret += "最新一集\n"
	doc.Find("#main0 .movurl").Each(func(i int, s *goquery.Selection) {
		h, exist := s.Find("li:last-child a").Attr("href")
		if exist {
			ret += "http://www.ntyou.cc" + h + "\n"
		}
	})
	if ret != "" {
		return strings.TrimRight(ret, "\n")
	}
	return ""
}

func getNtInfo(tag string) ([]Bangumi, error) {
	var doc *goquery.Document
	var err error
	var res *http.Response
	for i := 0; i < 3; i++ {
		res, err = http.Get("http://www.ntyou.cc/search/-------------.html?wd=" + url.QueryEscape(tag) + "&page=1")
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次连接NT动漫城网站时出现错误,error:", err)
			continue
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			myUtil.ErrLog.Println("第", i+1, "次连接NT动漫网status code error:\nStatusCode:", res.StatusCode, "\nStatus:", res.Status)
			err = errors.New("status code error:\nStatusCode:" + strconv.Itoa(res.StatusCode) + "\nStatus:" + res.Status)
			continue
		}
		doc, err = goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次解析NT动漫城网站时出现错误,error:", err)
			continue
		}
		break
	}
	if err != nil {
		myUtil.ErrLog.Println("获取NT动漫数据时出现错误,error:", err)
		return nil, err
	}
	wg := &sync.WaitGroup{}
	lock := sync.Mutex{}
	var bangumiArray []Bangumi
	doc.Find("#container .cell").Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func() {
			var bangumi Bangumi
			bangumi.Index = i
			bangumi.Image, _ = s.Find("a img").Attr("src")
			bangumi.Status = s.Find("a span").Text()
			bangumi.Name = s.Find(".cell_imform .cell_imform_name").Text()
			bangumi.Link = func() string {
				str, _ := s.Find(".cell_poster").Attr("href")
				link := getNtBangumiLink("http://www.ntyou.cc" + str)
				if link == "" {
					return "http://www.ntyou.cc" + str
				} else {
					return link
				}
			}()
			lock.Lock()
			bangumiArray = append(bangumiArray, bangumi)
			lock.Unlock()
			wg.Done()
		}()
	})
	wg.Wait()
	println("NT搜索结束！")
	sort.Slice(bangumiArray, func(i, j int) bool {
		return bangumiArray[i].Index < bangumiArray[j].Index
	})
	return bangumiArray, nil
}
