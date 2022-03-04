package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/qyqx233/go-tunel/outer/pub"
	"github.com/rs/zerolog/log"
)

func StartRest(addr string) {
	log.Debug().Str("addr", addr).Msg("监听地址")
	app := fiber.New()
	app.Get("/rest/memStor", func(c *fiber.Ctx) error {
		return c.JSON(&pub.MemStor)
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello")
	})
	err := app.Listen(addr)
	if err != nil {
		log.Error().Err(err).Msg("fiber listen error")
	}
}
