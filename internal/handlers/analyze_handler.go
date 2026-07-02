package handlers

import (
	"cv-detector/internal/services"

	"github.com/gofiber/fiber/v2"
)

func AnalyzeCV(c *fiber.Ctx) error {
	result, err := services.AnalyzeCV(c)
	if err != nil {
		return err
	}

	return c.JSON(result)
}
