import type { Incident, PaginatedResponse, TimelineEvent, AnalyticsOverview } from './types';

const API_BASE = import.meta.env.VITE_API_URL || '';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

// ─── Incidents ──────────────────────────────────────────────────────────────
export function listIncidents(params?: Record<string, string>) {
  const query = params ? '?' + new URLSearchParams(params).toString() : '';
  return request<PaginatedResponse<Incident>>(`/api/v1/incidents${query}`);
}

export function getIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}`);
}

export function createIncident(data: { title: string; severity: string; services?: string[] }) {
  return request<Incident>('/api/v1/incidents', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateIncident(id: string, data: Partial<Incident>) {
  return request<Incident>(`/api/v1/incidents/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// ─── Status Transitions ─────────────────────────────────────────────────────
export function acknowledgeIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}/acknowledge`, { method: 'POST' });
}
export function investigateIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}/investigate`, { method: 'POST' });
}
export function mitigateIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}/mitigate`, { method: 'POST' });
}
export function resolveIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}/resolve`, { method: 'POST' });
}
export function closeIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}/close`, { method: 'POST' });
}
export function reopenIncident(id: string) {
  return request<Incident>(`/api/v1/incidents/${id}/reopen`, { method: 'POST' });
}

// ─── Timeline ───────────────────────────────────────────────────────────────
export function getTimeline(incidentId: string) {
  return request<{ data: TimelineEvent[] }>(`/api/v1/incidents/${incidentId}/timeline`);
}

export function addTimelineEvent(incidentId: string, data: { description: string; event_type?: string }) {
  return request<TimelineEvent>(`/api/v1/incidents/${incidentId}/timeline`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// ─── Similar ────────────────────────────────────────────────────────────────
export function getSimilarIncidents(incidentId: string) {
  return request<{ data: Incident[]; message?: string }>(`/api/v1/incidents/${incidentId}/similar`);
}

// ─── RCA ────────────────────────────────────────────────────────────────────
export function triggerRCA(incidentId: string) {
  return request<{ status: string; message: string }>(`/api/v1/incidents/${incidentId}/rca`, { method: 'POST' });
}

// ─── Analytics ──────────────────────────────────────────────────────────────
export function getAnalyticsOverview() {
  return request<AnalyticsOverview>('/api/v1/analytics/overview');
}
