import { render, screen } from '@testing-library/react'
import { describe, expect, test } from 'vitest'
import { StatusBadge } from './StatusBadge'
import type { NodeStatus } from '../../types/infra'

describe('StatusBadge', () => {
  test.each([
    ['healthy', 'Healthy'],
    ['warning', 'Warning'],
    ['critical', 'Critical'],
    ['unknown', 'Unknown'],
  ])('status "%s" expose l\'aria-label "%s"', (status, label) => {
    render(<StatusBadge status={status as NodeStatus} />)
    expect(screen.getByRole('status', { name: label })).toBeInTheDocument()
  })
})
