# Checker 🚀 System Monitoring For Go Programmers

Welcome to **Checker** — a refreshingly simple, dependency-light replacement for complex monitoring systems like Nagios or Zabbix.

## Why Checker?

Monitoring systems often come with a steep learning curve. You're forced to wrestle with complex abstractions (Hosts, HostGroups, ServiceGroups) and decipher proprietary configuration languages or DSLs.

**Checker** takes a fundamentally different approach. We believe monitoring should be as straightforward as writing code:

- 🧠 **No new DSL to learn, just use Go:** Configure your entire monitoring setup directly in Go.
- 🛠️ **Native IDE Support:** Because it's just Go code, you get full autocompletion, type safety, and debugging support right out of the box.
- 🎯 **Keep it Simple:** There are only **Checks** (what to monitor) and **Notifiers** (who to alert). No convoluted hierarchies.
- 🔋 **Batteries Included:** A rich standard library (`checker/lib`) of ready-to-use checks (HTTP, Ping, CPU, Memory, SSH, DNS) and notifiers (Email, Logging, Pushover) is included.
- 🌐 **Peer-to-Peer Monitoring:** Every Checker instance can act as a server. Build a decentralized, resilient monitoring grid where nodes exchange their global state trees seamlessly.
- 📜 **JSON Configuration:** Prefer config files over binaries? Easily load your entire setup from a clean JSON file.

---

## Quick Start Guide

We've designed Checker to get out of your way. Here are the most common ways to use it. You can find fully runnable versions of these snippets in the [`./examples`](./examples) directory!

### 1. Write a Custom Check in Go

Writing a custom check is as simple as defining a function that takes a context and returns a state. Perfect for checking your internal app health!

*From `examples/1-custom-check/main.go`*
```go
package main

import (
	"context"
	"fmt"
	"time"

	chkr "github.com/tweithoener/checker"
)

func main() {
	// 1. Create a new Checker instance
	c := chkr.New()

	// 2. Add your custom check logic
	c.AddCheck("My Custom Check", func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
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

	// 4. Start the engine
	c.SetInterval(2 * time.Second)
	c.Start()

	time.Sleep(10 * time.Second)
	c.Shutdown(context.Background())
}
```

### 2. Use the Standard Library

Don't reinvent the wheel. The `checker/lib` package provides battle-tested checks and notifiers.

*From `examples/2-library-checks/main.go`*
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
	
	// Plug in robust checks with a single line
	c.AddCheck("Ping Webserver", lib.Ping("example.com", 50, 300))
	c.AddCheck("Check My Website", lib.Http("GET", "https://example.com/", http.StatusOK))
	
	// Add an out-of-the-box structured logging notifier
	c.AddNotifier(lib.Logging(nil))

	c.SetInterval(5 * time.Second)
	c.Start()
	
	// ... (Wait and Shutdown logic)
}
```

### 3. Data-Driven: JSON Configuration

Deploy Checker as a standalone binary and configure it completely via JSON. No recompilation required!

*From `examples/3-json-config/config.json`*
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
      "Maker": "Logging",
      "Args": {
        "Prefix": "ALERT: "
      }
    }
  ]
}
```

*From `examples/3-json-config/main.go`*
```go
package main

import (
	"log"
	"os"

	chkr "github.com/tweithoener/checker"
	_ "github.com/tweithoener/checker/lib" // Register all standard library makers
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

### 4. Distributed Peer-to-Peer Grid

Instances can monitor each other. Set `Server.Enabled: true` on Node A, and add it to the `Peers` array of Node B. 

When Node A detects a failure, Node B pulls that state and triggers its own notifiers. Plus, you get a beautiful HTML dashboard of your entire infrastructure just by visiting the server's port in your browser! See `examples/4-peer-to-peer/` for a full multi-node demo.
