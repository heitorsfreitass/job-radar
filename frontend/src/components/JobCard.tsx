import type { Job } from '../types'

interface Props {
  job: Job
  onSelect: (job: Job) => void
}

function timeAgo(iso: string): string {
  const days = Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000)
  if (days <= 0) return 'today'
  if (days === 1) return '1 day ago'
  if (days < 30) return `${days} days ago`
  const months = Math.floor(days / 30)
  return months === 1 ? '1 month ago' : `${months} months ago`
}

export default function JobCard({ job, onSelect }: Props) {
  return (
    <button className="job-card" onClick={() => onSelect(job)}>
      <div className="job-card-header">
        <h3>{job.Title}</h3>
        <span className={`badge badge-${job.Workplace}`}>{job.Workplace}</span>
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
  )
}
