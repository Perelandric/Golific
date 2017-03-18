package main

import (
	"encoding/json"
	"fmt"
)

//go:generate Golific $GOFILE

/*
@enum json:"string"
*/
type __Animal struct {
	Dog   int `gString:"doggie" gDescription:"Loves to lick your face"`
	Cat   int `gString:"kitty" gDescription:"Loves to scratch your face"`
	Horse int `gString:"horsie" gDescription:"Has a very long face"`
}

// Use the resulting AnimalEnum in your code
type Resident struct {
	Name string
	Pet  AnimalEnum
}

func main() {
	res := Resident{
		Name: "Charlie Brown",
		Pet:  Animal.Dog, // Use the Animal namespace to assign a variant
	}

	// The `json:"string"` option causes our `gString` value to be used when marshaled as JSON
	j, err := json.Marshal(&res)

	fmt.Printf("%s %v\n", j, err) // {"Name":"Charlie Brown","Pet":"doggie"} <nil>

	// Enumerate all the variants in a range loop
	for _, animal := range Animal.Values {
		fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
	}
}
