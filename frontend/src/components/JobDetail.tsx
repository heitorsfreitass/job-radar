import DOMPurify from 'dompurify'
import type { Job } from '../types'

interface Props {
  job: Job
  onClose: () => void
}

export default function JobDetail({ job, onClose }: Props) {
  const safeDescription = DOMPurify.sanitize(job.Description)

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose} aria-label="Close">
          ×
        </button>
        <h2>{job.Title}</h2>
        <p className="job-detail-company">
          {job.CompanyName} · {job.Country || 'Location unspecified'}
        </p>
        <div className="job-detail-badges">
          <span className={`badge badge-${job.Workplace}`}>{job.Workplace}</span>
          {job.Seniority !== 'unknown' && <span className="badge">{job.Seniority}</span>}
          {job.Employment !== 'unknown' && <span className="badge">{job.Employment.replace('_', ' ')}</span>}
        </div>
        {job.Tags.length > 0 && (
          <div className="job-detail-tags">
            {job.Tags.map((tag) => (
              <span key={tag} className="tag">
                {tag}
              </span>
            ))}
          </div>
        )}
        <div
          className="job-detail-description"
          dangerouslySetInnerHTML={{ __html: safeDescription }}
        />
        <a className="apply-link" href={job.URL} target="_blank" rel="noopener noreferrer">
          Apply on {job.Source} →
        </a>
      </div>
    </div>
  )
}
