package funnyReply

import (
	"fmt"
	"github.com/yanyiwu/gojieba"
	"log"
	"os"
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	println(len(string([]rune("啥叫哦"))))
}
func TestGenerate(t *testing.T) {
	input := "摆"
	//file, err := os.ReadFile("../textGenerate/material.txt")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//text := []rune(string(file))
	//analyzeText(text, 0, "", "原神")
	//if err != nil {
	//	log.Fatal(err)
	//}
	gen, _ := generateText(nil, 0, "原神", input)
	println(gen)
}

func TestAnalyze(t *testing.T) {
	file, err := os.ReadFile("textGenerate/material.txt")
	if err != nil {
		log.Fatal(err)
	}
	text := []rune(string(file))
	analyzeText(string(text), 0, "", "原神")
}

func TestTrie(t *testing.T) {
	println(find(AnalyzeTrie, []rune("原h,hello")))
}

func TestCut(t *testing.T) {
	x := gojieba.NewJieba()
	defer x.Free()
	file, err := os.ReadFile("textGenerate/material.txt")
	if err != nil {
		log.Fatal(err)
	}
	s := string(file)
	s = "刻晴说 奥利安费！All in！\n偶内的手，啊？啊？盲僧在为什么我。行不行。上路我给你吗？诶呀米诺？你能说明白就看不惯你告诉我为什么不在上路被打野不在我怎么保得住他么？打野都没有在的臭jb尊尼获加的，打野先越塔我怎么帮，你告诉我说话来！\n诶乌兹！"
	words := x.Cut(s, true)
	text := strings.Join(words, "/")
	fmt.Println("全模式:", text)
}

func TestGse(t *testing.T) {
	// 加载默认词典
	tokenizer.LoadDict("jp")
	println(tokenizer.DictPath)
	s := "刻晴说 奥利安费！All in！\n偶内的手，啊？啊？盲僧在为什么我。行不行。上路我给你吗？诶呀米诺？你能说明白就看不惯你告诉我为什么不在上路被打野不在我怎么保得住他么？打野都没有在的臭jb尊尼获加的，打野先越塔我怎么帮，你告诉我说话来！\n诶乌兹！"
	words := tokenizer.Cut(s)
	text := strings.Join(words, "/")
	fmt.Println(text)
	words = tokenizer.Cut(s, true)
	text = strings.Join(words, "/")
	fmt.Println(text)
}
