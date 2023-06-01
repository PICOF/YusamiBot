package HZYS

import "testing"

func TestGetTables(t *testing.T) {
	getTables()
}

func TestGenerator(t *testing.T) {
	var g generatorImpl
	g.GetAudioFileList("电棍", []rune("大家好啊，我是丁真"))
	g.GenerateAudio()
}
