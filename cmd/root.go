// Package cmd implements the gem-image CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nlink-jp/gem-image/internal/client"
	"github.com/nlink-jp/gem-image/internal/config"
	"github.com/nlink-jp/gem-image/internal/image"
	"github.com/nlink-jp/gem-image/internal/security"
	"github.com/spf13/cobra"
)

var version string

// CLI flags
var (
	flagPrompt     string
	flagInputs     []string
	flagOutput     string
	flagFormat     string
	flagConfigPath string
	flagModel      string
	flagDebug      bool
)

// Exit codes
const (
	exitOK             = 0
	exitGeneralError   = 1
	exitInputError     = 2
	exitAPIError       = 3
	exitSafetyBlocked  = 4
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gem-image",
		Short:        "Generate and edit images using Gemini 2.5 Flash",
		SilenceUsage: true,
		RunE:         runGenerate,
	}

	cmd.Flags().StringVarP(&flagPrompt, "prompt", "p", "", "image generation prompt (reads stdin if omitted)")
	cmd.Flags().StringSliceVarP(&flagInputs, "input", "i", nil, "input image path (repeatable)")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "output file path (required)")
	cmd.Flags().StringVar(&flagFormat, "format", "png", "output format: png or jpeg")
	cmd.Flags().StringVarP(&flagConfigPath, "config", "c", "", "config file path")
	cmd.Flags().StringVarP(&flagModel, "model", "m", "", "model name override")
	cmd.Flags().BoolVar(&flagDebug, "debug", false, "enable debug output")

	_ = cmd.MarkFlagRequired("output")

	return cmd
}

// Execute runs the root command.
func Execute(v string) {
	version = v
	cmd := newRootCmd()
	cmd.Version = version

	if err := cmd.Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		os.Exit(exitGeneralError)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// 1. Load configuration
	cfg, err := config.Load(flagConfigPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	cfg.ApplyFlags(flagModel)

	// 2. Obtain prompt
	prompt, err := resolvePrompt(flagPrompt, os.Stdin)
	if err != nil {
		cmd.SilenceUsage = false
		return err
	}

	// 3. Validate output path
	if err := security.ValidateOutputPath(flagOutput); err != nil {
		return exitWithCode(fmt.Errorf("output: %w", err), exitInputError)
	}

	// 4. Read and validate input images
	var images []*client.ImageInput
	for _, path := range flagInputs {
		img, err := image.ReadImageFile(path)
		if err != nil {
			return exitWithCode(fmt.Errorf("input image: %w", err), exitInputError)
		}
		images = append(images, &client.ImageInput{
			Data:     img.Data,
			MIMEType: img.MIMEType,
		})
	}

	// 5. Wrap prompt for injection protection
	sysPrompt, wrappedPrompt, err := security.WrapPrompt(prompt)
	if err != nil {
		return fmt.Errorf("prompt protection: %w", err)
	}

	// 6. Resolve output format
	outputFormat := image.ResolveFormat(flagOutput, flagFormat)

	if flagDebug {
		fmt.Fprintf(os.Stderr, "[debug] model=%s images=%d format=%s\n",
			cfg.Model.Name, len(images), outputFormat)
	}

	// 7. Create client and generate
	ctx := context.Background()
	c, err := client.New(ctx, cfg)
	if err != nil {
		return exitWithCode(fmt.Errorf("client: %w", err), exitAPIError)
	}
	defer c.Close()

	result, err := c.Generate(ctx, &client.GenerateOpts{
		SystemPrompt: sysPrompt,
		UserPrompt:   wrappedPrompt,
		Images:       images,
		OutputFormat: outputFormat,
	})
	if err != nil {
		if errors.Is(err, client.ErrSafetyBlock) {
			return exitWithCode(fmt.Errorf("blocked by safety filter"), exitSafetyBlocked)
		}
		return exitWithCode(fmt.Errorf("generate: %w", err), exitAPIError)
	}

	// 8. Write output image
	if err := image.WriteFile(flagOutput, result.ImageData, outputFormat); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	// 9. Display token usage on stderr
	if result.Usage != nil {
		fmt.Fprintf(os.Stderr, "tokens: input=%d output=%d total=%d\n",
			result.Usage.InputTokens, result.Usage.OutputTokens, result.Usage.TotalTokens)
	}

	if flagDebug && result.Text != "" {
		fmt.Fprintf(os.Stderr, "[debug] model text: %s\n", result.Text)
	}

	return nil
}

// resolvePrompt obtains the prompt from -p flag or stdin.
func resolvePrompt(flagValue string, stdin *os.File) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	// Check if stdin has data
	stat, _ := stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("prompt is required: use -p flag or pipe input via stdin")
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	s := strings.TrimSpace(string(data))
	if s == "" {
		return "", fmt.Errorf("empty prompt from stdin")
	}

	return s, nil
}

// exitError wraps an error with an exit code.
type exitError struct {
	err  error
	code int
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

func exitWithCode(err error, code int) error {
	return &exitError{err: err, code: code}
}
