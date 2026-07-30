import { useEffect, useState } from 'react'
import { fetchNodeMetrics } from '../api/client'
import type { MetricsResponse } from '../types/infra'

interface NodeMetricsState {
  data: MetricsResponse | null
  isLoading: boolean
  error: string
}

export function useNodeMetrics(nodeId: string | null): NodeMetricsState {
  const [state, setState] = useState<NodeMetricsState>({ data: null, isLoading: false, error: '' })

  useEffect(() => {
    if (!nodeId) return
    let cancelled = false
    setState({ data: null, isLoading: true, error: '' })

    fetchNodeMetrics(nodeId, { step: '60s' })
      .then((data) => {
        if (!cancelled) setState({ data, isLoading: false, error: '' })
      })
      .catch((err: unknown) => {
        if (!cancelled)
          setState({
            data: null,
            isLoading: false,
            error: err instanceof Error ? err.message : 'Unknown error',
          })
      })
    return () => {
      cancelled = true
    }
  }, [nodeId])

  return state
}
