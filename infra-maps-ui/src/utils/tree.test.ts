import { describe, expect, test } from 'vitest'
import { findNodeById, flattenTree, getAncestors } from './tree'
import type { UnifiedNode } from '../types/infra'

const base = { type: 'pod', status: 'healthy', x: 0, y: 0, z: 0, source: 'test', lastSeen: '' } as const

const tree: UnifiedNode = {
  ...base,
  id: 'root',
  name: 'Infrastructure',
  type: 'root',
  children: [
    {
      ...base,
      id: 'cluster-prod',
      name: 'cluster',
      type: 'cluster',
      children: [{ ...base, id: 'pod-1', name: 'pod-1' }],
    },
    { ...base, id: 'vm-island-dc1', name: 'dc1', type: 'vm-island' },
  ],
}

describe('tree utils', () => {
  test('findNodeById trouve un nœud profond', () => {
    expect(findNodeById(tree, 'pod-1')?.name).toBe('pod-1')
  })

  test('findNodeById retourne null si absent ou arbre null', () => {
    expect(findNodeById(tree, 'nope')).toBeNull()
    expect(findNodeById(null, 'pod-1')).toBeNull()
  })

  test('flattenTree retourne tous les nœuds', () => {
    expect(flattenTree(tree).map((n) => n.id)).toEqual([
      'root',
      'cluster-prod',
      'pod-1',
      'vm-island-dc1',
    ])
  })

  test('getAncestors retourne le chemin racine → nœud', () => {
    expect(getAncestors(tree, 'pod-1').map((n) => n.id)).toEqual([
      'root',
      'cluster-prod',
      'pod-1',
    ])
  })

  test('getAncestors retourne [] si nœud absent', () => {
    expect(getAncestors(tree, 'nope')).toEqual([])
  })
})
