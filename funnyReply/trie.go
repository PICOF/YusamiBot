package funnyReply

import (
	"encoding/json"
	"os"
	"strings"
)

var AnalyzeTrie *wordTrie

type wordTrie struct {
	Count    int                `json:"count"`
	NextNode map[rune]*wordTrie `json:"next_node"`
}

func readWord(path string) ([]string, error) {
	var words []string
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		w, err := os.ReadFile(path + "/" + file.Name())
		if err != nil {
			return nil, err
		}
		words = append(words, strings.Fields(string(w))...)
	}
	return words, nil
}

func getTrie() (*wordTrie, error) {
	words, err := readWord("./textGenerate/word")
	if err != nil {
		return nil, err
	}
	trie := &wordTrie{
		Count:    0,
		NextNode: make(map[rune]*wordTrie),
	}
	for _, v := range words {
		insert(trie, []rune(v))
	}
	return trie, nil
}
func insert(trie *wordTrie, word []rune) {
	var ok bool
	for i, v := range word {
		_, ok = trie.NextNode[v]
		if !ok {
			trie.NextNode[v] = &wordTrie{
				Count:    0,
				NextNode: make(map[rune]*wordTrie),
			}
		}
		trie = trie.NextNode[v]
		if i == len(word)-1 {
			trie.Count++
		}
	}
}

func find(trie *wordTrie, msg []rune) int {
	for i, v := range msg {
		if _, ok := trie.NextNode[v]; ok {
			trie = trie.NextNode[v]
		} else {
			return len(msg[:i])
		}
	}
	return len(msg)
}

func writeTrie(trie *wordTrie) error {
	marshal, err := json.Marshal(trie)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("textGenerate/word.json", os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
