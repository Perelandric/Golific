package main

import (
	"log"
	"strconv"
	"strings"
)

/*****************************

FooEnum - bit flags

******************************/

type FooEnum struct{ value_4v1mir4wruhy uint8 }

var Foo = struct {
	Bar FooEnum
	Baz FooEnum
	Buz FooEnum
}{
	Bar: FooEnum{value_4v1mir4wruhy: 1},
	Baz: FooEnum{value_4v1mir4wruhy: 2},
	Buz: FooEnum{value_4v1mir4wruhy: 4},
}

// Used to iterate in range loops
var foobar = [...]FooEnum{
	Foo.Bar, Foo.Baz, Foo.Buz,
}

// Get the integer value of the enum variant
func (self FooEnum) Value() uint8 {
	return self.value_4v1mir4wruhy
}

func (self FooEnum) IntValue() int {
	return int(self.value_4v1mir4wruhy)
}

// Get the string representation of the enum variant
func (self FooEnum) String() string {
	switch self.value_4v1mir4wruhy {
	case 1:
		return "bar"
	case 2:
		return "baz"
	case 4:
		return "Buz"
	}

	if self.value_4v1mir4wruhy == 0 {
		return ""
	}

	var vals = make([]string, 0, 3/2)

	for _, item := range foobar {
		if self.value_4v1mir4wruhy&item.value_4v1mir4wruhy == item.value_4v1mir4wruhy {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, ",")
}

// Get the string description of the enum variant
func (self FooEnum) Description() string {
	switch self.value_4v1mir4wruhy {
	case 1:
		return "bar"
	case 2:
		return "This is the description"
	case 4:
		return "Buz"
	}
	return ""
}

// XML marshaling methods to come
// Bitflag enum methods
func (self FooEnum) Add(v FooEnum) FooEnum {
	self.value_4v1mir4wruhy |= v.value_4v1mir4wruhy
	return self
}

func (self FooEnum) AddAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value_4v1mir4wruhy |= item.value_4v1mir4wruhy
	}
	return self
}

func (self FooEnum) Remove(v FooEnum) FooEnum {
	self.value_4v1mir4wruhy &^= v.value_4v1mir4wruhy
	return self
}

func (self FooEnum) RemoveAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		self.value_4v1mir4wruhy &^= item.value_4v1mir4wruhy
	}
	return self
}

func (self FooEnum) Has(v FooEnum) bool {
	return self.value_4v1mir4wruhy&v.value_4v1mir4wruhy == v.value_4v1mir4wruhy
}

func (self FooEnum) HasAny(v ...FooEnum) bool {
	for _, item := range v {
		if self.value_4v1mir4wruhy&item.value_4v1mir4wruhy == item.value_4v1mir4wruhy {
			return true
		}
	}
	return false
}

func (self FooEnum) HasAll(v ...FooEnum) bool {
	for _, item := range v {
		if self.value_4v1mir4wruhy&item.value_4v1mir4wruhy != item.value_4v1mir4wruhy {
			return false
		}
	}
	return true
}

/*****************************

OofEnum

******************************/

type OofEnum struct{ value_n09eoz2ee3ka uint8 }

var Oof = struct {
	Bar OofEnum
	Baz OofEnum
	Buz OofEnum
}{
	Bar: OofEnum{value_n09eoz2ee3ka: 1},
	Baz: OofEnum{value_n09eoz2ee3ka: 123},
	Buz: OofEnum{value_n09eoz2ee3ka: 3},
}

// Used to iterate in range loops
var OofValues = [...]OofEnum{
	Oof.Bar, Oof.Baz, Oof.Buz,
}

// Get the integer value of the enum variant
func (self OofEnum) Value() uint8 {
	return self.value_n09eoz2ee3ka
}

func (self OofEnum) IntValue() int {
	return int(self.value_n09eoz2ee3ka)
}

// Get the string representation of the enum variant
func (self OofEnum) String() string {
	switch self.value_n09eoz2ee3ka {
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
	switch self.value_n09eoz2ee3ka {
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
	self.value_n09eoz2ee3ka = uint8(n)
	return nil
}

// XML marshaling methods to come

/*****************************

AnimalEnum

******************************/

type AnimalEnum struct{ value_16myto1zex81q uint8 }

var Animal = struct {
	Dog   AnimalEnum
	Cat   AnimalEnum
	Horse AnimalEnum
}{
	Dog:   AnimalEnum{value_16myto1zex81q: 1},
	Cat:   AnimalEnum{value_16myto1zex81q: 2},
	Horse: AnimalEnum{value_16myto1zex81q: 3},
}

// Used to iterate in range loops
var AnimalValues = [...]AnimalEnum{
	Animal.Dog, Animal.Cat, Animal.Horse,
}

// Get the integer value of the enum variant
func (self AnimalEnum) Value() uint8 {
	return self.value_16myto1zex81q
}

func (self AnimalEnum) IntValue() int {
	return int(self.value_16myto1zex81q)
}

// Get the string representation of the enum variant
func (self AnimalEnum) String() string {
	switch self.value_16myto1zex81q {
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
	switch self.value_16myto1zex81q {
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
		self.value_16myto1zex81q = 1
		return nil
	case "kitty":
		self.value_16myto1zex81q = 2
		return nil
	case "horsie":
		self.value_16myto1zex81q = 3
		return nil
	default:
		log.Printf("Unexpected value: %q while unmarshaling AnimalEnum\n", s)
	}

	return nil
}

// XML marshaling methods to come
