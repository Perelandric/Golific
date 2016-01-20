package main

import (
	"log"
	"strconv"
	"strings"
)

/*****************************

FooEnum - bit flags

******************************/

type FooEnum struct{ value_faqb48c0j3ju uint8 }

var Foo = struct {
	Bar FooEnum
	Baz FooEnum
	Buz FooEnum
}{
	Bar: FooEnum{value_faqb48c0j3ju: 1},
	Baz: FooEnum{value_faqb48c0j3ju: 2},
	Buz: FooEnum{value_faqb48c0j3ju: 4},
}

// Used to iterate in range loops
var foobar = [...]FooEnum{
	Foo.Bar, Foo.Baz, Foo.Buz,
}

// Get the integer value of the enum variant
func (self FooEnum) Value() uint8 {
	return self.value_faqb48c0j3ju
}

func (self FooEnum) IntValue() int {
	return int(self.value_faqb48c0j3ju)
}

// Get the string representation of the enum variant
func (self FooEnum) String() string {
	switch self.value_faqb48c0j3ju {
	case 1:
		return "bar"
	case 2:
		return "baz"
	case 4:
		return "Buz"
	}

	if self.value_faqb48c0j3ju == 0 {
		return ""
	}

	var vals = make([]string, 0, 3/2)

	for _, item := range foobar {
		if self.value_faqb48c0j3ju&item.value_faqb48c0j3ju == item.value_faqb48c0j3ju {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, ",")
}

// Get the string description of the enum variant
func (self FooEnum) Description() string {
	switch self.value_faqb48c0j3ju {
	case 1:
		return "bar"
	case 2:
		return "This is the description"
	case 4:
		return "Buz"
	}
	return ""
}

// JSON marshaling methods
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
		self.value_faqb48c0j3ju = 1
		return nil
	case "baz":
		self.value_faqb48c0j3ju = 2
		return nil
	case "Buz":
		self.value_faqb48c0j3ju = 4
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

	self.value_faqb48c0j3ju = uint8(val)
	return nil
}

// Bitflag enum methods
func (self FooEnum) Add(v FooEnum) FooEnum {
	self.value_faqb48c0j3ju |= v.value_faqb48c0j3ju
	return self
}

func (self FooEnum) AddAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value_faqb48c0j3ju |= item.value_faqb48c0j3ju
	}
	return self
}

func (self FooEnum) Remove(v FooEnum) FooEnum {
	self.value_faqb48c0j3ju &^= v.value_faqb48c0j3ju
	return self
}

func (self FooEnum) RemoveAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value_faqb48c0j3ju &^= item.value_faqb48c0j3ju
	}
	return self
}

func (self FooEnum) Has(v FooEnum) bool {
	return self.value_faqb48c0j3ju&v.value_faqb48c0j3ju == v.value_faqb48c0j3ju
}

func (self FooEnum) HasAny(v ...FooEnum) bool {
	for _, item := range v {
		if self.value_faqb48c0j3ju&item.value_faqb48c0j3ju == item.value_faqb48c0j3ju {
			return true
		}
	}
	return false
}

func (self FooEnum) HasAll(v ...FooEnum) bool {
	for _, item := range v {
		if self.value_faqb48c0j3ju&item.value_faqb48c0j3ju != item.value_faqb48c0j3ju {
			return false
		}
	}
	return true
}

/*****************************

OofEnum

******************************/

type OofEnum struct{ value_7ngc4szyjxi0 uint8 }

var Oof = struct {
	Bar OofEnum
	Baz OofEnum
	Buz OofEnum
}{
	Bar: OofEnum{value_7ngc4szyjxi0: 1},
	Baz: OofEnum{value_7ngc4szyjxi0: 123},
	Buz: OofEnum{value_7ngc4szyjxi0: 3},
}

// Used to iterate in range loops
var OofValues = [...]OofEnum{
	Oof.Bar, Oof.Baz, Oof.Buz,
}

// Get the integer value of the enum variant
func (self OofEnum) Value() uint8 {
	return self.value_7ngc4szyjxi0
}

func (self OofEnum) IntValue() int {
	return int(self.value_7ngc4szyjxi0)
}

// Get the string representation of the enum variant
func (self OofEnum) String() string {
	switch self.value_7ngc4szyjxi0 {
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
	switch self.value_7ngc4szyjxi0 {
	case 1:
		return "bar"
	case 123:
		return "Baz"
	case 3:
		return "Some description"
	}
	return ""
}

// JSON marshaling methods
func (self OofEnum) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(self.IntValue())), nil
}

func (self *OofEnum) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	self.value_7ngc4szyjxi0 = uint8(n)
	return nil
}

/*****************************

AnimalEnum

******************************/

type AnimalEnum struct{ value_326b4iofg5bz uint8 }

var Animal = struct {
	Dog   AnimalEnum
	Cat   AnimalEnum
	Horse AnimalEnum
}{
	Dog:   AnimalEnum{value_326b4iofg5bz: 1},
	Cat:   AnimalEnum{value_326b4iofg5bz: 2},
	Horse: AnimalEnum{value_326b4iofg5bz: 3},
}

// Used to iterate in range loops
var AnimalValues = [...]AnimalEnum{
	Animal.Dog, Animal.Cat, Animal.Horse,
}

// Get the integer value of the enum variant
func (self AnimalEnum) Value() uint8 {
	return self.value_326b4iofg5bz
}

func (self AnimalEnum) IntValue() int {
	return int(self.value_326b4iofg5bz)
}

// Get the string representation of the enum variant
func (self AnimalEnum) String() string {
	switch self.value_326b4iofg5bz {
	case 1:
		return "doggy"
	case 2:
		return "kitty"
	case 3:
		return "horsie"
	}

	return ""
}

// Get the string description of the enum variant
func (self AnimalEnum) Description() string {
	switch self.value_326b4iofg5bz {
	case 1:
		return "Your best friend, and you know it."
	case 2:
		return "Your best friend, but doesn't always show it."
	case 3:
		return "Everyone loves horses."
	}
	return ""
}

// JSON marshaling methods
func (self AnimalEnum) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(self.String())), nil
}

func (self *AnimalEnum) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {
	case "doggy":
		self.value_326b4iofg5bz = 1
		return nil
	case "kitty":
		self.value_326b4iofg5bz = 2
		return nil
	case "horsie":
		self.value_326b4iofg5bz = 3
		return nil
	default:
		log.Printf("Unexpected value: %q while unmarshaling AnimalEnum\n", s)
	}

	return nil
}
