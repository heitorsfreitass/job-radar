// Package domain holds the core business model, independent of any
// transport, storage, or external API concerns.
package domain

import "time"

// EmploymentType represents the working arrangement of a job posting.
type EmploymentType string

const (
	EmploymentTypeFullTime   EmploymentType = "full_time"
	EmploymentTypePartTime   EmploymentType = "part_time"
	EmploymentTypeContract   EmploymentType = "contract"
	EmploymentTypeInternship EmploymentType = "internship"
	EmploymentTypeUnknown    EmploymentType = "unknown"
)

// WorkplaceType represents where the job is performed.
type WorkplaceType string

const (
	WorkplaceRemote  WorkplaceType = "remote"
	WorkplaceOnsite  WorkplaceType = "onsite"
	WorkplaceHybrid  WorkplaceType = "hybrid"
	WorkplaceUnknown WorkplaceType = "unknown"
)

// SeniorityLevel represents the seniority the job posting targets.
type SeniorityLevel string

const (
	SeniorityJunior  SeniorityLevel = "junior"
	SeniorityMid     SeniorityLevel = "mid"
	SenioritySenior  SeniorityLevel = "senior"
	SeniorityLead    SeniorityLevel = "lead"
	SeniorityUnknown SeniorityLevel = "unknown"
)

// Source identifies which upstream job board a Job was ingested from.
type Source string

const (
	SourceArbeitnow Source = "arbeitnow"
	SourceRemotive  Source = "remotive"
)

// Job is the normalized representation of a job posting, shared across
// every upstream source. Source-specific adapters are responsible for
// mapping their raw payloads into this shape (see the mapper.go file in
// each adapters/outbound/<source> package).
type Job struct {
	ID          string
	Title       string
	CompanyName string
	Description string
	URL         string // canonical apply URL; for Remotive this MUST stay the original link
	Source      Source
	SourceJobID string // upstream identifier, used for dedup and traceability
	Country     string
	Workplace   WorkplaceType
	Employment  EmploymentType
	Seniority   SeniorityLevel
	Tags        []string
	PublishedAt time.Time
	IngestedAt  time.Time
}
