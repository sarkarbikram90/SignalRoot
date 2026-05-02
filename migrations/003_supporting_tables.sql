-- +goose Up
-- Timeline events (ordered causal chain)
CREATE TABLE timeline_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id     UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    signal_id       UUID REFERENCES signals(id),
    event_type      TEXT NOT NULL,
    actor_type      TEXT,
    actor_id        UUID,
    description     TEXT NOT NULL,
    metadata        JSONB NOT NULL DEFAULT '{}',
    occurred_at     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_timeline_incident ON timeline_events(incident_id, occurred_at ASC);

-- Incident responders
CREATE TABLE incident_responders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id     UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),
    role            TEXT NOT NULL DEFAULT 'responder' CHECK (role IN ('commander','lead','responder','observer')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at         TIMESTAMPTZ,
    UNIQUE(incident_id, user_id)
);

-- Compliance reports
CREATE TABLE compliance_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    incident_id     UUID NOT NULL REFERENCES incidents(id),
    report_type     TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','generated','reviewed','finalized')),
    generated_by    UUID REFERENCES users(id),
    content         JSONB NOT NULL DEFAULT '{}',
    pdf_url         TEXT,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized_at    TIMESTAMPTZ
);

-- Incident DNA (similarity metadata)
CREATE TABLE incident_dna (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id     UUID NOT NULL REFERENCES incidents(id) UNIQUE,
    org_id          UUID NOT NULL,
    feature_vector  JSONB NOT NULL,
    top_matches     JSONB NOT NULL DEFAULT '[]',
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Knowledge entities
CREATE TABLE knowledge_entities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    entity_type     TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    incident_count  INT NOT NULL DEFAULT 0,
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, entity_type, name)
);

-- Audit log (immutable)
CREATE TABLE audit_log (
    id          BIGSERIAL PRIMARY KEY,
    org_id      UUID NOT NULL,
    user_id     UUID,
    action      TEXT NOT NULL,
    resource    TEXT NOT NULL,
    resource_id TEXT,
    details     JSONB NOT NULL DEFAULT '{}',
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_org_created ON audit_log(org_id, created_at DESC);

-- Notification rules
CREATE TABLE notification_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    severities      TEXT[] NOT NULL DEFAULT '{}',
    services        TEXT[] NOT NULL DEFAULT '{}',
    channels        TEXT[] NOT NULL DEFAULT '{}',
    notify_users    UUID[] NOT NULL DEFAULT '{}',
    cooldown_minutes INT NOT NULL DEFAULT 5,
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Background jobs table (pgqueue pattern)
CREATE TABLE jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type        TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','completed','failed','dead')),
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    last_error      TEXT,
    run_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_jobs_pending ON jobs(status, run_at ASC) WHERE status = 'pending';
CREATE INDEX idx_jobs_type ON jobs(job_type, status);

-- LLM usage tracking
CREATE TABLE llm_usage (
    id              BIGSERIAL PRIMARY KEY,
    org_id          UUID NOT NULL,
    model           TEXT NOT NULL,
    job_type        TEXT NOT NULL,
    input_tokens    INT NOT NULL DEFAULT 0,
    output_tokens   INT NOT NULL DEFAULT 0,
    cost_cents      FLOAT NOT NULL DEFAULT 0,
    cached          BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_llm_usage_org ON llm_usage(org_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS llm_usage;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS notification_rules;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS knowledge_entities;
DROP TABLE IF EXISTS incident_dna;
DROP TABLE IF EXISTS compliance_reports;
DROP TABLE IF EXISTS incident_responders;
DROP TABLE IF EXISTS timeline_events;
