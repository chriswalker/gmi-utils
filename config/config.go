package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	// Default directory name for holding Gemini utility configuration
	configPath = "gemini"
	// Default name for gmifmt configuration files
	defaultConfigFilename = ".gmifmtconf"
)

// Configuration stretches to just colour values for
// Gemtext elements currently
type Config map[string]string

// Load attempts to load configuration files from a file. A file
// spcified on the the command-line with '-c' or '--config' takes
// precedence, and after that we check the following:
//
//   $XDG_CONFIG_HOME/gemini/.gmifmtconf
//   $HOME/.config/gemini/.gmifmtconf
//   $HOME/.gmifmtconf
func Load(file string) (*Config, error) {
	filepaths := []string{file}

	if val, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		filepaths = append(filepaths, fmt.Sprintf("%s/%s/%s", val, configPath, defaultConfigFilename))
	}
	val, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filepaths = append(filepaths, fmt.Sprintf("%s/.config/%s/%s", val, configPath, defaultConfigFilename))
	filepaths = append(filepaths, fmt.Sprintf("%s/%s", val, defaultConfigFilename))

	var configFile *os.File
	for i, path := range filepaths {
		configFile, err = os.Open(path)
		if err != nil {
			if i == 0 && file != "" {
				// Config file specified on command line
				if os.IsNotExist(err) {
					return nil, err
				}
			}
			// Otherwise, only return out if we have an error that's not related
			// to file existance on one of the other paths
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		break
	}
	defer configFile.Close()

	if configFile != nil {
		return load(configFile)
	}

	return nil, nil
}

// load loads the given file and parses it into a config
// map.
func load(file *os.File) (*Config, error) {
	conf := make(Config)

	s := bufio.NewScanner(file)
	i := 1
	for s.Scan() {
		line := s.Text()
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid configuration item at line %d ('%s')", i, line)
		}

		conf[parts[0]] = parts[1]
		i++
	}

	return &conf, nil
}
