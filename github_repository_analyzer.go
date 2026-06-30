package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GitHubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type GitHubTreeResponse struct {
	Tree []GitHubTreeItem `json:"tree"`
}

type RepositoryIntelligence struct {
	RepositoryName        string   `json:"repository_name"`
	SelectionReason       string   `json:"selection_reason"`
	PrimaryLanguage       string   `json:"primary_language"`
	ImportantFilesFound   []string `json:"important_files_found"`
	ImportantFilesMissing []string `json:"important_files_missing"`
	LanguageFilesFound    []string `json:"language_files_found"`
	ArchitectureSignals   []string `json:"architecture_signals"`
	TestingSignals        []string `json:"testing_signals"`
	DeploymentSignals     []string `json:"deployment_signals"`
	QualitySignals        []string `json:"quality_signals"`
	ImprovementAreas      []string `json:"improvement_areas"`
	RepositoryScore       int      `json:"repository_score"`
}

func selectRepositoriesForAnalysis(repos []GitHubRepo) []GitHubRepo {

	// 1. Repo yoksa boş slice dön
	if len(repos) == 0 {
		return []GitHubRepo{}
	}

	selectedRepos := []GitHubRepo{}

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

func fetchRepositoryTree(username string, repo GitHubRepo) ([]GitHubTreeItem, error) {
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
		return []GitHubTreeItem{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []GitHubTreeItem{}, fmt.Errorf("failed to fetch repository tree")
	}

	var treeResponse GitHubTreeResponse

	if err := json.NewDecoder(resp.Body).Decode(&treeResponse); err != nil {
		return []GitHubTreeItem{}, err
	}

	return treeResponse.Tree, nil
}

func analyzeRepositoryTree(repo GitHubRepo, treeItems []GitHubTreeItem, selectionReason string) RepositoryIntelligence {
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

	intelligence := RepositoryIntelligence{
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
