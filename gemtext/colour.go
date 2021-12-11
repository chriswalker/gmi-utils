package gemtext

import (
	"fmt"
	"strconv"
	"strings"
)

/*
 ------------------------------------------------------------------------------
 The following code is lifted in its entirety from:

 https://github.com/jwalton/gchalk
 ------------------------------------------------------------------------------
*/

// Close is the reset code for 16m ansi color codes.
const Close = "\u001B[39m"

func isHexDigit(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func parseHexColor(str string) string {
	index := 0

	// Find the "#"
	if index < len(str) && str[index] == '#' {
		index++
	}

	hexStart := index
	for index < len(str) && isHexDigit(str[index]) {
		index++
	}

	colorStr := str[hexStart:index]
	if len(colorStr) != 3 && len(colorStr) != 6 {
		return ""
	}

	return colorStr
}

// HexToRGB converts from the RGB HEX color space to the RGB color space.
//
// "hex" should be a hexadecimal string containing RGB data (e.g. "#2340ff" or "00f").
// Note that the leading "#" is optional.  If the hex code passed in is invalid,
// this will return 0, 0, 0 - it's up to you to validate the input if you want to
// detect invalid values.
func hexToRGB(hex string) (red uint8, green uint8, blue uint8) {
	s := parseHexColor(hex)
	if s == "" {
		return 0, 0, 0
	}

	// Adapted from https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		default:
			panic(fmt.Sprintf("Unhandled char %v", b))
		}
	}

	switch len(s) {
	case 6:
		red = hexToByte(s[0])<<4 + hexToByte(s[1])
		green = hexToByte(s[2])<<4 + hexToByte(s[3])
		blue = hexToByte(s[4])<<4 + hexToByte(s[5])
	case 3:
		red = hexToByte(s[0]) * 17
		green = hexToByte(s[1]) * 17
		blue = hexToByte(s[2]) * 17
	default:
		return 0, 0, 0
	}
	return
}

type Colour struct {
	red, green, blue uint8
}

func NewColour(hex string) *Colour {
	r, g, b := hexToRGB(hex)
	return &Colour{red: r, green: g, blue: b}
}

// WriteStringAnsi16m writes the string used to set a 24bit foreground color.
// WrapInAnsi16mColour wraps the suppled str in Ansi 16m escape codes for the
// supplied colour.
func WriteAnsi16mColour(out *strings.Builder, colour Colour) {
	out.WriteString("\u001B[38;2;")
	out.WriteString(strconv.Itoa(int(colour.red)))
	out.WriteString(";")
	out.WriteString(strconv.Itoa(int(colour.green)))
	out.WriteString(";")
	out.WriteString(strconv.Itoa(int(colour.blue)))
	out.WriteString("m")
}
