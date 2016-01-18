package main

import (
	"bytes"
	"go/format"
	"os"
	"text/template"
)

func (self *EnumData) generateCode() error {
	if len(self.Reprs) == 0 {
		return nil
	}

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

// If any EnumRepr is `bitflag`, `log` is needed
func (self *EnumData) NeedsLog() bool {
	for _, repr := range self.Reprs {
		if repr.flags&bitflags == bitflags {
			return true
		}
	}
	return false
}

var tmpl = template.Must(template.New("generate_enum").Parse(
	`package {{.Package}}

import (
	"strconv"
	"strings"
	{{if .NeedsLog}}"log"{{end}}
)
{{- range $repr := .Reprs}}
{{- $intType := .GetIntType}}

/*****************************

{{$repr.Name}}Enum{{if .IsBitflag}} - bit flags{{end}}

******************************/

type {{$repr.Name}}Enum struct{ value {{$intType}} }

var {{$repr.Name}} = struct {
	{{- range $f := .Fields}}
	{{$f.Name}} {{$repr.Name}}Enum
	{{- end}}
}{
	{{- range $f := .Fields}}
	{{$f.Name}}: {{$repr.Name}}Enum{ value: {{$f.Value}} },
	{{- end}}
}

// Used to iterate in range loops
var {{.GetIterName}} = [...]{{$repr.Name}}Enum{
	{{range $f := .Fields}} {{$repr.Name}}.{{$f.Name}},{{end}}
}

// Get the integer value of the enum variant
func (self {{$repr.Name}}Enum) Value() {{$intType}} {
	return self.value
}

func (self {{$repr.Name}}Enum) IntValue() int {
	return int(self.value)
}

// Get the string representation of the enum variant
func (self {{$repr.Name}}Enum) String() string {
	switch self.value {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return "{{$f.String}}"
	{{end -}}
  }

	{{if .IsBitflag -}}
	if self.value == 0 {
		return ""
	}

	var vals = make([]string, 0, {{len .Fields}}/2)

	for _, item := range {{.GetIterName}} {
		if self.value & item.value == item.value {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, "{{.FlagSep}}")
	{{else -}}
	return ""
	{{end -}}
}

// Get the string description of the enum variant
func (self {{$repr.Name}}Enum) Description() string {
  switch self.value {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return "{{$f.Description}}"
	{{end -}}
  }
  return ""
}

{{if $repr.JsonMarshalIsString -}}
func (self {{$repr.Name}}Enum) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Quote(self.String())), nil
}
{{- else -}}
func (self {{$repr.Name}}Enum) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Itoa(self.IntValue())), nil
}
{{- end}}

{{if $repr.JsonUnmarshalIsString -}}
func (self *{{$repr.Name}}Enum) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {
	{{range $f := .Fields -}}
	case "{{$f.String}}":
		self.value = {{$f.Value}}
		return nil
	{{end -}}
	{{if not .IsBitflag -}}
	default:
		log.Printf("Unexpected value: %q while unmarshaling {{$repr.Name}}Enum\n", s)
	{{end -}}
	}

	{{if .IsBitflag -}}
	var val = 0

	for _, part := range strings.Split(string(b), "{{.FlagSep}}") {
		switch part {
		{{range $f := .Fields -}}
		case "{{$f.String}}":
			val |= {{$f.Value}}
		{{end -}}
  	default:
			log.Printf("Unexpected value: %q while unmarshaling {{$repr.Name}}Enum\n", part)
		}
	}

	self.value = {{$intType}}(val)
	{{end -}}

	return nil
}
{{else -}}
func (self *{{$repr.Name}}Enum) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	self.value = {{$intType}}(n)
	return nil
}
{{- end}}

{{- if .IsBitflag}}
// Bitflag enum methods
func (self {{$repr.Name}}Enum) Add(v {{$repr.Name}}Enum) {{$repr.Name}}Enum {
	self.value |= v.value
	return self
}

func (self {{$repr.Name}}Enum) AddAll(v ...{{$repr.Name}}Enum) {{$repr.Name}}Enum {
	for _, item := range v {
		self.value |= item.value
	}
	return self
}

func (self {{$repr.Name}}Enum) Remove(v {{$repr.Name}}Enum) {{$repr.Name}}Enum {
	self.value &^= v.value
	return self
}

func (self {{$repr.Name}}Enum) RemoveAll(v ...{{$repr.Name}}Enum) {{$repr.Name}}Enum {
	for _, item := range v {
		self.value &^= item.value
	}
	return self
}

func (self {{$repr.Name}}Enum) Has(v {{$repr.Name}}Enum) bool {
	return self.value&v.value == v.value
}

func (self {{$repr.Name}}Enum) HasAny(v ...{{$repr.Name}}Enum) bool {
	for _, item := range v {
		if self.value&item.value == item.value {
			return true
		}
	}
	return false
}

func (self {{$repr.Name}}Enum) HasAll(v ...{{$repr.Name}}Enum) bool {
	for _, item := range v {
		if self.value&item.value != item.value {
			return false
		}
	}
	return true
}
{{end -}}
{{end -}}
`))
