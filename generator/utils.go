package generator

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
)

// generate index.html by template
func GenerateIndexFile(t *template.Template, data interface{}, dir string) error {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("Create directory error:%v", err)
	}
	f, err := os.Create(dir + "/index.html")
	if err != nil {
		return fmt.Errorf("Creating file %s Err:%v", dir, err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)

	if err := t.Execute(writer, data); err != nil {
		return fmt.Errorf("Executing template Error: %v", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Writing file Err: %v", err)
	}
	return nil
}
