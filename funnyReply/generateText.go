package funnyReply

import (
	"Lealra/data"
	"Lealra/myUtil"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	"strings"
	"time"
)

var terminator = []string{"!", "?", ".", "！", "？", "~", "。"}

type nextT struct {
	frequencyMap map[string]int
}

type randomNode struct {
	weight int
	value  string
}

func randomSelect(options *map[string]int) (res string) {
	var arr []randomNode
	sum := 0
	for k, v := range *options {
		arr = append(arr, randomNode{
			weight: sum,
			value:  k,
		})
		sum += v
	}
	if sum == 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(sum)
	for _, node := range arr {
		if randNum >= node.weight {
			res = node.value
		} else {
			break
		}
	}
	(*options)[res]--
	return
}

func getNextWord(uidList []string, groupId int64, tag string, preText string, wordMap *map[string]nextT) (string, error) {
	if options, ok := (*wordMap)[preText]; ok {
		return randomSelect(&options.frequencyMap), nil
	} else {
		newNext := nextT{frequencyMap: make(map[string]int)}
		var userFilter bson.E
		var groupFilter bson.E
		var tagFilter bson.E
		if len(uidList) > 0 {
			userFilter = bson.E{Key: "user_id", Value: bson.D{{"$in", uidList}}}
		}
		if groupId != 0 {
			groupFilter = bson.E{Key: "group_id", Value: groupId}
		}
		if tag != "" {
			userFilter = bson.E{}
			groupFilter = bson.E{}
		}
		tagFilter = bson.E{Key: "tag", Value: tag}
		filter := bson.D{tagFilter, groupFilter, userFilter, {"pre_text", preText}}
		c := data.Db.Collection("msgAnalysisResults")
		cur, err := c.Find(context.TODO(), filter)
		if err != nil {
			return "", err
		}
		for cur.Next(context.TODO()) {
			var elem mappingRecord
			err = cur.Decode(&elem)
			if err != nil {
				myUtil.ErrLog.Println("文字生成记录获取失败，err: ", err)
			}
			newNext.frequencyMap[elem.NextText] = elem.Frequency
		}
		ret := randomSelect(&newNext.frequencyMap)
		(*wordMap)[preText] = newNext
		return ret, nil
	}
}

// db.learnedRespAccurate.aggregate([
// { $match: { resp: "！" } },
// { $match: { text: {$in: ["？","虹夏"]} } },
// { $sample: { size: 1 } }
// ])
func getRandomWord(uidList []string, groupId int64, tag string) (string, error) {
	var userStage bson.E
	var groupStage bson.E
	var tagStage bson.E
	if len(uidList) > 0 {
		userStage = bson.E{Key: "user_id", Value: bson.D{{"$in", uidList}}}
	}
	if groupId != 0 {
		groupStage = bson.E{Key: "group_id", Value: groupId}
	}
	if tag != "" {
		userStage = bson.E{}
		groupStage = bson.E{}
	}
	tagStage = bson.E{Key: "tag", Value: tag}
	punctuation := []string{",", "，", "~", "`", ".", "。", ";", ":", "\\", "·", "{", "}"}
	pipeline := mongo.Pipeline{bson.D{{"$match", bson.D{tagStage, groupStage, userStage, bson.E{Key: "pre_text", Value: bson.D{{"$nin", punctuation}}}}}}, bson.D{{"$sample", bson.D{{"size", 1}}}}}
	c := data.Db.Collection("msgAnalysisResults")
	cur, err := c.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return "", err
	}
	if cur.Next(context.TODO()) {
		var elem mappingRecord
		err = cur.Decode(&elem)
		return elem.PreText, nil
	}
	return "", nil
}
func generateText(uidList []string, groupId int64, tag string, input string) (output string, err error) {
	if len(input) == 0 {
		input, err = getRandomWord(uidList, groupId, tag)
		if err != nil {
			return "", err
		}
	}
	result := tokenizer.Cut(input, true)
	wordMap := make(map[string]nextT)
	var pre string
	var gen string
	var index int
	for len(result) > 0 {
		gen, err = getNextWord(uidList, groupId, tag, result[len(result)-1], &wordMap)
		if err != nil {
			return "", err
		}
		if len(gen) != 0 {
			result = append(result, gen)
			pre = gen
			break
		}
		result = result[:len(result)-1]
	}
	if len(result) > 0 {
		for i := 0; i < maxTimes; i++ {
			gen, err = getNextWord(uidList, groupId, tag, pre, &wordMap)
			if err != nil {
				return "", err
			}
			result = append(result, gen)
			pre = gen
			if in, _ := myUtil.IsInArray(terminator, gen); in {
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(maxTimes) <= i {
					break
				}
			}
		}
	}
	output = strings.Join(result, "")
	for _, v := range terminator {
		k := strings.LastIndex(output, v)
		if k == -1 {
			continue
		}
		i := k + len(v)
		if i > index {
			index = i
		}
	}
	if index != 0 {
		return output[:index], nil
	}
	return output, nil
}

func myGenerateText(uidList []string, groupId int64, tag string, input []rune) (output string, err error) {
	if len(input) == 0 {
		return "", nil
	}
	mid := input
	wordMap := make(map[string]nextT)
	var pre string
	var index int
	var gen string
	var length int
	for len(mid) > 0 {
		length = find(AnalyzeTrie, mid)
		if length == 0 {
			length = 1
		}
		pre = string(mid[len(mid)-length:])
		gen, err = getNextWord(uidList, groupId, tag, pre, &wordMap)
		if err != nil {
			return "", err
		}
		if len(gen) != 0 {
			mid = append(mid, []rune(gen)...)
			pre = gen
			break
		}
		//遇见开头就不能生成的试图倒退出可生成的串
		mid = mid[:len(mid)-1]
	}
	for i := 0; i < maxTimes && len(mid) > 0; i++ {
		gen, err = getNextWord(uidList, groupId, tag, pre, &wordMap)
		if err != nil {
			return "", err
		}
		mid = append(mid, []rune(gen)...)
		pre = gen
		last := mid[len(mid)-1]
		if in, _ := myUtil.IsInArray(terminator, last); in {
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(maxTimes) <= i {
				break
			}
		}
	}
	//没有生成成功，返回原文本
	if len(mid) < len(input) {
		mid = input
	}
	for _, v := range terminator {
		k := strings.LastIndex(string(mid), v)
		if k == -1 {
			continue
		}
		i := k + len(v)
		if i > index {
			index = i
		}
	}
	if index != 0 {
		return string(mid)[:index], nil
	}
	return string(mid), nil
}
