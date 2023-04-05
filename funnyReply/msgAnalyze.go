package funnyReply

import (
	"Lealra/data"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"context"
	"github.com/dlclark/regexp2"
	"github.com/go-ego/gse"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strconv"
	"time"
)

var maxTimes int
var tokenizer gse.Segmenter

type mappingRecord struct {
	Tag       string `bson:"tag"`
	PreText   string `bson:"pre_text"`
	NextText  string `bson:"next_text"`
	Frequency int    `bson:"frequency"`
	GroupID   int64  `bson:"group_id"`
	UserID    string `bson:"user_id"`
	LastTime  int64  `bson:"last_time"`
}

func getAllUid(list []string) []string {
	for i, uid := range list {
		list[i] = uid[10 : len(uid)-1]
	}
	return list
}

func ReplyHandler(mjson returnStruct.Message) (string, error) {
	var tag string
	var compile *regexp.Regexp
	var listenFlag bool
	match, err := regexp2.MustCompile("(?<=说点).*(?=道理)", 0).FindStringMatch(mjson.RawMessage)
	if err != nil {
		return "", err
	}
	if match != nil {
		tag = match.String()
		compile = regexp.MustCompile("\\[CQ:at,qq=[0-9]+\\]")
		uidList := getAllUid(compile.FindAllString(mjson.RawMessage, -1))
		compile = regexp.MustCompile("\\[CQ:.*\\]")
		pureText := compile.ReplaceAllString(mjson.RawMessage, "")
		compile = regexp.MustCompile("说点.*道理\\s*")
		pureText = compile.ReplaceAllString(pureText, "")
		gen, err := generateText(uidList, mjson.GroupID, tag, pureText)
		if err != nil {
			myUtil.ErrLog.Println("文本生成时失败！tag: ", tag, "text: ", pureText)
		}
		if len(gen) == 0 {
			return "说不出道理", nil
		} else {
			return gen, nil
		}
	}
	match, err = regexp2.MustCompile("(?<=听点).*(?=道理)", 0).FindStringMatch(mjson.RawMessage)
	if err != nil {
		return "", err
	}
	if match != nil {
		tag = match.String()
		listenFlag = true
		compile = regexp.MustCompile("听点.*道理\\s*")
		mjson.RawMessage = compile.ReplaceAllString(mjson.RawMessage, "")
	}
	compile = regexp.MustCompile("\\[CQ:.*\\]")
	analyzeText(compile.ReplaceAllString(mjson.RawMessage, ""), mjson.GroupID, strconv.FormatInt(mjson.UserID, 10), tag)
	if listenFlag {
		return "啊——" + tag + "说的道理~", nil
	} else {
		return "", nil
	}
}

func analyzeText(text string, groupId int64, uid string, tag string) {
	result := tokenizer.Cut(text, true)
	for i, _ := range result {
		if i == len(result)-1 {
			break
		}
		err := storeAnalysisResult(groupId, uid, tag, result[i], result[i+1])
		if err != nil {
			myUtil.ErrLog.Println("存入文本解析结果时出错！tag: ", tag, "pre_text: ", result[i], "next_text: ", result[i+1])
			continue
		}
	}
}

func myAnalyzeText(text []rune, groupId int64, uid string, tag string) {
	var length int
	var preText string
	var nextText string
	index := 0
	for {
		length = find(AnalyzeTrie, text[index:])
		if length == 0 {
			length = 1
		}
		preText = string(text[index : index+length])
		if index+length >= len(text) {
			break
		} else {
			index += length
		}
		length = find(AnalyzeTrie, text[index:])
		if length == 0 {
			length = 1
		}
		nextText = string(text[index : index+length])
		err := storeAnalysisResult(groupId, uid, tag, preText, nextText)
		if err != nil {
			myUtil.ErrLog.Println("存入文本解析结果时出错！tag: ", tag, "pre_text: ", preText, "next_text: ", nextText)
			continue
		}
	}
}

func storeAnalysisResult(groupId int64, uid string, tag string, preText string, nextText string) error {
	var filter bson.D
	var update bson.D
	if tag != "" {
		groupId = 0
		uid = ""
	}
	opt := options.Update().SetUpsert(true)
	c := data.Db.Collection("msgAnalysisResults")
	filter = bson.D{{"group_id", groupId}, {"user_id", uid}, {"tag", tag}, {"pre_text", preText}, {"next_text", nextText}}
	update = bson.D{{"$inc", bson.D{{"frequency", 1}}}, {"$setOnInsert", bson.D{{"group_id", groupId}, {"user_id", uid}, {"tag", tag}, {"pre_text", preText}, {"next_text", nextText}}}, {"$set", bson.D{{"last_time", time.Now().Unix()}}}}
	_, err := c.UpdateOne(context.TODO(), filter, update, opt)
	if err != nil {
		return err
	}
	return nil
}
