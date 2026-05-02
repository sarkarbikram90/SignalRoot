import { useEffect, useState } from 'react';
import { Plus, Search, Filter } from 'lucide-react';
import { IncidentCard } from '../components/IncidentCard';
import { useIncidentStore } from '../store/incidents';
import type { Incident } from '../api/types';

const demoIncidents: Incident[] = [
  { id: '1', org_id: '1', number: 42, title: 'Payment service latency spike — p99 exceeding 2000ms', severity: 'critical', status: 'investigating', services_affected: ['payment-service', 'checkout-api'], environments: ['production'], tags: ['latency'], detected_at: new Date(Date.now() - 3600000).toISOString(), created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 12, summary: 'Payment processing latency has exceeded thresholds.' },
  { id: '2', org_id: '1', number: 41, title: 'Database connection pool exhaustion on user-service', severity: 'high', status: 'acknowledged', services_affected: ['user-service'], environments: ['production'], tags: ['database'], detected_at: new Date(Date.now() - 7200000).toISOString(), created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 8 },
  { id: '3', org_id: '1', number: 40, title: 'CDN cache invalidation failure', severity: 'medium', status: 'mitigated', services_affected: ['cdn'], environments: ['production'], tags: ['cdn'], detected_at: new Date(Date.now() - 86400000).toISOString(), created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 5 },
  { id: '4', org_id: '1', number: 39, title: 'Deployment rollback — API version mismatch', severity: 'high', status: 'resolved', services_affected: ['api-gateway'], environments: ['production'], tags: ['deployment'], detected_at: new Date(Date.now() - 172800000).toISOString(), resolved_at: new Date(Date.now() - 169200000).toISOString(), created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 15, root_cause: 'Incompatible API schema change.', root_cause_confidence: 0.85 },
  { id: '5', org_id: '1', number: 38, title: 'Kafka consumer lag exceeding 10k messages', severity: 'low', status: 'closed', services_affected: ['event-processor'], environments: ['production'], tags: ['kafka'], detected_at: new Date(Date.now() - 259200000).toISOString(), resolved_at: new Date(Date.now() - 255600000).toISOString(), created_at: new Date().toISOString(), updated_at: new Date().toISOString(), signal_count: 3 },
];

const statuses = ['all', 'open', 'acknowledged', 'investigating', 'mitigated', 'resolved', 'closed'];
const severities = ['all', 'critical', 'high', 'medium', 'low'];

export function IncidentListPage() {
  const { incidents, fetchIncidents, filters, setFilter } = useIncidentStore();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [severityFilter, setSeverityFilter] = useState('all');
  const [useDemoData, setUseDemoData] = useState(false);
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    fetchIncidents().catch(() => setUseDemoData(true));
  }, []);

  let displayIncidents = useDemoData || incidents.length === 0 ? demoIncidents : incidents;

  if (statusFilter !== 'all') {
    displayIncidents = displayIncidents.filter(i => i.status === statusFilter);
  }
  if (severityFilter !== 'all') {
    displayIncidents = displayIncidents.filter(i => i.severity === severityFilter);
  }
  if (search) {
    const q = search.toLowerCase();
    displayIncidents = displayIncidents.filter(i =>
      i.title.toLowerCase().includes(q) ||
      i.summary?.toLowerCase().includes(q) ||
      i.services_affected.some(s => s.toLowerCase().includes(q))
    );
  }

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div>
          <h1 className="page-title">Incidents</h1>
          <p className="page-subtitle">{displayIncidents.length} incidents</p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
          <Plus size={16} />
          New Incident
        </button>
      </div>

      {/* Search & Filters */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ position: 'relative', marginBottom: 16 }}>
          <Search size={18} style={{ position: 'absolute', left: 14, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
          <input
            className="input"
            placeholder="Search incidents by title, service, or keyword..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            style={{ paddingLeft: 42 }}
          />
        </div>

        <div className="filter-bar">
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, color: 'var(--text-muted)', fontSize: '0.8125rem' }}>
            <Filter size={14} /> Status:
          </div>
          {statuses.map(s => (
            <button
              key={s}
              className={`filter-chip ${statusFilter === s ? 'active' : ''}`}
              onClick={() => setStatusFilter(s)}
            >
              {s === 'all' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>

        <div className="filter-bar">
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, color: 'var(--text-muted)', fontSize: '0.8125rem' }}>
            <Filter size={14} /> Severity:
          </div>
          {severities.map(s => (
            <button
              key={s}
              className={`filter-chip ${severityFilter === s ? 'active' : ''}`}
              onClick={() => setSeverityFilter(s)}
            >
              {s === 'all' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* Incident List */}
      {displayIncidents.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">
            <AlertTriangle size={28} color="var(--text-muted)" />
          </div>
          <h3>No incidents found</h3>
          <p>No incidents match your current filters. Try adjusting your search or connect your first integration.</p>
          <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
            <Plus size={16} /> Create Incident
          </button>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {displayIncidents.map((inc, i) => (
            <div key={inc.id} className="animate-fade-in" style={{ animationDelay: `${i * 0.03}s` }}>
              <IncidentCard incident={inc} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function AlertTriangle2(props: { size: number; color: string }) {
  return <AlertTriangle {...props} />;
}
