import { useEffect, useState } from 'react';
import { Activity, AlertTriangle, Clock, CheckCircle, TrendingUp, Zap } from 'lucide-react';
import type { Incident } from '../api/types';
import { IncidentCard } from '../components/IncidentCard';
import { useIncidentStore } from '../store/incidents';

// Demo data for when backend isn't connected
const demoIncidents: Incident[] = [
  {
    id: '1', org_id: '1', number: 42, title: 'Payment service latency spike — p99 exceeding 2000ms',
    severity: 'critical', status: 'investigating', services_affected: ['payment-service', 'checkout-api'],
    environments: ['production'], tags: ['latency', 'customer-facing'], detected_at: new Date(Date.now() - 3600000).toISOString(),
    created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 12,
    summary: 'Payment processing latency has exceeded acceptable thresholds. Multiple customers reporting failed transactions.',
  },
  {
    id: '2', org_id: '1', number: 41, title: 'Database connection pool exhaustion on user-service',
    severity: 'high', status: 'acknowledged', services_affected: ['user-service', 'auth-service'],
    environments: ['production'], tags: ['database', 'connection-pool'], detected_at: new Date(Date.now() - 7200000).toISOString(),
    created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 8,
    summary: 'Connection pool reaching maximum capacity, causing intermittent 500 errors on authentication endpoints.',
  },
  {
    id: '3', org_id: '1', number: 40, title: 'CDN cache invalidation failure — stale assets served',
    severity: 'medium', status: 'mitigated', services_affected: ['cdn', 'frontend'],
    environments: ['production'], tags: ['cdn', 'caching'], detected_at: new Date(Date.now() - 86400000).toISOString(),
    resolved_at: new Date(Date.now() - 82800000).toISOString(),
    created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 5,
  },
  {
    id: '4', org_id: '1', number: 39, title: 'Deployment rollback — API version mismatch',
    severity: 'high', status: 'resolved', services_affected: ['api-gateway'],
    environments: ['production'], tags: ['deployment', 'rollback'], detected_at: new Date(Date.now() - 172800000).toISOString(),
    resolved_at: new Date(Date.now() - 169200000).toISOString(),
    created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 15,
    root_cause: 'Incompatible API schema change deployed without backward compatibility layer.',
    root_cause_confidence: 0.85,
  },
  {
    id: '5', org_id: '1', number: 38, title: 'Kafka consumer lag exceeding 10k messages',
    severity: 'low', status: 'closed', services_affected: ['event-processor', 'notification-service'],
    environments: ['production'], tags: ['kafka', 'consumer-lag'], detected_at: new Date(Date.now() - 259200000).toISOString(),
    resolved_at: new Date(Date.now() - 255600000).toISOString(),
    created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 3,
  },
];

const stats = [
  { label: 'Open Incidents', value: '3', icon: AlertTriangle, color: 'var(--critical)', change: '+1', positive: false },
  { label: 'MTTA', value: '4.2m', icon: Clock, color: 'var(--warning)', change: '-12%', positive: true },
  { label: 'MTTR', value: '47m', icon: CheckCircle, color: 'var(--success)', change: '-8%', positive: true },
  { label: 'This Week', value: '7', icon: TrendingUp, color: 'var(--brand-400)', change: '-3', positive: true },
];

export function DashboardPage() {
  const { incidents, fetchIncidents, loading } = useIncidentStore();
  const [useDemoData, setUseDemoData] = useState(false);

  useEffect(() => {
    fetchIncidents().catch(() => setUseDemoData(true));
  }, []);

  const displayIncidents = useDemoData || incidents.length === 0 ? demoIncidents : incidents;
  const openIncidents = displayIncidents.filter(i => !['resolved', 'closed'].includes(i.status));

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div>
          <h1 className="page-title">Dashboard</h1>
          <p className="page-subtitle">Incident intelligence at a glance</p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{
            width: 8, height: 8, borderRadius: '50%',
            background: useDemoData ? 'var(--warning)' : 'var(--success)',
            animation: 'pulse-dot 2s ease-in-out infinite',
          }} />
          <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
            {useDemoData ? 'Demo mode' : 'Live'}
          </span>
        </div>
      </div>

      {/* Stats Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16, marginBottom: 32 }}>
        {stats.map(stat => (
          <div key={stat.label} className="stat-card" style={{ animationDelay: '0.1s' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span className="stat-label">{stat.label}</span>
              <stat.icon size={18} color={stat.color} />
            </div>
            <span className="stat-value">{stat.value}</span>
            <span className={`stat-change ${stat.positive ? 'positive' : 'negative'}`}>
              {stat.change} vs last week
            </span>
          </div>
        ))}
      </div>

      {/* Active Incidents */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 600, display: 'flex', alignItems: 'center', gap: 8 }}>
            <Activity size={20} color="var(--critical)" />
            Active Incidents
            <span style={{
              background: 'var(--critical-soft)',
              color: 'var(--critical)',
              padding: '2px 10px',
              borderRadius: '9999px',
              fontSize: '0.8125rem',
              fontWeight: 700,
            }}>
              {openIncidents.length}
            </span>
          </h2>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {openIncidents.map((inc, i) => (
            <div key={inc.id} style={{ animationDelay: `${i * 0.05}s` }} className="animate-fade-in">
              <IncidentCard incident={inc} />
            </div>
          ))}
        </div>
      </div>

      {/* Recent Resolved */}
      <div>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
          <CheckCircle size={20} color="var(--success)" />
          Recently Resolved
        </h2>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {displayIncidents.filter(i => ['resolved', 'closed'].includes(i.status)).slice(0, 3).map(inc => (
            <IncidentCard key={inc.id} incident={inc} />
          ))}
        </div>
      </div>
    </div>
  );
}
