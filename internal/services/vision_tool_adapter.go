package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// VisionToolAdapter wraps vision capabilities into the tools.VisionClient
// interface for use in the debate ensemble. It delegates to a vision-capable
// LLM provider for actual image analysis.
type VisionToolAdapter struct {
	// provider holds the vision-capable LLM provider. Will be wired with
	// the actual vision handler during router initialization.
	provider VisionProvider
	logger   *logrus.Logger
}

// VisionProvider abstracts vision-capable LLM providers for image analysis.
type VisionProvider interface {
	// AnalyzeImageData sends raw image bytes along with a prompt to a
	// vision-capable model and returns the analysis result.
	AnalyzeImageData(ctx context.Context, imageData []byte, prompt string) (string, error)

	// AnalyzeImageURL sends an image URL along with a prompt to a
	// vision-capable model and returns the analysis result.
	AnalyzeImageURL(ctx context.Context, imageURL string, prompt string) (string, error)
}

// NewVisionToolAdapter creates a new VisionToolAdapter. The provider may be nil;
// in that case all calls will return an error indicating the provider is not
// configured.
func NewVisionToolAdapter(provider VisionProvider, logger *logrus.Logger) *VisionToolAdapter {
	if logger == nil {
		logger = logrus.New()
	}
	return &VisionToolAdapter{
		provider: provider,
		logger:   logger,
	}
}

// AnalyzeImage analyzes raw image data with a text prompt. Implements the
// tools.VisionClient interface.
func (v *VisionToolAdapter) AnalyzeImage(
	ctx context.Context,
	imageData []byte,
	prompt string,
) (interface{}, error) {
	if v.provider == nil {
		v.logger.Warn("vision provider not configured, cannot analyze image data")
		return nil, fmt.Errorf("vision provider not configured")
	}

	if len(imageData) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	if prompt == "" {
		prompt = "Describe this image in detail."
	}

	v.logger.WithFields(logrus.Fields{
		"image_size": len(imageData),
		"prompt":     truncatePrompt(prompt, 100),
	}).Debug("analyzing image data via vision provider")

	result, err := v.provider.AnalyzeImageData(ctx, imageData, prompt)
	if err != nil {
		v.logger.WithError(err).Error("vision provider failed to analyze image data")
		return nil, fmt.Errorf("vision analysis failed: %w", err)
	}

	return result, nil
}

// AnalyzeURL analyzes an image at a URL with a text prompt. Implements the
// tools.VisionClient interface.
func (v *VisionToolAdapter) AnalyzeURL(
	ctx context.Context,
	imageURL string,
	prompt string,
) (interface{}, error) {
	if v.provider == nil {
		v.logger.Warn("vision provider not configured, cannot analyze image URL")
		return nil, fmt.Errorf("vision provider not configured")
	}

	if imageURL == "" {
		return nil, fmt.Errorf("image URL is empty")
	}

	if prompt == "" {
		prompt = "Describe this image in detail."
	}

	v.logger.WithFields(logrus.Fields{
		"image_url": imageURL,
		"prompt":    truncatePrompt(prompt, 100),
	}).Debug("analyzing image URL via vision provider")

	result, err := v.provider.AnalyzeImageURL(ctx, imageURL, prompt)
	if err != nil {
		v.logger.WithError(err).Error("vision provider failed to analyze image URL")
		return nil, fmt.Errorf("vision URL analysis failed: %w", err)
	}

	return result, nil
}

// truncatePrompt truncates a prompt string to maxLen characters for logging.
func truncatePrompt(prompt string, maxLen int) string {
	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen] + "..."
}
