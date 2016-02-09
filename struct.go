package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
)

const (
	read = 1 << iota
	write
	dropCtor
	embedded
)

type StructDefaults struct {
	BaseRepr
	Constructor string
}

type StructRepr struct {
	StructDefaults
	Name   string
	Fields []*StructFieldRepr
}

type StructFieldRepr struct {
	BaseFieldRepr
	Name        string // Field name
	Type        string // Field data type
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
	cgText, flags, _, err := self.genericGatherFlags(cgText, self == &structDefaults)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {
		case "drop_json": // Do not generate JSON marshaling methods
			if err = self.doBooleanFlag(flag, dropJson); err != nil {
				return cgText, err
			}

		case "drop_ctor": // Do not generate default constructor function
			if err = self.doBooleanFlag(flag, dropCtor); err != nil {
				return cgText, err
			}

		case "ctor_name": // Custom name for the default constructor
			if self == (&structDefaults) { // Only if we're not setting defaults
				if self.Constructor, err = flag.getIdent(); err != nil {
					return cgText, err

				} else {
					continue
				}
			}

			fallthrough // Not available as a default value

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
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
func (self *StructRepr) GetCtorName() string {
	if len(self.Constructor) == 0 {
		return "New" + strings.Title(self.Name)
	}
	return self.Constructor
}

func (self *StructRepr) DoCtor() bool { return self.flags&dropCtor == 0 }
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

// embedded fields use the `.Name` to store the embedded type. Since the type
// also serves as a name, the leading '*' may need to be stripped away.
func (self *StructFieldRepr) GetNameMaybeType() string {
	return strings.TrimLeft(self.Name, "*")
}

func (self *FileData) doStructDefaults(cgText string) (string, error) {
	return structDefaults.gatherFlags(cgText)
}

func (self *FileData) doStruct(cgText string, docs []string) (string, string, error) {
	var err error

	strct := StructRepr{
		StructDefaults: structDefaults, // copy of current defaults
	}
	strct.StructDefaults.Base.docs = docs

	if !unicode.IsSpace(rune(cgText[0])) {
		return cgText, "",
			fmt.Errorf("@struct is expected to be followed by a space and the name.")
	}

	cgText, foundNewline := trimLeftCheckNewline(cgText)
	if foundNewline {
		return cgText, "",
			fmt.Errorf("The name must be on the same line as the @struct")
	}

	if cgText, strct.Name, err = getIdent(cgText); err != nil {
		return cgText, strct.Name, err
	}

	if cgText, err = strct.gatherFlags(cgText); err != nil {
		return cgText, strct.Name, err
	}

	if cgText, err = strct.doFields(cgText); err != nil {
		return cgText, strct.Name, err
	}

	self.Structs = append(self.Structs, &strct)

	return cgText, strct.Name, strct.validate()
}

func (self *StructRepr) validate() error {

	if self.flags&dropCtor == dropCtor && len(self.Constructor) > 0 {
		log.Printf("WARNING: %q: found --drop_ctor and --ctor_name\n", self.Name)
	}

	return nil
}

func (self *StructRepr) doFields(cgText string) (_ string, err error) {

	for len(cgText) > 0 {
		var foundPrefix bool
		var f = StructFieldRepr{}

		if cgText, foundPrefix = f.gatherCodeComments(cgText); foundPrefix {
			return cgText, nil
		}

		var leadingNewline, foundStr, isEmbedded, wasQuote bool

		if cgText, f.Name, err = getIdentOrType(cgText); err != nil {
			return cgText, err
		}

		// Check if the field is embedded (next is `--`, newline or quote)
		if cgText, isEmbedded, wasQuote = checkEmbedded(cgText); isEmbedded {
			f.flags |= embedded
			self.Fields = append(self.Fields, &f)

			if wasQuote { // Was a quote, so gather the field tag
				if cgText, f.Tag, _, _, err = getString(cgText, false); err != nil {
					return cgText, err
				}
			}
			// Now gather flags (if any)
			if cgText, err = f.gatherEmbeddedFlags(cgText); err != nil {
				return cgText, err
			}

			continue
		}

		// We know it's not an embedded type, so make sure it's a valid ident
		// instead of a type. Using getIdent() just so we get the expected error.
		if _, _, err = getIdent(f.Name); err != nil {
			return cgText, err
		}

		if cgText, f.Type, err = getType(cgText); err != nil {
			return cgText, err
		}

		// See if there's a tag.
		cgText, f.Tag, foundStr, leadingNewline, err = getString(cgText, false)
		if err != nil {
			return cgText, err
		}

		if leadingNewline && foundStr { // The string was on the next line
			return cgText, fmt.Errorf("Found string; expected flags or field name")
		}

		if cgText, err = f.gatherFlags(cgText); err != nil {
			return cgText, err
		}

		if f.flags&(read|write) == (read | write) { // if `read` AND `write`
			if f.Name == f.Read {
				return cgText,
					fmt.Errorf("read method name conflicts with property name %q", f.Name)
			}
			if f.Name == f.Write {
				return cgText,
					fmt.Errorf("write method name conflicts with property name %q", f.Name)
			}

			// if `read` or `write` are set (but not both), set default name if needed
		} else if f.flags&read == read || f.flags&write == write {
			if f.flags&read == read && len(f.Read) == 0 {
				f.Read = f.Name
			}
			if f.flags&write == write && len(f.Write) == 0 {
				f.Write = "Set" + strings.Title(f.Name)
			}
		}

		self.Fields = append(self.Fields, &f)
	}

	if len(self.Fields) == 0 {
		return cgText, fmt.Errorf("Structs must have at least one field defined")
	}

	return cgText, nil
}

func checkEmbedded(cgText string) (_ string, isEmbed, wasQuote bool) {
	temp, isEmbed := trimLeftCheckNewline(cgText)

	if isFlag := strings.HasPrefix(temp, "--"); isFlag {
		return temp, true, false
	}

	if isEmbed || len(temp) == 0 {
		return cgText, true, false
	}

	if temp[0] == '`' || temp[0] == '"' {
		return temp, true, true
	}
	return temp, false, false
}

func (self *StructFieldRepr) gatherEmbeddedFlags(
	cgText string) (_ string, err error) {

	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "default_expr": // Set default expression
			if self.DefaultExpr, err = flag.getNonEmpty(); err != nil {
				return cgText, err
			}
			// TODO: Should I parse the expression here? Use the formatter to verify?

		default:
			return cgText, fmt.Errorf("Unknown flag %q for embedded field", flag.Name)
		}
	}

	return cgText, nil
}

func (self *StructFieldRepr) gatherFlags(cgText string) (string, error) {
	const warnExported = "WARNING: The %s method %q is not exported.\n"

	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "read": // Set read access
			self.flags |= read
			if flag.FoundEqual {
				if self.Read, err = flag.getIdent(); err != nil {
					return cgText, err
				}
				if isExportedIdent(flag.Value) == false {
					log.Printf(warnExported, flag.Name, flag.Value)
				}
			}

		case "write": // Set write access
			self.flags |= write
			if flag.FoundEqual {
				if self.Write, err = flag.getIdent(); err != nil {
					return cgText, err
				}
				if isExportedIdent(flag.Value) == false {
					log.Printf(warnExported, flag.Name, flag.Value)
				}
			}

		case "default_expr": // Set default expression
			if self.DefaultExpr, err = flag.getNonEmpty(); err != nil {
				return cgText, err
			}
			// TODO: Should I parse the expression here? Use the formatter to verify?

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *FileData) GatherStructImports() {
	if len(self.Structs) == 0 {
		return
	}
	self.Imports["encoding/json"] = true
}

func (self *FileData) DoStructSummary() bool {
	return false
}

var struct_tmpl = `
{{- define "generate_struct"}}
{{- range $struct := .}}
{{- $privateType := $struct.GetPrivateTypeName}}
{{- $jsonType := $struct.GetJSONTypeName}}

/*****************************

{{$struct.Name}} struct

******************************/

{{- if $struct.DoCtor}}
func {{$struct.GetCtorName}}() *{{$struct.Name}} {
  return &{{$struct.Name}} {
    private: {{$privateType}} {
      {{range $f := $struct.Fields}}
      {{- if and $f.IsPrivate $f.DoDefaultExpr -}}
      {{printf "%s: %s," $f.Name $f.DefaultExpr}}
      {{end -}}
      {{end -}}
    },
    {{range $f := $struct.Fields}}
    {{- if and (or $f.IsEmbedded $f.IsPublic) $f.DoDefaultExpr -}}
    {{printf "%s: %s," $f.GetNameMaybeType $f.DefaultExpr}}
    {{end -}}
    {{end}}
  }
}
{{end}}

{{$struct.DoDocs -}}
type {{$struct.Name}} struct {
  private {{$privateType}}
  {{- range $f := $struct.Fields}}
	{{- if $f.IsEmbedded}}
	{{printf "%s%s%s" $f.DoDocs $f.Name $f.GetSpaceAndTag}}
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
	{{printf "%s%s" $f.Name $f.GetSpaceAndTag}}
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
		switch key {
		{{- range $f := $struct.Fields -}}
		{{if $f.CouldBeJSON}}
		case {{$f.PossibleJSONKeys}}:
		{{- if $f.IsPublic}}
	    err = json.Unmarshal(rawMsg, &self.{{$f.Name}})
		{{- else -}}
	    err = json.Unmarshal(rawMsg, &self.private.{{$f.Name}})
		{{- end -}}
		{{end -}}
		{{end}}
		default:
			// Ignoring unknown property
		}

	  if err != nil {
	    return err
	  }
	}
  return nil
}
{{end -}}

{{end -}}
{{end -}}
`
