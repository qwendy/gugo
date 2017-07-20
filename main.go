package main

import (
	gugo "gugo/generator"
	"html/template"
	"log"
	"net/http"
)

func main() {
	indexTpl, _ := template.ParseFiles("./themes/微/index.html")
	postTpl, _ := template.ParseFiles("./themes/微/post.html")
	tagTpl, _ := template.ParseFiles("./themes/微/tag.html")
	// categoryTpl, _ := template.ParseFiles("../themes/微/category.html")
	listTpl, _ := template.ParseFiles("./themes/微/list.html")
	des := "./public"
	hp := gugo.NewHomePage("./source/_post", des, indexTpl, postTpl, 5, 5)
	if err := hp.BatchHandle(); err != nil {
		log.Println(err)
		return
	}
	tags := gugo.NewTags(hp.Posts, tagTpl, listTpl, des)
	tags.SetInfos()
	if err := tags.Generate(); err != nil {
		log.Println(err)
		return
	}
	category := gugo.NewCategory(hp.Posts, tagTpl, listTpl, des)
	category.SetInfos()
	if err := category.Generate(); err != nil {
		log.Println(err)
		return
	}
}

func StaticServer() {
	// 前缀去除 ;列出dir
	http.Handle("", http.StripPrefix("/public/", http.FileServer(http.Dir(""))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
