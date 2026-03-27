// Package checker provides a lightweight, zero-dependency core framework
// for periodic health monitoring and peer-to-peer status exchange.
//
// At its core, the package is built around the [Checker] type, which runs
// a series of user-defined [Check] functions at a given interval. A check
// evaluates the health of a specific component (e.g., a database, an API,
// or system metrics) and returns a [State] (OK, Warn, Fail, or Skipped)
// along with a descriptive message.
//
// When a check changes its state or remains in a non-OK state, the
// [Checker] dispatches the new state to all registered [Notifier] functions.
// This allows you to easily plug in alerting systems like email, Slack,
// or Pushover.
//
// # Configuration
//
// You can build your Checker programmatically using [New], [AddCheck], and
// [AddNotifier], or you can load an entire setup dynamically from a JSON file
// using [ReadConfig]. The JSON configuration is especially powerful when
// combined with the "lib" subpackage, which registers makers for many common
// checks and notifiers.
//
// # Peer-to-Peer Monitoring
//
// Checker natively supports a decentralized monitoring approach. By using
// [EnableServer] and [AddPeer], a Checker instance can act as both a
// monitoring server and a client. It will periodically pull the state
// of its peers and integrate their check results into its own state tree.
// Because peer states are propagated transitively, a single node can give
// you a global view of your entire infrastructure.
//
// # Inner Workings & Concurrency
//
// The [Checker] utilizes a highly concurrent architecture:
//   - A single ticker loop manages the scheduling of checks, ensuring they
//     are spread evenly across the defined interval to prevent CPU/Network spikes.
//   - Each [Check] is executed in its own goroutine. Long-running checks will
//     not block the ticker. If a check is still running when its next execution
//     is due, the overlap is skipped safely using non-blocking mutexes.
//   - [Notifier] functions are also spawned in separate goroutines, ensuring
//     that slow alerting APIs do not stall the monitoring pipeline.
//   - The internal state tree is protected by read-write mutexes (`sync.RWMutex`),
//     allowing safe, lock-free deep copies for the HTTP status page and JSON
//     API endpoints.
package checker
