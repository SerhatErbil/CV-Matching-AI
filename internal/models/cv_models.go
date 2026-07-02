package models

type MatchResult struct {
	RequiredSkills []string `json:"required_skills"`
	MatchedSkills  []string `json:"matched_skills"`
	MissingSkills  []string `json:"missing_skills"`
	ExtraSkills    []string `json:"extra_skills"`
	MatchScore     int      `json:"match_score"`
	Summary        string   `json:"summary"`
}

type ParsedJobDescription struct {
	RequiredSkills       []string `json:"required_skills"`
	OptionalSkills       []string `json:"optional_skills"`
	ExperienceLevel      string   `json:"experience_level"`
	EducationRequirement string   `json:"education_requirement"`
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
