import type { Incident } from '../api/types';
import { formatRelativeTime, incidentNumber, truncate, statusLabel } from '../lib/utils';
import { useNavigate } from 'react-router-dom';
import { Clock, Layers, ExternalLink } from 'lucide-react';

interface Props {
  incident: Incident;
}

export function IncidentCard({ incident }: Props) {
  const navigate = useNavigate();

  return (
    <div
      className="card animate-fade-in"
      style={{ cursor: 'pointer', transition: 'all var(--transition-fast)' }}
      onClick={() => navigate(`/incidents/${incident.id}`)}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.8125rem', fontWeight: 600, fontFamily: 'monospace' }}>
              {incidentNumber(incident.number)}
            </span>
            <span className={`badge badge-${incident.severity}`}>
              {incident.severity}
            </span>
            <span className={`status-badge status-${incident.status}`}>
              <span className="status-dot" />
              {statusLabel(incident.status)}
            </span>
          </div>

          <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: 8, lineHeight: 1.4 }}>
            {truncate(incident.title, 100)}
          </h3>

          {incident.summary && (
            <p style={{ color: 'var(--text-muted)', fontSize: '0.8125rem', marginBottom: 12, lineHeight: 1.5 }}>
              {truncate(incident.summary, 150)}
            </p>
          )}

          <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
            {incident.services_affected?.length > 0 && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <Layers size={14} color="var(--text-muted)" />
                <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                  {incident.services_affected.slice(0, 3).map(svc => (
                    <span key={svc} style={{
                      padding: '2px 8px',
                      background: 'var(--surface-2)',
                      borderRadius: 'var(--radius-sm)',
                      fontSize: '0.75rem',
                      color: 'var(--text-secondary)',
                    }}>
                      {svc}
                    </span>
                  ))}
                  {incident.services_affected.length > 3 && (
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      +{incident.services_affected.length - 3}
                    </span>
                  )}
                </div>
              </div>
            )}

            <div style={{ display: 'flex', alignItems: 'center', gap: 4, color: 'var(--text-muted)', fontSize: '0.8125rem' }}>
              <Clock size={14} />
              {formatRelativeTime(incident.detected_at)}
            </div>

            {(incident.signal_count ?? 0) > 0 && (
              <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                {incident.signal_count} signals
              </span>
            )}
          </div>
        </div>

        <ExternalLink size={16} color="var(--text-muted)" style={{ flexShrink: 0, marginTop: 4 }} />
      </div>
    </div>
  );
}
