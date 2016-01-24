package main

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	read = 1 << iota
	write
)

type StructRepr struct {
	BaseRepr
	Name   string
	flags  uint
	Fields []*StructFieldRepr
}

type StructFieldRepr struct {
	BaseFieldRepr
	Name  string
	Tag   string
	flags uint
}

func (self *FileData) doStruct(cgText string) (string, string, error) {
	strct := StructRepr{}

	var err error

	if !unicode.IsSpace(rune(cgText[0])) {
		return cgText, "",
			fmt.Errorf("@struct is expected to be followed by a space and the name.")
	}

	cgText, foundNewline := trimLeftCheckNewline(cgText)
	if foundNewline {
		return cgText, "",
			fmt.Errorf("The name must be on the same line as the @enum")
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

	return cgText, strct.Name, nil
}

func (self *StructRepr) doFields(cgText string) (_ string, err error) {
	for len(cgText) > 0 && !strings.HasPrefix(cgText, "@struct") {
		var f = StructFieldRepr{}

		cgText, f.Name, err = getIdent(cgText)
		if err != nil {
			return cgText, err
		}

		cgText, err = f.gatherFlags(cgText)
		if err != nil {
			return cgText, err
		}

		self.Fields = append(self.Fields, &f)
	}

	if len(self.Fields) == 0 {
		return cgText, fmt.Errorf("Enums must have at least one variant defined")
	}

	return cgText, nil
}

func (self *StructRepr) gatherFlags(cgText string) (string, error) {
	cgText, flags, _, err := self.genericGatherFlags(cgText, false)
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

func (self *StructFieldRepr) gatherFlags(cgText string) (string, error) {
	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "tags": // The separator used when joining bitflags
			if !flag.FoundEqual {
				return cgText, fmt.Errorf("%q is meant to have a value", flag.Name)
			}
			self.Tag = flag.Value

		case "read": // The separator used when joining bitflags
			if err = self.doBooleanFlag(flag, read); err != nil {
				return cgText, err
			}

		case "write": // The separator used when joining bitflags
			if err = self.doBooleanFlag(flag, write); err != nil {
				return cgText, err
			}

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}
