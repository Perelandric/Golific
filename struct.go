package main

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
)

const (
	read = 1 << iota
	write
	embedded
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
	Tag         string // Typical struct field tags
	Read        string // Method name for reads
	Write       string // Method name for writes
	DefaultExpr string // Default expression
}

var structDefaults StructDefaults

func init() {
	structDefaults.flags = 0
}

func (self *StructDefaults) gatherFlags(cgText string) (string, error) {
	flags, err := genericGatherFlags(cgText)
	if err != nil {
		return cgText, err
	}

	for i := range flags {
		var flag = flags[i]

		switch strings.ToLower(flag.Name) {
		case "drop_json": // Do not generate JSON marshaling methods
			if err = self.doBooleanFlag(flag, dropJson); err != nil {
				return cgText, err
			}

			fallthrough // Not available as a default value

		default:
			flags[i].unknown = true
		}
	}

	return cgText, nil
}

func (self *StructRepr) GetPrivateTypeName() string {
	return "private_" + self.getUniqueId()
}
func (self *StructRepr) GetJSONTypeName() string {
	return "json_" + self.getUniqueId()
}
func (self *StructRepr) DoJson() bool { return self.flags&dropJson == 0 }

func (self *StructFieldRepr) DoRead() bool  { return len(self.Read) > 0 }
func (self *StructFieldRepr) DoWrite() bool { return len(self.Write) > 0 }
func (self *StructFieldRepr) DoDefaultExpr() bool {
	return len(self.DefaultExpr) > 0
}
func (self *StructFieldRepr) IsEmbedded() bool {
	return self.flags&embedded == embedded
}
func (self *StructFieldRepr) IsPrivate() bool {
	return !self.IsPublic() && !self.IsEmbedded()
}
func (self *StructFieldRepr) IsPublic() bool {
	return self.flags&(read|write) == (read | write)
}
func (self *StructFieldRepr) GetSpaceAndTag() string {
	if len(self.Tag) > 0 {
		return fmt.Sprintf(" `%s`", self.Tag)
	}
	return ""
}

func (self *StructFieldRepr) CouldBeJSON() bool {
	return isExportedIdent(self.Name) && !self.IsEmbedded()
}

func (self *StructFieldRepr) PossibleJSONKeys() string {
	if t := self.getJSONFieldTagName(); t != "" {
		return strconv.Quote(t)
	}
	return strings.Join([]string{
		strconv.Quote(self.Name),
		strconv.Quote(strings.ToLower(string(self.Name[0])) + self.Name[1:]),
	}, ", ")
}

func (self *StructFieldRepr) getJSONFieldTagName() string {
	if idx := strings.Index(self.Tag, `json:"`); idx != -1 {
		t := self.Tag[idx+6:] // Found a valid start to the JSON field tag

		if idx = strings.IndexByte(t, '"'); idx != -1 {
			t = t[0:idx] // Found the closing quote, so it's valid

			if idx = strings.IndexByte(t, ','); idx != -1 {
				return t[0:idx] // Found a comma, so the name comes before it
			}
			return t // No comma found, so all we have is a name
		}
	}
	return ""
}

// Gets the Name, which may be the Type for embedded fields. If so, it strips
// away any leading `*`
func (self *StructFieldRepr) GetNameMaybeType() string {
	if self.IsEmbedded() {
		return strings.TrimLeft(self.Type, "*")
	}
	return self.Name
}

func (self *FileData) doStructDefaults(cgText string) (string, error) {
	return structDefaults.gatherFlags(cgText)
}

func (self *FileData) newStruct(
	cgText string, docs []*ast.Comment, spec *ast.TypeSpec) error {

	strct, ok := spec.Type.(*ast.StructType)
	if !ok || strct.Incomplete {
		return fmt.Errorf("Expected 'struct' type for @enum")
	}

	var err error

	strct_repr := StructRepr{
		StructDefaults: structDefaults, // copy of current defaults
	}

	if err = strct_repr.setDocsAndName(docs, spec); err != nil {
		return err
	}

	if cgText, err = strct_repr.gatherFlags(strings.TrimSpace(cgText)); err != nil {
		return err
	}

	if err = strct_repr.doFields(strct.Fields); err != nil {
		return err
	}

	self.Structs = append(self.Structs, &strct_repr)

	return nil
}

func (self *StructRepr) doFields(fields *ast.FieldList) (err error) {
	const name_conflit = "%q method name conflicts with property name %q"

	if len(fields.List) == 0 {
		return fmt.Errorf("Structs must have at least one field defined")
	}

	for _, field := range fields.List {
		var f = StructFieldRepr{}

		if err := f.gatherCodeCommentsAndName(field, true); err != nil {
			return err
		}

		if f.flags&embedded == 0 {
			// TODO: I need `gatherFlags` to alter the string and keep any unrecognized
			// tags, since those will need to be added on to the resulting struct.
			if err = f.gatherFlags(getFlags(field.Tag)); err != nil {
				return err
			}

			if f.flags&(read|write) == (read | write) { // if `read` AND `write`
				if f.Name == f.Read {
					return fmt.Errorf(name_conflit, "read", f.Name)
				}
				if f.Name == f.Write {
					return fmt.Errorf(name_conflit, "write", f.Name)
				}

				// if `read` OR `write` are set (but not both), set default name if needed
			} else if f.flags&read == read || f.flags&write == write {
				if f.flags&read == read && len(f.Read) == 0 {
					f.Read = f.Name
				}

				if f.flags&write == write && len(f.Write) == 0 {
					f.Write = "Set" + strings.Title(f.Name)
				}
			}
		}

		self.Fields = append(self.Fields, &f)
	}

	return nil
}

func (self *StructFieldRepr) gatherFlags(cgText string) error {
	flags, err := genericGatherFlags(cgText)
	if err != nil {
		return err
	}

	for i := range flags {
		var flag = flags[i]

		switch strings.ToLower(flag.Name) {

		case "read": // Set read access
			self.flags |= read
			if flag.FoundColon {
				self.Read = flag.Value
			}

		case "write": // Set write access
			self.flags |= write
			if flag.FoundColon {
				self.Write = flag.Value
			}

		default:
			flags[i].unknown = true
		}
	}

	return nil
}

func (self *FileData) GatherStructImports() {
	if len(self.Structs) == 0 {
		return
	}
	self.Imports["encoding/json"] = true
}

var struct_tmpl = `
{{- define "generate_struct"}}
{{- range $struct := .}}
{{- $privateType := $struct.GetPrivateTypeName}}
{{- $jsonType := $struct.GetJSONTypeName}}

/*****************************

{{$struct.Name}} struct

******************************/

{{$struct.DoDocs -}}
type {{$struct.Name}} struct {
  private {{$privateType}}
  {{- range $f := $struct.Fields}}
	{{- if $f.IsEmbedded}}
	{{printf "%s%s%s" $f.DoDocs $f.Type $f.GetSpaceAndTag}}
  {{- else if $f.IsPublic}}
  {{printf "%s%s %s%s" $f.DoDocs $f.Name $f.Type $f.GetSpaceAndTag}}
  {{- end -}}
  {{end -}}
}

type {{$privateType}} struct {
  {{- range $f := $struct.Fields}}
  {{- if $f.IsPrivate}}
  {{printf "%s%s %s%s" $f.DoDocs $f.Name $f.Type $f.GetSpaceAndTag}}
  {{- end -}}
  {{end -}}
}

type {{$jsonType}} struct {
  *{{- $privateType}}
  {{- range $f := $struct.Fields}}
	{{- if $f.IsEmbedded}}
	{{printf "%s%s" $f.Type $f.GetSpaceAndTag}}
  {{- else if $f.IsPublic}}
  {{printf "%s %s%s" $f.Name $f.Type $f.GetSpaceAndTag}}
  {{- end -}}
  {{end -}}
}

{{- range $f := $struct.Fields}}
{{- if $f.DoRead}}
func (self *{{$struct.Name}}) {{$f.Read}} () {{$f.Type}} {
  {{- if $f.IsPrivate}}
  return self.private.{{$f.Name}}
  {{- else -}}
  return self.{{$f.Name}}
  {{end -}}
}
{{end -}}
{{if $f.DoWrite}}
func (self *{{$struct.Name}}) {{$f.Write}} ( v {{$f.Type}} ) {
  {{- if $f.IsPrivate}}
  self.private.{{$f.Name}} = v
  {{- else -}}
  self.{{$f.Name}} = v
  {{end -}}
}
{{end -}}
{{end -}}

{{- if $struct.DoJson}}
func (self *{{$struct.Name}}) MarshalJSON() ([]byte, error) {
  return json.Marshal({{$jsonType}} {
    &self.private,
    {{range $f := $struct.Fields -}}
    {{if or $f.IsEmbedded $f.IsPublic -}}
    self.{{$f.GetNameMaybeType}},
    {{end -}}
    {{end -}}
  })
}

func (self *{{$struct.Name}}) UnmarshalJSON(j []byte) error {
	if len(j) == 4 && string(j) == "null" {
		return nil
	}

	m := make(map[string]json.RawMessage)

	err := json.Unmarshal(j, &m)
	if err != nil {
		return err
	}

	// For every property found, perform a separate UnmarshalJSON operation. This
	// prevents overwrite of values in 'self' where properties are absent.
	for key, rawMsg := range m {
		// The anon structs in each case are needed for field tags

		switch key {
		{{- range $f := $struct.Fields -}}
		{{if $f.CouldBeJSON}}
		case {{$f.PossibleJSONKeys}}:

		var x struct {
			F {{$f.Type}}{{$f.GetSpaceAndTag}}
		}

		var msgForStruct = append(append(append(append(
			[]byte("{\""), key...), "\":"...), rawMsg...), '}')

		if err = json.Unmarshal(msgForStruct, &x); err != nil {
			return err
		}

		{{if $f.IsPublic}}
			self.{{$f.Name}} = x.F
		{{else -}}
			self.private.{{$f.Name}} = x.F
		{{- end -}}

		{{end -}}
		{{end}}
		default:
			// Ignoring unknown property
		}
	}
  return nil
}
{{end -}}

{{end -}}
{{end -}}
`
