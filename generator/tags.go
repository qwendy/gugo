package generator

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
)

type Tags struct {
	Posts       []Post
	Infos       map[string][]Post
	Template    *template.Template
	Destination string
}

type TagsTemplateData struct {
	Infos map[string][]Post
}

func NewTags(p []Post, t *template.Template, des string) *Tags {
	return &Tags{
		Infos:       make(map[string][]Post),
		Posts:       p,
		Template:    t,
		Destination: des,
	}
}

func (t *Tags) SetInfos() {
	for _, p := range t.Posts {
		for _, tag := range p.Meta.Tags {
			t.Infos[tag] = append(t.Infos[tag], p)
		}
	}
}

func (t *Tags) Generate() error {
	dir := t.Destination + "/tags"
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("Create directory error:%v", err)
	}
	f, err := os.Create(dir + "/index.html")
	if err != nil {
		return fmt.Errorf("Creating file %s Err:%v", dir, err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	data := TagsTemplateData{
		Infos: t.Infos,
	}
	if err := t.Template.Execute(writer, data); err != nil {
		return fmt.Errorf("Executing template Error: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Writing file Err: %v", err)
	}
	return nil
}
