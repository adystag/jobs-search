package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUnauthenticated = errors.New("user/password is not correct")
)

type UserAuthenticator interface {
	AuthenticateUser(ctx context.Context, req UserAuthenticationRequest) (User, error)
}

type UserRegistrator interface {
	RegisterUser(ctx context.Context, req UserRegistrationRequest) (User, error)
}

type UserGetterByUsername interface {
	GetUserByUsername(ctx context.Context, username string) (User, error)
}

type UserStorer interface {
	StoreUser(ctx context.Context, user *User) error
}

type UserAuthenticationRequest struct {
	Username string
	Password string
}

type UserRegistrationRequest struct {
	UserAuthenticationRequest
	PasswordConfirmation string
}

type User struct {
	ID        int64
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type userRegistrator struct {
	validator  Validator[UserRegistrationRequest]
	timer      Timer
	hasher     Hasher
	userStorer UserStorer
}

func (ur userRegistrator) RegisterUser(ctx context.Context, req UserRegistrationRequest) (User, error) {
	err := ur.validator.Validate(ctx, req)
	if err != nil {
		return User{}, fmt.Errorf("validating user registration request: %w", err)
	}

	req.Password, err = ur.hasher.Hash(req.Password)
	if err != nil {
		return User{}, fmt.Errorf("hashing plain user password: %w", err)
	}

	now := ur.timer.Now()
	user := User{
		Username:  req.Username,
		Password:  req.Password,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = ur.userStorer.StoreUser(ctx, &user)
	if err != nil {
		return User{}, fmt.Errorf("storing user: %w", err)
	}

	return user, nil
}

func NewUserRegistrator(
	validator Validator[UserRegistrationRequest],
	timer Timer,
	hasher Hasher,
	userStorer UserStorer,
) *userRegistrator {
	return &userRegistrator{
		validator:  validator,
		timer:      timer,
		hasher:     hasher,
		userStorer: userStorer,
	}
}

type userRegistrationRequestValidator struct {
	validate *validator.Validate
}

func (v userRegistrationRequestValidator) EvaluateErrorAs(err, target error) error {
	if errors.As(err, &validator.ValidationErrors{}) {
		err = target
	}

	return err
}

func (v userRegistrationRequestValidator) Validate(ctx context.Context, req UserRegistrationRequest) error {
	err := v.validate.VarCtx(ctx, req.Username, "required")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("username", "required"))
	}

	err = v.validate.VarCtx(ctx, req.Username, "min=3")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("username", "min=3"))
	}

	err = v.validate.VarCtx(ctx, req.Username, "max=15")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("username", "max=15"))
	}

	err = v.validate.VarCtx(ctx, req.Username, "alphanum")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("username", "alphanum"))
	}

	err = v.validate.VarCtx(ctx, req.Password, "required")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("password", "required"))
	}

	err = v.validate.VarCtx(ctx, req.Password, "min=6")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("password", "min=6"))
	}

	err = v.validate.VarWithValueCtx(ctx, req.Password, req.PasswordConfirmation, "eqfield")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("password", "eqfield=password_confirmation"))
	}

	return nil
}

func NewUserRegistrationRequestValidator(validate *validator.Validate) *userRegistrationRequestValidator {
	return &userRegistrationRequestValidator{
		validate: validate,
	}
}

type usernameUniquenessValidator struct {
	userGetterByUsername UserGetterByUsername
}

func (v usernameUniquenessValidator) Validate(ctx context.Context, req UserRegistrationRequest) error {
	user, err := v.userGetterByUsername.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("getting user by username: %w", err)
		}
	}

	if user.ID > 0 {
		return NewValidationError("username", "unique")
	}

	return nil
}

func NewUsernameUniquenessValidator(userGetterByUsername UserGetterByUsername) *usernameUniquenessValidator {
	return &usernameUniquenessValidator{
		userGetterByUsername: userGetterByUsername,
	}
}

type userAuthenticator struct {
	validator            Validator[UserAuthenticationRequest]
	userGetterByUsername UserGetterByUsername
	comparator           Comparator
}

func (ua userAuthenticator) AuthenticateUser(ctx context.Context, req UserAuthenticationRequest) (User, error) {
	err := ua.validator.Validate(ctx, req)
	if err != nil {
		return User{}, fmt.Errorf("validating user authentication request: %w", err)
	}

	user, err := ua.userGetterByUsername.GetUserByUsername(ctx, req.Username)
	if err != nil {
		err = fmt.Errorf("getting user by username: %w", err)

		if errors.Is(err, ErrUserNotFound) {
			err = ErrUnauthenticated
		}

		return User{}, err
	}

	err = ua.comparator.Compare(user.Password, req.Password)
	if err != nil {
		err = fmt.Errorf("comparing hashed with plain user password: %w", err)

		if errors.Is(err, ErrHashMismatched) {
			err = ErrUnauthenticated
		}

		return User{}, err
	}

	return user, nil
}

func NewUserAuthenticator(
	validator Validator[UserAuthenticationRequest],
	userGetterByUsername UserGetterByUsername,
	comparator Comparator,
) *userAuthenticator {
	return &userAuthenticator{
		validator:            validator,
		userGetterByUsername: userGetterByUsername,
		comparator:           comparator,
	}
}

type userAuthenticationRequestValidator struct {
	validate *validator.Validate
}

func (v userAuthenticationRequestValidator) EvaluateErrorAs(err, target error) error {
	if errors.As(err, &validator.ValidationErrors{}) {
		err = target
	}

	return err
}

func (v userAuthenticationRequestValidator) Validate(ctx context.Context, req UserAuthenticationRequest) error {
	err := v.validate.VarCtx(ctx, req.Username, "required")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("username", "required"))
	}

	err = v.validate.VarCtx(ctx, req.Password, "required")
	if err != nil {
		return v.EvaluateErrorAs(err, NewValidationError("password", "required"))
	}

	return nil
}

func NewUserAuthenticationRequestValidator(validate *validator.Validate) *userAuthenticationRequestValidator {
	return &userAuthenticationRequestValidator{
		validate: validate,
	}
}
