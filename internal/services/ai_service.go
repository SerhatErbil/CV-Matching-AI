package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func GenerateAIComment(prompt string) (string, error) {
	requestBody := fiber.Map{
		"model":  "llama3.2:3b",
		"prompt": prompt,
		"stream": false,
		"options": fiber.Map{
			"temperature": 0.2,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Response, nil
}
