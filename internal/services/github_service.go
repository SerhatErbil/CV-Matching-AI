package services

import (
	"cv-detector/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func SelectRepositoriesForAnalysis(repos []models.GitHubRepo) []models.GitHubRepo {

	// 1. Repo yoksa boş slice dön
	if len(repos) == 0 {
		return []models.GitHubRepo{}
	}

	selectedRepos := []models.GitHubRepo{}

	// 2. En güncel repoyu bul
	mostRecent := repos[0]

	for _, repo := range repos {

		if repo.UpdatedAt > mostRecent.UpdatedAt {

			mostRecent = repo

		}

	}

	selectedRepos = append(selectedRepos, mostRecent)

	// 3. En çok star alan repoyu bul

	mostStarred := repos[0]

	for _, repo := range repos {

		if repo.StargazersCount > mostStarred.StargazersCount {

			mostStarred = repo

		}

	}

	if mostRecent.Name != mostStarred.Name {

		selectedRepos = append(selectedRepos, mostStarred)

	}

	return selectedRepos
}

func FetchRepositoryTree(username string, repo models.GitHubRepo) ([]models.GitHubTreeItem, error) {
	branch := repo.DefaultBranch

	if branch == "" {
		branch = "main"
	}

	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1",
		username,
		repo.Name,
		branch,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return []models.GitHubTreeItem{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []models.GitHubTreeItem{}, fmt.Errorf("failed to fetch repository tree")
	}

	var treeResponse models.GitHubTreeResponse

	if err := json.NewDecoder(resp.Body).Decode(&treeResponse); err != nil {
		return []models.GitHubTreeItem{}, err
	}

	return treeResponse.Tree, nil
}

func AnalyzeRepositoryTree(repo models.GitHubRepo, treeItems []models.GitHubTreeItem, selectionReason string) models.RepositoryIntelligence {
	importantFiles := []string{
		"README.md",
		"LICENSE",
		"Dockerfile",
		"docker-compose.yml",
		".env.example",
		".gitignore",
	}

	languageFiles := []string{
		"go.mod",
		"go.sum",
		"package-lock.json",
		"package.json",
		"requirements.txt",
		"pyproject.toml",
		"tsconfig.json",
		"Package.swift",
	}
	importantFilesFound := []string{}
	importantFilesMissing := []string{}
	languageFilesFound := []string{}

	for _, item := range treeItems {
		if item.Type != "blob" {
			continue
		}

		path := item.Path

		if containsFile(importantFiles, path) {
			importantFilesFound = append(importantFilesFound, path)
		}

		if containsFile(languageFiles, path) {
			languageFilesFound = append(languageFilesFound, path)
		}
	}
	for _, file := range importantFiles {

		if !containsFile(importantFilesFound, file) {

			importantFilesMissing = append(
				importantFilesMissing,
				file,
			)

		}

	}

	score := 0

	architectureSignals := []string{}
	testingSignals := []string{}
	deploymentSignals := []string{}
	qualitySignals := []string{}

	if containsFile(importantFilesFound, "README.md") {
		score += 15
	}

	if containsFile(importantFilesFound, ".gitignore") {
		score += 10
	}

	if containsFile(importantFilesFound, "LICENSE") {
		score += 10
	}

	if len(languageFilesFound) > 0 {
		score += 15
	}

	if containsFile(importantFilesFound, "Dockerfile") {
		score += 15
	}

	if containsFile(importantFilesFound, "docker-compose.yml") {
		score += 10
	}

	if containsFile(importantFilesFound, ".env.example") {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	if containsFile(importantFilesFound, "README.md") {
		qualitySignals = append(
			qualitySignals,
			"Project includes a README for documentation.",
		)
	}

	if containsFile(importantFilesFound, ".gitignore") {
		qualitySignals = append(
			qualitySignals,
			"Repository contains a .gitignore file.",
		)
	}

	if containsFile(importantFilesFound, "LICENSE") {
		qualitySignals = append(
			qualitySignals,
			"Repository includes an open-source license.",
		)
	}

	if len(languageFilesFound) > 0 {
		qualitySignals = append(
			qualitySignals,
			"Dependency management files are present.",
		)
	}

	if containsFile(importantFilesFound, "Dockerfile") {
		deploymentSignals = append(
			deploymentSignals,
			"Repository includes a Dockerfile for containerized deployment.",
		)
	}

	if containsFile(importantFilesFound, "docker-compose.yml") {
		deploymentSignals = append(
			deploymentSignals,
			"Repository includes docker-compose.yml for local multi-service setup.",
		)
	}

	if containsFile(importantFilesFound, ".env.example") {
		deploymentSignals = append(
			deploymentSignals,
			"Repository documents environment variables with .env.example.",
		)
	}

	if containsFile(languageFilesFound, "go.mod") {
		architectureSignals = append(
			architectureSignals,
			"Go module architecture detected.",
		)
	}

	if containsFile(languageFilesFound, "package.json") {
		architectureSignals = append(
			architectureSignals,
			"Node.js project structure detected.",
		)
	}

	if containsFile(languageFilesFound, "requirements.txt") ||
		containsFile(languageFilesFound, "pyproject.toml") {

		architectureSignals = append(
			architectureSignals,
			"Python project structure detected.",
		)
	}

	if containsFile(languageFilesFound, "tsconfig.json") {
		architectureSignals = append(
			architectureSignals,
			"TypeScript project configuration detected.",
		)
	}

	if containsFile(languageFilesFound, "Package.swift") {
		architectureSignals = append(
			architectureSignals,
			"Swift Package Manager configuration detected.",
		)
	}

	for _, item := range treeItems {

		path := strings.ToLower(item.Path)

		if strings.HasSuffix(path, "_test.go") {

			testingSignals = append(
				testingSignals,
				"Go unit tests detected.",
			)

			break
		}

		if strings.Contains(path, "/test") ||
			strings.Contains(path, "/tests") ||
			strings.Contains(path, "__tests__") ||
			strings.Contains(path, "pytest") ||
			strings.Contains(path, "jest") {

			testingSignals = append(
				testingSignals,
				"Automated testing structure detected.",
			)

			break
		}
	}

	improvementAreas := []string{}

	if !containsFile(importantFilesFound, "README.md") {
		improvementAreas = append(improvementAreas, "Add a README.md file to explain the project purpose and usage")
	}

	if !containsFile(importantFilesFound, "LICENSE") {
		improvementAreas = append(improvementAreas, "Add a LICENSE file")
	}

	if !containsFile(importantFilesFound, "Dockerfile") {
		improvementAreas = append(improvementAreas, "Add a Dockerfile for containerized deployment")
	}

	if !containsFile(importantFilesFound, "docker-compose.yml") {
		improvementAreas = append(improvementAreas, "Add docker-compose.yml for local multi-service setup")
	}

	if !containsFile(importantFilesFound, ".env.example") {
		improvementAreas = append(improvementAreas, "Add .env.example to document required environment variables")
	}

	if len(languageFilesFound) == 0 {
		improvementAreas = append(improvementAreas, "Add dependency/configuration files such as go.mod, package.json, or requirements.txt")
	}

	intelligence := models.RepositoryIntelligence{
		RepositoryName:        repo.Name,
		SelectionReason:       selectionReason,
		PrimaryLanguage:       repo.Language,
		ImportantFilesFound:   importantFilesFound,
		ImportantFilesMissing: importantFilesMissing,
		LanguageFilesFound:    languageFilesFound,
		ArchitectureSignals:   architectureSignals,
		TestingSignals:        testingSignals,
		DeploymentSignals:     deploymentSignals,
		QualitySignals:        qualitySignals,
		ImprovementAreas:      improvementAreas,
		RepositoryScore:       score,
	}

	return intelligence
}

func containsFile(files []string, target string) bool {
	for _, file := range files {
		if file == target {
			return true
		}

		if strings.HasSuffix(target, "/"+file) {
			return true
		}
	}

	return false
}
func AnalyzeGitHubProfile(githubURL string) (models.GitHubAnalysis, error) {
	if githubURL == "" {
		return models.GitHubAnalysis{}, nil
	}

	username := strings.TrimPrefix(githubURL, "https://github.com/")
	username = strings.TrimPrefix(username, "http://github.com/")
	username = strings.Trim(username, "/")
	username = strings.TrimSpace(username)

	apiURL := "https://api.github.com/users/" + username

	resp, err := http.Get(apiURL)
	if err != nil {
		return models.GitHubAnalysis{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.GitHubAnalysis{}, fmt.Errorf("GitHub API Error: %d", resp.StatusCode)
	}

	var result struct {
		Login       string `json:"login"`
		HTMLURL     string `json:"html_url"`
		PublicRepos int    `json:"public_repos"`
		Followers   int    `json:"followers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.GitHubAnalysis{}, err
	}

	return models.GitHubAnalysis{
		Username:    result.Login,
		ProfileURL:  result.HTMLURL,
		PublicRepos: result.PublicRepos,
		Followers:   result.Followers,
	}, nil
}

func AnalyzeGitHubRepositories(username string) ([]models.GitHubRepo, error) {
	apiURL := "https://api.github.com/users/" + username + "/repos"

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github repositories couldn't be fetched")
	}

	var repos []models.GitHubRepo

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func CalculateGitHubAnalysis(githubAnalysis models.GitHubAnalysis, repos []models.GitHubRepo) models.GitHubAnalysis {
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

func calculateUpdateScore(repos []models.GitHubRepo) int {
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
