package main

import (
	"context"
	"github.com/sinderpl/AsyncTaskProcessor/api"
	"github.com/sinderpl/AsyncTaskProcessor/queue"
	"github.com/sinderpl/AsyncTaskProcessor/storage"
	"github.com/sinderpl/AsyncTaskProcessor/task"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var (
	cfg        = Config{}
	configPath = "config/Configuration.yml"
)

type Config struct {
	Api struct {
		ListenAddr string `yaml:"listenAddr"`
	} `yaml:"api"`
	Queue struct {
		MaxBufferSize  int `yaml:"maxBufferSize"`
		WorkerPoolSize int `yaml:"workerPoolSize,omitempty"`
		MaxTaskRetry   int `yaml:"maxTaskRetry"`
	} `yaml:"queue"`
	Storage struct {
		User     string `yaml:"user"`
		DBName   string `yaml:"dbname"`
		Password string `yaml:"password"`
	} `yaml:"storage"`
}

func main() {

	cfgFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read cfg file: %v", err)
	}

	if err := yaml.Unmarshal(cfgFile, &cfg); err != nil {
		log.Fatalf("Failed to unmarshal YAML cfg data: %v", err)
	}

	// TODO graceful shutdown
	mainCtx := context.Background()

	storage, err := storage.NewPostgresStore(cfg.Storage.User, cfg.Storage.DBName, cfg.Storage.Password)

	if err != nil {
		log.Fatalf("Failed to initialise database: %v", err)
	}
	err = storage.Init()
	if err != nil {
		log.Fatalf("Failed to run database migration: %v", err)
	}

	taskChan := make(chan []*task.Task)

	q, err := queue.CreateQueue(mainCtx,
		queue.WithMainQueue(&taskChan),
		queue.WithMaxBufferSize(cfg.Queue.MaxBufferSize),
		queue.WithMaxWorkerPoolSize(cfg.Queue.WorkerPoolSize),
		queue.WithMaxTaskRetry(cfg.Queue.MaxTaskRetry),
		queue.WithStorage(storage))

	if err != nil {
		log.Fatalf("failed to initialize queue: %v", err)
	}

	q.Start()

	server := api.CreateApiServer(
		api.WithListenAddr(cfg.Api.ListenAddr),
		api.WithQueue(&taskChan),
		api.WithStorage(storage))

	server.Run()

}
