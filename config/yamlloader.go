package utils

import (
	strut "github.com/pintobikez/correios-service/config/structures"
	errors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

var (
	ErrInvalidFile       = "Invalid file absolute path"
	ErrUnableToReadFile  = "Unable to read the file storage"
	ErrUnableToParseFile = "Unable to parse the file storage"
)

// Loads a Yaml file and returns it
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

// Loads the given Yaml file into a YmlConfig structure
func LoadDBConfigFile(filename string, c *strut.DbConfig) error {

	file, err := LoadYamlFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(file, c); err != nil {
		return errors.Wrap(err, ErrUnableToParseFile)
	}

	return nil
}

// Loads the given Yaml file into a YmlConfig structure
func LoadCorreiosConfigFile(filename string, c *strut.CorreiosConfig) error {

	file, err := LoadYamlFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(file, c); err != nil {
		return errors.Wrap(err, ErrUnableToParseFile)
	}

	return nil
}
