import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, test, vi } from 'vitest'
import App from './App'
import { useInfraStore } from './store/infra'
import type { UnifiedNode } from './types/infra'

const mockTree: UnifiedNode = {
  id: 'root',
  name: 'Infrastructure',
  type: 'root',
  status: 'warning',
  x: 0,
  y: 0,
  z: 0,
  source: 'internal',
  lastSeen: '2026-07-30T10:00:00Z',
  children: [
    {
      id: 'cluster-prod',
      name: 'cluster-production',
      type: 'cluster',
      status: 'warning',
      x: 0,
      y: 0,
      z: 0,
      source: 'prometheus',
      lastSeen: '2026-07-30T10:00:00Z',
      children: [
        {
          id: 'pod-api-1',
          name: 'api-frontend-1',
          type: 'pod',
          status: 'healthy',
          cpu: 42.5,
          memory: 256,
          memoryTotal: 512,
          x: 0,
          y: 0,
          z: 0,
          source: 'prometheus',
          lastSeen: '2026-07-30T10:00:00Z',
        },
      ],
    },
  ],
}

beforeEach(() => {
  useInfraStore.setState({
    tree: null,
    selectedNodeId: null,
    isLoading: true,
    isError: false,
    errorMessage: '',
    lastRefresh: null,
    cacheAgeSeconds: null,
  })
  vi.stubGlobal(
    'fetch',
    vi.fn(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockTree), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    ),
  )
})

describe('App', () => {
  test("charge l'arbre et affiche la zone cluster avec ses pods", async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('cluster-production')).toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: 'api-frontend-1 — healthy' })).toBeInTheDocument()
  })

  test('clic sur un pod ouvre le side panel, ✕ le ferme', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => screen.getByRole('button', { name: 'api-frontend-1 — healthy' }))
    await user.click(screen.getByRole('button', { name: 'api-frontend-1 — healthy' }))

    expect(screen.getByRole('complementary', { name: 'Détails de api-frontend-1' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Fermer' }))
    expect(screen.queryByRole('complementary')).not.toBeInTheDocument()
  })

  test('affiche un état d\'erreur si l\'API est down', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.reject(new Error('Failed to fetch'))),
    )
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText(/Aucune donnée — Failed to fetch/)).toBeInTheDocument()
    })
  })

  test('signale les données périmées via X-Cache-Age-Seconds (200 OK)', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() =>
        Promise.resolve(
          new Response(JSON.stringify(mockTree), {
            status: 200,
            headers: { 'Content-Type': 'application/json', 'X-Cache-Age-Seconds': '87' },
          }),
        ),
      ),
    )
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText(/Source indisponible — données périmées \(87s\)/)).toBeInTheDocument()
    })
  })

  test('pas de bandeau périmé quand le cache est frais', async () => {
    render(<App />)

    await waitFor(() => screen.getByText('cluster-production'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
