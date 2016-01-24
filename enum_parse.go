package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	bitflags = 1 << iota
	summary
	jsonMarshalIsString
	jsonUnmarshalIsString
	xmlMarshalIsString
	xmlUnmarshalIsString
	dropJson
	dropXml
)

type EnumRepr struct {
	BaseRepr
	Name string

	flags    uint
	FlagSep  string
	iterName string
	unique   string
	Fields   []*EnumFieldRepr
}

type EnumFieldRepr struct {
	BaseFieldRepr
	Name        string
	String      string
	Description string
	Value       int64
}

func (self *EnumRepr) GetUniqueName() string {
	if self.unique == "" {
		self.unique = "value_" + strconv.FormatInt(rand.Int63(), 36)
	}
	return self.unique
}

func (self *EnumRepr) GetReceiverName() string {
	r, _ := utf8.DecodeRuneInString(self.Name)
	return string(r) + "e"
}

func (self *EnumRepr) DoSummary() bool { return self.flags&summary == summary }
func (self *EnumRepr) DoJson() bool    { return self.flags&dropJson == 0 }
func (self *EnumRepr) DoXml() bool     { return self.flags&dropXml == 0 }
func (self *EnumRepr) IsBitflag() bool { return self.flags&bitflags == bitflags }

func (self *EnumRepr) JsonMarshalIsString() bool {
	return self.flags&jsonMarshalIsString == jsonMarshalIsString
}
func (self *EnumRepr) JsonUnmarshalIsString() bool {
	return self.flags&jsonUnmarshalIsString == jsonUnmarshalIsString
}

func (self *EnumRepr) GetIterName() string {
	if len(self.iterName) == 0 {
		return "Values"
	}
	return self.iterName
}

func (repr *EnumRepr) GetIntType() string {
	var bf = repr.flags&bitflags == bitflags
	var ln = int64(len(repr.Fields))

	switch {
	case (bf && ln <= 8) || ln < 256:
		return "uint8"
	case (bf && ln <= 16) || ln < 65536:
		return "uint16"
	case (bf && ln <= 32) || ln < 4294967296:
		return "uint32"
	}
	return "uint64"
}

func (self *FileData) doEnum(cgText string) (string, string, error) {
	enum := EnumRepr{
		iterName: "Values",
	}

	var err error

	if !unicode.IsSpace(rune(cgText[0])) {
		return cgText, "",
			fmt.Errorf("@enum is expected to be followed by a space and the name.")
	}

	cgText, foundNewline := trimLeftCheckNewline(cgText)
	if foundNewline {
		return cgText, "",
			fmt.Errorf("The name must be on the same line as the @enum")
	}

	if cgText, enum.Name, err = getIdent(cgText); err != nil {
		return cgText, enum.Name, err
	}

	if cgText, err = enum.gatherFlags(cgText); err != nil {
		return cgText, enum.Name, err
	}

	if cgText, err = enum.doFields(cgText); err != nil {
		return cgText, enum.Name, err
	}

	self.Enums = append(self.Enums, &enum)

	return cgText, enum.Name, nil
}

func (self *EnumRepr) doFields(cgText string) (_ string, err error) {
	for len(cgText) > 0 && !strings.HasPrefix(cgText, "@enum") {
		var f = EnumFieldRepr{
			Value: -1,
		}

		cgText, f.Name, err = getIdent(cgText)
		if err != nil {
			return cgText, err
		}

		if f.Name == self.iterName {
			return cgText,
				fmt.Errorf("The variant named %q conflicts with the iterator. Use "+
					"`--iterator_name=SomeOtherIdent` to resolve the conflict.", f.Name)
		}

		cgText, err = f.gatherFlags(cgText)
		if err != nil {
			return cgText, err
		}

		if self.flags&bitflags == bitflags && f.Value != -1 {
			return cgText, fmt.Errorf("bitflags may not have a custom --value")
		}

		if len(f.String) == 0 {
			f.String = f.Name
		}
		if len(f.Description) == 0 {
			f.Description = f.String
		}

		if f.Value == -1 {
			if self.flags&bitflags == bitflags {
				f.Value = 1 << uint(len(self.Fields))
			} else {
				// TODO: Make sure there are no Custom number conflicts
				f.Value = int64(len(self.Fields) + 1)
			}
		}

		self.Fields = append(self.Fields, &f)
	}

	if len(self.Fields) == 0 {
		return cgText, fmt.Errorf("Enums must have at least one variant defined")
	}

	return cgText, nil
}

func (self *EnumRepr) gatherFlags(cgText string) (string, error) {
	cgText, flags, _, err := self.genericGatherFlags(cgText, false)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "bitflags": // The enum values are to be bitflags
			if err = self.doBooleanFlag(flag, bitflags); err != nil {
				return cgText, err
			}

		case "bitflag_separator": // The separator used when joining bitflags
			if !flag.FoundEqual || len(flag.Value) == 0 {
				return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
			}
			self.FlagSep = flag.Value

		case "iterator_name": // Custom Name for Array of values
			if !flag.FoundEqual || !isIdent(flag.Value) {
				return cgText, fmt.Errorf("%q requires a valid identifier", flag.Name)
			}
			self.iterName = flag.Value

		case "summary": // Include a summary of this enum at the top of the file
			if err = self.doBooleanFlag(flag, summary); err != nil {
				return cgText, err
			}

		case "json": // Set type of JSON marshaler and unmarshaler
			if !flag.FoundEqual || len(flag.Value) == 0 {
				return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
			}
			err = self.setMarshal(flag, jsonMarshalIsString|jsonUnmarshalIsString)
			if err != nil {
				return cgText, err
			}
			/*
				case "xml": // Set type of XML marshaler and unmarshaler
					if !flag.FoundEqual || len(flag.Value) == 0 {
						return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
					}
					err = self.setMarshal(flag, xmlMarshalIsString|xmlUnmarshalIsString)
					if err != nil {
						return cgText, err
					}
			*/
		case "json_marshal": // Set type of JSON marshaler
			if !flag.FoundEqual || len(flag.Value) == 0 {
				return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
			}
			err = self.setMarshal(flag, jsonMarshalIsString)
			if err != nil {
				return cgText, err
			}

		case "json_unmarshal": // Set type of JSON unmarshaler
			if !flag.FoundEqual || len(flag.Value) == 0 {
				return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
			}
			if err = self.setMarshal(flag, jsonUnmarshalIsString); err != nil {
				return cgText, err
			}
			/*
				case "xml_marshal": // Set type of XML marshaler
					if !flag.FoundEqual || len(flag.Value) == 0 {
						return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
					}
					if err = self.setMarshal(flag, xmlMarshalIsString); err != nil {
						return cgText, err
					}

				case "xml_unmarshal": // Set type of XML unmarshaler
					if !flag.FoundEqual || len(flag.Value) == 0 {
						return cgText, fmt.Errorf("%q requires a non-empty value", flag.Name)
					}
					if err = self.setMarshal(flag, xmlUnmarshalIsString); err != nil {
						return cgText, err
					}
			*/
		case "drop_json": // Do not generate JSON marshaling methods
			if err = self.doBooleanFlag(flag, dropJson); err != nil {
				return cgText, err
			}
			/*
				case "drop_xml": // Do not generate XML marshaling methods
					if err = self.doBooleanFlag(flag, dropXml); err != nil {
						return cgText, err
					}
			*/
		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *EnumFieldRepr) gatherFlags(cgText string) (string, error) {
	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "string": // The string representation of the field
			if !flag.FoundEqual {
				return cgText, fmt.Errorf("Expected value after %q", flag.Name)
			}
			self.String = flag.Value

		case "description": // The description of the field
			if !flag.FoundEqual {
				return cgText, fmt.Errorf("Expected value after %q", flag.Name)
			}
			self.Description = flag.Value

		case "value": // Custom value for the field
			if !flag.FoundEqual {
				return cgText, fmt.Errorf("Expected value after %q", flag.Name)
			}

			if n, err := strconv.ParseUint(flag.Value, 10, 32); err != nil {
				return cgText, fmt.Errorf("%q is not a valid uint", flag.Value)
			} else {
				self.Value = int64(n)
			}

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *EnumRepr) setMarshal(flag Flag, flags uint) error {
	switch strings.ToLower(flag.Value) {
	case "string":
		self.flags |= flags
	case "value":
		self.flags &^= flags
	default:
		return fmt.Errorf("Unexpected value %q for %q", flag.Value, flag.Name)
	}
	return nil
}
