package main

import (
	"encoding/json"
	"fmt"
)

//go:generate Golific $GOFILE

/*
@union
*/
type __TestUnion struct {
	FooEnum
	AnimalEnum
	Tester
}

/*
@union
*/
type __AnotherUnion struct {
	Tester
	OofEnum
}

/*
@enum
bitflags bitflag_separator:"," iterator_name:"foobar" json:"string"
*/
type __Foo struct {
	Bar int `gString:"bar"`
	Baz int `gString:"baz" gDescription:"This is the description"`
	Buz int
}

/*
@enum
*/
type __Oof struct {
	Bar int `gString:"bar"`
	Baz int `gValue:"123"`
	Buz int `gDescription:"Some description"`
}

/*
@enum
json:"string"
*/
// An enum to test the @enum descriptor
type __Animal struct {
	// Dog is a dog
	Dog int `gString:"doggy"
			gDescription:"Your best friend, and you know it."`

	// Cat is a cat
	Cat int `gDefault
			gString:"kitty"
			gDescription:"Your best friend, but doesn't always show it."`

	// Horse is a horse
	Horse int `gString:"horsie"
			gDescription:"Everyone loves horses."`
}

/*
@struct
*/
// A struct to test the @struct descriptor
//
// And another line or two for good measure
type __Tester struct {
	*AnimalEnum

	Arr []*AnimalEnum

	// The first item
	Test1 string `json:"tester1"`

	Test2 string `gRead`
	Test3 string `gWrite json:"tester3,omitempty"`

	// The fourth item
	Test4 string `gRead gWrite`
	Test5 string `gRead gWrite` /*`
	gRead
	gWrite
	gString:"the string"
	gDescription:"This is the description"`
	*/

	*FooEnum
}

type tester struct {
	F FooEnum
	O OofEnum
}

func main() {
	var t = tester{
		F: Foo.Baz,
		O: Oof.Buz,
	}

	var j, err = json.Marshal(&t)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(j))

	type Resident struct {
		Name string
		Pet  AnimalEnum // The generated type for your Animal enum
	}

	res := Resident{
		Name: "Charlie Brown",
		Pet:  Animal.Dog, // Assign one of the variants
	}

	// The `--json=string` flag causes our custom `--string` value to be used in the resulting JSON
	jj, err := json.Marshal(&res)
	fmt.Printf("%s\n", jj) // {"Name":"Charlie Brown","Pet":"doggie"}

	// Enumerate all the variants in a range loop
	for _, animal := range Animal.Values {
		fmt.Printf("Kind: %s, Description: %q, Type: %q, Namespace: %q\n", animal, animal.Description(), animal.Type(), animal.Namespace())
	}
}
