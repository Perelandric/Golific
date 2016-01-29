package main

import (
	"encoding/json"
	"fmt"
)

//go:generate Golific $GOFILE

/*
@enum Foo --bitflags --bitflag_separator="," --iterator_name="foobar" --json=string --summary
Bar --string=bar
Baz --string=baz --description="This is the description"
Buz

@enum Oof
Bar --string="bar"
Baz --value = 123
Buz --description="Some description"
*/

/*
// An enum to test the @enum descriptor
@enum Animal --json=string --summary

// Dog is a dog
Dog --string=doggy
		--description="Your best friend, and you know it."

// Cat is a cat
Cat --string=kitty --description="Your best friend, but doesn't always show it."

// Horse is a horse (of course)
Horse --string=horsie --description="Everyone loves horses."

*/

/*
// A struct to test the @struct descriptor
//
// And another line or two for good measure
@struct Tester
	*AnimalEnum --default_expr="&Animal.Dog"

	// The first item
	Test1 string `json:"test1"`

	Test2 string --read --default_expr=`"foo"`
	Test3 string --write

	// The fourth item
	Test4 string --read --write --default_expr=`"bar"`
	Test5 string --read --write
	*FooEnum "whatever"
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
