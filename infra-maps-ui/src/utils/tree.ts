import type { UnifiedNode } from '../types/infra'

export function findNodeById(root: UnifiedNode | null, id: string): UnifiedNode | null {
  if (!root) return null
  if (root.id === id) return root
  for (const child of root.children ?? []) {
    const found = findNodeById(child, id)
    if (found) return found
  }
  return null
}

export function flattenTree(root: UnifiedNode | null): UnifiedNode[] {
  if (!root) return []
  return [root, ...(root.children ?? []).flatMap(flattenTree)]
}

export function getAncestors(root: UnifiedNode | null, id: string): UnifiedNode[] {
  if (!root) return []
  if (root.id === id) return [root]
  for (const child of root.children ?? []) {
    const path = getAncestors(child, id)
    if (path.length > 0) return [root, ...path]
  }
  return []
}
