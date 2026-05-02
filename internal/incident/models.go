package incident

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a tenant in the system.
type Organization struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Slug      string                 `json:"slug" db:"slug"`
	Plan      string                 `json:"plan" db:"plan"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// User represents an authenticated user.
type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	OrgID       uuid.UUID  `json:"org_id" db:"org_id"`
	Email       string     `json:"email" db:"email"`
	Name        string     `json:"name" db:"name"`
	Role        string     `json:"role" db:"role"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Provider    string     `json:"provider" db:"provider"`
	ProviderID  *string    `json:"-" db:"provider_id"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// APIKey represents an API key for programmatic access.
type APIKey struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	OrgID      uuid.UUID  `json:"org_id" db:"org_id"`
	UserID     *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Name       string     `json:"name" db:"name"`
	KeyHash    string     `json:"-" db:"key_hash"`
	KeyPrefix  string     `json:"key_prefix" db:"key_prefix"`
	Scopes     []string   `json:"scopes" db:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// Integration represents a connected external tool.
type Integration struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	OrgID        uuid.UUID              `json:"org_id" db:"org_id"`
	Type         string                 `json:"type" db:"type"`
	Name         string                 `json:"name" db:"name"`
	Config       map[string]interface{} `json:"config" db:"config"`
	Status       string                 `json:"status" db:"status"`
	LastSyncedAt *time.Time             `json:"last_synced_at,omitempty" db:"last_synced_at"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Signal represents a raw event ingested from an external source.
type Signal struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	OrgID         uuid.UUID              `json:"org_id" db:"org_id"`
	IntegrationID *uuid.UUID             `json:"integration_id,omitempty" db:"integration_id"`
	SourceType    string                 `json:"source_type" db:"source_type"`
	SourceID      *string                `json:"source_id,omitempty" db:"source_id"`
	SignalType    string                 `json:"signal_type" db:"signal_type"`
	Severity      *string                `json:"severity,omitempty" db:"severity"`
	Title         string                 `json:"title" db:"title"`
	Body          *string                `json:"body,omitempty" db:"body"`
	RawPayload    map[string]interface{} `json:"raw_payload" db:"raw_payload"`
	ServiceName   *string                `json:"service_name,omitempty" db:"service_name"`
	Environment   *string                `json:"environment,omitempty" db:"environment"`
	Tags          []string               `json:"tags" db:"tags"`
	OccurredAt    time.Time              `json:"occurred_at" db:"occurred_at"`
	ReceivedAt    time.Time              `json:"received_at" db:"received_at"`
	DedupeKey     *string                `json:"dedupe_key,omitempty" db:"dedupe_key"`
	IncidentID    *uuid.UUID             `json:"incident_id,omitempty" db:"incident_id"`
}

// Incident represents a correlated group of signals.
type Incident struct {
	ID                  uuid.UUID   `json:"id" db:"id"`
	OrgID               uuid.UUID   `json:"org_id" db:"org_id"`
	Number              int64       `json:"number" db:"number"`
	Title               string      `json:"title" db:"title"`
	Summary             *string     `json:"summary,omitempty" db:"summary"`
	Severity            string      `json:"severity" db:"severity"`
	Status              string      `json:"status" db:"status"`
	RootCause           *string     `json:"root_cause,omitempty" db:"root_cause"`
	RootCauseConfidence *float64    `json:"root_cause_confidence,omitempty" db:"root_cause_confidence"`
	ImpactSummary       *string     `json:"impact_summary,omitempty" db:"impact_summary"`
	ServicesAffected    []string    `json:"services_affected" db:"services_affected"`
	Environments        []string    `json:"environments" db:"environments"`
	Tags                []string    `json:"tags" db:"tags"`
	DetectedAt          time.Time   `json:"detected_at" db:"detected_at"`
	AcknowledgedAt      *time.Time  `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	MitigatedAt         *time.Time  `json:"mitigated_at,omitempty" db:"mitigated_at"`
	ResolvedAt          *time.Time  `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt            *time.Time  `json:"closed_at,omitempty" db:"closed_at"`
	CreatedAt           time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy           *uuid.UUID  `json:"created_by,omitempty" db:"created_by"`
	DNAVectorID         *string     `json:"dna_vector_id,omitempty" db:"dna_vector_id"`
	SimilarIncidentIDs  []uuid.UUID `json:"similar_incident_ids,omitempty" db:"similar_incident_ids"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`

	// Computed fields (not stored, populated by queries)
	SignalCount int `json:"signal_count,omitempty" db:"signal_count"`
}

// TimelineEvent represents a single event in an incident timeline.
type TimelineEvent struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	IncidentID  uuid.UUID              `json:"incident_id" db:"incident_id"`
	SignalID    *uuid.UUID             `json:"signal_id,omitempty" db:"signal_id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	ActorType   *string                `json:"actor_type,omitempty" db:"actor_type"`
	ActorID     *uuid.UUID             `json:"actor_id,omitempty" db:"actor_id"`
	Description string                 `json:"description" db:"description"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	OccurredAt  time.Time              `json:"occurred_at" db:"occurred_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// IncidentResponder represents a user assigned to an incident.
type IncidentResponder struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	IncidentID uuid.UUID  `json:"incident_id" db:"incident_id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	Role       string     `json:"role" db:"role"`
	JoinedAt   time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt     *time.Time `json:"left_at,omitempty" db:"left_at"`

	// Joined fields
	UserName  string `json:"user_name,omitempty" db:"user_name"`
	UserEmail string `json:"user_email,omitempty" db:"user_email"`
}

// ComplianceReport represents an audit-ready incident report.
type ComplianceReport struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	OrgID        uuid.UUID              `json:"org_id" db:"org_id"`
	IncidentID   uuid.UUID              `json:"incident_id" db:"incident_id"`
	ReportType   string                 `json:"report_type" db:"report_type"`
	Status       string                 `json:"status" db:"status"`
	GeneratedBy  *uuid.UUID             `json:"generated_by,omitempty" db:"generated_by"`
	Content      map[string]interface{} `json:"content" db:"content"`
	PDFURL       *string                `json:"pdf_url,omitempty" db:"pdf_url"`
	GeneratedAt  time.Time              `json:"generated_at" db:"generated_at"`
	FinalizedAt  *time.Time             `json:"finalized_at,omitempty" db:"finalized_at"`
}

// IncidentDNA stores the similarity feature vector for an incident.
type IncidentDNA struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	IncidentID    uuid.UUID              `json:"incident_id" db:"incident_id"`
	OrgID         uuid.UUID              `json:"org_id" db:"org_id"`
	FeatureVector map[string]interface{} `json:"feature_vector" db:"feature_vector"`
	TopMatches    []MatchResult          `json:"top_matches" db:"top_matches"`
	ComputedAt    time.Time              `json:"computed_at" db:"computed_at"`
}

// MatchResult represents a similarity match with another incident.
type MatchResult struct {
	IncidentID  uuid.UUID `json:"incident_id"`
	Score       float64   `json:"score"`
	Reason      string    `json:"reason"`
	MatchedAt   time.Time `json:"matched_at"`
}

// KnowledgeEntity represents a known entity in the org's knowledge graph.
type KnowledgeEntity struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	OrgID         uuid.UUID              `json:"org_id" db:"org_id"`
	EntityType    string                 `json:"entity_type" db:"entity_type"`
	Name          string                 `json:"name" db:"name"`
	Description   *string                `json:"description,omitempty" db:"description"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	IncidentCount int                    `json:"incident_count" db:"incident_count"`
	LastSeenAt    *time.Time             `json:"last_seen_at,omitempty" db:"last_seen_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// AuditLogEntry represents an immutable audit trail entry.
type AuditLogEntry struct {
	ID         int64                  `json:"id" db:"id"`
	OrgID      uuid.UUID              `json:"org_id" db:"org_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	Action     string                 `json:"action" db:"action"`
	Resource   string                 `json:"resource" db:"resource"`
	ResourceID *string                `json:"resource_id,omitempty" db:"resource_id"`
	Details    map[string]interface{} `json:"details" db:"details"`
	IPAddress  *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// Job represents a background processing job.
type Job struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	JobType     string                 `json:"job_type" db:"job_type"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Status      string                 `json:"status" db:"status"`
	Attempts    int                    `json:"attempts" db:"attempts"`
	MaxAttempts int                    `json:"max_attempts" db:"max_attempts"`
	LastError   *string                `json:"last_error,omitempty" db:"last_error"`
	RunAt       time.Time              `json:"run_at" db:"run_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// NotificationRule defines when and where to send notifications.
type NotificationRule struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	OrgID           uuid.UUID   `json:"org_id" db:"org_id"`
	Name            string      `json:"name" db:"name"`
	Severities      []string    `json:"severities" db:"severities"`
	Services        []string    `json:"services" db:"services"`
	Channels        []string    `json:"channels" db:"channels"`
	NotifyUsers     []uuid.UUID `json:"notify_users" db:"notify_users"`
	CooldownMinutes int         `json:"cooldown_minutes" db:"cooldown_minutes"`
	Enabled         bool        `json:"enabled" db:"enabled"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
}

// Pagination helpers
type PaginationMeta struct {
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"has_more"`
	Total   int    `json:"total,omitempty"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// Status constants
const (
	StatusOpen          = "open"
	StatusAcknowledged  = "acknowledged"
	StatusInvestigating = "investigating"
	StatusMitigated     = "mitigated"
	StatusResolved      = "resolved"
	StatusClosed        = "closed"
)

// Severity constants
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
	SeverityUnknown  = "unknown"
)

// Role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// ValidStatusTransitions maps the allowed state machine transitions.
var ValidStatusTransitions = map[string][]string{
	StatusOpen:          {StatusAcknowledged, StatusInvestigating, StatusMitigated, StatusResolved, StatusClosed},
	StatusAcknowledged:  {StatusInvestigating, StatusMitigated, StatusResolved, StatusClosed},
	StatusInvestigating: {StatusMitigated, StatusResolved, StatusClosed},
	StatusMitigated:     {StatusResolved, StatusClosed},
	StatusResolved:      {StatusClosed, StatusOpen}, // reopen
	StatusClosed:        {StatusOpen},                // reopen
}

// IsValidTransition checks if the status transition is allowed.
func IsValidTransition(from, to string) bool {
	allowed, ok := ValidStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// SeverityScore returns a numeric score for sorting.
func SeverityScore(s string) float64 {
	switch s {
	case SeverityCritical:
		return 1.0
	case SeverityHigh:
		return 0.75
	case SeverityMedium:
		return 0.5
	case SeverityLow:
		return 0.25
	case SeverityInfo:
		return 0.1
	default:
		return 0.0
	}
}

// MaxSeverity returns the higher severity between two.
func MaxSeverity(a, b string) string {
	if SeverityScore(a) >= SeverityScore(b) {
		return a
	}
	return b
}
