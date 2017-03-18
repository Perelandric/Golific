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

type Resident struct {
	Name string
	Pet  AnimalEnum // The generated type for your Animal enum
}

func main() {
	res := Resident{
		Name: "Charlie Brown",
		Pet:  Animal.Dog, // Assign one of the variants
	}

	// The `json:"string"` option we included causes our custom `gString` value to be used when marshaled as JSON data
	j, err := json.Marshal(&res)

	fmt.Printf("%s %v\n", j, err) // {"Name":"Charlie Brown","Pet":"doggie"} <nil>

	// Enumerate all the variants in a range loop
	for _, animal := range Animal.Values {
		fmt.Printf("Kind: %s, Description: %q\n", animal, animal.Description())
	}
}
