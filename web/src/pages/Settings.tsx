import { useState } from 'react';
import { Plug, Key, Users, Bell, Shield, Plus, Check, X, ExternalLink } from 'lucide-react';

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<'integrations' | 'api-keys' | 'team' | 'notifications'>('integrations');

  const integrations = [
    { id: '1', name: 'PagerDuty', type: 'pagerduty', status: 'active', lastSync: '2m ago', icon: '🔔' },
    { id: '2', name: 'Slack #incidents', type: 'slack', status: 'active', lastSync: '1m ago', icon: '💬' },
    { id: '3', name: 'Datadog', type: 'datadog', status: 'paused', lastSync: 'Never', icon: '📊' },
  ];

  const apiKeys = [
    { id: '1', name: 'CI Pipeline', prefix: 'sr_live_abc...', lastUsed: '5m ago', created: '30d ago' },
    { id: '2', name: 'Terraform', prefix: 'sr_live_def...', lastUsed: '2h ago', created: '14d ago' },
  ];

  const members = [
    { id: '1', name: 'Sarah Chen', email: 'sarah@company.com', role: 'owner', lastActive: 'now' },
    { id: '2', name: 'Alex Kim', email: 'alex@company.com', role: 'admin', lastActive: '1h ago' },
    { id: '3', name: 'Jordan Rivera', email: 'jordan@company.com', role: 'member', lastActive: '3h ago' },
  ];

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div>
          <h1 className="page-title">Settings</h1>
          <p className="page-subtitle">Manage your integrations, team, and preferences</p>
        </div>
      </div>

      {/* Tab Navigation */}
      <div style={{ display: 'flex', gap: 0, borderBottom: '1px solid var(--surface-3)', marginBottom: 24 }}>
        {[
          { key: 'integrations', label: 'Integrations', icon: Plug },
          { key: 'api-keys', label: 'API Keys', icon: Key },
          { key: 'team', label: 'Team', icon: Users },
          { key: 'notifications', label: 'Notifications', icon: Bell },
        ].map(tab => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key as any)}
            style={{
              padding: '12px 20px', background: 'none', border: 'none', cursor: 'pointer',
              fontSize: '0.875rem', fontWeight: 600, display: 'flex', alignItems: 'center', gap: 8,
              color: activeTab === tab.key ? 'var(--brand-400)' : 'var(--text-muted)',
              borderBottom: activeTab === tab.key ? '2px solid var(--brand-400)' : '2px solid transparent',
            }}
          >
            <tab.icon size={16} /> {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'integrations' && (
        <div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
            <button className="btn btn-primary"><Plus size={16} /> Add Integration</button>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {integrations.map(int => (
              <div key={int.id} className="card" style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <span style={{ fontSize: '1.5rem' }}>{int.icon}</span>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 600 }}>{int.name}</div>
                  <div style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>Last sync: {int.lastSync}</div>
                </div>
                <div style={{
                  display: 'flex', alignItems: 'center', gap: 6, padding: '4px 12px',
                  background: int.status === 'active' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(245, 158, 11, 0.15)',
                  color: int.status === 'active' ? 'var(--success)' : 'var(--warning)',
                  borderRadius: 'var(--radius-sm)', fontSize: '0.8125rem', fontWeight: 600,
                }}>
                  {int.status === 'active' ? <Check size={14} /> : <X size={14} />}
                  {int.status}
                </div>
                <button className="btn btn-ghost btn-sm"><ExternalLink size={14} /></button>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'api-keys' && (
        <div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
            <button className="btn btn-primary"><Plus size={16} /> Create API Key</button>
          </div>
          <div className="table-container">
            <table className="data-table">
              <thead><tr><th>Name</th><th>Key</th><th>Last Used</th><th>Created</th><th></th></tr></thead>
              <tbody>
                {apiKeys.map(k => (
                  <tr key={k.id}>
                    <td style={{ fontWeight: 600 }}>{k.name}</td>
                    <td><code style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>{k.prefix}</code></td>
                    <td style={{ color: 'var(--text-muted)' }}>{k.lastUsed}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{k.created}</td>
                    <td><button className="btn btn-danger btn-sm">Revoke</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'team' && (
        <div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
            <button className="btn btn-primary"><Plus size={16} /> Invite Member</button>
          </div>
          <div className="table-container">
            <table className="data-table">
              <thead><tr><th>Name</th><th>Email</th><th>Role</th><th>Last Active</th></tr></thead>
              <tbody>
                {members.map(m => (
                  <tr key={m.id}>
                    <td style={{ fontWeight: 600 }}>{m.name}</td>
                    <td style={{ color: 'var(--text-muted)' }}>{m.email}</td>
                    <td><span className={`badge badge-${m.role === 'owner' ? 'info' : 'low'}`}>{m.role}</span></td>
                    <td style={{ color: 'var(--text-muted)' }}>{m.lastActive}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'notifications' && (
        <div className="card">
          <h3 style={{ fontWeight: 600, marginBottom: 16 }}>Notification Rules</h3>
          <p style={{ color: 'var(--text-muted)', marginBottom: 24 }}>Configure which incidents trigger notifications and where they are sent.</p>
          <div style={{ padding: '20px', background: 'var(--surface-2)', borderRadius: 'var(--radius-md)', border: '1px solid var(--surface-3)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
              <span style={{ fontWeight: 600 }}>Critical incidents → #incidents-p1</span>
              <div style={{ width: 36, height: 20, borderRadius: 10, background: 'var(--success)', position: 'relative', cursor: 'pointer' }}>
                <div style={{ width: 16, height: 16, borderRadius: '50%', background: 'white', position: 'absolute', top: 2, right: 2 }} />
              </div>
            </div>
            <div style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
              Severities: Critical • Cooldown: 5 minutes
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
