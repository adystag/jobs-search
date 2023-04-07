package http

import (
	"encoding/json"
	"fmt"

	"github.com/adystag/jobs-search/internal"

	"github.com/gofiber/fiber/v2"
)

type PresentableJob internal.Job

func (pj PresentableJob) MarshalJSON() ([]byte, error) {
	tmp := struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		CreatedAt   string `json:"created_at"`
		Company     string `json:"company"`
		CompanyURL  string `json:"company_url"`
		Location    string `json:"location"`
		Title       string `json:"title"`
		Description string `json:"description"`
		HowToApply  string `json:"how_to_apply"`
		CompanyLogo string `json:"company_logo"`
	}{
		ID:          pj.ID.String(),
		Type:        pj.Type,
		URL:         pj.URL,
		CreatedAt:   pj.CreatedAt,
		Company:     pj.Company,
		CompanyURL:  pj.CompanyURL,
		Location:    pj.Location,
		Title:       pj.Title,
		Description: pj.Description,
		HowToApply:  pj.HowToApply,
		CompanyLogo: pj.CompanyLogo,
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		return nil, fmt.Errorf("marshalling job to json: %w", err)
	}

	return b, nil
}

type JobsListPresenter struct{}

func (JobsListPresenter) Present(ctx *fiber.Ctx, jobs []internal.Job) error {
	presentableJobs := []PresentableJob{}

	for _, each := range jobs {
		presentableJobs = append(presentableJobs, PresentableJob(each))
	}

	return ctx.Status(fiber.StatusOK).JSON(presentableJobs)
}

func NewJobsListPresenter() *JobsListPresenter {
	return &JobsListPresenter{}
}

type JobsListingHandler struct {
	jobsLister internal.JobsLister
	presenter  Presenter[[]internal.Job]
}

func (h JobsListingHandler) Handle(ctx *fiber.Ctx) error {
	opts := []internal.Option[internal.JobsListerOption]{}

	description := ctx.Query("description")
	if len(description) > 0 {
		opts = append(opts, internal.WithJobsListerDescription(description))
	}

	location := ctx.Query("location")
	if len(location) > 0 {
		opts = append(opts, internal.WithJobsListerLocation(location))
	}

	fullTime := ctx.QueryBool("full_time")
	if fullTime {
		opts = append(opts, internal.WithJobsListerFullTime(fullTime))
	}

	page := ctx.QueryInt("page")
	if page >= 1 {
		opts = append(opts, internal.WithJobsListerPage(page))
	}

	jobs, err := h.jobsLister.ListJobs(ctx.Context(), opts...)
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}

	return h.presenter.Present(ctx, jobs)
}

func NewJobsListingHandler(
	jobsLister internal.JobsLister,
	presenter Presenter[[]internal.Job],
) *JobsListingHandler {
	return &JobsListingHandler{
		jobsLister: jobsLister,
		presenter:  presenter,
	}
}

type JobGetterByIDHandler struct {
	jobGetterByID internal.JobGetterByID
}

func (h JobGetterByIDHandler) Handle(ctx *fiber.Ctx) error {
	job, err := h.jobGetterByID.GetJobByID(ctx.Context(), ctx.Params("jobID"))
	if err != nil {
		return fmt.Errorf("getting job by id: %w", err)
	}

	return ctx.Status(fiber.StatusOK).JSON(PresentableJob(job))
}

func NewJobGetterByIDHandler(jobGetterByID internal.JobGetterByID) *JobGetterByIDHandler {
	return &JobGetterByIDHandler{
		jobGetterByID: jobGetterByID,
	}
}
