package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Configuration struct {
	Frontend frontendConfig
	Backend  []backendConfig
	Users    []User
}

type frontendConfig struct {
	ListenAddr string `json:"ListenAddr"`
	ListenPort string `json:"ListenPort"`
	ListenTLS  bool   `json:"ListenTLS"`
}

type backendConfig struct {
	ServerAddr string `json:"ListenAddr"`
	ServerPort string `json:"ListenPort"`
	ServerTLS  bool   `json:"ListenTLS"`
	ServerUser string `json:"ServerUser"`
	ServerPass string `json:"ServerPass"`
	ServerConn int    `json:"ServerConn"`
}

type User struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
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
