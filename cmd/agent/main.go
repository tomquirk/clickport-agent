package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type RuntimeScriptParameter struct {
	Type        string
	Description string
}

type RuntimeScript struct {
	Description string
	Parameters  []RuntimeScriptParameter
}

type RuntimeScripts map[string]RuntimeScript

func loadScriptConfig() RuntimeScripts {
	yamlFile, err := ioutil.ReadFile("examples/example.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	runtimeScripts := make(RuntimeScripts)

	err = yaml.Unmarshal(yamlFile, &runtimeScripts)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return runtimeScripts
}

func main() {
	runtimeScripts := loadScriptConfig()
	// test loader works
	for k := range runtimeScripts {
		fmt.Println(runtimeScripts[k].Parameters[0].Description)
	}
}
