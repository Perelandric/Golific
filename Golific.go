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
	log.SetPrefix("golific: ")

	rand.Seed(time.Now().UnixNano())

	var data = FileData{
		Imports: make(map[string]bool, 3),
	}

	for _, file := range os.Args[1:] {
		fmt.Printf("Processing file: %q\n", file)
		if err := data.DoFile(file); err != nil {
			fmt.Printf("File not generated; error: %s\n", err.Error())
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

	self.File = filepath.Join(dir, "golific____"+filename)

	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "/*") { // Only do multi-line comments
				self.doComment(c)
			}
		}
	}

	if err := self.generateCode(); err != nil {
		return err
	}

	return nil
}

func (self *FileData) doComment(c *ast.Comment) {
	cgText := c.Text[2 : len(c.Text)-2]

	var err error
	var name, prefix string
	var docs []string

	for {
		cgText = strings.TrimSpace(cgText)

		if prefix = getPrefix(cgText, true); prefix == "" {
			break
		}

		cgText = cgText[len(prefix):] // Strip away the leading prefix

		if len(cgText) == 0 {
			log.Printf("Found %s with no definition.\n", prefix)
			break
		}

		var parser func(string, []string) (string, string, error)

		switch prefix {
		case "//":
			var line string
			cgText, line = getLine(cgText)
			docs = append(docs, line)
			continue

		case "@enum":
			log.SetPrefix("golific-enum: ")
			parser = self.doEnum

		case "@struct":
			log.SetPrefix("golific-struct: ")
			parser = self.doStruct

		default:
			log.Fatalf("Unknown prefix %q\n", prefix)
		}

		if cgText, name, err = parser(cgText, docs); err != nil {
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

		docs = nil
	}
}

func getPrefix(cgText string, includeComment bool) string {
	for i, prefix := range [...]string{"//", "@enum", "@struct"} {
		if i == 0 && !includeComment {
			continue
		}
		if strings.HasPrefix(cgText, prefix) {
			return prefix
		}
	}
	return ""
}

func nextDescriptor(cgText string) int {
	for i := 0; i < len(cgText); i++ {
		if cgText[i] == '@' {
			if prefix := getPrefix(cgText, false); prefix != "" {
				return i
			}
		}
	}
	return -1
}

// Grab the remainder of a line. Consume but don't include the line break.
func getLine(cgText string) (_, line string) {
	i := 0
	for ; i < len(cgText); i++ {
		if cgText[i] == '\n' {
			break
		}
		if cgText[i] == '\r' {
			if i+1 < len(cgText) && cgText[i+1] == '\n' {
				i++
			}
			break
		}
	}
	if i == len(cgText) {
		log.Println("Code comments not preceeding a descriptor line are ignored")
		return "", strings.TrimSpace(cgText)
	}
	return cgText[i+1:], strings.TrimSpace(cgText[0:i])
}
