import { describe, expect, test } from 'vitest'
import { formatCPU, formatMemory, formatPercent } from './format'

describe('format', () => {
  test('formatCPU garde une décimale', () => {
    expect(formatCPU(42.53)).toBe('42.5%')
    expect(formatCPU(0)).toBe('0.0%')
  })

  test('formatMemory affiche en MB sous 1024', () => {
    expect(formatMemory(256)).toBe('256 MB')
    expect(formatMemory(1023.4)).toBe('1023 MB')
  })

  test('formatMemory convertit en GB à partir de 1024 MB', () => {
    expect(formatMemory(1024)).toBe('1.0 GB')
    expect(formatMemory(8192)).toBe('8.0 GB')
  })

  test('formatPercent arrondit sans décimale', () => {
    expect(formatPercent(84.6)).toBe('85%')
  })
})
