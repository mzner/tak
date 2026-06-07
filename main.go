package main

import (
	"log"

	"github.com/mzner/tak/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Execute(version); err != nil {
		log.Fatal(err)
	}
}
