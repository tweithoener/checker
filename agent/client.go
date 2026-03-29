package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// doHttpRequest is a package-level variable for the HTTP client's Do method.
// This allows for easy mocking in unit tests without starting a real server.
var doHttpRequest = http.DefaultClient.Do

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type restClient struct {
	endpoint string
	token    string
	model    string
}

// NewRESTClient creates a new LLM client that talks to an OpenAI-compatible REST API.
func NewRESTClient(endpoint, token, model string) LLMClient {
	return &restClient{
		endpoint: endpoint,
		token:    token,
		model:    model,
	}
}

// Analyze sends the events to the LLM for analysis.
func (c *restClient) Analyze(ctx context.Context, systemPrompt string, events []Event, lastAnalysis time.Time) (string, error) {
	if len(events) == 0 {
		return "No events to analyze.", nil
	}

	// Append timeframe instruction if we have analyzed events before
	if !lastAnalysis.IsZero() {
		latestEvent := events[len(events)-1].ReceivedAt
		systemPrompt += fmt.Sprintf("\n\nIMPORTANT: Focus your analysis strictly on the log entries from %s to %s. Entries prior to %s have already been analyzed previously and should only be used as context if they directly influence the recent events.", 
			lastAnalysis.Format("2006-01-02 15:04:05"), 
			latestEvent.Format("2006-01-02 15:04:05"),
			lastAnalysis.Format("2006-01-02 15:04:05"),
		)
	}

	// 1. Format events into a readable string for the LLM
	var b strings.Builder
	b.WriteString("Here is the current log of monitoring events:\n\n")
	for _, ev := range events {
		fmt.Fprintf(&b, "[%s] Check '%s': %s (%s)\n",
			ev.ReceivedAt.Format("2006-01-02 15:04:05"),
			ev.Name,
			ev.CheckState.State,
			ev.CheckState.Message,
		)
	}

	// 2. Prepare the request payload
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: b.String()},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 3. Create the HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// 4. Execute the request using our mockable variable
	resp, err := doHttpRequest(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 5. Parse the response
	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM API returned no choices")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
