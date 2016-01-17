package test

import (
  "strconv"
  "strings"
)

/*****************************

	FooEnum - bit flags

******************************/

type FooEnum struct{ value uint8 }

var Foo = struct {
  Bar FooEnum
  Baz FooEnum
  Buz FooEnum
}{
  Bar: FooEnum{1},
  Baz: FooEnum{2},
  Buz: FooEnum{4},
}

var foobar = [...]FooEnum{Foo.Bar, Foo.Baz, Foo.Buz, }

func (self FooEnum) Value() uint8 {
	return self.value
}
func (self FooEnum) IntValue() int {
	return int(self.value)
}

func (self FooEnum) String() string {
	switch self.value {
  case 1:
    return "bar"
  case 2:
    return "baz"
  case 4:
    return "Buz"
  }

	if self.value == 0 {
		return ""
	}

  var vals = make([]string, 0, 2)

  for _, item := range foobar {
    if self.value & item.value == item.value {
      vals = append(vals, item.String())
    }
  }
  return strings.Join(vals, ",")
}

func (self FooEnum) Description() string {
  switch self.value {
  case 1:
    return "bar"
  case 2:
    return "This is the description"
  case 4:
    return "Buz"
  }
  return ""
}

func (self FooEnum) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Quote(self.String())), nil
}

func (self *FooEnum) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {
  case "bar":
  	self.value = 1
  	return nil
  case "baz":
  	self.value = 2
  	return nil
  case "Buz":
  	self.value = 4
  	return nil
	}

  if true {

    var val = 0

    for _, part := range strings.Split(string(b), ",") {
      switch part {
      case "bar":
        val &= 1
      case "baz":
        val &= 2
      case "Buz":
        val &= 4
    //  default:
        // log.Printf("Unexpected value: %q while unmarshaling FooEnum\n", part)
      }
    }

    self.value = uint8(val)
    return nil
    }

	return nil // fmt.Errorf("Invalid enum string value: %s\n", b)
}

func (self FooEnum) Add(v FooEnum) FooEnum {
	self.value |= v.value
	return self
}

func (self FooEnum) AddAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value |= item.value
	}
	return self
}

func (self FooEnum) Remove(v FooEnum) FooEnum {
	self.value &^= v.value
	return self
}

func (self FooEnum) RemoveAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value &^= item.value
	}
	return self
}

func (self FooEnum) Has(v FooEnum) bool {
	return self.value&v.value == v.value
}

func (self FooEnum) HasAny(v ...FooEnum) bool {
	for _, item := range v {
		if self.value&item.value == item.value {
			return true
		}
	}
	return false
}

func (self FooEnum) HasAll(v ...FooEnum) bool {
	for _, item := range v {
		if self.value&item.value != item.value {
			return false
		}
	}
	return true
}

/*****************************

	OofEnum

******************************/

type OofEnum struct{ value int8 }

var Oof = struct {
  Bar OofEnum
  Baz OofEnum
  Buz OofEnum
}{
  Bar: OofEnum{0},
  Baz: OofEnum{1},
  Buz: OofEnum{2},
}

var OofValues = [...]OofEnum{Oof.Bar, Oof.Baz, Oof.Buz, }

func (self OofEnum) Value() int8 {
	return self.value
}
func (self OofEnum) IntValue() int {
	return int(self.value)
}

func (self OofEnum) String() string {
	switch self.value {
  case 1:
    return "bar"
  case 123:
    return "Baz"
  case 3:
    return "Buz"
  }
	return ""
}

func (self OofEnum) Description() string {
  switch self.value {
  case 1:
    return "bar"
  case 123:
    return "Baz"
  case 3:
    return "Some description"
  }
  return ""
}

func (self OofEnum) MarshalJSON() ([]byte, error) {
  return []byte(strconv.Itoa(self.IntValue())), nil
}

func (self *OofEnum) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	self.value = int8(n)
	return nil
}
