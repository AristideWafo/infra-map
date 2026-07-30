import { useEffect } from 'react'
import { fetchConnections } from '../api/client'
import { useInfraStore } from '../store/infra'

const POLL_INTERVAL = 30_000

export function useConnections() {
  useEffect(() => {
    let cancelled = false
    const { setConnections } = useInfraStore.getState()

    async function load() {
      try {
        const res = await fetchConnections()
        if (!cancelled) setConnections(res.connections)
      } catch {
        // Connexions non critiques : on garde les dernières connues
      }
    }

    void load()
    const interval = setInterval(() => void load(), POLL_INTERVAL)
    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [])
}
