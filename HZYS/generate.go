package HZYS

import (
	"Lealra/returnStruct"
	"strings"
)

func Generate(mjson returnStruct.Message) (string, error) {
	sentence := []rune(mjson.RawMessage)
	args := strings.Fields(mjson.RawMessage)
	var g generatorImpl
	if string(sentence[:4]) == "大家好啊" {
		g.GetAudioFileList("电棍", sentence)
	} else if len(args) >= 3 && args[0] == "活字印刷" {
		g.GetAudioFileList(args[1], []rune(strings.Join(args[2:], " ")))
	}
	if g.FileList != nil {
		base64, err := g.GenerateAudio()
		if err != nil {
			return "", err
		}
		return "[CQ:record,file=base64://" + base64 + "]", nil
	}
	return "", nil
}
