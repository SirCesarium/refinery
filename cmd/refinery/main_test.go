package main

import (
	"os"
	"testing"
)

func TestMainExists(t *testing.T) {
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		t.Error("main.go should exist in cmd/refinery")
	}
}

func TestMainFunction(t *testing.T) {
	_, err := os.Stat("main.go")
	if err != nil {
		t.Fatalf("main.go should be readable: %v", err)
	}

	content, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "func main()") {
		t.Error("main.go should contain func main()")
	}
	if !contains(contentStr, "cli.Execute()") {
		t.Error("main.go should call cli.Execute()")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && search(s, substr)
}

func search(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
