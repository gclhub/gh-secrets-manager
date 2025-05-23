package version

import (
	"testing"
)

func TestVersionVariables(t *testing.T) {
	if Version == "" {
		t.Error("Expected non-empty version string")
	}

	// CommitHash might be "unknown" in test environment, but should not be empty
	if CommitHash == "" {
		t.Error("Expected non-empty commit hash")
	}
}