import { useLayoutEffect, useState } from 'react'
import type { Connection, UnifiedNode } from '../../types/infra'

interface Line {
  id: string
  x1: number
  y1: number
  x2: number
  y2: number
}

interface ConnectionSVGProps {
  connections: Connection[]
  containerRef: React.RefObject<HTMLDivElement | null>
  // Recalcule les lignes quand l'arbre change (positions DOM potentiellement différentes)
  tree: UnifiedNode | null
}

// Overlay SVG : relie les cartes présentes dans le DOM (data-node-id).
// Les connexions dont un des deux bouts n'est pas rendu sont ignorées.
export function ConnectionSVG({ connections, containerRef, tree }: ConnectionSVGProps) {
  const [lines, setLines] = useState<Line[]>([])

  useLayoutEffect(() => {
    const container = containerRef.current
    if (!container) return

    function compute() {
      if (!container) return
      const containerRect = container.getBoundingClientRect()
      const centers = new Map<string, { x: number; y: number }>()
      container.querySelectorAll<HTMLElement>('[data-node-id]').forEach((el) => {
        const r = el.getBoundingClientRect()
        centers.set(el.dataset.nodeId ?? '', {
          x: r.left + r.width / 2 - containerRect.left + container.scrollLeft,
          y: r.top + r.height / 2 - containerRect.top + container.scrollTop,
        })
      })

      setLines(
        connections.flatMap((c) => {
          const from = centers.get(c.fromId)
          const to = centers.get(c.toId)
          if (!from || !to) return []
          return [{ id: c.id, x1: from.x, y1: from.y, x2: to.x, y2: to.y }]
        }),
      )
    }

    compute()
    window.addEventListener('resize', compute)
    return () => window.removeEventListener('resize', compute)
  }, [connections, containerRef, tree])

  if (lines.length === 0) return null

  return (
    <svg className="connection-svg" aria-hidden="true">
      <defs>
        <marker id="conn-arrow" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
          <path d="M0,0 L8,3 L0,6 Z" fill="var(--connection-default)" />
        </marker>
      </defs>
      {lines.map((l) => (
        <line
          key={l.id}
          x1={l.x1}
          y1={l.y1}
          x2={l.x2}
          y2={l.y2}
          stroke="var(--connection-default)"
          strokeWidth="1.5"
          strokeDasharray="6 4"
          opacity="0.6"
          markerEnd="url(#conn-arrow)"
        />
      ))}
    </svg>
  )
}
