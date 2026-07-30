interface MetricValueProps {
  label: string
  value: string
}

export function MetricValue({ label, value }: MetricValueProps) {
  return (
    <span className="metric-value">
      <span className="metric-value__label">{label}</span>
      <span className="metric-value__num">{value}</span>
    </span>
  )
}
