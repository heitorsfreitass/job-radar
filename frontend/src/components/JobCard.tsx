import type { Job } from '../types'

interface Props {
  job: Job
  isSaved: boolean
  onSelect: (job: Job) => void
  onToggleSave: (job: Job) => void
  onHide: (id: string) => void
}

function timeAgo(iso: string): string {
  const days = Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000)
  if (days <= 0) return 'today'
  if (days === 1) return '1 day ago'
  if (days < 30) return `${days} days ago`
  const months = Math.floor(days / 30)
  return months === 1 ? '1 month ago' : `${months} months ago`
}

function isNew(ingestedAt: string): boolean {
  return Date.now() - new Date(ingestedAt).getTime() < 24 * 60 * 60 * 1000
}

export default function JobCard({ job, isSaved, onSelect, onToggleSave, onHide }: Props) {
  return (
    <div className="job-card">
      <div className="job-card-actions">
        <button
          className={`icon-btn ${isSaved ? 'icon-btn-active' : ''}`}
          title={isSaved ? 'Remove from saved' : 'Save job'}
          onClick={(e) => {
            e.stopPropagation()
            onToggleSave(job)
          }}
        >
          {isSaved ? '★' : '☆'}
        </button>
        <button
          className="icon-btn"
          title="Hide this job"
          onClick={(e) => {
            e.stopPropagation()
            onHide(job.ID)
          }}
        >
          ✕
        </button>
      </div>
      <button className="job-card-main" onClick={() => onSelect(job)}>
        <div className="job-card-header">
          <h3>{job.Title}</h3>
          <div className="job-card-badges">
            {isNew(job.IngestedAt) && <span className="badge badge-new">new</span>}
            <span className={`badge badge-${job.Workplace}`}>{job.Workplace}</span>
          </div>
        </div>
        <p className="job-card-company">{job.CompanyName}</p>
        <div className="job-card-meta">
          {job.Country && <span>{job.Country}</span>}
          {job.Seniority !== 'unknown' && <span>{job.Seniority}</span>}
          {job.Employment !== 'unknown' && <span>{job.Employment.replace('_', ' ')}</span>}
          <span className="job-card-source">via {job.Source}</span>
          <span className="job-card-time">{timeAgo(job.PublishedAt)}</span>
        </div>
      </button>
    </div>
  )
}
