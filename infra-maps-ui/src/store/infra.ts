import { create } from 'zustand'
import type { UnifiedNode, Connection, Alert } from '../types/infra'

interface InfraState {
  tree: UnifiedNode | null
  connections: Connection[]
  alerts: Alert[]

  selectedNodeId: string | null

  filterNamespace: string
  filterStatus: string
  filterTag: string

  isLoading: boolean
  isError: boolean
  errorMessage: string
  lastRefresh: Date | null
  cacheAgeSeconds: number | null

  setTree: (tree: UnifiedNode, cacheAgeSeconds: number | null) => void
  setConnections: (connections: Connection[]) => void
  addAlert: (alert: Alert) => void
  removeAlert: (id: string) => void

  selectNode: (nodeId: string | null) => void

  setFilterNamespace: (ns: string) => void
  setFilterStatus: (status: string) => void
  setFilterTag: (tag: string) => void

  setLoading: (v: boolean) => void
  setError: (message: string) => void
  clearError: () => void
}

export const useInfraStore = create<InfraState>((set) => ({
  tree: null,
  connections: [],
  alerts: [],
  selectedNodeId: null,
  filterNamespace: '',
  filterStatus: '',
  filterTag: '',
  isLoading: true,
  isError: false,
  errorMessage: '',
  lastRefresh: null,
  cacheAgeSeconds: null,

  setTree: (tree, cacheAgeSeconds) =>
    set({ tree, isLoading: false, isError: false, lastRefresh: new Date(), cacheAgeSeconds }),
  setConnections: (connections) => set({ connections }),
  addAlert: (alert) => set((s) => ({ alerts: [...s.alerts, alert] })),
  removeAlert: (id) => set((s) => ({ alerts: s.alerts.filter((a) => a.id !== id) })),

  selectNode: (selectedNodeId) => set({ selectedNodeId }),

  setFilterNamespace: (filterNamespace) => set({ filterNamespace }),
  setFilterStatus: (filterStatus) => set({ filterStatus }),
  setFilterTag: (filterTag) => set({ filterTag }),

  setLoading: (isLoading) => set({ isLoading }),
  setError: (errorMessage) => set({ isError: true, errorMessage, isLoading: false }),
  clearError: () => set({ isError: false, errorMessage: '' }),
}))
