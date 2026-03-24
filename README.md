# Checker -- System Monitoring For Go Programmers

A very simple, reduced, and lightweight replacement for complex monitoring systems like Nagios. 

## Motivation

Monitoring systems often come with a steep learning curve, requiring you to understand complex abstractions like Hosts, HostGroups, Services, and ServiceGroups, along with custom configuration languages or DSLs.

**Checker** takes a different approach:
- **No new syntax to learn:** It can be configured entirely in Go.
- **IDE Support:** Because it's just Go code, you get full autocompletion, type safety, and debugging support right in your IDE.
- **Keep it simple:** There are only **Checks** and **Notifiers**. No complex hierarchies. You define what to check and who to notify when the state changes. 
- **JSON Configuration (Optional):** If you prefer, you can easily load your setup from a simple JSON file using the provided `lib` components.
- **Peer-to-Peer Monitoring:** Every Checker instance can act as a server. Instances can monitor each other, providing a decentralized and resilient monitoring network with summary reporting.

## Examples

### 1. A Simple Custom Check in Go

You can write your checks directly in Go. A check is simply a function that takes a context and history, and returns a state and a message.

```go
package main

import (
	"context"
	"fmt"
	"time"

	chkr "github.com/tweithoener/checker"
)

func main() {
	// 1. Create a new Checker
	c := chkr.New()

	// 2. Add a custom check
	c.AddCheck("My Custom Check", func(ctx context.Context, h chkr.History) (chkr.State, string) {
		// Your custom logic here...
		if time.Now().Second()%2 == 0 {
			return chkr.OK, "Everything is fine"
		}
		return chkr.Fail, "Something went wrong!"
	})

	// 3. Add a simple notifier to print to the console
	c.AddNotifier(func(ctx context.Context, name string, h chkr.History) {
		fmt.Printf("[%s] Check '%s' state changed to: %s (%s)\n", time.Now().Format(time.RFC3339), name, h.State(), h.Message())
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
```

### 2. Using the Standard Library from Go
The `checker/lib` package provides ready-to-use checks (like Ping, HTTP, DNS, etc.) and notifiers (like Logging, Pushover, Rate-Limiting).
You can easily include these into your checker.

```go
	...
	
	// 2. Add checks and notifier from the standard library
	c.AddCheck("Ping Webserver", lib.Ping("example.com", 50, 300))
	c.AddCheck("Check My Website ", lib.Http("GET", "https://example.com/", http.StatusOK))
	c.AddNotifier(lib.Logging("ALERT: "))

	...
```


### 3. Using the Standard Library and JSON Configuration

The  checks from the Standard Library can also be configured easily via a JSON file.

**`config.json`**
```json
{
  "Interval": 5,
  "Checks": [
    {
      "Maker": "Http",
      "Name": "Check My Website",
      "Args": {
        "Method": "GET",
        "Url": "http://example.com",
        "Expected": 200
      }
    }
  ],
  "Notifiers": [
    {
      "Maker": "Less",
      "Args": {
        "Notifier": {
          "Maker": "Logging",
          "Args": {
            "Prefix": "ALERT: "
          }
        }
      }
    }
  ]
}
```

**`main.go`**
```go
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
```

### 4. Distributed Peer-to-Peer Monitoring

Enable the built-in HTTP server on one instance and monitor it from another. This allows you to build a resilient monitoring grid where instances "watch the watchers".

**Remote Instance (`config.json`):**
```json
{
  "Server": {
    "Enabled": true,
    "Listen": ":8080"
  },
  "Checks": [ ... ]
}
```

**Local Instance Monitoring the Remote (`config.json`):**
```json
{
  "Checks": [
    {
      "Maker": "Peer",
      "Name": "Remote Office Checker",
      "Args": {
        "Address": "http://192.168.1.50:8080/"
      }
    }
  ]
}
```

When the remote instance reports a non-OK state, the `Peer` check automatically summarizes the failure, showing the number of affected checks and the most recent error.
