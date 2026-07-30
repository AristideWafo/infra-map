import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, test, vi } from 'vitest'
import { FilterBar } from './FilterBar'

const noop = () => {}

describe('FilterBar', () => {
  test('liste les namespaces et tags fournis', () => {
    render(
      <FilterBar
        namespaces={['default', 'jobs']}
        tags={['prod', 'db']}
        filterNamespace=""
        filterStatus=""
        filterTag=""
        onNamespaceChange={noop}
        onStatusChange={noop}
        onTagChange={noop}
      />,
    )

    expect(screen.getByRole('option', { name: 'jobs' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'db' })).toBeInTheDocument()
  })

  test('sélection namespace remonte la valeur', async () => {
    const user = userEvent.setup()
    const onNamespaceChange = vi.fn()
    render(
      <FilterBar
        namespaces={['default', 'jobs']}
        tags={[]}
        filterNamespace=""
        filterStatus=""
        filterTag=""
        onNamespaceChange={onNamespaceChange}
        onStatusChange={noop}
        onTagChange={noop}
      />,
    )

    await user.selectOptions(screen.getByRole('combobox', { name: 'Filtrer par namespace' }), 'jobs')
    expect(onNamespaceChange).toHaveBeenCalledWith('jobs')
  })

  test('sélection statut remonte la valeur', async () => {
    const user = userEvent.setup()
    const onStatusChange = vi.fn()
    render(
      <FilterBar
        namespaces={[]}
        tags={[]}
        filterNamespace=""
        filterStatus=""
        filterTag=""
        onNamespaceChange={noop}
        onStatusChange={onStatusChange}
        onTagChange={noop}
      />,
    )

    await user.selectOptions(screen.getByRole('combobox', { name: 'Filtrer par statut' }), 'critical')
    expect(onStatusChange).toHaveBeenCalledWith('critical')
  })
})
