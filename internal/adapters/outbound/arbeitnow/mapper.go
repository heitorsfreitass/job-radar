package arbeitnow

import (
	"strings"
	"time"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

// rawJob mirrors the shape of a single entry in Arbeitnow's
// `job-board-api` `data` array.
type rawJob struct {
	Slug        string   `json:"slug"`
	CompanyName string   `json:"company_name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Remote      bool     `json:"remote"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	JobTypes    []string `json:"job_types"`
	Location    string   `json:"location"`
	CreatedAt   int64    `json:"created_at"` // unix seconds
}

func (r rawJob) toDomain() *domain.Job {
	now := time.Now().UTC()
	return &domain.Job{
		Title:       r.Title,
		CompanyName: r.CompanyName,
		Description: r.Description,
		URL:         r.URL,
		Source:      domain.SourceArbeitnow,
		SourceJobID: r.Slug,
		Country:     r.Location,
		Workplace:   workplaceFrom(r),
		Employment:  employmentFrom(r.JobTypes),
		Seniority:   seniorityFrom(r.Title),
		Tags:        r.Tags,
		PublishedAt: time.Unix(r.CreatedAt, 0).UTC(),
		IngestedAt:  now,
	}
}

func workplaceFrom(r rawJob) domain.WorkplaceType {
	if r.Remote {
		return domain.WorkplaceRemote
	}
	for _, t := range r.Tags {
		if strings.Contains(strings.ToLower(t), "hybrid") {
			return domain.WorkplaceHybrid
		}
	}
	if strings.Contains(strings.ToLower(r.Location), "remote") {
		return domain.WorkplaceRemote
	}
	return domain.WorkplaceOnsite
}

var employmentByArbeitnowType = map[string]domain.EmploymentType{
	"full-time": domain.EmploymentTypeFullTime,
	"part-time": domain.EmploymentTypePartTime,
	"contract":  domain.EmploymentTypeContract,
	"internship": domain.EmploymentTypeInternship,
	"freelance":  domain.EmploymentTypeContract,
}

func employmentFrom(jobTypes []string) domain.EmploymentType {
	for _, jt := range jobTypes {
		if e, ok := employmentByArbeitnowType[strings.ToLower(jt)]; ok {
			return e
		}
	}
	return domain.EmploymentTypeUnknown
}

// seniorityFrom applies a best-effort keyword heuristic over the job
// title, since Arbeitnow does not expose a structured seniority field.
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
