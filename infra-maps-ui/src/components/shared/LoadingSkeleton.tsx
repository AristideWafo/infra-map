export function LoadingSkeleton() {
  return (
    <div className="loading-skeleton" role="status" aria-label="Chargement">
      {[0, 1, 2].map((i) => (
        <div key={i} className="loading-skeleton__block" />
      ))}
    </div>
  )
}
