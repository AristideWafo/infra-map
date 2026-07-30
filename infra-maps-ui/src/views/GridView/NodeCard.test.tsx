import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, test, vi } from 'vitest'
import { NodeCard } from './NodeCard'
import type { UnifiedNode } from '../../types/infra'

const node: UnifiedNode = {
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
}

describe('NodeCard', () => {
  test('affiche nom, statut et métriques', () => {
    render(<NodeCard node={node} isSelected={false} isAlerting={false} onClick={() => {}} />)

    expect(screen.getByRole('button', { name: 'api-frontend-1 — healthy' })).toBeInTheDocument()
    expect(screen.getByText('42.5%')).toBeInTheDocument()
    expect(screen.getByText('256 MB')).toBeInTheDocument()
  })

  test('clic remonte le nœud', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()
    render(<NodeCard node={node} isSelected={false} isAlerting={false} onClick={onClick} />)

    await user.click(screen.getByRole('button'))
    expect(onClick).toHaveBeenCalledWith(node)
  })

  test('sélection reflétée via aria-pressed', () => {
    render(<NodeCard node={node} isSelected={true} isAlerting={false} onClick={() => {}} />)
    expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'true')
  })

  test("masque le bloc métriques quand il n'y en a pas", () => {
    const bare = { ...node, cpu: undefined, memory: undefined }
    render(<NodeCard node={bare} isSelected={false} isAlerting={false} onClick={() => {}} />)
    expect(screen.queryByText('CPU')).not.toBeInTheDocument()
  })
})
