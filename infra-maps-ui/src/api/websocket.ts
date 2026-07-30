import type { Alert } from '../types/infra'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

interface AlertWSHandlers {
  onAlertFired: (alert: Alert) => void
  onAlertResolved: (payload: { id: string }) => void
}

export interface AlertWS {
  close: () => void
}

// WebSocket alertes avec reconnexion en backoff exponentiel (1s → 30s max).
export function createAlertWebSocket(handlers: AlertWSHandlers): AlertWS {
  // Environnements sans WebSocket (tests jsdom) : no-op
  if (typeof WebSocket === 'undefined') {
    return { close: () => {} }
  }
  let ws: WebSocket | null = null
  let closed = false
  let retryMs = 1_000
  let timer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    if (closed) return
    const url = BASE_URL.replace(/^http/, 'ws') + '/api/v1/ws/alerts'
    ws = new WebSocket(url)

    ws.onopen = () => {
      retryMs = 1_000
    }

    ws.onmessage = (event: MessageEvent<string>) => {
      try {
        const msg = JSON.parse(event.data) as { type: string; payload?: unknown }
        if (msg.type === 'alert_fired' && msg.payload) {
          handlers.onAlertFired(msg.payload as Alert)
        } else if (msg.type === 'alert_resolved' && msg.payload) {
          handlers.onAlertResolved(msg.payload as { id: string })
        }
      } catch {
        // message non JSON — ignoré
      }
    }

    ws.onclose = () => {
      if (closed) return
      timer = setTimeout(connect, retryMs)
      retryMs = Math.min(retryMs * 2, 30_000)
    }
  }

  connect()

  return {
    close: () => {
      closed = true
      if (timer) clearTimeout(timer)
      ws?.close()
    },
  }
}
