package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("enum: ")

	var data EnumData

	for _, file := range os.Args[1:] {
		fmt.Printf("Processing file: %q\n", file)
		data.DoFile(file)
	}
}
