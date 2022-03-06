package config

import (
	"errors"
	"io/ioutil"
	"net/url"

	"gopkg.in/yaml.v2"
)

type ClientType string

const (
	MatrixType ClientType = "matrix"
	SlackType  ClientType = "slack"
)

type ClientConfig struct {
	// The chat platform's username to connect with
	Username string `yaml:"username"`
	// The password to authenticate the requests with.
	Password string `yaml:"password"`
	// The access token to authenticate the requests with.
	AccessToken string `yaml:"accesstoken"`
	// A URL with the host and port of the matrix server. E.g. https://matrix.org:8448
	HomeserverURL string `yaml:"homeserverurl"`
	// The desired display name for this client.
	DisplayName string `yaml:"displayname"`
	// The type of this client.
	ClientType ClientType `yaml:"clienttype"`
}

type ConfigFile struct {
	Clients []ClientConfig `yaml:"clients"`
}

// Check that the client has supplied the correct fields.
func (c *ClientConfig) Check() error {
	if c.Username == "" || c.HomeserverURL == "" || (c.AccessToken == "" && c.Password == "") {
		return errors.New(`must supply a "Username", a "HomeserverURL", and an "AccessToken/Password"`)
	}
	if _, err := url.Parse(c.HomeserverURL); err != nil {
		return err
	}
	return nil
}

// Load the ConfigFile
func Load(configPath string) (*ConfigFile, error) {
	var configFile ConfigFile

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	// Pass the current working directory and ioutil.ReadFile so that they can
	// be mocked in the tests
	if err = yaml.Unmarshal(configData, &configFile); err != nil {
		return nil, err
	}
	return &configFile, nil
}
