package note

import (
	"html/template"
	"regexp"
)

func Analyse(notes2 notes) []template.HTML {
	var ret []template.HTML
	compile := regexp.MustCompile("\\[CQ:image[^\\]]*\\]")
	urlComp := regexp.MustCompile("(http|https)://([\\w-]+\\.)+[\\w-]+(/[\\w-./?%&=]*)?")
	for _, note := range notes2.Note {
		ret = append(ret, template.HTML("<div class=\"time\">"+note[11:22]+"</div>"))
		note = note[24:]
		pre := 0
		arr := compile.FindAllStringIndex(note, -1)
		for _, j := range arr {
			if j[0] != pre {
				ret = append(ret, template.HTML("<div>"+note[pre:j[0]]+"</div>"))
			}
			ret = append(ret, template.HTML("<div><img src=\""+urlComp.FindString(note[j[0]:j[1]])+"\"></div>"))
			pre = j[1]
		}
		if pre != len(note) {
			ret = append(ret, template.HTML("<div>"+note[pre:]+"</div>"))
		}
	}
	return ret
}
