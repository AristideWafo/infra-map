import type {
  UnifiedNode,
  ConnectionsResponse,
  HealthStatus,
  MetricsResponse,
  LogsResponse,
} from '../types/infra'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

async function apiFetch<T>(path: string, params?: Record<string, string>): Promise<T> {
  const url = new URL(`${BASE_URL}/api/v1${path}`)
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      if (v) url.searchParams.set(k, v)
    })
  }

  const res = await fetch(url.toString(), {
    headers: { 'Content-Type': 'application/json' },
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(typeof body.error === 'string' ? body.error : `HTTP ${res.status}`)
  }

  return res.json() as Promise<T>
}

export interface TreeResult {
  tree: UnifiedNode
  cacheAgeSeconds: number | null
}

// fetchTree lit X-Cache-Age-Seconds : le backend sert des données périmées en
// 200 (dégradation gracieuse) quand un scraper est down — sans ce header, le
// front ne verrait jamais qu'un arbre affiché est en fait obsolète.
export async function fetchTree(params?: {
  namespace?: string
  status?: string
  tag?: string
}): Promise<TreeResult> {
  const url = new URL(`${BASE_URL}/api/v1/tree`)
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      if (v) url.searchParams.set(k, v)
    })
  }

  const res = await fetch(url.toString(), { headers: { 'Content-Type': 'application/json' } })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(typeof body.error === 'string' ? body.error : `HTTP ${res.status}`)
  }

  const raw = res.headers.get('X-Cache-Age-Seconds')
  return {
    tree: (await res.json()) as UnifiedNode,
    cacheAgeSeconds: raw !== null ? Number(raw) : null,
  }
}

export const fetchNodeDetail = (nodeId: string) => apiFetch<UnifiedNode>(`/nodes/${nodeId}`)

export const fetchConnections = () => apiFetch<ConnectionsResponse>('/connections')

export const fetchHealth = () => apiFetch<HealthStatus>('/health')

export const fetchNodeMetrics = (nodeId: string, params?: { from?: string; to?: string; step?: string; metric?: string }) =>
  apiFetch<MetricsResponse>(`/nodes/${nodeId}/metrics`, params as Record<string, string>)

export const fetchNodeLogs = (nodeId: string, params?: { limit?: string; level?: string }) =>
  apiFetch<LogsResponse>(`/nodes/${nodeId}/logs`, params as Record<string, string>)
