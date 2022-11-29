package schoolTimeTable

import (
	"Lealra/myUtil"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type OriginalData struct {
	Weekday     string `json:"xqj"`
	Week        int    `json:"djz"`
	Position    string `json:"room_name"`
	Start       string `json:"ksjc"`
	Finish      string `json:"jsjc"`
	Teacher     string `json:"teacher"`
	Cname       string `json:"c_name"`
	OpeningTime string `json:"rq"`
}

type Class struct {
	Name     string
	Position string
	Start    int
	Finish   int
	Teacher  string
}

func getSchoolTimetable(sid string) ([]OriginalData, error) {
	get, err := http.Get("https://studyapi.uestc.edu.cn/ckd/getWeekClassSchedule?userId=" + sid)
	body, err := io.ReadAll(get.Body)
	get.Body.Close()
	var data []OriginalData
	json.Unmarshal(body, &data)
	return data, err
}

func getClassInfo(sid string, weekday int) ([]Class, error) {
	res, err := getSchoolTimetable(sid)
	var data []Class
	for _, c := range res {
		if i, _ := strconv.Atoi(c.Weekday); i != weekday || c.OpeningTime[c.Week] != '1' {
			continue
		} else {
			s, _ := strconv.Atoi(c.Start)
			f, _ := strconv.Atoi(c.Finish)
			data = append(data, Class{
				Name:     c.Cname,
				Position: c.Position,
				Start:    s,
				Finish:   f,
				Teacher:  c.Teacher,
			})
		}
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Start < data[j].Start
	})
	return data, err
}
func GetClass(uid int64, tomo bool) ([]Class, error) {
	sid, err := GetSid(uid)
	if err != nil {
		myUtil.ErrLog.Println("获取用户学号时出错:", err)
		return nil, err
	}
	if sid == "" {
		myUtil.ErrLog.Println("用户 ", uid, " 未绑定学号")
		return nil, err
	}
	d := int(time.Now().Weekday())
	if tomo {
		d++
	} else {
		if d == 0 {
			d = 7
		}
	}
	return getClassInfo(sid, d)
}
