import type { NodeStatus } from '../types/infra'

export const STATUS_CONFIG: Record<
  NodeStatus,
  { icon: string; label: string; cssVar: string }
> = {
  healthy: { icon: '✓', label: 'Healthy', cssVar: 'var(--status-healthy)' },
  warning: { icon: '⚠', label: 'Warning', cssVar: 'var(--status-warning)' },
  critical: { icon: '✕', label: 'Critical', cssVar: 'var(--status-critical)' },
  unknown: { icon: '?', label: 'Unknown', cssVar: 'var(--status-unknown)' },
}

export const statusToColor = (s: NodeStatus): string => STATUS_CONFIG[s].cssVar
export const statusToIcon = (s: NodeStatus): string => STATUS_CONFIG[s].icon
export const statusToLabel = (s: NodeStatus): string => STATUS_CONFIG[s].label
