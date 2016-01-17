package generate_enum

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

func DoFile(file string) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return
	}

	pckg := f.Name.Name
	if len(pckg) == 0 {
		log.Println("No package name")
		return
	}

	var dir, filename = filepath.Split(file)

	osFile, err := os.Create(filepath.Join(dir, "enum____"+filename))
	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	var mux sync.Mutex
	var first = true

	for _, cg := range f.Comments {
		wg.Add(1)

		go func(txt, pckg string) {
			var buf bytes.Buffer

			if generateCode(doComment(txt), &buf) {
				mux.Lock()

				if first {
					first = false
					osFile.WriteString(doOpen(pckg))
				}
				osFile.Write(buf.Bytes())

				mux.Unlock()
			}
			wg.Done()
		}(cg.Text(), pckg)
	}

	wg.Wait()
	osFile.Close()
}

func checkValidity(res []*EnumRepr, flgs, flds, errs bool) []*EnumRepr {
	if !flgs || !flds || errs {
		if errs {
			log.Println("Enums with errors are excluded")
		} else {
			log.Println("Incomplete definition. Both flags and fields are required.")
		}
		if len(res) > 0 {
			return res[0 : len(res)-1]
		}
	}
	return res
}

func doComment(cgText string) []*EnumRepr {
	s := bufio.NewScanner(strings.NewReader(cgText))
	res := make([]*EnumRepr, 0)

	firstPass := true
	doFlags, doFields, hasErrors := true, true, false

	var repr *EnumRepr

	for line := skipEmptyLines(s); len(line) > 0; line = skipEmptyLines(s) {

		if strings.TrimSpace(line) == "@enum" {
			firstPass = false
			res = checkValidity(res, doFlags, doFields, hasErrors)

			repr = &EnumRepr{
				flag_sep: ",", // Go ahead and set the default even if not needed
			}
			res = append(res, repr)

			doFlags, doFields, hasErrors = false, false, false

		} else if firstPass {
			return nil // comment group didn't start with @enum

		} else if hasErrors {
			continue

		} else if strings.HasPrefix(strings.TrimSpace(line), "--") {
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

	return checkValidity(res, doFlags, doFields, hasErrors)
}

const unexpectedValue = "Unexpected value %q for %q\n"

func (self *EnumRepr) setFlags(flags string) bool {
	var name, value string
	var found, foundEqual bool

	for len(flags) > 0 {
		if flags, found = trashUntil(strings.TrimSpace(flags), "--", true); !found {
			if len(strings.TrimSpace(flags)) > 0 {
				log.Printf(`Expected "--", but found: %q`, flags)
				return false
			}
			break
		}

		flags, name = getWord(flags)

		switch strings.ToLower(name) {

		case "name": // Set the base name for the enum
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}

			if len(self.name) > 0 {
				log.Printf("Name is already set: %q, but found: %q\n", self.name, value)
				return false
			}

			self.name = value

		case "bitflags": // The enum values are to be bitflags
			self.flags |= bitflags

		case "bitflag_separator": // The separator used when joining bitflags
			if flags, self.flag_sep, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}

		case "iterator_name": // Custom name for Array of values
			if flags, self.iter_name, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}

		case "marshaler": // Set type of JSON and XML marshalers
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(name, value, jsonMarshalIsString|xmlMarshalIsString)

		case "unmarshaler": // Set type of JSON and XML unmarshalers
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(name, value, jsonUnmarshalIsString|xmlUnmarshalIsString)

		case "json_marshaler": // Set type of JSON marshaler
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(name, value, jsonMarshalIsString)

		case "json_unmarshaler": // Set type of JSON unmarshaler
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(name, value, jsonUnmarshalIsString)

		case "xml_marshaler": // Set type of XML marshaler
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(name, value, xmlMarshalIsString)

		case "xml_unmarshaler": // Set type of XML unmarshaler
			if flags, value, foundEqual = getValue(name, flags); !foundEqual {
				return false
			}
			self.setMarshalingFlags(name, value, xmlUnmarshalIsString)

		default:
			log.Printf("Unknown flag %q\n", name)
		}
	}
	return true
}

func (self *EnumRepr) setField(field string) bool {
	var f = FieldRepr{
		Value: -1,
	}

	field, f.Name = getWord(strings.TrimSpace(field))

	if len(f.Name) == 0 {
		log.Println("Field name is empty")
		return false
	}

	var name string
	var foundEqual, found bool

	for len(field) > 0 {
		if field, found = trashUntil(strings.TrimSpace(field), "--", true); !found {
			if len(strings.TrimSpace(field)) > 0 {
				log.Printf(`Expected "--", but found: %q`, field)
			}
			break
		}

		field, name = getWord(field)

		switch strings.ToLower(name) {

		case "string": // The string representation of the field
			if field, f.String, foundEqual = getValue(name, field); !foundEqual {
				return false
			}

		case "description": // The description of the field
			if field, f.Description, foundEqual = getValue(name, field); !foundEqual {
				return false
			}

		case "value": // Custom value for the field
			var v string

			if field, v, foundEqual = getValue(name, field); !foundEqual {
				return false
			}
			if n, err := strconv.ParseUint(v, 10, 32); err != nil {
				log.Printf("%q is not a valid uint\n", v)
				return false
			} else {
				f.Value = int(n)
			}
		default:
			log.Printf("Unknown flag %q\n", name)
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
			f.Value = 1 << uint(len(self.fields))
		} else {
			// TODO: Make sure there are no Custom number conflicts
			f.Value = len(self.fields) + 1
		}
	}

	self.fields = append(self.fields, &f)

	return true
}

func (self *EnumRepr) setMarshalingFlags(name, value string, flags uint) {
	switch strings.ToLower(value) {
	case "string":
		self.flags |= flags
	case "value":
		self.flags &^= flags
	default:
		log.Printf(unexpectedValue, value, name)
	}
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

func getWord(source string) (_ string, word string) {
	var i = 0
	for _, c := range source {
		if ('a' <= c && c <= 'z') || c == '_' || ('A' <= c && c <= 'Z') {
			word += string(c)
			i += utf8.RuneLen(c)
		} else {
			break
		}
	}
	return source[i:], word
}

func getValue(name, source string) (_, value string, foundEqual bool) {
	if len(source) == 0 {
		return "", "", false
	}

	source, foundEqual = trashUntil(strings.TrimSpace(source), "=", true)
	if !foundEqual {
		log.Printf(`Expected "=" after %q.\n`, name)
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

const (
	bitflags = 1 << iota
	jsonMarshalIsString
	jsonUnmarshalIsString
	xmlMarshalIsString
	xmlUnmarshalIsString
)

type EnumRepr struct {
	name string

	flags     uint
	flag_sep  string
	iter_name string
	fields    []*FieldRepr
}

type FieldRepr struct {
	Name        string
	String      string
	Description string
	Value       int // Only used for custom values on non-bitflag enums
}
