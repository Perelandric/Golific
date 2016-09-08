package main

/*
// Generated
type MyUnion interface{
	_MyUnion_oijfjslkj()
}

// Generated
type AnotherUnion interface{
	_AnotherUnion_oijfjslkj()
}

// @union
type MyUnion struct {
	Foo
	Bar
	String
}

type Foo_MyUnion_oijfjslkj struct { *Foo }
func (x *Foo_MyUnion_oijfjslkj) _MyUnion_oijfjslkj() {} // generated

type Bar_MyUnion_oijfjslkj struct { *Bar }
func (x *Bar_oijfjslkj) _MyUnion_oijfjslkj() {} // generated

type String_MyUnion_oijfjslkj struct { String }
func (x *String_MyUnion_oijfjslkj) _MyUnion_oijfjslkj() {} // generated
*/

type UnionDefaults struct {
	BaseRepr
}

type UnionRepr struct {
	UnionDefaults
	Name   string
	Fields []*UnionFieldRepr
}

type UnionFieldRepr struct {
	BaseFieldRepr
	Type string // Union member type
}

var unionDefaults UnionDefaults

func init() {
	unionDefaults.flags = 0
}

/*

func (self *UnionDefaults) gatherFlags(cgText string) (string, error) {
	flags, _, err := self.genericGatherFlags(cgText, self == &unionDefaults)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *UnionRepr) GetUniqueMethodName() string {
	return self.Name + "_union_" + self.getUniqueId()
}

func (self *FileData) doUnionDefaults(cgText string) (string, error) {
	return unionDefaults.gatherFlags(cgText)
}

func (self *FileData) doUnion(cgText string, docs []string) (string, string, error) {
	var err error

	union := UnionRepr{
		UnionDefaults: unionDefaults, // copy of current defaults
	}
	union.UnionDefaults.Base.docs = docs

	if !unicode.IsSpace(rune(cgText[0])) {
		return cgText, "",
			fmt.Errorf("@union is expected to be followed by a space and the name.")
	}

	cgText, foundNewline := trimLeftCheckNewline(cgText)
	if foundNewline {
		return cgText, "",
			fmt.Errorf("The name must be on the same line as the @union")
	}

	if cgText, union.Name, err = getIdent(cgText); err != nil {
		return cgText, union.Name, err
	}

	if cgText, err = union.gatherFlags(cgText); err != nil {
		return cgText, union.Name, err
	}

	if cgText, err = union.doFields(cgText); err != nil {
		return cgText, union.Name, err
	}

	self.Unions = append(self.Unions, &union)

	return cgText, union.Name, union.validate()
}

func (self *UnionRepr) validate() error {
	return nil
}

func (self *UnionRepr) doFields(cgText string) (_ string, err error) {

	for len(cgText) > 0 {
		var foundPrefix bool
		var f = UnionFieldRepr{}

		if foundPrefix = f.gatherCodeComments(cgText); foundPrefix {
			return cgText, nil
		}

		if cgText, f.Type, err = getType(cgText); err != nil {
			return cgText, err
		}

		if cgText, err = f.gatherFlags(cgText); err != nil {
			return cgText, err
		}

		self.Fields = append(self.Fields, &f)
	}

	if len(self.Fields) == 0 {
		return cgText, fmt.Errorf("Unions must have at least one member defined")
	}

	return cgText, nil
}

func (self *UnionFieldRepr) gatherFlags(cgText string) (string, error) {
	const warnExported = "WARNING: The %s method %q is not exported.\n"

	flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *FileData) GatherUnionImports() {
	if len(self.Unions) == 0 {
		return
	}
	self.Imports["encoding/json"] = true
}

func (self *FileData) DoUnionSummary() bool {
	return false
}
*/
var union_tmpl = `
{{- define "generate_union"}}
{{- range $union := .}}

{{- $methodName := $union.GetUniqueMethodName}}

/*****************************

{{$union.Name}} union

******************************/
/*
type {{$union.Name}} interface {
  {{$methodName}}()
}

{{- range $f := $union.Fields}}
func (self {{$f.Type}}) {{$methodName}}(){}
{{- end -}}

{{end -}}
{{end -}}
`
