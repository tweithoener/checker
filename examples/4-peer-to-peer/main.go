package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	chkr "github.com/tweithoener/checker"
	_ "github.com/tweithoener/checker/lib"
)

func main() {
	configPath := flag.String("config", "server.json", "path to configuration file (e.g. server.json or client.json)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	f, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("can't open config file: %v", err)
	}
	defer f.Close()

	c := chkr.New()
	if err := c.ReadConfig(f); err != nil {
		log.Fatalf("can't configure checker from config file: %v", err)
	}

	c.SetInterval(5 * time.Second)
	c.Start()
	log.Printf("Started checker with config: %s", *configPath)
	log.Println("Press Ctrl+C to exit")

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Shutdown(shutdownCtx)
}
