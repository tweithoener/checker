// Package lib provides a robust collection of built-in Checks and Notifiers
// for the checker framework. It functions as the "Standard Library" of the
// checker ecosystem, allowing you to monitor common infrastructure without
// writing custom check logic.
//
// Included components range from system metrics (CPU, Memory, Disk, Uptime)
// powered by gopsutil, to network diagnostics (Ping, DNS, HTTP, Proxy),
// and command execution (Cmd, SSH). It also provides notifiers for Logging,
// debugging, and mobile alerts via Pushover.
//
// # Usage in Go
//
// You can instantiate checks directly using their exported functions (e.g.,
// [Http], [Cpu], [Ping]) and add them to a Checker instance via `AddCheck`.
//
// # The JSON Registry (Inner Workings)
//
// What makes the lib package powerful is its integration with the checker's
// JSON configuration system.
//
// Upon package initialization (in `init()`), the lib package automatically
// registers a series of "Makers" (implementing [checker.CheckMaker] and
// [checker.NotifierMaker]) with the core checker registry. A Maker knows how
// to translate a generic JSON payload into a concrete arguments struct
// (like [HttpArgs] or [CpuArgs]) and eventually instantiate the actual
// Check or Notifier function.
//
// This enables a completely data-driven monitoring setup: you can compile
// a single Go binary that imports this package anonymously (`_ "github.com/.../lib"`),
// and users can configure complex monitoring pipelines entirely through JSON,
// including recursive wrappers like [Less] and [Fail].
//
// # Testing and Mocking
//
// The lib package is designed with testability in mind. Most hardware- and
// network-dependent calls (like reading from /proc, spawning OS processes,
// or dialing raw ICMP sockets) are abstracted into package-level variables
// internally. This allows the test suite to safely mock these interactions
// without relying on the host system's state or privileges.
package lib
