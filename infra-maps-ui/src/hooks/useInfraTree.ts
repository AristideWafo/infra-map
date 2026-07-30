import { useEffect } from 'react'
import { fetchTree } from '../api/client'
import { useInfraStore } from '../store/infra'

const POLL_INTERVAL = 30_000 // synchronisé avec le SCRAPE_INTERVAL backend

export function useInfraTree() {
  const filterNamespace = useInfraStore((s) => s.filterNamespace)
  const filterStatus = useInfraStore((s) => s.filterStatus)
  const filterTag = useInfraStore((s) => s.filterTag)

  useEffect(() => {
    let cancelled = false
    const { setTree, setLoading, setError } = useInfraStore.getState()

    async function load() {
      try {
        const { tree, cacheAgeSeconds } = await fetchTree({
          namespace: filterNamespace,
          status: filterStatus,
          tag: filterTag,
        })
        if (!cancelled) setTree(tree, cacheAgeSeconds)
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Unknown error')
      }
    }

    setLoading(true)
    void load()
    const interval = setInterval(() => void load(), POLL_INTERVAL)
    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [filterNamespace, filterStatus, filterTag])
}
