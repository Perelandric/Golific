package main

import (
	"fmt"
	"log"
	"strings"
	"unicode"
)

const (
	read = 1 << iota
	write
)

type StructRepr struct {
	BaseRepr
	Name        string
	Constructor string
	Fields      []*StructFieldRepr
}

type StructFieldRepr struct {
	BaseFieldRepr
	Name  string // Field name
	Type  string // Field data type
	Tag   string // Typical struct field tags
	Read  string // Method name for reads
	Write string // Method name for writes
}

func (self *StructRepr) GetPrivateTypeName() string {
	return "private_" + self.getUniqueId()
}

func (self *StructRepr) GetJSONTypeName() string {
	return "json_" + self.getUniqueId()
}

func (self *StructRepr) DoCtor() bool { return len(self.Constructor) > 0 }
func (self *StructRepr) DoJson() bool { return self.flags&dropJson == 0 }

func (self *StructFieldRepr) DoRead() bool  { return len(self.Read) > 0 }
func (self *StructFieldRepr) DoWrite() bool { return len(self.Write) > 0 }

func (self *StructFieldRepr) IsPrivate() bool {
	return self.flags&(read|write) != (read | write)
}
func (self *StructFieldRepr) IsPublic() bool {
	return !self.IsPrivate()
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
	for len(cgText) > 0 && getPrefix(cgText) == "" {
		var f = StructFieldRepr{}

		if cgText, f.Name, err = getIdent(cgText); err != nil {
			return cgText, err
		}

		if isExportedIdent(f.Name) == false {
			return cgText,
				fmt.Errorf("@struct fields must be exported. Found %q", f.Name)
		}

		if cgText, f.Type, err = getType(cgText); err != nil {
			return cgText, err
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
				f.Write = "Set" + f.Name
			}

			// If `read` and `write` are set, make sure accessor names (if any) don't
			// conflict with property names
		}

		self.Fields = append(self.Fields, &f)
	}

	if len(self.Fields) == 0 {
		return cgText, fmt.Errorf("Enums must have at least one variant defined")
	}

	// TODO: Should run through all fields and make sure that no method names
	// conflict with property names. This isn't entirely necessary since the
	// compiler would report it, but still helpful.

	return cgText, nil
}

func (self *StructRepr) gatherFlags(cgText string) (string, error) {
	cgText, flags, _, err := self.genericGatherFlags(cgText, false)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {
		case "drop_json": // Do not generate JSON marshaling methods
			if err = self.doBooleanFlag(flag, dropJson); err != nil {
				return cgText, err
			}

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *StructFieldRepr) gatherFlags(cgText string) (string, error) {
	const errValidId = "%s method name must be a valid identifier."
	const warnExported = "WARNING: The %s method %q is not exported.\n"

	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "tags": // Typical struct field tags
			if !flag.FoundEqual {
				return cgText, fmt.Errorf("%q is meant to have a value", flag.Name)
			}
			self.Tag = flag.Value

		case "read": // Set read access
			self.flags |= read
			self.Read = flag.Value
			if len(flag.Value) > 0 {
				if !isIdent(flag.Value) {
					return cgText, fmt.Errorf(errValidId, flag.Name)

				} else if isExportedIdent(flag.Value) == false {
					log.Printf(warnExported, flag.Name, flag.Value)
				}
			}

		case "write": // Set write access
			self.flags |= write
			self.Write = flag.Value
			if len(flag.Value) > 0 {
				if !isIdent(flag.Value) {
					return cgText, fmt.Errorf(errValidId, flag.Name)

				} else if isExportedIdent(flag.Value) == false {
					log.Printf(warnExported, flag.Name, flag.Value)
				}
			}

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}
