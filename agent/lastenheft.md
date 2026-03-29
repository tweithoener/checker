# AI Agent für Checker

Dieses Dokument beschreibt die geplante Funktionalität für einen in Checker integrierten AI-Agenten. Ziel ist es, das System von reiner Fehlererkennung und Benachrichtigung hin zu intelligenter Aggregation und autonomer Fehleranalyse zu erweitern.

## Rahmenbedingungen

- **Optionalität:** Der hier beschriebene AI-Agent ist ein rein optionales Feature. Das Checker-System muss weiterhin vollständig ohne den AI-Agenten nutzbar sein.
- **Minimalinvasiv:** Die Implementierung des AI-Agenten belässt die bestehende API des Checkers im Idealfall völlig unverändert. Sind minimale Anpassungen unumgänglich, werden diese so abstrahiert, dass sie auch für zukünftige Komponenten (außerhalb des AI-Kontextes) nutzbar sind.
- **Konsistenter Stil:** Der Programmier- und Dokumentationsstil entspricht strikt dem restlichen Projekt.
- **Dokumentation:** Die Dokumentation erfolgt im Source-Code, in der Package-Dokumentation (`doc.go`) und in einer dedizierten `README.md`-Datei. Die Anforderungen an die Dokumentation richten sich nach den Best Practices des Agenten-Skills "godoc-polisher".
- **Abhängigkeiten:** Es wird eine minimale Anzahl an externen Abhängigkeiten angestrebt. Soweit möglich und sinnvoll, sollen Sprach-Standards bevorzugt werden, um das Projekt leichtgewichtig zu halten und die Komplexität bei der Integration gering zu halten.
- **Testabdeckung & Mocking:** Es wird eine hohe Testabdeckung durch automatisierte Unit-Tests angestrebt. Geringfügige architektonische Anpassungen im produktiven Code zur Erleichterung der Testbarkeit sind ausdrücklich erwünscht. Insbesondere sind auf Paketebene exportierte oder private Funktionsvariablen für API-Calls (die im Test durch Mock-Funktionen überschrieben werden können) der bevorzugte Weg, um komplizierte Mock-Endpunkte für externe APIs zu vermeiden.

## Basis-Anforderungen

### Architektur & Integration
- **Modell-Agnostik:** Der AI-Agent soll sich an möglichst viele unterschiedliche Modelle verschiedener Anbieter anschließen können (OpenAI, Anthropic, Gemini, lokale Modelle etc.). Technisch wird dies idealerweise über eine standardisierte API (z. B. eine OpenAI-kompatible REST-Schnittstelle) realisiert.
- **Initialisierung:** Der AI-Agent wird minimal mit einem API-Endpunkt und einem API-Token initialisiert.
- **Der Agent als Notifier:** Nach der Initialisierung kann der AI-Agent als regulärer Notifier im Checker registriert werden. Das führt dazu, dass er über alle Zustandsänderungen (`Events`) im System informiert ist.
- **Ereignis-Puffer (Memory):** Der Agent puffert diese Ereignisse in einem eigenen internen Speicher. Um Kontext-Limits der LLMs und Speicherüberläufe zu vermeiden, wird der Puffer sinnvoll limitiert (z. B. als Sliding-Window der letzten X Ereignisse).
- **Ausgabe-Kanal:** Der AI-Agent wird mit mindestens einem weiteren, klassischen Notifier konfiguriert (z. B. Email, Pushover), der ihm dazu dient, seine generierten Erkenntnisse an den Anwender zu senden (Ausgabe-Notifier).

### Trigger-Logik (Wann analysiert die KI?)
- Es gibt eine **Triggerfunktion**, die entscheidet, wann der AI-Agent die gepufferten Ereignisse zusammen mit einem geeigneten Prompt an sein Modell schickt. Die Antwort des Modells wird anschließend über den Ausgabe-Notifier an den Anwender gesendet.
- **Integrierte Standard-Trigger:** Die Triggerfunktion verfügt über sinnvolle Schwellenwerte, zum Beispiel:
  - Zeitbasiert: Letzte Aktivierung liegt länger als 24h zurück.
  - Volumenbasiert: Mehr als X Ereignisse haben sich im Puffer angesammelt (Spike-Erkennung).
  - Zustandsbasiert: Mehr als Y Checks sind länger als Z Minuten im Status `Fail` (Dauerstörung).
- **Custom Triggers:** Die integrierte Triggerfunktion kann durch eine eigene, benutzerdefinierte Triggerfunktion überschrieben werden.

### Prompting & Kontext
- **Standard-System-Prompt:** Der Prompt, der mit den Ereignissen an das Modell geschickt wird, hat einen geeigneten, praxisnahen Standard. Zum Beispiel:
  > *"Analysiere den folgenden Log-Auszug. Unterscheide zwischen transienten Fehlern (selbstheilend) und kritischen Fehlern. Wenn du ein Muster erkennst, das auf einen drohenden Ausfall hindeutet, erkläre die Kausalität."*
- **Custom Prompts:** Der System-Prompt kann vom Anwender bei der Initialisierung überschrieben oder durch domänenspezifisches Wissen ergänzt werden.

---

## Erweiterte Anforderungen

### Chat (Bidirektionale Kommunikation)
- Ergänzend zu den bisherigen unidirektionalen Ausgabe-Notifiern wird das System um bidirektionale Kommunikationswege ("Chats") erweitert, z. B. für Telegram oder Slack.
- Der Anwender kann über diesen Chat direkte Rückfragen zu den Ereignissen und Systemzuständen stellen, über die der AI-Agent durch seinen Puffer Kenntnis hat. Die Antwort erfolgt im selben Kanal.
- **Verbindungsaufbau (Polling):** Um Probleme mit restriktiven Firewalls (Inbound-Traffic) zu vermeiden, wird auf Webhooks verzichtet. Sofern die Chat-Anbieter keine dauerhaften Verbindungen (wie WebSockets) anbieten, nutzt der Agent eine eigene Goroutine für das **Polling**.
- **Adaptives Polling:** Um Ressourcen zu schonen, wird die Polling-Frequenz adaptiv gesteuert (z. B. dramatische Reduzierung der Frequenz, wenn der Anwender im Chat länger inaktiv war).

### Autonome Fehleranalyse (Investigativer AI-Agent)
- Das Checker-Framework verfügt mit den `Cmd()`- und `Ssh()`-Checks bereits über etablierte Wege, auf lokale Systeme und Remote-Peers zuzugreifen. Diese Werkzeuge werden dem AI-Agenten (via LLM "Tool Calling" / "Function Calling") zur Verfügung gestellt.
- **Ziel:** Der Agent kann im Bedarfsfall autonom Systembefehle ausführen, um Fehlerursachen präventiv zu ermitteln, **bevor** er sich das erste Mal beim Anwender meldet. (Beispiel: Die CPU-Load ist hoch -> Der Agent führt autonom `top` aus oder analysiert das Access-Log des Web-Servers, um die Ursache direkt mit in die Alarmierung aufzunehmen).
- **Sicherheitskonzept (Whitelist & Rate-Limiting):** 
  - Um kritische Fehlfunktionen (z. B. versehentliche Löschbefehle durch Halluzinationen) vollständig auszuschließen, operiert der Agent **ohne** "Human-in-the-Loop", wird jedoch durch eine **strikte Whitelist** auf vordefinierte, rein lesende Diagnosebefehle (wie `systemctl status`, `tail`, `cat`, `top`, `journalctl`) limitiert.
  - Die Whitelist muss zukünftig plattformübergreifend (Linux, Windows, macOS) gedacht werden.
  - Ergänzend wird ein **Rate-Limiting** für die Ausführung von Tools implementiert, um zu verhindern, dass die KI in einer Schleife hängenbleibt und das Zielsystem durch ständige Abfragen (z. B. Endlos-SSH-Verbindungen) auslastet.
