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
		maxQueueSize   int `yaml:"maxQueueSize" json:"maxQueueSize,omitempty"`
		workerPoolSize int `json:"workerPoolSize,omitempty"`
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

	mainCtx := context.Background()
	taskChan := make(chan []*task.Task)

	//stopChan := make(chan os.Signal, 2)
	//signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	//
	//go func() {
	//	<-stopChan
	//	slog.Info("os exit signal called, shutting down")
	//	close(taskChan)
	//	mainCtx.Done()
	//}()

	// TODO add main CTX

	q, err := queue.CreateQueue(mainCtx,
		queue.WithMainQueue(&taskChan),
		queue.WithMaxQueueSize(config.Queue.maxQueueSize),
		queue.WithMaxWorkerPoolSize(config.Queue.maxQueueSize))

	if err != nil {
		log.Fatalf("failed to initialize queue: %v", err)
	}

	q.Start()

	server := api.CreateApiServer(
		api.WithListenAddr(config.Api.ListenAddr),
		api.WithQueue(&taskChan)) // TODO handle passing this chan to task better

	server.Run()

}
