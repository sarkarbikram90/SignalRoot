package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/signalroot/signalroot/internal/incident"
)

// SignalRepo handles signal persistence.
type SignalRepo struct {
	db *DB
}

func NewSignalRepo(db *DB) *SignalRepo {
	return &SignalRepo{db: db}
}

func (r *SignalRepo) Insert(ctx context.Context, s *incident.Signal) error {
	rawPayload, _ := json.Marshal(s.RawPayload)
	query := `INSERT INTO signals (id, org_id, integration_id, source_type, source_id, signal_type, severity, title, body, raw_payload, service_name, environment, tags, occurred_at, received_at, dedupe_key, incident_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) ON CONFLICT (org_id, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, s.ID, s.OrgID, s.IntegrationID, s.SourceType, s.SourceID, s.SignalType, s.Severity, s.Title, s.Body, rawPayload, s.ServiceName, s.Environment, s.Tags, s.OccurredAt, s.ReceivedAt, s.DedupeKey, s.IncidentID)
	return err
}

func (r *SignalRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*incident.Signal, error) {
	query := `SELECT id,org_id,integration_id,source_type,source_id,signal_type,severity,title,body,raw_payload,service_name,environment,tags,occurred_at,received_at,dedupe_key,incident_id FROM signals WHERE id=$1 AND org_id=$2`
	var s incident.Signal
	var rawPayload []byte
	err := r.db.Pool.QueryRow(ctx, query, id, orgID).Scan(&s.ID, &s.OrgID, &s.IntegrationID, &s.SourceType, &s.SourceID, &s.SignalType, &s.Severity, &s.Title, &s.Body, &rawPayload, &s.ServiceName, &s.Environment, &s.Tags, &s.OccurredAt, &s.ReceivedAt, &s.DedupeKey, &s.IncidentID)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(rawPayload, &s.RawPayload)
	return &s, nil
}

func (r *SignalRepo) LinkToIncident(ctx context.Context, orgID, signalID, incidentID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE signals SET incident_id=$1 WHERE id=$2 AND org_id=$3`, incidentID, signalID, orgID)
	return err
}

func (r *SignalRepo) CountRecentByService(ctx context.Context, orgID uuid.UUID, serviceName string, since time.Time) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM signals WHERE org_id=$1 AND service_name=$2 AND occurred_at>=$3`, orgID, serviceName, since).Scan(&count)
	return count, err
}

// IncidentRepo handles incident persistence.
type IncidentRepo struct {
	db *DB
}

func NewIncidentRepo(db *DB) *IncidentRepo {
	return &IncidentRepo{db: db}
}

func (r *IncidentRepo) Create(ctx context.Context, inc *incident.Incident) error {
	metadata, _ := json.Marshal(inc.Metadata)
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var number int64
	err = tx.QueryRow(ctx, "SELECT next_incident_number($1)", inc.OrgID).Scan(&number)
	if err != nil {
		return fmt.Errorf("get incident number: %w", err)
	}
	inc.Number = number

	_, err = tx.Exec(ctx, `INSERT INTO incidents (id,org_id,number,title,summary,severity,status,root_cause,root_cause_confidence,impact_summary,services_affected,environments,tags,detected_at,acknowledged_at,mitigated_at,resolved_at,closed_at,created_at,updated_at,created_by,dna_vector_id,similar_incident_ids,metadata) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		inc.ID, inc.OrgID, inc.Number, inc.Title, inc.Summary, inc.Severity, inc.Status, inc.RootCause, inc.RootCauseConfidence, inc.ImpactSummary, inc.ServicesAffected, inc.Environments, inc.Tags, inc.DetectedAt, inc.AcknowledgedAt, inc.MitigatedAt, inc.ResolvedAt, inc.ClosedAt, inc.CreatedAt, inc.UpdatedAt, inc.CreatedBy, inc.DNAVectorID, inc.SimilarIncidentIDs, metadata)
	if err != nil {
		return fmt.Errorf("insert incident: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *IncidentRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*incident.Incident, error) {
	query := `SELECT i.id,i.org_id,i.number,i.title,i.summary,i.severity,i.status,i.root_cause,i.root_cause_confidence,i.impact_summary,i.services_affected,i.environments,i.tags,i.detected_at,i.acknowledged_at,i.mitigated_at,i.resolved_at,i.closed_at,i.created_at,i.updated_at,i.created_by,i.dna_vector_id,i.similar_incident_ids,i.metadata,COALESCE(sc.cnt,0) FROM incidents i LEFT JOIN (SELECT incident_id,COUNT(*) as cnt FROM signals WHERE incident_id IS NOT NULL GROUP BY incident_id) sc ON sc.incident_id=i.id WHERE i.id=$1 AND i.org_id=$2`
	return r.scanRow(r.db.Pool.QueryRow(ctx, query, id, orgID))
}

func (r *IncidentRepo) scanRow(row pgx.Row) (*incident.Incident, error) {
	var inc incident.Incident
	var metadata []byte
	err := row.Scan(&inc.ID, &inc.OrgID, &inc.Number, &inc.Title, &inc.Summary, &inc.Severity, &inc.Status, &inc.RootCause, &inc.RootCauseConfidence, &inc.ImpactSummary, &inc.ServicesAffected, &inc.Environments, &inc.Tags, &inc.DetectedAt, &inc.AcknowledgedAt, &inc.MitigatedAt, &inc.ResolvedAt, &inc.ClosedAt, &inc.CreatedAt, &inc.UpdatedAt, &inc.CreatedBy, &inc.DNAVectorID, &inc.SimilarIncidentIDs, &metadata, &inc.SignalCount)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadata, &inc.Metadata)
	return &inc, nil
}

func (r *IncidentRepo) Update(ctx context.Context, inc *incident.Incident) error {
	metadata, _ := json.Marshal(inc.Metadata)
	_, err := r.db.Pool.Exec(ctx, `UPDATE incidents SET title=$3,summary=$4,severity=$5,status=$6,root_cause=$7,root_cause_confidence=$8,impact_summary=$9,services_affected=$10,environments=$11,tags=$12,acknowledged_at=$13,mitigated_at=$14,resolved_at=$15,closed_at=$16,updated_at=$17,dna_vector_id=$18,similar_incident_ids=$19,metadata=$20 WHERE id=$1 AND org_id=$2`,
		inc.ID, inc.OrgID, inc.Title, inc.Summary, inc.Severity, inc.Status, inc.RootCause, inc.RootCauseConfidence, inc.ImpactSummary, inc.ServicesAffected, inc.Environments, inc.Tags, inc.AcknowledgedAt, inc.MitigatedAt, inc.ResolvedAt, inc.ClosedAt, time.Now().UTC(), inc.DNAVectorID, inc.SimilarIncidentIDs, metadata)
	return err
}

// IncidentFilter defines filtering options.
type IncidentFilter struct {
	Statuses    []string
	Severities  []string
	Service     string
	Environment string
	From        *time.Time
	To          *time.Time
	HasRCA      *bool
	SearchQuery string
	Limit       int
}

func (r *IncidentRepo) List(ctx context.Context, orgID uuid.UUID, filter IncidentFilter) ([]incident.Incident, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 25
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	var conds []string
	var args []interface{}
	idx := 1
	conds = append(conds, fmt.Sprintf("i.org_id=$%d", idx))
	args = append(args, orgID)
	idx++

	if len(filter.Statuses) > 0 {
		conds = append(conds, fmt.Sprintf("i.status=ANY($%d)", idx))
		args = append(args, filter.Statuses)
		idx++
	}
	if len(filter.Severities) > 0 {
		conds = append(conds, fmt.Sprintf("i.severity=ANY($%d)", idx))
		args = append(args, filter.Severities)
		idx++
	}
	if filter.Service != "" {
		conds = append(conds, fmt.Sprintf("$%d=ANY(i.services_affected)", idx))
		args = append(args, strings.TrimSuffix(filter.Service, "*"))
		idx++
	}
	if filter.From != nil {
		conds = append(conds, fmt.Sprintf("i.detected_at>=$%d", idx))
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		conds = append(conds, fmt.Sprintf("i.detected_at<=$%d", idx))
		args = append(args, *filter.To)
		idx++
	}
	if filter.SearchQuery != "" {
		conds = append(conds, fmt.Sprintf("(i.title ILIKE $%d OR i.summary ILIKE $%d)", idx, idx))
		args = append(args, "%"+filter.SearchQuery+"%")
		idx++
	}

	where := strings.Join(conds, " AND ")
	var total int
	r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM incidents i WHERE "+where, args...).Scan(&total)

	args = append(args, filter.Limit)
	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`SELECT i.id,i.org_id,i.number,i.title,i.summary,i.severity,i.status,i.root_cause,i.root_cause_confidence,i.impact_summary,i.services_affected,i.environments,i.tags,i.detected_at,i.acknowledged_at,i.mitigated_at,i.resolved_at,i.closed_at,i.created_at,i.updated_at,i.created_by,i.dna_vector_id,i.similar_incident_ids,i.metadata,COALESCE(sc.cnt,0) FROM incidents i LEFT JOIN (SELECT incident_id,COUNT(*) as cnt FROM signals WHERE incident_id IS NOT NULL GROUP BY incident_id) sc ON sc.incident_id=i.id WHERE %s ORDER BY i.detected_at DESC LIMIT $%d`, where, idx), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var incidents []incident.Incident
	for rows.Next() {
		inc, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		incidents = append(incidents, *inc)
	}
	return incidents, total, nil
}

func (r *IncidentRepo) FindOpenByService(ctx context.Context, orgID uuid.UUID, serviceName, env string, since time.Time) ([]incident.Incident, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT i.id,i.org_id,i.number,i.title,i.summary,i.severity,i.status,i.root_cause,i.root_cause_confidence,i.impact_summary,i.services_affected,i.environments,i.tags,i.detected_at,i.acknowledged_at,i.mitigated_at,i.resolved_at,i.closed_at,i.created_at,i.updated_at,i.created_by,i.dna_vector_id,i.similar_incident_ids,i.metadata,0 FROM incidents i WHERE i.org_id=$1 AND i.status IN ('open','acknowledged','investigating','mitigated') AND $2=ANY(i.services_affected) AND $3=ANY(i.environments) AND i.detected_at>=$4 ORDER BY i.detected_at DESC`, orgID, serviceName, env, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []incident.Incident
	for rows.Next() {
		inc, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, *inc)
	}
	return res, nil
}

func (r *IncidentRepo) FindRecentlyResolved(ctx context.Context, orgID uuid.UUID, serviceName string, since time.Time) ([]incident.Incident, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT i.id,i.org_id,i.number,i.title,i.summary,i.severity,i.status,i.root_cause,i.root_cause_confidence,i.impact_summary,i.services_affected,i.environments,i.tags,i.detected_at,i.acknowledged_at,i.mitigated_at,i.resolved_at,i.closed_at,i.created_at,i.updated_at,i.created_by,i.dna_vector_id,i.similar_incident_ids,i.metadata,0 FROM incidents i WHERE i.org_id=$1 AND i.status IN ('resolved','closed') AND $2=ANY(i.services_affected) AND COALESCE(i.resolved_at,i.closed_at)>=$3 ORDER BY i.detected_at DESC LIMIT 5`, orgID, serviceName, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []incident.Incident
	for rows.Next() {
		inc, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, *inc)
	}
	return res, nil
}

// TimelineRepo handles timeline events.
type TimelineRepo struct {
	db *DB
}

func NewTimelineRepo(db *DB) *TimelineRepo { return &TimelineRepo{db: db} }

func (r *TimelineRepo) Insert(ctx context.Context, evt *incident.TimelineEvent) error {
	metadata, _ := json.Marshal(evt.Metadata)
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO timeline_events (id,incident_id,signal_id,event_type,actor_type,actor_id,description,metadata,occurred_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, evt.ID, evt.IncidentID, evt.SignalID, evt.EventType, evt.ActorType, evt.ActorID, evt.Description, metadata, evt.OccurredAt, evt.CreatedAt)
	return err
}

func (r *TimelineRepo) ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]incident.TimelineEvent, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT id,incident_id,signal_id,event_type,actor_type,actor_id,description,metadata,occurred_at,created_at FROM timeline_events WHERE incident_id=$1 ORDER BY occurred_at ASC`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []incident.TimelineEvent
	for rows.Next() {
		var e incident.TimelineEvent
		var md []byte
		if err := rows.Scan(&e.ID, &e.IncidentID, &e.SignalID, &e.EventType, &e.ActorType, &e.ActorID, &e.Description, &md, &e.OccurredAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(md, &e.Metadata)
		events = append(events, e)
	}
	return events, nil
}

// JobRepo handles background jobs.
type JobRepo struct {
	db *DB
}

func NewJobRepo(db *DB) *JobRepo { return &JobRepo{db: db} }

func (r *JobRepo) Enqueue(ctx context.Context, job *incident.Job) error {
	payload, _ := json.Marshal(job.Payload)
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO jobs (id,job_type,payload,status,max_attempts,run_at,created_at) VALUES ($1,$2,$3,'pending',$4,$5,$6)`, job.ID, job.JobType, payload, job.MaxAttempts, job.RunAt, job.CreatedAt)
	return err
}

func (r *JobRepo) Dequeue(ctx context.Context) (*incident.Job, error) {
	var job incident.Job
	var payload []byte
	err := r.db.Pool.QueryRow(ctx, `UPDATE jobs SET status='running',started_at=NOW(),attempts=attempts+1 WHERE id=(SELECT id FROM jobs WHERE status='pending' AND run_at<=NOW() ORDER BY run_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING id,job_type,payload,status,attempts,max_attempts,last_error,run_at,started_at,completed_at,created_at`).Scan(&job.ID, &job.JobType, &payload, &job.Status, &job.Attempts, &job.MaxAttempts, &job.LastError, &job.RunAt, &job.StartedAt, &job.CompletedAt, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(payload, &job.Payload)
	return &job, nil
}

func (r *JobRepo) Complete(ctx context.Context, jobID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE jobs SET status='completed',completed_at=NOW() WHERE id=$1`, jobID)
	return err
}

func (r *JobRepo) Fail(ctx context.Context, jobID uuid.UUID, errMsg string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE jobs SET status=CASE WHEN attempts>=max_attempts THEN 'dead' ELSE 'failed' END,last_error=$2,completed_at=NOW() WHERE id=$1`, jobID, errMsg)
	return err
}

func (r *JobRepo) RetryFailed(ctx context.Context) (int, error) {
	result, err := r.db.Pool.Exec(ctx, `UPDATE jobs SET status='pending',run_at=NOW()+INTERVAL '30 seconds' WHERE status='failed' AND attempts<max_attempts`)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

// AuditRepo handles immutable audit log.
type AuditRepo struct {
	db *DB
}

func NewAuditRepo(db *DB) *AuditRepo { return &AuditRepo{db: db} }

func (r *AuditRepo) Log(ctx context.Context, entry *incident.AuditLogEntry) error {
	details, _ := json.Marshal(entry.Details)
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO audit_log (org_id,user_id,action,resource,resource_id,details,ip_address,user_agent) VALUES ($1,$2,$3,$4,$5,$6,$7::inet,$8)`, entry.OrgID, entry.UserID, entry.Action, entry.Resource, entry.ResourceID, details, entry.IPAddress, entry.UserAgent)
	return err
}
