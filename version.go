package main

import (
	"fmt"
	"os"
)

var version = "v0.0.3"

func ShowVersion() {
	fmt.Fprintf(os.Stderr, "%s version %s\n", Options.Command, version)
}
