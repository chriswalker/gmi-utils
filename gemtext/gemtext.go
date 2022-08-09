package gemtext

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/chriswalker/gmi-utils/config"
)

// lineType represents the various lines types in Gemtext.
type lineType struct {
	// The line type's prefix
	prefix string
	// Output colour; if nil no colour assigned
	Colour *Colour
}

var (
	text               = lineType{prefix: ""}
	link               = lineType{prefix: "=>"}
	preformattedToggle = lineType{prefix: "```"}
	preformatted       = lineType{prefix: ""}
	header             = lineType{prefix: "#"}
	header2            = lineType{prefix: "##"}
	header3            = lineType{prefix: "###"}
	listItem           = lineType{prefix: "*"}
	quoted             = lineType{prefix: ">"}
)

// Configure takes any colour values set in the given
// Config struct and applies them to the defined line types.
func Configure(conf config.Config) {
	if val, ok := conf["preformatted"]; ok {
		preformatted.Colour = NewColour(val)
	}
	if val, ok := conf["header"]; ok {
		header.Colour = NewColour(val)
	}
	if val, ok := conf["header2"]; ok {
		header2.Colour = NewColour(val)
	}
	if val, ok := conf["header3"]; ok {
		header3.Colour = NewColour(val)
	}
	if val, ok := conf["quoted"]; ok {
		quoted.Colour = NewColour(val)
	}
	if val, ok := conf["link"]; ok {
		link.Colour = NewColour(val)
	}
}

// Output formats the supplied byte slice and emits to
// the supplied writer (usually os.Stdout). The caller
// provides the current width of the terminal, which is
// used to determine text wrapping and margins.
func Output(width, margin int, r io.Reader, w io.Writer) {
	preformatted := false
	var block block

	s := bufio.NewScanner(r)
	for s.Scan() {
		block = NewBlock(width, margin, preformatted, s.Text())
		if block.lineType == preformattedToggle {
			preformatted = !preformatted
			continue
		}
		fmt.Fprint(w, block.String(margin))
	}
}

// getLineType returns the lineType for the supplied based on its
// prefix value.
func getLineType(isPreformatted bool, line string) lineType {
	switch {
	case strings.HasPrefix(line, preformattedToggle.prefix):
		return preformattedToggle
	case isPreformatted:
		return preformatted
	case strings.HasPrefix(line, quoted.prefix):
		return quoted
	case strings.HasPrefix(line, header3.prefix):
		return header3
	case strings.HasPrefix(line, header2.prefix):
		return header2
	case strings.HasPrefix(line, header.prefix):
		return header
	case strings.HasPrefix(line, listItem.prefix):
		return listItem
	case strings.HasPrefix(line, link.prefix):
		return link
	default:
		return text
	}
}

// block represents a parsed, formatted single line of gemtext
// to output. Parsed gemtext may generate multiple physical lines
// in the event of word-wrapping.
//
// Each line is comprised of:
//
// - A margin,
// - Optional prefix columns (for headers, quotes, list items)
// - One or more lines of formatted text, wrapped to adjust for
//   terminal width
type block struct {
	// The type of line this block represents
	lineType lineType
	// The text for this block; if the line's length is
	// greater than the terminal width, it is word-wrapped
	// into a number of lines
	lines []string
}

// NewBlock takes the supplied line of bytes and converts it into
// a 'block' - that is, an internal structure that contains the
// type of the line (URL, bullet list item, header, etc) and a
// slice of lines representing a word-wrapped sequence of text.
func NewBlock(width, margin int, preformatted bool, line string) block {
	lineType := getLineType(preformatted, line)
	s := line

	// strip any prefixes
	s = s[len(lineType.prefix):]

	b := block{
		lineType: lineType,
	}

	// TODO revisit this
	if preformatted {
		// Preformatted text output as-is
		b.lines = append(b.lines, s)
	} else if lineType == link {
		b.lines = []string{parseLink(s)}
	} else if lineType != preformattedToggle {
		// Available width is calculated as:
		// the current terminal width - (L + R margin) - prefix width
		availableWidth := width - margin*2 - (len(lineType.prefix) + 1)
		b.lines = wrap(availableWidth, s)
	}

	return b
}

// String outputs all lines in this block as a single string.
func (b *block) String(margin int) string {
	builder := strings.Builder{}

	for i, line := range b.lines {
		// Generate margin
		builder.WriteString(strings.Repeat(" ", margin))

		if b.lineType.Colour != nil {
			WriteAnsi16mColour(&builder, *b.lineType.Colour)
		}

		// Output prefix
		if b.lineType != link && b.lineType.prefix != "" {
			// TODO - not happy with this, revisit
			if i == 0 || b.lineType != listItem && i > 0 {
				builder.WriteString(b.lineType.prefix)
				builder.WriteString(" ")
			} else if b.lineType == listItem {
				builder.WriteString("  ")
			}
		}

		// add line
		builder.WriteString(line)

		if b.lineType.Colour != nil {
			builder.WriteString(Close)
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// wrap takes the supplied line and wraps it on word boundaries
// based on the supplied width. It returns a slice of wrapped lines.
func wrap(width int, line string) []string {
	var wrapped []string

	if len(line) == 0 {
		wrapped = append(wrapped, line)
		return wrapped
	}

	words := strings.Fields(strings.TrimSpace(line))

	// May just have spaces on a line
	if len(words) == 0 {
		return wrapped
	}

	wrappedLine := words[0]
	spaceLeft := width - len(wrappedLine)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped = append(wrapped, wrappedLine)
			wrappedLine = word
			spaceLeft = width - len(wrappedLine)
		} else {
			wrappedLine += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return append(wrapped, wrappedLine)
}

// parseLink parses the supplied URL line type into a formatted string.
// A URL line type is defined in the Gemini spec as:
//
//   =>[<whitespace>]<URL>[<whitespace><USER-FRIENDLY LINK NAME>]
//
// where:
//
// * <whitespace> is any non-zero number of consecutive spaces or tabs
// * Square brackets indicate that the enclosed content is optional.
// * <URL> is a URL, which may be absolute or relative.
//
// (lifted directly from gemini://gemini.circumlunar.space/docs/spcification.gmi)
//
// Note that URL lines have already had their prefix stripped when
// parsing.
func parseLink(line string) string {
	cutSet := " \t"
	line = strings.TrimLeft(line, cutSet)

	var link string

	// Get index between URL and link name
	idx := strings.IndexAny(line, cutSet)
	if idx > 0 {
		link = fmt.Sprintf("%s [%s]", strings.Trim(line[idx:], cutSet),
			strings.Trim(line[:idx], cutSet))
	} else {
		// No link name specified, just a URL
		link = fmt.Sprintf("[%s]", strings.Trim(line, cutSet))
	}

	return link
}

// ExtractLinks constructs a map of links and their link text from
// the provided io.Reader.
func ExtractLinks(r io.Reader) map[string]string {
	links := make(map[string]string)

	cutSet := " \t"
	preformatted := false

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, preformattedToggle.prefix) {
			preformatted = !preformatted
			continue
		}
		if preformatted {
			continue
		}
		if strings.HasPrefix(line, link.prefix) {
			// strip any prefixes
			line = line[len(link.prefix):]
			// Strip any trailing stuff left of the cutset
			line = strings.TrimLeft(line, cutSet)

			// Get index between URL and link name
			idx := strings.IndexAny(line, cutSet)
			if idx > 0 {
				links[strings.Trim(line[:idx], cutSet)] = strings.Trim(line[idx:], cutSet)
			} else {
				// No link name specified, just a URL
				links[strings.Trim(line, cutSet)] = ""
			}
		}
	}

	return links
}
