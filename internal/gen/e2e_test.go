package gen

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func generateToString(t *testing.T, schemaPath, pkgName, structName string) string {
	generator, err := NewGenerator(schemaPath, pkgName, structName)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	var buf bytes.Buffer
	_, err = generator.Generate(&buf)
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	return buf.String()
}

func TestE2E_MovieSchema(t *testing.T) {
	code := generateToString(t, "../../testdata/schemas/movie.schema.json", "movie", "MovieLens")

	expectedMethods := []string{
		"func (l *MovieLens) GetCastAt(index0 int) string",
		"func (l *MovieLens) SetCastAt(index0 int, value string) error",
		"func (l *MovieLens) LenCast() int64",
		"func (l *MovieLens) ForEachCast(callback func(index int, value string) bool)",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(code, method) {
			t.Errorf("expected generated code to contain: %s", method)
		}
	}
}

func TestE2E_BlogPostSchema(t *testing.T) {
	code := generateToString(t, "../../testdata/schemas/blog-post.schema.json", "blogpost", "BlogPostLens")

	expectedMethods := []string{
		"func (l *BlogPostLens) GetTagsAt(index0 int) string",
		"func (l *BlogPostLens) SetTagsAt(index0 int, value string) error",
		"func (l *BlogPostLens) LenTags() int64",
		"func (l *BlogPostLens) ForEachTags(callback func(index int, value string) bool)",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(code, method) {
			t.Errorf("expected generated code to contain: %s", method)
		}
	}
}

func TestIntegration_RealPayload(t *testing.T) {
	code := generateToString(t, "../../testdata/schemas/movie.schema.json", "main", "MovieLens")

	tmpDir, err := os.MkdirTemp("", "lens-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "movie.go"), []byte(code), 0644); err != nil {
		t.Fatalf("failed to write movie.go: %v", err)
	}

	testCode := `package main

import (
	"testing"
)

func TestMovieLensIntegration(t *testing.T) {
	jsonPayload := []byte(` + "`" + `{
		"title": "Inception",
		"director": "Christopher Nolan",
		"releaseDate": "2010-07-16",
		"cast": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"]
	}` + "`" + `)

	lens := NewMovieLens(jsonPayload)

	if lens.LenCast() != 2 {
		t.Errorf("expected LenCast() == 2, got %d", lens.LenCast())
	}

	if val := lens.GetCastAt(0); val != "Leonardo DiCaprio" {
		t.Errorf("expected 'Leonardo DiCaprio', got '%s'", val)
	}

	if err := lens.SetCastAt(1, "Tom Hardy"); err != nil {
		t.Errorf("SetCastAt failed: %v", err)
	}
	if val := lens.GetCastAt(1); val != "Tom Hardy" {
		t.Errorf("expected 'Tom Hardy', got '%s'", val)
	}

	count := 0
	lens.ForEachCast(func(index int, value string) bool {
		if index == 0 && value != "Leonardo DiCaprio" {
			t.Errorf("index 0 wrong: %s", value)
		}
		if index == 1 && value != "Tom Hardy" {
			t.Errorf("index 1 wrong: %s", value)
		}
		count++
		return true
	})
	if count != 2 {
		t.Errorf("expected 2 iterations, got %d", count)
	}
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "movie_test.go"), []byte(testCode), 0644); err != nil {
		t.Fatalf("failed to write movie_test.go: %v", err)
	}

	cmd := exec.Command("go", "mod", "init", "testint")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("go mod init failed: %v", err)
	}

	cmd = exec.Command("go", "get", "github.com/tidwall/gjson")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("go get gjson failed: %v", err)
	}
	cmd = exec.Command("go", "get", "github.com/tidwall/sjson")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("go get sjson failed: %v", err)
	}

	cmd = exec.Command("go", "test", "-v")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("integration test failed: %v\nOutput: %s", err, string(output))
	}
}
