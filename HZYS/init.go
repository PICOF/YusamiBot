package HZYS

import (
	"Lealra/myUtil"
	"github.com/mozillazg/go-pinyin"
)

func init() {
	args = pinyin.NewArgs()
	err := getTables()
	if err != nil {
		myUtil.ErrLog.Println("Error getting hzys tables, error: ", err)
	}
}
