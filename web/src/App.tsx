import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Sidebar } from './components/Sidebar';
import { DashboardPage } from './pages/Dashboard';
import { IncidentListPage } from './pages/IncidentList';
import { IncidentDetailPage } from './pages/IncidentDetail';
import { AnalyticsPage } from './pages/Analytics';
import { SettingsPage } from './pages/Settings';

function App() {
  return (
    <BrowserRouter>
      <div className="app-layout">
        <Sidebar />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/incidents" element={<IncidentListPage />} />
            <Route path="/incidents/:id" element={<IncidentDetailPage />} />
            <Route path="/analytics" element={<AnalyticsPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/knowledge" element={<KnowledgePlaceholder />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
}

function KnowledgePlaceholder() {
  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div>
          <h1 className="page-title">Knowledge Base</h1>
          <p className="page-subtitle">Services, failure modes, and institutional memory</p>
        </div>
      </div>
      <div className="empty-state">
        <div className="empty-state-icon" style={{ background: 'var(--surface-2)' }}>
          📚
        </div>
        <h3>Coming soon</h3>
        <p>The knowledge graph builds automatically as incidents are resolved. Keep investigating and resolving incidents to grow your institutional memory.</p>
      </div>
    </div>
  );
}

export default App;
