package config

import (
	"io/ioutil"
	"log"
	"os"

	rt "github.com/runtime-hq/runtime-agent/internal/runtime"
	"gopkg.in/yaml.v2"
)

const defaultConfigFilepath = ".runtime-config.yml"

type Config struct {
	RuntimeScripts *rt.RuntimeScripts
}

func loadConfig(configFilepath *string) *Config {
	yamlFile, err := ioutil.ReadFile(*configFilepath)
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}

	runtimeScripts := make(rt.RuntimeScripts)

	err = yaml.Unmarshal(yamlFile, &runtimeScripts)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	config := &Config{
		RuntimeScripts: &runtimeScripts,
	}

	return config
}

func LoadConfig() *Config {
	configFilepath := os.Getenv("CONFIG_FILEPATH")
	if configFilepath == "" {
		configFilepath = defaultConfigFilepath
	}

	return loadConfig(&configFilepath)

}
