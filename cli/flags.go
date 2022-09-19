// Package cli provides some wrapper functionality around the
// standard library flags package; it currently provides code
// for emitting Usage functions to replace the defatult Go
// usage output.
//
// Callers should all Usage() to obtain a function assignable
// to flag.Usage. cli.Usage should be provided with a UsaggeOptions
// structure providing basic information about the program, and
// an io.Writer (usually os.Stdout, but can be substituted for
// testing).

package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// Padding for flag output
const namePadding = 15

// UsageOptions wraps some descriptive text for help output - currently
// just its textual description, and usage examples.
type UsageOptions struct {
	Description string
	Usage       string
}

// flagInfo is a private structure for recording details about
// registered flags; it's used to eventually output the usage
// text.
type flagInfo struct {
	// Each element in the flags slice is a formatted string
	// of short- and long-form flags, matched by usage strings.
	// This takes advantage of the fact VisitAll visits these
	// alphabetically and that short- and long-form flags
	// currentlt start with the same character.
	// If this changes, it'll need refactoring.
	flags      []string
	usage      string
	defaultVal string
}

// Usage generates the help text for the program. It returns
// a regular function assignable to flags.Usage in the standard
// library's flag package.
func Usage(opts UsageOptions, w io.Writer) func() {
	return func() {
		fmt.Fprintf(w, "%s\n\n", opts.Description)

		if opts.Usage != "" {
			fmt.Fprintln(w, "Usage:")
			fmt.Fprintf(w, "%s\n\n", opts.Usage)
		}

		fmt.Fprintln(w, "Flags:")
		var flags []*flagInfo
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			var added bool
			for _, fl := range flags {
				if fl.usage == f.Usage {
					fl.flags = append(fl.flags, fmt.Sprintf("--%s", f.Name))
					added = true
				}
			}
			if !added {
				f := &flagInfo{
					flags:      []string{fmt.Sprintf("-%s", f.Name)},
					usage:      f.Usage,
					defaultVal: f.DefValue,
				}
				flags = append(flags, f)
			}
		})

		for _, fl := range flags {
			flags := strings.Join(fl.flags, ", ")

			usage := fl.usage
			if fl.defaultVal != "" {
				usage = fmt.Sprintf("%s (default: %v)", usage, fl.defaultVal)
			}
			fmt.Fprintf(w, "  %s %s\n", rightPad(flags, namePadding), usage)
		}
	}
}

// rightPad adds padding to the right of a string - copied from:
//
// https://github.com/cli/cli/blob/537dcb27398bd193a094d5fcd3aa9bc67a546328/pkg/cmd/root/help.go#L229-L232
func rightPad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds ", padding)
	return fmt.Sprintf(template, s)
}
