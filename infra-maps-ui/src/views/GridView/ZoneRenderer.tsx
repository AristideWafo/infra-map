import type { UnifiedNode } from '../../types/infra'
import { StatusBadge } from '../../components/shared/StatusBadge'
import { NodeCard } from './NodeCard'
import { EmptyState } from '../../components/shared/EmptyState'

interface ZoneRendererProps {
  zone: UnifiedNode
  selectedNodeId: string | null
  alertingIds: ReadonlySet<string>
  onSelect: (node: UnifiedNode) => void
}

// Rend une zone (cluster, vm-island, node K8s) et ses enfants.
// Les zones intermédiaires sont rendues récursivement ; les feuilles en NodeCard.
export function ZoneRenderer({ zone, selectedNodeId, alertingIds, onSelect }: ZoneRendererProps) {
  const children = zone.children ?? []
  const isLeaf = (n: UnifiedNode) => !n.children || n.children.length === 0

  return (
    <section className={`zone zone--${zone.type} zone--${zone.status}`} aria-label={zone.name}>
      <header className="zone__header">
        <StatusBadge status={zone.status} />
        <h2 className="zone__title">{zone.name}</h2>
        {zone.pods !== undefined && <span className="zone__meta">{zone.pods} pods</span>}
      </header>
      <div className="zone__grid">
        {children.length === 0 ? (
          <EmptyState message={`Aucune donnée — vérifier ${zone.source}`} />
        ) : (
          children.map((child) =>
            isLeaf(child) ? (
              <NodeCard
                key={child.id}
                node={child}
                isSelected={selectedNodeId === child.id}
                isAlerting={alertingIds.has(child.id)}
                onClick={onSelect}
              />
            ) : (
              <ZoneRenderer
                key={child.id}
                zone={child}
                selectedNodeId={selectedNodeId}
                alertingIds={alertingIds}
                onSelect={onSelect}
              />
            ),
          )
        )}
      </div>
    </section>
  )
}
