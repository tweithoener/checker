package agent

import (
	"context"
	"slices"
	"sync"
	"time"

	chkr "github.com/tweithoener/checker"
)

// Event encapsulates a CheckState with metadata for the agent.
type Event struct {
	Name       string
	CheckState chkr.CheckState
	ReceivedAt time.Time
}

// LLMClient abstracts the communication with the AI model.
type LLMClient interface {
	Analyze(ctx context.Context, systemPrompt string, events []Event, lastAnalysis time.Time) (string, error)
}

// Trigger evaluates whether an analysis by the AI should be started.
type Trigger interface {
	ShouldTrigger(buffer *EventBuffer, lastAnalysis time.Time) bool
}

// AIAgent represents the AI assistant acting as a notifier.
type AIAgent struct {
	mu           sync.Mutex
	client       LLMClient
	buffer       *EventBuffer
	trigger      Trigger
	out          chkr.Notifier
	systemPrompt string
	lastAnalysis time.Time
}

// AIAgentOption defines functional options for the agent.
type AIAgentOption func(*AIAgent)

// TimeNow is a package-level variable to allow time mocking in tests.
var TimeNow = time.Now

const defaultSystemPrompt = `Analyze the following extract from an IT monitoring system log and generate a report.
Begin your report with a concise summary of the most critical information, then go into the details of your observations.
Clearly distinguish between critical failures and transient (self-healing) issues.
If you identify patterns in the logs that indicate impending failures, highlight them prominently.
Where applicable, explain the causal relationships you recognize.

If you have specific actionable advice for the system administrator, include it in your report.
Be as concrete as possible so the administrator can react quickly. Good example: "Run 'systemctl restart nginx' on www.company.local". Bad example: "Check if the http-server process on the web server is still running".

- Keep it brief. Do not use *any* formatting. Use plain text only. No Markdown, no *bold*, no _italics_, etc.
- The reader of your report might be reading it on a mobile phone, is in a hurry, and needs to grasp the most important information quickly. Keep this in mind when phrasing and structuring your text.
- Avoid quoting lines directly from the logs. Simply state what happened and when. You do not need to provide proof.`

// New initializes a new AIAgent.
func New(client LLMClient, out chkr.Notifier, opts ...AIAgentOption) *AIAgent {
	a := &AIAgent{
		client:       client,
		out:          out,
		buffer:       NewEventBuffer(100), // Default limit: 100 events
		systemPrompt: defaultSystemPrompt,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// WithTrigger sets a custom trigger logic.
func WithTrigger(t Trigger) AIAgentOption {
	return func(a *AIAgent) {
		a.trigger = t
	}
}

// WithSystemPrompt allows overriding the default prompt.
func WithSystemPrompt(p string) AIAgentOption {
	return func(a *AIAgent) {
		a.systemPrompt = p
	}
}

// WithBufferLimit sets the maximum number of events in the buffer.
func WithBufferLimit(limit int) AIAgentOption {
	return func(a *AIAgent) {
		a.buffer = NewEventBuffer(limit)
	}
}

// Notifier returns the chkr.Notifier function to be registered in Checker.
func (a *AIAgent) Notifier() chkr.Notifier {
	return func(ctx context.Context, cs chkr.CheckState) {
		a.mu.Lock()

		a.buffer.Add(cs)

		// Early return if there's no trigger or it shouldn't trigger yet
		if a.trigger == nil || !a.trigger.ShouldTrigger(a.buffer, a.lastAnalysis) {
			a.mu.Unlock()
			return
		}

		// Grab a copy of the current buffer history for analysis
		eventsSeq := a.buffer.Events()
		events := slices.Collect(eventsSeq)
		analysisTime := TimeNow()
		previousAnalysis := a.lastAnalysis

		a.lastAnalysis = analysisTime

		a.mu.Unlock()

		// Start analysis asynchronously to avoid blocking the checker's notifier loop
		go func(evs []Event, last time.Time) {
			analysis, err := a.client.Analyze(ctx, a.systemPrompt, evs, last)
			if err != nil {
				a.out(ctx, chkr.CheckState{
					Name:    "AI-Agent Error",
					State:   chkr.Fail,
					Message: err.Error(),
				})
				return
			}

			a.out(ctx, chkr.CheckState{
				Name:    "AI-Analysis",
				State:   chkr.OK,
				Message: analysis,
			})
		}(events, previousAnalysis)	}
}
