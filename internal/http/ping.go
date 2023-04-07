package http

import "github.com/gofiber/fiber/v2"

type PingHandler struct{}

func (PingHandler) Handle(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "pong",
	})
}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}
