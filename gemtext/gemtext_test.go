package gemtext

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chriswalker/gmi-utils/config"
)

var (
	availableWidth = 25
	margin         = 5
	update         = flag.Bool("update", false, "whether to regenerate .golden files")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestGetLineType(t *testing.T) {
	testCases := map[string]struct {
		line           string
		isPreformatted bool
		expected       lineType
	}{
		"text line":           {line: "Just a regular text line", expected: text},
		"preformatted toggle": {line: "```", expected: preformattedToggle},
		"preformatted":        {line: "preformatted text", isPreformatted: true, expected: preformatted},
		"quoted":              {line: "> A quote", expected: quoted},
		// "link":                {line: "=> gemini://some.url/ Some URL", expected: link},
		"list item": {line: "* Bullet item", expected: listItem},
		"header 1":  {line: "#", expected: header},
		"header 2":  {line: "##", expected: header2},
		"header 3":  {line: "###", expected: header3},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := getLineType(tc.isPreformatted, tc.line)
			if got != tc.expected{
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestWrapLine(t *testing.T) {
	testCases := map[string]struct {
		lineType     lineType
		line         string
		wrappedLines []string
	}{
		"basic": {
			lineType:     text,
			line:         "one two three four five six",
			wrappedLines: []string{"one two three four five", "six"},
		},
		"newline only": {
			lineType:     text,
			line:         "",
			wrappedLines: []string{""},
		},
		"spaces only": {
			lineType:     text,
			line:         "     ",
			wrappedLines: []string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := wrap(availableWidth, tc.line)
			if len(got) != len(tc.wrappedLines) {
				t.Errorf("wrong number of wrapped lines returned; got %d, want %d", len(got), len(tc.wrappedLines))
			}
			for i, line := range got {
				if line != tc.wrappedLines[i]{
					t.Errorf("line %d does not match; got '%s', want '%s'", i, line, tc.wrappedLines[i])
				}
			}
		})
	}
}

func TestNewBlock(t *testing.T) {
	testCases := map[string]struct {
		line           string
		isPreformatted bool
		expected       block
	}{
		"text block": {
			line: "basic text block here",
			expected: block{
				lineType: text,
				lines:    []string{"basic text", "block here"},
			},
		},
		"link block with link text": {
			line: "=> gemini://some.url/ This is a link",
			expected: block{
				lineType: link,
				lines:    []string{"This is a link [gemini://some.url/]"},
			},
		},
		"link block with no link text": {
			line: "=> gemini://some.url/",
			expected: block{
				lineType: link,
				lines:    []string{"[gemini://some.url/]"},
			},
		},
		"preformatted toggle": {
			line: "```",
			expected: block{
				lineType: preformattedToggle,
				lines:    []string{},
			},
		},
		"preformatted": {
			line:           "preformatted text",
			isPreformatted: true,
			expected: block{
				lineType: preformatted,
				lines:    []string{"preformatted text"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When calling NewBlock, need to account for additional impacts to overall width
			block := NewBlock(availableWidth, margin, tc.isPreformatted, tc.line)
			if block.lineType != tc.expected.lineType {
				t.Errorf("got line type of '%v', want '%v'", block.lineType, tc.expected.lineType)
			}
			if len(block.lines) != len(tc.expected.lines) {
				t.Errorf("got %d parsed lines, want %d", len(block.lines), len(tc.expected.lines))
			}
			for i, line := range block.lines {
				if line != tc.expected.lines[i] {
					t.Errorf("line %d does not match; got '%s', want '%s'", i, line, tc.expected.lines[i])
				}
			}
		})
	}
}

// Tests output of complete line - gutter, prefixes and text
func TestBlockString(t *testing.T) {
	marginStr := strings.Repeat(" ", margin)
	testCases := map[string]struct {
		block    block
		expected string
	}{
		// TODO rework expected a bit
		"text block": {
			block: block{
				lineType: text,
				lines:    []string{"line one of the text block", "line two of the text block"},
			},
			expected: fmt.Sprintf("%sline one of the text block\n%sline two of the text block\n", marginStr, marginStr),
		},
		"list item block": {
			block: block{
				lineType: listItem,
				lines:    []string{"line one of the text block", "line two of the text block"},
			},
			expected: fmt.Sprintf("%s%s line one of the text block\n%s  line two of the text block\n", marginStr, listItem.prefix, marginStr),
		},
		"quote block": {
			block: block{
				lineType: quoted,
				lines:    []string{"line one of the text block", "line two of the text block"},
			},
			expected: fmt.Sprintf("%s%s line one of the text block\n%s%s line two of the text block\n",
				marginStr,
				quoted.prefix,
				marginStr,
				quoted.prefix),
		},
		"header block": {
			block: block{
				lineType: header,
				lines:    []string{"line one of the text block", "line two of the text block"},
			},
			expected: fmt.Sprintf("%s%s line one of the text block\n%s%s line two of the text block\n",
				marginStr,
				header.prefix,
				marginStr,
				header.prefix),
		},
		"header 2 block": {
			block: block{
				lineType: header2,
				lines:    []string{"line one of the text block", "line two of the text block"},
			},
			expected: fmt.Sprintf("%s%s line one of the text block\n%s%s line two of the text block\n",
				marginStr,
				header2.prefix,
				marginStr,
				header2.prefix),
		},
		"header 3 block": {
			block: block{
				lineType: header3,
				lines:    []string{"line one of the text block", "line two of the text block"},
			},
			expected: fmt.Sprintf("%s%s line one of the text block\n%s%s line two of the text block\n",
				marginStr,
				header3.prefix,
				marginStr,
				header3.prefix),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := tc.block.String(margin)
			if s != tc.expected {
				t.Errorf("got formatted string of '%s', want '%s'", s, tc.expected)
			}
		})
	}
}

func TestOutput(t *testing.T) {
	conf := make(config.Config)
	conf["preformatted"] = "#ff0000"
	conf["header"] = "#00ffff"
	conf["header2"] = "#00ffff"
	conf["header3"] = "#ffff00"
	conf["link"] = "#00ff00"
	conf["quoted"] = "#ff00ff"
	Configure(conf)

	paths, err := filepath.Glob(filepath.Join("testdata", "*.input"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		_, fileName := filepath.Split(path)
		testName := fileName[:len(fileName)-len(filepath.Ext(path))]

		t.Run(testName, func(t *testing.T) {
			input, err := os.Open(path)
			if err != nil {
				t.Fatalf("error reading test input file: %s", err)
			}
			defer input.Close()

			var b []byte
			got := bytes.NewBuffer(b)
			Output(30, 2, input, got)

			golden := filepath.Join("testdata", testName+".golden")
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

func TestExtractLinks(t *testing.T) {
	expected := map[string]string{
		"gemini://some.url/fragment": "textual name",
		"gemini://some.other/url":    "closer text",
		// TODO handle non-text URLs
		"gemini://url/with/no/text/": "",
		"/relative/url":              "a relative URL",
	}

	f, err := os.Open("./testdata/extract")
	if err != nil {
		t.Fatal(err)
	}

	links := ExtractLinks(f)

	for url, want := range expected {
		var got string
		got, ok := links[url]
		if !ok {
			t.Errorf("expected URL of '%s' not found", url)
		}
		if got != want {
			t.Errorf("got value of '%s' for URL '%s', want '%s'", got , url, want)
		}
	}
}
