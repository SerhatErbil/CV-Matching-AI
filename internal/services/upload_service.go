package services

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func UploadCV(c *fiber.Ctx) (fiber.Map, error) {
	file, err := c.FormFile("cv")
	if err != nil {
		return nil, c.Status(400).JSON(fiber.Map{
			"error": "Failed to get CV file",
		})
	}

	os.MkdirAll("uploads", os.ModePerm)

	filePath := filepath.Join("uploads", file.Filename)

	if err := c.SaveFile(file, filePath); err != nil {
		return nil, c.Status(500).JSON(fiber.Map{
			"error": "Failed to save CV file",
		})
	}

	cvText, err := ReadPdfText(filePath)
	if err != nil {
		return nil, c.Status(500).JSON(fiber.Map{
			"error": "Could not read PDF text",
		})
	}

	foundSkills := ExtractSkills(cvText)

	return fiber.Map{
		"message":        "CV uploaded and read successfully",
		"file_path":      filePath,
		"file_name":      file.Filename,
		"cv_text":        cvText,
		"cv_text_length": len(cvText),
		"skills_found":   foundSkills,
	}, nil
}
