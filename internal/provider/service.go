package provider

import (
	"github.com/adystag/jobs-search/internal"
	"github.com/adystag/jobs-search/internal/repository/http"
	"github.com/adystag/jobs-search/internal/repository/mysql"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

func (Service) Provide(module *internal.Module) error {
	module.Timer = internal.NewTimer()

	validate := validator.New()
	bcryptHasher := internal.NewBcryptHasher(bcrypt.DefaultCost)
	userRepository := mysql.NewUserRepository(module.DB)

	module.UserRegistrator = internal.NewUserRegistrator(
		internal.NewValidationAggregator[internal.UserRegistrationRequest](
			internal.NewUserRegistrationRequestValidator(validate),
			internal.NewUsernameUniquenessValidator(userRepository),
		),
		module.Timer,
		bcryptHasher,
		userRepository,
	)
	module.UserAuthenticator = internal.NewUserAuthenticator(
		internal.NewUserAuthenticationRequestValidator(validate),
		userRepository,
		bcryptHasher,
	)

	jobRepository := http.NewJobRepository(module.Configuration.DANS.BaseURL)

	module.JobsLister = jobRepository
	module.JobGetterByID = jobRepository

	return nil
}
