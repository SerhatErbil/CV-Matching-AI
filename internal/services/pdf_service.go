package services

import "os/exec"

func ReadPdfText(path string) (string, error) {
	cmd := exec.Command("pdftotext", path, "-")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
