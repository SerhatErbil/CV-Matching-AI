package services

import (
	"regexp"
	"strings"
)

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
	"postgresql",
	"rest api",
	"git",
	"ci/cd",
	"microservices",
	"linux",
}

func ExtractSkills(text string) []string {
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
