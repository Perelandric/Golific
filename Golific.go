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

	for _, filePath := range os.Args[1:] {
		fmt.Printf("Processing file: %q\n", filePath)

		if err := data.DoFile(filePath); err != nil {
			fmt.Printf("File not generated; error: %s\n", err.Error())
		}
	}
}

type golificObj interface {
	finish(*ast.Node, []string) error
}

type FileData struct {
	Package string
	Name    string
	File    string
	Enums   []*EnumRepr
	Structs []*StructRepr
	Unions  []*UnionRepr
	Imports map[string]bool
}

func (self *FileData) DoFile(filePath string) error {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	self.Package = f.Name.Name

	var dir, filename = filepath.Split(filePath)

	self.Name = filename
	self.File = filepath.Join(dir, "golific____"+filename)

	ast.Walk(self, f)

	if err := self.generateCode(); err != nil {
		return err
	}

	return nil
}

/*
Returns the proper function to process an annotation, if found.
*/
func (self *FileData) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return self
	}

	d, ok := node.(*ast.GenDecl)

	if !ok || d.Doc == nil || len(d.Doc.List) == 0 || len(d.Specs) == 0 {
		return self
	}

	if spec, ok := d.Specs[0].(*ast.TypeSpec); ok {
		self.tryDecl(d.Doc.List, spec)
	}

	return self
}

func (self *FileData) tryDecl(cList []*ast.Comment, spec *ast.TypeSpec) {
	var c = cList[0]
	cgText := strings.TrimSpace(c.Text[2:])

	if strings.HasPrefix(c.Text, "/*") {
		cgText = strings.TrimSpace(cgText[0 : len(cgText)-2])
	}

	var err error
	var name, prefix string

	if prefix = getPrefix(cgText); prefix == "" {
		return
	}

	strct, ok := spec.Type.(*ast.StructType)
	if !ok || strct.Incomplete {
		err = fmt.Errorf("Expected 'struct' type for %s", prefix)
	}

	if err == nil {
		cgText = strings.TrimSpace(cgText[len(prefix):]) // Strip away the prefix

		log.SetPrefix(fmt.Sprintf("golific-%s (%s): ", prefix, self.Name))

		switch prefix {
		case "@enum":
			err = self.newEnum(cgText, cList[1:], spec, strct)

		case "@struct":
			err = self.newStruct(cgText, cList[1:], spec, strct)

		case "@union":
			err = self.newUnion(cgText, cList[1:], spec, strct)

		case "@enum-defaults":
			err = self.doEnumDefaults(cgText)

		case "@struct-defaults":
			err = self.doStructDefaults(cgText)

		case "@union-defaults":
			err = self.doUnionDefaults(cgText)

		default:
			log.Fatalf("Unknown prefix %q\n", prefix)
		}
	}

	if err != nil {
		if len(name) > 0 {
			log.Printf("%s: %s\n", name, err)
		} else {
			log.Println(err)
		}
	}
}

func getPrefix(cgText string) string {
	for _, prefix := range [...]string{
		"@enum-defaults", "@enum",
		"@struct-defaults", "@struct",
		"@union-defaults", "@union",
	} {
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
