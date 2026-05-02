package ingestion

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/signalroot/signalroot/internal/incident"
)

// CanonicalSignal is the normalized signal produced by all adapters.
type CanonicalSignal struct {
	SourceType  string
	SourceID    string
	SignalType  string
	Severity    string
	Title       string
	Body        string
	ServiceName string
	Environment string
	Tags        []string
	RawPayload  map[string]interface{}
	OccurredAt  time.Time
	DedupeKey   string
}

// Adapter normalizes external payloads into canonical signals.
type Adapter interface {
	Normalize(payload []byte) ([]CanonicalSignal, error)
	VerifySignature(payload []byte, signature string, secret string) bool
	SourceType() string
}

// ToSignal converts a canonical signal to a persistent Signal model.
func (cs *CanonicalSignal) ToSignal(orgID uuid.UUID, integrationID *uuid.UUID) *incident.Signal {
	title := cs.Title
	if title == "" && cs.Body != "" {
		if len(cs.Body) > 200 {
			title = cs.Body[:200]
		} else {
			title = cs.Body
		}
	}
	if title == "" {
		title = fmt.Sprintf("Untitled signal from %s", cs.SourceType)
	}
	if len(title) > 500 {
		title = title[:500]
	}

	env := cs.Environment
	if env == "" {
		env = "production"
	}

	sev := cs.Severity
	var sevPtr *string
	if sev != "" {
		sevPtr = &sev
	}
	var body *string
	if cs.Body != "" {
		body = &cs.Body
	}
	var srcID *string
	if cs.SourceID != "" {
		srcID = &cs.SourceID
	}
	var svcName *string
	if cs.ServiceName != "" {
		svcName = &cs.ServiceName
	}
	var dedupeKey *string
	if cs.DedupeKey != "" {
		dedupeKey = &cs.DedupeKey
	}

	return &incident.Signal{
		ID:            uuid.New(),
		OrgID:         orgID,
		IntegrationID: integrationID,
		SourceType:    cs.SourceType,
		SourceID:      srcID,
		SignalType:    cs.SignalType,
		Severity:      sevPtr,
		Title:         title,
		Body:          body,
		RawPayload:    cs.RawPayload,
		ServiceName:   svcName,
		Environment:   &env,
		Tags:          cs.Tags,
		OccurredAt:    cs.OccurredAt,
		ReceivedAt:    time.Now().UTC(),
		DedupeKey:     dedupeKey,
	}
}

// computeDedupeKey generates a deterministic SHA-256 hash for deduplication.
func computeDedupeKey(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, ":")))
	return hex.EncodeToString(h[:])
}

// verifyHMACSHA256 verifies an HMAC-SHA256 signature.
func verifyHMACSHA256(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// PagerDutyAdapter handles PagerDuty webhook payloads.
type PagerDutyAdapter struct{}

func (a *PagerDutyAdapter) SourceType() string { return "pagerduty" }

func (a *PagerDutyAdapter) VerifySignature(payload []byte, signature, secret string) bool {
	return verifyHMACSHA256(payload, strings.TrimPrefix(signature, "v1="), secret)
}

func (a *PagerDutyAdapter) Normalize(payload []byte) ([]CanonicalSignal, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	messages, ok := raw["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing messages array")
	}

	var signals []CanonicalSignal
	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		eventType, _ := m["event"].(string)
		incData, _ := m["incident"].(map[string]interface{})
		if incData == nil {
			continue
		}

		title, _ := incData["title"].(string)
		desc, _ := incData["description"].(string)
		incID, _ := incData["id"].(string)
		urgency, _ := incData["urgency"].(string)

		var serviceName string
		if svc, ok := incData["service"].(map[string]interface{}); ok {
			serviceName, _ = svc["name"].(string)
		}

		severity := mapPDSeverity(urgency)
		signalType := mapPDEventType(eventType)

		var occurredAt time.Time
		if ts, ok := incData["created_at"].(string); ok {
			occurredAt, _ = time.Parse(time.RFC3339, ts)
		}
		if occurredAt.IsZero() {
			occurredAt = time.Now().UTC()
		}

		signals = append(signals, CanonicalSignal{
			SourceType:  "pagerduty",
			SourceID:    incID,
			SignalType:  signalType,
			Severity:    severity,
			Title:       title,
			Body:        desc,
			ServiceName: serviceName,
			Tags:        []string{fmt.Sprintf("event=%s", eventType)},
			RawPayload:  m,
			OccurredAt:  occurredAt,
			DedupeKey:   computeDedupeKey("pagerduty", incID, eventType),
		})
	}
	return signals, nil
}

func mapPDSeverity(urgency string) string {
	switch urgency {
	case "high":
		return "critical"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

func mapPDEventType(event string) string {
	switch event {
	case "incident.triggered":
		return "alert"
	case "incident.acknowledged":
		return "status_change"
	case "incident.resolved":
		return "status_change"
	case "incident.annotated":
		return "comment"
	default:
		return "alert"
	}
}

// SlackAdapter handles Slack Events API payloads.
type SlackAdapter struct{}

func (a *SlackAdapter) SourceType() string { return "slack" }

func (a *SlackAdapter) VerifySignature(payload []byte, signature, secret string) bool {
	// Slack v0 signing: v0=sha256(v0:timestamp:body)
	// For simplicity, we verify the hash portion
	return verifyHMACSHA256(payload, strings.TrimPrefix(signature, "v0="), secret)
}

func (a *SlackAdapter) Normalize(payload []byte) ([]CanonicalSignal, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	event, ok := raw["event"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing event object")
	}

	eventType, _ := event["type"].(string)
	text, _ := event["text"].(string)
	channel, _ := event["channel"].(string)
	ts, _ := event["ts"].(string)
	user, _ := event["user"].(string)

	signalType := "comment"
	if eventType == "channel_topic" {
		signalType = "status_change"
	}

	title := text
	if len(title) > 200 {
		title = title[:200]
	}
	if title == "" {
		title = fmt.Sprintf("Slack message in %s", channel)
	}

	var occurredAt time.Time
	if ts != "" {
		// Slack timestamps are Unix epoch with decimal
		parts := strings.Split(ts, ".")
		if len(parts) > 0 {
			if sec, err := fmt.Sscanf(parts[0], "%d", new(int64)); err == nil && sec > 0 {
				var tsInt int64
				fmt.Sscanf(parts[0], "%d", &tsInt)
				occurredAt = time.Unix(tsInt, 0).UTC()
			}
		}
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	return []CanonicalSignal{{
		SourceType:  "slack",
		SourceID:    ts,
		SignalType:  signalType,
		Severity:    "info",
		Title:       title,
		Body:        text,
		Tags:        []string{fmt.Sprintf("channel=%s", channel), fmt.Sprintf("user=%s", user)},
		RawPayload:  raw,
		OccurredAt:  occurredAt,
		DedupeKey:   computeDedupeKey("slack", channel, ts),
	}}, nil
}

// GenericWebhookAdapter handles arbitrary JSON payloads.
type GenericWebhookAdapter struct{}

func (a *GenericWebhookAdapter) SourceType() string { return "webhook" }

func (a *GenericWebhookAdapter) VerifySignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true
	}
	return verifyHMACSHA256(payload, signature, secret)
}

func (a *GenericWebhookAdapter) Normalize(payload []byte) ([]CanonicalSignal, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	title, _ := raw["title"].(string)
	if title == "" {
		title, _ = raw["summary"].(string)
	}
	body, _ := raw["body"].(string)
	if body == "" {
		body, _ = raw["message"].(string)
	}
	severity, _ := raw["severity"].(string)
	service, _ := raw["service"].(string)

	return []CanonicalSignal{{
		SourceType:  "webhook",
		SignalType:  "alert",
		Severity:    severity,
		Title:       title,
		Body:        body,
		ServiceName: service,
		RawPayload:  raw,
		OccurredAt:  time.Now().UTC(),
		DedupeKey:   computeDedupeKey("webhook", fmt.Sprintf("%x", sha256.Sum256(payload))),
	}}, nil
}

// GetAdapter returns the appropriate adapter for a source type.
func GetAdapter(sourceType string) Adapter {
	switch sourceType {
	case "pagerduty":
		return &PagerDutyAdapter{}
	case "slack":
		return &SlackAdapter{}
	default:
		return &GenericWebhookAdapter{}
	}
}
