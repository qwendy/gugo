package generator

import (
	"html/template"
	"testing"
)

func TestNewPost(t *testing.T) {
	tpl, _ := template.ParseFiles("../template/post.html")
	type args struct {
		sourcePath  string
		destination string
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
			if err := post.ParseMetaData(); err != nil {
				t.Error(err)
			}
			if err := post.Convert(); err != nil {
				t.Error(err)
			}

			if err := post.Generate(); err != nil {
				t.Error(err)
			}
			t.Errorf("Meta:%+v", post.Meta)
		})
	}
}
