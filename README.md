# Checker -- System Monitoring For Go Programmers

A very simple, reduced, and lightweight replacement for complex monitoring systems like Nagios.

## Motivation

Monitoring systems often come with a steep learning curve, requiring you to understand complex abstractions like Hosts, HostGroups, Services, and ServiceGroups, along with custom configuration languages or DSLs.

**Checker** takes a different approach:
- **No new syntax to learn:** It can be configured entirely in Go.
- **IDE Support:** Because it's just Go code, you get full autocompletion, type safety, and debugging support right in your IDE.
- **Keep it simple:** There are only **Checks** and **Notifiers**. No complex hierarchies. You define what to check and who to notify when the state changes.
- **Standard Library:** A rich collection of commonly used checks (HTTP, Ping, CPU, Memory, SSH, etc.) is ready to be included. See the [Standard Library Documentation (`lib/README.md`)](lib/README.md) for a full list of available components.
- **JSON Configuration (Optional):** If you prefer, you can easily load your entire setup from a simple JSON file using the provided `lib` components.
- **Peer-to-Peer Monitoring:** Every Checker instance can act as a server. Instances can monitor each other, providing a decentralized and resilient monitoring network where peers exchange their global state trees.

## Examples

### 1. A Simple Custom Check in Go

You can write your checks directly in Go. A check is simply a function that takes a context and the current `CheckState`, and returns a new state and a message.

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
	c.AddCheck("My Custom Check", func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		// Your custom logic here...
		if time.Now().Second()%2 == 0 {
			return chkr.OK, "Everything is fine"
		}
		return chkr.Fail, "Something went wrong!"
	})

	// 3. Add a simple notifier to print to the console
	c.AddNotifier(func(ctx context.Context, name string, cs chkr.CheckState) {
		fmt.Printf("[%s] %s is now: %s (%s)\n", time.Now().Format(time.TimeOnly), name, cs.State, cs.Message)
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

The `checker/lib` package provides ready-to-use checks (like Ping, HTTP, DNS, System Metrics) and notifiers (like Logging, Pushover, Rate-Limiting).

```go
package main

import (
	"context"
	"net/http"
	"time"

	chkr "github.com/tweithoener/checker"
	"github.com/tweithoener/checker/lib"
)

func main() {
	c := chkr.New()
	
	// Add checks and a notifier from the standard library
	c.AddCheck("Ping Webserver", lib.Ping("example.com", 50, 300))
	c.AddCheck("Check My Website", lib.Http("GET", "https://example.com/", http.StatusOK))
	c.AddNotifier(lib.Logging("ALERT: "))

	c.SetInterval(5 * time.Second)
	c.Start()
	// ... (Shutdown logic)
}
```

### 3. Using the Standard Library and JSON Configuration

You can configure the exact same setup using a JSON file. This is perfect for deploying Checker as a standalone binary without recompiling.

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
	"log"
	"os"

	chkr "github.com/tweithoener/checker"
	// Blank import registers all standard checks and notifiers
	_ "github.com/tweithoener/checker/lib"
)

func main() {
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("can't open config file: %v", err)
	}
	defer f.Close()

	c := chkr.New()
	if err := c.ReadConfig(f); err != nil {
		log.Fatalf("config error: %v", err)
	}

	c.Start()
	// ... (Wait and Shutdown logic)
}
```

### 4. Distributed Peer-to-Peer Monitoring

Enable the built-in HTTP server on one instance and monitor it from another. This allows you to build a resilient monitoring grid where instances "watch the watchers".

**Node A (Server):**
```json
{
  "Interval": 10,
  "Server": {
    "Enabled": true,
    "Listen": ":8080"
  },
  "Checks": [
    {
      "Maker": "Cpu",
      "Name": "CPU Usage",
      "Args": { "WarnPercent": 80, "FailPercent": 90 }
    }
  ]
}
```

**Node B (Monitoring Node A):**
```json
{
  "Interval": 10,
  "Peers": [
    "192.168.1.50:8080"
  ],
  "Notifiers": [
    {
      "Maker": "Logging",
      "Args": { "Prefix": "GLOBAL-STATE: " }
    }
  ]
}
```

When Node A reports a non-OK state, Node B automatically pulls that state, summarizes the failure, and triggers its own notifiers. By visiting `http://192.168.1.50:8080/` in your browser, you also get a neat HTML dashboard of the current health!
