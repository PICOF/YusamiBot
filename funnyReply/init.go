package funnyReply

import (
	"Lealra/config"
	"log"
)

func init() {
	maxTimes = config.Settings.Daoli.MaxTime
	var err error
	tokenizer.DictPath = "./tokenizerDict/"
	err = tokenizer.LoadDict()
	if err != nil {
		log.Fatal(err)
	}
	err = tokenizer.LoadDict("jp")
	if err != nil {
		log.Fatal(err)
	}
	//AnalyzeTrie, err = getTrie()
	//if err != nil {
	//	log.Fatal(err)
	//}
}
