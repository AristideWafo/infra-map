import type { NodeStatus } from '../../types/infra'
import { STATUS_CONFIG } from '../../utils/status'

export function StatusBadge({ status }: { status: NodeStatus }) {
  const cfg = STATUS_CONFIG[status]
  return (
    <span role="status" aria-label={cfg.label} style={{ color: cfg.cssVar }}>
      {cfg.icon}
    </span>
  )
}
