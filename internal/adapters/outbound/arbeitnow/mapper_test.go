package arbeitnow

import (
	"encoding/json"
	"testing"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

func TestToDomain(t *testing.T) {
	raw := rawJob{
		Slug:        "senior-go-engineer-berlin-1",
		CompanyName: "Acme",
		Title:       "Senior Go Engineer",
		Description: "desc",
		Remote:      true,
		URL:         "https://www.arbeitnow.com/jobs/companies/acme/senior-go-engineer-berlin-1",
		Tags:        []string{"Go", "Backend"},
		JobTypes:    []string{"Full-time"},
		Location:    "Berlin",
		CreatedAt:   1_700_000_000,
	}

	job := raw.toDomain()

	if job.Source != domain.SourceArbeitnow {
		t.Errorf("Source = %q, want %q", job.Source, domain.SourceArbeitnow)
	}
	if job.SourceJobID != "senior-go-engineer-berlin-1" {
		t.Errorf("SourceJobID = %q, want slug", job.SourceJobID)
	}
	if job.URL != raw.URL {
		t.Errorf("URL = %q, want %q", job.URL, raw.URL)
	}
	if job.Workplace != domain.WorkplaceRemote {
		t.Errorf("Workplace = %q, want remote (raw.Remote=true)", job.Workplace)
	}
	if job.Employment != domain.EmploymentTypeFullTime {
		t.Errorf("Employment = %q, want full_time", job.Employment)
	}
	if job.Seniority != domain.SenioritySenior {
		t.Errorf("Seniority = %q, want senior (title contains \"Senior\")", job.Seniority)
	}
	if job.PublishedAt.Unix() != raw.CreatedAt {
		t.Errorf("PublishedAt = %v, want unix %d", job.PublishedAt, raw.CreatedAt)
	}
}

func TestStringList_HandlesObjectShapeInPlaceOfArray(t *testing.T) {
	var raw rawJob
	// Arbeitnow serializes tags/job_types as a JSON object keyed by a
	// non-contiguous index (e.g. `{"1":"..."}`) for at least some
	// listings — seen live in production for "job_types". The values
	// must still be captured, not dropped.
	err := json.Unmarshal([]byte(`{"slug":"x","tags":{},"job_types":{"1":"professional / experienced"}}`), &raw)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if len(raw.Tags) != 0 {
		t.Errorf("Tags = %v, want empty (empty object)", raw.Tags)
	}
	if want := []string{"professional / experienced"}; len(raw.JobTypes) != 1 || raw.JobTypes[0] != want[0] {
		t.Errorf("JobTypes = %v, want %v", raw.JobTypes, want)
	}
}

func TestWorkplaceFrom(t *testing.T) {
	cases := []struct {
		name string
		raw  rawJob
		want domain.WorkplaceType
	}{
		{"explicit remote flag", rawJob{Remote: true}, domain.WorkplaceRemote},
		{"hybrid tag", rawJob{Tags: []string{"Hybrid"}}, domain.WorkplaceHybrid},
		{"remote in location text", rawJob{Location: "Remote - Germany Only"}, domain.WorkplaceRemote},
		{"defaults onsite", rawJob{Location: "Berlin"}, domain.WorkplaceOnsite},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := workplaceFrom(tc.raw); got != tc.want {
				t.Errorf("workplaceFrom() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEmploymentFrom(t *testing.T) {
	cases := []struct {
		jobTypes []string
		want     domain.EmploymentType
	}{
		{[]string{"Full-time"}, domain.EmploymentTypeFullTime},
		{[]string{"Internship"}, domain.EmploymentTypeInternship},
		{[]string{"Freelance"}, domain.EmploymentTypeContract},
		{nil, domain.EmploymentTypeUnknown},
		{[]string{"Something Unrecognized"}, domain.EmploymentTypeUnknown},
	}
	for _, tc := range cases {
		if got := employmentFrom(tc.jobTypes); got != tc.want {
			t.Errorf("employmentFrom(%v) = %q, want %q", tc.jobTypes, got, tc.want)
		}
	}
}

func TestSeniorityFrom(t *testing.T) {
	cases := []struct {
		title string
		want  domain.SeniorityLevel
	}{
		{"Senior Backend Engineer", domain.SenioritySenior},
		{"Junior Developer", domain.SeniorityJunior},
		{"Staff Engineer", domain.SeniorityLead},
		{"Software Engineer", domain.SeniorityUnknown},
	}
	for _, tc := range cases {
		if got := seniorityFrom(tc.title); got != tc.want {
			t.Errorf("seniorityFrom(%q) = %q, want %q", tc.title, got, tc.want)
		}
	}
}
