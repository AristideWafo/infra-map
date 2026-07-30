import '@testing-library/jest-dom/vitest'

// jsdom n'implémente pas ResizeObserver (requis par Recharts)
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = ResizeObserverStub as unknown as typeof ResizeObserver
}

// Neutraliser WebSocket en test : pas de connexion réseau réelle
class WebSocketStub {
  onopen: (() => void) | null = null
  onmessage: ((e: unknown) => void) | null = null
  onclose: (() => void) | null = null
  close() {}
  send() {}
}

globalThis.WebSocket = WebSocketStub as unknown as typeof WebSocket
