import type { Job } from '../types'
import JobCard from './JobCard'

interface Props {
  jobs: Job[]
  loading: boolean
  error: string | null
  onSelect: (job: Job) => void
}

export default function JobList({ jobs, loading, error, onSelect }: Props) {
  if (error) {
    return <p className="state-message state-error">Failed to load jobs: {error}</p>
  }
  if (loading) {
    return <p className="state-message">Loading jobs…</p>
  }
  if (jobs.length === 0) {
    return <p className="state-message">No jobs match these filters.</p>
  }

  return (
    <div className="job-list">
      {jobs.map((job) => (
        <JobCard key={job.ID} job={job} onSelect={onSelect} />
      ))}
    </div>
  )
}
