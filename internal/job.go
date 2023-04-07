package internal

import (
	"context"

	"github.com/google/uuid"
)

type JobsLister interface {
	ListJobs(ctx context.Context, opts ...Option[JobsListerOption]) ([]Job, error)
}

type JobGetterByID interface {
	GetJobByID(ctx context.Context, jobID string) (Job, error)
}

type JobsListerOption struct {
	Description string
	Location    string
	FullTime    bool
	Page        int
}

func WithJobsListerDescription(description string) Option[JobsListerOption] {
	return func(opt *JobsListerOption) {
		opt.Description = description
	}
}

func WithJobsListerLocation(location string) Option[JobsListerOption] {
	return func(opt *JobsListerOption) {
		opt.Location = location
	}
}

func WithJobsListerFullTime(fullTime bool) Option[JobsListerOption] {
	return func(opt *JobsListerOption) {
		opt.FullTime = fullTime
	}
}

func WithJobsListerPage(page int) Option[JobsListerOption] {
	return func(opt *JobsListerOption) {
		opt.Page = page
	}
}

type Job struct {
	ID          uuid.UUID
	Company     string
	CompanyURL  string
	CompanyLogo string
	URL         string
	Type        string
	Location    string
	Title       string
	Description string
	HowToApply  string
	CreatedAt   string
}
