package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadImageFile_ValidPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	// Minimal PNG magic + padding
	data := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 100)...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	img, err := ReadImageFile(path)
	if err != nil {
		t.Fatalf("ReadImageFile: %v", err)
	}
	if img.MIMEType != "image/png" {
		t.Errorf("MIMEType = %q, want image/png", img.MIMEType)
	}
	if len(img.Data) != len(data) {
		t.Errorf("data length = %d, want %d", len(img.Data), len(data))
	}
}

func TestReadImageFile_ValidJPEG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	data := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 100)...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	img, err := ReadImageFile(path)
	if err != nil {
		t.Fatalf("ReadImageFile: %v", err)
	}
	if img.MIMEType != "image/jpeg" {
		t.Errorf("MIMEType = %q, want image/jpeg", img.MIMEType)
	}
}

func TestReadImageFile_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake.png")
	if err := os.WriteFile(path, []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadImageFile(path)
	if err == nil {
		t.Error("expected error for invalid image format")
	}
}

func TestReadImageFile_NonExistent(t *testing.T) {
	_, err := ReadImageFile("/nonexistent/file.png")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
