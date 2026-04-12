package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_PNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.png")
	data := []byte("image data")

	if err := WriteFile(path, data, "image/png"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(written) != string(data) {
		t.Errorf("written data mismatch")
	}

	info, _ := os.Stat(path)
	perm := info.Mode().Perm()
	if perm != 0o644 {
		t.Errorf("permission = %o, want 644", perm)
	}
}

func TestResolveFormat(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		formatFlag string
		want       string
	}{
		{"png extension", "out.png", "", "image/png"},
		{"jpeg extension", "out.jpeg", "", "image/jpeg"},
		{"jpg extension", "out.jpg", "", "image/jpeg"},
		{"PNG extension uppercase", "out.PNG", "", "image/png"},
		{"extension overrides flag", "out.jpg", "png", "image/jpeg"},
		{"flag jpeg", "out.bin", "jpeg", "image/jpeg"},
		{"flag jpg", "out.bin", "jpg", "image/jpeg"},
		{"flag png", "out.bin", "png", "image/png"},
		{"default png", "out.bin", "", "image/png"},
		{"no extension no flag", "output", "", "image/png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveFormat(tt.outputPath, tt.formatFlag)
			if got != tt.want {
				t.Errorf("ResolveFormat(%q, %q) = %q, want %q", tt.outputPath, tt.formatFlag, got, tt.want)
			}
		})
	}
}

func TestIsPNG(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	if !isPNG(png) {
		t.Error("expected PNG to be detected")
	}

	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	if isPNG(jpeg) {
		t.Error("JPEG should not be detected as PNG")
	}
}
