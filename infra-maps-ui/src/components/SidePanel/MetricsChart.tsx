import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type { MetricPoint } from '../../types/infra'

interface MetricsChartProps {
  title: string
  unit: string
  points: MetricPoint[]
}

export function MetricsChart({ title, unit, points }: MetricsChartProps) {
  const data = points.map((p) => ({
    time: new Date(p.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    value: Math.round(p.value * 10) / 10,
  }))

  return (
    <div className="metrics-chart">
      <h3 className="metrics-chart__title">{title}</h3>
      <ResponsiveContainer width="100%" height={120}>
        <LineChart data={data} margin={{ top: 4, right: 8, bottom: 0, left: -16 }}>
          <XAxis
            dataKey="time"
            tick={{ fill: 'var(--text-secondary)', fontSize: 10 }}
            stroke="var(--border-default)"
            minTickGap={40}
          />
          <YAxis
            tick={{ fill: 'var(--text-secondary)', fontSize: 10 }}
            stroke="var(--border-default)"
            unit={unit}
          />
          <Tooltip
            contentStyle={{
              background: 'var(--bg-elevated)',
              border: '1px solid var(--border-default)',
              borderRadius: 6,
              fontSize: 11,
            }}
            labelStyle={{ color: 'var(--text-secondary)' }}
          />
          <Line
            type="monotone"
            dataKey="value"
            stroke="var(--accent)"
            strokeWidth={1.5}
            dot={false}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
