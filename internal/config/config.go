package config

import (
	"fmt"
	"io/ioutil"
	"os"

	rt "gitlab.com/clickport/clickport-agent/internal/clickport"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilepath = ".clickport.yml"

	signingSecretEnvKey  = "CLICKPORT_SIGNING_SECRET"
	configFilepathEnvKey = "CLICKPORT_CONFIG_FILEPATH"
)

var (
	errNoSigningSecret = fmt.Errorf("must specify `%s` environment variable", signingSecretEnvKey)
)

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

	signingSecret, hasSigningSecret := os.LookupEnv(signingSecretEnvKey)
	if !hasSigningSecret {
		return nil, errNoSigningSecret
	}

	config := &Config{
		ClickportScripts: &clickportScripts,
		SigningSecret:    signingSecret,
	}

	return config, nil
}

func LoadConfig() (*Config, error) {
	configFilepath := os.Getenv(configFilepathEnvKey)
	if configFilepath == "" {
		configFilepath = defaultConfigFilepath
	}

	return loadConfig(&configFilepath)
}
