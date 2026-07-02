package services

import (
	"cv-detector/internal/models"
	"regexp"
	"strings"
)

func ParseJobDescription(jobDescription string) models.ParsedJobDescription {
	text := strings.ToLower(jobDescription)

	requiredSkills := []string{}
	optionalSkills := []string{}

	requiredSection := extractSection(text, []string{"required skills", "requirements"}, []string{"preferred skills", "optional skills", "nice to have", "responsibilities"})
	optionalSection := extractSection(text, []string{"preferred skills", "optional skills", "nice to have"}, []string{"responsibilities", "education", "experience"})

	if requiredSection != "" {
		requiredSkills = ExtractSkills(requiredSection)
	} else {
		requiredSkills = ExtractSkills(jobDescription)
	}

	if optionalSection != "" {
		optionalSkills = ExtractSkills(optionalSection)
	}

	experienceLevel := "Not specified"
	if strings.Contains(text, "junior") {
		experienceLevel = "Junior"
	} else if strings.Contains(text, "mid") || strings.Contains(text, "middle") {
		experienceLevel = "Mid-level"
	} else if strings.Contains(text, "senior") {
		experienceLevel = "Senior"
	} else if strings.Contains(text, "intern") || strings.Contains(text, "internship") {
		experienceLevel = "Internship"
	}

	educationRequirement := "Not specified"
	if strings.Contains(text, "computer science") || strings.Contains(text, "software engineering") || strings.Contains(text, "related field") {
		educationRequirement = "Computer Science, Software Engineering or related field"
	} else if strings.Contains(text, "bachelor") || strings.Contains(text, "bsc") {
		educationRequirement = "Bachelor degree"
	}

	return models.ParsedJobDescription{
		RequiredSkills:       requiredSkills,
		OptionalSkills:       optionalSkills,
		ExperienceLevel:      experienceLevel,
		EducationRequirement: educationRequirement,
	}
}

func extractSection(text string, startKeywords []string, endKeywords []string) string {
	startIndex := -1

	for _, keyword := range startKeywords {
		index := strings.Index(text, keyword)
		if index != -1 {
			startIndex = index + len(keyword)
			break
		}
	}

	if startIndex == -1 {
		return ""
	}

	endIndex := len(text)

	for _, keyword := range endKeywords {
		index := strings.Index(text[startIndex:], keyword)
		if index != -1 && startIndex+index < endIndex {
			endIndex = startIndex + index
		}
	}

	return text[startIndex:endIndex]
}

func CalculateMatch(jobDescription string, cvText string) models.MatchResult {
	parsedJob := ParseJobDescription(jobDescription)
	requiredSkills := parsedJob.RequiredSkills
	cvSkills := ExtractSkills(cvText)

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

	return models.MatchResult{
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
