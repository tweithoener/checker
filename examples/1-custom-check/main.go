package main

import (
	"context"
	"log/slog"
	"time"

	chkr "github.com/tweithoener/checker"
)

func main() {
	// 1. Create a new Checker
	c := chkr.New()

	// 2. Add a custom check
	c.AddCheck("My Custom Check", func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		// Your custom logic here...
		if time.Now().Second()%2 == 0 {
			return chkr.OK, "Everything is fine"
		}
		return chkr.Fail, "Something went wrong!"
	})

	// 3. Add a simple structured console notifier
	c.AddNotifier(func(ctx context.Context, cs chkr.CheckState) {
		slog.Info("notifier event",
			"check", cs.Name,
			"state", cs.State,
			"message", cs.Message,
		)
	})

	// 4. Set the check interval and start
	c.SetInterval(2 * time.Second)
	c.Start()

	// Let it run for a while
	time.Sleep(10 * time.Second)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.Shutdown(ctx)
}
