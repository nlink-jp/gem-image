// Package image handles image file I/O for gem-image.
package image

import (
	"fmt"
	"os"

	"github.com/nlink-jp/gem-image/internal/security"
)

// ImageInput holds validated image data ready for API submission.
type ImageInput struct {
	Data     []byte
	MIMEType string
}

// ReadImageFile reads and validates an image file, returning the image data
// and its detected MIME type.
func ReadImageFile(path string) (*ImageInput, error) {
	resolved, err := security.ValidateImagePath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("read image %s: %w", path, err)
	}

	if err := security.ValidateImageData(data); err != nil {
		return nil, fmt.Errorf("validate image %s: %w", path, err)
	}

	mime := security.DetectMIMEFromBytes(data)
	return &ImageInput{Data: data, MIMEType: mime}, nil
}
