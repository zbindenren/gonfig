package gonfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
)

// parseMapOpts parses options from a map[string]interface{}.  This is used
// for configuration file encodings that can decode to such a map.
func parseMapOpts(j map[string]interface{}, opts []*option) error {
	for _, opt := range opts {
		val, set := j[opt.id]
		if !set {
			continue
		}

		if opt.isParent {
			if casted, ok := val.(map[string]interface{}); ok {
				if err := parseMapOpts(casted, opt.subOpts); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("error parsing config file: "+
					"value of type %s given for composite config var %s",
					reflect.TypeOf(val), opt.fullID())
			}
		} else {
			if err := opt.setValue(reflect.ValueOf(val)); err != nil {
				return err
			}
		}
	}

	return nil
}

// openConfigFile trues to open the config file at path.  If it fails
// it returns a nice error.
func openConfigFile(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file at %s does not exist", path)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading config file at %s: %s", path, err)
	}

	return content, nil
}

// parseFile parses the config file for all config options by delegating
// the call to the method specific to the config file encoding specified.
func parseFile(s *setup) error {
	content, err := openConfigFile(s.configFilePath)
	if err != nil {
		return err
	}

	decoder := s.conf.FileDecoder
	if decoder == nil {
		// Look for the config file extension to determine the encoding.
		switch path.Ext(s.configFilePath) {
		case "json":
			decoder = DecoderJSON
		case "toml":
			decoder = DecoderTOML
		case "yaml", "yml":
			decoder = DecoderYAML
		default:
			decoder = DecoderTryAll
		}
	}

	m, err := decoder(content)
	if err != nil {
		return fmt.Errorf("failed to parse file at %s: %s",
			s.configFilePath, err)
	}

	// Parse the map for the options.
	if err := parseMapOpts(m, s.opts); err != nil {
		return fmt.Errorf("error loading config vars from config file: %s", err)
	}

	return nil
}
