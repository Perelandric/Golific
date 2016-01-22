package main

import (
	"encoding/json"
	"fmt"
)

//go:generate GoEnum $GOFILE

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
@enum Animal --json=string --summary
Dog --string=doggy
		--description="Your best friend, and you know it."
Cat --string=kitty --description="Your best friend, but doesn't always show it."
Horse --string=horsie --description="Everyone loves horses."
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
		fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
	}
}
