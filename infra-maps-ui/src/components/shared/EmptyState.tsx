interface EmptyStateProps {
  message: string
}

export function EmptyState({ message }: EmptyStateProps) {
  return (
    <div className="empty-state" role="status">
      <span aria-hidden="true">⚠</span>
      <p>{message}</p>
    </div>
  )
}
