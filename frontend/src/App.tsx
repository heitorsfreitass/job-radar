import { useEffect, useState } from 'react'
import { getMe, savePreferences, searchJobs } from './api'
import AuthModal from './components/AuthModal'
import FilterBar from './components/FilterBar'
import JobDetail from './components/JobDetail'
import JobList from './components/JobList'
import Pagination from './components/Pagination'
import {
  clearAuth,
  loadAuth,
  loadHiddenIds,
  loadJSON,
  loadSavedJobs,
  saveAuth,
  saveJSON,
  toggleHidden,
  toggleSaved,
  type StoredAuth,
} from './storage'
import { EMPTY_FILTERS, filtersToPreferences, preferencesToFilters, type AuthResponse, type Filters, type Job } from './types'
import './App.css'

const PAGE_SIZE = 20

type View = 'all' | 'saved'
type SaveStatus = 'idle' | 'saving' | 'saved' | 'error'

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
  const [auth, setAuth] = useState<StoredAuth | null>(loadAuth)
  const [showAuthModal, setShowAuthModal] = useState(false)
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle')

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

  // A stored token might have expired since the last visit; drop it
  // quietly rather than showing a signed-in state that 401s on save.
  useEffect(() => {
    if (!auth) return
    getMe(auth.token).catch(() => {
      clearAuth()
      setAuth(null)
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleToggleSave = (job: Job) => setSavedJobs((current) => toggleSaved(current, job))
  const handleHide = (id: string) => setHiddenIds((current) => toggleHidden(current, id))

  const handleAuthSuccess = (resp: AuthResponse) => {
    const next: StoredAuth = { token: resp.token, email: resp.email }
    saveAuth(next)
    setAuth(next)
    setShowAuthModal(false)

    getMe(resp.token)
      .then((me) => {
        const prefFilters = preferencesToFilters(me.preferences)
        if (Object.values(prefFilters).some((v) => v !== '')) {
          setFilters(prefFilters)
        }
      })
      .catch(() => {})
  }

  const handleLogout = () => {
    clearAuth()
    setAuth(null)
  }

  const handleSaveDefault = () => {
    if (!auth) return
    setSaveStatus('saving')
    savePreferences(auth.token, filtersToPreferences(filters))
      .then(() => {
        setSaveStatus('saved')
        setTimeout(() => setSaveStatus('idle'), 2000)
      })
      .catch(() => setSaveStatus('error'))
  }

  const savedList = Object.values(savedJobs).sort(
    (a, b) => new Date(b.PublishedAt).getTime() - new Date(a.PublishedAt).getTime(),
  )
  const visibleJobs = view === 'all' ? jobs.filter((j) => !hiddenIds.has(j.ID)) : savedList
  const hiddenOnPage = view === 'all' ? jobs.length - visibleJobs.length : 0

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-header-top">
          <div>
            <h1>job-radar</h1>
            <p>European job postings, aggregated from Arbeitnow and Remotive.</p>
          </div>
          {auth ? (
            <div className="auth-bar">
              <span className="auth-email">{auth.email}</span>
              <button className="link-btn" onClick={handleLogout}>
                Log out
              </button>
            </div>
          ) : (
            <button className="auth-cta" onClick={() => setShowAuthModal(true)}>
              Log in / Sign up
            </button>
          )}
        </div>
      </header>

      <div className="view-tabs">
        <button className={view === 'all' ? 'active' : ''} onClick={() => setView('all')}>
          Search
        </button>
        <button className={view === 'saved' ? 'active' : ''} onClick={() => setView('saved')}>
          Saved ({savedList.length})
        </button>
      </div>

      {view === 'all' && (
        <div className="filter-row">
          <FilterBar filters={filters} onChange={setFilters} />
          {auth && (
            <button className="save-default-btn" onClick={handleSaveDefault} disabled={saveStatus === 'saving'}>
              {saveStatus === 'saved' ? 'Saved ✓' : saveStatus === 'saving' ? 'Saving…' : 'Save as default'}
            </button>
          )}
        </div>
      )}

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
      {showAuthModal && <AuthModal onSuccess={handleAuthSuccess} onClose={() => setShowAuthModal(false)} />}
    </div>
  )
}
