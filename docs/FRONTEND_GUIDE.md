# Frontend Development Guide — React

> Lire `docs/DATA_MODELS.md` et `docs/DESIGN_SYSTEM.md` avant de coder.
> Stack : React 19, TypeScript strict, Zustand, Recharts, Vite.

---

## Structure du Projet

```
infra-maps-ui/
├── src/
│   ├── api/
│   │   ├── client.ts          # fetch wrapper avec error handling centralisé
│   │   └── websocket.ts       # WebSocket manager (reconnexion auto)
│   ├── store/
│   │   └── infra.ts           # Zustand store — état global
│   ├── hooks/
│   │   ├── useInfraTree.ts    # Polling arbre toutes les 30s
│   │   ├── useNodeDetail.ts   # Fetch détails d'un nœud
│   │   ├── useNodeMetrics.ts  # Fetch métriques historiques
│   │   ├── useNodeLogs.ts     # Fetch logs récents
│   │   └── useAlerts.ts       # WebSocket alertes
│   ├── views/
│   │   ├── GridView/
│   │   │   ├── index.tsx        # Vue Grid 2D — entry point
│   │   │   ├── ZoneRenderer.tsx # Rendu d'une zone (cluster, island)
│   │   │   ├── NodeCard.tsx     # Carte individuelle d'un nœud
│   │   │   └── ConnectionSVG.tsx # Lignes SVG de connexion
│   │   └── MapView/             # Phase 4 — Three.js — ne pas créer avant
│   ├── components/
│   │   ├── SidePanel/
│   │   │   ├── index.tsx
│   │   │   ├── MetricsChart.tsx # Recharts pour l'historique
│   │   │   └── LogViewer.tsx
│   │   ├── Toolbar/
│   │   │   ├── index.tsx
│   │   │   ├── FilterBar.tsx
│   │   │   └── SearchInput.tsx
│   │   └── shared/
│   │       ├── StatusBadge.tsx  # Badge statut (couleur + icône + aria-label)
│   │       ├── MetricValue.tsx  # Valeur formatée (CPU, RAM) en JetBrains Mono
│   │       ├── EmptyState.tsx   # État vide ou source indisponible
│   │       └── LoadingSkeleton.tsx
│   ├── styles/
│   │   ├── variables.css      # CSS variables du design system (voir DESIGN_SYSTEM.md)
│   │   └── globals.css        # Reset minimal, body background
│   ├── utils/
│   │   ├── status.ts          # statusToColor(), statusToIcon(), statusToLabel()
│   │   ├── format.ts          # formatCPU(), formatMemory(), formatBytes()
│   │   └── tree.ts            # findNodeById(), flattenTree(), getAncestors()
│   ├── types/
│   │   └── infra.ts           # Types TypeScript (voir DATA_MODELS.md)
│   ├── App.tsx
│   └── main.tsx               # Import variables.css ici
├── public/
├── vite.config.ts
├── tsconfig.json              # strict: true obligatoire
└── package.json
```

---

## State Management — Zustand

### Store principal

```typescript
// src/store/infra.ts
import { create } from 'zustand'
import type { UnifiedNode, Connection, Alert } from '../types/infra'

interface InfraState {
  // Données
  tree: UnifiedNode | null
  connections: Connection[]
  alerts: Alert[]

  // Navigation
  selectedNodeId: string | null
  expandedNodeIds: Set<string>
  breadcrumb: { id: string; name: string }[]

  // Filtres
  filterNamespace: string
  filterStatus: string
  filterTag: string
  searchQuery: string

  // UI
  isLoading: boolean
  isError: boolean
  errorMessage: string
  lastRefresh: Date | null

  // Actions — Données
  setTree: (tree: UnifiedNode) => void
  setConnections: (connections: Connection[]) => void
  addAlert: (alert: Alert) => void
  removeAlert: (id: string) => void

  // Actions — Navigation
  selectNode: (nodeId: string | null) => void
  expandNode: (nodeId: string) => void
  collapseNode: (nodeId: string) => void
  navigateTo: (node: UnifiedNode) => void

  // Actions — Filtres
  setFilterNamespace: (ns: string) => void
  setFilterStatus: (status: string) => void
  setFilterTag: (tag: string) => void
  setSearchQuery: (q: string) => void

  // Actions — UI
  setLoading: (v: boolean) => void
  setError: (message: string) => void
  clearError: () => void
}

export const useInfraStore = create<InfraState>((set) => ({
  tree: null,
  connections: [],
  alerts: [],
  selectedNodeId: null,
  expandedNodeIds: new Set(),
  breadcrumb: [],
  filterNamespace: '',
  filterStatus: '',
  filterTag: '',
  searchQuery: '',
  isLoading: true,
  isError: false,
  errorMessage: '',
  lastRefresh: null,

  setTree: (tree) => set({ tree, isLoading: false, isError: false, lastRefresh: new Date() }),
  setConnections: (connections) => set({ connections }),
  addAlert: (alert) => set((s) => ({ alerts: [...s.alerts, alert] })),
  removeAlert: (id) => set((s) => ({ alerts: s.alerts.filter((a) => a.id !== id) })),

  selectNode: (nodeId) => set({ selectedNodeId: nodeId }),
  expandNode: (nodeId) => set((s) => ({
    expandedNodeIds: new Set([...s.expandedNodeIds, nodeId])
  })),
  collapseNode: (nodeId) => set((s) => {
    const next = new Set(s.expandedNodeIds)
    next.delete(nodeId)
    return { expandedNodeIds: next }
  }),
  navigateTo: (node) => set((s) => ({
    breadcrumb: [...s.breadcrumb, { id: node.id, name: node.name }],
    expandedNodeIds: new Set([...s.expandedNodeIds, node.id]),
  })),

  setFilterNamespace: (filterNamespace) => set({ filterNamespace }),
  setFilterStatus: (filterStatus) => set({ filterStatus }),
  setFilterTag: (filterTag) => set({ filterTag }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),

  setLoading: (isLoading) => set({ isLoading }),
  setError: (errorMessage) => set({ isError: true, errorMessage, isLoading: false }),
  clearError: () => set({ isError: false, errorMessage: '' }),
}))
```

---

## Pattern Hooks

### Règle fondamentale

> Toute logique de fetch et de transformation se fait dans les hooks.  
> Les composants sont visuellement purs : ils reçoivent des props et rendent.

### `useInfraTree` — Polling

```typescript
// src/hooks/useInfraTree.ts
import { useEffect } from 'react'
import { fetchTree } from '../api/client'
import { useInfraStore } from '../store/infra'

const POLL_INTERVAL = 30_000 // 30s — synchronisé avec le backend

export function useInfraTree() {
  const { filterNamespace, filterStatus, filterTag, setTree, setLoading, setError } = useInfraStore()

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const tree = await fetchTree({ namespace: filterNamespace, status: filterStatus, tag: filterTag })
        if (!cancelled) setTree(tree)
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Unknown error')
      }
    }

    setLoading(true)
    load()
    const interval = setInterval(load, POLL_INTERVAL)
    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [filterNamespace, filterStatus, filterTag]) // Reload si filtres changent
}
```

### `useAlerts` — WebSocket

```typescript
// src/hooks/useAlerts.ts
import { useEffect } from 'react'
import { createAlertWebSocket } from '../api/websocket'
import { useInfraStore } from '../store/infra'

export function useAlerts() {
  const { addAlert, removeAlert } = useInfraStore()

  useEffect(() => {
    const ws = createAlertWebSocket({
      onAlertFired: (alert) => addAlert(alert),
      onAlertResolved: ({ id }) => removeAlert(id),
    })
    return () => ws.close()
  }, [])
}
```

---

## API Client

```typescript
// src/api/client.ts

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

async function apiFetch<T>(path: string, params?: Record<string, string>): Promise<T> {
  const url = new URL(`${BASE_URL}/api/v1${path}`)
  if (params) {
    Object.entries(params).forEach(([k, v]) => { if (v) url.searchParams.set(k, v) })
  }

  const res = await fetch(url.toString(), {
    headers: { 'Content-Type': 'application/json' },
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? `HTTP ${res.status}`)
  }

  return res.json()
}

export const fetchTree = (params?: { namespace?: string; status?: string; tag?: string }) =>
  apiFetch<UnifiedNode>('/tree', params as Record<string, string>)

export const fetchNodeDetail = (nodeId: string) =>
  apiFetch<UnifiedNode>(`/nodes/${nodeId}`)

export const fetchNodeMetrics = (nodeId: string, params?: { from?: string; to?: string; step?: string }) =>
  apiFetch<MetricsResponse>(`/nodes/${nodeId}/metrics`, params as Record<string, string>)

export const fetchNodeLogs = (nodeId: string, params?: { limit?: string; level?: string }) =>
  apiFetch<LogsResponse>(`/nodes/${nodeId}/logs`, params as Record<string, string>)

export const fetchConnections = () =>
  apiFetch<ConnectionsResponse>('/connections')
```

---

## Composants — Conventions

### NodeCard — composant pur

```typescript
// src/views/GridView/NodeCard.tsx

interface NodeCardProps {
  node: UnifiedNode
  isSelected: boolean
  isAlerting: boolean
  onClick: (node: UnifiedNode) => void
}

// ✅ Composant pur — pas d'appel API, pas d'accès au store
export function NodeCard({ node, isSelected, isAlerting, onClick }: NodeCardProps) {
  return (
    <div
      className={cn(
        'node-card',
        `node-card--${node.type}`,
        `node-card--${node.status}`,
        isSelected && 'node-card--selected',
        isAlerting && 'node-card--alerting',
      )}
      onClick={() => onClick(node)}
      role="button"
      tabIndex={0}
      aria-label={`${node.name} — ${node.status}`}
      aria-pressed={isSelected}
    >
      <div className="node-card__header">
        <StatusBadge status={node.status} />
        <span className="node-card__name">{node.name}</span>
      </div>
      {(node.cpu !== undefined || node.memory !== undefined) && (
        <div className="node-card__metrics">
          {node.cpu !== undefined && (
            <MetricValue label="CPU" value={node.cpu} unit="%" />
          )}
          {node.memory !== undefined && node.memoryTotal !== undefined && (
            <MetricValue label="MEM" value={node.memory} unit="MB" />
          )}
        </div>
      )}
    </div>
  )
}
```

### StatusBadge — accessibilité obligatoire

```typescript
// src/components/shared/StatusBadge.tsx

const STATUS_CONFIG = {
  healthy:  { icon: '✓', label: 'Healthy',  color: 'var(--status-healthy)'  },
  warning:  { icon: '⚠', label: 'Warning',  color: 'var(--status-warning)'  },
  critical: { icon: '✕', label: 'Critical', color: 'var(--status-critical)' },
  unknown:  { icon: '?', label: 'Unknown',  color: 'var(--status-unknown)'  },
} as const

export function StatusBadge({ status }: { status: NodeStatus }) {
  const cfg = STATUS_CONFIG[status]
  return (
    <span
      role="status"
      aria-label={cfg.label}
      style={{ color: cfg.color }}
    >
      {cfg.icon}
    </span>
  )
}
```

---

## Conventions TypeScript

```typescript
// ✅ Interfaces pour les objets API
interface UnifiedNode { id: string; name: string; ... }

// ✅ Type pour les unions littérales
type NodeStatus = 'healthy' | 'warning' | 'critical' | 'unknown'

// ✅ unknown pour les erreurs inconnues
} catch (err: unknown) {
  const message = err instanceof Error ? err.message : 'Unknown error'
}

// ❌ Interdit — any
const nodes: any[] = []

// ❌ Interdit — assertion forcée sans vérification
const node = data as UnifiedNode
```

---

## Tests

### Hook test avec Vitest

```typescript
// src/hooks/useInfraTree.test.ts
import { renderHook, act } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '../mocks/server' // MSW setup
import { useInfraTree } from './useInfraTree'

test('charge l\'arbre et met à jour le store', async () => {
  server.use(
    http.get('/api/v1/tree', () => HttpResponse.json({ id: 'root', name: 'Infra', type: 'root' }))
  )

  const { result } = renderHook(() => useInfraTree())
  await act(async () => { await new Promise(r => setTimeout(r, 50)) })

  // Vérifier le store Zustand
  const { tree } = useInfraStore.getState()
  expect(tree?.id).toBe('root')
})
```

### Composant test avec RTL

```typescript
// src/components/shared/StatusBadge.test.tsx
import { render, screen } from '@testing-library/react'
import { StatusBadge } from './StatusBadge'

test.each([
  ['healthy', 'Healthy'],
  ['warning', 'Warning'],
  ['critical', 'Critical'],
])('StatusBadge "%s" a l\'aria-label correct', (status, label) => {
  render(<StatusBadge status={status as NodeStatus} />)
  expect(screen.getByRole('status', { name: label })).toBeInTheDocument()
})
```

---

## Anti-patterns

```typescript
// ❌ Fetch dans un composant
function NodeCard({ nodeId }) {
  const [data, setData] = useState(null)
  useEffect(() => { fetch(`/api/nodes/${nodeId}`).then(...) }, []) // ← INTERDIT
}

// ❌ Accès au store dans un composant de visualisation
function NodeCard({ nodeId }) {
  const { tree } = useInfraStore() // ← INTERDIT — passer les données en props
}

// ❌ CSS hardcodé
<div style={{ color: '#f85149' }}> // ← INTERDIT — utiliser var(--status-critical)

// ❌ any
const handleData = (data: any) => { ... } // ← INTERDIT
```
