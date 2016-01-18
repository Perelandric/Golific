package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("enum: ")

	rand.Seed(time.Now().UnixNano())

	var data EnumData

	for _, file := range os.Args[1:] {
		fmt.Printf("Processing file: %q\n", file)
		data.DoFile(file)
	}
}
