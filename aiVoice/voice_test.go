package aiVoice

import (
	"Lealra/returnStruct"
	"testing"
)

func TestVoiceGenerateHandler(t *testing.T) {
	handler, err := VoiceGenerateHandler(returnStruct.Message{})
	if err != nil {
		return
	}
	println(handler)
}

func TestInit(t *testing.T) {
	println("Initial")
}

func TestMultiLanguageTextHandler(t *testing.T) {
	ret := MultiLanguageTextHandler("你好我是[JA]sas[JA]阿尔托莉雅[JA]sadsad[JA]很想认识你", "ja")
	println(ret)
}
