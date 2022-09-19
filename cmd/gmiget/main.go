package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chriswalker/gmi-utils/cli"
	"github.com/chriswalker/gmi-utils/gemini"
)

const (
	desc  = "gmiget - gets Gemini pages"
	usage = `  gmiget [flags...] <url>`
)

var (
	help       bool
	statusOnly bool
)

func main() {
	flag.BoolVar(&help, "help", false, "Show help for gmiget")
	flag.BoolVar(&help, "h", false, "Show help for gmiget")
	flag.BoolVar(&statusOnly, "I", false, "Output response header only")

	flag.Usage = cli.Usage(cli.UsageOptions{
		Description: desc,
		Usage:       usage,
	}, os.Stdout)
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(1)
	}

	geminiURL, err := getURL(os.Stdin)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

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

// getURL gets a URL from either stdin (if being piped in) or from the
// command line via args. It returns an error if a URL is not supplied.
func getURL(in *os.File) (string, error) {
	var url string

	f, err := in.Stat()
	if err != nil {
		return "", nil
	}

	// URL must either be passed in via a pipe, or flag
	if f.Mode()&os.ModeNamedPipe != 0 {
		// Got something, read it
		// TODO - this needs a little refactoring
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		url = string(b)
		url = strings.TrimRight(url, "\n")
	} else {
		args := flag.Args()
		if len(args) != 1 {
			return "", errors.New("gmiget: missing Gemini URL")
		}
		url = args[0]
	}

	return url, nil
}
