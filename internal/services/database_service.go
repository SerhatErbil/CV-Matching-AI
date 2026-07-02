package services

import (
	"cv-detector/internal/models"
	"database/sql"
	"github.com/lib/pq"
	"log"
	"fmt"
	"os"
)

func SaveAnalysis(db *sql.DB, result models.CandidateResult, jobDescription string) error {
	_, err := db.Exec(`
		INSERT INTO cv_analysis_results (
			file_name,
			job_description,
			required_skills,
			matched_skills,
			missing_skills,
			match_score,
			github_url,
			github_score,
			github_ai_comment,
			final_score,
			final_recommendation
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
	`,
		result.FileName,
		jobDescription,
		pq.Array(result.RequiredSkills),
		pq.Array(result.MatchedSkills),
		pq.Array(result.MissingSkills),
		result.MatchScore,
		result.GitHubURL,
		result.GitHubAnalysis.GitHubScore,
		result.GitHubAnalysis.AIComment,
		result.FinalScore,
		result.FinalRecommendation,
	)

	return err
}

func ConnectDB() *sql.DB {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "cv_matching_ai")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database ping error:", err)
	}

	log.Println("PostgreSQL connected successfully")

	return db
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}