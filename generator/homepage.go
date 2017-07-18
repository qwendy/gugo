package generator

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const (
	PAGE_DIR = "page"
)

type HomePage struct {
	Posts            []Post
	SourceDir        string
	Destination      string
	IndexTemplate    *template.Template
	PostTemplate     *template.Template
	NumPerPage       int
	PageNumShowCount int
}
type HomePageTemplateData struct {
	FirstUrl    string
	LastUrl     string
	CurPage     int
	AllPageNum  int
	PrePageUrl  string
	CurPageUrl  string
	NextPageUrl string
	Posts       []Post
	PageData    []PageData
}
type PageData struct {
	Num int
	Url string
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// f := &log.TextFormatter{}
	// f.QuoteEmptyFields = true
	f := &prefixed.TextFormatter{}
	log.SetFormatter(f)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}
func NewHomePage(source, des string, indexTpl, postTpl *template.Template, numPerPage, pageNumShowCount int) *HomePage {
	return &HomePage{
		SourceDir:        source,
		Destination:      des,
		IndexTemplate:    indexTpl,
		PostTemplate:     postTpl,
		NumPerPage:       numPerPage,
		PageNumShowCount: pageNumShowCount,
	}
}
func (hp *HomePage) BatchHandle() error {
	if err := hp.GetPosts(); err != nil {
		return err
	}
	hp.GeneratePages()
	return nil
}
func (hp *HomePage) GetPosts() error {
	dir, err := ioutil.ReadDir(hp.SourceDir)
	if err != nil {
		return err
	}
	for i, file := range dir {
		log.Infof("%v - Start parse %v", i, file.Name())
		p := NewPost(hp.SourceDir+"/"+file.Name(), hp.Destination, hp.PostTemplate)
		if err := p.BatchHandle(); err != nil {
			log.Errorf("Handle %v error:%v", file.Name(), err)
		}
		hp.Posts = append(hp.Posts, *p)
	}
	hp.sortPosts()
	return nil
}

// separate posts into several pages
func (hp *HomePage) GeneratePages() {
	allPageNum := int(math.Ceil(float64(len(hp.Posts)) / float64(hp.NumPerPage)))
	for i := 1; i <= allPageNum; i++ {
		log.Infof("Start generating page %v", i)
		tplData := hp.getPageData(i, allPageNum)
		if err := hp.writeToHtml(tplData); err != nil {
			log.Errorf("Generating page %v failed", err)
		}
	}
}

func (hp *HomePage) sortPosts() {
	sort.Slice(hp.Posts, func(i, j int) bool {
		if hp.Posts[i].Meta.Date > hp.Posts[j].Meta.Date {
			return true
		}
		return false
	})
}

func (hp *HomePage) getPageData(page, allPageNum int) (tplData HomePageTemplateData) {
	lastUrl := fmt.Sprintf("page/%v", allPageNum)
	min := 0
	max := hp.NumPerPage * page
	prePageUrl := "/"
	nextPageUrl := fmt.Sprintf("/%v/%v", PAGE_DIR, page+1)
	curPageUrl := "/"
	if page != 1 {
		min = hp.NumPerPage * (page - 1)
		prePageUrl = fmt.Sprintf("/%v/%v", PAGE_DIR, page-1)
		curPageUrl = fmt.Sprintf("/%v/%v", PAGE_DIR, page)
	}
	if page == 2 {
		prePageUrl = "/"
	}
	if page == allPageNum {
		max = len(hp.Posts)
		nextPageUrl = fmt.Sprintf("/%v/%v", PAGE_DIR, allPageNum)
	}
	p := hp.Posts[min:max]
	log.Infof("page:%v,post:%v-%v", page, min, max)
	for _, xxx := range p {
		log.Infoln(xxx.Meta)
	}
	pds := []PageData{}
	for index := page; index-page < hp.PageNumShowCount && index <= allPageNum; index++ {
		pds = append(pds, PageData{
			Num: index,
			Url: fmt.Sprintf("/page/%v", index),
		})
	}
	tplData = HomePageTemplateData{
		FirstUrl:    "",
		LastUrl:     lastUrl,
		CurPage:     page,
		AllPageNum:  allPageNum,
		PrePageUrl:  prePageUrl,
		CurPageUrl:  curPageUrl,
		NextPageUrl: nextPageUrl,
		PageData:    pds,
		Posts:       p,
	}
	return
}

func (hp *HomePage) writeToHtml(tplData HomePageTemplateData) error {
	if err := os.MkdirAll(hp.Destination+"/"+tplData.CurPageUrl, 0777); err != nil {
		return fmt.Errorf("Create directory error:%v", err)
	}
	f, err := os.Create(hp.Destination + "/" + tplData.CurPageUrl + "/index.html")
	if err != nil {
		return fmt.Errorf("Creating file %s Err: %v", tplData.CurPageUrl, err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	if err := hp.IndexTemplate.Execute(writer, tplData); err != nil {
		return fmt.Errorf("Executing template Error: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Writing file Err: %v", err)
	}
	return nil
}
