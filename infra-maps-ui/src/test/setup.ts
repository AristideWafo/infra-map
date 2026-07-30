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
