package remotive

import (
	"strconv"
	"strings"
	"time"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// remotiveAttributionTag is appended to every job's Tags so API consumers
// can see, without inspecting the Source field, that this listing must
// keep linking back to Remotive (see README's "Attribution" rule).
const remotiveAttributionTag = "via Remotive"

// rawJob mirrors a single entry in Remotive's `remote-jobs` `jobs` array.
type rawJob struct {
	ID                        int64    `json:"id"`
	URL                       string   `json:"url"`
	Title                     string   `json:"title"`
	CompanyName               string   `json:"company_name"`
	Category                  string   `json:"category"`
	Tags                      []string `json:"tags"`
	JobType                   string   `json:"job_type"`
	PublicationDate           string   `json:"publication_date"`
	CandidateRequiredLocation string   `json:"candidate_required_location"`
	Description               string   `json:"description"`
}

func (r rawJob) toDomain() *domain.Job {
	now := time.Now().UTC()

	published := now
	if t, err := time.Parse("2006-01-02T15:04:05", r.PublicationDate); err == nil {
		published = t.UTC()
	}

	return &domain.Job{
		Title:       r.Title,
		CompanyName: r.CompanyName,
		Description: r.Description,
		URL:         r.URL, // MUST stay the original Remotive apply URL, untouched
		Source:      domain.SourceRemotive,
		SourceJobID: strconv.FormatInt(r.ID, 10),
		Country:     r.CandidateRequiredLocation,
		Workplace:   domain.WorkplaceRemote, // every Remotive listing is a remote job by definition
		Employment:  employmentFrom(r.JobType),
		Seniority:   seniorityFrom(r.Title),
		Tags:        append(append([]string{}, r.Tags...), remotiveAttributionTag),
		PublishedAt: published,
		IngestedAt:  now,
	}
}

var employmentByRemotiveType = map[string]domain.EmploymentType{
	"full_time":  domain.EmploymentTypeFullTime,
	"part_time":  domain.EmploymentTypePartTime,
	"contract":   domain.EmploymentTypeContract,
	"freelance":  domain.EmploymentTypeContract,
	"internship": domain.EmploymentTypeInternship,
}

func employmentFrom(jobType string) domain.EmploymentType {
	if e, ok := employmentByRemotiveType[strings.ToLower(jobType)]; ok {
		return e
	}
	return domain.EmploymentTypeUnknown
}

// seniorityFrom applies a best-effort keyword heuristic over the job
// title, since Remotive does not expose a structured seniority field.
func seniorityFrom(title string) domain.SeniorityLevel {
	t := strings.ToLower(title)
	switch {
	case strings.Contains(t, "senior") || strings.Contains(t, "sr."):
		return domain.SenioritySenior
	case strings.Contains(t, "lead") || strings.Contains(t, "principal") || strings.Contains(t, "staff"):
		return domain.SeniorityLead
	case strings.Contains(t, "junior") || strings.Contains(t, "jr.") || strings.Contains(t, "entry level") || strings.Contains(t, "intern"):
		return domain.SeniorityJunior
	case strings.Contains(t, "mid-level") || strings.Contains(t, "mid level"):
		return domain.SeniorityMid
	default:
		return domain.SeniorityUnknown
	}
}
