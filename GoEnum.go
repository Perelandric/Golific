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

	var data = FileData{
		Imports: make(map[string]bool, 3),
	}

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
	Imports map[string]bool
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

	if err := self.generateCode(); err != nil {
		return err
	}

	return nil
}

func (self *FileData) doComment(cg *ast.CommentGroup) {
	cgText := cg.Text()

	var err error
	var name, prefix string

	for {
		cgText = strings.TrimSpace(cgText)

		if prefix = getPrefix(cgText); prefix == "" {
			break
		}

		cgText = cgText[len(prefix):] // Strip away the `@prefix`

		if len(cgText) == 0 {
			log.Printf("Found %s with no definition.\n", prefix)
			break
		}

		var parser func(string) (string, string, error)

		switch prefix {
		case "@enum":
			parser = self.doEnum
		case "@struct":
			parser = self.doStruct
		default:
			log.Fatalf("Unknown prefix %q\n", prefix)
		}

		if cgText, name, err = parser(cgText); err != nil {
			if len(name) > 0 {
				log.Printf("%s: %s\n", name, err)
			} else {
				log.Println(err)
			}

			if idx := nextDescriptor(cgText); idx == -1 {
				break

			} else {
				if between := cgText[0:idx]; len(strings.TrimSpace(between)) != 0 {
					log.Printf("Skipping invalid data in comment: %q\n", between)
				}
				cgText = cgText[idx:] // Slice away everyting until the `@prefix`
			}
		}
	}
}

func getPrefix(cgText string) string {
	for _, prefix := range [...]string{"@enum", "@struct"} {
		if strings.HasPrefix(cgText, prefix) {
			return prefix
		}
	}
	return ""
}

func nextDescriptor(cgText string) int {
	for i := 0; i < len(cgText); i++ {
		if cgText[i] == '@' {
			if prefix := getPrefix(cgText); prefix != "" {
				return i
			}
		}
	}
	return -1
}
