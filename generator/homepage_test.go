package generator

import (
	"html/template"
	"testing"
)

func TestNewHomePage(t *testing.T) {
	indexTpl, _ := template.ParseFiles("../template/index.html")
	postTpl, _ := template.ParseFiles("../template/post.html")
	hp := NewHomePage("../source/_post", "../public", indexTpl, postTpl, 5, 5)
	if err := hp.BatchHandle(); err != nil {
		t.Error(err)
	}
}
