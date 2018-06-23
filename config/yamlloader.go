package utils

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// Loads a Yaml file and returns it
const (
	//ErrInvalidFile message
	ErrInvalidFile = "Invalid file absolute path"
	//ErrUnableToReadFile message
	ErrUnableToReadFile = "Unable to read the file storage"
	//ErrUnableToParseFile message
	ErrUnableToParseFile = "Unable to parse the file storage"
)

//LoadYamlFile Loads a Yaml file and returns it
func LoadYamlFile(filename string) ([]byte, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidFile)
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, ErrUnableToReadFile)
	}

	return file, nil
}

//LoadConfigFile Loads the given Yaml file into the Structure
func LoadConfigFile(filename string, c interface{}) error {

	file, err := LoadYamlFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(file, c); err != nil {
		return errors.Wrap(err, ErrUnableToParseFile)
	}

	return nil
}
