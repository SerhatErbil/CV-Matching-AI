package models

type GitHubRepo struct {
	Name            string `json:"name"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	UpdatedAt       string `json:"updated_at"`
	DefaultBranch   string `json:"default_branch"`
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
