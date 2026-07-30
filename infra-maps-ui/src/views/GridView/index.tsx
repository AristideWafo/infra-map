import type { UnifiedNode } from '../../types/infra'
import { EmptyState } from '../../components/shared/EmptyState'
import { LoadingSkeleton } from '../../components/shared/LoadingSkeleton'
import { ZoneRenderer } from './ZoneRenderer'

interface GridViewProps {
  tree: UnifiedNode | null
  isLoading: boolean
  isError: boolean
  errorMessage: string
  selectedNodeId: string | null
  onSelect: (node: UnifiedNode) => void
}

export function GridView({
  tree,
  isLoading,
  isError,
  errorMessage,
  selectedNodeId,
  onSelect,
}: GridViewProps) {
  if (isLoading && !tree) return <LoadingSkeleton />
  if (isError && !tree) return <EmptyState message={`Aucune donnée — ${errorMessage}`} />
  if (!tree || !tree.children?.length)
    return <EmptyState message="Aucune donnée — vérifier les sources" />

  return (
    <div className="grid-view">
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
