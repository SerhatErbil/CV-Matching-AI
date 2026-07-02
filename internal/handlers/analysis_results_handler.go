package handlers

import (
	"database/sql"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetAnalysisResults(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(`
		SELECT 
			id,
			file_name,
			match_score,
			final_score,
			final_recommendation,
			created_at
		FROM cv_analysis_results
		ORDER BY final_score DESC
	`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch analysis results",
			})
		}
		defer rows.Close()

		type AnalysisResultResponse struct {
			ID                  int       `json:"id"`
			FileName            string    `json:"file_name"`
			MatchScore          int       `json:"match_score"`
			FinalScore          int       `json:"final_score"`
			FinalRecommendation string    `json:"final_recommendation"`
			CreatedAt           time.Time `json:"created_at"`
		}

		results := []AnalysisResultResponse{}

		for rows.Next() {
			var result AnalysisResultResponse

			err := rows.Scan(
				&result.ID,
				&result.FileName,
				&result.MatchScore,
				&result.FinalScore,
				&result.FinalRecommendation,
				&result.CreatedAt,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to scan analysis result",
				})
			}

			results = append(results, result)
		}

		return c.JSON(fiber.Map{
			"count":   len(results),
			"results": results,
		})
	}
}

func GetAnalysisResultByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")

		type AnalysisResultResponse struct {
			ID                  int       `json:"id"`
			FileName            string    `json:"file_name"`
			JobDescription      string    `json:"job_description"`
			MatchScore          int       `json:"match_score"`
			FinalScore          int       `json:"final_score"`
			FinalRecommendation string    `json:"final_recommendation"`
			CreatedAt           time.Time `json:"created_at"`
		}

		var result AnalysisResultResponse

		err := db.QueryRow(`
		SELECT
			id,
			file_name,
			job_description,
			match_score,
			final_score,
			final_recommendation,
			created_at
		FROM cv_analysis_results
		WHERE id = $1
	`, id).Scan(
			&result.ID,
			&result.FileName,
			&result.JobDescription,
			&result.MatchScore,
			&result.FinalScore,
			&result.FinalRecommendation,
			&result.CreatedAt,
		)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Analysis result not found",
			})
		}

		return c.JSON(result)
	}
}

func ExportAnalysisResultsJSON(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		rows, err := db.Query(`
		SELECT
			id,
			file_name,
			job_description,
			match_score,
			final_score,
			final_recommendation,
			created_at
		FROM cv_analysis_results
		ORDER BY created_at DESC
	`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to export analysis results",
			})
		}
		defer rows.Close()

		type ExportResult struct {
			ID                  int       `json:"id"`
			FileName            string    `json:"file_name"`
			JobDescription      string    `json:"job_description"`
			MatchScore          int       `json:"match_score"`
			FinalScore          int       `json:"final_score"`
			FinalRecommendation string    `json:"final_recommendation"`
			CreatedAt           time.Time `json:"created_at"`
		}

		results := []ExportResult{}

		for rows.Next() {
			var result ExportResult

			err := rows.Scan(
				&result.ID,
				&result.FileName,
				&result.JobDescription,
				&result.MatchScore,
				&result.FinalScore,
				&result.FinalRecommendation,
				&result.CreatedAt,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to scan export result",
				})
			}

			results = append(results, result)
		}

		return c.JSON(fiber.Map{
			"export_type": "json",
			"count":       len(results),
			"results":     results,
		})
	}
}

func ExportAnalysisResultsCSV(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		rows, err := db.Query(`
		SELECT
			id,
			file_name,
			match_score,
			final_score,
			final_recommendation,
			created_at
		FROM cv_analysis_results
		ORDER BY final_score DESC
	`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch analysis results",
			})
		}
		defer rows.Close()

		os.MkdirAll("exports", os.ModePerm)

		filePath := filepath.Join("exports", "analysis_results.csv")

		csvFile, err := os.Create(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create CSV file",
			})
		}
		defer csvFile.Close()

		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		writer.Write([]string{
			"ID",
			"File Name",
			"Match Score",
			"Final Score",
			"Final Recommendation",
			"Created At",
		})

		for rows.Next() {
			var (
				id                  int
				fileName            string
				matchScore          int
				finalScore          int
				finalRecommendation string
				createdAt           time.Time
			)

			err := rows.Scan(
				&id,
				&fileName,
				&matchScore,
				&finalScore,
				&finalRecommendation,
				&createdAt,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to scan row",
				})
			}

			writer.Write([]string{
				strconv.Itoa(id),
				fileName,
				strconv.Itoa(matchScore),
				strconv.Itoa(finalScore),
				finalRecommendation,
				createdAt.Format(time.RFC3339),
			})
		}

		return c.JSON(fiber.Map{
			"message": "CSV exported successfully",
			"file":    filePath,
		})
	}
}
