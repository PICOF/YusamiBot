package hf

import "testing"

var testAction Manipulator
var c Client

func init() {
	c, _ = GetClientWithUrl("https://huggingface.co/spaces/skytnt/moe-tts", true)
	testAction = NewManipulator()
	_, err := testAction.SetClient(c)
	if err != nil {
		println(err.Error())
		return
	}
}

func TestManipulatorImpl_Find(t *testing.T) {
	m := testAction.Find("TTS").Find("model0")
	_, err := m.Find("Generate").Execute(LinkModeWs)
	if err != nil {
		println(err.Error())
		return
	}
	get, err := m.Find("Output Audio").Get()
	if err != nil {
		println(err.Error())
		return
	}
	println(get.Props.Value.(string))
}
