package main

import "fmt"

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

func (self *FileData) doStruct(cgText string) (string, string, error) {
	return "", "", fmt.Errorf("Not yet implemented\n")
}
