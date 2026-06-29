package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	return RepositoryIntelligence{}
}
