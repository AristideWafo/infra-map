import { useEffect, useState } from 'react'
import { fetchNodeLogs } from '../api/client'
import type { LogEntry } from '../types/infra'

interface NodeLogsState {
  entries: LogEntry[]
  isLoading: boolean
  error: string
}

export function useNodeLogs(nodeId: string | null, level: string): NodeLogsState {
  const [state, setState] = useState<NodeLogsState>({ entries: [], isLoading: false, error: '' })

  useEffect(() => {
    if (!nodeId) return
    let cancelled = false
    setState({ entries: [], isLoading: true, error: '' })

    fetchNodeLogs(nodeId, level ? { level } : undefined)
      .then((res) => {
        if (!cancelled) setState({ entries: res.entries, isLoading: false, error: '' })
      })
      .catch((err: unknown) => {
        if (!cancelled)
          setState({
            entries: [],
            isLoading: false,
            error: err instanceof Error ? err.message : 'Unknown error',
          })
      })
    return () => {
      cancelled = true
    }
  }, [nodeId, level])

  return state
}
