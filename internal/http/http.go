package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/adystag/jobs-search/internal"

	"github.com/gofiber/fiber/v2"
)

type Presenter[T any] interface {
	Present(ctx *fiber.Ctx, val T) error
}

type Server struct {
	app  *fiber.App
	addr string
}

func (s Server) Run() error {
	err := s.app.Listen(s.addr)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listening http request on %s: %w", s.addr, err)
		}
	}

	return nil
}

func (s Server) Stop() error {
	err := s.app.Shutdown()
	if err != nil {
		return fmt.Errorf("shutting down http server: %w", err)
	}

	return nil
}

func NewServer(module *internal.Module) *Server {
	app := fiber.New()

	pingHandler := NewPingHandler()

	app.Get("/ping", pingHandler.Handle)

	api := app.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			user := v1.Group("/user")
			{
				jwtUserPresenter := NewJWTUserPresenter(
					module.Timer,
					module.Configuration.Application.URL,
					module.Configuration.JWT.LifeTime,
					module.Configuration.Application.Secret,
				)
				userRegistrationHandler := NewUserRegistrationHandler(module.UserRegistrator, jwtUserPresenter)

				user.Post("/registration", userRegistrationHandler.Handle)

				userAuthenticationHandler := NewUserAuthenticationHandler(module.UserAuthenticator, jwtUserPresenter)

				user.Post("/login", userAuthenticationHandler.Handle)
			}

			jwtAuthenticationMiddleware := NewJWTAuthenticationMiddleware(module.Configuration.Application.Secret)

			job := v1.Group("/job", jwtAuthenticationMiddleware.Handle)
			{
				jobsListingHandler := NewJobsListingHandler(module.JobsLister, NewJobsListPresenter())

				job.Get("/", jobsListingHandler.Handle)

				jobGetterByIDHandler := NewJobGetterByIDHandler(module.JobGetterByID)

				job.Get("/:jobID", jobGetterByIDHandler.Handle)
			}
		}
	}

	return &Server{
		app:  app,
		addr: fmt.Sprintf(":%s", module.Configuration.Application.Port),
	}
}
