package aiVoice

import (
	"Lealra/myUtil"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

var vitsConfig Config

const configName = "vits"

type Config struct {
	Spaces []Space `json:"spaces"`
	Single []struct {
		Speaker string `json:"speaker"`
		Space   Space  `json:"space"`
	} `json:"single"`
}

func GetConfig() (v *viper.Viper) {
	v = viper.New()
	v.SetConfigName(configName)
	v.AddConfigPath(".")
	v.AddConfigPath("../")
	v.AddConfigPath("./config")
	v.AddConfigPath("../../config")
	v.AddConfigPath("../../../config")
	err := v.ReadInConfig() // 查找并读取配置文件
	if err != nil {         // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = refreshConfig(v)
	if err != nil {
		log.Fatal("vits 配置导入失败！error: ", err)
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		err := refreshConfig(v)
		if err != nil {
			myUtil.ErrLog.Println("vits 配置刷新失败！error: ", err)
		}
		GetVits()
	})
	return
}
func refreshConfig(v *viper.Viper) error {
	err := v.Unmarshal(&vitsConfig)
	return err
}
