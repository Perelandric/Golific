package main

import (
	"log"
	"strconv"
	"strings"
)

/*****************************

FooEnum - bit flags

******************************/

type FooEnum struct{ value_hb36h415ag9i uint8 }

var Foo = struct {
	Bar FooEnum
	Baz FooEnum
	Buz FooEnum
}{
	Bar: FooEnum{value_hb36h415ag9i: 1},
	Baz: FooEnum{value_hb36h415ag9i: 2},
	Buz: FooEnum{value_hb36h415ag9i: 4},
}

// Used to iterate in range loops
var foobar = [...]FooEnum{
	Foo.Bar, Foo.Baz, Foo.Buz,
}

// Get the integer value of the enum variant
func (Fe FooEnum) Value() uint8 {
	return Fe.value_hb36h415ag9i
}

func (Fe FooEnum) IntValue() int {
	return int(Fe.value_hb36h415ag9i)
}

// Get the string representation of the enum variant
func (Fe FooEnum) String() string {
	switch Fe.value_hb36h415ag9i {
	case 1:
		return "bar"
	case 2:
		return "baz"
	case 4:
		return "Buz"
	}

	if Fe.value_hb36h415ag9i == 0 {
		return ""
	}

	var vals = make([]string, 0, 3/2)

	for _, item := range foobar {
		if Fe.value_hb36h415ag9i&item.value_hb36h415ag9i == item.value_hb36h415ag9i {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, ",")
}

// Get the string description of the enum variant
func (Fe FooEnum) Description() string {
	switch Fe.value_hb36h415ag9i {
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
func (Fe FooEnum) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(Fe.String())), nil
}

func (Fe *FooEnum) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {
	case "bar":
		Fe.value_hb36h415ag9i = 1
		return nil
	case "baz":
		Fe.value_hb36h415ag9i = 2
		return nil
	case "Buz":
		Fe.value_hb36h415ag9i = 4
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

	Fe.value_hb36h415ag9i = uint8(val)
	return nil
}

// XML marshaling methods to come
// Bitflag enum methods
func (Fe FooEnum) Add(v FooEnum) FooEnum {
	Fe.value_hb36h415ag9i |= v.value_hb36h415ag9i
	return Fe
}

func (Fe FooEnum) AddAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		Fe.value_hb36h415ag9i |= item.value_hb36h415ag9i
	}
	return Fe
}

func (Fe FooEnum) Remove(v FooEnum) FooEnum {
	Fe.value_hb36h415ag9i &^= v.value_hb36h415ag9i
	return Fe
}

func (Fe FooEnum) RemoveAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		Fe.value_hb36h415ag9i &^= item.value_hb36h415ag9i
	}
	return Fe
}

func (Fe FooEnum) Has(v FooEnum) bool {
	return Fe.value_hb36h415ag9i&v.value_hb36h415ag9i == v.value_hb36h415ag9i
}

func (Fe FooEnum) HasAny(v ...FooEnum) bool {
	for _, item := range v {
		if Fe.value_hb36h415ag9i&item.value_hb36h415ag9i == item.value_hb36h415ag9i {
			return true
		}
	}
	return false
}

func (Fe FooEnum) HasAll(v ...FooEnum) bool {
	for _, item := range v {
		if Fe.value_hb36h415ag9i&item.value_hb36h415ag9i != item.value_hb36h415ag9i {
			return false
		}
	}
	return true
}

/*****************************

OofEnum

******************************/

type OofEnum struct{ value_148yiuu7sigrx uint8 }

var Oof = struct {
	Bar OofEnum
	Baz OofEnum
	Buz OofEnum
}{
	Bar: OofEnum{value_148yiuu7sigrx: 1},
	Baz: OofEnum{value_148yiuu7sigrx: 123},
	Buz: OofEnum{value_148yiuu7sigrx: 3},
}

// Used to iterate in range loops
var OofValues = [...]OofEnum{
	Oof.Bar, Oof.Baz, Oof.Buz,
}

// Get the integer value of the enum variant
func (Oe OofEnum) Value() uint8 {
	return Oe.value_148yiuu7sigrx
}

func (Oe OofEnum) IntValue() int {
	return int(Oe.value_148yiuu7sigrx)
}

// Get the string representation of the enum variant
func (Oe OofEnum) String() string {
	switch Oe.value_148yiuu7sigrx {
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
func (Oe OofEnum) Description() string {
	switch Oe.value_148yiuu7sigrx {
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
func (Oe OofEnum) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(Oe.IntValue())), nil
}

func (Oe *OofEnum) UnmarshalJSON(b []byte) error {
	var n, err = strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	Oe.value_148yiuu7sigrx = uint8(n)
	return nil
}

// XML marshaling methods to come

/*****************************

AnimalEnum

******************************/

type AnimalEnum struct{ value_18hsg8max2cj2 uint8 }

var Animal = struct {
	Dog   AnimalEnum
	Cat   AnimalEnum
	Horse AnimalEnum
}{
	Dog:   AnimalEnum{value_18hsg8max2cj2: 1},
	Cat:   AnimalEnum{value_18hsg8max2cj2: 2},
	Horse: AnimalEnum{value_18hsg8max2cj2: 3},
}

// Used to iterate in range loops
var AnimalValues = [...]AnimalEnum{
	Animal.Dog, Animal.Cat, Animal.Horse,
}

// Get the integer value of the enum variant
func (Ae AnimalEnum) Value() uint8 {
	return Ae.value_18hsg8max2cj2
}

func (Ae AnimalEnum) IntValue() int {
	return int(Ae.value_18hsg8max2cj2)
}

// Get the string representation of the enum variant
func (Ae AnimalEnum) String() string {
	switch Ae.value_18hsg8max2cj2 {
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
func (Ae AnimalEnum) Description() string {
	switch Ae.value_18hsg8max2cj2 {
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
func (Ae AnimalEnum) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(Ae.String())), nil
}

func (Ae *AnimalEnum) UnmarshalJSON(b []byte) error {
	var s, err = strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return nil
	}

	switch s {
	case "doggy":
		Ae.value_18hsg8max2cj2 = 1
		return nil
	case "kitty":
		Ae.value_18hsg8max2cj2 = 2
		return nil
	case "horsie":
		Ae.value_18hsg8max2cj2 = 3
		return nil
	default:
		log.Printf("Unexpected value: %q while unmarshaling AnimalEnum\n", s)
	}

	return nil
}

// XML marshaling methods to come
