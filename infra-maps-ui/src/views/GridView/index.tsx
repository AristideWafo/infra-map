import { useRef } from 'react'
import type { Connection, UnifiedNode } from '../../types/infra'
import { EmptyState } from '../../components/shared/EmptyState'
import { LoadingSkeleton } from '../../components/shared/LoadingSkeleton'
import { ZoneRenderer } from './ZoneRenderer'
import { ConnectionSVG } from './ConnectionSVG'

interface GridViewProps {
  tree: UnifiedNode | null
  connections: Connection[]
  isLoading: boolean
  isError: boolean
  errorMessage: string
  selectedNodeId: string | null
  onSelect: (node: UnifiedNode) => void
}

export function GridView({
  tree,
  connections,
  isLoading,
  isError,
  errorMessage,
  selectedNodeId,
  onSelect,
}: GridViewProps) {
  const containerRef = useRef<HTMLDivElement>(null)

  if (isLoading && !tree) return <LoadingSkeleton />
  if (isError && !tree) return <EmptyState message={`Aucune donnée — ${errorMessage}`} />
  if (!tree || !tree.children?.length)
    return <EmptyState message="Aucune donnée — vérifier les sources" />

  return (
    <div className="grid-view" ref={containerRef}>
      <ConnectionSVG connections={connections} containerRef={containerRef} tree={tree} />
      {tree.children.map((zone) => (
        <ZoneRenderer
          key={zone.id}
          zone={zone}
          selectedNodeId={selectedNodeId}
          onSelect={onSelect}
        />
      ))}
    </div>
  )
}
