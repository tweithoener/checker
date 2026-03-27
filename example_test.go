package checker_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	chkr "github.com/tweithoener/checker"
	"github.com/tweithoener/checker/lib"
)

// Example demonstrates the basic usage of the checker package,
// including creating a new checker, adding a check, and a notifier.
func Example() {
	// 1. Create a new Checker instance.
	c := chkr.New()
	c.SetName("ExampleChecker")

	// 2. Add a simple custom check function that fails.
	c.AddCheck("HealthCheck", func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		// your logic goes here
		return chkr.Fail, "System is down"
	})

	// 3. Add a simple notifier to print status changes.
	// We use a channel to wait for the notification in this example.
	c.AddNotifier(lib.Debug(""))

	// 4. Start the checker.
	c.SetInterval(100 * time.Millisecond)
	c.Start()

	time.Sleep(180 * time.Millisecond)

	// 5. Gracefully shut down the checker.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	c.Shutdown(ctx)

	// Output:
	// Failed: HealthCheck: System is down (1x since 2001-01-01 01:01:01)
}

// Example_readConfig shows how to configure a checker instance using a JSON string.
// Note: This example uses a simplified JSON structure. In a real application,
// you would import the 'lib' package to use the standard checks and notifiers.
func Example_readConfig() {
	configJSON := `{
		"Checks": [
			{
				"Maker": "Http",
				"Name": "Web Check",
				"Args": {
					"Method": "GET",
					"Url": "http://example.com/doesnotexist",
					"Expected": 200
				}
			}
		],
		"Notifiers": [
			{
				"Maker": "Debug",
				"Args": {}
			}
		]
	}`

	// 1. Create new Checker instance and read the JSON configuraiton
	c := chkr.New()
	if err := c.ReadConfig(strings.NewReader(configJSON)); err != nil {
		fmt.Printf("Config error: %v\n", err)
		return
	}

	// 2. Start the checker.
	c.SetInterval(100 * time.Millisecond)
	c.Start()

	time.Sleep(180 * time.Millisecond)

	// 3. Gracefully shut down the checker.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	c.Shutdown(ctx)

	// Output:
	// Failed: Web Check: Unexpected status code: 404 (expected 200) (1x since 2001-01-01 01:01:01)
}
