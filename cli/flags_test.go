package cli_test

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/chriswalker/gmi-utils/cli"
)

var update = flag.Bool("update", false, "whether to regenerate .golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestUsage(t *testing.T) {
	testCases := map[string]struct {
		description string
		usage       string
		setup       func()
		goldenFile  string
	}{
		"basic": {
			description: "Test description of tool",
			usage:       `    a <flags>`,
			setup: func() {
				flag.String("c", "", "path to config file to load")
				flag.String("config", "", "path to config file to load")
			},
			goldenFile: "basic",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			// Setup
			var b []byte
			got := bytes.NewBuffer(b)
			usage := cli.UsageOptions{
				Description: tc.description,
				Usage:       tc.usage,
			}
			flag.Usage = cli.Usage(usage, got)

			tc.setup()
			flag.Parse()

			// Run Usage, catch output in 'got'
			flag.Usage()

			// Load golden file, and compare with usage output
			golden := filepath.Join("testdata", tc.goldenFile+".golden")
			f, err := os.OpenFile(golden, os.O_RDWR, 0644)
			if err != nil {
				t.Fatalf("error opening test golden file: %s", err)
			}
			defer f.Close()

			want, err := io.ReadAll(f)
			if err != nil {
				t.Fatalf("error reading test golden file: %s", err)
			}

			if *update {
				if err := os.WriteFile(golden, got.Bytes(), 0644); err != nil {
					t.Fatalf("error updating golden file '%s': %s", golden, err)
				}
				want = got.Bytes()
			}

			if !bytes.Equal(got.Bytes(), want) {
				t.Errorf("got '%s', want '%s'", got, want)
			}
		})
	}
}
