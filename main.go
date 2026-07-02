package main

import (
	"database/sql"

	"cv-detector/internal/handlers"
	"cv-detector/internal/services"

	"github.com/gofiber/fiber/v2"
)

var db *sql.DB

func main() {
	db = services.ConnectDB()
	app := fiber.New()

	app.Get("/health", handlers.Health)
	app.Post("/extract-skills", handlers.ExtractSkills)
	app.Get("/analysis-results", handlers.GetAnalysisResults(db))
	app.Get("/analysis-results/:id", handlers.GetAnalysisResultByID(db))
	app.Get("/analysis-results/export/json", handlers.ExportAnalysisResultsJSON(db))
	app.Get("/analysis-results/export/csv", handlers.ExportAnalysisResultsCSV(db))
	app.Post("/match-cv", handlers.MatchCV)
	app.Post("/upload-cv", handlers.UploadCV)
	app.Post("/analyze-cv", handlers.AnalyzeCV)
	app.Post("/analyze-multiple-cvs", handlers.AnalyzeMultipleCVs(db))
	app.Listen(":3000")
}
