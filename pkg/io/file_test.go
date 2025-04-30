package io

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadJSONSecrets_Array(t *testing.T) {
	// Create a temporary JSON file with test data
	content := `[
		{"name": "SECRET1", "value": "value1"},
		{"name": "SECRET2", "value": "value2"}
	]`
	tmpfile := createTempFile(t, "test_secrets_*.json", content)
	defer os.Remove(tmpfile)

	secrets, err := ReadJSONSecrets(tmpfile)
	if err != nil {
		t.Fatalf("ReadJSONSecrets failed: %v", err)
	}

	expected := []SecretData{
		{Name: "SECRET1", Value: "value1"},
		{Name: "SECRET2", Value: "value2"},
	}

	if !reflect.DeepEqual(secrets, expected) {
		t.Errorf("ReadJSONSecrets = %v, want %v", secrets, expected)
	}
}

func TestReadJSONSecrets_Map(t *testing.T) {
	content := `{
		"SECRET1": "value1",
		"SECRET2": "value2"
	}`
	tmpfile := createTempFile(t, "test_secrets_*.json", content)
	defer os.Remove(tmpfile)

	secrets, err := ReadJSONSecrets(tmpfile)
	if err != nil {
		t.Fatalf("ReadJSONSecrets failed: %v", err)
	}

	// Since map iteration order is non-deterministic, we need to check contents without order
	if len(secrets) != 2 {
		t.Errorf("ReadJSONSecrets returned %d secrets, want 2", len(secrets))
	}

	secretMap := make(map[string]string)
	for _, s := range secrets {
		secretMap[s.Name] = s.Value
	}

	expectedValues := map[string]string{
		"SECRET1": "value1",
		"SECRET2": "value2",
	}

	if !reflect.DeepEqual(secretMap, expectedValues) {
		t.Errorf("ReadJSONSecrets contents = %v, want %v", secretMap, expectedValues)
	}
}

func TestReadCSVSecrets(t *testing.T) {
	content := `name,value
SECRET1,value1
SECRET2,value2`
	tmpfile := createTempFile(t, "test_secrets_*.csv", content)
	defer os.Remove(tmpfile)

	secrets, err := ReadCSVSecrets(tmpfile)
	if err != nil {
		t.Fatalf("ReadCSVSecrets failed: %v", err)
	}

	expected := []SecretData{
		{Name: "SECRET1", Value: "value1"},
		{Name: "SECRET2", Value: "value2"},
	}

	if !reflect.DeepEqual(secrets, expected) {
		t.Errorf("ReadCSVSecrets = %v, want %v", secrets, expected)
	}
}

func TestReadCSVSecrets_AlternativeHeaders(t *testing.T) {
	content := `secret_name,secret_value
SECRET1,value1
SECRET2,value2`
	tmpfile := createTempFile(t, "test_secrets_*.csv", content)
	defer os.Remove(tmpfile)

	secrets, err := ReadCSVSecrets(tmpfile)
	if err != nil {
		t.Fatalf("ReadCSVSecrets failed: %v", err)
	}

	expected := []SecretData{
		{Name: "SECRET1", Value: "value1"},
		{Name: "SECRET2", Value: "value2"},
	}

	if !reflect.DeepEqual(secrets, expected) {
		t.Errorf("ReadCSVSecrets = %v, want %v", secrets, expected)
	}
}

func TestWriteJSONSecrets(t *testing.T) {
	secrets := []SecretData{
		{Name: "SECRET1", Value: "value1"},
		{Name: "SECRET2", Value: "value2"},
	}

	tmpfile := filepath.Join(t.TempDir(), "test_secrets.json")
	if err := WriteJSONSecrets(tmpfile, secrets); err != nil {
		t.Fatalf("WriteJSONSecrets failed: %v", err)
	}

	// Read back and verify
	readSecrets, err := ReadJSONSecrets(tmpfile)
	if err != nil {
		t.Fatalf("Failed to read written secrets: %v", err)
	}

	if !reflect.DeepEqual(readSecrets, secrets) {
		t.Errorf("WriteJSONSecrets wrote %v, read back %v", secrets, readSecrets)
	}
}

func TestWriteCSVSecrets(t *testing.T) {
	secrets := []SecretData{
		{Name: "SECRET1", Value: "value1"},
		{Name: "SECRET2", Value: "value2"},
	}

	tmpfile := filepath.Join(t.TempDir(), "test_secrets.csv")
	if err := WriteCSVSecrets(tmpfile, secrets); err != nil {
		t.Fatalf("WriteCSVSecrets failed: %v", err)
	}

	// Read back and verify
	readSecrets, err := ReadCSVSecrets(tmpfile)
	if err != nil {
		t.Fatalf("Failed to read written secrets: %v", err)
	}

	if !reflect.DeepEqual(readSecrets, secrets) {
		t.Errorf("WriteCSVSecrets wrote %v, read back %v", secrets, readSecrets)
	}
}

// Helper function to create temporary test files
func createTempFile(t *testing.T, pattern, content string) string {
	t.Helper()
	tmpfile := filepath.Join(t.TempDir(), pattern)
	if err := os.WriteFile(tmpfile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tmpfile
}