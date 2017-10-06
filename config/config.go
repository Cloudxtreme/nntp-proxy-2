package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Configuration struct {
	Frontend frontendConfig
	Backend  []backendConfig
	Users    []user
}

type frontendConfig struct {
	FrontendAddr            string             `json:"frontendAddr"`
	FrontendPort            string             `json:"frontendPort"`
	FrontendTLS             bool               `json:"frontendTLS"`
	FrontendTLSCert         string             `json:"frontendTLSCert"`
	FrontendTLSKey          string             `json:"frontendTLSKey"`
	FrontendAllowedCommands []frontendCommands `json:"frontendAllowedCommands"`
}

type frontendCommands struct {
	FrontendCommand string `json:"frontendCommand"`
}

type backendConfig struct {
	BackendName  string `json:"backendName"`
	BackendAddr  string `json:"backendAddr"`
	BackendPort  string `json:"backendPort"`
	BackendTLS   bool   `json:"backendTLS"`
	BackendUser  string `json:"backendUser"`
	BackendPass  string `json:"backendPass"`
	BackendConns int    `json:"backendConns"`
}

type user struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type SelectedBackend struct {
	BackendName  string
	BackendAddr  string
	BackendPort  string
	BackendTLS   bool
	BackendUser  string
	BackendPass  string
	BackendConns int
}

func LoadConfig(path string) Configuration {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File Missing. ", err)
	}

	var config Configuration
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Config Parse Error: ", err)
	}

	return config
}
