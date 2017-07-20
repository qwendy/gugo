package generator

import (
	"html/template"
	"testing"
)

func TestDisPlayGenerate(t *testing.T) {
	indexTpl, _ := template.ParseFiles("../themes/微/index.html")
	postTpl, _ := template.ParseFiles("../themes/微/post.html")
	tagTpl, _ := template.ParseFiles("../themes/微/tag.html")
	// categoryTpl, _ := template.ParseFiles("../themes/微/category.html")
	listTpl, _ := template.ParseFiles("../themes/微/list.html")
	des := "../public"
	hp := NewHomePage("../source/_post", des, indexTpl, postTpl, 5, 5)
	if err := hp.BatchHandle(); err != nil {
		t.Error(err)
	}
	tags := NewTags(hp.Posts, tagTpl, listTpl, des)
	tags.SetInfos()
	if err := tags.Generate(); err != nil {
		t.Error(err)
	}
	category := NewCategory(hp.Posts, tagTpl, listTpl, des)
	category.SetInfos()
	if err := category.Generate(); err != nil {
		t.Error(err)
	}
}
