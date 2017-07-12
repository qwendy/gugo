package generator

import (
	"fmt"
	"html/template"
	"math"
)

const (
	PAGE_DIR = "page"
)

type HomePage struct {
	Posts            []Post
	Destination      string
	Template         *template.Template
	NumPerPage       int
	PageNumShowCount int
}
type HomePageTemplateData struct {
	FirstUrl    string
	LastUrl     string
	CurPage     int
	AllPageNum  int
	PrePageUrl  string
	NextPageUrl string
	Posts       []Post
	PageData    []PageData
}
type PageData struct {
	Num int
	Url string
}

func NewHomePage(posts []Post, des string, t *template.Template, numPerPage, pageNumShowCount int) *HomePage {
	return &HomePage{
		Posts:            posts,
		Destination:      des,
		Template:         t,
		NumPerPage:       numPerPage,
		PageNumShowCount: pageNumShowCount,
	}
}

// separate posts into several pages
func (hp *HomePage) GeneratePages() {
	allPageNum := int(math.Ceil(float64(len(hp.Posts)) / float64(hp.NumPerPage)))
	for i := 1; i <= allPageNum; i++ {
		tplData := hp.getPageData(i, allPageNum)
		hp.writeToHtml(tplData)
	}
}

func (hp *HomePage) getPageData(page, allPageNum int) (tplData HomePageTemplateData) {
	lastUrl := fmt.Sprintf("page/%v", allPageNum)
	max := hp.PageNumShowCount
	prePageUrl := ""
	nextPageUrl := ""
	if page != 1 {
		prePageUrl = fmt.Sprintf("%v/%v", PAGE_DIR, page-1)
		nextPageUrl = fmt.Sprintf("%v/%v", PAGE_DIR, page+1)
	}
	if page == allPageNum {
		max = len(hp.Posts) - 1
		nextPageUrl = fmt.Sprintf("page/%v", allPageNum)
	}
	p := hp.Posts[page:max]
	pds := []PageData{}
	for index := page; index-page < hp.PageNumShowCount && index <= allPageNum; index++ {
		pds = append(pds, PageData{
			Num: index,
			Url: fmt.Sprintf("page/%v", index),
		})
	}
	tplData = HomePageTemplateData{
		FirstUrl:    "",
		LastUrl:     lastUrl,
		CurPage:     page,
		AllPageNum:  allPageNum,
		Posts:       p,
		PrePageUrl:  prePageUrl,
		NextPageUrl: nextPageUrl,
		PageData:    pds,
	}
	return
}

func (hp *HomePage) writeToHtml(tplData HomePageTemplateData) {

}
