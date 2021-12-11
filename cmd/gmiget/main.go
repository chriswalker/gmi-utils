package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/chriswalker/gmi-utils/gemini"
)

const usage = `
gmiget - gets Gemini pages

Usage:
    gmiget [flags...] <url>

Options:
    -h, --help     Show help for gmiget
    -I             Print response status code only

`

var (
	help       bool
	statusOnly bool
)

func main() {
	flag.BoolVar(&help, "help", false, "Show help for gmiget")
	flag.BoolVar(&help, "h", false, "Show help for gmiget")
	flag.BoolVar(&statusOnly, "I", false, "Output response header only")
	flag.Usage = func() {
		fmt.Print(usage)
		os.Exit(1)
	}
	flag.Parse()

	if help {
		fmt.Print(usage)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "gmiget: missing Gemini URL")
		flag.Usage()
	}
	geminiURL := args[0]

	resp, err := gemini.Get(geminiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gmiget: could not open URL: %s\n", err)
		os.Exit(1)
	}

	output := string(resp.Body)
	if statusOnly {
		output = gemini.Status(resp.StatusCode)
	}
	fmt.Printf("%s\n", output)
}
