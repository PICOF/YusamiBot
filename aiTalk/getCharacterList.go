package aiTalk

import (
	"Lealra/myUtil"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Character struct {
	BotName   string `yaml:"botName"`
	BotId     string `yaml:"botId"`
	Greetings string `yaml:"greetings"`
}

var CharList []Character

func GetCharacterList() bool {
	file, err := ioutil.ReadFile("./characterList.yaml")
	if err != nil {
		myUtil.ErrLog.Println("Error reading character list file: ", err)
		return false
	}
	yaml.Unmarshal(file, &CharList)
	return true
}
