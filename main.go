package main

import (
	"database/sql"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

var db *sql.DB

func extractSkills(text string, skills []string) []string {
	text = strings.ToLower(text)

	foundSkills := []string{}

	for _, skill := range skills {
		skillLower := strings.ToLower(skill)

		pattern := `\b` + regexp.QuoteMeta(skillLower) + `\b`
		re := regexp.MustCompile(pattern)

		if re.MatchString(text) {
			foundSkills = append(foundSkills, skill)
		}
	}
	if containsString(foundSkills, "react native") {
		foundSkills = removeString(foundSkills, "react")
	}

	return foundSkills
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func removeString(items []string, target string) []string {
	result := []string{}

	for _, item := range items {
		if item != target {
			result = append(result, item)
		}
	}

	return result
}

func readPdfText(path string) (string, error) {
	cmd := exec.Command("pdftotext", path, "-")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func generateAIComment(prompt string) (string, error) {
	requestBody := fiber.Map{
		"model":  "llama3.2:3b",
		"prompt": prompt,
		"stream": false,
		"options": fiber.Map{
			"temperature": 0.2,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Response, nil
}

var skills = []string{

	"react",
	"golang",
	"mongodb",
	"react native",
	"docker",
	"python",
	"sql",
	"java",
	"c#",
	"kubernetes",
	"machine learning",
	"nodejs",
	"c++",
}

type MatchResult struct {
	RequiredSkills []string `json:"required_skills"`
	MatchedSkills  []string `json:"matched_skills"`
	MissingSkills  []string `json:"missing_skills"`
	ExtraSkills    []string `json:"extra_skills"`
	MatchScore     int      `json:"match_score"`
	Summary        string   `json:"summary"`
}

type CandidateResult struct {
	FileName            string         `json:"file_name"`
	RedFlags            []string       `json:"red_flags"`
	GitHubURL           string         `json:"github_url"`
	GitHubAnalysis      GitHubAnalysis `json:"github_analysis"`
	Repos               []GitHubRepo   `json:"repos"`
	FinalScore          int            `json:"final_score"`
	FinalRecommendation string         `json:"final_recommendation"`
	AICandidateSummary  string         `json:"ai_candidate_summary"`

	MatchResult
}

func calculateMatch(jobDescription string, cvText string) MatchResult {
	requiredSkills := extractSkills(jobDescription, skills)
	cvSkills := extractSkills(cvText, skills)

	matchedSkill := []string{}
	missingSkills := []string{}
	extraSkills := []string{}

	for _, skill := range requiredSkills {
		if containsSkill(cvText, skill) {
			matchedSkill = append(matchedSkill, skill)
		} else {
			missingSkills = append(missingSkills, skill)
		}
	}

	for _, skill := range cvSkills {
		if !containsSkill(jobDescription, skill) {
			extraSkills = append(extraSkills, skill)
		}
	}

	matchScore := 0
	if len(requiredSkills) > 0 {
		matchScore = len(matchedSkill) * 100 / len(requiredSkills)
	}

	summary := ""
	if matchScore >= 70 {
		summary = "The candidate is a strong match for the job description"
	} else if matchScore >= 40 {
		summary = "The candidate is a partial match for the job description"
	} else {
		summary = "The candidate does not meet most of the required criteria"
	}

	return MatchResult{
		RequiredSkills: requiredSkills,
		MatchedSkills:  matchedSkill,
		MissingSkills:  missingSkills,
		ExtraSkills:    extraSkills,
		MatchScore:     matchScore,
		Summary:        summary,
	}

}

func containsSkill(text string, skill string) bool {
	text = strings.ToLower(text)
	skill = strings.ToLower(skill)

	pattern := `\b` + regexp.QuoteMeta(skill) + `\b`
	re := regexp.MustCompile(pattern)

	return re.MatchString(text)
}

func detectRedFlags(cvText string, result MatchResult) []string {
	redFlags := []string{}

	if len(result.MissingSkills) > 0 {
		redFlags = append(redFlags, "İş ilanındaki bazı zorunlu teknik yetenekler CV'de bulunamadı.")
	}

	if !strings.Contains(strings.ToLower(cvText), "github.com") {
		redFlags = append(redFlags, "GitHub bağlantısı bulunamadı.")
	}

	if len(cvText) < 600 {
		redFlags = append(redFlags, "CV içeriği kısa veya yetersiz görünüyor.")
	}

	if !strings.Contains(strings.ToLower(cvText), "proje") && !strings.Contains(strings.ToLower(cvText), "project") {
		redFlags = append(redFlags, "CV'de proje bilgisi bulunamadı.")
	}

	return redFlags
}

func extractGitHubURL(cvText string) string {
	ser := regexp.MustCompile(`(?i)(https?://)?(www\.)?github\.com/[a-zA-Z0-9-]+`)
	match := ser.FindString(cvText)

	if match == "" {
		return ""
	}

	if !strings.HasPrefix(strings.ToLower(match), "http") {
		match = "https://" + match
	}

	return match
}

type GitHubAnalysis struct {
	Username               string                   `json:"username"`
	ProfileURL             string                   `json:"profile_url"`
	PublicRepos            int                      `json:"public_repos"`
	Followers              int                      `json:"followers"`
	TotalStars             int                      `json:"total_stars"`
	TotalForks             int                      `json:"total_forks"`
	TopLanguages           []string                 `json:"top_languages"`
	GitHubScore            int                      `json:"github_score"`
	AIComment              string                   `json:"ai_comment"`
	RepositoryIntelligence []RepositoryIntelligence `json:"repository_intelligence"`
}

type GitHubRepo struct {
	Name            string `json:"name"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	UpdatedAt       string `json:"updated_at"`
	DefaultBranch   string `json:"default_branch"`
}

func analyzeGitHubProfile(githubURL string) (GitHubAnalysis, error) {
	if githubURL == "" {
		return GitHubAnalysis{}, nil
	}
	username := strings.TrimPrefix(githubURL, "https://github.com/")
	username = strings.TrimPrefix(username, "http://github.com/")
	username = strings.Trim(username, "/")
	username = strings.TrimSpace(username)

	apiURL := "https://api.github.com/users/" + username

	resp, err := http.Get(apiURL)
	if err != nil {
		return GitHubAnalysis{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return GitHubAnalysis{}, fmt.Errorf(
			"GitHub API Error %d: %s",
			resp.StatusCode,
			string(body),
		)
	}
	var result struct {
		Login       string `json:"login"`
		HTMLURL     string `json:"html_url"`
		PublicRepos int    `json:"public_repos"`
		Followers   int    `json:"followers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return GitHubAnalysis{}, err
	}

	fmt.Println("GitHub URL:", githubURL)
	fmt.Println("Username:", username)

	return GitHubAnalysis{
		Username:    result.Login,
		ProfileURL:  result.HTMLURL,
		PublicRepos: result.PublicRepos,
		Followers:   result.Followers,
	}, nil
}

func analyzeGitHubRepositories(username string) ([]GitHubRepo, error) {
	apiURL := "https://api.github.com/users/" + username + "/repos"

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github repositories couldn't be fetched")
	}
	var repos []GitHubRepo

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func calculateGitHubAnalysis(githubAnalysis GitHubAnalysis, repos []GitHubRepo) GitHubAnalysis {
	totalStars := 0
	totalForks := 0
	languageCounts := map[string]int{}

	for _, repo := range repos {
		totalStars += repo.StargazersCount
		totalForks += repo.ForksCount

		if repo.Language != "" {
			languageCounts[repo.Language]++
		}
	}
	githubAnalysis.TotalStars = totalStars
	githubAnalysis.TotalForks = totalForks

	topLanguages := []string{}

	for language := range languageCounts {
		topLanguages = append(topLanguages, language)
	}

	githubAnalysis.TopLanguages = topLanguages

	score := 0
	repoCount := len(repos)
	if repoCount >= 5 {
		score += 10
	} else if repoCount >= 3 {
		score += 7
	} else if repoCount >= 1 {
		score += 4
	}
	languageCount := len(githubAnalysis.TopLanguages)
	if languageCount >= 4 {
		score += 25
	} else if languageCount == 3 {
		score += 20
	} else if languageCount == 2 {
		score += 15
	} else if languageCount == 1 {
		score += 8
	}

	if githubAnalysis.TotalStars >= 10 || githubAnalysis.TotalForks >= 5 {
		score += 15
	} else if githubAnalysis.TotalStars >= 3 || githubAnalysis.TotalForks >= 2 {
		score += 10
	} else if githubAnalysis.TotalStars >= 1 || githubAnalysis.TotalForks >= 1 {
		score += 5
	}

	score += calculateUpdateScore(repos)

	githubAnalysis.GitHubScore = score

	return githubAnalysis
}

func calculateFinalCandidateScore(
	matchScore int,
	githubScore int,
	redFlags []string,
	repositoryIntelligence []RepositoryIntelligence,
) int {
	score := 0

	score += matchScore * 60 / 100
	score += githubScore * 25 / 100

	repoScore := 0

	if len(repositoryIntelligence) > 0 {
		totalRepoScore := 0

		for _, repo := range repositoryIntelligence {
			totalRepoScore += repo.RepositoryScore
		}

		repoScore = totalRepoScore / len(repositoryIntelligence)
	}

	score += repoScore * 15 / 100

	redFlagPenalty := len(redFlags) * 5

	if redFlagPenalty > 20 {
		redFlagPenalty = 20
	}

	score -= redFlagPenalty

	if score < 0 {
		score = 0
	}

	if score > 100 {
		score = 100
	}

	return score
}

func generateFinalRecommendation(finalScore int) string {
	if finalScore >= 70 {
		return "Strong candidate. Recommended for interview."
	}

	if finalScore >= 50 {
		return "Potential candidate. Consider for interview after manual review."
	}

	return "Weak match. Not recommended for interview at this stage."
}

func generateGitHubAIComment(githubAnalysis GitHubAnalysis, repos []GitHubRepo) string {
	if githubAnalysis.Username == "" {
		return "An evaluation could not be performed because the candidate's GitHub profile was not found in their resume."
	}

	repoSummaries := []string{}

	for _, repo := range repos {
		repoSummaries = append(repoSummaries, fmt.Sprintf(
			"Repository: %s, Language: %s, Stars: %d, Forks: %d, Last Update: %s",
			repo.Name,
			repo.Language,
			repo.StargazersCount,
			repo.ForksCount,
			repo.UpdatedAt,
		))
	}

	prompt := fmt.Sprintf(`
You are an experienced software engineer specializing in technical recruiting.

Evaluate the following GitHub profile:
Username: %s
Profile URL: %s
Number of public repos: %d
Number of followers: %d
Total stars: %d
Total forks: %d
Languages ​​used: %v
GitHub score: %d

Repo details:
%v

Write a professional evaluation in English consisting of 4–6 concise sentences.
Focus on:
- Repository quality
- Project diversity
- Technology stack
- Activity level
- Overall impression from a technical recruiter's perspective

Do not make assumptions.
Base your evaluation only on the provided data.
Keep the tone objective and constructive.
`, githubAnalysis.Username, githubAnalysis.ProfileURL, githubAnalysis.PublicRepos,
		githubAnalysis.Followers, githubAnalysis.TotalStars, githubAnalysis.TotalForks,
		githubAnalysis.TopLanguages, githubAnalysis.GitHubScore, repoSummaries)

	comment, err := generateAIComment(prompt)
	if err != nil {
		return "GitHub için AI yorumu oluşturulamadı."
	}

	return comment
}

func generateCandidateAISummary(
	jobDescription string,
	result MatchResult,
	githubAnalysis GitHubAnalysis,
	redFlags []string,
	finalScore int,
	finalRecommendation string,
) string {
	repoSummaries := []string{}

	for _, repo := range githubAnalysis.RepositoryIntelligence {
		repoSummaries = append(repoSummaries, fmt.Sprintf(
			"Repository: %s, Score: %d, Architecture: %v, Quality: %v, Improvements: %v",
			repo.RepositoryName,
			repo.RepositoryScore,
			repo.ArchitectureSignals,
			repo.QualitySignals,
			repo.ImprovementAreas,
		))
	}

	prompt := fmt.Sprintf(`
You are a senior technical recruiter.

Your task is NOT to calculate a new score.
Your task is ONLY to explain the backend-calculated hiring decision.

IMPORTANT RULES:
- Do not create a new score.
- Do not disagree with the backend decision.
- Do not change the final recommendation.
- Stay fully consistent with the Final Candidate Score and Final Recommendation.
- Use only the provided backend data.
- Do not invent experience, skills, projects, or achievements.
- If something is missing, mention it as a limitation.
- Use professional recruiter language.
- Write in English.
- Write exactly 4-6 concise sentences.
- No bullet points.
- No headings.
- Do not mention that you are an AI.

Never use phrases like:
- I think
- I believe
- I agree
- I concur
- In my opinion

Do not express personal judgement.

Only explain why the backend reached its decision.

Never mention any weakness unless it exists inside:
- Missing Skills
- Red Flags
- Repository Intelligence

Never infer missing experience.

Never assume technologies.

Never speculate.

CRITICAL INSTRUCTIONS

You are NOT allowed to reinterpret backend data.

Do NOT summarize lists incorrectly.

If a skill appears inside Missing Skills,
you must never describe it as matched.

If a skill appears inside Matched Skills,
you must never describe it as missing.

Do not compare or reinterpret the backend evaluation.

Treat every backend field as absolute truth.

Treat Matched Skills, Missing Skills, Extra Skills, Red Flags, and Repository Intelligence as independent categories.

Never merge information between categories.

Never move an item from one category to another.

Never state or imply that a missing skill is matched.

Never state or imply that a matched skill is missing.

Only reference skills that explicitly exist inside the provided backend lists.

Do not mention technologies that are not present.

Do not infer additional technical abilities.

When explaining the backend decision:

- Start with the overall evaluation.
- Mention the strongest matched skills.
- Mention only confirmed missing skills.
- Mention repository findings only if they exist.
- Finish by reinforcing the backend's final recommendation.

Do not force every category into the explanation if it is empty.

Avoid generic openings such as:
- Based on the provided data
- Based on the provided evaluation
- Overall assessment
- The candidate is evaluated as

Do not explain what a missing skill means.

Simply state that it was not identified.

Example:

Correct:
React Native experience was not identified in the CV.

Incorrect:
React Native may not be fully developed.

Do not minimize or exaggerate backend findings.

Describe limitations objectively without judging their impact unless explicitly stated by the backend.

Backend Evaluation Data:
Job Description: %s

Backend Match Score: %d/100
Backend GitHub Score: %d/100
Backend Final Candidate Score: %d/100
Backend Final Recommendation: %s

Matched Skills: %v
Missing Skills: %v
Extra Skills: %v
Red Flags: %v
Repository Intelligence: %v

Write one natural recruiter-style paragraph that explains the candidate's strengths, limitations, the reason behind the backend's decision, and the interview recommendation while staying fully consistent with the backend's final recommendation.
The explanation must sound like a professional ATS recruitment report rather than a general AI summary.
`,
		jobDescription,
		result.MatchScore,
		githubAnalysis.GitHubScore,
		finalScore,
		finalRecommendation,
		result.MatchedSkills,
		result.MissingSkills,
		result.ExtraSkills,
		redFlags,
		repoSummaries,
	)

	comment, err := generateAIComment(prompt)
	if err != nil {
		return "Candidate AI summary could not be generated."
	}

	return comment
}

func calculateUpdateScore(repos []GitHubRepo) int {
	if len(repos) == 0 {
		return 0
	}
	latestUpdate := time.Time{}

	for _, repo := range repos {
		parsedTime, err := time.Parse(time.RFC3339, repo.UpdatedAt)
		if err != nil {
			continue
		}

		if parsedTime.After(latestUpdate) {
			latestUpdate = parsedTime
		}
	}
	daysSinceUpdate := time.Since(latestUpdate).Hours() / 24

	if daysSinceUpdate <= 30 {
		return 20
	}

	if daysSinceUpdate <= 90 {
		return 15
	}

	if daysSinceUpdate <= 180 {
		return 10
	}

	return 0
}

func connectDB() {
	var err error

	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=cv_matching_ai sslmode=disable"

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Database ping error:", err)
	}

	log.Println("PostgreSQL connected successfully")
}

func saveAnalysis(result CandidateResult, jobDescription string) error {

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

func main() {
	connectDB()
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "cv-detector",
		})
	})

	app.Post("/extract-skills", func(c *fiber.Ctx) error {
		type Request struct {
			Text string `json:"text"`
		}

		var req Request

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		foundSkills := extractSkills(req.Text, skills)

		return c.JSON(fiber.Map{
			"skills_found": foundSkills,
		})
	})
	app.Get("/analysis-results", func(c *fiber.Ctx) error {
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
	})

	app.Get("/analysis-results/:id", func(c *fiber.Ctx) error {

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
	})

	app.Post("/match-cv", func(c *fiber.Ctx) error {
		type Request struct {
			JobDescription string `json:"job_description"`
			CVText         string `json:"cv_text"`
		}

		var req Request

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		result := calculateMatch(req.JobDescription, req.CVText)

		return c.JSON(result)
	})

	app.Post("/upload-cv", func(c *fiber.Ctx) error {
		file, err := c.FormFile("cv")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Failed to get CV file",
			})
		}

		os.MkdirAll("uploads", os.ModePerm)

		filePath := filepath.Join("uploads", file.Filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to save CV file",
			})
		}
		cvText, err := readPdfText(filePath)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not read PDF text",
			})
		}
		foundSkills := extractSkills(cvText, skills)

		return c.JSON(fiber.Map{
			"message":        "CV uploaded and read successfully",
			"file_path":      filePath,
			"file_name":      file.Filename,
			"cv_text":        cvText,
			"cv_text_length": len(cvText),
			"skills_found":   foundSkills,
		})
	})
	app.Post("/analyze-cv", func(c *fiber.Ctx) error {
		jobDescription := c.FormValue("job_description")

		if jobDescription == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Job description is required",
			})
		}

		file, err := c.FormFile("cv")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "CV file is required",
			})
		}

		os.MkdirAll("uploads", os.ModePerm)

		filePath := filepath.Join("uploads", file.Filename)

		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to save CV file",
			})
		}

		cvText, err := readPdfText(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not read PDF text",
			})
		}

		result := calculateMatch(jobDescription, cvText)
		cvSkills := extractSkills(cvText, skills)

		prompt := fmt.Sprintf(`
        You are an experienced technical recruiter.

        Analyze both the job criteria and the full CV content. Do not rely only on the extracted skills.

		Comment on how any criteria I'm not looking for in candidates could benefit us.

		State whether the candidate should be invited to an interview.

        jobDescription: %s
		cvText: %s

		result.MatchScore: %d

        Required Skills: %v
        Matched Skills: %v
        Missing Skills: %v
        Extra Skills: %v

       Write 5-8 concise sentences in Turkish. Evaluate the candidate's technical skills, relevant experience, strengths and weaknesses, and overall suitability for the role. State whether the candidate will be invited for an interview and explain the reason. Do not add bullet points, headings, or point distributions.

	   Important:
       Write only in Turkish.
       Do not use English words unless they are technology names.
       Do not mix languages.
       Use professional HR language.

	   If information is not mentioned in the CV, say that it is not mentioned. Do not make assumptions.
       `, jobDescription, cvText, result.MatchScore, result.RequiredSkills, result.MatchedSkills, result.MissingSkills, result.ExtraSkills)

		fmt.Println(prompt)
		aiComment, err := generateAIComment(prompt)
		if err != nil {
			aiComment = "AI evaluation could not be generated."
		}

		return c.JSON(fiber.Map{
			"file_name":       file.Filename,
			"required_skills": result.RequiredSkills,
			"cv_skills":       cvSkills,
			"matched_skills":  result.MatchedSkills,
			"missing_skills":  result.MissingSkills,
			"extra_skills":    result.ExtraSkills,
			"match_score":     result.MatchScore,
			"summary":         result.Summary,
			"ai_comment":      aiComment,
		})
	})

	app.Post("/analyze-multiple-cvs", func(c *fiber.Ctx) error {
		jobDescription := c.FormValue("job_description")

		if jobDescription == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Job description is required",
			})
		}

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid form date",
			})
		}
		files := form.File["cvs"]
		if len(files) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Upload at least one CV",
			})
		}

		fileNames := []string{}
		results := []CandidateResult{}

		os.MkdirAll("uploads", os.ModePerm)

		for _, file := range files {
			fileNames = append(fileNames, file.Filename)

			filePath := filepath.Join("uploads", file.Filename)

			if err := c.SaveFile(file, filePath); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to save CV File",
				})
			}
			cvText, err := readPdfText(filePath)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Could not read PDF text",
				})
			}

			result := calculateMatch(jobDescription, cvText)

			redFlags := detectRedFlags(cvText, result)

			githubURL := extractGitHubURL(cvText)

			githubAnalysis, err := analyzeGitHubProfile(githubURL)
			if err != nil {
				fmt.Println("GitHub Profile Error:", err)
				githubAnalysis = GitHubAnalysis{}
			}

			repos := []GitHubRepo{}

			if githubAnalysis.Username != "" {
				repos, err = analyzeGitHubRepositories(githubAnalysis.Username)
				if err != nil {
					fmt.Println("GitHub Repositories Error:", err)
					repos = []GitHubRepo{}
				}
			}

			fmt.Println("Repo Count:", len(repos))

			selectedRepos := selectRepositoriesForAnalysis(repos)

			repositoryIntelligence := []RepositoryIntelligence{}

			for _, selectedRepo := range selectedRepos {

				treeItems, err := fetchRepositoryTree(githubAnalysis.Username, selectedRepo)
				if err != nil {
					continue
				}

				intelligence := analyzeRepositoryTree(
					selectedRepo,
					treeItems,
					"Selected for repository analysis",
				)

				repositoryIntelligence = append(repositoryIntelligence, intelligence)
			}

			githubAnalysis = calculateGitHubAnalysis(githubAnalysis, repos)

			githubAnalysis.RepositoryIntelligence = repositoryIntelligence

			githubAnalysis.AIComment = ""
			finalScore := calculateFinalCandidateScore(
				result.MatchScore,
				githubAnalysis.GitHubScore,
				redFlags,
				githubAnalysis.RepositoryIntelligence,
			)

			finalRecommendation := generateFinalRecommendation(finalScore)

			aiCandidateSummary := generateCandidateAISummary(
				jobDescription,
				result,
				githubAnalysis,
				redFlags,
				finalScore,
				finalRecommendation,
			)

			candidate := CandidateResult{
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
			err = saveAnalysis(candidate, jobDescription)
			if err != nil {
				fmt.Println("Database Save Error:", err)
			}

			results = append(results, candidate)
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].MatchScore > results[j].MatchScore
		})

		return c.JSON(fiber.Map{
			"message":         "Multiple CVs received successfully",
			"job_description": jobDescription,
			"file_count":      len(files),
			"file_names":      fileNames,
			"results":         results,
		})
	})

	app.Listen(":3000")
}
