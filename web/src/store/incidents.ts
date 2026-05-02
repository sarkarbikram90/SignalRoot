import { create } from 'zustand';
import type { Incident, IncidentStatus } from '../api/types';
import * as api from '../api/client';

interface IncidentStore {
  incidents: Incident[];
  currentIncident: Incident | null;
  loading: boolean;
  error: string | null;
  filters: {
    status?: string;
    severity?: string;
    search?: string;
  };
  total: number;

  fetchIncidents: () => Promise<void>;
  fetchIncident: (id: string) => Promise<void>;
  setFilter: (key: string, value: string | undefined) => void;
  transitionStatus: (id: string, action: string) => Promise<void>;
  createIncident: (data: { title: string; severity: string; services?: string[] }) => Promise<Incident>;
}

export const useIncidentStore = create<IncidentStore>((set, get) => ({
  incidents: [],
  currentIncident: null,
  loading: false,
  error: null,
  filters: {},
  total: 0,

  fetchIncidents: async () => {
    set({ loading: true, error: null });
    try {
      const { filters } = get();
      const params: Record<string, string> = {};
      if (filters.status) params.status = filters.status;
      if (filters.severity) params.severity = filters.severity;
      if (filters.search) params.q = filters.search;

      const res = await api.listIncidents(params);
      set({ incidents: res.data, total: res.meta.total, loading: false });
    } catch (err) {
      set({ error: (err as Error).message, loading: false });
    }
  },

  fetchIncident: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const incident = await api.getIncident(id);
      set({ currentIncident: incident, loading: false });
    } catch (err) {
      set({ error: (err as Error).message, loading: false });
    }
  },

  setFilter: (key, value) => {
    set(state => ({
      filters: { ...state.filters, [key]: value },
    }));
    get().fetchIncidents();
  },

  transitionStatus: async (id: string, action: string) => {
    const actionMap: Record<string, (id: string) => Promise<Incident>> = {
      acknowledge: api.acknowledgeIncident,
      investigate: api.investigateIncident,
      mitigate: api.mitigateIncident,
      resolve: api.resolveIncident,
      close: api.closeIncident,
      reopen: api.reopenIncident,
    };
    const fn = actionMap[action];
    if (!fn) throw new Error(`Unknown action: ${action}`);

    // Optimistic update
    const oldIncident = get().currentIncident;
    try {
      const updated = await fn(id);
      set({ currentIncident: updated });
      // Also update in list
      set(state => ({
        incidents: state.incidents.map(i => i.id === id ? updated : i),
      }));
    } catch (err) {
      set({ currentIncident: oldIncident, error: (err as Error).message });
      throw err;
    }
  },

  createIncident: async (data) => {
    const incident = await api.createIncident(data);
    set(state => ({ incidents: [incident, ...state.incidents] }));
    return incident;
  },
}));
