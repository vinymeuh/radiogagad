// Copyright 2020 VinyMeuh. All rights reserved.
// Use of the source code is governed by a MIT-style license that can be found in the LICENSE file.

package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config is the format of the application's configuration file
type Config struct {
	MPD struct {
		Server string `yaml:"host"`
	} `yaml:"mpd"`
	PowerButton struct {
		Enabled     bool   `yaml:"enabled"`
		ShutdownCmd string `yaml:"shutdown_cmd"`
	} `yaml:"powerbutton"`
}

// NewConfig creates a new Config with default values
func NewConfig() Config {
	var c Config
	c.MPD.Server = "localhost:6600"
	c.PowerButton.Enabled = true
	return c
}

// LoadFromFile fills the Config from a YAML file
func (c *Config) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(c)
	return err
}
