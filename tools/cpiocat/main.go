package main

import (
	"flag"
	"os"

	"novit.nc/direktil/initrd/cpiocat"
)

func main() {
	flag.Parse()

	cpiocat.Append(os.Stdout, os.Stdin, flag.Args())
}
