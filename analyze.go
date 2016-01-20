package main

import (
	"bufio"
	"fmt"
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
		self.doComment(cg.Text())
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

func (self *EnumRepr) JsonMarshalIsString() bool {
	return self.flags&jsonMarshalIsString == jsonMarshalIsString
}

func (self *EnumRepr) JsonUnmarshalIsString() bool {
	return self.flags&jsonUnmarshalIsString == jsonUnmarshalIsString
}

func (self *EnumRepr) GetIterName() string {
	if len(self.iterName) == 0 {
		return self.Name + "Values"
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

func (self *EnumData) checkValidity(flgs, flds, errs bool, repr *EnumRepr) {
	const warnMarshal = "WARNING: %s marshal and unmarshal do not match for %q\n"

	if !flgs || !flds || errs {
		if errs {
			log.Println("Enums with errors are excluded")
		} else {
			log.Println("Incomplete definition. Both flags and Fields are required.")
		}
		if len(self.Reprs) > 0 {
			self.Reprs = self.Reprs[0 : len(self.Reprs)-1]
		}

	} else if repr != nil {
		// Warn if marshaling to string and unmarshaling to number, or vice versa.
		if (repr.flags&jsonMarshalIsString == 0) !=
			(repr.flags&jsonUnmarshalIsString == 0) {
			log.Printf(warnMarshal, "JSON", repr.Name)
		}
		if (repr.flags&xmlMarshalIsString == 0) !=
			(repr.flags&xmlUnmarshalIsString == 0) {
			log.Printf(warnMarshal, "XML", repr.Name)
		}
	}
}

func (self *EnumData) doComment(cgText string) {
	s := bufio.NewScanner(strings.NewReader(cgText))

	firstPass := true
	doFlags, doFields, hasErrors := true, true, false

	var repr *EnumRepr

	for line := skipEmptyLines(s); len(line) > 0; line = skipEmptyLines(s) {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "@enum") {
			firstPass = false
			self.checkValidity(doFlags, doFields, hasErrors, repr)

			repr = &EnumRepr{
				FlagSep: ",", // Go ahead and set the default even if not needed
			}
			self.Reprs = append(self.Reprs, repr)

			doFlags, doFields, hasErrors = false, false, false

			line = line[5:] // strip out the @enum

			if len(strings.TrimSpace(line)) == 0 {
				continue

			} else if unicode.IsSpace(rune(line[0])) {
				line = strings.TrimSpace(line)

			} else {
				hasErrors = true
				continue
			}
		} else if firstPass {
			return // comment group didn't start with @enum

		} else if hasErrors {
			continue
		}

		if strings.HasPrefix(line, "--") {
			if doFields {
				log.Println("Flags must come before enum variant definitions")
				hasErrors = true
				continue
			}
			doFlags = true

			if repr.setFlags(line) == false {
				hasErrors = true
			}

		} else { // Get the field definitions
			if !doFlags {
				log.Println("No flags were defined before the first enum variant")
				hasErrors = true
				continue
			}
			doFields = true

			if repr.setField(line) == false {
				hasErrors = true
			}
		}
	}

	self.checkValidity(doFlags, doFields, hasErrors, repr)
}

const unexpectedValue = "Unexpected value %q for %q\n"

func (self *EnumRepr) setFlags(flags string) bool {
	var Name, value string
	var found, foundEqual bool

	for len(flags) > 0 {
		if flags, found = trashUntil(strings.TrimSpace(flags), "--", true); !found {
			if len(strings.TrimSpace(flags)) > 0 {
				log.Printf(`Expected "--", but found: %q`, flags)
				return false
			}
			break
		}

		flags, Name = getIdent(flags)

		switch strings.ToLower(Name) {

		case "name": // Set the base Name for the enum
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}

			if len(self.Name) > 0 {
				log.Printf("Name is already set: %q, but found: %q\n", self.Name, value)
				return false
			}

			self.Name = value

		case "bitflags": // The enum values are to be bitflags
			self.flags |= bitflags

		case "bitflag_separator": // The separator used when joining bitflags
			if flags, self.FlagSep, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			if self.FlagSep == "" {
				log.Println("`--flag_separator` must have at least one character")
				return false
			}

		case "iterator_name": // Custom Name for Array of values
			if flags, self.iterName, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}

		case "json": // Set type of JSON marshaler and unmarshaler
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(Name, value, jsonMarshalIsString|jsonUnmarshalIsString)

		case "xml": // Set type of XML marshaler and unmarshaler
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(Name, value, xmlMarshalIsString|xmlUnmarshalIsString)

		case "json_marshal": // Set type of JSON marshaler
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(Name, value, jsonMarshalIsString)

		case "json_unmarshal": // Set type of JSON unmarshaler
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(Name, value, jsonUnmarshalIsString)

		case "xml_marshal": // Set type of XML marshaler
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(Name, value, xmlMarshalIsString)

		case "xml_unmarshal": // Set type of XML unmarshaler
			if flags, value, foundEqual = getValue(Name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(Name, value, xmlUnmarshalIsString)

		default:
			log.Printf("Unknown flag %q\n", Name)
		}
	}
	return true
}

func (self *EnumRepr) setField(field string) bool {
	var f = FieldRepr{
		Value: -1,
	}

	field, f.Name = getIdent(strings.TrimSpace(field))

	if len(f.Name) == 0 {
		log.Println("Field Name is empty")
		return false
	}

	var Name string
	var foundEqual, found bool

	for len(field) > 0 {
		if field, found = trashUntil(strings.TrimSpace(field), "--", true); !found {
			if len(strings.TrimSpace(field)) > 0 {
				log.Printf(`Expected "--", but found: %q`, field)
			}
			break
		}

		field, Name = getIdent(field)

		switch strings.ToLower(Name) {

		case "string": // The string representation of the field
			if field, f.String, foundEqual = getValue(Name, field); !foundEqual {
				return false
			}

		case "description": // The description of the field
			if field, f.Description, foundEqual = getValue(Name, field); !foundEqual {
				return false
			}

		case "value": // Custom value for the field
			var v string

			if field, v, foundEqual = getValue(Name, field); !foundEqual {
				return false
			}
			if n, err := strconv.ParseUint(v, 10, 32); err != nil {
				log.Printf("%q is not a valid uint\n", v)
				return false
			} else {
				f.Value = int64(n)
			}
		default:
			log.Printf("Unknown flag %q\n", Name)
		}
	}

	if self.flags&bitflags == bitflags && f.Value != -1 {
		log.Println("bitflag enums may not have a custom --value setting")
		return false
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

	return true
}

func (self *EnumRepr) setMarshalingFlags(Name, value string, flags uint) {
	switch strings.ToLower(value) {
	case "string":
		self.flags |= flags
	case "value":
		self.flags &^= flags
	default:
		log.Printf(unexpectedValue, value, Name)
	}
}

func (self *EnumRepr) IsBitflag() bool {
	return self.flags&bitflags == bitflags
}

func trashUntil(source, search string, exclude bool) (string, bool) {
	var idx = strings.Index(source, search)
	if idx == -1 {
		return source, false
	}

	if idx > 0 {
		log.Printf("Expected %q, but found: %q\n", search, source[0:idx])
	}

	if exclude { // caller wants the leading search string removed
		idx += len(search)
	}
	return source[idx:], true
}

func getIdent(source string) (_ string, word string) {
	var i = 0
	for j, c := range source {
		if ('a' <= c && c <= 'z') || c == '_' || ('A' <= c && c <= 'Z') ||
			(j > 0 && '0' <= c && c <= '9') {
			word += string(c)
			i += utf8.RuneLen(c)
		} else {
			break
		}
	}
	return source[i:], word
}

func getValue(Name, source string) (_, value string, foundEqual bool) {
	if len(source) == 0 {
		return "", "", false
	}

	source, foundEqual = trashUntil(strings.TrimSpace(source), "=", true)
	if !foundEqual {
		log.Printf(`Expected "=" after %q.\n`, Name)
		return source, "", foundEqual
	}

	source = strings.TrimSpace(source)

	var idx = 0

	if source[0] == '"' || source[0] == '\'' {
		idx = strings.IndexByte(source[1:], source[0])
		if idx == -1 {
			return source, "", false // Expected closing quote
		} else {
			idx += 1 // Because we started searching on the second character
		}
		return source[idx+1:], source[1:idx], foundEqual
	} else { // Get unquoted value
		for _, r := range source {
			if unicode.IsSpace(r) {
				break
			}
			idx += utf8.RuneLen(r)
		}
		return source[idx:], source[0:idx], foundEqual
	}
}

// Returns an empty string when scanner is complete
func skipEmptyLines(s *bufio.Scanner) string {
	for s.Scan() {
		t := strings.TrimSpace(s.Text())

		if err := s.Err(); err != nil {
			fmt.Println(err)
		}

		if len(t) == 0 {
			continue
		}
		return t
	}
	return ""
}
