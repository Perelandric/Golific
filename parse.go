package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	bitflags = 1 << iota
	jsonMarshalIsString
	jsonUnmarshalIsString
	xmlMarshalIsString
	xmlUnmarshalIsString
	dropJson
	dropXml
)

type EnumData struct {
	Package string
	File    string
	Reprs   []*EnumRepr
}

type EnumRepr struct {
	Name string

	flags    uint
	FlagSep  string
	iterName string
	unique   string
	Fields   []*FieldRepr
}

type FieldRepr struct {
	Name        string
	String      string
	Description string
	Value       int64
}

type Flag struct {
	Name       string
	Value      string
	FoundEqual bool
}

func (self *EnumData) DoFile(file string) error {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	self.Package = f.Name.Name
	if len(self.Package) == 0 {
		return fmt.Errorf("No package Name")
	}

	var dir, filename = filepath.Split(file)

	self.File = filepath.Join(dir, "enum____"+filename)

	for _, cg := range f.Comments {
		self.doComment(cg)
	}

	if err := self.generateCode(); err != nil {
		return err
	}

	return nil
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

func (self *EnumRepr) DoJson() bool { return self.flags&dropJson == 0 }
func (self *EnumRepr) DoXml() bool  { return self.flags&dropXml == 0 }

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

func (self *EnumData) doComment(cg *ast.CommentGroup) {
	cgText := strings.TrimSpace(cg.Text())

	if !strings.HasPrefix(cgText, "@enum") { // First item must be @enum
		return
	}

	var err error
	var name string

	for {
		cgText = strings.TrimSpace(cgText)

		var idx = strings.Index(cgText, "@enum")
		if idx != 0 {
			break
		}

		cgText = cgText[5:] // Strip away the `@enum`

		if cgText, name, err = self.doEnum(cgText); err != nil {
			if len(name) > 0 {
				log.Printf("%s: %s\n", name, err)
			} else {
				log.Println(err)
			}

			if idx := strings.Index(cgText, "@enum"); idx == -1 {
				break
			} else {
				cgText = cgText[idx:] // Slice away everyting until the `@enum`
			}
		}
	}
}

func (self *EnumData) doEnum(cgText string) (string, string, error) {
	var enum = EnumRepr{
		iterName: "Values",
	}

	var err error

	if cgText, enum.Name, err = getIdent(strings.TrimSpace(cgText)); err != nil {
		return cgText, enum.Name, err
	}

	if cgText, err = enum.gatherFlags(cgText); err != nil {
		return cgText, enum.Name, err
	}

	if cgText, err = enum.doFields(cgText); err != nil {
		return cgText, enum.Name, err
	}

	self.Reprs = append(self.Reprs, &enum)

	return cgText, enum.Name, nil
}

func (self *EnumRepr) doFields(cgText string) (_ string, err error) {
	for len(cgText) > 0 && !strings.HasPrefix(cgText, "@enum") {
		var f = FieldRepr{
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
	cgText, flags, foundNewline, err := genericGatherFlags(cgText)
	if err != nil {
		return cgText, err
	}

	if foundNewline == false {
		return cgText, fmt.Errorf("Expected line break after last descriptor flag")
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

func (self *EnumRepr) doBooleanFlag(flag Flag, toSet int) error {
	if !flag.FoundEqual || flag.Value == "true" {
		self.flags |= dropJson
	} else if flag.Value == "false" {
		self.flags &^= dropJson
	} else {
		return fmt.Errorf("Invalid value %q for %q", flag.Value, flag.Name)
	}
	return nil
}

func (self *FieldRepr) gatherFlags(cgText string) (string, error) {
	cgText, flags, foundNewline, err := genericGatherFlags(cgText)
	if err != nil {
		return cgText, err
	}

	if len(cgText) > 0 && !foundNewline {
		return cgText, fmt.Errorf("Expected line break after variant definition")
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

func (self *EnumRepr) IsBitflag() bool {
	return self.flags&bitflags == bitflags
}

func getFlagWord(source string) (_, word string, err error) {
	var n = 0

	for _, r := range source {
		if ('a' <= r && r <= 'z') || r == '_' {
			n += utf8.RuneLen(r)
		} else if r == '=' || unicode.IsSpace(r) {
			break
		} else {
			return source, "", fmt.Errorf("Invalid flag: %q", source[:n])
		}
	}

	if n == 0 {
		return "", "", fmt.Errorf("Invalid flag: %q", "")
	}

	return source[n:], source[:n], nil
}

func getIdent(source string) (_, ident string, err error) {
	source = strings.TrimSpace(source)

	var n = 0

	for i, r := range source {
		if isIdentRune(i, r) {
			n += utf8.RuneLen(r)
		} else if unicode.IsSpace(r) {
			break
		} else {
			return source, "", fmt.Errorf("Invalid identifier: %q", source[:n])
		}
	}

	if n == 0 {
		return "", "", fmt.Errorf("Invalid identifier: %q", "")
	}

	return source[n:], source[:n], nil
}

func isIdent(word string) bool {
	if len(word) == 0 {
		return false
	}
	for i, r := range word {
		if !isIdentRune(i, r) {
			return false
		}
	}
	return true
}

func isIdentRune(i int, r rune) bool {
	if unicode.IsLetter(r) == false && unicode.IsDigit(r) == false && r != '_' {
		return false
	}
	if i == 0 && unicode.IsDigit(r) {
		return false
	}
	return true
}

// Does a left trim, but also checks if a newline was found
func trimLeftCheckNewline(s string) (string, bool) {
	var n = 0
	var found = false

	for _, r := range s {
		if unicode.IsSpace(r) {
			n += utf8.RuneLen(r)

			if r == '\n' || r == '\r' {
				found = true
			}
		} else {
			break
		}
	}
	return s[n:], found
}

func genericGatherFlags(cgText string) (string, []Flag, bool, error) {
	var flags = make([]Flag, 0)
	var foundNewline bool
	var err error

	cgText, foundNewline = trimLeftCheckNewline(cgText)

	for strings.HasPrefix(cgText, "--") {

		cgText = cgText[2:] // strip away the "--"

		var f Flag

		if cgText, f.Name, err = getFlagWord(cgText); err != nil {
			return cgText, flags, foundNewline, err
		}

		cgText, foundNewline = trimLeftCheckNewline(cgText)

		if strings.HasPrefix(cgText, "=") {
			f.FoundEqual = true

			if foundNewline {
				return cgText, flags, foundNewline, fmt.Errorf("Invalid line break before '='")
			}

			cgText = cgText[1:] // Strip away the `=`

			cgText, foundNewline = trimLeftCheckNewline(cgText)
			if foundNewline {
				return cgText, flags, foundNewline, fmt.Errorf("Invalid line break after '='")
			}

			if len(cgText) == 0 {
				return cgText, flags, false, fmt.Errorf("Expected value after '='")
			}

			if cgText[0] == '"' || cgText[0] == '\'' {
				var idx = strings.IndexByte(cgText[1:], cgText[0])

				if idx == -1 {
					return cgText, flags, false, fmt.Errorf("Missing closing quote")
				}
				idx += 1 // Because we started searching on the second character

				f.Value, cgText = cgText[1:idx], cgText[idx+1:]

			} else { // Get unquoted value
				var idx = 0
				for _, r := range cgText {
					if unicode.IsSpace(r) {
						break
					}
					idx += utf8.RuneLen(r)
				}
				f.Value, cgText = cgText[0:idx], cgText[idx:]
			}

			cgText, foundNewline = trimLeftCheckNewline(cgText)
		}

		flags = append(flags, f)
	}

	return cgText, flags, foundNewline, nil
}
