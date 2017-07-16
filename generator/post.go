// convert post.md to post.html
package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/henrylee2cn/pholcus/common/goquery"
	"github.com/russross/blackfriday"
	"github.com/sourcegraph/syntaxhighlight"
)

type Post struct {
	SourcePath   string
	sourceData   []byte
	overviewData []byte
	Destination  string // the folder to create directory and html file
	Location     string // the location(directory) of the html file
	htmlData     []byte
	overViewHtml []byte
	Meta         *Meta
	Template     *template.Template
	TemplateData *PostTemplateData
}

type PostTemplateData struct {
	Meta     *Meta
	Overview template.HTML
	Content  template.HTML
}
type Meta struct {
	Title      string
	Date       string
	Tags       []string
	Category   []string
	ParsedDate time.Time
}

func NewPost(sourcePath string, destination string, t *template.Template) *Post {
	return &Post{
		SourcePath:  sourcePath,
		Destination: destination,
		Meta:        new(Meta),
		Template:    t,
	}
}

func (post *Post) BatchHandle() error {
	if err := post.Fetch(); err != nil {
		return err
	}
	if err := post.ParseMetaData(); err != nil {
		return err
	}
	if err := post.Convert(); err != nil {
		return err
	}
	if err := post.CreateDestinationPath(); err != nil {
		return err
	}
	if err := post.Generate(); err != nil {
		return err
	}
	return nil
}

// get source markdown file
func (p *Post) Fetch() error {
	input, err := ioutil.ReadFile(p.SourcePath)
	p.sourceData = input
	return err
}

// parse the metedata from markdown file and remove it from sourcefile
// parse overview data
func (p *Post) ParseMetaData() (err error) {
	buf := bufio.NewReader(bytes.NewReader(p.sourceData))
	metaData := []byte{}
	overview := []byte{}
	content := []byte{}
	metaStart := false
	metaFinished := false
	overviewFinished := false
	// get the data between the line only have "---"
	for true {
		line, isPrefix, lineErr := buf.ReadLine()
		if lineErr != nil {
			break
		}
		for _, s := range []string{"---", "--- ", " --- ", " ---", "----"} {
			if string(line) == s {
				if metaStart == true {
					metaFinished = true
				}
				metaStart = true
			}
		}
		for _, s := range []string{"<!-- more -->", "<!-- more-->", "<!--more -->"} {
			if strings.Contains(string(line), s) {
				overviewFinished = true
			}
		}

		if !isPrefix {
			line = append(line, []byte("\n")...)
		}
		if metaStart && !metaFinished {
			metaData = append(metaData, line...)
		}
		if metaFinished && !overviewFinished {
			overview = append(overview, line...)
		}
		if metaFinished {
			content = append(content, line...)
		}
	}
	if !metaFinished {
		return fmt.Errorf("Cannot find the metadata of the post!")
	}
	err = yaml.Unmarshal(metaData, p.Meta)
	if err != nil {
		return
	}
	if p.Meta.Title == "" {
		return fmt.Errorf("The title of the post is empty!")
	}
	if p.Meta.Date == "" {
		return fmt.Errorf("The Date of the post is empty!")
	}
	p.sourceData = content
	p.overviewData = overview
	return
}

// convert mardown file data to html
func (p *Post) Convert() (err error) {
	p.htmlData = blackfriday.MarkdownCommon(p.sourceData)
	err = p.fixHtmlData(p.htmlData)
	if err != nil {
		return err
	}
	p.overViewHtml = blackfriday.MarkdownCommon(p.overviewData)
	err = p.fixHtmlData(p.overViewHtml)
	if err != nil {
		return err
	}

	return
}

// create the catelog of the post .
// the catelog of the post is the meta.date
func (p *Post) CreateDestinationPath() (err error) {
	d := strings.Split(p.Meta.Date, "-")
	if len(d) < 3 {
		return fmt.Errorf("The date in the metadata of the post is incorrect!")
	}
	d = strings.Split(p.Meta.Date, " ")
	if len(d) < 1 {
		return fmt.Errorf("The date in the metadata of the post is incorrect!")
	}
	p.Location = fmt.Sprintf("%s/%s", strings.Replace(d[0], "-", "/", -1), p.Meta.Title)
	if mdErr := os.MkdirAll(p.Destination+"/"+p.Location, 0777); mdErr != nil {
		return mdErr
	}
	return
}

// write html data to file
func (p *Post) Generate() (err error) {
	f, err := os.Create(p.Destination + "/" + p.Location + "/index.html")
	if err != nil {
		return fmt.Errorf("Creating file %s Err: %v", p.Destination, err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	p.TemplateData = &PostTemplateData{
		Meta:     p.Meta,
		Content:  template.HTML(string(p.htmlData)),
		Overview: template.HTML(string(p.overViewHtml)),
	}
	if err := p.Template.Execute(writer, p.TemplateData); err != nil {
		return fmt.Errorf("Executing template Error: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Writing file Err: %v", err)
	}
	return nil
}

// replace code in html file,use syntaxhighlight code
// remove some html tags such as <head>
func (p *Post) fixHtmlData(data []byte) (err error) {
	reader := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return fmt.Errorf("Creating NewDocumentFromReader Err:%v", err)
	}
	doc.Find("code[class*=\"language-\"]").Each(func(i int, s *goquery.Selection) {
		oldCode := s.Text()
		formatted, _ := syntaxhighlight.AsHTML([]byte(oldCode))
		s.ReplaceWithHtml(string(formatted))
	})
	newDoc, err := doc.Html()
	if err != nil {
		return fmt.Errorf("Generating new html err: %v", err)
	}
	// replace the html tags that we donnot need
	for _, tag := range []string{"<html>", "</html>", "<head>", "</head>", "<body>", "</body>"} {
		newDoc = strings.Replace(newDoc, tag, "", 1)
	}
	data = []byte(newDoc)
	return
}
