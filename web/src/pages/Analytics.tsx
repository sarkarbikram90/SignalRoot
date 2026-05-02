import { BarChart3, TrendingDown, Clock, AlertTriangle } from 'lucide-react';

const mockServices = [
  { name: 'payment-service', incidents: 12, mttr: 34, lastIncident: '2h ago', severity: 'critical' },
  { name: 'user-service', incidents: 8, mttr: 22, lastIncident: '1d ago', severity: 'high' },
  { name: 'api-gateway', incidents: 6, mttr: 15, lastIncident: '3d ago', severity: 'medium' },
  { name: 'notification-service', incidents: 4, mttr: 48, lastIncident: '5d ago', severity: 'low' },
  { name: 'search-service', incidents: 2, mttr: 12, lastIncident: '2w ago', severity: 'low' },
];

export function AnalyticsPage() {
  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div>
          <h1 className="page-title">Analytics</h1>
          <p className="page-subtitle">Incident trends and service health</p>
        </div>
      </div>

      {/* Key Metrics */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16, marginBottom: 32 }}>
        <div className="stat-card">
          <span className="stat-label">Total Incidents (30d)</span>
          <span className="stat-value">32</span>
          <span className="stat-change negative">+12% vs previous period</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Avg MTTR</span>
          <span className="stat-value">47m</span>
          <span className="stat-change positive">-15% improvement</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">Avg MTTA</span>
          <span className="stat-value">4.2m</span>
          <span className="stat-change positive">-22% improvement</span>
        </div>
        <div className="stat-card">
          <span className="stat-label">P1 Incidents</span>
          <span className="stat-value">5</span>
          <span className="stat-change negative">+2 vs previous</span>
        </div>
      </div>

      {/* Services Table */}
      <div style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
          <BarChart3 size={20} color="var(--brand-400)" />
          Top Services by Incident Volume
        </h2>
        <div className="table-container">
          <table className="data-table">
            <thead>
              <tr>
                <th>Service</th>
                <th>Incidents (30d)</th>
                <th>Avg MTTR</th>
                <th>Last Incident</th>
                <th>Trend</th>
              </tr>
            </thead>
            <tbody>
              {mockServices.map(svc => (
                <tr key={svc.name}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <div style={{
                        width: 8, height: 8, borderRadius: '50%',
                        background: svc.severity === 'critical' ? 'var(--critical)' : svc.severity === 'high' ? 'var(--high)' : 'var(--success)',
                      }} />
                      <span style={{ fontWeight: 600 }}>{svc.name}</span>
                    </div>
                  </td>
                  <td>{svc.incidents}</td>
                  <td>{svc.mttr}m</td>
                  <td style={{ color: 'var(--text-muted)' }}>{svc.lastIncident}</td>
                  <td>
                    <TrendingDown size={16} color="var(--success)" />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Weekly Trend Placeholder */}
      <div className="card">
        <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: 16 }}>Weekly Incident Trend</h3>
        <div style={{
          height: 200, display: 'flex', alignItems: 'flex-end', gap: 12, padding: '0 20px',
          borderBottom: '1px solid var(--surface-3)',
        }}>
          {[5, 8, 3, 12, 7, 4, 6].map((val, i) => (
            <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
              <div style={{
                width: '100%', height: val * 15,
                background: `linear-gradient(to top, var(--brand-600), var(--brand-400))`,
                borderRadius: '4px 4px 0 0',
                transition: 'height var(--transition-slow)',
              }} />
            </div>
          ))}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 20px 0', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
          {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(d => <span key={d}>{d}</span>)}
        </div>
      </div>
    </div>
  );
}
