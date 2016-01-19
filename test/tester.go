package main

import (
	"encoding/json"
	"fmt"
)

//go:generate GoEnum $GOFILE

/*
@enum --name=Foo --bitflags --bitflag_separator="," --iterator_name="foobar" --json=string
Bar --string=bar
Baz --string=baz --description="This is the description"
Buz

@enum --name=Oof
Bar --string="bar"
Baz --value = 123
Buz --description="Some description"
*/

/*
@enum --name=Animal --json=string
Dog --string=dog --description="Your best friend, and you know it."
Cat --string=cat --description="Your best friend, but doesn't always show it."
Horse --string=horse --description="Everyone loves horses."
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
		Pet  AnimalEnum
	}

	res := Resident{
		Name: "Bart Simpson",
		Pet:  Animal.Dog,
	}

	jj, err := json.Marshal(&res)
	fmt.Printf("%s\n", jj) // {"Name":"Bart Simpson","Pet":"dog"}

	for _, animal := range AnimalValues {
		fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
	}
}
