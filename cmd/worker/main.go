package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	queueasynq "github.com/dovetaill/article-sentinel/internal/queue/asynq"
)

func main() {
	configPath := flag.String("config", envOrDefault("CONFIG_PATH", "configs/config.yaml"), "config file path")
	flag.Parse()

	rt, err := bootstrap.BuildWorkerRuntime(*configPath)
	if err != nil {
		log.Fatalf("build worker runtime: %v", err)
	}
	defer func() {
		if closeErr := rt.Shutdown(); closeErr != nil {
			log.Printf("shutdown resources: %v", closeErr)
		}
	}()

	srv, err := queueasynq.NewServer(rt)
	if err != nil {
		log.Fatalf("build worker server: %v", err)
	}
	mux := queueasynq.RegisterHandlers(rt)

	if err := srv.Start(mux); err != nil {
		log.Fatalf("worker server: %v", err)
	}
	defer srv.Shutdown()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop
	log.Printf("received signal: %s", sig)
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
