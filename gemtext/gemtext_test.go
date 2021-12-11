package gemtext

import (
	"fmt"
	"strings"
	"testing"
)

var (
	availableWidth = 25
	margin         = 5
)

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
			if tc.expected != got {
				t.Errorf("expected %v, got %v", tc.expected, got)
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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := wrap(availableWidth, tc.line)
			if len(tc.wrappedLines) != len(got) {
				t.Errorf("wrong number of wrapped lines returned; expected %d, got %d", len(tc.wrappedLines), len(got))
			}
			for i, line := range got {
				if tc.wrappedLines[i] != line {
					t.Errorf("line %d does not match; expected '%s', got '%s'", i, tc.wrappedLines[i], line)
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
				t.Errorf("expected line type of '%v', got '%v'", block.lineType, tc.expected.lineType)
			}
			if len(block.lines) != len(tc.expected.lines) {
				t.Errorf("expected %d parsed lines, got %d", len(tc.expected.lines), len(block.lines))
			}
			for i, line := range block.lines {
				if line != tc.expected.lines[i] {
					t.Errorf("line %d does not match; expected '%s', got '%s'", i, tc.expected.lines[i], line)
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
				t.Errorf("expected formatted string of '%s', got '%s'", tc.expected, s)
			}
		})
	}
}
