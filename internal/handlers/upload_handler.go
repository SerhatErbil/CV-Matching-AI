package handlers

import (
	"cv-detector/internal/services"

	"github.com/gofiber/fiber/v2"
)

func UploadCV(c *fiber.Ctx) error {
	result, err := services.UploadCV(c)
	if err != nil {
		return err
	}

	return c.JSON(result)
}
