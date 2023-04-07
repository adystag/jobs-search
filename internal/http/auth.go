package http

import (
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

var UserIDContextValue = ContextValueKey{"UserID"}

type ContextValueKey struct {
	s string
}

func (c ContextValueKey) String() string {
	return c.s
}

type JWTAuthenticationMiddleware struct {
	secret []byte
}

func (m JWTAuthenticationMiddleware) Handle(ctx *fiber.Ctx) error {
	authHeader := ctx.GetReqHeaders()[fiber.HeaderAuthorization]

	if !strings.Contains(authHeader, "Bearer") {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	authHeaderComps := strings.SplitN(authHeader, " ", 2)

	if len(authHeaderComps) < 2 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	claims := jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(authHeaderComps[1], &claims, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	userID, err := strconv.ParseInt(claims.Audience, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ctx.Context().SetUserValue(UserIDContextValue, userID)

	return ctx.Next()
}

func NewJWTAuthenticationMiddleware(secret []byte) *JWTAuthenticationMiddleware {
	return &JWTAuthenticationMiddleware{
		secret: secret,
	}
}
