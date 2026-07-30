// Formatage des métriques — la mémoire arrive TOUJOURS en MB depuis l'API.

export function formatCPU(cpu: number): string {
  return `${cpu.toFixed(1)}%`
}

export function formatMemory(mb: number): string {
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`
  return `${Math.round(mb)} MB`
}

export function formatPercent(v: number): string {
  return `${v.toFixed(0)}%`
}
