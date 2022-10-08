package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/chriswalker/gmi-utils/config"
)

func setup(xdgConfigHome, homeDir string) (home, xdg string) {
	// Grab original values
	home, _ = os.UserHomeDir()
	xdg, _ = os.LookupEnv("XDG_CONFIG_HOME")
	// If XDG_CONFIG_HOME was set, unset it; we might not be
	// using it in the following tests
	if xdg != "" {
		os.Unsetenv("XDG_CONFIG_HOME")
	}
	if xdgConfigHome != "" {
		os.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	}
	if homeDir != "" {
		os.Setenv("HOME", homeDir)
	}

	return
}

func reset(homeDir, xdgDir string) {
	if xdgDir != "" {
		os.Setenv("XDG_CONFIG_HOME", xdgDir)
	} else {
		os.Unsetenv("XDG_CONFIG_HOME")
	}
	os.Setenv("HOME", homeDir)
}

func TestLoad(t *testing.T) {
	testCases := map[string]struct {
		fileName          string
		xdgConfigHome     string
		homeDir           string
		expectedConfigLen int
		errMsg            string
	}{
		"not found": {
			fileName: "./testdata/does_not_exist",
			errMsg:   "no such file or directory",
		},
		"invalid config": {
			fileName: "./testdata/invalid",
			errMsg:   "invalid configuration item at line",
		},
		"load from XDG_CONFIG_HOME": {
			xdgConfigHome:     "./testdata/XDG",
			expectedConfigLen: 1,
		},
		"load from HOME/.config": {
			homeDir:           "./testdata/home",
			expectedConfigLen: 2,
		},
		"load from HOME": {
			homeDir:           "./testdata/home2",
			expectedConfigLen: 3,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			home, xdg := setup(tc.xdgConfigHome, tc.homeDir)
			defer reset(home, xdg)

			config, err := config.Load(tc.fileName)

			if tc.errMsg != "" {
				if err == nil {
					t.Error("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("got error '%s', want '%s", err, tc.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %q:", err)
				return
			}
			if tc.expectedConfigLen > 0 && len(*config) != tc.expectedConfigLen {
				t.Errorf("parsed %d config items, want %d", len(*config), tc.expectedConfigLen)
			}
		})
	}
}
