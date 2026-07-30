import { useEffect, useState } from 'react'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export function useTags(): string[] {
  const [tags, setTags] = useState<string[]>([])

  useEffect(() => {
    let cancelled = false
    fetch(`${BASE_URL}/api/v1/tags`)
      .then((r) => (r.ok ? r.json() : { tags: [] }))
      .then((body: { tags: string[] }) => {
        if (!cancelled) setTags(body.tags ?? [])
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [])

  return tags
}
