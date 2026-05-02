export interface Organization {
  id: string;
  name: string;
  slug: string;
  plan: string;
  settings: Record<string, unknown>;
  created_at: string;
}

export interface User {
  id: string;
  org_id: string;
  email: string;
  name: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
  avatar_url?: string;
}

export interface Incident {
  id: string;
  org_id: string;
  number: number;
  title: string;
  summary?: string;
  severity: Severity;
  status: IncidentStatus;
  root_cause?: string;
  root_cause_confidence?: number;
  impact_summary?: string;
  services_affected: string[];
  environments: string[];
  tags: string[];
  detected_at: string;
  acknowledged_at?: string;
  mitigated_at?: string;
  resolved_at?: string;
  closed_at?: string;
  created_at: string;
  updated_at: string;
  signal_count?: number;
  similar_incident_ids?: string[];
  metadata?: Record<string, unknown>;
}

export interface Signal {
  id: string;
  org_id: string;
  source_type: string;
  source_id?: string;
  signal_type: string;
  severity?: string;
  title: string;
  body?: string;
  service_name?: string;
  environment?: string;
  tags: string[];
  occurred_at: string;
  received_at: string;
  incident_id?: string;
}

export interface TimelineEvent {
  id: string;
  incident_id: string;
  signal_id?: string;
  event_type: string;
  actor_type?: string;
  actor_id?: string;
  description: string;
  metadata: Record<string, unknown>;
  occurred_at: string;
}

export interface SimilarMatch {
  incident_id: string;
  score: number;
  reason: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    cursor?: string;
    has_more: boolean;
    total: number;
  };
}

export interface AnalyticsOverview {
  total_incidents: number;
  open_incidents: number;
  mttr_minutes: number;
  mtta_minutes: number;
  incidents_this_week: number;
}

export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info' | 'unknown';
export type IncidentStatus = 'open' | 'acknowledged' | 'investigating' | 'mitigated' | 'resolved' | 'closed';
