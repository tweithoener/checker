package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	chkr "github.com/tweithoener/checker"
	// Import the lib package to register the standard checks and notifiers
	_ "github.com/tweithoener/checker/lib"
)

func main() {
	// Configure global structured logging as JSON for this example.
	// This will affect both the checker core and the structured log notifier.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Open the configuration file
	f, err := os.Open("config.json")
	if err != nil {
		slog.Error("can't open config file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	c := chkr.New()

	// Load checks and notifiers from the JSON config
	if err := c.ReadConfig(f); err != nil {
		slog.Error("can't configure checker from config file", "error", err)
		os.Exit(1)
	}

	c.SetInterval(5 * time.Second)
	c.Start()

	// Keep running
	time.Sleep(30 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Shutdown(ctx)
}
