package main

import (
	"log"
	"strconv"
	"strings"
)

/*****************************

FooEnum - bit flags

******************************/

type FooEnum struct{ value_19bh2vnljbmq1 uint8 }

var Foo = struct {
	Bar FooEnum
	Baz FooEnum
	Buz FooEnum

	// Used to iterate in range loops
	foobar [3]FooEnum
}{
	Bar: FooEnum{value_19bh2vnljbmq1: 1},
	Baz: FooEnum{value_19bh2vnljbmq1: 2},
	Buz: FooEnum{value_19bh2vnljbmq1: 4},
}

func init() {
	Foo.foobar = [3]FooEnum{
		Foo.Bar, Foo.Baz, Foo.Buz,
	}
}

// Get the integer value of the enum variant
func (Fe FooEnum) Value() uint8 {
	return Fe.value_19bh2vnljbmq1
}

func (Fe FooEnum) IntValue() int {
	return int(Fe.value_19bh2vnljbmq1)
}

// Name returns the name of the variant as a string.
func (Fe FooEnum) Name() string {
	switch Fe.value_19bh2vnljbmq1 {
	case 1:
		return "Bar"
	case 2:
		return "Baz"
	case 4:
		return "Buz"
	}

	return ""
}

// String returns the given string value of the variant. If none has been set,
// its return value is as though 'Name()' had been called.
// If multiple bit values are assigned, the string values will be joined into a
// single string using "," as a separator.
func (Fe FooEnum) String() string {
	switch Fe.value_19bh2vnljbmq1 {
	case 1:
		return "bar"
	case 2:
		return "baz"
	case 4:
		return "Buz"
	}

	if Fe.value_19bh2vnljbmq1 == 0 {
		return ""
	}

	var vals = make([]string, 0, 3/2)

	for _, item := range Foo.foobar {
		if Fe.value_19bh2vnljbmq1&item.value_19bh2vnljbmq1 == item.value_19bh2vnljbmq1 {
			vals = append(vals, item.String())
		}
	}
	return strings.Join(vals, ",")
}

// Description returns the description of the variant. If none has been set, its
// return value is as though 'String()' had been called.
func (Fe FooEnum) Description() string {
	switch Fe.value_19bh2vnljbmq1 {
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
		Fe.value_19bh2vnljbmq1 = 1
		return nil
	case "baz":
		Fe.value_19bh2vnljbmq1 = 2
		return nil
	case "Buz":
		Fe.value_19bh2vnljbmq1 = 4
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

	Fe.value_19bh2vnljbmq1 = uint8(val)
	return nil
}

// XML marshaling methods to come
// Bitflag enum methods
func (Fe FooEnum) Add(v FooEnum) FooEnum {
	Fe.value_19bh2vnljbmq1 |= v.value_19bh2vnljbmq1
	return Fe
}

func (Fe FooEnum) AddAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		Fe.value_19bh2vnljbmq1 |= item.value_19bh2vnljbmq1
	}
	return Fe
}

func (Fe FooEnum) Remove(v FooEnum) FooEnum {
	Fe.value_19bh2vnljbmq1 &^= v.value_19bh2vnljbmq1
	return Fe
}

func (Fe FooEnum) RemoveAll(v ...FooEnum) FooEnum {
	for _, item := range v {
		Fe.value_19bh2vnljbmq1 &^= item.value_19bh2vnljbmq1
	}
	return Fe
}

func (Fe FooEnum) Has(v FooEnum) bool {
	return Fe.value_19bh2vnljbmq1&v.value_19bh2vnljbmq1 == v.value_19bh2vnljbmq1
}

func (Fe FooEnum) HasAny(v ...FooEnum) bool {
	for _, item := range v {
		if Fe.value_19bh2vnljbmq1&item.value_19bh2vnljbmq1 == item.value_19bh2vnljbmq1 {
			return true
		}
	}
	return false
}

func (Fe FooEnum) HasAll(v ...FooEnum) bool {
	for _, item := range v {
		if Fe.value_19bh2vnljbmq1&item.value_19bh2vnljbmq1 != item.value_19bh2vnljbmq1 {
			return false
		}
	}
	return true
}

/*****************************

OofEnum

******************************/

type OofEnum struct{ value_aa5rvdo9g1iu uint8 }

var Oof = struct {
	Bar OofEnum
	Baz OofEnum
	Buz OofEnum

	// Used to iterate in range loops
	Values [3]OofEnum
}{
	Bar: OofEnum{value_aa5rvdo9g1iu: 1},
	Baz: OofEnum{value_aa5rvdo9g1iu: 123},
	Buz: OofEnum{value_aa5rvdo9g1iu: 3},
}

func init() {
	Oof.Values = [3]OofEnum{
		Oof.Bar, Oof.Baz, Oof.Buz,
	}
}

// Get the integer value of the enum variant
func (Oe OofEnum) Value() uint8 {
	return Oe.value_aa5rvdo9g1iu
}

func (Oe OofEnum) IntValue() int {
	return int(Oe.value_aa5rvdo9g1iu)
}

// Name returns the name of the variant as a string.
func (Oe OofEnum) Name() string {
	switch Oe.value_aa5rvdo9g1iu {
	case 1:
		return "Bar"
	case 123:
		return "Baz"
	case 3:
		return "Buz"
	}

	return ""
}

// String returns the given string value of the variant. If none has been set,
// its return value is as though 'Name()' had been called.

func (Oe OofEnum) String() string {
	switch Oe.value_aa5rvdo9g1iu {
	case 1:
		return "bar"
	case 123:
		return "Baz"
	case 3:
		return "Buz"
	}

	return ""
}

// Description returns the description of the variant. If none has been set, its
// return value is as though 'String()' had been called.
func (Oe OofEnum) Description() string {
	switch Oe.value_aa5rvdo9g1iu {
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
	Oe.value_aa5rvdo9g1iu = uint8(n)
	return nil
}

// XML marshaling methods to come

/*****************************

AnimalEnum

******************************/

type AnimalEnum struct{ value_1k9dwobfqc99t uint8 }

var Animal = struct {
	Dog   AnimalEnum
	Cat   AnimalEnum
	Horse AnimalEnum

	// Used to iterate in range loops
	Values [3]AnimalEnum
}{
	Dog:   AnimalEnum{value_1k9dwobfqc99t: 1},
	Cat:   AnimalEnum{value_1k9dwobfqc99t: 2},
	Horse: AnimalEnum{value_1k9dwobfqc99t: 3},
}

func init() {
	Animal.Values = [3]AnimalEnum{
		Animal.Dog, Animal.Cat, Animal.Horse,
	}
}

// Get the integer value of the enum variant
func (Ae AnimalEnum) Value() uint8 {
	return Ae.value_1k9dwobfqc99t
}

func (Ae AnimalEnum) IntValue() int {
	return int(Ae.value_1k9dwobfqc99t)
}

// Name returns the name of the variant as a string.
func (Ae AnimalEnum) Name() string {
	switch Ae.value_1k9dwobfqc99t {
	case 1:
		return "Dog"
	case 2:
		return "Cat"
	case 3:
		return "Horse"
	}

	return ""
}

// String returns the given string value of the variant. If none has been set,
// its return value is as though 'Name()' had been called.

func (Ae AnimalEnum) String() string {
	switch Ae.value_1k9dwobfqc99t {
	case 1:
		return "doggy"
	case 2:
		return "kitty"
	case 3:
		return "horsie"
	}

	return ""
}

// Description returns the description of the variant. If none has been set, its
// return value is as though 'String()' had been called.
func (Ae AnimalEnum) Description() string {
	switch Ae.value_1k9dwobfqc99t {
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
		Ae.value_1k9dwobfqc99t = 1
		return nil
	case "kitty":
		Ae.value_1k9dwobfqc99t = 2
		return nil
	case "horsie":
		Ae.value_1k9dwobfqc99t = 3
		return nil
	default:
		log.Printf("Unexpected value: %q while unmarshaling AnimalEnum\n", s)
	}

	return nil
}

// XML marshaling methods to come
