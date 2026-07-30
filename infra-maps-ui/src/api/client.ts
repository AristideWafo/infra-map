import type { UnifiedNode, ConnectionsResponse, HealthStatus } from '../types/infra'

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

export const fetchTree = (params?: { namespace?: string; status?: string; tag?: string }) =>
  apiFetch<UnifiedNode>('/tree', params as Record<string, string>)

export const fetchNodeDetail = (nodeId: string) => apiFetch<UnifiedNode>(`/nodes/${nodeId}`)

export const fetchConnections = () => apiFetch<ConnectionsResponse>('/connections')

export const fetchHealth = () => apiFetch<HealthStatus>('/health')
