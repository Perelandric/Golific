package main

func (self *FileData) GatherEnumImports() {
	if len(self.Enums) == 0 {
		return
	}
	self.Imports["strconv"] = true

	// If any EnumRepr is includes a JSON unmarshaler, "log" is needed
	for _, repr := range self.Enums {
		// If we don't dropJson and we are unmarshaling as a string, we need "log"
		if repr.flags&(dropJson|jsonUnmarshalIsString) == jsonUnmarshalIsString {
			self.Imports["log"] = true
			break
		}
	}

	// If any EnumRepr is `bitflag`, "strings" is needed
	for _, repr := range self.Enums {
		if repr.flags&bitflags == bitflags {
			self.Imports["strings"] = true
			break
		}
	}
}

// If any EnumRepr is `bitflag`, `strings` is needed
func (self *FileData) DoEnumSummary() bool {
	for _, repr := range self.Enums {
		if repr.flags&summary == summary {
			return true
		}
	}
	return false
}

var enum_tmpl = `
{{- define "generate_enum"}}
{{- range $enum := .}}
{{- $intType := .GetIntType}}
{{- $uniqField := .GetUniqueName}}
{{- $variantType := printf "%sEnum" $enum.Name}}

/*****************************

{{$variantType}}{{if .IsBitflag}} - bit flags{{end}}

******************************/

{{$enum.DoDocs -}}
type {{$variantType}} struct{ {{$uniqField}} {{$intType}} }

var {{$enum.Name}} = struct {
	{{- range $f := .Fields}}
	{{printf "%s%s %s" $f.DoDocs $f.Name $variantType}}
	{{- end}}

	// {{.GetIterName}} is an array of all variants. Useful in range loops.
	{{.GetIterName}} [{{len .Fields}}]{{$variantType}}
}{
	{{- range $f := .Fields}}
	{{$f.Name}}: {{$variantType}}{ {{$uniqField}}: {{$f.Value}} },
	{{- end}}
}

func init() {
	{{$enum.Name}}.{{.GetIterName}} = [{{len .Fields}}]{{$variantType}}{
		{{range $f := .Fields}} {{$enum.Name}}.{{$f.Name}},{{end}}
	}
}

// Value returns the numeric value of the variant as a {{$intType}}.
func (self {{$variantType}}) Value() {{$intType}} {
	return self.{{$uniqField}}
}

// IntValue is the same as 'Value()', except that the value is cast to an 'int'.
func (self {{$variantType}}) IntValue() int {
	return int(self.{{$uniqField}})
}

// Name returns the name of the variant as a string.
func (self {{$variantType}}) Name() string {
	switch self.{{$uniqField}} {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return {{printf "%q" $f.Name}}
	{{end -}}
	}

	return ""
}

// Type returns the variant's type name as a string.
func (self {{$variantType}}) Type() string {
	return {{printf "%q" $variantType}}
}

// Namespace returns the variant's namespace name as a string.
func (self {{$variantType}}) Namespace() string {
	return {{printf "%q" $enum.Name}}
}

// IsDefault returns true if the variant was designated as the default value.
func (self {{$variantType}}) IsDefault() bool {
	return {{printf "%t" $enum.HasDefault}} && self.{{$uniqField}} == 0
}

// String returns the given string value of the variant. If none has been set,
// its return value is as though 'Name()' had been called.
{{if .IsBitflag -}}
// If multiple bit values are assigned, the string values will be joined into a
// single string using "{{.FlagSep}}" as a separator.
{{- end}}
func (self {{$variantType}}) String() string {
	switch self.{{$uniqField}} {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return {{printf "%q" $f.String}}
	{{end -}}
  }

	{{if .IsBitflag -}}
	if self.{{$uniqField}} == 0 {
		return ""
	}

	var vals = make([]string, 0, {{len .Fields}}/2)

	for _, item := range {{$enum.Name}}.{{.GetIterName}} {
		if self.{{$uniqField}} & item.{{$uniqField}} == item.{{$uniqField}} {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, {{printf "%q" .FlagSep}})
	{{else -}}
	return ""
	{{end -}}
}

// Description returns the description of the variant. If none has been set, its
// return value is as though 'String()' had been called.
func (self {{$variantType}}) Description() string {
  switch self.{{$uniqField}} {
	{{range $f := .Fields -}}
	case {{$f.Value}}:
		return {{printf "%q" $f.Description}}
	{{end -}}
  }
  return ""
}

{{if $enum.DoJson -}}
// JSON marshaling methods
{{if $enum.JsonMarshalIsString -}}
func (self {{$variantType}}) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Quote(self.String())), nil
}
{{- else -}}
func (self {{$variantType}}) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Itoa(int(self.{{$uniqField}}))), nil
}
{{- end}}

{{if $enum.JsonUnmarshalIsString -}}
func (self *{{$variantType}}) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {
	{{range $f := .Fields -}}
	case {{printf "%q" $f.String}}:
		self.{{$uniqField}} = {{$f.Value}}
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
		case {{printf "%q" $f.String}}:
			val |= {{$f.Value}}
		{{end -}}
  	default:
			log.Printf("Unexpected value: %q while unmarshaling {{$variantType}}\n", part)
		}
	}

	self.{{$uniqField}} = {{$intType}}(val)
	{{end -}}

	return nil
}
{{else -}}
func (self *{{$variantType}}) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	self.{{$uniqField}} = {{$intType}}(n)
	return nil
}
{{- end}}
{{- end}}

{{if $enum.DoXml -}}

{{- end}}

{{- if .IsBitflag}}
// Bitflag enum methods

// Add returns a copy of the variant with the value of 'v' added to it.
func (self {{$variantType}}) Add(v {{$variantType}}) {{$variantType}} {
	self.{{$uniqField}} |= v.{{$uniqField}}
	return self
}

// AddAll returns a copy of the variant with all the values of 'v' added to it.
func (self {{$variantType}}) AddAll(v ...{{$variantType}}) {{$variantType}} {
	for _, item := range v {
		self.{{$uniqField}} |= item.{{$uniqField}}
	}
	return self
}

// Remove returns a copy of the variant with the value of 'v' removed from it.
func (self {{$variantType}}) Remove(v {{$variantType}}) {{$variantType}} {
	self.{{$uniqField}} &^= v.{{$uniqField}}
	return self
}

// RemoveAll returns a copy of the variant with all the values of 'v' removed
// from it.
func (self {{$variantType}}) RemoveAll(v ...{{$variantType}}) {{$variantType}} {
	for _, item := range v {
		self.{{$uniqField}} &^= item.{{$uniqField}}
	}
	return self
}

// Has returns 'true' if the receiver contains the value of 'v', otherwise
// 'false'.
func (self {{$variantType}}) Has(v {{$variantType}}) bool {
	return self.{{$uniqField}}&v.{{$uniqField}} == v.{{$uniqField}}
}

// HasAny returns 'true' if the receiver contains any of the values of 'v',
// otherwise 'false'.
func (self {{$variantType}}) HasAny(v ...{{$variantType}}) bool {
	for _, item := range v {
		if self.{{$uniqField}}&item.{{$uniqField}} == item.{{$uniqField}} {
			return true
		}
	}
	return false
}

// HasAll returns 'true' if the receiver contains all the values of 'v',
// otherwise 'false'.
func (self {{$variantType}}) HasAll(v ...{{$variantType}}) bool {
	for _, item := range v {
		if self.{{$uniqField}}&item.{{$uniqField}} != item.{{$uniqField}} {
			return false
		}
	}
	return true
}
{{end -}}
{{end -}}
{{end -}}
`
