package security

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateImagePath_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	resolved, err := ValidateImagePath(path)
	if err != nil {
		t.Fatalf("ValidateImagePath: %v", err)
	}
	if !filepath.IsAbs(resolved) {
		t.Errorf("expected absolute path, got %q", resolved)
	}
}

func TestValidateImagePath_NonExistent(t *testing.T) {
	_, err := ValidateImagePath("/nonexistent/file.png")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidateImagePath_Directory(t *testing.T) {
	dir := t.TempDir()
	_, err := ValidateImagePath(dir)
	if err == nil {
		t.Error("expected error for directory")
	}
}

func TestValidateOutputPath_ValidDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.png")
	err := ValidateOutputPath(path, false)
	if err != nil {
		t.Fatalf("ValidateOutputPath: %v", err)
	}
}

func TestValidateOutputPath_NonExistentDir(t *testing.T) {
	err := ValidateOutputPath("/nonexistent/dir/output.png", false)
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
	if !errors.Is(err, ErrOutputDirNoDir) {
		t.Errorf("expected ErrOutputDirNoDir, got %v", err)
	}
}

func TestValidateOutputPath_FileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.png")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := ValidateOutputPath(path, false)
	if !errors.Is(err, ErrFileExists) {
		t.Errorf("expected ErrFileExists, got %v", err)
	}
}

func TestValidateOutputPath_FileExistsWithForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.png")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := ValidateOutputPath(path, true)
	if err != nil {
		t.Fatalf("expected no error with --force, got %v", err)
	}
}

func TestValidateImageData_ValidPNG(t *testing.T) {
	// Minimal PNG magic bytes + some data
	data := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 100)...)
	if err := ValidateImageData(data); err != nil {
		t.Fatalf("ValidateImageData (PNG): %v", err)
	}
}

func TestValidateImageData_ValidJPEG(t *testing.T) {
	data := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 100)...)
	if err := ValidateImageData(data); err != nil {
		t.Fatalf("ValidateImageData (JPEG): %v", err)
	}
}

func TestValidateImageData_InvalidFormat(t *testing.T) {
	data := []byte("not an image file")
	err := ValidateImageData(data)
	if !errors.Is(err, ErrInvalidImage) {
		t.Errorf("expected ErrInvalidImage, got %v", err)
	}
}

func TestValidateImageData_TooLarge(t *testing.T) {
	data := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, MaxImageSize+1)...)
	err := ValidateImageData(data)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestDetectMIMEFromBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"PNG", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "image/png"},
		{"JPEG", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"GIF", []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, "image/gif"},
		{"WEBP", []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}, "image/webp"},
		{"unknown", []byte{0x00, 0x01, 0x02}, ""},
		{"empty", []byte{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectMIMEFromBytes(tt.data)
			if got != tt.want {
				t.Errorf("DetectMIMEFromBytes = %q, want %q", got, tt.want)
			}
		})
	}
}
