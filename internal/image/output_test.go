package image

import (
	"bytes"
	stdimage "image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// makeImage builds a small solid-colour image encoded as the given MIME type.
func makeImage(t *testing.T, mimeType string) []byte {
	t.Helper()
	img := stdimage.NewRGBA(stdimage.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: 200, G: 40, B: 40, A: 255})
		}
	}
	var buf bytes.Buffer
	var err error
	switch mimeType {
	case "image/png":
		err = png.Encode(&buf, img)
	case "image/jpeg":
		err = jpeg.Encode(&buf, img, nil)
	default:
		t.Fatalf("unsupported test format %q", mimeType)
	}
	if err != nil {
		t.Fatalf("encode %s: %v", mimeType, err)
	}
	return buf.Bytes()
}

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

// The flash-lite image models return JPEG, so a PNG output path has to be
// transcoded rather than handed the model bytes verbatim.
func TestWriteFile_JPEGModelDataToPNGPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.png")

	if err := WriteFile(path, makeImage(t, "image/jpeg"), "image/png"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !isPNG(written) {
		t.Errorf("file at %s is not PNG", path)
	}
}

func TestWriteFile_PNGModelDataToJPEGPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.jpg")

	if err := WriteFile(path, makeImage(t, "image/png"), "image/jpeg"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !isJPEG(written) {
		t.Errorf("file at %s is not JPEG", path)
	}
}

func TestTranscode(t *testing.T) {
	pngData := makeImage(t, "image/png")
	jpegData := makeImage(t, "image/jpeg")

	tests := []struct {
		name    string
		data    []byte
		want    string
		checkFn func([]byte) bool
	}{
		{"png to jpeg", pngData, "image/jpeg", isJPEG},
		{"jpeg to png", jpegData, "image/png", isPNG},
		{"png stays png", pngData, "image/png", isPNG},
		{"jpeg stays jpeg", jpegData, "image/jpeg", isJPEG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Transcode(tt.data, tt.want)
			if err != nil {
				t.Fatalf("Transcode: %v", err)
			}
			if !tt.checkFn(got) {
				t.Errorf("result is not %s", tt.want)
			}
		})
	}
}

func TestTranscode_UnknownFormatPassesThrough(t *testing.T) {
	data := []byte("not an image")
	got, err := Transcode(data, "image/png")
	if err != nil {
		t.Fatalf("Transcode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("data was modified: %q", got)
	}
}

func TestTranscode_UnsupportedTarget(t *testing.T) {
	if _, err := Transcode(makeImage(t, "image/png"), "image/webp"); err == nil {
		t.Error("expected an error for an unsupported target format")
	}
}

func TestIsJPEG(t *testing.T) {
	if !isJPEG(makeImage(t, "image/jpeg")) {
		t.Error("expected JPEG to be detected")
	}
	if isJPEG(makeImage(t, "image/png")) {
		t.Error("PNG should not be detected as JPEG")
	}
}

// Go's PNG encoder widens YCbCr (what a JPEG decodes to) to 16 bits per
// channel unless it is converted first, doubling the file size for no gain.
func TestTranscode_JPEGToPNGStays8Bit(t *testing.T) {
	got, err := Transcode(makeImage(t, "image/jpeg"), "image/png")
	if err != nil {
		t.Fatalf("Transcode: %v", err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("DecodeConfig: %v", err)
	}
	switch cfg.ColorModel {
	case color.NRGBA64Model, color.RGBA64Model, color.Gray16Model:
		t.Errorf("PNG was encoded at 16 bits per channel (%T)", cfg.ColorModel)
	}
}
