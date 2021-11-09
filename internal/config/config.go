package config

import (
	"errors"
	"io/ioutil"
	"os"

	rt "gitlab.com/clickport/clickport-agent/internal/clickport"
	"gopkg.in/yaml.v2"
)

const defaultConfigFilepath = ".clickport-config.yml"

type Config struct {
	ClickportScripts *rt.ClickportScripts
	SigningSecret    string
	APISecret        string
}

func loadConfig(configFilepath *string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(*configFilepath)
	if err != nil {
		return nil, err
	}

	clickportScripts := make(rt.ClickportScripts)

	if err = yaml.Unmarshal(yamlFile, &clickportScripts); err != nil {
		return nil, err
	}

	signingSecret, hasSigningSecret := os.LookupEnv("SIGNING_SECRET")
	if !hasSigningSecret {
		return nil, errors.New("must specify `SIGNING_SECRET` environment variable")
	}

	config := &Config{
		ClickportScripts: &clickportScripts,
		SigningSecret:    signingSecret,
	}

	return config, nil
}

func LoadConfig() (*Config, error) {
	configFilepath := os.Getenv("CONFIG_FILEPATH")
	if configFilepath == "" {
		configFilepath = defaultConfigFilepath
	}

	return loadConfig(&configFilepath)
}
