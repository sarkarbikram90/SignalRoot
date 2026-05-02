import type { Severity, IncidentStatus } from '../api/types';

export function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;

  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;

  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 30) return `${diffDays}d ago`;

  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function formatDuration(startStr: string, endStr?: string): string {
  const start = new Date(startStr);
  const end = endStr ? new Date(endStr) : new Date();
  const diffMs = end.getTime() - start.getTime();
  const mins = Math.floor(diffMs / 60000);

  if (mins < 60) return `${mins}m`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ${mins % 60}m`;
  const days = Math.floor(hours / 24);
  return `${days}d ${hours % 24}h`;
}

export function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function severityColor(severity: Severity): string {
  const colors: Record<string, string> = {
    critical: 'var(--critical)',
    high: 'var(--high)',
    medium: 'var(--medium)',
    low: 'var(--low)',
    info: 'var(--info)',
    unknown: 'var(--text-muted)',
  };
  return colors[severity] || colors.unknown;
}

export function statusLabel(status: IncidentStatus): string {
  const labels: Record<string, string> = {
    open: 'Open',
    acknowledged: 'Acknowledged',
    investigating: 'Investigating',
    mitigated: 'Mitigated',
    resolved: 'Resolved',
    closed: 'Closed',
  };
  return labels[status] || status;
}

export function incidentNumber(num: number): string {
  return `INC-${String(num).padStart(4, '0')}`;
}

export function truncate(text: string, maxLen: number = 80): string {
  if (text.length <= maxLen) return text;
  return text.substring(0, maxLen) + '…';
}
