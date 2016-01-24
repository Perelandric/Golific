package main

type StructRepr struct {
	Name   string
	flags  uint
	Fields []*StructFieldRepr
}

type StructFieldRepr struct {
	Name  string
	Tag   string
	flags uint
}
