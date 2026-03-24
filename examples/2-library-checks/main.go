package main

import (
	"context"
	"net/http"
	"time"

	chkr "github.com/tweithoener/checker"
	"github.com/tweithoener/checker/lib"
)

func main() {
	// 1. Create a new Checker
	c := chkr.New()

	// 2. Add checks and notifier from the standard library
	c.AddCheck("Ping Webserver", lib.Ping("example.com", 50, 300))
	c.AddCheck("Check My Website ", lib.Http("GET", "https://example.com/", http.StatusOK))
	c.AddNotifier(lib.Logging("ALERT: "))

	// 3. Set the check interval and start
	c.SetInterval(2 * time.Second)
	c.Start()

	// Let it run for a while
	time.Sleep(10 * time.Second)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Shutdown(ctx)
}
