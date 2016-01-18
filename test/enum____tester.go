package test

import (
	"log"
	"strconv"
	"strings"
)

/*****************************

FooEnum - bit flags

******************************/

type FooEnum struct{ value_p6kpk974c9pm uint8 }

var Foo = struct {
	Bar FooEnum
	Baz FooEnum
	Buz FooEnum
}{
	Bar: FooEnum{value_p6kpk974c9pm: 1},
	Baz: FooEnum{value_p6kpk974c9pm: 2},
	Buz: FooEnum{value_p6kpk974c9pm: 4},
}

// Used to iterate in range loops
var foobar = [...]FooEnum{
	Foo.Bar, Foo.Baz, Foo.Buz,
}

// Get the integer value of the enum variant
func (self FooEnum) Value() uint8 {
	return self.value_p6kpk974c9pm
}

func (self FooEnum) IntValue() int {
	return int(self.value_p6kpk974c9pm)
}

// Get the string representation of the enum variant
func (self FooEnum) String() string {
	switch self.value_p6kpk974c9pm {
	case 1:
		return "bar"
	case 2:
		return "baz"
	case 4:
		return "Buz"
	}

	if self.value_p6kpk974c9pm == 0 {
		return ""
	}

	var vals = make([]string, 0, 3/2)

	for _, item := range foobar {
		if self.value_p6kpk974c9pm&item.value_p6kpk974c9pm == item.value_p6kpk974c9pm {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, ",")
}

// Get the string description of the enum variant
func (self FooEnum) Description() string {
	switch self.value_p6kpk974c9pm {
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
		self.value_p6kpk974c9pm = 1
		return nil
	case "baz":
		self.value_p6kpk974c9pm = 2
		return nil
	case "Buz":
		self.value_p6kpk974c9pm = 4
		return nil
	}

	var val = 0

	for _, part := range strings.Split(string(b), ",") {
		switch part {
		case "bar":
			val |= 1
		case "baz":
			val |= 2
		case "Buz":
			val |= 4
		default:
			log.Printf("Unexpected value: %q while unmarshaling FooEnum\n", part)
		}
	}

	self.value_p6kpk974c9pm = uint8(val)
	return nil
}

// Bitflag enum methods
func (self FooEnum) Add(v FooEnum) FooEnum {
	self.value_p6kpk974c9pm |= v.value_p6kpk974c9pm
	return self
}

func (self FooEnum) AddAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value_p6kpk974c9pm |= item.value_p6kpk974c9pm
	}
	return self
}

func (self FooEnum) Remove(v FooEnum) FooEnum {
	self.value_p6kpk974c9pm &^= v.value_p6kpk974c9pm
	return self
}

func (self FooEnum) RemoveAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value_p6kpk974c9pm &^= item.value_p6kpk974c9pm
	}
	return self
}

func (self FooEnum) Has(v FooEnum) bool {
	return self.value_p6kpk974c9pm&v.value_p6kpk974c9pm == v.value_p6kpk974c9pm
}

func (self FooEnum) HasAny(v ...FooEnum) bool {
	for _, item := range v {
		if self.value_p6kpk974c9pm&item.value_p6kpk974c9pm == item.value_p6kpk974c9pm {
			return true
		}
	}
	return false
}

func (self FooEnum) HasAll(v ...FooEnum) bool {
	for _, item := range v {
		if self.value_p6kpk974c9pm&item.value_p6kpk974c9pm != item.value_p6kpk974c9pm {
			return false
		}
	}
	return true
}

/*****************************

OofEnum

******************************/

type OofEnum struct{ value_1404ufk3hj2w uint8 }

var Oof = struct {
	Bar OofEnum
	Baz OofEnum
	Buz OofEnum
}{
	Bar: OofEnum{value_1404ufk3hj2w: 1},
	Baz: OofEnum{value_1404ufk3hj2w: 123},
	Buz: OofEnum{value_1404ufk3hj2w: 3},
}

// Used to iterate in range loops
var OofValues = [...]OofEnum{
	Oof.Bar, Oof.Baz, Oof.Buz,
}

// Get the integer value of the enum variant
func (self OofEnum) Value() uint8 {
	return self.value_1404ufk3hj2w
}

func (self OofEnum) IntValue() int {
	return int(self.value_1404ufk3hj2w)
}

// Get the string representation of the enum variant
func (self OofEnum) String() string {
	switch self.value_1404ufk3hj2w {
	case 1:
		return "bar"
	case 123:
		return "Baz"
	case 3:
		return "Buz"
	}

	return ""
}

// Get the string description of the enum variant
func (self OofEnum) Description() string {
	switch self.value_1404ufk3hj2w {
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
	self.value_1404ufk3hj2w = uint8(n)
	return nil
}
