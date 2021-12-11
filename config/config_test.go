package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/chriswalker/gmi-utils/config"
)

func setup(xdgConfigHome, homeDir string) {
	if xdgConfigHome != "" {
		os.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	}
	if homeDir != "" {
		os.Setenv("HOME", homeDir)
	}
}

func teardown(homeDir string) {
	os.Unsetenv("XDG_CONFIG_HOME")
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

	home, _ := os.UserHomeDir()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup(tc.xdgConfigHome, tc.homeDir)
			defer teardown(home)

			config, err := config.Load(tc.fileName)

			if tc.errMsg != "" {
				if err == nil {
					t.Error("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error '%s', got '%s", tc.errMsg, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %q:", err)
				return
			}
			if tc.expectedConfigLen > 0 && tc.expectedConfigLen != len(*config) {
				t.Errorf("expected %d config items to have been parsed, got %d", tc.expectedConfigLen, len(*config))
			}
		})
	}
}
