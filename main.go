package main

import (
	"flag"
	gugo "gugo/generator"
	"html/template"
	"log"
	"net/http"
)

func Generate(sourceDir, des, themeDir string) {
	indexTpl, _ := template.ParseFiles(themeDir + "/index.html")
	postTpl, _ := template.ParseFiles(themeDir + "/post.html")
	tagTpl, _ := template.ParseFiles(themeDir + "/tag.html")
	// categoryTpl, _ := template.ParseFiles("../themes/微/category.html")
	listTpl, _ := template.ParseFiles(themeDir + "/list.html")
	hp := gugo.NewHomePage(sourceDir+"/_post", des, indexTpl, postTpl, 5, 5)
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
	s := gugo.NewStatic(sourceDir, themeDir, des)
	if err := s.BatchHandle(); err != nil {
		log.Println(err)
		return
	}
}

func main() {
	methed := flag.String("m", "g", "run method. use g  to generate site. use s to run site by the local server")
	des := flag.String("des", "./public", "destination directory")
	source := flag.String("source", "./source", "source directory")
	theme := flag.String("theme", "./themes/微", "theme directory")
	port := flag.String("port", "9090", "local serfer`s port")

	flag.Parse()
	switch *methed {
	case "g":
		Generate(*source, *des, *theme)
	case "s":
		StaticServer(*des, *port)
	default:
		log.Println("Sorry! The Method Not Found!")
	}

}

func StaticServer(des string, port string) {
    // 设置静态目录
    http.Handle("/", http.FileServer(http.Dir(des)))
	log.Println("listen: http://127.0.0.1:"+port)
	err := http.ListenAndServe(":"+port, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
