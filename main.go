package main

import (
	"context"
	"github.com/sinderpl/AsyncTaskProcessor/api"
	"github.com/sinderpl/AsyncTaskProcessor/queue"
	"github.com/sinderpl/AsyncTaskProcessor/task"
	"gopkg.in/yaml.v2"
	"log"
	"os"
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
		maxBufferSize  int `yaml:"maxBufferSize"`
		workerPoolSize int `yaml:"workerPoolSize,omitempty"`
		maxTaskRetry   int `yaml:"maxTaskRetry"`
		bufferSize     int `yaml:"bufferSize"`
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

	// TODO graceful shutdown
	mainCtx := context.Background()
	taskChan := make(chan []*task.Task)

	q, err := queue.CreateQueue(mainCtx,
		queue.WithMainQueue(&taskChan),
		queue.WithMaxBufferSize(config.Queue.maxBufferSize),
		queue.WithMaxWorkerPoolSize(config.Queue.workerPoolSize),
		queue.WithMaxTaskRetry(config.Queue.maxTaskRetry))

	if err != nil {
		log.Fatalf("failed to initialize queue: %v", err)
	}

	q.Start()

	server := api.CreateApiServer(
		api.WithListenAddr(config.Api.ListenAddr),
		api.WithQueue(&taskChan))

	server.Run()

}
