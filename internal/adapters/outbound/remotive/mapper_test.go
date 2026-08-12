package remotive

import (
	"testing"

	"github.com/heitorsfreitass/job-radar/internal/domain"
)

func TestToDomain(t *testing.T) {
	raw := rawJob{
		ID:                        2086540,
		URL:                       "https://remotive.com/remote-jobs/sales/inside-sales-contractor-2086540",
		Title:                     "Senior Backend Engineer",
		CompanyName:               "Acme",
		Category:                  "Software Development",
		Tags:                      []string{"Go", "Backend"},
		JobType:                   "full_time",
		PublicationDate:           "2026-08-08T21:48:06",
		CandidateRequiredLocation: "Worldwide",
		Description:               "desc",
	}

	job := raw.toDomain()

	if job.Source != domain.SourceRemotive {
		t.Errorf("Source = %q, want %q", job.Source, domain.SourceRemotive)
	}
	if job.SourceJobID != "2086540" {
		t.Errorf("SourceJobID = %q, want \"2086540\"", job.SourceJobID)
	}
	if job.URL != raw.URL {
		t.Errorf("URL = %q, want untouched original %q", job.URL, raw.URL)
	}
	if job.Workplace != domain.WorkplaceRemote {
		t.Errorf("Workplace = %q, want remote (every Remotive job is remote)", job.Workplace)
	}
	if job.Employment != domain.EmploymentTypeFullTime {
		t.Errorf("Employment = %q, want full_time", job.Employment)
	}
	if job.Seniority != domain.SenioritySenior {
		t.Errorf("Seniority = %q, want senior", job.Seniority)
	}

	found := false
	for _, tag := range job.Tags {
		if tag == remotiveAttributionTag {
			found = true
		}
	}
	if !found {
		t.Errorf("Tags = %v, want to include %q", job.Tags, remotiveAttributionTag)
	}
	if len(job.Tags) != len(raw.Tags)+1 {
		t.Errorf("Tags = %v, want original tags plus attribution", job.Tags)
	}

	wantPublished := "2026-08-08T21:48:06Z"
	if job.PublishedAt.Format("2006-01-02T15:04:05Z") != wantPublished {
		t.Errorf("PublishedAt = %v, want %s", job.PublishedAt, wantPublished)
	}
}

func TestEmploymentFrom(t *testing.T) {
	cases := []struct {
		jobType string
		want    domain.EmploymentType
	}{
		{"full_time", domain.EmploymentTypeFullTime},
		{"part_time", domain.EmploymentTypePartTime},
		{"contract", domain.EmploymentTypeContract},
		{"freelance", domain.EmploymentTypeContract},
		{"internship", domain.EmploymentTypeInternship},
		{"unknown_type", domain.EmploymentTypeUnknown},
	}
	for _, tc := range cases {
		if got := employmentFrom(tc.jobType); got != tc.want {
			t.Errorf("employmentFrom(%q) = %q, want %q", tc.jobType, got, tc.want)
		}
	}
}
