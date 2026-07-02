package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func AnalyzeCV(c *fiber.Ctx) (fiber.Map, error) {
	jobDescription := c.FormValue("job_description")

	if jobDescription == "" {
		return nil, c.Status(400).JSON(fiber.Map{
			"error": "Job description is required",
		})
	}

	file, err := c.FormFile("cv")
	if err != nil {
		return nil, c.Status(400).JSON(fiber.Map{
			"error": "CV file is required",
		})
	}

	os.MkdirAll("uploads", os.ModePerm)

	filePath := filepath.Join("uploads", file.Filename)

	if err := c.SaveFile(file, filePath); err != nil {
		return nil, c.Status(500).JSON(fiber.Map{
			"error": "Failed to save CV file",
		})
	}

	cvText, err := ReadPdfText(filePath)
	if err != nil {
		return nil, c.Status(500).JSON(fiber.Map{
			"error": "Could not read PDF text",
		})
	}

	result := CalculateMatch(jobDescription, cvText)
	cvSkills := ExtractSkills(cvText)

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

	aiComment, err := GenerateAIComment(prompt)
	if err != nil {
		aiComment = "AI evaluation could not be generated."
	}

	return fiber.Map{
		"file_name":       file.Filename,
		"required_skills": result.RequiredSkills,
		"cv_skills":       cvSkills,
		"matched_skills":  result.MatchedSkills,
		"missing_skills":  result.MissingSkills,
		"extra_skills":    result.ExtraSkills,
		"match_score":     result.MatchScore,
		"summary":         result.Summary,
		"ai_comment":      aiComment,
	}, nil
}
