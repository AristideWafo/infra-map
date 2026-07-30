import { render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'
import { ZoneRenderer } from './ZoneRenderer'
import type { UnifiedNode } from '../../types/infra'

const base = {
  status: 'healthy' as const,
  x: 0,
  y: 0,
  z: 0,
  lastSeen: '2026-07-30T10:00:00Z',
}

describe('ZoneRenderer', () => {
  test('affiche un état vide quand la zone n\'a aucun enfant (DESIGN_SYSTEM.md)', () => {
    const zone: UnifiedNode = {
      ...base,
      id: 'cluster-empty',
      name: 'cluster-empty',
      type: 'cluster',
      source: 'kubernetes',
    }

    render(
      <ZoneRenderer
        zone={zone}
        selectedNodeId={null}
        alertingIds={new Set()}
        onSelect={vi.fn()}
      />,
    )

    expect(screen.getByText('Aucune donnée — vérifier kubernetes')).toBeInTheDocument()
  })

  test('rend les nœuds enfants quand présents', () => {
    const zone: UnifiedNode = {
      ...base,
      id: 'cluster-prod',
      name: 'cluster-production',
      type: 'cluster',
      source: 'kubernetes',
      children: [
        { ...base, id: 'pod-1', name: 'pod-1', type: 'pod', source: 'kubernetes' },
      ],
    }

    render(
      <ZoneRenderer
        zone={zone}
        selectedNodeId={null}
        alertingIds={new Set()}
        onSelect={vi.fn()}
      />,
    )

    expect(screen.getByRole('button', { name: 'pod-1 — healthy' })).toBeInTheDocument()
    expect(screen.queryByText(/Aucune donnée/)).not.toBeInTheDocument()
  })
})
