import { useEffect, useRef } from 'react'
import type { LogEntry } from '../../types/infra'

interface LogViewerProps {
  entries: LogEntry[]
  isLoading: boolean
  error: string
  level: string
  onLevelChange: (level: string) => void
}

const LEVELS = ['error', 'warn', 'info']

export function LogViewer({ entries, isLoading, error, level, onLevelChange }: LogViewerProps) {
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: 'end' })
  }, [entries])

  return (
    <div className="log-viewer">
      <div className="log-viewer__controls">
        <select
          value={level}
          onChange={(e) => onLevelChange(e.target.value)}
          aria-label="Filtrer par niveau"
        >
          <option value="">Tous niveaux</option>
          {LEVELS.map((l) => (
            <option key={l} value={l}>
              {l}
            </option>
          ))}
        </select>
      </div>

      <div className="log-viewer__list">
        {isLoading && <p className="log-viewer__info">Chargement…</p>}
        {error && <p className="log-viewer__info">Logs indisponibles — {error}</p>}
        {!isLoading && !error && entries.length === 0 && (
          <p className="log-viewer__info">Aucun log sur la période</p>
        )}
        {entries.map((e, i) => (
          <div key={i} className={`log-line log-line--${e.level}`}>
            <span className="log-line__time">
              {new Date(e.timestamp).toLocaleTimeString()}
            </span>
            <span className="log-line__level">{e.level.toUpperCase()}</span>
            <span className="log-line__msg">{e.message}</span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
