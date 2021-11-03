package config

import (
	"errors"
	"io/ioutil"
	"os"

	rt "gitlab.com/runtime-hq/runtime-agent/internal/runtime"
	"gopkg.in/yaml.v2"
)

const defaultConfigFilepath = ".runtime-config.yml"

type Config struct {
	RuntimeScripts *rt.RuntimeScripts
	SigningSecret  string
	APISecret      string
}

func loadConfig(configFilepath *string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(*configFilepath)
	if err != nil {
		return nil, err
	}

	runtimeScripts := make(rt.RuntimeScripts)

	if err = yaml.Unmarshal(yamlFile, &runtimeScripts); err != nil {
		return nil, err
	}

	signingSecret, hasSigningSecret := os.LookupEnv("SIGNING_SECRET")
	if !hasSigningSecret {
		return nil, errors.New("must specify `SIGNING_SECRET` environment variable")
	}

	config := &Config{
		RuntimeScripts: &runtimeScripts,
		SigningSecret:  signingSecret,
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
