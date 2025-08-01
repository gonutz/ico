package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonutz/ico"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/gonutz/bmp"
)

func usage() {
	fmt.Println(`usage: ico IMAGE-FILE [OUTPUT-FILE]

    Converts the input image (BMP, JPEG, PNG or GIF) to an icon file (ICO).
    Saves to the given output file name. If no output file name is given, saves
    to the input file name with the extension changed to .ico.`)
}

func main() {
	os.Exit(run())
}

func run() int {
	args := os.Args[1:]
	if !(1 <= len(args) && len(args) <= 2) {
		fmt.Fprintln(os.Stderr, "wrong number of arguments")
		usage()
		return 2
	}

	input := args[0]
	output := strings.TrimSuffix(input, filepath.Ext(input)) + ".ico"
	if len(args) == 2 {
		output = args[1]
	}

	icon, err := ico.FromFile(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	err = os.WriteFile(output, icon, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output file '%s': %v\n", output, err)
		return 2
	}

	return 0
}
