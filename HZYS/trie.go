package HZYS

import (
	"Lealra/config"
	"encoding/json"
	"math/rand"
	"os"
)

var meaningless = make(map[string][]string)
var trieTrees = make(map[string]*wordTrie)

type wordTrie struct {
	FileName []string
	NextNode map[rune]*wordTrie
}

type table struct {
	Meaningful  map[string][]string `json:"meaningful"`
	Meaningless []string            `json:"meaningless"`
}

func getTrie(t table) *wordTrie {
	trie := &wordTrie{
		NextNode: make(map[rune]*wordTrie),
	}
	for k, v := range t.Meaningful {
		insert(trie, []rune(k), v)
	}
	return trie
}
func insert(trie *wordTrie, word []rune, filename []string) {
	var ok bool
	for i, v := range word {
		_, ok = trie.NextNode[v]
		if !ok {
			trie.NextNode[v] = &wordTrie{
				NextNode: make(map[rune]*wordTrie),
			}
		}
		trie = trie.NextNode[v]
		if i == len(word)-1 {
			trie.FileName = filename
		}
	}
}

func find(path string, voiceName string, msg []rune, index int) (end int, filename string) {
	trie := trieTrees[voiceName]
	if trie == nil {
		return
	}
	end = index
	for _, ok := trie.NextNode[msg[index]]; ok; _, ok = trie.NextNode[msg[index]] {
		trie = trie.NextNode[msg[index]]
		index++
		if trie.FileName != nil {
			filename = path + "sentence/" + trie.FileName[rand.Int()%len(trie.FileName)] + ".wav"
			end = index
		}
		if index >= len(msg) {
			break
		}
	}
	return
}

func getTables() error {
	path := config.Settings.HZYS.SourceDir
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		w, err := os.ReadFile(path + "/" + file.Name() + "/table.json")
		if err != nil {
			return err
		}
		var t table
		err = json.Unmarshal(w, &t)
		if err != nil {
			return err
		}
		meaningless[file.Name()] = t.Meaningless
		trieTrees[file.Name()] = getTrie(t)
	}
	return nil
}
