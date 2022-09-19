package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chriswalker/gmi-utils/cli"
	"github.com/chriswalker/gmi-utils/gemtext"
)

const (
	name  = "gmilinks"
	desc  = "gmilinks - extracts the links from the supplied raw gemtext"
	usage = `  gmilinks -f <file>,

  # Pipe in gemtext via stdin
  gmiget gemini://some-url/ | gmilinks | fzf`
)

var (
	help      bool
	inputFile string
)

func main() {
	flag.BoolVar(&help, "help", false, "Show help for gmilinks")
	flag.BoolVar(&help, "h", false, "Show help for gmilinks")
	flag.StringVar(&inputFile, "file", "", "Gemtext file to extract links from")
	flag.StringVar(&inputFile, "f", "", "Gemtext file to extract links from")

	flag.Usage = cli.Usage(cli.UsageOptions{
		Description: desc,
		Usage:       usage,
	}, os.Stdout)
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(1)
	}

	var input io.Reader

	// Require either a file, or something piped in on stdin
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
			os.Exit(1)
		}
		input = f
	} else {
		f, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
			os.Exit(1)
		}

		if f.Mode()&os.ModeNamedPipe == 0 {
			fmt.Println("nothing passed into stdin - exiting.")
			os.Exit(1)
		}

		input = os.Stdin
	}

	links := gemtext.ExtractLinks(input)
	for link, text := range links {
		fmt.Printf("%s|%s\n", text, link)
	}
}
