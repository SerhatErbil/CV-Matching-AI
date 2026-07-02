package services

import (
	"cv-detector/internal/models"
	"regexp"
	"strings"
)

func DetectRedFlags(cvText string, result models.MatchResult) []string {
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

func ExtractGitHubURL(cvText string) string {
	ser := regexp.MustCompile(`(?i)(https?://)?(www\.)?github\.com/[a-zA-Z0-9-]+`)
	match := ser.FindString(cvText)

	if match == "" {
		return ""
	}

	if !strings.HasPrefix(strings.ToLower(match), "http") {
		match = "https://" + match
	}

	return match
}
