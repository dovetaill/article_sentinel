package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/dovetaill/go-auth-demo/internal/api/register"
	"github.com/dovetaill/go-auth-demo/internal/app/bootstrap"
)

func main() {
	configPath := flag.String("config", envOrDefault("CONFIG_PATH", "configs/config.yaml"), "config file path")
	flag.Parse()

	// server 负责加载配置、初始化日志并装配 HTTP 运行时。
	rt, err := bootstrap.BuildServerRuntime(*configPath)
	if err != nil {
		log.Fatalf("build server runtime: %v", err)
	}
	defer func() {
		if closeErr := rt.Shutdown(); closeErr != nil {
			log.Printf("shutdown resources: %v", closeErr)
		}
	}()

	// handler 内已经完成路由注册与中间件链装配。
	handler := register.NewRouter(rt)
	addr := rt.Config.App.Host + ":" + strconv.Itoa(rt.Config.App.Port)
	readTimeout := durationFromSeconds(rt.Config.HTTP.ReadTimeoutSeconds, 15)
	writeTimeout := durationFromSeconds(rt.Config.HTTP.WriteTimeoutSeconds, 15)
	idleTimeout := durationFromSeconds(rt.Config.HTTP.IdleTimeoutSeconds, 60)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("received signal: %s", sig)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http server shutdown: %v", err)
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func durationFromSeconds(seconds int, fallback int) time.Duration {
	if seconds <= 0 {
		seconds = fallback
	}
	return time.Duration(seconds) * time.Second
}
