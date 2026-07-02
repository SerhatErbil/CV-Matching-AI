package services

import (
	"cv-detector/internal/models"
	"fmt"
)

func CalculateFinalCandidateScore(
	matchScore int,
	githubScore int,
	redFlags []string,
	repositoryIntelligence []models.RepositoryIntelligence,
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

func GenerateFinalRecommendation(finalScore int) string {
	if finalScore >= 70 {
		return "Strong candidate. Recommended for interview."
	}

	if finalScore >= 50 {
		return "Potential candidate. Consider for interview after manual review."
	}

	return "Weak match. Not recommended for interview at this stage."
}

func GenerateCandidateAISummary(
	jobDescription string,
	parsedJobDescription models.ParsedJobDescription,
	result models.MatchResult,
	githubAnalysis models.GitHubAnalysis,
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

Do not mention optional skills as weaknesses unless they appear in Missing Skills or Red Flags.
Do not mention Agile, teamwork, experience level, soft skills, or education unless they explicitly appear in Missing Skills, Red Flags, or Parsed fields.
Do not say a candidate lacks something unless it is explicitly listed in Missing Skills or Red Flags.
Preferred Skills are not weaknesses.
Optional Skills are not weaknesses.

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

Parsed Required Skills: %v
Parsed Optional Skills: %v
Parsed Experience Level: %s
Parsed Education Requirement: %s

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
		parsedJobDescription.RequiredSkills,
		parsedJobDescription.OptionalSkills,
		parsedJobDescription.ExperienceLevel,
		parsedJobDescription.EducationRequirement,
	)

	comment, err := GenerateAIComment(prompt)
	if err != nil {
		return "Candidate AI summary could not be generated."
	}

	return comment
}
