package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type StructDefaults struct {
	BaseRepr
}

type StructRepr struct {
	StructDefaults
	Fields []*StructFieldRepr
}

type StructFieldRepr struct {
	BaseFieldRepr
	DefaultExpr string // Default expression
	JsonName    string // Name used for json [un]marshaling
	JsonNameCI  string // Case insensitive version of JsonName
	astField    *ast.Field
}

var structDefaults StructDefaults

func (self *StructDefaults) gatherFlags(tagText string) error {
	return self.genericGatherFlags(tagText, func(flag Flag) error {
		switch flag.Name {
		case "drop_json": // Do not generate JSON marshaling methods
			return self.doBooleanFlag(flag, dropJson)

		default:
			return UnknownFlag
		}
	})
}

func (self *StructRepr) HasPrivateJSON() bool {
	return self.flags&hasPrivateJSON == hasPrivateJSON
}

func (self *StructFieldRepr) HasJSONOmitEmpty() bool {
	return self.flags&jsonOmitEmpty == jsonOmitEmpty
}
func (self *StructFieldRepr) IsEmbedded() bool {
	return self.flags&embedded == embedded
}

func (sf *StructFieldRepr) IsPrivateField() bool {
	return sf.flags&privateJSON == privateJSON
}
func (sf *StructFieldRepr) IsPrivateJSON() bool {
	return sf.flags&(privateJSON|hasJsonTag) == (privateJSON | hasJsonTag)
}
func (sf *StructFieldRepr) HasJsonTag() bool {
	return sf.flags&hasJsonTag == hasJsonTag
}

// Gets the Name, which may be the Type for embedded fields. If so, it strips
// away any leading `*`
func (self *StructFieldRepr) GetNameMaybeType() string {
	if self.IsEmbedded() {
		return strings.TrimLeft(self.Type, "*")
	}
	return self.Name
}

func (self *FileData) doStructDefaults(tagText string) error {
	return structDefaults.gatherFlags(tagText)
}

func (self *FileData) newStruct(fset *token.FileSet, tagText string,
	docs []*ast.Comment, spec *ast.TypeSpec, strct *ast.StructType) error {

	var err error

	strct_repr := StructRepr{
		StructDefaults: structDefaults, // copy of current defaults
	}
	strct_repr.fset = fset

	if err = strct_repr.setDocsAndName(docs, spec, false); err != nil {
		return err
	}

	/*
		if err = strct_repr.gatherFlags(tagText); err != nil {
			return err
		}
	*/

	if err = strct_repr.doFields(strct.Fields); err != nil {
		return err
	}

	self.Structs = append(self.Structs, &strct_repr)

	return nil
}

func (self *StructRepr) doFields(fields *ast.FieldList) (err error) {
	if len(fields.List) == 0 {
		return fmt.Errorf("@structs must have at least one field defined")
	}

	for _, field := range fields.List {
		var f = StructFieldRepr{astField: field}
		f.fset = self.fset

		if err := f.gatherCodeCommentsAndName(field, true); err != nil {
			return err
		}

		f.JsonName = f.Name

		if f.flags&embedded == embedded {
			self.flags |= hasEmbeddedFields

		} else {
			if err = f.gatherFlags(getFlags(field.Tag)); err != nil {
				return err
			}

			if !isExportedIdent(f.Name) && f.flags&hasJsonTag == hasJsonTag {
				f.flags |= privateJSON
				self.flags |= hasPrivateJSON
			}
		}

		f.JsonNameCI = strings.ToLower(f.JsonName)

		self.Fields = append(self.Fields, &f)
	}

	return nil
}

func (self *StructFieldRepr) gatherFlags(tagText string) error {
	return self.genericGatherFlags(tagText, func(flag Flag) error {
		switch flag.Name {
		case "json": // Just to find out if it has `omitempty`
			self.flags |= hasJsonTag

			if len(flag.Value) > 0 {
				if idx := strings.IndexByte(flag.Value, ','); idx == -1 {
					self.JsonName = flag.Value

				} else {
					jsonName := strings.TrimSpace(flag.Value[0:idx])
					if len(jsonName) > 0 {
						self.JsonName = jsonName
					}

					if strings.Contains(flag.Value[idx:], "omitempty") {
						self.flags |= jsonOmitEmpty
					}
				}
			} else {
				self.JsonName = self.Name
			}

			return UnknownFlag

		default:
			return UnknownFlag
		}
	})
}

func (self *FileData) GatherStructImports() {
	if len(self.Structs) == 0 {
		return
	}
	self.Imports["Golific/gJson"] = true
	self.Imports["reflect"] = true

	self.Imports["encoding/json"] = true
}

func (self *StructFieldRepr) MaybeStruct() bool {
	switch n := self.astField.Type.(type) {
	case *ast.ArrayType, *ast.MapType:
		return false

	case *ast.Ident:
		switch n.Name {
		case "bool", "string",
			"int", "int64", "int32", "int16", "int8", "uint", "uint64", "uint32",
			"uint16", "uint8", "float64", "float32":
			return false
		}
	}
	return true
}

func (self *StructFieldRepr) CantAvoidEncodingAttempt() string {
	if self.IsPrivateField() && self.HasJsonTag() == false {
		return "false"
	}

	if self.HasJSONOmitEmpty() {
		switch n := self.astField.Type.(type) {
		case *ast.ArrayType, *ast.MapType:
			return "len(self." + self.GetNameMaybeType() + ") != 0"

		case *ast.Ident:
			switch n.Name {
			case "bool":
				return "!self." + self.GetNameMaybeType()

			case "string":
				return "len(self." + self.GetNameMaybeType() + ") != 0"

			case "int", "int64", "int32", "int16", "int8", "uint", "uint64", "uint32",
				"uint16", "uint8", "float64", "float32":
				return "self." + self.GetNameMaybeType() + " != 0"
			}
		}

		return "z, ok := interface{}(self." + self.GetNameMaybeType() + ").(gJson.Zeroable); !ok || !z.IsZero()"
	}

	return "true"
}

var struct_tmpl = `

{{- define "generate_struct"}}
{{- range $struct := .}}

/*****************************

{{$struct.Name}} struct

******************************/

// JSONEncode implements part of Golific's JSONEncodable interface.
func (self *{{$struct.Name}}) JSONEncode(encoder *gJson.Encoder) bool {
	if self == nil {
		return encoder.EncodeNull(false)
	}

	encoder.WriteRawByte('{')
	var first = true

	{{ range $f := $struct.Fields -}}
	{{if $f.IsEmbedded -}}

	if je, ok := interface{}(self.{{$f.GetNameMaybeType}}).(gJson.JSONEncodable); ok {
		first = !encoder.EmbedEncodedStruct(je, first) && first
	} else {
		first = !encoder.EmbedMarshaledStruct(self.{{$f.GetNameMaybeType}}, first) && first
	}

	{{else -}}

	if {{$f.CantAvoidEncodingAttempt}} {
		var d interface{} = self.{{$f.Name}}

		if _, ok := d.(gJson.JSONEncodable); !ok {
			if {{$f.MaybeStruct}} && reflect.ValueOf(self.{{$f.Name}}).Kind() == reflect.Struct {
				d = &self.{{$f.Name}}
			}
		}

		var doEncode = true
		if {{$f.HasJSONOmitEmpty}} { // has omitempty?
			if zer, okCanZero := d.(gJson.Zeroable); okCanZero {
				doEncode = !zer.IsZero()
			}
		}

		if doEncode {
			first = !encoder.EncodeKeyVal({{printf "%q" $f.JsonName}}, d, first, {{$f.HasJSONOmitEmpty}}) && first
		}
	}

	{{end -}}
	{{end -}}

	encoder.WriteRawByte('}')

  return true || !first
}


func (self *{{$struct.Name}}) MarshalJSON() ([]byte, error) {
	var encoder gJson.Encoder
	self.JSONEncode(&encoder)
	return encoder.Bytes(), nil
}

func (self *{{$struct.Name}}) UnmarshalJSON(j []byte) error {
	if len(j) == 4 && string(j) == "null" {
		return nil
	}

	// First unmarshal using the default unmarshaler. The temp type is so that
	// this method is not called recursively.
	type temp *{{$struct.Name}}
	if err := json.Unmarshal(j, temp(self)); err != nil {
		return err
	}

	{{if $struct.HasPrivateJSON}}

	// For every property found, perform a separate UnmarshalJSON operation. This
	// prevents overwrite of values in 'self' where properties are absent.
	m := make(map[string]json.RawMessage)

	err := json.Unmarshal(j, &m)
	if err != nil {
		return err
	}

	// JSON key comparisons are case-insensitive
	for k, v := range m {
		m[strings.ToLower(k)] = v
	}

	var data json.RawMessage
	var ok bool

	{{- range $f := $struct.Fields -}}
	{{- if $f.IsPrivateJSON}}
	if data, ok = m[{{printf "%q" $f.JsonNameCI}}]; ok {
		var temp struct{ {{$f.OriginalCode}} }
		data = append(append([]byte("{ \"{{$f.JsonNameCI}}\":"), data...), '}')

		if err = json.Unmarshal(data, &temp); err != nil {
			return fmt.Errorf(
				"Field: %s, Error: %s", {{printf "%q" $f.JsonNameCI}}, err.Error(),
			)
		}

		self.{{$f.Name}} = temp.{{$f.Name}}
	}
	{{end -}}
	{{end -}}

	{{end -}}

  return nil
}

{{end -}}
{{end -}}
`
