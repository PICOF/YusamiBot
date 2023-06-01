package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
	"time"
)

var config *viper.Viper

type Setting struct {
	Mode             MsgFilterMode    `yaml:"msgFilterMode"`
	DataSource       DataSource       `yaml:"dataSource"`
	Func             Function         `yaml:"function"`
	Auth             Auth             `yaml:"auth"`
	Logs             Logs             `yaml:"logs"`
	Server           Server           `yaml:"server"`
	AiPaint          AiPaint          `yaml:"aiPaint"`
	Setu             Setu             `yaml:"setu"`
	BotName          BotName          `yaml:"botName"`
	Bangumi          BangumiSettings  `yaml:"bangumi"`
	OpenAi           OpenAi           `yaml:"openAi"`
	CharacterAi      CharacterAi      `yaml:"characterAi"`
	LearnAndResponse LearnAndResponse `yaml:"learnAndResponse"`
	Bilibili         Bilibili         `yaml:"bilibili"`
	Daoli            Daoli            `yaml:"daoli"`
	HZYS             HZYS             `yaml:"hzys"`
	Music            Music            `yaml:"music"`
	AntiCf           AntiCf           `yaml:"antiCf"`
	Proxy            Proxy            `yaml:"proxy"`
}

type BotName struct {
	Name           string `yaml:"name"`
	FullName       string `yaml:"fullName"`
	LowerCasedName string `yaml:"lowerCasedName"`
	Id             int64  `yaml:"id"`
}

type BangumiSettings struct {
	BangumiSearch bool `yaml:"bangumiSearch"`
	MaxPoolSize   int  `yaml:"maxPoolSize"`
}

type OpenAiSetting struct {
	ChatSetting struct {
		Memory           bool   `yaml:"memory"`
		Model            string `yaml:"model"`
		MaxTokens        string `yaml:"maxTokens"`
		Temperature      string `yaml:"temperature"`
		TopP             string `yaml:"topP"`
		FrequencyPenalty string `yaml:"frequencyPenalty"`
		PresencePenalty  string `yaml:"presencePenalty"`
	} `yaml:"chat"`
	EditSetting struct {
		Model string `yaml:"model"`
	} `yaml:"edit"`
	CommSetting struct {
		Preset string `yaml:"preset"`
		MsgCap int    `yaml:"msgCap"`
		MaxLen int    `yaml:"maxLen"`
	} `yaml:"comm"`
}

type OpenAi struct {
	Token   string        `yaml:"token"`
	Retries int           `yaml:"retries"`
	Setting OpenAiSetting `yaml:"settings"`
}

type Proxy struct {
	HttpsProxy []string `yaml:"httpsProxy"`
}

type LearnAndResponse struct {
	RenewSwitch  bool  `yaml:"renewSwitch"`
	GroupToRenew int64 `yaml:"groupToRenew"`
	MsgInterval  int64 `yaml:"msgInterval"`
	UseBase64    bool  `yaml:"useBase64"`
	Compress     bool  `yaml:"compress"`
}

type AntiCf struct {
	Ja3       string `yaml:"ja3"`
	UserAgent string `yaml:"userAgent"`
	Timeout   int    `yaml:"timeout"`
	Proxy     string `yaml:"proxy"`
}

type Bilibili struct {
	Status   bool    `yaml:"status"`
	Interval float32 `yaml:"interval"`
}

type Music struct {
	Card bool `yaml:"card"`
}

type CharacterAi struct {
	Token   string `yaml:"token"`
	Timeout int    `yaml:"timeout"`
}

type Daoli struct {
	MaxTime    int `yaml:"maxTime"`
	Expiration int `yaml:"expiration"`
}

type HZYS struct {
	SourceDir string `yaml:"sourceDir"`
}

type MsgFilterMode struct {
	GroupFilterMode   int     `yaml:"groupFilterMode"`
	PrivateFilterMode int     `yaml:"privateFilterMode"`
	BanId             []int64 `yaml:"banId"`
}

type DataSource struct {
	Port         string `yaml:"port"`
	Auth         bool   `yaml:"auth"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"databaseName"`
	MaxPoolSize  uint64 `yaml:"maxPoolSize"`
	TimeOut      int64  `yaml:"timeout"`
}

type Server struct {
	Hostname string `yaml:"hostname"`
	Port     string `yaml:"port"`
}

type Function struct {
	Repeat       bool `yaml:"repeat"`
	MakeChoice   bool `yaml:"makeChoice"`
	PrivateDiary bool `yaml:"privateDiary"`
}

type AiPaint struct {
	Free struct {
		Api   string   `yaml:"api"`
		Token []string `yaml:"token"`
	} `yaml:"free"`
}

type Auth struct {
	Admin []int64 `yaml:"admin"`
}

type Logs struct {
	MsgLogsRefreshCycle time.Duration `yaml:"msgLogsRefreshCycle"`
	ErrLogsRefreshCycle int           `yaml:"errLogsRefreshCycle"`
}

type Setu struct {
	SetuGroupSender bool   `yaml:"setuGroupSender"`
	Api             string `yaml:"api"`
	PicMode         int    `yaml:"picMode"`
}

var Settings = &Setting{}
var BilibiliStatusChan = make(chan struct{})
var PicRenewChan = make(chan struct{})

func GetSetting() {
	err := config.Unmarshal(&Settings, func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	})
	if err != nil {
		return
	}
	if err != nil {
		fmt.Println("Error in unmarshal config: ", err)
	}
	if Settings.Bilibili.Status {
		select {
		case BilibiliStatusChan <- struct{}{}:
		default:
		}
	}
	if Settings.LearnAndResponse.RenewSwitch {
		select {
		case PicRenewChan <- struct{}{}:
		default:
		}
	}
}
func GetConfig(configName string) (v *viper.Viper) {
	v = viper.New()
	v.SetConfigName(configName)
	v.AddConfigPath(".")
	v.AddConfigPath("..")
	v.AddConfigPath("./config")
	v.AddConfigPath("../../config")
	v.AddConfigPath("../../../config")
	err := v.ReadInConfig() // 查找并读取配置文件
	if err != nil {         // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("config file changes: ", e.String())
		GetSetting()
	})
	return
}
