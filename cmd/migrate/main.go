package main

import (
	"flag"
	"log"
	"os"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
)

func main() {
	configPath := flag.String("config", envOrDefault("CONFIG_PATH", "configs/config.yaml"), "config file path")
	flag.Parse()

	if err := bootstrap.RunMigrateCommand(*configPath); err != nil {
		log.Fatalf("run migrate command: %v", err)
	}
	log.Print("starter schema sync complete")
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
