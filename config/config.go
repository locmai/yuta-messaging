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

type NluClientType string

const (
	DiaglogflowClientType NluClientType = "diaglog"
	LuisClientType        NluClientType = "luis"
)

type ServerConfig struct {
	// The host which the server run on
	Host string `yaml:"host"`
	// The port which the server run on
	Port string `yaml:"port"`
	// Timeout for both read and write operations
	Timeout int `yaml:"timeout"`
}

type NluClientConfig struct {
	projectID string `yaml:"projectid"`
	sessionID string `yaml:"sessionid"`
}

type ConfigFile struct {
	Server     ServerConfig      `yaml:"server"`
	Clients    []ClientConfig    `yaml:"clients"`
	NluClients []NluClientConfig `yaml:"nluclients"`
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
