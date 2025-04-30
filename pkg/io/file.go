package io

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SecretData represents a secret or variable entry from a file
type SecretData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ReadJSONSecrets reads secrets from a JSON file
func ReadJSONSecrets(filePath string) ([]SecretData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try to parse as array of secrets first
	var secretsArray []SecretData
	if err := json.Unmarshal(data, &secretsArray); err == nil {
		return secretsArray, nil
	}

	// Try to parse as map of name:value pairs
	var secretsMap map[string]string
	if err := json.Unmarshal(data, &secretsMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert map to array
	secrets := make([]SecretData, 0, len(secretsMap))
	for name, value := range secretsMap {
		secrets = append(secrets, SecretData{
			Name:  name,
			Value: value,
		})
	}

	return secrets, nil
}

// ReadCSVSecrets reads secrets from a CSV file
func ReadCSVSecrets(filePath string) ([]SecretData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Find name and value column indices
	var nameIdx, valueIdx int = -1, -1
	for i, col := range header {
		col = strings.ToLower(strings.TrimSpace(col))
		switch col {
		case "name", "secret_name", "variable_name":
			nameIdx = i
		case "value", "secret_value", "variable_value":
			valueIdx = i
		}
	}

	if nameIdx == -1 || valueIdx == -1 {
		return nil, fmt.Errorf("CSV must have 'name' and 'value' columns (or variations)")
	}

	var secrets []SecretData
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		if len(record) <= nameIdx || len(record) <= valueIdx {
			continue // Skip malformed rows
		}

		name := strings.TrimSpace(record[nameIdx])
		value := strings.TrimSpace(record[valueIdx])
		if name == "" || value == "" {
			continue // Skip empty rows
		}

		secrets = append(secrets, SecretData{
			Name:  name,
			Value: value,
		})
	}

	return secrets, nil
}

// WriteJSONSecrets writes secrets to a JSON file
func WriteJSONSecrets(filePath string, secrets []SecretData) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(secrets); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

// WriteCSVSecrets writes secrets to a CSV file
func WriteCSVSecrets(filePath string, secrets []SecretData) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"name", "value"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records
	for _, secret := range secrets {
		if err := writer.Write([]string{secret.Name, secret.Value}); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}