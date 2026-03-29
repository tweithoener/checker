# Pflichtenheft: AI-Agent für Checker

Dieses Pflichtenheft übersetzt die Anforderungen des `lastenheft.md` in eine konkrete technische Architektur und Implementierungsstrategie. Es dient als Bauplan und Diskussionsgrundlage für die Umsetzung im neuen Package `agent`.

---

## Phase 1: Die Kern-Engine (Puffer, Trigger, LLM)

In der ersten Phase bauen wir den unidirektionalen Agenten, der Events sammelt, bewertet und per klassischem Notifier alarmiert.

### 1. Kernkomponente: `AIAgent` und Konfiguration
Der Agent ist das zentrale Bindeglied. Er hält den Zustand, orchestriert die Schnittstellen und stellt das `chkr.Notifier`-Interface bereit. Die Initialisierung erfolgt flexibel über das "Functional Options"-Pattern.

```go
package agent

import (
	"context"
	"sync"
	chkr "github.com/tweithoener/checker"
)

// AIAgent repräsentiert den KI-Assistenten.
type AIAgent struct {
	mu           sync.Mutex
	client       LLMClient
	buffer       *EventBuffer
	trigger      Trigger
	out          chkr.Notifier
	systemPrompt string
}

// AIAgentOption definiert funktionale Optionen für den Agenten (z.B. WithTrigger, WithPrompt).
type AIAgentOption func(*AIAgent)

// New initialisiert einen neuen Agenten.
func New(endpoint, token string, out chkr.Notifier, opts ...AIAgentOption) *AIAgent

// Notifier gibt die chkr.Notifier-Funktion zurück, die in Checker registriert wird.
func (a *AIAgent) Notifier() chkr.Notifier
```

### 2. Ereignis-Puffer (`EventBuffer`) und Datenmodell
Der Agent darf das LLM nicht mit jedem einzelnen Event überfluten. Events werden in einem threadsicheren Ringpuffer / Sliding-Window gespeichert. Um die zeitbasierte Auswertung zu vereinfachen, kapseln wir den `CheckState` in einem eigenen Typen.

```go
// Event kapselt einen CheckState mit Metadaten für den Agenten.
type Event struct {
	Name       string
	CheckState chkr.CheckState
	ReceivedAt time.Time // Präziser Zeitstempel beim Eintreffen im Agenten
}

// EventBuffer speichert eine begrenzte Anzahl an Events.
type EventBuffer struct {
	mu     sync.Mutex
	events []Event
	limit  int 
}

// Add fügt ein neues Event hinzu.
func (b *EventBuffer) Add(name string, cs chkr.CheckState)

// Flush gibt alle gepufferten Events zurück und leert den Puffer.
func (b *EventBuffer) Flush() []Event
```

### 3. Trigger-System (`Trigger`)
Die Trigger-Logik entscheidet, *wann* der Puffer an die KI geschickt wird. Durch ein Interface ist dies beliebig erweiterbar.

```go
// Trigger evaluiert, ob eine Analyse durch die KI gestartet werden soll.
type Trigger interface {
	ShouldTrigger(buffer *EventBuffer) bool
}
```
**Geplante Standard-Trigger (Implementierungen):**
- `3.1 VolumeTrigger`: Löst aus, wenn `len(events) >= threshold`.
- `3.2 TimeTrigger`: Löst aus, wenn seit dem letzten Flush `interval` vergangen ist.
- `3.3 StateTrigger`: Löst aus, wenn eine gewisse Anzahl an Checks den Status `Fail` hat.
- `3.4 AndTrigger`: Zieht nur, wenn *alle* enthaltenen Trigger wahr sind.
- `3.5 OrTrigger`: Zieht, wenn *mindestens einer* der enthaltenen Trigger wahr ist.

### 4. LLM-Client (`LLMClient`)
Um externe Abhängigkeiten zu minimieren, wird ein schlanker, generischer HTTP-Client implementiert, der das OpenAI-REST-Format (das auch von lokalen Modellen wie Ollama und anderen Anbietern verstanden wird) via `net/http` spricht.

```go
// LLMClient abstrahiert die Kommunikation mit dem KI-Modell.
type LLMClient interface {
	Analyze(ctx context.Context, systemPrompt string, events []Event) (string, error)
}
```

### 4.1 Mocking-Strategie für Unit Tests
Diese Variable auf Paketebene erlaubt das Testen der API-Logik ohne echten Netzwerk-Traffic.
```go
var doHttpRequest = http.DefaultClient.Do
```

### 4.2 Integrationstest (Abschluss Phase 1)
Bevor mit Phase 2 begonnen wird, muss ein lauffähiges Gesamtsystem (Agent -> Puffer -> Trigger -> LLM -> Notifier) existieren. 
- **Validierung:** Ein simuliertes Fehlerszenario (Trigger löst aus) muss erfolgreich eine KI-Analyse durchlaufen und das Ergebnis über den Ausgabe-Notifier zustellen.
- **Voraussetzung:** Erfolgreicher Durchlauf stellt den Meilenstein für die Erweiterung um interaktive Komponenten dar.

---

## Phase 2: Chat & Bidirektionale Kommunikation

Sobald Phase 1 stabil ist, wird das System um die Chat-Fähigkeit erweitert.

### 5. Chat-Provider (`ChatProvider`)
Der Agent muss in der Lage sein, Nachrichten aus einem Chat-System zu lesen (Polling) und dorthin zu antworten.

```go
// ChatMessage repräsentiert eine eingehende oder ausgehende Nachricht.
type ChatMessage struct {
	ChannelID string
	UserID    string
	Text      string
}

// ChatProvider ist die Schnittstelle zu Telegram, Slack etc.
type ChatProvider interface {
	// Poll blockiert und liefert neue Nachrichten (mit adaptivem Backoff intern).
	Poll(ctx context.Context) (<-chan ChatMessage, error)
	// Send schickt eine Nachricht in den Kanal zurück.
	Send(ctx context.Context, msg ChatMessage) error
}
```

### 6. Chat-Loop im AIAgent
Der `AIAgent` erhält eine Methode `StartChat(ctx context.Context, provider ChatProvider)`.
- Diese startet eine Goroutine, die den `Poll()`-Channel liest.
- Eingehende Fragen (z.B. "Wie ist der Status von Datenbank X?") werden an den `LLMClient` geschickt, *inklusive* des aktuellen Inhalts des `EventBuffer` als Kontext.
- Die Antwort wird via `Send()` zurückgespielt.

---

## Phase 3: Autonome Fehleranalyse (Investigativer Agent)

In der finalen Phase erhält die KI Werkzeuge ("Function Calling"), um Ursachen selbstständig zu ermitteln.

### 7. Tool-Registry und Whitelist
Um RCE-Risiken (Remote Code Execution) zu minimieren, wird eine harte Whitelist implementiert. Die KI darf nicht beliebige Strings in `Cmd()` werfen.

```go
// Tool repräsentiert einen ausführbaren, lesenden Befehl.
type Tool struct {
	Name        string
	Description string // Wichtig für den Prompt der KI, damit sie weiß, was das Tool tut
	Execute     func(ctx context.Context, args map[string]string) (string, error)
}

// 7.1 Standard-Whitelist-Tools
var StandardTools = []Tool{
	{
		Name: "GetProcessStatus",
		Description: "Returns the status of a systemd service.",
		Execute: func(...) { /* calls systemctl status <arg> */ },
	},
	{
		Name: "ReadLogTail",
		Description: "Reads the last N lines of a specified log file.",
		Execute: func(...) { /* calls tail -n <arg> */ },
	},
}
```

### 8. Rate-Limiting & Execution Guard
- **Token-Bucket:** Dem `AIAgent` wird ein Rate-Limiter hinzugefügt (z. B. `golang.org/x/time/rate`), der auf max. 5 Tool-Calls pro Minute limitiert ist.
- **LLM Function Calling:** Der `restClient` (aus 4.) wird erweitert, um dem LLM das JSON-Schema der verfügbaren `StandardTools` im API-Request mitzuteilen, sodass das Modell bei Analysebedarf eine Funktion anfordern kann.
