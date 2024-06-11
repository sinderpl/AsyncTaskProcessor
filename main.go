package main

import (
	"fmt"
	"github.com/sinderpl/AsyncTaskProcessor/api"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	config     = Config{}
	configPath = "config/Configuration.yml"
)

type Config struct {
	Api struct {
		ListenAddr string `yaml:"listenAddr"`
	} `yaml:"api"`
}

func main() {

	cfgFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(cfgFile, &config); err != nil {
		log.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	fmt.Println(config.Api.ListenAddr)

	fmt.Println("Hello world")

	server := api.CreateApiServer(
		api.WithListenAddr(config.Api.ListenAddr))

	server.Run()
	fmt.Println("Hello world 2")

}
