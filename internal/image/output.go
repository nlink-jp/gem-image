package image

import (
	"bytes"
	"errors"
	"fmt"
	stdimage "image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// WriteFile writes image data to a file with 0644 permissions.
// The model picks the encoding of what it returns (PNG for most models, JPEG
// for the flash-lite image models), so the data is transcoded whenever it does
// not already match desiredFormat. Data in neither PNG nor JPEG is written
// through unchanged.
func WriteFile(path string, data []byte, desiredFormat string) error {
	out, err := Transcode(data, desiredFormat)
	if err != nil {
		return err
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

func isJPEG(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
}

// sourceFormat reports the MIME type of data, or "" when it is neither PNG
// nor JPEG.
func sourceFormat(data []byte) string {
	switch {
	case isPNG(data):
		return "image/png"
	case isJPEG(data):
		return "image/jpeg"
	}
	return ""
}

// MaxImageDimension is the maximum allowed width or height for image decoding.
// Prevents Image Bomb attacks that cause OOM via extremely large pixel dimensions.
const MaxImageDimension = 10000

// ErrImageTooLarge is returned when image dimensions exceed MaxImageDimension.
var ErrImageTooLarge = errors.New("image dimensions too large for decoding")

// toNRGBA converts img to 8-bit NRGBA unless the PNG encoder already writes it
// at 8 bits per channel. Decoded JPEG data is YCbCr, which the encoder would
// otherwise widen to a 16-bit PNG — twice the bytes for no extra detail.
func toNRGBA(img stdimage.Image) stdimage.Image {
	switch img.(type) {
	case *stdimage.NRGBA, *stdimage.RGBA, *stdimage.Gray, *stdimage.Paletted:
		return img
	}
	bounds := img.Bounds()
	dst := stdimage.NewNRGBA(bounds)
	draw.Draw(dst, bounds, img, bounds.Min, draw.Src)
	return dst
}

// Transcode re-encodes PNG/JPEG data into desiredFormat. Data that already
// matches, or that is in neither format, is returned unchanged.
func Transcode(data []byte, desiredFormat string) ([]byte, error) {
	src := sourceFormat(data)
	if src == "" || src == desiredFormat {
		return data, nil
	}

	// Check dimensions before full decode to prevent Image Bomb attacks
	cfg, _, err := stdimage.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("read %s config: %w", src, err)
	}
	if cfg.Width > MaxImageDimension || cfg.Height > MaxImageDimension {
		return nil, fmt.Errorf("%w: %dx%d (max %d)", ErrImageTooLarge, cfg.Width, cfg.Height, MaxImageDimension)
	}

	img, _, err := stdimage.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", src, err)
	}

	var buf bytes.Buffer
	switch desiredFormat {
	case "image/jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	case "image/png":
		err = png.Encode(&buf, toNRGBA(img))
	default:
		return nil, fmt.Errorf("unsupported output format %q", desiredFormat)
	}
	if err != nil {
		return nil, fmt.Errorf("convert %s to %s: %w", src, desiredFormat, err)
	}
	return buf.Bytes(), nil
}
