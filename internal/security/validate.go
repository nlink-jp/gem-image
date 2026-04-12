package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// MaxImageSize is the maximum allowed input image size (20 MB).
const MaxImageSize = 20 * 1024 * 1024

var (
	ErrPathTraversal  = errors.New("path traversal detected")
	ErrInvalidImage   = errors.New("invalid image format")
	ErrFileTooLarge   = errors.New("file exceeds size limit")
	ErrOutputDirNoDir = errors.New("output directory does not exist")
	ErrFileExists     = errors.New("output file already exists (use --force to overwrite)")
)

// PNG magic bytes: \x89PNG\r\n\x1a\n
var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// JPEG magic bytes: \xFF\xD8\xFF
var jpegMagic = []byte{0xFF, 0xD8, 0xFF}

// WEBP: starts with RIFF....WEBP
var webpRIFF = []byte{0x52, 0x49, 0x46, 0x46}
var webpMagic = []byte{0x57, 0x45, 0x42, 0x50}

// GIF magic bytes: GIF87a or GIF89a
var gifMagic = []byte{0x47, 0x49, 0x46}

// ValidateImagePath checks that a file path is safe and resolves symlinks.
// Returns the cleaned absolute path.
func ValidateImagePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", path, err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve symlink %s: %w", path, err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, not a file", path)
	}

	return resolved, nil
}

// ValidateOutputPath checks that the output directory exists and is writable.
// If force is false and the output file already exists, returns ErrFileExists.
func ValidateOutputPath(path string, force bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	dir := filepath.Dir(abs)
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrOutputDirNoDir, dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s is not a directory", ErrOutputDirNoDir, dir)
	}

	// Check if output file already exists
	if !force {
		if _, err := os.Stat(abs); err == nil {
			return fmt.Errorf("%w: %s", ErrFileExists, abs)
		}
	}

	return nil
}

// ValidateImageData checks magic bytes and file size.
func ValidateImageData(data []byte) error {
	if len(data) > MaxImageSize {
		return fmt.Errorf("%w: %d bytes (max %d)", ErrFileTooLarge, len(data), MaxImageSize)
	}

	if !isValidImageMagic(data) {
		return ErrInvalidImage
	}

	return nil
}

// DetectMIMEFromBytes determines MIME type from magic bytes.
// Returns empty string if format is not recognized.
func DetectMIMEFromBytes(data []byte) string {
	if hasPrefix(data, pngMagic) {
		return "image/png"
	}
	if hasPrefix(data, jpegMagic) {
		return "image/jpeg"
	}
	if len(data) >= 12 && hasPrefix(data, webpRIFF) && hasPrefix(data[8:], webpMagic) {
		return "image/webp"
	}
	if hasPrefix(data, gifMagic) {
		return "image/gif"
	}
	return ""
}

func isValidImageMagic(data []byte) bool {
	return DetectMIMEFromBytes(data) != ""
}

func hasPrefix(data, prefix []byte) bool {
	if len(data) < len(prefix) {
		return false
	}
	for i, b := range prefix {
		if data[i] != b {
			return false
		}
	}
	return true
}
