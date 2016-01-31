package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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
	hasDefault
	hasCustomValue
)

type EnumDefaults struct {
	BaseRepr
	FlagSep  string // ""
	iterName string // "Values"
}

type EnumRepr struct {
	EnumDefaults
	Name   string
	Fields []*EnumFieldRepr
}

type EnumFieldRepr struct {
	BaseFieldRepr
	Name        string
	String      string
	Description string
	Value       int64
}

var enumDefaults EnumDefaults

func init() {
	enumDefaults.FlagSep = ""
	enumDefaults.iterName = "Values"
	enumDefaults.flags = 0
}

func (self *EnumDefaults) gatherFlags(cgText string) (string, error) {
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
			if self.FlagSep, err = flag.getNonEmpty(); err != nil {
				return cgText, err
			}

		case "iterator_name": // Custom Name for Array of values
			if self.iterName, err = flag.getIdent(); err != nil {
				return cgText, err
			}

		case "summary": // Include a summary of this enum at the top of the file
			if err = self.doBooleanFlag(flag, summary); err != nil {
				return cgText, err
			}

		case "json": // Set type of JSON marshaler and unmarshaler
			err = self.setMarshal(flag, jsonMarshalIsString|jsonUnmarshalIsString)
			if err != nil {
				return cgText, err
			}
			/*
				case "xml": // Set type of XML marshaler and unmarshaler
					err = self.setMarshal(flag, xmlMarshalIsString|xmlUnmarshalIsString)
					if err != nil {
						return cgText, err
					}
			*/
		case "json_marshal": // Set type of JSON marshaler
			if err = self.setMarshal(flag, jsonMarshalIsString); err != nil {
				return cgText, err
			}

		case "json_unmarshal": // Set type of JSON unmarshaler
			if err = self.setMarshal(flag, jsonUnmarshalIsString); err != nil {
				return cgText, err
			}
			/*
				case "xml_marshal": // Set type of XML marshaler
					if err = self.setMarshal(flag, xmlMarshalIsString); err != nil {
						return cgText, err
					}

				case "xml_unmarshal": // Set type of XML unmarshaler
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

func (self *EnumRepr) GetUniqueName() string {
	return "value_" + self.getUniqueId()
}

func (self *EnumRepr) DoSummary() bool { return self.flags&summary == summary }
func (self *EnumRepr) DoJson() bool    { return self.flags&dropJson == 0 }
func (self *EnumRepr) DoXml() bool     { return self.flags&dropXml == 0 }
func (self *EnumRepr) IsBitflag() bool { return self.flags&bitflags == bitflags }
func (self *EnumRepr) HasDefault() bool {
	return self.flags&hasDefault == hasDefault
}

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

func (self *FileData) doEnumDefaults(cgText string) (string, error) {
	return enumDefaults.gatherFlags(cgText)
}

func (self *FileData) doEnum(cgText string, docs []string) (string, string, error) {
	var err error

	enum := EnumRepr{
		EnumDefaults: enumDefaults, // copy of current defaults
	}
	enum.EnumDefaults.Base.docs = docs

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

	return cgText, enum.Name, enum.validate()
}

func (self *EnumRepr) validate() error {
	var def string

	for _, f := range self.Fields {
		if f.flags&hasDefault == hasDefault { // Only one --default variant allowed
			if len(def) > 0 {
				return fmt.Errorf("--default was previously defined on %q", def)
			}

			self.flags |= hasDefault // Needed for `IsDefault()` method.
			def = f.Name
		}
	}
	return nil
}

func (self *EnumRepr) doFields(cgText string) (_ string, err error) {
	for len(cgText) > 0 {
		var foundPrefix bool
		var f = EnumFieldRepr{}

		if cgText, foundPrefix = f.gatherCodeComments(cgText); foundPrefix {
			return cgText, nil
		}

		if cgText, f.Name, err = getIdent(cgText); err != nil {
			return cgText, err
		}

		if f.Name == self.iterName {
			return cgText,
				fmt.Errorf("The variant named %q conflicts with the iterator. Use "+
					"`--iterator_name=SomeOtherIdent` to resolve the conflict.", f.Name)
		}

		if cgText, err = f.gatherFlags(cgText); err != nil {
			return cgText, err
		}

		if self.flags&bitflags == bitflags && f.Value != 0 {
			return cgText, fmt.Errorf("bitflags may not have a custom --value")
		}

		if len(f.String) == 0 {
			f.String = f.Name
		}
		if len(f.Description) == 0 {
			f.Description = f.String
		}

		if f.Value == 0 && f.flags&hasDefault == 0 {
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

func (self *EnumFieldRepr) gatherFlags(cgText string) (string, error) {

	const errCustomDefault = "A --value can not be assigned on a --default variant"

	cgText, flags, _, err := self.genericGatherFlags(cgText, true)
	if err != nil {
		return cgText, err
	}

	for _, flag := range flags {
		switch strings.ToLower(flag.Name) {

		case "default": // The default value used when [un]marshaling
			if err = self.doBooleanFlag(flag, hasDefault); err != nil {
				return cgText, err
			}
			if self.flags&(hasDefault|hasCustomValue) == (hasDefault | hasCustomValue) {
				return cgText, fmt.Errorf(errCustomDefault)
			}

		case "string": // The string representation of the field
			if self.String, err = flag.getWithEqualSign(); err != nil {
				return cgText, err
			}

		case "description": // The description of the field
			if self.Description, err = flag.getWithEqualSign(); err != nil {
				return cgText, err
			}

		case "value": // Custom value for the field
			if _, err = flag.getWithEqualSign(); err != nil {
				return cgText, err
			}

			if n, err := strconv.ParseUint(flag.Value, 10, 32); err != nil {
				return cgText, fmt.Errorf("%q is not a valid uint", flag.Value)

			} else {
				self.Value = int64(n)

				if self.flags&hasDefault == hasDefault {
					return cgText, fmt.Errorf(errCustomDefault)
				}

				if self.Value == 0 {
					return cgText,
						fmt.Errorf("The 0 value is reserved for the --default flag.")
				}

				self.flags |= hasCustomValue
			}

		default:
			return cgText, fmt.Errorf("Unknown flag %q", flag.Name)
		}
	}

	return cgText, nil
}

func (self *EnumDefaults) setMarshal(flag Flag, flags uint) error {
	if _, err := flag.getNonEmpty(); err != nil {
		return err
	}
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
