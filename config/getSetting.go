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

type OpenAi struct {
	Token string `yaml:"token"`
}

type Setting struct {
	Mode       MsgFilterMode   `yaml:"msgFilterMode"`
	DataSource DataSource      `yaml:"dataSource"`
	Func       Function        `yaml:"function"`
	Auth       Auth            `yaml:"auth"`
	Logs       Logs            `yaml:"logs"`
	Server     Server          `yaml:"server"`
	AiPaint    AiPaint         `yaml:"aiPaint"`
	Setu       Setu            `yaml:"setu"`
	BotName    BotName         `yaml:"botName"`
	Bangumi    BangumiSettings `yaml:"bangumi"`
	OpenAi     OpenAi          `yaml:"openAi"`
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
}

var Settings = &Setting{}

func GetSetting() {
	file, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println("Error reading config file: ", err)
	}
	yaml.Unmarshal(file, &Settings)
}
