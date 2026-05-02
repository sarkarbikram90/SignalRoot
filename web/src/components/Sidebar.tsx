import { NavLink } from 'react-router-dom';
import { LayoutDashboard, AlertTriangle, Settings, BarChart3, Zap, BookOpen } from 'lucide-react';

export function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="sidebar-logo">
        <div style={{ width: 32, height: 32, borderRadius: 8, background: 'linear-gradient(135deg, var(--brand-500), var(--brand-700))', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Zap size={18} color="white" />
        </div>
        <h1>SignalRoot</h1>
      </div>
      <nav className="sidebar-nav">
        <NavLink to="/" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`} end>
          <LayoutDashboard size={18} />
          Dashboard
        </NavLink>
        <NavLink to="/incidents" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <AlertTriangle size={18} />
          Incidents
        </NavLink>
        <NavLink to="/analytics" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <BarChart3 size={18} />
          Analytics
        </NavLink>
        <NavLink to="/knowledge" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <BookOpen size={18} />
          Knowledge
        </NavLink>
        <div style={{ flex: 1 }} />
        <NavLink to="/settings" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <Settings size={18} />
          Settings
        </NavLink>
      </nav>
    </aside>
  );
}
