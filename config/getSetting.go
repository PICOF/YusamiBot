package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

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
}

type OpenAi struct {
	Token   string        `yaml:"token"`
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
	Music            Music            `yaml:"music"`
	AntiCf           AntiCf           `yaml:"antiCf"`
	Proxy            Proxy            `yaml:"proxy"`
}

type MsgFilterMode struct {
	GroupFilterMode   int `yaml:"groupFilterMode"`
	PrivateFilterMode int `yaml:"privateFilterMode"`
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
	Api   string   `yaml:"api"`
	Token []string `yaml:"token"`
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
var BilibiliStatusChan = make(chan bool)
var PicRenewChan = make(chan bool)

func GetSetting() {
	file, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println("Error reading config file: ", err)
	}
	yaml.Unmarshal(file, &Settings)
	if Settings.Bilibili.Status {
		select {
		case BilibiliStatusChan <- true:
		default:
		}
	}
	if Settings.LearnAndResponse.RenewSwitch {
		select {
		case PicRenewChan <- true:
		default:
		}
	}
}
