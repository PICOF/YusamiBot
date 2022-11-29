package bangumi

import (
	"Lealra/myUtil"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type dmhyInfo struct {
	Data dmhyData `json:"data"`
}
type dmhyData struct {
	ResultSet []dmhyResultSet `json:"searchData"`
}

type dmhyResultSet struct {
	Id    int64  `json:"id"`
	Date  string `json:"date"`
	Title string `json:"title"`
	Link  string `json:"link"`
	Size  string `json:"size"`
}
type detailInfo struct {
	Data detailData `json:"data"`
}
type detailData struct {
	FileLink    string `json:"fileLink"`
	MagnetLink1 string `json:"magnetLink1"`
	MagnetLink2 string `json:"magnetLink2"`
}

func dmhyInfoHandler(info dmhyInfo) []Bangumi {
	var ret []Bangumi
	wg := &sync.WaitGroup{}
	lock := sync.Mutex{}
	for i, r := range info.Data.ResultSet {
		wg.Add(1)
		go func() {
			postData := url.Values{}
			postData.Add("link", r.Link)
			postData.Add("id", strconv.FormatInt(r.Id, 10))
			post, err := http.Post("https://dongmanhuayuan.myheartsite.com/api/acg/detail", "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
			if err != nil {
				myUtil.ErrLog.Println("获取动漫花园详细结果时出现错误,error:", err)
				wg.Done()
				return
			}
			defer post.Body.Close()
			all, err := ioutil.ReadAll(post.Body)
			if err != nil {
				myUtil.ErrLog.Println("解析动漫花园详细结果时出现错误,error:", err)
				wg.Done()
				return
			}
			var res detailInfo
			err = json.Unmarshal(all, &res)
			if err != nil {
				myUtil.ErrLog.Println("转换动漫花园详细结果为数据类型时出现错误,error:", err)
				wg.Done()
				return
			}
			var a Bangumi
			a.Index = i
			a.Name = r.Title + "【" + r.Size + "】"
			a.Status = r.Date
			a.Link = "【torrent】\nhttps:" + res.Data.FileLink
			if res.Data.MagnetLink1 != "" {
				a.Link += "\n【magnet1】\n" + res.Data.MagnetLink1
			}
			if res.Data.MagnetLink2 != "" {
				a.Link += "\n【magnet2】\n" + res.Data.MagnetLink2
			}
			lock.Lock()
			ret = append(ret, a)
			lock.Unlock()
			wg.Done()
		}()
		if i == 10 {
			break
		}
	}
	wg.Wait()
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Index < ret[j].Index
	})
	return ret
}

func getDmhyInfo(tag string) ([]Bangumi, error) {
	postData := url.Values{}
	postData.Add("keyword", tag)
	postData.Add("page", "1")
	postData.Add("searchType", "0")
	var err error
	var post *http.Response
	var all []byte
	var res dmhyInfo
	for i := 0; i < 3; i++ {
		post, err = http.Post("https://dongmanhuayuan.myheartsite.com/api/acg/search", "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次获取动漫花园搜索结果时出现错误,error:", err)
			continue
		}
		defer post.Body.Close()
		all, err = ioutil.ReadAll(post.Body)
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次解析动漫花园搜索结果时出现错误,error:", err)
			continue
		}
		err = json.Unmarshal(all, &res)
		if err != nil {
			myUtil.ErrLog.Println("第", i+1, "次转换动漫花园搜索结果为数据类型时出现错误,error:", err)
			continue
		}
		break
	}
	if err != nil {
		myUtil.ErrLog.Println("获取动漫花园数据时出现错误,error:", err)
		return nil, err
	}
	ret := dmhyInfoHandler(res)
	println("动漫花园搜索结束！")
	return ret, nil
}
