import type { UnifiedNode } from '../../types/infra'
import { StatusBadge } from '../shared/StatusBadge'
import { formatCPU, formatMemory, formatPercent } from '../../utils/format'
import { statusToLabel } from '../../utils/status'

interface SidePanelProps {
  node: UnifiedNode
  onClose: () => void
}

// Phase 0 : métriques instantanées du nœud sélectionné.
// L'historique Recharts (proxy Prometheus) arrive en Phase 2.
export function SidePanel({ node, onClose }: SidePanelProps) {
  const memPct =
    node.memory !== undefined && node.memoryTotal ? (node.memory / node.memoryTotal) * 100 : null

  return (
    <aside className="side-panel" aria-label={`Détails de ${node.name}`}>
      <header className="side-panel__header">
        <StatusBadge status={node.status} />
        <h2 className="side-panel__title">{node.name}</h2>
        <button type="button" className="side-panel__close" onClick={onClose} aria-label="Fermer">
          ✕
        </button>
      </header>

      <dl className="side-panel__details">
        <div className="side-panel__row">
          <dt>Statut</dt>
          <dd>{statusToLabel(node.status)}</dd>
        </div>
        <div className="side-panel__row">
          <dt>Type</dt>
          <dd className="mono">{node.type}</dd>
        </div>
        <div className="side-panel__row">
          <dt>ID</dt>
          <dd className="mono">{node.id}</dd>
        </div>
        <div className="side-panel__row">
          <dt>Source</dt>
          <dd className="mono">{node.source}</dd>
        </div>
        {node.namespace && (
          <div className="side-panel__row">
            <dt>Namespace</dt>
            <dd className="mono">{node.namespace}</dd>
          </div>
        )}
        {node.cpu !== undefined && (
          <div className="side-panel__row">
            <dt>CPU</dt>
            <dd className="mono">{formatCPU(node.cpu)}</dd>
          </div>
        )}
        {node.memory !== undefined && node.memoryTotal !== undefined && (
          <div className="side-panel__row">
            <dt>Mémoire</dt>
            <dd className="mono">
              {formatMemory(node.memory)} / {formatMemory(node.memoryTotal)}
              {memPct !== null && ` (${formatPercent(memPct)})`}
            </dd>
          </div>
        )}
        {node.disk !== undefined && (
          <div className="side-panel__row">
            <dt>Disque</dt>
            <dd className="mono">{formatPercent(node.disk)}</dd>
          </div>
        )}
        {node.restarts !== undefined && (
          <div className="side-panel__row">
            <dt>Restarts</dt>
            <dd className="mono">{node.restarts}</dd>
          </div>
        )}
        {node.tags && node.tags.length > 0 && (
          <div className="side-panel__row">
            <dt>Tags</dt>
            <dd>
              {node.tags.map((t) => (
                <span key={t} className="tag">
                  {t}
                </span>
              ))}
            </dd>
          </div>
        )}
        <div className="side-panel__row">
          <dt>Vu à</dt>
          <dd className="mono">{new Date(node.lastSeen).toLocaleTimeString()}</dd>
        </div>
      </dl>
    </aside>
  )
}
