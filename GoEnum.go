package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("enum: ")

	rand.Seed(time.Now().UnixNano())

	var data FileData

	for _, file := range os.Args[1:] {
		fmt.Printf("Processing file: %q\n", file)
		if err := data.DoFile(file); err != nil {
			fmt.Println(err)
		}
	}
}

type FileData struct {
	Package string
	File    string
	Enums   []*EnumRepr
	Structs []*StructRepr
}

func (self *FileData) DoFile(file string) error {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	self.Package = f.Name.Name
	if len(self.Package) == 0 {
		return fmt.Errorf("No package Name")
	}

	var dir, filename = filepath.Split(file)

	self.File = filepath.Join(dir, "enum____"+filename)

	for _, cg := range f.Comments {
		self.doComment(cg)
	}

	if err := self.generateEnumCode(); err != nil {
		return err
	}

	return nil
}

func (self *FileData) doComment(cg *ast.CommentGroup) {
	cgText := strings.TrimSpace(cg.Text())

	if !strings.HasPrefix(cgText, "@enum") { // First item must be @enum
		return
	}

	var err error
	var name string

	for {
		cgText = strings.TrimSpace(cgText)

		var idx = strings.Index(cgText, "@enum")
		if idx != 0 {
			break
		}

		cgText = cgText[5:] // Strip away the `@enum`

		if len(cgText) == 0 {
			log.Println("Found @enum with no definition.")
			break
		}

		if cgText, name, err = self.doEnum(cgText); err != nil {
			if len(name) > 0 {
				log.Printf("%s: %s\n", name, err)
			} else {
				log.Println(err)
			}

			if idx := strings.Index(cgText, "@enum"); idx == -1 {
				break
			} else {
				cgText = cgText[idx:] // Slice away everyting until the `@enum`
			}
		}
	}
}
