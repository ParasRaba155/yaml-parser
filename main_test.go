package main

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

// TestParsing is the sole test to check if parser is validating the yaml files
// correctly or not
func TestParsing(t *testing.T) {
	testDataDir := os.DirFS("testdata")
	err := fs.WalkDir(
		testDataDir,
		".",
		func(path string, d fs.DirEntry, err error) error {
			// skip the directory
			if d.IsDir() {
				return nil
			}
			t.Run("testdata/"+path, func(t *testing.T) {
				fileContent, err := os.ReadFile("testdata/" + path)
				if err != nil {
					t.Errorf(
						"expected no error in reading the file, got: %v",
						err,
					)
				}

				parser := NewParser(fileContent)
				_, err = parser.Parse()

				// invalid yaml should return some error
				if strings.Contains(path, "invalid") && err == nil {
					t.Errorf("expected invalid yaml, got valid")
				}

				// valid yaml should not return any error
				if strings.Contains(path, "valid") &&
					!strings.Contains(path, "invalid") &&
					err != nil {
					t.Errorf("expected valid yaml, got: %v", err)
				}
			})
			return nil
		},
	)
	if err != nil {
		t.Errorf("expected no error in walking dir, got: %v", err)
	}
}
