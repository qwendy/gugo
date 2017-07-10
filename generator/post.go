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

	"github.com/henrylee2cn/pholcus/common/goquery"
	"github.com/russross/blackfriday"
	"github.com/sourcegraph/syntaxhighlight"
)

type Post struct {
	SourcePath  string
	sourceData  []byte
	Destination string
	htmlData    []byte
	Meta        *Meta
	Template    *template.Template
}

type PostTemplateData struct {
	Title     string
	Date      string
	Tags      []string
	Catalogue string
	Content   template.HTML
}

func NewPost(sourcePath string, destination string, m *Meta, t *template.Template) *Post {
	return &Post{
		SourcePath:  sourcePath,
		Destination: destination,
		Meta:        m,
		Template:    t,
	}
}

// get source markdown file
func (p *Post) Fetch() error {
	input, err := ioutil.ReadFile(p.SourcePath)
	p.sourceData = input
	return err
}

// convert mardown file data to html
func (p *Post) Convert() (err error) {
	p.htmlData = blackfriday.MarkdownCommon(p.sourceData)
	err = p.fixHtmlData()
	if err != nil {
		return err
	}
	return
}

// write html data to file
func (p *Post) Generate() (err error) {
	f, err := os.Create(p.Destination)
	if err != nil {
		return fmt.Errorf("Creating file %s Err: %v", p.Destination, err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	td := PostTemplateData{
		Title:     p.Meta.Title,
		Date:      p.Meta.Date,
		Tags:      p.Meta.Tags,
		Catalogue: p.Meta.Catalogue,
		Content:   template.HTML(string(p.htmlData)),
	}
	if err := p.Template.Execute(writer, td); err != nil {
		return fmt.Errorf("Executing template Error: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Writing file Err: %v", err)
	}
	return nil
}

// replace code in html file,use syntaxhighlight code
// remove some html tags such as <head>
func (p *Post) fixHtmlData() (err error) {
	reader := bytes.NewReader(p.htmlData)
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
	p.htmlData = []byte(newDoc)
	return
}
