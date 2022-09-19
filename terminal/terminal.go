package terminal

import (
	"os"
	"syscall"

	"golang.org/x/term"
)

const consoleDev = "/dev/tty"

func GetWidth() int {
	ttyFile := getTTY()
	width, _, _ := term.GetSize(int(ttyFile.Fd())) // TODO err
	return width
}

// Below functions taken from fzf, which has a great way of determining
// the terminal device, and falls back to stdin - see
//
// https://github.com/junegunn/fzf/blob/ef67a45702c01ff93e0ea99a51594c8160f66cc1/src/tui/ttyname_unix.go

// getTTY returns terminal device to be used as stdin, falls back to os.Stdin
func getTTY() *os.File {
	in, err := os.OpenFile(consoleDev, syscall.O_RDONLY, 0)
	if err != nil {
		tty := ttyname()
		if len(tty) > 0 {
			if in, err := os.OpenFile(tty, syscall.O_RDONLY, 0); err == nil {
				return in
			}
		}
		return os.Stdin
	}
	return in
}

func ttyname() string {
	var stderr syscall.Stat_t
	if syscall.Fstat(2, &stderr) != nil {
		return ""
	}

	devDirs := [2]string{"/dev/pts/", "/dev/"}

	for _, prefix := range devDirs {
		files, err := os.ReadDir(prefix)
		if err != nil {
			continue
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				// Caller checks length of return; we'll end up
				// trying to use stdin if this happens
				return ""
			}
			if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Rdev == stderr.Rdev {
				return prefix + file.Name()
			}
		}
	}
	return ""
}
