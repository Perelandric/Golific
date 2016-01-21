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
func (self *EnumData) AnyBitflags() bool {
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
	"log"
	"strconv"
	{{if .AnyBitflags}}"strings"{{end -}}
)
{{- range $repr := .Reprs}}
{{- $intType := .GetIntType}}
{{- $uniqField := .GetUniqueName}}
{{- $self := .GetReceiverName}}
{{- $variantType := printf "%sEnum" $repr.Name}}

/*****************************

{{$variantType}}{{if .IsBitflag}} - bit flags{{end}}

******************************/

type {{$variantType}} struct{ {{$uniqField}} {{$intType}} }

var {{$repr.Name}} = struct {
	{{- range $f := .Fields}}
	{{$f.Name}} {{$variantType}}
	{{- end}}
}{
	{{- range $f := .Fields}}
	{{$f.Name}}: {{$variantType}}{ {{$uniqField}}: {{$f.Value}} },
	{{- end}}
}

// Used to iterate in range loops
var {{.GetIterName}} = [...]{{$variantType}}{
	{{range $f := .Fields}} {{$repr.Name}}.{{$f.Name}},{{end}}
}

// Get the integer value of the enum variant
func ({{$self}} {{$variantType}}) Value() {{$intType}} {
	return {{$self}}.{{$uniqField}}
}

func ({{$self}} {{$variantType}}) IntValue() int {
	return int({{$self}}.{{$uniqField}})
}

// Name returns the name of the variant as a string.
func ({{$self}} {{$variantType}}) Name() string {
	switch {{$self}}.{{$uniqField}} {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return "{{$f.Name}}"
	{{end -}}
	}

	return ""
}

// String returns the given string value of the variant. If none has been set,
// its return value is as though 'Name()' had been called.
{{if .IsBitflag -}}
// If multiple bit values are assigned, the string values will be joined into a
// single string using "{{.FlagSep}}" as a separator.
{{- end}}
func ({{$self}} {{$variantType}}) String() string {
	switch {{$self}}.{{$uniqField}} {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return "{{$f.String}}"
	{{end -}}
  }

	{{if .IsBitflag -}}
	if {{$self}}.{{$uniqField}} == 0 {
		return ""
	}

	var vals = make([]string, 0, {{len .Fields}}/2)

	for _, item := range {{.GetIterName}} {
		if {{$self}}.{{$uniqField}} & item.{{$uniqField}} == item.{{$uniqField}} {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, "{{.FlagSep}}")
	{{else -}}
	return ""
	{{end -}}
}

// Description returns the description of the variant. If none has been set, its
// return value is as though 'String()' had been called.
func ({{$self}} {{$variantType}}) Description() string {
  switch {{$self}}.{{$uniqField}} {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return "{{$f.Description}}"
	{{end -}}
  }
  return ""
}

{{if $repr.DoJson -}}
// JSON marshaling methods
{{if $repr.JsonMarshalIsString -}}
func ({{$self}} {{$variantType}}) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Quote({{$self}}.String())), nil
}
{{- else -}}
func ({{$self}} {{$variantType}}) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Itoa({{$self}}.IntValue())), nil
}
{{- end}}

{{if $repr.JsonUnmarshalIsString -}}
func ({{$self}} *{{$variantType}}) UnmarshalJSON(b []byte) error {
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
		{{$self}}.{{$uniqField}} = {{$f.Value}}
		return nil
	{{end -}}
	{{if not .IsBitflag -}}
	default:
		log.Printf("Unexpected value: %q while unmarshaling {{$variantType}}\n", s)
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
			log.Printf("Unexpected value: %q while unmarshaling {{$variantType}}\n", part)
		}
	}

	{{$self}}.{{$uniqField}} = {{$intType}}(val)
	{{end -}}

	return nil
}
{{else -}}
func ({{$self}} *{{$variantType}}) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	{{$self}}.{{$uniqField}} = {{$intType}}(n)
	return nil
}
{{- end}}
{{- end}}

{{if $repr.DoXml -}}
// XML marshaling methods to come
{{- end}}

{{- if .IsBitflag}}
// Bitflag enum methods
func ({{$self}} {{$variantType}}) Add(v {{$variantType}}) {{$variantType}} {
	{{$self}}.{{$uniqField}} |= v.{{$uniqField}}
	return {{$self}}
}

func ({{$self}} {{$variantType}}) AddAll(v ...{{$variantType}}) {{$variantType}} {
	for _, item := range v {
		{{$self}}.{{$uniqField}} |= item.{{$uniqField}}
	}
	return {{$self}}
}

func ({{$self}} {{$variantType}}) Remove(v {{$variantType}}) {{$variantType}} {
	{{$self}}.{{$uniqField}} &^= v.{{$uniqField}}
	return {{$self}}
}

func ({{$self}} {{$variantType}}) RemoveAll(v ...{{$variantType}}) {{$variantType}} {
	for _, item := range v {
		{{$self}}.{{$uniqField}} &^= item.{{$uniqField}}
	}
	return {{$self}}
}

func ({{$self}} {{$variantType}}) Has(v {{$variantType}}) bool {
	return {{$self}}.{{$uniqField}}&v.{{$uniqField}} == v.{{$uniqField}}
}

func ({{$self}} {{$variantType}}) HasAny(v ...{{$variantType}}) bool {
	for _, item := range v {
		if {{$self}}.{{$uniqField}}&item.{{$uniqField}} == item.{{$uniqField}} {
			return true
		}
	}
	return false
}

func ({{$self}} {{$variantType}}) HasAll(v ...{{$variantType}}) bool {
	for _, item := range v {
		if {{$self}}.{{$uniqField}}&item.{{$uniqField}} != item.{{$uniqField}} {
			return false
		}
	}
	return true
}
{{end -}}
{{end -}}
`))
