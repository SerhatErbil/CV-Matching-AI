package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

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

	return foundSkills
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
	"golang",
	"react",
	"react native",
	"mongodb",
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
	FileName       string         `json:"file_name"`
	RedFlags       []string       `json:"red_flags"`
	GitHubURL      string         `json:"github_url"`
	GitHubAnalysis GitHubAnalysis `json:"github_analysis"`
	Repos []GitHubRepo  `json:"repos"`

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
	ser := regexp.MustCompile(`https?://github\.com/[^\s]+`)
	githubURL := ser.FindString(cvText)
	return githubURL

}

type GitHubAnalysis struct {
	Username     string   `json:"username"`
	ProfileURL   string   `json:"profile_url"`
	PublicRepos  int      `json:"public_repos"`
	Followers    int      `json:"followers"`
	TotalStars   int      `json:"total_stars"`
	TotalForks   int      `json:"total_forks"`
	TopLanguages []string `json:"top_languages"`
}

type GitHubRepo struct {
	Name            string `json:"name"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	UpdatedAt       string `json:"updated_at"`
}

func analyzeGitHubProfile(githubURL string) (GitHubAnalysis, error) {
	if githubURL == "" {
		return GitHubAnalysis{}, nil
	}
	username := strings.TrimPrefix(githubURL, "https://github.com/")
	username = strings.TrimPrefix(username, "http://github.com/")
	username = strings.Trim(username, "/")

	apiURL := "https://api.github.com/users/" + username

	resp, err := http.Get(apiURL)
	if err != nil {
		return GitHubAnalysis{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GitHubAnalysis{}, fmt.Errorf("github profile could not be fetched")

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

func main() {
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
				githubAnalysis = GitHubAnalysis{}
			}

			repos, err := analyzeGitHubRepositories(githubAnalysis.Username)
			if err != nil {
				repos = []GitHubRepo{}
			}

			fmt.Println("Repo Count:", len(repos))

			candidate := CandidateResult{
				FileName:       file.Filename,
				MatchResult:    result,
				RedFlags:       redFlags,
				GitHubURL:      githubURL,
				GitHubAnalysis: githubAnalysis,
				Repos:          repos,
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
