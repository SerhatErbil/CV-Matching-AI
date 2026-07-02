package handlers

import (
	"cv-detector/internal/services"

	"github.com/gofiber/fiber/v2"
)

type MatchCVRequest struct {
	JobDescription string `json:"job_description"`
	CVText         string `json:"cv_text"`
}

func MatchCV(c *fiber.Ctx) error {
	var req MatchCVRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	result := services.CalculateMatch(req.JobDescription, req.CVText)

	return c.JSON(result)
}
