package main

import (
	"context"
	"log"
	"os"
	"time"

	chkr "github.com/tweithoener/checker"
	// Import the lib package to register the standard checks and notifiers
	_ "github.com/tweithoener/checker/lib"
)

func main() {
	// Open the configuration file
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("can't open config file: %v", err)
	}
	defer f.Close()

	c := chkr.New()
	
	// Load checks and notifiers from the JSON config
	if err := c.ReadConfig(f); err != nil {
		log.Fatalf("can't configure checker from config file: %v", err)
	}

	c.SetInterval(5 * time.Second)
	c.Start()

	// Keep running
	time.Sleep(30 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Shutdown(ctx)
}
