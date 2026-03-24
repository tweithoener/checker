# Checker Standard Library (`lib`)

The `lib` package provides a robust collection of built-in **Checks** and **Notifiers** for the Checker framework. These components are automatically registered when the package is imported and can be easily configured via JSON or used directly in your Go code.

## Available Checks

### Network & Connectivity
* **`Http`**: Performs an HTTP request (e.g., GET) and verifies that the response matches the expected status code (e.g., 200).
* **`Ping`**: Sends ICMP echo requests to a specific IP address to verify connectivity. Allows configuration of warning and failure latency thresholds.
* **`Dns`**: Verifies that a specific hostname resolves to an expected IP address using a custom DNS server.
* **`Proxy`**: Performs an HTTP request through a specified proxy server and checks the response status code.
* **`Peer`**: Connects to the built-in HTTP server of a remote Checker instance. It summarizes the remote state, alerting if any remote checks are failing or warning.

### System & Hardware (Powered by gopsutil)
* **`Cpu`**: Monitors the total CPU usage percentage of the system.
* **`Mem`**: Monitors the virtual memory (RAM) usage percentage.
* **`Swap`**: Monitors the swap memory usage percentage.
* **`Disk`**: Monitors the disk space usage percentage on a specific path (e.g., `/`).
* **`Load`**: Monitors the 5-minute system load average.
* **`Uptime`**: Verifies that the system has been running for at least a specified minimum time (useful for detecting unexpected reboots).

### Processes & Execution
* **`ProcExists`**: Verifies if at least one process with the exact given name (e.g., `nginx`) is currently running.
* **`SysProcs`**: Monitors the total number of running processes on the system to prevent PID exhaustion.
* **`Cmd`**: Executes a local shell command and evaluates its exit status.
* **`Ssh`**: Connects to a remote host via SSH and can execute commands to verify remote state.

### Utilities
* **`Fail`**: A wrapper check that inverts the result of another check. It succeeds only if the inner check fails (useful for testing negative scenarios, e.g., verifying an endpoint is *not* reachable).

---

## Available Notifiers

* **`Logging`**: A simple notifier that outputs check results to the standard Go log, prepended with a custom prefix.
* **`Pushover`**: Sends real-time push notifications to your mobile devices using the Pushover service. It routes different states (Fail, Warn, OK) with varying priorities and notification sounds.
* **`Less`**: A wrapper notifier that rate-limits notifications from an inner notifier. It prevents alert fatigue by suppressing repeated alerts for the same state, only notifying on state changes or after a prolonged period (e.g., hourly).

## Usage Example (Go)

The primary way to use Checker is directly within your Go code. This provides type safety, autocompletion, and the ability to seamlessly mix built-in checks with your own custom logic.

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

	// Add a built-in notifier (wrapped in rate-limiting)
	logger := lib.Logging("[PROD] ")
	c.AddNotifier(lib.Less(logger))

	c.SetInterval(30 * time.Second)
	c.Start()

	// ... wait for shutdown ...
}
```

## Usage Example (JSON)

Alternatively, you can easily mix and match these components using a JSON configuration file:

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
    },
    {
      "Maker": "Cpu",
      "Name": "System CPU",
      "Args": {
        "WarnPercent": 75.0,
        "FailPercent": 90.0
      }
    }
  ],
  "Notifiers": [
    {
      "Maker": "Less",
      "Args": {
        "Notifier": {
          "Maker": "Pushover",
          "Args": {
            "Prefix": "[PROD]",
            "App": "YOUR_APP_TOKEN",
            "Recipient": "YOUR_USER_KEY"
          }
        }
      }
    }
  ]
}
```
