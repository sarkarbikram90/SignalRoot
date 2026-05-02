package correlation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/signalroot/signalroot/internal/db"
	"github.com/signalroot/signalroot/internal/incident"
)

// Engine handles signal-to-incident correlation.
type Engine struct {
	signalRepo   *db.SignalRepo
	incidentRepo *db.IncidentRepo
	timelineRepo *db.TimelineRepo
	jobRepo      *db.JobRepo
	logger       *zap.Logger
	windowMins   int // correlation window in minutes
}

// NewEngine creates a new correlation engine.
func NewEngine(sr *db.SignalRepo, ir *db.IncidentRepo, tr *db.TimelineRepo, jr *db.JobRepo, logger *zap.Logger) *Engine {
	return &Engine{
		signalRepo:   sr,
		incidentRepo: ir,
		timelineRepo: tr,
		jobRepo:      jr,
		logger:       logger,
		windowMins:   30,
	}
}

// Correlate determines if a signal belongs to an existing incident or creates a new one.
func (e *Engine) Correlate(ctx context.Context, sig *incident.Signal) (*incident.Incident, error) {
	start := time.Now()
	defer func() {
		e.logger.Debug("correlation completed",
			zap.Duration("duration", time.Since(start)),
			zap.String("signal_id", sig.ID.String()),
		)
	}()

	// 1. Explicit linking — signal already has an incident_id
	if sig.IncidentID != nil {
		inc, err := e.incidentRepo.GetByID(ctx, sig.OrgID, *sig.IncidentID)
		if err == nil {
			return inc, nil
		}
		e.logger.Warn("explicit incident link failed", zap.Error(err))
	}

	serviceName := ""
	if sig.ServiceName != nil {
		serviceName = *sig.ServiceName
	}
	env := "production"
	if sig.Environment != nil {
		env = *sig.Environment
	}

	// 2. Active incident window — find open incidents for same service/env
	if serviceName != "" {
		since := time.Now().UTC().Add(-time.Duration(e.windowMins) * time.Minute)
		openIncidents, err := e.incidentRepo.FindOpenByService(ctx, sig.OrgID, serviceName, env, since)
		if err != nil {
			e.logger.Error("failed to find open incidents", zap.Error(err))
		} else if len(openIncidents) > 0 {
			inc := &openIncidents[0]
			if err := e.linkSignalToIncident(ctx, sig, inc); err != nil {
				return nil, err
			}
			return inc, nil
		}
	}

	// 3. Incident reopening — check recently resolved for same service
	if serviceName != "" {
		twoHoursAgo := time.Now().UTC().Add(-2 * time.Hour)
		recentlyResolved, err := e.incidentRepo.FindRecentlyResolved(ctx, sig.OrgID, serviceName, twoHoursAgo)
		if err == nil && len(recentlyResolved) > 0 {
			severity := ""
			if sig.Severity != nil {
				severity = *sig.Severity
			}
			// Only reopen for non-info signals
			if severity != "info" {
				inc := &recentlyResolved[0]
				return e.reopenIncident(ctx, sig, inc)
			}
		}
	}

	// 4. Info signals don't create new incidents
	if sig.Severity != nil && *sig.Severity == "info" {
		e.logger.Debug("info signal not correlated, skipping incident creation")
		return nil, nil
	}

	// 5. Signal burst detection
	if serviceName != "" {
		twoMinsAgo := time.Now().UTC().Add(-2 * time.Minute)
		count, err := e.signalRepo.CountRecentByService(ctx, sig.OrgID, serviceName, twoMinsAgo)
		if err == nil && count >= 5 {
			e.logger.Info("signal burst detected", zap.String("service", serviceName), zap.Int("count", count))
		}
	}

	// 6. Create new incident
	return e.createIncidentFromSignal(ctx, sig)
}

func (e *Engine) linkSignalToIncident(ctx context.Context, sig *incident.Signal, inc *incident.Incident) error {
	if err := e.signalRepo.LinkToIncident(ctx, sig.OrgID, sig.ID, inc.ID); err != nil {
		return fmt.Errorf("link signal to incident: %w", err)
	}

	// Add timeline event
	now := time.Now().UTC()
	evt := &incident.TimelineEvent{
		ID:          uuid.New(),
		IncidentID:  inc.ID,
		SignalID:    &sig.ID,
		EventType:   "signal",
		ActorType:   strPtr("system"),
		Description: fmt.Sprintf("Signal linked: %s", sig.Title),
		Metadata:    map[string]interface{}{"source_type": sig.SourceType},
		OccurredAt:  sig.OccurredAt,
		CreatedAt:   now,
	}
	e.timelineRepo.Insert(ctx, evt)

	// Update incident severity if signal is more severe
	if sig.Severity != nil && incident.SeverityScore(*sig.Severity) > incident.SeverityScore(inc.Severity) {
		inc.Severity = *sig.Severity
		e.incidentRepo.Update(ctx, inc)
	}

	return nil
}

func (e *Engine) reopenIncident(ctx context.Context, sig *incident.Signal, inc *incident.Incident) (*incident.Incident, error) {
	inc.Status = incident.StatusOpen
	inc.ResolvedAt = nil
	inc.ClosedAt = nil
	if err := e.incidentRepo.Update(ctx, inc); err != nil {
		return nil, fmt.Errorf("reopen incident: %w", err)
	}

	if err := e.linkSignalToIncident(ctx, sig, inc); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	evt := &incident.TimelineEvent{
		ID:          uuid.New(),
		IncidentID:  inc.ID,
		EventType:   "status_change",
		ActorType:   strPtr("system"),
		Description: "Incident reopened after resolution — possible incomplete fix",
		OccurredAt:  now,
		CreatedAt:   now,
	}
	e.timelineRepo.Insert(ctx, evt)

	e.logger.Info("incident reopened", zap.String("incident_id", inc.ID.String()), zap.Int64("number", inc.Number))
	return inc, nil
}

func (e *Engine) createIncidentFromSignal(ctx context.Context, sig *incident.Signal) (*incident.Incident, error) {
	now := time.Now().UTC()
	severity := incident.SeverityUnknown
	if sig.Severity != nil {
		severity = *sig.Severity
	}

	var services []string
	if sig.ServiceName != nil {
		services = []string{*sig.ServiceName}
	}
	var envs []string
	if sig.Environment != nil {
		envs = []string{*sig.Environment}
	}

	inc := &incident.Incident{
		ID:               uuid.New(),
		OrgID:            sig.OrgID,
		Title:            sig.Title,
		Severity:         severity,
		Status:           incident.StatusOpen,
		ServicesAffected: services,
		Environments:     envs,
		Tags:             sig.Tags,
		DetectedAt:       sig.OccurredAt,
		CreatedAt:        now,
		UpdatedAt:        now,
		Metadata:         make(map[string]interface{}),
	}

	if err := e.incidentRepo.Create(ctx, inc); err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}

	// Link the originating signal
	if err := e.signalRepo.LinkToIncident(ctx, sig.OrgID, sig.ID, inc.ID); err != nil {
		e.logger.Error("failed to link signal to new incident", zap.Error(err))
	}

	// Add timeline events
	evt := &incident.TimelineEvent{
		ID:          uuid.New(),
		IncidentID:  inc.ID,
		SignalID:    &sig.ID,
		EventType:   "signal",
		ActorType:   strPtr("system"),
		Description: fmt.Sprintf("Incident created from signal: %s", sig.Title),
		Metadata:    map[string]interface{}{"source_type": sig.SourceType},
		OccurredAt:  sig.OccurredAt,
		CreatedAt:   now,
	}
	e.timelineRepo.Insert(ctx, evt)

	// Enqueue async jobs
	e.enqueueJob(ctx, "generate_incident_summary", map[string]interface{}{
		"incident_id": inc.ID.String(),
		"org_id":      inc.OrgID.String(),
	}, 3)

	e.enqueueJob(ctx, "compute_incident_dna", map[string]interface{}{
		"incident_id": inc.ID.String(),
		"org_id":      inc.OrgID.String(),
	}, 3)

	e.logger.Info("incident created",
		zap.String("incident_id", inc.ID.String()),
		zap.Int64("number", inc.Number),
		zap.String("severity", inc.Severity),
	)

	return inc, nil
}

func (e *Engine) enqueueJob(ctx context.Context, jobType string, payload map[string]interface{}, maxAttempts int) {
	now := time.Now().UTC()
	job := &incident.Job{
		ID:          uuid.New(),
		JobType:     jobType,
		Payload:     payload,
		Status:      "pending",
		MaxAttempts: maxAttempts,
		RunAt:       now,
		CreatedAt:   now,
	}
	if err := e.jobRepo.Enqueue(ctx, job); err != nil {
		e.logger.Error("failed to enqueue job", zap.String("type", jobType), zap.Error(err))
	}
}

func strPtr(s string) *string {
	return &s
}
