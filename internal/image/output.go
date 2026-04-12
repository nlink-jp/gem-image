package image

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// WriteFile writes image data to a file with 0644 permissions.
// If the desired format is JPEG and the data is PNG, it converts automatically.
func WriteFile(path string, data []byte, desiredFormat string) error {
	out := data
	if desiredFormat == "image/jpeg" && isPNG(data) {
		converted, err := pngToJPEG(data)
		if err != nil {
			return fmt.Errorf("convert PNG to JPEG: %w", err)
		}
		out = converted
	}

	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write image %s: %w", path, err)
	}
	return nil
}

// ResolveFormat determines the output MIME type from the output path extension
// and the --format flag. Extension takes precedence over the flag.
func ResolveFormat(outputPath string, formatFlag string) string {
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	}

	switch strings.ToLower(formatFlag) {
	case "jpeg", "jpg":
		return "image/jpeg"
	default:
		return "image/png"
	}
}

func isPNG(data []byte) bool {
	return len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 &&
		data[2] == 0x4E && data[3] == 0x47
}

func pngToJPEG(data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
