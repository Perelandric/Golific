package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

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
	Fields []*UnionFieldRepr
}

type UnionFieldRepr struct {
	BaseFieldRepr
	astField *ast.Field
}

var unionDefaults UnionDefaults

func (self *UnionDefaults) gatherFlags(tagText string) error {
	return self.genericGatherFlags(tagText, func(flag Flag) error {
		switch flag.Name {
		default:
			return UnknownFlag
		}
		return nil
	})
}

func (self *FileData) doUnionDefaults(tagText string) error {
	return unionDefaults.gatherFlags(tagText)
}

func (self *FileData) newUnion(fset *token.FileSet, tagText string,
	docs []*ast.Comment, spec *ast.TypeSpec, strct *ast.StructType) error {

	var err error

	union_repr := UnionRepr{
		UnionDefaults: unionDefaults, // copy of current defaults
	}
	union_repr.fset = fset

	if err = union_repr.setDocsAndName(docs, spec, true); err != nil {
		return err
	}

	if err = union_repr.gatherFlags(tagText); err != nil {
		return err
	}

	if err = union_repr.doFields(strct.Fields); err != nil {
		return err
	}

	self.Unions = append(self.Unions, &union_repr)

	return nil
}

func (self *UnionRepr) doFields(fields *ast.FieldList) (err error) {
	if len(fields.List) == 0 {
		return fmt.Errorf("@unions must have at least one member defined")
	}

	for _, field := range fields.List {
		var f = UnionFieldRepr{astField: field}
		f.fset = self.fset

		if err := f.gatherCodeCommentsAndName(field, true); err != nil {
			return err
		}

		if f.flags&embedded == 0 {
			return fmt.Errorf("@union members must be defined as embedded fields")
		}

		if err = f.gatherFlags(getFlags(field.Tag)); err != nil {
			return err
		}

		self.Fields = append(self.Fields, &f)
	}

	return nil
}

func (self *UnionFieldRepr) gatherFlags(tagText string) error {
	return self.genericGatherFlags(tagText, func(flag Flag) error {
		switch flag.Name {
		default:
			return UnknownFlag
		}
	})
}

func (self *FileData) GatherUnionImports() {
	if len(self.Unions) == 0 {
		return
	}
	self.Imports["encoding/json"] = true
	self.Imports["Golific/gJson"] = true
}

func (self *UnionRepr) GetUniqueMethodName() string {
	return self.Name + "_union_" + self.getUniqueId()
}

var union_tmpl = `
{{- define "generate_union"}}
{{- range $union := .}}

{{- $methodName := $union.GetUniqueMethodName}}

/*****************************

{{$union.Name}} union

******************************/

type {{$union.Name}} interface {
  {{$methodName}}()
}

{{- range $f := $union.Fields}}
func (self {{$f.Type}}) {{$methodName}}(){}
{{- end -}}

{{end -}}
{{end -}}
`
