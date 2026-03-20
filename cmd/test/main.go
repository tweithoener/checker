package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	chkr "github.com/tweithoener/checker"
	_ "github.com/tweithoener/checker/lib"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	f, err := os.Open("full.json")
	if err != nil {
		log.Fatalf("can't open config file: %v", err)
	}
	c := chkr.New()
	if err := c.ReadConfig(f); err != nil {
		log.Fatalf("can't configure checker from config file: %v", err)
	}
	c.SetInterval(5 * time.Second)
	c.Start()

	select {
	case <-ctx.Done():
	case <-time.After(60 * time.Second):
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c.Shutdown(ctx)
}
