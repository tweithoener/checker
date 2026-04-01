# Checker Standard Library (`lib`) 🔋

Welcome to the **Standard Library** of the Checker ecosystem!

Why write custom checks for common infrastructure when you can just plug and play? The `lib` package provides a robust collection of built-in checks and notifiers, ready to be used in your Go code or dynamically configured via JSON. No new DSL to learn, just pure Go!

## 🕵️‍♂️ Available Checks

Here is a glimpse of what you get out of the box:

### Network & Connectivity
* **`Http`**: Hits an HTTP/HTTPS endpoint and validates the expected status code.
* **`Ping`**: Sends ICMP echo requests to measure latency and verify connectivity.
* **`Dns`**: Queries a specific DNS server to verify domain resolution.
* **`Proxy`**: Tests HTTP proxy servers to ensure they are forwarding requests correctly.
* **`Peer`**: Connects to the built-in HTTP server of a remote Checker instance to monitor the watcher.

### System & Hardware (Powered by gopsutil)
* **`Cpu`**: Monitors CPU utilization and alerts on high load.
* **`Mem`**: Tracks virtual memory (RAM) usage.
* **`Disk`**: Keeps an eye on your disk space usage.
* **`Load`**: Monitors the 5-minute system load average.
* **`Swap`**: Monitors swap memory usage.
* **`Uptime`**: Ensures the system has been running for a minimum required time.

### Processes & Execution
* **`Cmd`**: Executes any local CLI command and analyzes its exit code or stdout.
* **`Ssh`**: Attempts to connect, authenticate, and run commands against an SSH server.
* **`ProcExists`**: Verifies if a vital background process (e.g., `nginx`, `postgres`) is currently running.
* **`SysProcs`**: Tracks the total number of running processes to detect fork bombs or leaks.

### Utilities
* **`Fail`**: A meta-check useful for debugging or recursive logic.

---

## 📢 Available Notifiers

When things go wrong, you need to know immediately.

* **`Email`**: Sends full-featured, customizable HTML or plaintext alerts via SMTP.
* **`Pushover`**: Sends real-time push notifications straight to your phone, complete with priority routing and custom notification sounds!
* **`Logging`**: The classic. Dumps check results to the standard Go log with a custom prefix.
* **`Less`**: A smart meta-notifier that limits alert spam.

## 🛠️ Usage

### From Go Code

Using the standard library directly in your Go code gives you perfect type safety and IDE support:

```go
package main

import (
	"context"
	"time"

	chkr "github.com/tweithoener/checker"
	"github.com/tweithoener/checker/lib"
)

func main() {
	c := chkr.New()

	// Add built-in checks directly using their exported functions
	c.AddCheck("API Health", lib.Http("GET", "https://api.example.com/health", 200))
	c.AddCheck("System CPU", lib.Cpu(75.0, 90.0))

	// Add a built-in structured logging notifier (wrapped in rate-limiting)
	logger := lib.Logging(nil)
	c.AddNotifier(lib.Less(logger))

	c.SetInterval(30 * time.Second)
	c.Start()

	// ... wait for shutdown ...
}
```

### From JSON Configuration

Prefer no recompiles? You can define everything in a `config.json`. The `lib` package automatically registers "Makers" for all components during initialization.

```json
{
  "Checks": [
    {
      "Maker": "Http",
      "Name": "API Health",
      "Args": {
        "Method": "GET",
        "Url": "https://api.example.com/health",
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
            "Attributes": {
              "env": "production",
              "team": "backend"
            }
          }
        }
      }
    }
  ]
}
```

Just remember to blank-import the lib package in your `main.go` so the init hooks fire:

```go
import _ "github.com/tweithoener/checker/lib"
```