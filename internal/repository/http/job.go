package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"

	gourl "net/url"

	"github.com/adystag/jobs-search/internal"
	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID `json:"id"`
	Company     string    `json:"company"`
	CompanyURL  string    `json:"company_url"`
	CompanyLogo string    `json:"company_logo"`
	URL         string    `json:"url"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	HowToApply  string    `json:"how_to_apply"`
	CreatedAt   string    `json:"created_at"`
}

type jobRepository struct {
	baseURL string
}

func (jr jobRepository) ListJobs(ctx context.Context, opts ...internal.Option[internal.JobsListerOption]) ([]internal.Job, error) {
	url, err := gourl.Parse(jr.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %w", err)
	}

	url.Path = path.Join(url.Path, "api/recruitment/positions.json")
	opt := internal.JobsListerOption{}

	internal.ApplyOptions(&opt, opts...)

	values := url.Query()

	if len(opt.Description) > 0 {
		values.Set("description", opt.Description)
	}

	if len(opt.Location) > 0 {
		values.Set("location", opt.Location)
	}

	if opt.FullTime {
		values.Set("full_time", strconv.FormatBool(opt.FullTime))
	}

	if opt.Page > 0 {
		values.Set("page", strconv.Itoa(opt.Page))
	}

	url.RawQuery = values.Encode()
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("initializing new request: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing http request: %w", err)
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading http response body: %w", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http request returns %d:%s", res.StatusCode, string(b))
	}

	var mJobs []Job

	err = json.Unmarshal(b, &mJobs)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling body response from json: %w", err)
	}

	var jobs []internal.Job

	for _, each := range mJobs {
		jobs = append(jobs, internal.Job{
			ID:          each.ID,
			Company:     each.Company,
			CompanyURL:  each.CompanyURL,
			CompanyLogo: each.CompanyLogo,
			URL:         each.URL,
			Type:        each.Type,
			Location:    each.Location,
			Title:       each.Title,
			Description: each.Description,
			HowToApply:  each.HowToApply,
			CreatedAt:   each.CreatedAt,
		})
	}

	return jobs, nil
}

func (jr jobRepository) GetJobByID(ctx context.Context, jobID string) (internal.Job, error) {
	id, err := uuid.Parse(jobID)
	if err != nil {
		return internal.Job{}, internal.NewValidationError("job_id", "uuid")
	}

	url, err := gourl.Parse(jr.baseURL)
	if err != nil {
		return internal.Job{}, fmt.Errorf("parsing base url: %w", err)
	}

	url.Path = path.Join(url.Path, fmt.Sprintf("/api/recruitment/positions/%s", id.String()))
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return internal.Job{}, fmt.Errorf("initializing new request: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return internal.Job{}, fmt.Errorf("doing http request: %w", err)
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return internal.Job{}, fmt.Errorf("reading http response body: %w", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return internal.Job{}, fmt.Errorf("http request returns %d:%s", res.StatusCode, string(b))
	}

	var job Job

	err = json.Unmarshal(b, &job)
	if err != nil {
		return internal.Job{}, fmt.Errorf("unmarshalling body response from json: %w", err)
	}

	return internal.Job{
		ID:          job.ID,
		Company:     job.Company,
		CompanyURL:  job.CompanyURL,
		CompanyLogo: job.CompanyLogo,
		URL:         job.URL,
		Type:        job.Type,
		Location:    job.Location,
		Title:       job.Title,
		Description: job.Description,
		HowToApply:  job.HowToApply,
		CreatedAt:   job.CreatedAt,
	}, nil
}

func NewJobRepository(baseURL string) *jobRepository {
	return &jobRepository{
		baseURL: baseURL,
	}
}
