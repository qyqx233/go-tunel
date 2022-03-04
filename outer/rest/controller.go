package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/qyqx233/go-tunel/outer/pub"
)

func StartRest(addr string) {
	app := fiber.New()
	app.Get("/rest/memStor", func(c *fiber.Ctx) error {
		return c.JSON(&pub.MemStor)
	})
	app.Server().ListenAndServe(addr)
}
