package handlers

import (
	"cv-detector/internal/services"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func AnalyzeMultipleCVs(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		result, err := services.AnalyzeMultipleCVs(c, db)
		if err != nil {
			return err
		}

		return c.JSON(result)
	}
}
