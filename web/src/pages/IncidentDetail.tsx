import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Clock, AlertTriangle, CheckCircle, Users, Brain, FileText, Zap, MessageSquare, Shield } from 'lucide-react';
import { useIncidentStore } from '../store/incidents';
import { formatRelativeTime, formatDuration, formatDateTime, incidentNumber, statusLabel, severityColor } from '../lib/utils';
import type { Incident, TimelineEvent } from '../api/types';
import * as api from '../api/client';

// Demo data
const demoIncident: Incident = {
  id: '1', org_id: '1', number: 42,
  title: 'Payment service latency spike — p99 exceeding 2000ms',
  summary: 'Payment processing latency has exceeded acceptable thresholds. Multiple PagerDuty alerts triggered for payment-service and checkout-api. Customer support reporting increased failed transaction complaints.',
  severity: 'critical', status: 'investigating',
  services_affected: ['payment-service', 'checkout-api', 'stripe-integration'],
  environments: ['production'], tags: ['latency', 'customer-facing', 'payments'],
  detected_at: new Date(Date.now() - 3600000).toISOString(),
  acknowledged_at: new Date(Date.now() - 3500000).toISOString(),
  created_at: new Date().toISOString(), updated_at: new Date().toISOString(),
  signal_count: 12,
  root_cause: 'Database connection pool exhaustion in payment-service caused by a connection leak introduced in deploy v2.14.3. The connection leak was triggered when Stripe webhook processing timed out, leaving connections in a half-open state.',
  root_cause_confidence: 0.82,
  metadata: { cascading: false },
};

const demoTimeline: TimelineEvent[] = [
  { id: '1', incident_id: '1', event_type: 'signal', actor_type: 'system', description: 'PagerDuty alert: payment-service p99 latency > 2000ms', metadata: { source_type: 'pagerduty' }, occurred_at: new Date(Date.now() - 3600000).toISOString() },
  { id: '2', incident_id: '1', event_type: 'signal', actor_type: 'system', description: 'PagerDuty alert: checkout-api error rate > 5%', metadata: { source_type: 'pagerduty' }, occurred_at: new Date(Date.now() - 3540000).toISOString() },
  { id: '3', incident_id: '1', event_type: 'status_change', actor_type: 'user', description: 'Status changed to acknowledged', metadata: {}, occurred_at: new Date(Date.now() - 3500000).toISOString() },
  { id: '4', incident_id: '1', event_type: 'comment', actor_type: 'user', description: 'Investigating database connection metrics. Seeing unusual pool exhaustion pattern.', metadata: {}, occurred_at: new Date(Date.now() - 3200000).toISOString() },
  { id: '5', incident_id: '1', event_type: 'status_change', actor_type: 'user', description: 'Status changed to investigating', metadata: {}, occurred_at: new Date(Date.now() - 3100000).toISOString() },
  { id: '6', incident_id: '1', event_type: 'ai_insight', actor_type: 'ai', description: 'Root cause analysis generated: Database connection pool exhaustion likely caused by connection leak in v2.14.3 deploy', metadata: {}, occurred_at: new Date(Date.now() - 2800000).toISOString() },
  { id: '7', incident_id: '1', event_type: 'signal', actor_type: 'system', description: 'Slack: @oncall investigating DB connection pool — rolling back v2.14.3', metadata: { source_type: 'slack' }, occurred_at: new Date(Date.now() - 2500000).toISOString() },
];

const statusActions: Record<string, { label: string; action: string; icon: any; variant: string }[]> = {
  open: [{ label: 'Acknowledge', action: 'acknowledge', icon: CheckCircle, variant: 'btn-secondary' }],
  acknowledged: [{ label: 'Investigate', action: 'investigate', icon: Zap, variant: 'btn-primary' }],
  investigating: [
    { label: 'Mitigate', action: 'mitigate', icon: Shield, variant: 'btn-secondary' },
    { label: 'Resolve', action: 'resolve', icon: CheckCircle, variant: 'btn-primary' },
  ],
  mitigated: [{ label: 'Resolve', action: 'resolve', icon: CheckCircle, variant: 'btn-primary' }],
  resolved: [{ label: 'Close', action: 'close', icon: CheckCircle, variant: 'btn-secondary' }],
  closed: [{ label: 'Reopen', action: 'reopen', icon: AlertTriangle, variant: 'btn-danger' }],
};

export function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentIncident, fetchIncident, transitionStatus } = useIncidentStore();
  const [timeline, setTimeline] = useState<TimelineEvent[]>([]);
  const [useDemoData, setUseDemoData] = useState(false);
  const [activeTab, setActiveTab] = useState<'timeline' | 'similar' | 'rca' | 'report'>('timeline');
  const [comment, setComment] = useState('');

  useEffect(() => {
    if (id) {
      fetchIncident(id).catch(() => setUseDemoData(true));
      api.getTimeline(id).then(r => setTimeline(r.data)).catch(() => {});
    }
  }, [id]);

  const incident = useDemoData ? demoIncident : currentIncident;
  const displayTimeline = useDemoData || timeline.length === 0 ? demoTimeline : timeline;

  if (!incident) {
    return (
      <div style={{ padding: 40, textAlign: 'center' }}>
        <div className="skeleton" style={{ width: 300, height: 24, margin: '0 auto 16px' }} />
        <div className="skeleton" style={{ width: 500, height: 16, margin: '0 auto' }} />
      </div>
    );
  }

  const actions = statusActions[incident.status] || [];

  const handleAction = async (action: string) => {
    try {
      await transitionStatus(incident.id, action);
    } catch {
      // Error handled in store
    }
  };

  const eventIcon = (type: string) => {
    switch (type) {
      case 'signal': return <Zap size={14} color="var(--brand-400)" />;
      case 'status_change': return <CheckCircle size={14} color="var(--success)" />;
      case 'comment': return <MessageSquare size={14} color="var(--text-secondary)" />;
      case 'ai_insight': return <Brain size={14} color="var(--info)" />;
      default: return <Clock size={14} color="var(--text-muted)" />;
    }
  };

  return (
    <div className="animate-fade-in">
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <button onClick={() => navigate('/incidents')} className="btn btn-ghost btn-sm" style={{ marginBottom: 12 }}>
          <ArrowLeft size={16} /> Back to incidents
        </button>

        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 16 }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
              <span style={{ color: 'var(--text-muted)', fontFamily: 'monospace', fontWeight: 600 }}>
                {incidentNumber(incident.number)}
              </span>
              <span className={`badge badge-${incident.severity}`}>{incident.severity}</span>
              <span className={`status-badge status-${incident.status}`}>
                <span className="status-dot" />
                {statusLabel(incident.status)}
              </span>
            </div>
            <h1 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: 8 }}>{incident.title}</h1>
            <div style={{ display: 'flex', gap: 16, color: 'var(--text-muted)', fontSize: '0.8125rem' }}>
              <span><Clock size={14} style={{ verticalAlign: -2 }} /> Detected {formatRelativeTime(incident.detected_at)}</span>
              <span>Duration: {formatDuration(incident.detected_at, incident.resolved_at)}</span>
              <span>{incident.signal_count ?? 0} signals</span>
            </div>
          </div>

          <div style={{ display: 'flex', gap: 8 }}>
            {actions.map(a => (
              <button key={a.action} className={`btn ${a.variant}`} onClick={() => handleAction(a.action)}>
                <a.icon size={16} /> {a.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Services */}
      {incident.services_affected.length > 0 && (
        <div style={{ display: 'flex', gap: 8, marginBottom: 24, flexWrap: 'wrap' }}>
          {incident.services_affected.map(svc => (
            <span key={svc} style={{
              padding: '6px 14px', background: 'var(--surface-2)', border: '1px solid var(--surface-3)',
              borderRadius: 'var(--radius-md)', fontSize: '0.8125rem', color: 'var(--text-secondary)',
            }}>{svc}</span>
          ))}
        </div>
      )}

      {/* Summary */}
      {incident.summary && (
        <div className="card" style={{ marginBottom: 24 }}>
          <h3 style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: 8, display: 'flex', alignItems: 'center', gap: 6 }}>
            <Brain size={14} color="var(--brand-400)" /> AI Summary
          </h3>
          <p style={{ lineHeight: 1.7 }}>{incident.summary}</p>
        </div>
      )}

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 0, borderBottom: '1px solid var(--surface-3)', marginBottom: 24 }}>
        {(['timeline', 'rca', 'similar', 'report'] as const).map(tab => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            style={{
              padding: '12px 20px', background: 'none', border: 'none', cursor: 'pointer',
              fontSize: '0.875rem', fontWeight: 600,
              color: activeTab === tab ? 'var(--brand-400)' : 'var(--text-muted)',
              borderBottom: activeTab === tab ? '2px solid var(--brand-400)' : '2px solid transparent',
              transition: 'all var(--transition-fast)',
            }}
          >
            {tab === 'timeline' && 'Timeline'}
            {tab === 'rca' && 'Root Cause'}
            {tab === 'similar' && 'Similar'}
            {tab === 'report' && 'Report'}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      {activeTab === 'timeline' && (
        <div>
          <div className="timeline">
            {displayTimeline.map(evt => (
              <div key={evt.id} className="timeline-item animate-slide-in">
                <div className="timeline-dot">{eventIcon(evt.event_type)}</div>
                <div className="timeline-content">
                  <p style={{ fontSize: '0.875rem', marginBottom: 4 }}>{evt.description}</p>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    {formatDateTime(evt.occurred_at)}
                    {evt.actor_type === 'ai' && <span style={{ marginLeft: 8, color: 'var(--info)' }}>🤖 AI</span>}
                  </span>
                </div>
              </div>
            ))}
          </div>

          {/* Add comment */}
          <div style={{ display: 'flex', gap: 12, marginTop: 24 }}>
            <input
              className="input"
              placeholder="Add a comment to the timeline..."
              value={comment}
              onChange={e => setComment(e.target.value)}
            />
            <button className="btn btn-secondary" onClick={() => { setComment(''); }}>
              <MessageSquare size={16} /> Add
            </button>
          </div>
        </div>
      )}

      {activeTab === 'rca' && (
        <div className="card">
          {incident.root_cause ? (
            <>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 600, display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Brain size={18} color="var(--brand-400)" /> Root Cause Analysis
                </h3>
                {incident.root_cause_confidence != null && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>Confidence</span>
                    <div style={{ width: 100 }}>
                      <div className="confidence-bar">
                        <div className="confidence-fill" style={{
                          width: `${incident.root_cause_confidence * 100}%`,
                          background: incident.root_cause_confidence > 0.7 ? 'var(--success)' : incident.root_cause_confidence > 0.4 ? 'var(--warning)' : 'var(--error)',
                        }} />
                      </div>
                    </div>
                    <span style={{ fontSize: '0.8125rem', fontWeight: 600 }}>{Math.round(incident.root_cause_confidence * 100)}%</span>
                  </div>
                )}
              </div>
              <p style={{ lineHeight: 1.8, color: 'var(--text-secondary)' }}>{incident.root_cause}</p>
              <div style={{ marginTop: 16 }}>
                <button className="btn btn-ghost btn-sm">🔄 Regenerate RCA</button>
              </div>
            </>
          ) : (
            <div className="empty-state">
              <div className="empty-state-icon"><Brain size={28} color="var(--text-muted)" /></div>
              <h3>No root cause analysis yet</h3>
              <p>RCA is generated automatically when an incident is resolved, or you can trigger it manually.</p>
              <button className="btn btn-primary" onClick={() => id && api.triggerRCA(id)}>
                <Brain size={16} /> Generate RCA
              </button>
            </div>
          )}
        </div>
      )}

      {activeTab === 'similar' && (
        <div className="card">
          <div className="empty-state">
            <div className="empty-state-icon"><Zap size={28} color="var(--text-muted)" /></div>
            <h3>Not enough history yet</h3>
            <p>Similarity search requires at least 10 resolved incidents to provide meaningful matches. Keep resolving incidents and the system will learn.</p>
          </div>
        </div>
      )}

      {activeTab === 'report' && (
        <div className="card">
          <div className="empty-state">
            <div className="empty-state-icon"><FileText size={28} color="var(--text-muted)" /></div>
            <h3>Generate a compliance report</h3>
            <p>Create audit-ready reports for SOC2 compliance requirements.</p>
            <button className="btn btn-primary">
              <FileText size={16} /> Generate SOC2 Report
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
