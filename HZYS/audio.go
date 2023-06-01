package HZYS

import (
	"Lealra/config"
	"encoding/base64"
	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"
	"github.com/mozillazg/go-pinyin"
	"github.com/pkg/errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var args pinyin.Args
var specialMap = map[string][]string{
	"a": {"ei"},
	"b": {"bi"},
	"c": {"xi"},
	"d": {"di"},
	"e": {"yi"},
	"f": {"ai fu"},
	"g": {"ji"},
	"h": {"ai", "chi"},
	"i": {"ai"},
	"j": {"zhei"},
	"k": {"kei"},
	"l": {"ai", "lu"},
	"m": {"ai,mu"},
	"n": {"en"},
	"o": {"ou"},
	"p": {"pi"},
	"q": {"kiu"},
	"r": {"a"},
	"s": {"ai", "si"},
	"t": {"ti"},
	"u": {"you"},
	"v": {"wei"},
	"w": {"da", "bu", "liu"},
	"x": {"ai", "ke", "si"},
	"y": {"wai"},
	"z": {"zei"},
	"0": {"ling"},
	"1": {"yi"},
	"2": {"er"},
	"3": {"san"},
	"4": {"si"},
	"5": {"wu"},
	"6": {"liu"},
	"7": {"qi"},
	"8": {"ba"},
	"9": {"jiu"},
}

type generatorImpl struct {
	FileList []string
}

type Generator interface {
	GetAudioFileList(voiceName string, sentence []rune)
	GenerateAudio() (string, error)
}

func NewGenerator() Generator {
	return &generatorImpl{}
}

func (g *generatorImpl) GetAudioFileList(voiceName string, sentence []rune) {
	var end int
	var space int
	var filename string
	var fileList []string
	var p []string
	voiceName = strings.Join(pinyin.LazyConvert(voiceName, nil), "")
	path := config.Settings.HZYS.SourceDir + "/" + voiceName + "/"
	for i := 0; i < len(sentence); {
		end, filename = find(path, voiceName, sentence, i)
		if end == i {
			if _, ok := specialMap[strings.ToLower(string(sentence[i]))]; ok {
				p = specialMap[strings.ToLower(string(sentence[i]))]
			} else {
				p = pinyin.SinglePinyin(sentence[i], args)
			}
			if len(p) == 0 {
				space++
				if space > 2 {
					fileList = fileList[:len(fileList)-2]
					if _, ok := meaningless[voiceName]; ok && len(meaningless[voiceName]) > 0 {
						fileList = append(fileList, path+"sentence/"+meaningless[voiceName][rand.Int()%len(meaningless[voiceName])]+".wav")
					}
					space = 0
				} else {
					fileList = append(fileList, config.Settings.HZYS.SourceDir+"/public/space.wav")
				}
			} else {
				space = 0
			}
			for _, v := range p {
				filename = path + "word/" + v + ".wav"
				_, err := os.Stat(filename)
				if err != nil {
					filename = config.Settings.HZYS.SourceDir + "/public/space.wav"
				}
				fileList = append(fileList, filename)
			}
			i++
		} else {
			i = end
			fileList = append(fileList, filename)
			space = 0
		}
	}
	g.FileList = fileList
	return
}

func (g *generatorImpl) GenerateAudio() (string, error) {
	var buffer *beep.Buffer
	for i, s := range g.FileList {
		if s != "" {
			f, err := os.Open(s)
			if err != nil {
				return "", errors.WithMessage(err, "open file failed")
			}
			streamer, format, err := wav.Decode(f)
			if err != nil {
				return "", errors.WithMessage(err, "decode file failed")
			}
			if i == 0 {
				buffer = beep.NewBuffer(format)
			}
			buffer.Append(streamer)
		}
	}
	targetPath := config.Settings.HZYS.SourceDir + "/generate/" + strconv.Itoa(time.Now().Second()) + ".wav"
	f, _ := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	err := wav.Encode(f, buffer.Streamer(0, buffer.Len()), buffer.Format())
	if err != nil {
		return "", errors.WithMessage(err, "write file failed")
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		return "", errors.WithMessage(err, "read file failed")
	}
	return base64.StdEncoding.EncodeToString(content), nil
}
