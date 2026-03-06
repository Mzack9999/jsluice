package main

import (
	"strings"

	"github.com/Mzack9999/jsluice"
)

func printTree(opts options, filename string, source []byte, output chan string, errs chan error) {

	buf := strings.Builder{}
	buf.WriteString(filename + ":\n")

	buf.WriteString(jsluice.PrintTree(source))

	output <- buf.String()
}
