import { useEffect, useState } from 'react'
import { searchJobs } from './api'
import FilterBar from './components/FilterBar'
import JobDetail from './components/JobDetail'
import JobList from './components/JobList'
import Pagination from './components/Pagination'
import {
  loadHiddenIds,
  loadJSON,
  loadSavedJobs,
  saveJSON,
  toggleHidden,
  toggleSaved,
} from './storage'
import { EMPTY_FILTERS, type Filters, type Job } from './types'
import './App.css'

const PAGE_SIZE = 20

type View = 'all' | 'saved'

export default function App() {
  const [filters, setFilters] = useState<Filters>(() => loadJSON('filters', EMPTY_FILTERS))
  const [page, setPage] = useState(1)
  const [jobs, setJobs] = useState<Job[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedJob, setSelectedJob] = useState<Job | null>(null)
  const [view, setView] = useState<View>('all')
  const [hiddenIds, setHiddenIds] = useState<Set<string>>(loadHiddenIds)
  const [savedJobs, setSavedJobs] = useState<Record<string, Job>>(loadSavedJobs)

  useEffect(() => {
    saveJSON('filters', filters)
    setPage(1)
  }, [filters])

  useEffect(() => {
    if (view !== 'all') return

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
  }, [filters, page, view])

  const handleToggleSave = (job: Job) => setSavedJobs((current) => toggleSaved(current, job))
  const handleHide = (id: string) => setHiddenIds((current) => toggleHidden(current, id))

  const savedList = Object.values(savedJobs).sort(
    (a, b) => new Date(b.PublishedAt).getTime() - new Date(a.PublishedAt).getTime(),
  )
  const visibleJobs = view === 'all' ? jobs.filter((j) => !hiddenIds.has(j.ID)) : savedList
  const hiddenOnPage = view === 'all' ? jobs.length - visibleJobs.length : 0

  return (
    <div className="app">
      <header className="app-header">
        <h1>job-radar</h1>
        <p>European job postings, aggregated from Arbeitnow and Remotive.</p>
      </header>

      <div className="view-tabs">
        <button className={view === 'all' ? 'active' : ''} onClick={() => setView('all')}>
          Search
        </button>
        <button className={view === 'saved' ? 'active' : ''} onClick={() => setView('saved')}>
          Saved ({savedList.length})
        </button>
      </div>

      {view === 'all' && <FilterBar filters={filters} onChange={setFilters} />}

      {view === 'all' && hiddenOnPage > 0 && (
        <p className="hidden-note">
          {hiddenOnPage} job{hiddenOnPage > 1 ? 's' : ''} hidden on this page.{' '}
          <button className="link-btn" onClick={() => setHiddenIds(new Set())}>
            Show all
          </button>
        </p>
      )}

      <JobList
        jobs={visibleJobs}
        loading={view === 'all' && loading}
        error={view === 'all' ? error : null}
        savedIds={new Set(Object.keys(savedJobs))}
        onSelect={setSelectedJob}
        onToggleSave={handleToggleSave}
        onHide={handleHide}
      />

      {view === 'all' && (
        <Pagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />
      )}

      {selectedJob && <JobDetail job={selectedJob} onClose={() => setSelectedJob(null)} />}
    </div>
  )
}
