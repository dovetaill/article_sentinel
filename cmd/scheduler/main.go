package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	queueasynq "github.com/dovetaill/article-sentinel/internal/queue/asynq"
	"github.com/dovetaill/article-sentinel/internal/scheduler"
)

func main() {
	configPath := flag.String("config", envOrDefault("CONFIG_PATH", "configs/config.yaml"), "config file path")
	flag.Parse()

	rt, err := bootstrap.BuildSchedulerRuntime(*configPath)
	if err != nil {
		log.Fatalf("build scheduler runtime: %v", err)
	}
	defer func() {
		if closeErr := rt.Shutdown(); closeErr != nil {
			log.Printf("shutdown resources: %v", closeErr)
		}
	}()

	client, err := queueasynq.NewClient(rt)
	if err != nil {
		log.Fatalf("build scheduler client: %v", err)
	}

	cronScheduler := scheduler.New()
	if err := scheduler.RegisterJobs(
		cronScheduler,
		rt,
		scheduler.NewAsynqEnqueuer(client, rt.Config.Queue.Asynq.QueueName),
	); err != nil {
		log.Fatalf("register scheduler jobs: %v", err)
	}

	cronScheduler.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop
	log.Printf("received signal: %s", sig)

	ctx := cronScheduler.Stop()
	<-ctx.Done()
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
