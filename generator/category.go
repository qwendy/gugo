package generator

import (
	"html/template"
	"sort"
)

type Category struct {
	Posts        []Post
	Infos        map[string][]Post
	Template     *template.Template
	ListTemplate *template.Template
	Destination  string
}

type CategoryTemplateData struct {
	Kind  string
	Infos map[string][]Post
}

func NewCategory(p []Post, t, listTpl *template.Template, des string) *Category {
	return &Category{
		Infos:        make(map[string][]Post),
		Posts:        p,
		Template:     t,
		ListTemplate: listTpl,
		Destination:  des,
	}
}

func (c *Category) SetInfos() {
	for _, p := range c.Posts {
		for _, name := range p.Meta.Category {
			c.Infos[name] = append(c.Infos[name], p)
		}
	}
}

// Generate the index.html file of tag directory
// and index.html files of kinds of tags directories
func (c *Category) Generate() error {
	data := CategoryTemplateData{
		Kind:  "category",
		Infos: c.Infos,
	}
	dir := c.Destination + "/category"
	err := GenerateIndexFile(c.Template, data, dir)
	if err != nil {
		return err
	}
	for name, info := range c.Infos {
		sort.Slice(info, func(i, j int) bool {
			if info[i].Meta.Date > info[j].Meta.Date {
				return true
			}
			return false
		})
		d := dir + "/" + name
		if err := GenerateIndexFile(c.ListTemplate, struct {
			Kind  string
			Name  string
			Infos []Post
		}{Kind: "category", Name: name, Infos: info}, d); err != nil {
			return err
		}
	}
	return nil
}
