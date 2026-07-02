package handlers

import (
	"cv-detector/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ExtractSkillsRequest struct {
	Text string `json:"text"`
}

func ExtractSkills(c *fiber.Ctx) error {
	var req ExtractSkillsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	foundSkills := services.ExtractSkills(req.Text)

	return c.JSON(fiber.Map{
		"skills_found": foundSkills,
	})
}
