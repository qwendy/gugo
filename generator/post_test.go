package generator

import (
	"html/template"
	"testing"
	"time"
)

func TestNewPost(t *testing.T) {
	tpl, _ := template.ParseFiles("../template/post.html")
	type args struct {
		sourcePath  string
		destination string
		m           *Meta
		t           *template.Template
	}
	tests := []struct {
		name string
		args args
		want *Post
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				sourcePath:  "../source/_post/test.md",
				destination: "../public/test.html",
				m: &Meta{
					Title: "清风江上游",
					Date:  time.Now().Format("2006-01-02 15:04:05"),
				},
				t: tpl,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := NewPost(tt.args.sourcePath, tt.args.destination, tt.args.m, tt.args.t)
			if err := post.Fetch(); err != nil {
				t.Error(err)
			}
			if err := post.Convert(); err != nil {
				t.Error(err)
			}
			if err := post.Generate(); err != nil {
				t.Error(err)
			}
		})
	}
}
