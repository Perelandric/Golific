package main

import (
	"encoding/json"
	"fmt"
)

//go:generate Golific $GOFILE

/*
@union
*/
/*
type __TestUnion struct {
	FooEnum
	AnimalEnum
	Tester
}
*/

/*
@union
*/
/*
type AnotherUnion struct {
	Tester
	OofEnum
}
*/

/*
@enum
bitflags bitflag_separator:"," iterator_name="foobar" json:"string"
*/
type Foo struct {
	Bar int `string:bar`
	Baz int `string:baz description:"This is the description"`
	Buz int
}

/*
@enum-defaults summary
*/

/*
@enum
*/
type Oof struct {
	Bar int `string:"bar"`
	Baz int `value:123`
	Buz int `description:"Some description"`
}


/*
@enum
json:"string"
*/
// An enum to test the @enum descriptor
type Animal struct {
	// Dog is a dog
	Dog int `string:"doggy"
			description:"Your best friend, and you know it."`

	// Cat is a cat
	Cat int `string:"kitty"
			description:"Your best friend, but doesn't always show it."
			default`

	// Horse is a horse (of course)
	Horse int `string:"horsie" description:"Everyone loves horses."`
}


/*
@struct
*/
// A struct to test the @struct descriptor
//
// And another line or two for good measure
type Tester struct {
	*AnimalEnum

	// The first item
	Test1 string `json:"tester1"`

	Test2 string `read`
	Test3 string `write json:"tester3,omitempty"`

	// The fourth item
	Test4 string `read write`
	Test5 string `read write`
	*FooEnum
}

/*
@struct
*/
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
