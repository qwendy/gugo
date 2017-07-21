package generator

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
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

func CopyDir(from, to string) error {
	dir, err := ioutil.ReadDir(from)
	if err != nil {
		return err
	}
	for _, file := range dir {
		if file.IsDir() {
			if err := CopyDir(from+"/"+file.Name(), to+"/"+file.Name()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(to, 0777); err != nil {
			return err
		}
		out, err := os.Create(to + "/" + file.Name())
		if err != nil {
			return err
		}
		in, err := os.Open(from + "/" + file.Name())
		if err != nil {
			return err
		}
		reader := bufio.NewReader(in)
		_, err = reader.WriteTo(out)
		if err != nil {
			return err
		}
	}
	return nil
}
