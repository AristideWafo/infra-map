interface FilterBarProps {
  namespaces: string[]
  tags: string[]
  filterNamespace: string
  filterStatus: string
  filterTag: string
  onNamespaceChange: (v: string) => void
  onStatusChange: (v: string) => void
  onTagChange: (v: string) => void
}

const STATUSES = ['healthy', 'warning', 'critical', 'unknown']

// Composant pur — les valeurs et callbacks viennent du parent.
export function FilterBar({
  namespaces,
  tags,
  filterNamespace,
  filterStatus,
  filterTag,
  onNamespaceChange,
  onStatusChange,
  onTagChange,
}: FilterBarProps) {
  return (
    <div className="filter-bar">
      <label className="filter-bar__field">
        <span className="filter-bar__label">Namespace</span>
        <select
          value={filterNamespace}
          onChange={(e) => onNamespaceChange(e.target.value)}
          aria-label="Filtrer par namespace"
        >
          <option value="">Tous</option>
          {namespaces.map((ns) => (
            <option key={ns} value={ns}>
              {ns}
            </option>
          ))}
        </select>
      </label>
      <label className="filter-bar__field">
        <span className="filter-bar__label">Statut</span>
        <select
          value={filterStatus}
          onChange={(e) => onStatusChange(e.target.value)}
          aria-label="Filtrer par statut"
        >
          <option value="">Tous</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
      </label>
      <label className="filter-bar__field">
        <span className="filter-bar__label">Tag</span>
        <select
          value={filterTag}
          onChange={(e) => onTagChange(e.target.value)}
          aria-label="Filtrer par tag"
        >
          <option value="">Tous</option>
          {tags.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </label>
    </div>
  )
}
