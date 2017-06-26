package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Config struct {
	InputPath     string `json:"input_path"`
	OutputPath    string `json:"output_path"`
	ErrorPath     string `json:"error_path"`
	CompletedPath string `json:"completed_path"`
}

func GetConfigFromFile(fileName string) (*Config, error) {
	c := Config{}
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		msg := fmt.Sprintf("encountered an error [%s] reading config file: %s", err, fileName)
		return &c, errors.New(msg)

	}
	err = json.Unmarshal(bytes, &c)

	return &c, err
}

func (c *Config) IsValid() bool {
	if c.InputPath == "" || c.OutputPath == "" || c.ErrorPath == "" || c.CompletedPath == "" {
		return false
	}
	return true
}
