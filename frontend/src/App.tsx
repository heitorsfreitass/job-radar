import { useEffect, useState } from 'react'
import { searchJobs } from './api'
import FilterBar from './components/FilterBar'
import JobDetail from './components/JobDetail'
import JobList from './components/JobList'
import Pagination from './components/Pagination'
import { EMPTY_FILTERS, type Filters, type Job } from './types'
import './App.css'

const PAGE_SIZE = 20

export default function App() {
  const [filters, setFilters] = useState<Filters>(EMPTY_FILTERS)
  const [page, setPage] = useState(1)
  const [jobs, setJobs] = useState<Job[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedJob, setSelectedJob] = useState<Job | null>(null)

  useEffect(() => {
    setPage(1)
  }, [filters])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    const debounce = setTimeout(() => {
      searchJobs(filters, page, PAGE_SIZE)
        .then((res) => {
          if (cancelled) return
          setJobs(res.data)
          setTotal(res.meta.total)
        })
        .catch((err: Error) => {
          if (cancelled) return
          setError(err.message)
        })
        .finally(() => {
          if (!cancelled) setLoading(false)
        })
    }, 300)

    return () => {
      cancelled = true
      clearTimeout(debounce)
    }
  }, [filters, page])

  return (
    <div className="app">
      <header className="app-header">
        <h1>job-radar</h1>
        <p>European job postings, aggregated from Arbeitnow and Remotive.</p>
      </header>

      <FilterBar filters={filters} onChange={setFilters} />

      <JobList jobs={jobs} loading={loading} error={error} onSelect={setSelectedJob} />

      <Pagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />

      {selectedJob && <JobDetail job={selectedJob} onClose={() => setSelectedJob(null)} />}
    </div>
  )
}
