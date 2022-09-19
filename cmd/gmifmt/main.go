package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chriswalker/gmi-utils/cli"
	"github.com/chriswalker/gmi-utils/config"
	"github.com/chriswalker/gmi-utils/gemtext"
	"github.com/chriswalker/gmi-utils/terminal"
)

const (
	desc  = "gmifmt - formats and colours Gemtext"
	usage = `  gmifmt -f <file>,

  # Pipe in gemtext via stdin
  gmiget gemini://some-url/ | gmifmt [flags...]`
)

var (
	help       bool
	margin     int
	inputFile  string
	configFile string
)

func main() {
	flag.BoolVar(&help, "help", false, "Show help for gmifmt")
	flag.BoolVar(&help, "h", false, "Show help for gmifmt")
	flag.IntVar(&margin, "margin", 0, "Width of margin to apply to formatted gemtext")
	flag.IntVar(&margin, "m", 0, "Width of margin to apply to formatted gemtext")
	flag.StringVar(&inputFile, "file", "", "Gemtext file to format")
	flag.StringVar(&inputFile, "f", "", "Gemtext file to format")
	flag.StringVar(&configFile, "config", "", "Path to gmifmt configuration file")
	flag.StringVar(&configFile, "c", "", "Path to gmifmt configuration file")

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
