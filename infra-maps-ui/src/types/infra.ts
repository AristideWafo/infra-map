// Types miroir des structs Go — voir docs/DATA_MODELS.md (Go fait foi).

export type NodeType =
  | 'root'
  | 'vm-island'
  | 'compose-group'
  | 'cluster'
  | 'node'
  | 'pod'
  | 'vm'
  | 'container'

export type NodeStatus = 'healthy' | 'warning' | 'critical' | 'unknown'

export interface UnifiedNode {
  id: string
  name: string
  type: NodeType
  status: NodeStatus
  tags?: string[]

  cpu?: number
  memory?: number
  memoryTotal?: number
  disk?: number
  pods?: number
  restarts?: number

  x: number
  y: number
  z: number

  parentId?: string
  children?: UnifiedNode[]

  source: string
  rawLabels?: Record<string, string>

  lastSeen: string
  createdAt?: string

  namespace?: string
}

export interface Connection {
  id: string
  fromId: string
  toId: string
  traffic: number
  latency: number
  errors: number
  protocol: string
  type: string
}

export interface Alert {
  id: string
  nodeId: string
  name: string
  severity: 'warning' | 'critical'
  message: string
  firedAt: string
}

export interface HealthStatus {
  status: 'ok' | 'degraded' | 'error'
  scrapers: Record<string, { status: 'ok' | 'disabled' | 'error'; message?: string }>
  cacheAgeSeconds: number
}

export interface ApiError {
  error: string
  code: string
}

export interface ConnectionsResponse {
  connections: Connection[]
  total: number
}

export interface MetricPoint {
  timestamp: string
  value: number
}

export interface MetricsResponse {
  nodeId: string
  from: string
  to: string
  step: string
  series: Record<string, MetricPoint[]>
}

export interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  pod?: string
  container?: string
  node?: string
}

export interface LogsResponse {
  nodeId: string
  entries: LogEntry[]
}
