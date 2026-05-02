-- +goose Up
-- Raw signals (immutable append-only log)
CREATE TABLE signals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    integration_id  UUID REFERENCES integrations(id),
    source_type     TEXT NOT NULL,
    source_id       TEXT,
    signal_type     TEXT NOT NULL,
    severity        TEXT,
    title           TEXT NOT NULL,
    body            TEXT,
    raw_payload     JSONB NOT NULL DEFAULT '{}',
    service_name    TEXT,
    environment     TEXT,
    tags            TEXT[] NOT NULL DEFAULT '{}',
    occurred_at     TIMESTAMPTZ NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dedupe_key      TEXT,
    incident_id     UUID
);

-- Partial unique index for deduplication
CREATE UNIQUE INDEX idx_signals_dedupe ON signals(org_id, dedupe_key) WHERE dedupe_key IS NOT NULL;
CREATE INDEX idx_signals_org_occurred ON signals(org_id, occurred_at DESC);
CREATE INDEX idx_signals_incident ON signals(incident_id) WHERE incident_id IS NOT NULL;
CREATE INDEX idx_signals_service ON signals(org_id, service_name, occurred_at DESC);

-- Incidents (core aggregate)
CREATE TABLE incidents (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                UUID NOT NULL,
    number                BIGINT NOT NULL,
    title                 TEXT NOT NULL,
    summary               TEXT,
    severity              TEXT NOT NULL DEFAULT 'unknown' CHECK (severity IN ('critical','high','medium','low','unknown')),
    status                TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','acknowledged','investigating','mitigated','resolved','closed')),
    root_cause            TEXT,
    root_cause_confidence FLOAT,
    impact_summary        TEXT,
    services_affected     TEXT[] NOT NULL DEFAULT '{}',
    environments          TEXT[] NOT NULL DEFAULT '{}',
    tags                  TEXT[] NOT NULL DEFAULT '{}',
    detected_at           TIMESTAMPTZ NOT NULL,
    acknowledged_at       TIMESTAMPTZ,
    mitigated_at          TIMESTAMPTZ,
    resolved_at           TIMESTAMPTZ,
    closed_at             TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by            UUID REFERENCES users(id),
    dna_vector_id         TEXT,
    similar_incident_ids  UUID[] DEFAULT '{}',
    metadata              JSONB NOT NULL DEFAULT '{}',
    UNIQUE(org_id, number)
);

CREATE INDEX idx_incidents_org_detected ON incidents(org_id, detected_at DESC);
CREATE INDEX idx_incidents_status ON incidents(org_id, status);

-- Per-org incident number sequence function
CREATE OR REPLACE FUNCTION next_incident_number(p_org_id UUID) RETURNS BIGINT AS $$
DECLARE
    next_num BIGINT;
BEGIN
    SELECT COALESCE(MAX(number), 0) + 1 INTO next_num
    FROM incidents
    WHERE org_id = p_org_id;
    RETURN next_num;
END;
$$ LANGUAGE plpgsql;

-- Add FK from signals to incidents
ALTER TABLE signals ADD CONSTRAINT fk_signals_incident
    FOREIGN KEY (incident_id) REFERENCES incidents(id);

-- +goose Down
ALTER TABLE signals DROP CONSTRAINT IF EXISTS fk_signals_incident;
DROP FUNCTION IF EXISTS next_incident_number;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS signals;
