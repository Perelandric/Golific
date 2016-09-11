package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

const _defaults = "-defaults"

type Flag struct {
	Name            string
	Value           string
	FoundColon      bool
	ValueWasBoolean bool
	unknown         bool // marks if the flag was unknown to Golific
}

func getFlags(lit *ast.BasicLit) string {
	// Flags come from the struct field tag
	if lit != nil {
		tag, _ := strconv.Unquote(lit.Value)
		return tag
	}
	return ""
}

func isExportedIdent(id string) bool {
	return len(id) > 0 && 'A' <= id[0] && id[0] <= 'Z'
}

func (self *Flag) getWithColon() (string, error) {
	if !self.FoundColon {
		return self.Value,
			fmt.Errorf("%q expects a ':' followed by a value", self.Name)
	}
	return self.Value, nil
}
func (self *Flag) getNonEmpty() (string, error) {
	if !self.FoundColon || len(self.Value) == 0 {
		return self.Value, fmt.Errorf("%q requires a non-empty value", self.Name)
	}
	return self.Value, nil
}

func (self *Base) getUniqueId() string {
	if self.unique == "" {
		self.unique = strconv.FormatInt(rand.Int63(), 36)
	}
	return self.unique
}

type Base struct {
	flags  uint
	unique string
	Name   string
	docs   []string
}

func (self *Base) setDocsAndName(docs []*ast.Comment, spec *ast.TypeSpec) error {
	for _, d := range docs {
		self.docs = append(self.docs, d.Text)
	}

	if self.Name = spec.Name.Name; !strings.HasPrefix(self.Name, "__") {
		return fmt.Errorf("struct %q must start with '__'", self.Name)
	}
	self.Name = self.Name[2:] // slice away the '__'
	return nil
}

func (self *Base) doBooleanFlag(flag Flag, toSet uint) error {
	if !flag.FoundColon || (flag.ValueWasBoolean && flag.Value == "true") {
		self.flags |= toSet

	} else if flag.ValueWasBoolean && flag.Value == "false" {
		self.flags &^= toSet

	} else {
		return fmt.Errorf("Invalid value %q for %q", flag.Value, flag.Name)
	}
	return nil
}

func (self *Base) DoDocs() string {
	if len(self.docs) > 0 {
		return strings.Join(self.docs, "\n") + "\n"
	}
	return ""
}

type BaseRepr struct {
	Base
}
type BaseFieldRepr struct {
	Base
	Type string
}

// Gathers code comments. Comments are abandoned if a @prefix is found after.
func (self *BaseFieldRepr) gatherCodeCommentsAndName(
	f *ast.Field, allow_embedded bool) error {

	// Comes from any comment lines before a field
	if f.Doc != nil {
		for _, c := range f.Doc.List {
			self.docs = append(self.docs, c.Text)
		}
	}

	if len(f.Names) == 0 && allow_embedded {
		self.flags |= embedded

	} else if len(f.Names) != 1 {
		// TODO: Need to support multiple names for a single definition
		return fmt.Errorf("Struct field must have exactly one name")
	}

	if len(f.Names) >= 1 {
		self.Name = f.Names[0].Name
	}

	if ident, ok := f.Type.(*ast.Ident); ok {
		self.Type = ident.Name

	} else if star, ok := f.Type.(*ast.StarExpr); ok {
		if ident, ok := star.X.(*ast.Ident); ok {
			self.Type = "*" + ident.Name
		}
	}
	return nil
}

func genericGatherFlags(cgText string) ([]Flag, error) {

	var flags = make([]Flag, 0)
	var err error

	cgText = strings.TrimSpace(cgText)

	for len(cgText) > 0 {
		var f Flag

		// Get flag name
		var n = 0

		for n < len(cgText) {
			var r = cgText[n]

			if ('a' <= r && r <= 'z') || r == '_' {
				n += 1
			} else if r == ':' || unicode.IsSpace(rune(r)) {
				break
			} else {
				return flags, fmt.Errorf("Invalid flag: %q", cgText[:n+1])
			}
		}

		if n == 0 {
			return flags, fmt.Errorf("Expected flag name")
		}

		cgText, f.Name = strings.TrimSpace(cgText[n:]), cgText[:n]

		// Get possible colon and value
		if strings.HasPrefix(cgText, ":") {
			f.FoundColon = true

			cgText = strings.TrimSpace(cgText[1:]) // Strip away the `:`

			if len(cgText) == 0 {
				return flags, fmt.Errorf("Expected value after '%s:'", f.Name)
			}

			if strings.HasPrefix(cgText, "true") {
				f.Value = "true"
				f.ValueWasBoolean = true

			} else if strings.HasPrefix(cgText, "false") {
				f.Value = "false"
				f.ValueWasBoolean = true

			} else if cgText[0] == '"' {
				cgText = cgText[1:]
				var idx = strings.IndexByte(cgText, '"')

				if idx == -1 {
					return flags, fmt.Errorf("Expected closing quote")
				}

				f.Value = cgText[0:idx]

				cgText = strings.TrimSpace(cgText[idx+1:])

			} else {
				return flags, fmt.Errorf("Expected value after '%s:'", f.Name)
			}
		}

		flags = append(flags, f)
	}

	return flags, err
}

func (self *FileData) generateCode() error {
	if len(self.Enums) == 0 && len(self.Structs) == 0 && len(self.Unions) == 0 {
		return nil
	}

	//	self.GatherUnionImports()
	self.GatherEnumImports()
	self.GatherStructImports()

	// Execute the template on the data gathered
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, self); err != nil {
		return err
	}

	// Run the go code formatter to make sure syntax is correct before writing.
	b, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	file, err := os.Create(self.File)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	return err
}

var tmpl = template.Must(template.New("generate_golific").Parse(
	union_tmpl +
		struct_tmpl +
		enum_tmpl +
		`/****************************************************************************
	This file was generated by Golific.

	Do not edit this file. If you do, your changes will be overwritten the next
	time 'generate' is invoked.
******************************************************************************/

package {{.Package}}

import (
  {{- range $imp, $_ := .Imports}}
  {{printf "%q" $imp -}}
  {{end -}}
)


/******************************************************************************
  STRUCT SUMMARY

******************************************************************************/



/******************************************************************************
	ENUM SUMMARY
{{range $enum := .Enums}}
{{$enum.Name}} (type {{printf "%sEnum" $enum.Name}}, {{$enum.GetIntType}})
{{- range $f := $enum.Fields}}
	{{ printf "%s %d %q %q" $f.Name $f.Value $f.String $f.Description -}}
{{end}}
{{end -}}
******************************************************************************/


{{- template "generate_union" .Unions}}

{{- template "generate_struct" .Structs}}

{{- template "generate_enum" .Enums}}

`))
