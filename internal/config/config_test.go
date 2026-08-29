package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")

	cfg, err := Load("/nonexistent/path.toml")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GCP.Project != "test-project" {
		t.Errorf("project = %q, want test-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "global" {
		t.Errorf("location = %q, want global (Gemini 3 image models are global-only)", cfg.GCP.Location)
	}
	if cfg.Model.Name != DefaultModel {
		t.Errorf("model = %q, want %s", cfg.Model.Name, DefaultModel)
	}
	if !strings.HasPrefix(DefaultModel, "gemini-3") {
		t.Errorf("DefaultModel = %q, want a Gemini 3 model", DefaultModel)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("GEMIMAGE_PROJECT", "env-project")
	t.Setenv("GEMIMAGE_LOCATION", "asia-northeast1")
	t.Setenv("GEMIMAGE_MODEL", "gemini-2.5-pro")

	cfg, err := Load("/nonexistent/path.toml")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GCP.Project != "env-project" {
		t.Errorf("project = %q, want env-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "asia-northeast1" {
		t.Errorf("location = %q, want asia-northeast1", cfg.GCP.Location)
	}
	if cfg.Model.Name != "gemini-2.5-pro" {
		t.Errorf("model = %q, want gemini-2.5-pro", cfg.Model.Name)
	}
}

func TestLoad_TOMLFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `[gcp]
project  = "toml-project"
location = "europe-west1"

[model]
name = "custom-model"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GEMIMAGE_PROJECT", "")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GCP.Project != "toml-project" {
		t.Errorf("project = %q, want toml-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "europe-west1" {
		t.Errorf("location = %q, want europe-west1", cfg.GCP.Location)
	}
	if cfg.Model.Name != "custom-model" {
		t.Errorf("model = %q, want custom-model", cfg.Model.Name)
	}
}

func TestLoad_EnvOverridesToml(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `[gcp]
project = "toml-project"
[model]
name = "toml-model"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GEMIMAGE_PROJECT", "")
	t.Setenv("GEMIMAGE_MODEL", "env-model")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Model.Name != "env-model" {
		t.Errorf("model = %q, want env-model (env override)", cfg.Model.Name)
	}
	if cfg.GCP.Project != "toml-project" {
		t.Errorf("project = %q, want toml-project", cfg.GCP.Project)
	}
}

func TestLoad_MissingProject(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GEMIMAGE_PROJECT", "")

	_, err := Load("/nonexistent/path.toml")
	if err == nil {
		t.Error("expected error when project is not set")
	}
}

func TestApplyFlags(t *testing.T) {
	cfg := &Config{Model: ModelConfig{Name: "original"}}
	cfg.ApplyFlags("override-model")
	if cfg.Model.Name != "override-model" {
		t.Errorf("model = %q, want override-model", cfg.Model.Name)
	}
}

func TestApplyFlags_Empty(t *testing.T) {
	cfg := &Config{Model: ModelConfig{Name: "original"}}
	cfg.ApplyFlags("")
	if cfg.Model.Name != "original" {
		t.Errorf("model = %q, want original (empty flag should not override)", cfg.Model.Name)
	}
}
