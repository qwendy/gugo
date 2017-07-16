package generator

import (
	"html/template"
	"testing"
)

func TestNewHomePage(t *testing.T) {
	indexTpl, _ := template.ParseFiles("../themes/微/index.html")
	postTpl, _ := template.ParseFiles("../themes/微/post.html")
	hp := NewHomePage("../source/_post", "../public", indexTpl, postTpl, 5, 5)
	if err := hp.BatchHandle(); err != nil {
		t.Error(err)
	}
}
