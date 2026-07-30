import { useState } from 'react'
import type { UnifiedNode } from '../../types/infra'
import { StatusBadge } from '../shared/StatusBadge'
import { formatCPU, formatMemory, formatPercent } from '../../utils/format'
import { statusToLabel } from '../../utils/status'
import { useNodeMetrics } from '../../hooks/useNodeMetrics'
import { useNodeLogs } from '../../hooks/useNodeLogs'
import { MetricsChart } from './MetricsChart'
import { LogViewer } from './LogViewer'

interface SidePanelProps {
  node: UnifiedNode
  onClose: () => void
}

type Tab = 'metrics' | 'logs' | 'details'

export function SidePanel({ node, onClose }: SidePanelProps) {
  const [tab, setTab] = useState<Tab>('metrics')
  const [logLevel, setLogLevel] = useState('')
  const metrics = useNodeMetrics(tab === 'metrics' ? node.id : null)
  const logs = useNodeLogs(tab === 'logs' ? node.id : null, logLevel)

  return (
    <aside className="side-panel" aria-label={`Détails de ${node.name}`}>
      <header className="side-panel__header">
        <StatusBadge status={node.status} />
        <h2 className="side-panel__title">{node.name}</h2>
        <button type="button" className="side-panel__close" onClick={onClose} aria-label="Fermer">
          ✕
        </button>
      </header>

      <nav className="side-panel__tabs" role="tablist">
        {(
          [
            ['metrics', 'Métriques'],
            ['logs', 'Logs'],
            ['details', 'Détails'],
          ] as [Tab, string][]
        ).map(([id, label]) => (
          <button
            key={id}
            type="button"
            role="tab"
            aria-selected={tab === id}
            className={`side-panel__tab ${tab === id ? 'side-panel__tab--active' : ''}`}
            onClick={() => setTab(id)}
          >
            {label}
          </button>
        ))}
      </nav>

      {tab === 'metrics' && (
        <div className="side-panel__content">
          {metrics.isLoading && <p className="side-panel__info">Chargement…</p>}
          {metrics.error && (
            <p className="side-panel__info">Historique indisponible — {metrics.error}</p>
          )}
          {metrics.data?.series?.cpu && (
            <MetricsChart title="CPU" unit="%" points={metrics.data.series.cpu} />
          )}
          {metrics.data?.series?.memory && (
            <MetricsChart title="Mémoire" unit="MB" points={metrics.data.series.memory} />
          )}
          {metrics.data?.series?.disk && (
            <MetricsChart title="Disque" unit="%" points={metrics.data.series.disk} />
          )}
        </div>
      )}

      {tab === 'logs' && (
        <LogViewer
          entries={logs.entries}
          isLoading={logs.isLoading}
          error={logs.error}
          level={logLevel}
          onLevelChange={setLogLevel}
        />
      )}

      {tab === 'details' && (
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
      )}
    </aside>
  )
}
