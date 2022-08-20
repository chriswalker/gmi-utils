package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chriswalker/gmi-utils/config"
	"github.com/chriswalker/gmi-utils/gemtext"
	"github.com/chriswalker/gmi-utils/terminal"
)

const usage = `
gmifmt - formats and colours Gemtext

Usage:
    gmifmt -f <file>, or pipe in via stdin:
    gmiget gemini://some-url/ | gmifmt [options]

Options:
    -h, --help     Show help for gmifmt
    -f, --file     Format gemtext from file
    -c, --config   Read configuration from file
    -m, --margin   Set margins for output

`

var (
	help       bool
	margin     int
	inputFile  string
	configFile string
)

func main() {
	flag.BoolVar(&help, "help", false, "Show help for gmifmt")
	flag.BoolVar(&help, "h", false, "Show help for gmifmt")
	flag.IntVar(&margin, "margin", 0, "width of margin to apply to formatted gemtext")
	flag.IntVar(&margin, "m", 0, "width of margin to apply to formatted gemtext")
	flag.StringVar(&inputFile, "file", "", "gemtext file to format")
	flag.StringVar(&inputFile, "f", "", "gemtext file to format")
	flag.StringVar(&configFile, "config", "", "path to gmifmt configuration file")
	flag.StringVar(&configFile, "c", "", "path to gmifmt configuration file")

	flag.Usage = func() {
		fmt.Print(usage)
	}
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
			fmt.Println(err)
			os.Exit(1)
		}
		input = f
	} else {
		f, err := os.Stdin.Stat()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if f.Mode()&os.ModeNamedPipe == 0 {
			fmt.Println("nothing passed into stdin - exiting.")
			os.Exit(1)
		}

		input = os.Stdin
	}

	config, err := config.Load(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if config != nil {
		gemtext.Configure(*config)
	}

	fmt.Println()

	width := terminal.GetWidth()
	gemtext.Output(width, margin, input, os.Stdout)
}
