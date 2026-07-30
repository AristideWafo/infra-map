import { useState } from 'react'
import type { UnifiedNode } from '../../types/infra'
import { StatusBadge } from '../shared/StatusBadge'

interface SearchInputProps {
  nodes: UnifiedNode[]
  onSelect: (nodeId: string) => void
}

const MAX_RESULTS = 8

export function SearchInput({ nodes, onSelect }: SearchInputProps) {
  const [query, setQuery] = useState('')

  const q = query.trim().toLowerCase()
  const results = q
    ? nodes
        .filter(
          (n) =>
            n.name.toLowerCase().includes(q) ||
            (n.namespace ?? '').toLowerCase().includes(q),
        )
        .slice(0, MAX_RESULTS)
    : []

  return (
    <div className="search">
      <input
        type="search"
        className="search__input"
        placeholder="Rechercher…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        aria-label="Rechercher un nœud"
      />
      {results.length > 0 && (
        <ul className="search__results" role="listbox">
          {results.map((n) => (
            <li key={n.id}>
              <button
                type="button"
                className="search__result"
                onClick={() => {
                  onSelect(n.id)
                  setQuery('')
                }}
              >
                <StatusBadge status={n.status} />
                <span className="search__name">{n.name}</span>
                {n.namespace && <span className="search__ns">{n.namespace}</span>}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
