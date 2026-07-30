import type { UnifiedNode } from '../../types/infra'

interface BreadcrumbProps {
  path: UnifiedNode[]
  onNavigate: (nodeId: string) => void
}

// Chemin racine → nœud sélectionné. Cliquer un ancêtre le sélectionne.
export function Breadcrumb({ path, onNavigate }: BreadcrumbProps) {
  if (path.length === 0) return null
  return (
    <nav className="breadcrumb" aria-label="Fil d'Ariane">
      {path.map((node, i) => (
        <span key={node.id} className="breadcrumb__item">
          {i > 0 && <span className="breadcrumb__sep">›</span>}
          {i === path.length - 1 ? (
            <span className="breadcrumb__current" aria-current="location">
              {node.name}
            </span>
          ) : (
            <button
              type="button"
              className="breadcrumb__link"
              onClick={() => onNavigate(node.id)}
            >
              {node.name}
            </button>
          )}
        </span>
      ))}
    </nav>
  )
}
