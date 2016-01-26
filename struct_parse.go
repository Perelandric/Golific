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
	dropCtor
	embedded
)

type StructRepr struct {
	BaseRepr
	Name        string
	Constructor string
	Fields      []*StructFieldRepr
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

	if strct.flags&dropCtor == dropCtor && len(strct.Constructor) > 0 {
		log.Printf("WARNING: %q: found --drop_ctor and --ctor_name\n", strct.Name)
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
		var foundNewline, foundStr bool

		if cgText, f.Name, err = getIdentOrType(cgText); err != nil {
			return cgText, err
		}

		if cgText, f.Tag, foundStr, foundNewline, err = getString(cgText, false); err != nil {
			return cgText, err
		}
		if foundStr || foundNewline {
			f.flags |= embedded
			self.Fields = append(self.Fields, &f)
			if foundNewline {
				continue
			}
		}

		if cgText, foundNewline, err = f.gatherEmbeddedFlags(cgText); err != nil {
			return cgText, err
		}
		if foundNewline {
			f.flags |= embedded
			self.Fields = append(self.Fields, &f)
			continue
		}

		// We know it's not an embedded type, so make sure a `*` wasn't given at the
		// start of the name. Using getIdent() just so we get the expected error.
		if _, _, err = getIdent(f.Name); err != nil {
			return cgText, err
		}

		if isExportedIdent(f.Name) == false {
			return cgText,
				fmt.Errorf("@struct fields must be exported. Found %q", f.Name)
		}

		if cgText, f.Type, err = getType(cgText); err != nil {
			return cgText, err
		}

		// A linebreak here means this field is done. This is necessary before the
		// `getString()` call.
		if cgText, foundNewline = trimLeftCheckNewline(cgText); foundNewline {
			self.Fields = append(self.Fields, &f)
			continue
		}

		// See if there's a tag.
		if cgText, f.Tag, _, foundNewline, err = getString(cgText, false); err != nil {
			return cgText, err
		}
		if foundNewline {
			continue
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
		case "drop_json": // Do not generate JSON marshaling methods
			if err = self.doBooleanFlag(flag, dropJson); err != nil {
				return cgText, err
			}

		case "drop_ctor": // Do not generate default constructor function
			if err = self.doBooleanFlag(flag, dropCtor); err != nil {
				return cgText, err
			}

		case "ctor_name": // Custom name for the default constructor
			if self.Constructor, err = flag.getIdent(); err != nil {
				return cgText, err
			}

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *StructFieldRepr) gatherEmbeddedFlags(
	cgText string) (string, bool, error) {

	// If there's a newline, we have an embedded field
	cgText, foundNewline := trimLeftCheckNewline(cgText)
	if foundNewline {
		return cgText, true, nil
	}

	// If there's no leading `--` (and no newline, see above), it's not embedded
	if strings.HasPrefix(cgText, "--") == false {
		return cgText, false, nil
	}

	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, false, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "default_expr": // Set default expression
			if self.DefaultExpr, err = flag.getNonEmpty(); err != nil {
				return cgText, false, err
			}
			// TODO: Should I parse the expression here? Use the formatter to verify?

		default:
			return cgText, false, fmt.Errorf("Unknown flag %q for embedded field", flag.Name)
		}
	}

	return cgText, true, nil
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
