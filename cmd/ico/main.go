package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gonutz/ico"
)

func usage() {
	fmt.Println(`usage: ico IMAGE-FILE [OUTPUT-FILE]
  provide an input image (BMP, JPEG, PNG or GIF)
  optionally provide an output file (ICO)
  if no output file is given the input path will be used with extension .ico`)
}

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "not enough arguments")
		usage()
		return 2
	}
	if len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "too many arguments")
		usage()
		return 2
	}
	in := os.Args[1]
	var out string
	if len(os.Args) == 3 {
		out = os.Args[2]
	} else {
		ext := filepath.Ext(in)
		out = strings.TrimSuffix(in, ext) + ".ico"
	}

	f, err := os.Open(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open input file:", err)
		return 2
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot read input image: ", err)
		return 2
	}

	icon := ico.FromImage(img)
	err = ioutil.WriteFile(out, icon, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot write output file:", err)
		return 2
	}

	return 0
}
