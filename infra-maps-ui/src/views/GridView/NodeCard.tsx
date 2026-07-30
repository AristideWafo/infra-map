import type { UnifiedNode } from '../../types/infra'
import { StatusBadge } from '../../components/shared/StatusBadge'
import { MetricValue } from '../../components/shared/MetricValue'
import { formatCPU, formatMemory } from '../../utils/format'

interface NodeCardProps {
  node: UnifiedNode
  isSelected: boolean
  onClick: (node: UnifiedNode) => void
}

// Composant pur — pas d'appel API, pas d'accès au store.
export function NodeCard({ node, isSelected, onClick }: NodeCardProps) {
  const classes = [
    'node-card',
    `node-card--${node.type}`,
    `node-card--${node.status}`,
    isSelected ? 'node-card--selected' : '',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div
      className={classes}
      onClick={() => onClick(node)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') onClick(node)
      }}
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
          {node.cpu !== undefined && <MetricValue label="CPU" value={formatCPU(node.cpu)} />}
          {node.memory !== undefined && (
            <MetricValue label="MEM" value={formatMemory(node.memory)} />
          )}
        </div>
      )}
    </div>
  )
}
