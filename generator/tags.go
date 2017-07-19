package generator

import (
	"html/template"
)

type Tags struct {
	Posts        []Post
	Infos        map[string][]Post
	Template     *template.Template
	ListTemplate *template.Template
	Destination  string
}

type TagsTemplateData struct {
	Infos map[string][]Post
}

func NewTags(p []Post, t, listTpl *template.Template, des string) *Tags {
	return &Tags{
		Infos:        make(map[string][]Post),
		Posts:        p,
		Template:     t,
		ListTemplate: listTpl,
		Destination:  des,
	}
}

func (t *Tags) SetInfos() {
	for _, p := range t.Posts {
		for _, tag := range p.Meta.Tags {
			t.Infos[tag] = append(t.Infos[tag], p)
		}
	}
}

// Generate the index.html file of tag directory
// and index.html files of kinds of tags directories
func (t *Tags) Generate() error {
	data := TagsTemplateData{
		Infos: t.Infos,
	}
	dir := t.Destination + "/tags"
	err := GenerateIndexFile(t.Template, data, dir)
	if err != nil {
		return err
	}
	for name, info := range t.Infos {
		d := dir + "/" + name
		if err := GenerateIndexFile(t.ListTemplate, struct {
			Info []Post
		}{Info: info}, d); err != nil {
			return err
		}
	}
	return nil
}
