package main

import (
	"flag"
	"log"
	"os"

	"novit.nc/direktil/initrd/cpiocat"
)

func main() {
	flag.Parse()

	err := cpiocat.Append(os.Stdout, os.Stdin, flag.Args())
	if err != nil {
		log.Fatal(err)
	}
}
