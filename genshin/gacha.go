package genshin

import (
	"Lealra/myUtil"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type GachaRes struct {
	Status int    `json:"status"`
	Base64 string `json:"base64,omitempty"`
	Err    string `json:"err,omitempty"`
}

func GachaHandler(uid int64, uname string, ml []string, mflen int) (string, error) {
	if mflen > 0 && mflen <= 3 && len(ml[0]) >= 6 && ml[0][:6] == "原神" {
		switch mflen {
		case 1:
			switch ml[0][6:] {
			case "十连":
				res, err := getGachaRes(uid, uname, "角色1", "1", false)
				return res, err
			case "648":
				res, err := getGachaRes(uid, uname, "角色1", "8", false)
				return res, err
			case "十连展示":
				res, err := getGachaRes(uid, uname, "角色1", "1", true)
				return res, err
			case "648展示":
				res, err := getGachaRes(uid, uname, "角色1", "8", true)
				return res, err
			case "仓库重置":
				res, err := resetGacha(uid)
				return res, err
			}
		case 2:
			switch ml[0][6:] {
			case "十连":
				res, err := getGachaRes(uid, uname, ml[1], "1", false)
				return res, err
			case "648":
				res, err := getGachaRes(uid, uname, ml[1], "8", false)
				return res, err
			case "十连展示":
				res, err := getGachaRes(uid, uname, ml[1], "1", true)
				return res, err
			case "648展示":
				res, err := getGachaRes(uid, uname, ml[1], "8", true)
				return res, err
			case "定轨":
				if ml[1] == "查看" {
					res, err := getTargetWeapon()
					return res, err
				} else {
					res, err := setTargetWeapon(uid, ml[1])
					return res, err
				}
			}
		}
	}
	return "", nil
}

func resetGacha(uid int64) (string, error) {
	get, err := http.Get("http://127.0.0.1:5000/ResetUserInfo?uid=" + strconv.FormatInt(uid, 10))
	if err != nil {
		myUtil.ErrLog.Println("重置原神模拟抽卡结果时出现错误,error:", err)
		return "遭到了天理的攻击！", err
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var bd GachaRes
	err = json.Unmarshal(body, &bd)
	if err != nil {
		myUtil.ErrLog.Println("解析重置原神模拟抽卡结果响应时出现错误,error:", err)
		return "*丘丘语粗口*", err
	}
	switch bd.Status {
	case 1:
		return "重置成功！", nil
	case 2:
		return "找不到你的记录呢，看来戒赌成功了~", nil
	case 0:
		myUtil.ErrLog.Println("原神模拟抽卡服务器出现错误,error:", bd.Err)
		return "重置失败，请用🔨敲打旅行者", errors.New(bd.Err)
	}
	return "", nil
}

func getGachaRes(uid int64, uname string, pool string, num string, display bool) (string, error) {
	displayParam := ""
	if display {
		displayParam = "&display=True"
	}
	get, err := http.Get("http://127.0.0.1:5000/PrayTen?uid=" + strconv.FormatInt(uid, 10) + "&uname=" + url.QueryEscape(uname) + "&pool=" + url.QueryEscape(pool) + "&num=" + url.QueryEscape(num) + displayParam)
	if err != nil {
		myUtil.ErrLog.Println("获取原神模拟抽卡结果时出现错误,error:", err)
		return "遭到了天理的攻击！", err
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var bd GachaRes
	err = json.Unmarshal(body, &bd)
	if err != nil {
		myUtil.ErrLog.Println("解析原神模拟抽卡结果时出现错误,error:", err)
		return "原石打水漂啦~", err
	}
	switch bd.Status {
	case 1:
		return "[CQ:at,qq=" + strconv.FormatInt(uid, 10) + "]\n[CQ:image,file=base64://" + bd.Base64 + "]", nil
	case 2:
		return "没有这个祈愿池呢~", nil
	case 0:
		myUtil.ErrLog.Println("原神模拟抽卡服务器出现错误,error:", bd.Err)
		return "祈愿失败，请用🔨敲打旅行者", errors.New(bd.Err)
	}
	return "", nil
}

type WeaponInfo struct {
	ItemImg    string `json:"item_img"`
	ItemName   string `json:"item_name"`
	WeaponType string `json:"weapon_type"`
}

func getTargetWeapon() (string, error) {
	get, err := http.Get("http://127.0.0.1:5000/GetTargetWeapon")
	if err != nil {
		myUtil.ErrLog.Println("获取原神本期定轨武器时出现错误,error:", err)
		return "遭到了天理的攻击！", err
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var bd []WeaponInfo
	err = json.Unmarshal(body, &bd)
	if err != nil {
		myUtil.ErrLog.Println("解析原神本期定轨武器时出现错误,error:", err)
		return "*丘丘语粗口*", err
	}
	if len(bd) == 0 {
		return "没有获取到有关这期池子的定轨武器的信息！", errors.New("接收到原神本期定轨武器列表为空")
	} else {
		msg := "本期武器池的可定轨武器有~"
		for _, m := range bd {
			msg += "\n\n" + "武器名称：" + m.ItemName + "\n[CQ:image,file=" + m.ItemImg + "]\n" + "武器类型：" + m.WeaponType
		}
		return msg, nil
	}
}

func setTargetWeapon(uid int64, weapon string) (string, error) {
	get, err := http.Get("http://127.0.0.1:5000/SetTargetWeapon?uid=" + strconv.FormatInt(uid, 10) + "&weapon=" + url.QueryEscape(weapon))
	if err != nil {
		myUtil.ErrLog.Println("设定定轨武器时出现错误,error:", err)
		return "遭到了天理的攻击！", err
	}
	defer get.Body.Close()
	body, err := ioutil.ReadAll(get.Body)
	var bd GachaRes
	err = json.Unmarshal(body, &bd)
	if err != nil {
		myUtil.ErrLog.Println("解析武器定轨响应时出现错误,error:", err)
		return "*丘丘语粗口*", err
	}
	switch bd.Status {
	case 1:
		return "[CQ:at,qq=" + strconv.FormatInt(uid, 10) + "] 定轨成功！", nil
	case 3:
		myUtil.ErrLog.Println("原神模拟抽卡服务器出现错误,error:", bd.Err)
		return "这期池子没有这个武器的up哦~", errors.New(bd.Err)
	case 4:
		myUtil.ErrLog.Println("原神模拟抽卡服务器出现错误,error:", bd.Err)
		return "不要重复定轨哦~", errors.New(bd.Err)
	}
	return "", nil
}
