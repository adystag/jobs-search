package internal

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Provider interface {
	Provide(module *Module) error
}

type Module struct {
	Configuration struct {
		Application struct {
			Env    string
			Port   string
			URL    string
			Secret []byte
		}
		JWT struct {
			LifeTime time.Duration
		}
		DB struct {
			Host        string
			Port        string
			User        string
			Password    string
			Name        string
			AutoMigrate bool
		}
		DANS struct {
			BaseURL string
		}
	}

	DB *sqlx.DB

	Timer Timer

	UserRegistrator   UserRegistrator
	UserAuthenticator UserAuthenticator

	JobsLister    JobsLister
	JobGetterByID JobGetterByID
}

func NewModule(providers ...Provider) (*Module, error) {
	module := Module{}

	for _, each := range providers {
		err := each.Provide(&module)
		if err != nil {
			return nil, fmt.Errorf("providing module: %w", err)
		}
	}

	return &module, nil
}
