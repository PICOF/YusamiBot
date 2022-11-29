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

func getCycdmBangumiLink(url string) string {
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
	h, exist := doc.Find(".anthology-list-play li:first-child a").Attr("href")
	if exist {
		ret += "第一集\n" + "https://www.cycdm01.top" + h + "\n"
	}
	h, exist = doc.Find(".anthology-list-play li:last-child a").Attr("href")
	if exist {
		ret += "最新一集\n" + "https://www.cycdm01.top" + h
	}
	if ret != "" {
		return ret
	} else {
		return ""
	}
}

func getCycdmInfo(tag string) ([]Bangumi, error) {
	var doc *goquery.Document
	var err error
	var res *http.Response
	for i := 0; i < 3; i++ {
		res, err = http.Get("https://www.cycdm01.top/search.html?wd=" + url.QueryEscape(tag))
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次连接次元动漫城网站时出现错误,error:", err)
			continue
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			myUtil.ErrLog.Println("第", i+1, "次连接次元动漫城网站status code error:\nStatusCode:", res.StatusCode, "\nStatus:", res.Status)
			err = errors.New("status code error:\nStatusCode:" + strconv.Itoa(res.StatusCode) + "\nStatus:" + res.Status)
			continue
		}
		doc, err = goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次解析次元动漫城网站时出现错误,error:", err)
			continue
		}
		break
	}
	if err != nil {
		myUtil.ErrLog.Println("获取次元动漫城数据时出现错误,error:", err)
		return nil, err
	}
	wg := &sync.WaitGroup{}
	lock := sync.Mutex{}
	var bangumiArray []Bangumi
	doc.Find(".box-width .search-box").Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func() {
			var bangumi Bangumi
			bangumi.Index = i
			backgroundImg, _ := s.Find(".cover").Attr("style")
			bangumi.Image = backgroundImg[strings.LastIndex(backgroundImg, "(")+1 : len(backgroundImg)-2]
			bangumi.Status = s.Find(".left .public-list-exp .public-list-prb").Text()
			bangumi.Name = s.Find(".right .thumb-content .thumb-txt").Text()
			bangumi.Link = func() string {
				str, _ := s.Find(".left .public-list-exp").Attr("href")
				link := getCycdmBangumiLink("https://www.cycdm01.top" + str)
				if link == "" {
					return "https://www.cycdm01.top" + str
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
	sort.Slice(bangumiArray, func(i, j int) bool {
		return bangumiArray[i].Index < bangumiArray[j].Index
	})
	println("次元动漫城搜索结束！")
	return bangumiArray, nil
}
