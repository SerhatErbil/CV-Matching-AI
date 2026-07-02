package services

import (
	"cv-detector/internal/models"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gofiber/fiber/v2"
)

func AnalyzeMultipleCVs(c *fiber.Ctx, db *sql.DB) (fiber.Map, error) {
	jobDescription := c.FormValue("job_description")
	parsedJobDescription := ParseJobDescription(jobDescription)

	if jobDescription == "" {
		return nil, c.Status(400).JSON(fiber.Map{
			"error": "Job description is required",
		})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, c.Status(400).JSON(fiber.Map{
			"error": "Invalid form date",
		})
	}

	files := form.File["cvs"]
	if len(files) == 0 {
		return nil, c.Status(400).JSON(fiber.Map{
			"error": "Upload at least one CV",
		})
	}

	fileNames := []string{}
	results := []models.CandidateResult{}

	os.MkdirAll("uploads", os.ModePerm)

	for _, file := range files {
		fileNames = append(fileNames, file.Filename)

		filePath := filepath.Join("uploads", file.Filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return nil, c.Status(500).JSON(fiber.Map{
				"error": "Failed to save CV File",
			})
		}

		cvText, err := ReadPdfText(filePath)
		if err != nil {
			return nil, c.Status(500).JSON(fiber.Map{
				"error": "Could not read PDF text",
			})
		}

		result := CalculateMatch(jobDescription, cvText)
		redFlags := DetectRedFlags(cvText, result)
		githubURL := ExtractGitHubURL(cvText)

		githubAnalysis, err := AnalyzeGitHubProfile(githubURL)
		if err != nil {
			fmt.Println("GitHub Profile Error:", err)
			githubAnalysis = models.GitHubAnalysis{}
		}

		repos := []models.GitHubRepo{}

		if githubAnalysis.Username != "" {
			repos, err = AnalyzeGitHubRepositories(githubAnalysis.Username)
			if err != nil {
				fmt.Println("GitHub Repositories Error:", err)
				repos = []models.GitHubRepo{}
			}
		}

		selectedRepos := SelectRepositoriesForAnalysis(repos)

		repositoryIntelligence := []models.RepositoryIntelligence{}

		for _, selectedRepo := range selectedRepos {
			treeItems, err := FetchRepositoryTree(githubAnalysis.Username, selectedRepo)
			if err != nil {
				continue
			}

			intelligence := AnalyzeRepositoryTree(
				selectedRepo,
				treeItems,
				"Selected for repository analysis",
			)

			repositoryIntelligence = append(repositoryIntelligence, intelligence)
		}

		githubAnalysis = CalculateGitHubAnalysis(githubAnalysis, repos)
		githubAnalysis.RepositoryIntelligence = repositoryIntelligence
		githubAnalysis.AIComment = ""

		finalScore := CalculateFinalCandidateScore(
			result.MatchScore,
			githubAnalysis.GitHubScore,
			redFlags,
			githubAnalysis.RepositoryIntelligence,
		)

		finalRecommendation := GenerateFinalRecommendation(finalScore)

		aiCandidateSummary := GenerateCandidateAISummary(
			jobDescription,
			parsedJobDescription,
			result,
			githubAnalysis,
			redFlags,
			finalScore,
			finalRecommendation,
		)

		candidate := models.CandidateResult{
			FileName:            file.Filename,
			MatchResult:         result,
			RedFlags:            redFlags,
			GitHubURL:           githubURL,
			GitHubAnalysis:      githubAnalysis,
			Repos:               repos,
			FinalScore:          finalScore,
			FinalRecommendation: finalRecommendation,
			AICandidateSummary:  aiCandidateSummary,
		}

		err = SaveAnalysis(db, candidate, jobDescription)
		if err != nil {
			fmt.Println("Database Save Error:", err)
		}

		results = append(results, candidate)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})

	return fiber.Map{
		"message":                "Multiple CVs received successfully",
		"job_description":        jobDescription,
		"file_count":             len(files),
		"file_names":             fileNames,
		"results":                results,
		"parsed_job_description": parsedJobDescription,
	}, nil
}
