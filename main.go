package main

import (
	"github.com/sinderpl/AsyncTaskProcessor/queue"
	"log"
	"os"

	"github.com/sinderpl/AsyncTaskProcessor/api"
	"github.com/sinderpl/AsyncTaskProcessor/task"

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
	Queue struct {
		maxQueueSize int32 `yaml:"maxQueueSize"`
	} `yaml:"queue"`
}

func main() {

	cfgFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(cfgFile, &config); err != nil {
		log.Fatalf("Failed to unmarshal YAML config data: %v", err)
	}

	// TODO add main CTX

	taskChan := make(chan []task.Task)

	q := queue.CreateQueue(
		queue.WithMainQueue(&taskChan),
		queue.WithMaxQueueSize(config.Queue.maxQueueSize))

	q.Start()

	server := api.CreateApiServer(
		api.WithListenAddr(config.Api.ListenAddr),
		api.WithQueue(&taskChan)) // TODO handle passing this chan to task better

	server.Run()

}
