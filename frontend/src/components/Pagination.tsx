interface Props {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
}

export default function Pagination({ page, pageSize, total, onPageChange }: Props) {
  const lastPage = Math.max(1, Math.ceil(total / pageSize))
  if (lastPage <= 1) return null

  return (
    <div className="pagination">
      <button disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
        ← Prev
      </button>
      <span>
        Page {page} of {lastPage} ({total} jobs)
      </span>
      <button disabled={page >= lastPage} onClick={() => onPageChange(page + 1)}>
        Next →
      </button>
    </div>
  )
}
