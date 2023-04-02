package myUtil

import (
	"Lealra/config"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var MsgLog, ErrLog *log.Logger

func SetLogger() {
	mf, err := os.OpenFile("log/diaryofyusami/"+time.Now().Format("2006-01-02T3PM")+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("日志创建失败！error:", err)
	}
	MsgLog = log.New(io.MultiWriter(mf, os.Stdout), "", log.LstdFlags)
	ef, err := os.OpenFile("log/error/"+time.Now().Format("2006-01-02")+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("日志创建失败！error:", err)
	}
	ErrLog = log.New(io.MultiWriter(ef, os.Stdout), "[Error]", log.Lshortfile|log.LstdFlags)
}
func cleanLogs() {
	cleanMsgLogs()
	cleanErrLogs()
}

func cleanMsgLogs() {
	date := time.Now().Add(-config.Settings.Logs.MsgLogsRefreshCycle * time.Hour).Format("2006-01-02T3PM")
	root := "./log/diaryofyusami"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if info.Name()[:13] == date {
				err := os.Remove(path)
				if err != nil {
					ErrLog.Println("刪除信息日志時出現問題！Error:", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		ErrLog.Println("清理信息日志文件時出現問題！Error:", err)
	}
}
func cleanErrLogs() {
	date := time.Now().AddDate(0, 0, -config.Settings.Logs.ErrLogsRefreshCycle).Format("2006-01-02")
	root := "./log/error"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if info.Name()[:10] == date {
				err := os.Remove(path)
				if err != nil {
					ErrLog.Println("刪除錯誤日志時出現問題！Error:", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		ErrLog.Println("清理錯誤日志文件時出現問題！Error:", err)
	}
}
func RenewLoggers() {
	for true {
		time.Sleep(time.Second * time.Duration(60*(60-time.Now().Minute())-time.Now().Second()))
		SetLogger()
		go cleanLogs()
	}
}
