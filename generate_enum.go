package main

import (
	"fmt"
	"generate_enum/generate_enum"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("enum: ")

	var data generate_enum.EnumData

	for _, file := range os.Args[1:] {
		fmt.Printf("Processing file: %q\n", file)
		data.DoFile(file)
	}
}
