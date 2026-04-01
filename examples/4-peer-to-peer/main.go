package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	chkr "github.com/tweithoener/checker"
	_ "github.com/tweithoener/checker/lib"
)

func main() {
	// Configure global structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	configPath := flag.String("config", "one.json", "path to configuration file (e.g. one.json or another.json)")
	flag.Parse()

	f, err := os.Open(*configPath)
	if err != nil {
		slog.Error("can't open config file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	c := chkr.New()
	if err := c.ReadConfig(f); err != nil {
		slog.Error("can't configure checker from config file", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c.SetInterval(5 * time.Second)
	c.Start()
	slog.Info("started checker", "config", *configPath)

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Shutdown(shutdownCtx)
}
