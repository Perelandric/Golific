package main

import (
	"fmt"
	"go/ast"
	"strconv"
)

const (
	bitflags = 1 << iota
	jsonMarshalIsString
	jsonUnmarshalIsString
	xmlMarshalIsString
	xmlUnmarshalIsString
	dropJson
	dropXml
	hasDefault
	hasCustomValue
)

type EnumDefaults struct {
	BaseRepr
	FlagSep  string // ""
	iterName string // "Values"
}

type EnumRepr struct {
	EnumDefaults
	Fields []*EnumFieldRepr
}

type EnumFieldRepr struct {
	BaseFieldRepr
	String      string
	Description string
	Value       int64
}

var enumDefaults EnumDefaults

func init() {
	enumDefaults.FlagSep = ""
	enumDefaults.iterName = "Values"
	enumDefaults.flags = 0
}

func (self *EnumRepr) GetUniqueName() string {
	return "value_" + self.getUniqueId()
}

func (self *EnumRepr) DoJson() bool    { return self.flags&dropJson == 0 }
func (self *EnumRepr) DoXml() bool     { return self.flags&dropXml == 0 }
func (self *EnumRepr) IsBitflag() bool { return self.flags&bitflags == bitflags }
func (self *EnumRepr) HasDefault() bool {
	return self.flags&hasDefault == hasDefault
}

func (self *EnumRepr) JsonMarshalIsString() bool {
	return self.flags&jsonMarshalIsString == jsonMarshalIsString
}
func (self *EnumRepr) JsonUnmarshalIsString() bool {
	return self.flags&jsonUnmarshalIsString == jsonUnmarshalIsString
}

func (self *EnumRepr) GetIterName() string {
	if len(self.iterName) == 0 {
		return "Values"
	}
	return self.iterName
}

func (repr *EnumRepr) GetIntType() string {
	var bf = repr.flags&bitflags == bitflags
	var ln = int64(len(repr.Fields))

	switch {
	case (bf && ln <= 8) || ln < 256:
		return "uint8"
	case (bf && ln <= 16) || ln < 65536:
		return "uint16"
	case (bf && ln <= 32) || ln < 4294967296:
		return "uint32"
	}
	return "uint64"
}

func (self *FileData) doEnumDefaults(tagText string) error {
	return enumDefaults.gatherFlags(tagText)
}

func (ed *EnumDefaults) gatherFlags(tagText string) error {
	return ed.genericGatherFlags(tagText, func(flag Flag) (err error) {
		switch flag.Name {

		case "bitflags": // The enum values are to be bitflags
			return ed.doBooleanFlag(flag, bitflags)

		case "bitflag_separator": // The separator used when joining bitflags
			if ed.FlagSep, err = flag.getNonEmpty(); err != nil {
				return err
			}
		case "iterator_name": // Custom Name for Array of values
			ed.iterName = flag.Value

		case "json": // Set type of JSON marshaler and unmarshaler
			return ed.setMarshal(flag, jsonMarshalIsString|jsonUnmarshalIsString)

		case "json_marshal": // Set type of JSON marshaler
			return ed.setMarshal(flag, jsonMarshalIsString)

		case "json_unmarshal": // Set type of JSON unmarshaler
			return ed.setMarshal(flag, jsonUnmarshalIsString)

		case "drop_json": // Do not generate JSON marshaling methods
			return ed.doBooleanFlag(flag, dropJson)

		default:
			return UnknownFlag
		}
		return nil
	})
}

func (self *FileData) newEnum(tagText string, docs []*ast.Comment,
	spec *ast.TypeSpec, strct *ast.StructType) error {

	var err error

	enum := EnumRepr{
		EnumDefaults: enumDefaults, // copy of current defaults
	}

	if err = enum.setDocsAndName(docs, spec); err != nil {
		return err
	}

	if err = enum.gatherFlags(tagText); err != nil {
		return err
	}

	if err = enum.doFields(strct.Fields); err != nil {
		return err
	}

	self.Enums = append(self.Enums, &enum)

	var def string

	for _, f := range enum.Fields {
		if f.flags&hasDefault == hasDefault { // Only one --default variant allowed
			if len(def) > 0 {
				return fmt.Errorf("--default was previously defined on %q", def)
			}

			enum.flags |= hasDefault // Needed for `IsDefault()` method.
			def = f.Name
		}
	}
	return nil
}

func (self *EnumRepr) doFields(fields *ast.FieldList) (err error) {
	for _, field := range fields.List {
		var f = EnumFieldRepr{}

		if err = f.gatherCodeCommentsAndName(field, false); err != nil {
			return err
		}

		if f.Name == self.iterName {
			return fmt.Errorf("The variant named %q conflicts with the iterator. Use "+
				"`--iterator_name=SomeOtherIdent` to resolve the conflict.", f.Name)
		}

		// Flags come from the struct field tag
		if err = f.gatherFlags(getFlags(field.Tag)); err != nil {
			return err
		}

		if self.flags&bitflags == bitflags && f.Value != 0 {
			return fmt.Errorf("bitflags may not have a custom --value")
		}

		// Set values if no string or description value is given
		if len(f.String) == 0 {
			f.String = f.Name
		}
		if len(f.Description) == 0 {
			f.Description = f.String
		}

		// If no explicit value is set for the variant, and there's no default, then
		// provide a value.
		if f.Value == 0 && f.flags&hasDefault == 0 {
			if self.flags&bitflags == bitflags {
				f.Value = 1 << uint(len(self.Fields))
			} else {
				// TODO: Make sure there are no Custom number conflicts
				f.Value = int64(len(self.Fields) + 1)
			}
		}

		self.Fields = append(self.Fields, &f)
	}

	if len(self.Fields) == 0 {
		return fmt.Errorf("Enums must have at least one variant defined")
	}

	return nil
}

func (self *EnumFieldRepr) gatherFlags(tagText string) error {

	const errCustomDefault = "A 'value' can not be assigned on a 'default' variant"

	return self.genericGatherFlags(tagText, func(flag Flag) (err error) {
		switch flag.Name {
		case "gDefault": // The default value used when [un]marshaling
			if err = self.doBooleanFlag(flag, hasDefault); err != nil {
				return err
			}
			if self.flags&(hasDefault|hasCustomValue) == (hasDefault | hasCustomValue) {
				return fmt.Errorf(errCustomDefault)
			}

		case "gString": // The string representation of the field
			if self.String, err = flag.getWithColon(); err != nil {
				return err
			}

		case "gDescription": // The description of the field
			if self.Description, err = flag.getWithColon(); err != nil {
				return err
			}

		case "gValue": // Custom value for the field
			if _, err = flag.getWithColon(); err != nil {
				return err
			}

			if n, err := strconv.ParseUint(flag.Value, 10, 32); err != nil {
				return fmt.Errorf("%q is not a valid uint", flag.Value)

			} else {
				self.Value = int64(n)

				if self.flags&hasDefault == hasDefault {
					return fmt.Errorf(errCustomDefault)
				}

				if self.Value == 0 {
					return fmt.Errorf("The 0 value is reserved for the gDefault flag.")
				}

				self.flags |= hasCustomValue
			}
		default:
			return UnknownFlag
		}
		return nil
	})
}

func (self *EnumDefaults) setMarshal(flag Flag, flags uint) error {
	if _, err := flag.getNonEmpty(); err != nil {
		return err
	}
	switch flag.Value {
	case "string":
		self.flags |= flags
	case "value":
		self.flags &^= flags
	default:
		return fmt.Errorf("Unexpected value %q for %q", flag.Value, flag.Name)
	}
	return nil
}

func (self *FileData) GatherEnumImports() {
	if len(self.Enums) == 0 {
		return
	}
	self.Imports["strconv"] = true
	self.Imports["Golific/gJson"] = true

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

// IsZero returns true of the variant was designated as the default value. It is
// the same as IsDefault(), but implements the Zeroable interface.
func (self {{$variantType}}) IsZero() bool {
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

// JSONEncode implements part of Golific's JSONEncodable interface.
func (self {{$variantType}}) JSONEncode(encoder *gJson.Encoder) {
	{{if $enum.JsonMarshalIsString -}}
	encoder.EncodeString(self.String())
	{{- else -}}
	encoder.EncodeInt(int64(self.{{$uniqField}}))
	{{- end}}
}

{{if $enum.DoJson -}}
// JSON marshaling methods
func (self {{$variantType}}) MarshalJSON() ([]byte, error) {
	{{if $enum.JsonMarshalIsString -}}
	return []byte(strconv.Quote(self.String())), nil
	{{- else -}}
	return []byte(strconv.Itoa(int(self.{{$uniqField}}))), nil
	{{- end}}
}

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
