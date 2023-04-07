package http

import (
	"fmt"
	"strconv"
	"time"

	"github.com/adystag/jobs-search/internal"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

type JWTUserPresenter struct {
	timer    internal.Timer
	url      string
	lifeTime time.Duration
	secret   []byte
}

func (p JWTUserPresenter) Present(ctx *fiber.Ctx, user internal.User) error {
	now := p.timer.Now()
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		Audience:  strconv.FormatInt(user.ID, 10),
		Issuer:    p.url,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(p.lifeTime).Unix(),
	}).SignedString(p.secret)
	if err != nil {
		return fmt.Errorf("generating jwt token from user: %w", err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": accessToken,
		"expires_in":   p.lifeTime.Seconds(),
	})
}

func NewJWTUserPresenter(
	timer internal.Timer,
	url string,
	lifeTime time.Duration,
	secret []byte,
) *JWTUserPresenter {
	return &JWTUserPresenter{
		timer:    timer,
		url:      url,
		lifeTime: lifeTime,
		secret:   secret,
	}
}

type UserRegistrationHandler struct {
	userRegistrator internal.UserRegistrator
	userPresenter   Presenter[internal.User]
}

func (h UserRegistrationHandler) Handle(ctx *fiber.Ctx) error {
	var userRegistrationRequest struct {
		Username             string `json:"username"`
		Password             string `json:"password"`
		PasswordConfirmation string `json:"password_confirmation"`
	}

	err := ctx.BodyParser(&userRegistrationRequest)
	if err != nil {
		return fmt.Errorf("parsing http user registration request body: %w", err)
	}

	user, err := h.userRegistrator.RegisterUser(ctx.Context(), internal.UserRegistrationRequest{
		UserAuthenticationRequest: internal.UserAuthenticationRequest{
			Username: userRegistrationRequest.Username,
			Password: userRegistrationRequest.Password,
		},
		PasswordConfirmation: userRegistrationRequest.PasswordConfirmation,
	})
	if err != nil {
		return fmt.Errorf("registering user: %w", err)
	}

	return h.userPresenter.Present(ctx, user)
}

func NewUserRegistrationHandler(
	userRegistrator internal.UserRegistrator,
	userPresenter Presenter[internal.User],
) *UserRegistrationHandler {
	return &UserRegistrationHandler{
		userRegistrator: userRegistrator,
		userPresenter:   userPresenter,
	}
}

type UserAuthenticationHandler struct {
	userAuthenticator internal.UserAuthenticator
	userPresenter     Presenter[internal.User]
}

func (h UserAuthenticationHandler) Handle(ctx *fiber.Ctx) error {
	var userAuthenticationRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := ctx.BodyParser(&userAuthenticationRequest)
	if err != nil {
		return fmt.Errorf("parsing http user authentication request body: %w", err)
	}

	user, err := h.userAuthenticator.AuthenticateUser(ctx.Context(), internal.UserAuthenticationRequest{
		Username: userAuthenticationRequest.Username,
		Password: userAuthenticationRequest.Password,
	})
	if err != nil {
		return fmt.Errorf("authenticating user: %w", err)
	}

	return h.userPresenter.Present(ctx, user)
}

func NewUserAuthenticationHandler(
	userAuthenticator internal.UserAuthenticator,
	userPresenter Presenter[internal.User],
) *UserAuthenticationHandler {
	return &UserAuthenticationHandler{
		userAuthenticator: userAuthenticator,
		userPresenter:     userPresenter,
	}
}
