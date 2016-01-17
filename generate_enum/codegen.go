package generate_enum

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func generateCode(reprs []*EnumRepr, buf *bytes.Buffer) bool {
	if len(reprs) == 0 {
		return false
	}

	for _, repr := range reprs {
		// Make sure defaults are set
		if repr.iter_name == "" {
			repr.iter_name = repr.name + "Values"
		}

		var intType = getIntType(repr)

		doHeaderComment(buf, repr)
		doStructs(buf, repr, intType)
		doIterator(buf, repr)
		doValueFuncs(buf, repr, intType)
		doStringFunc(buf, repr)
		doDescFunc(buf, repr)
		doMarshalJSON(buf, repr)
		doUnmarshalJSON(buf, repr, intType)

		if repr.flags&bitflags == bitflags {
			doBitFuncs(buf, repr)
		}
	}
	return true
}

func getIntType(repr *EnumRepr) string {
	var bf = repr.flags&bitflags == bitflags
	var ln = len(repr.fields)
	var u = ""
	if bf {
		u = "u"
	}

	switch {
	case (bf && ln <= 8) || ln < 256:
		return u + "int8"
	case (bf && ln <= 16) || ln < 65536:
		return u + "int16"
	case (bf && ln <= 32) || ln < 4294967296:
		return u + "int32"
	}
	return u + "int64"
}

func doOpen(pckg string) string {
	var imps = []string{"strconv", "strings"}

	for i := range imps {
		imps[i] = strconv.Quote(imps[i])
	}

	return fmt.Sprintf(
		"package %s\n\nimport (\n  %s\n)\n", pckg, strings.Join(imps, "\n  "))
}

func doHeaderComment(buf *bytes.Buffer, repr *EnumRepr) {
	var typ string
	if repr.flags&bitflags == bitflags {
		typ = " - bit flags"
	}
	fmt.Fprintf(buf, `
/*****************************

	%sEnum%s

******************************/
`, repr.name, typ)
}

func doStructs(buf *bytes.Buffer, repr *EnumRepr, intType string) {
	fmt.Fprintf(buf, `
type %sEnum struct{ value %s }

var %s = struct {
%s}{
%s}
`, repr.name, intType, repr.name, getFieldDef(repr), getFieldInit(repr))
}

func getFieldDef(repr *EnumRepr) []byte {
	var buf bytes.Buffer
	for _, f := range repr.fields {
		fmt.Fprintf(&buf, "  %s %sEnum\n", f.Name, repr.name)
	}
	return buf.Bytes()
}

func getFieldInit(repr *EnumRepr) []byte {
	var buf bytes.Buffer
	for i, f := range repr.fields {
		if repr.flags&bitflags == bitflags {
			i = 1 << uint(i)
		}

		fmt.Fprintf(&buf, "  %s: %sEnum{%d},\n", f.Name, repr.name, i)
	}
	return buf.Bytes()
}

func doIterator(buf *bytes.Buffer, repr *EnumRepr) {
	var b bytes.Buffer
	for _, f := range repr.fields {
		fmt.Fprintf(&b, "%s.%s, ", repr.name, f.Name)
	}

	fmt.Fprintf(buf, "\nvar %s = [...]%sEnum{%s}\n",
		repr.iter_name, repr.name, b.String())
}

func doValueFuncs(buf *bytes.Buffer, repr *EnumRepr, intType string) {
	fmt.Fprintf(buf, `
func (self %sEnum) Value() %s {
	return self.value
}
func (self %sEnum) IntValue() int {
	return int(self.value)
}
`, repr.name, intType, repr.name)
}

func doStringFunc(buf *bytes.Buffer, repr *EnumRepr) {
	var b bytes.Buffer
	var bf = repr.flags&bitflags == bitflags

	var ret string

	for _, f := range repr.fields {
		if bf {
			ret = fmt.Sprintf(`
	if self.value == 0 {
		return ""
	}

  var vals = make([]string, 0, %d)

  for _, item := range %s {
    if self.value & item.value == item.value {
      vals = append(vals, item.String())
    }
  }
  return strings.Join(vals, %q)`,
				f.Value/2, repr.iter_name, repr.flag_sep)

		} else {
			ret = `	return ""`
		}

		fmt.Fprintf(&b, `
  case %d:
    return %q`, f.Value, f.String)
	}

	fmt.Fprintf(buf, `
func (self %sEnum) String() string {
	switch self.value {%s
  }
%s
}
`, repr.name, b.Bytes(), ret)
}

func doDescFunc(buf *bytes.Buffer, repr *EnumRepr) {
	var b bytes.Buffer

	for _, f := range repr.fields {
		fmt.Fprintf(&b, `
  case %d:
    return %q`, f.Value, f.Description)
	}

	fmt.Fprintf(buf, `
func (self %sEnum) Description() string {
  switch self.value {%s
  }
  return ""
}
`, repr.name, b.Bytes())
}

func doMarshalJSON(buf *bytes.Buffer, repr *EnumRepr) {
	if repr.flags&jsonMarshalIsString == jsonMarshalIsString {
		fmt.Fprintf(buf, `
func (self %sEnum) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Quote(self.String())), nil
}
`, repr.name)

	} else {
		fmt.Fprintf(buf, `
func (self %sEnum) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Itoa(self.IntValue())), nil
}
`, repr.name)

	}
}

func doUnmarshalJSON(buf *bytes.Buffer, repr *EnumRepr, intType string) {
	var b bytes.Buffer
	var bf = repr.flags&bitflags == bitflags

	for _, f := range repr.fields {
		fmt.Fprintf(&b, `
  case %q:
  	self.value = %d
  	return nil`, f.String, f.Value)
	}

	var bb bytes.Buffer

	if bf {
		fmt.Fprintf(&bb, `
    var val = 0

    for _, part := range strings.Split(string(b), %q) {
      switch part {`, repr.flag_sep)

		for _, f := range repr.fields {
			fmt.Fprintf(&bb, `
      case %q:
        val &= %d`, f.String, f.Value)
		}

		fmt.Fprintf(&bb, `
    //  default:
        // log.Printf("Unexpected value: %%q while unmarshaling %sEnum\n", part)
      }
    }

    self.value = %s(val)
    return nil
  `, repr.name, intType)
	}

	if repr.flags&jsonUnmarshalIsString == jsonUnmarshalIsString {
		fmt.Fprintf(buf, `
func (self *%sEnum) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {%s
	}

  if %t {
%s  }

	return nil // fmt.Errorf("Invalid enum string value: %%s\n", b)
}
`, repr.name, b.String(), bf, bb.String())

	} else {
		fmt.Fprintf(buf, `
func (self *%sEnum) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	self.value = %s(n)
	return nil
}
`, repr.name, intType)
	}
}

func doBitFuncs(buf *bytes.Buffer, repr *EnumRepr) {
	fmt.Fprintf(buf, `
func (self %sEnum) Add(v %sEnum) %sEnum {
	self.value |= v.value
	return self
}

func (self %sEnum) AddAll(v ...%sEnum) %sEnum {
	for _, item := range v {
		self.value |= item.value
	}
	return self
}

func (self %sEnum) Remove(v %sEnum) %sEnum {
	self.value &^= v.value
	return self
}

func (self %sEnum) RemoveAll(v ...%sEnum) %sEnum {
	for _, item := range v {
		self.value &^= item.value
	}
	return self
}

func (self %sEnum) Has(v %sEnum) bool {
	return self.value&v.value == v.value
}

func (self %sEnum) HasAny(v ...%sEnum) bool {
	for _, item := range v {
		if self.value&item.value == item.value {
			return true
		}
	}
	return false
}

func (self %sEnum) HasAll(v ...%sEnum) bool {
	for _, item := range v {
		if self.value&item.value != item.value {
			return false
		}
	}
	return true
}
`, repr.name, repr.name, repr.name,
		repr.name, repr.name, repr.name,
		repr.name, repr.name, repr.name,
		repr.name, repr.name, repr.name,
		repr.name, repr.name,
		repr.name, repr.name,
		repr.name, repr.name)
}
