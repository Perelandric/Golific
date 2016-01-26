package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Flag struct {
	Name       string
	Value      string
	FoundEqual bool
}

type Base struct {
	flags  uint
	unique string
}

func (self *Flag) getWithEqualSign() (string, error) {
	if !self.FoundEqual {
		return self.Value,
			fmt.Errorf("%q expects an '=' followed by a (possibly empty) value",
				self.Name)
	}
	return self.Value, nil
}
func (self *Flag) getNonEmpty() (string, error) {
	if !self.FoundEqual || len(self.Value) == 0 {
		return self.Value, fmt.Errorf("%q requires a non-empty value", self.Name)
	}
	return self.Value, nil
}
func (self *Flag) getIdent() (string, error) {
	if isIdent(self.Value) {
		return self.Value, nil
	}
	return self.Value, fmt.Errorf("%q requires a valid identifier", self.Name)
}

func (self *Base) getUniqueId() string {
	if self.unique == "" {
		self.unique = strconv.FormatInt(rand.Int63(), 36)
	}
	return self.unique
}

func (self *Base) doBooleanFlag(flag Flag, toSet uint) error {
	if !flag.FoundEqual || flag.Value == "true" {
		self.flags |= toSet
	} else if flag.Value == "false" {
		self.flags &^= toSet
	} else {
		return fmt.Errorf("Invalid value %q for %q", flag.Value, flag.Name)
	}
	return nil
}

type BaseRepr struct {
	Base
}
type BaseFieldRepr struct {
	Base
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

func getIdentOrType(source string) (_, ident_type string, err error) {
	if source, ident_type, err = getIdent(source); err != nil {
		if source, ident_type, err = getType(source); err != nil {
			return source, "", fmt.Errorf("Value is neither a valid identifier nor type")
		}
	}
	return source, ident_type, nil
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
		return source, "", fmt.Errorf("Invalid identifier: %q", "")
	}

	return source[n:], source[:n], nil
}

func getType(source string) (_, typ string, err error) {
	source = strings.TrimSpace(source)

	if len(source) == 0 {
		return "", "", fmt.Errorf("Invalid type: %q", "")
	}

	var n = 0
	var a string

	if source[0] == '*' {
		n += 1
		source = source[1:]
		a = "*"
	}

	if source, typ, err = getIdent(source); err != nil {
		return "", "", err
	}
	return source, a + typ, nil
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

func isExportedIdent(word string) bool {
	return isIdent(word) && ('A' <= word[0] && word[0] <= 'Z')
}

func isValidType(word string) bool {
	if len(word) == 0 {
		return false
	}
	if word[0] == '*' {
		return isIdent(word[1:])
	}
	return isIdent(word)
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

func getString(
	source string, doSingle bool) (_, val string, foundStr, foundNewline bool, err error) {

	source, foundNewline = trimLeftCheckNewline(source)

	if len(source) > 0 &&
		(source[0] == '"' || source[0] == '`' || (doSingle && source[0] == '\'')) {

		var idx = strings.IndexByte(source[1:], source[0])

		if idx == -1 {
			return source, "", false, foundNewline, fmt.Errorf("Missing closing quote")
		}
		idx += 1 // Because we started searching on the second character

		return source[idx+1:], source[1:idx], true, foundNewline, nil
	}

	return source, "", false, foundNewline, nil
}

func (self Base) genericGatherFlags(
	cgText string, possibleEnd bool) (string, []Flag, bool, error) {

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

			var foundStr bool

			// Don't need the foundNewline; we already trimmed it above
			if cgText, f.Value, foundStr, _, err = getString(cgText, true); err != nil {
				return cgText, flags, false, err
			}

			if !foundStr {
				// Get unquoted value
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

	if !foundNewline && (!possibleEnd || len(cgText) > 0) {
		err = fmt.Errorf("Expected line break.")
	}

	return cgText, flags, foundNewline, err
}
