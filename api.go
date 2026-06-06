package word2md

import (
	"context"
	"fmt"
	"io"

	"github.com/saiful-anwar/word2md/pkg/classifier"
	"github.com/saiful-anwar/word2md/pkg/engine"
)

// Source represents the input source for conversion.
type Source struct {
	Path   string
	Reader io.Reader
}

// FromFile creates a Source from a file path.
func FromFile(path string) Source {
	return Source{Path: path}
}

// FromReader creates a Source from an io.Reader.
// Note: file-based extraction requires a path; FromReader is for future streaming support.
func FromReader(r io.Reader) Source {
	return Source{Reader: r}
}

// Convert converts a DOCX source to Markdown.
func Convert(ctx context.Context, source Source, opts ...Option) (*Result, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if source.Path == "" {
		return nil, fmt.Errorf("file path is required for conversion")
	}

	engOpts := []engine.EngineOption{
		engine.WithImageDir(cfg.imageDir),
		engine.WithInlineFormatting(cfg.inlineFormatting),
		engine.WithHyperlinks(cfg.hyperlinks),
	}

	if cfg.headingDetection || cfg.listDetection {
		detector := classifier.NewDetector()
		if cfg.headingDetection {
			detector.UseFontSize = true
			detector.UseStyleNames = true
			detector.UseOutlineLevels = true
		} else {
			detector.UseFontSize = false
			detector.UseStyleNames = false
			detector.UseOutlineLevels = false
		}
		if cfg.headingThresholds.h1 != 0 || cfg.headingThresholds.h2 != 0 || cfg.headingThresholds.h3 != 0 {
			detector.HeadingThresholds = classifier.Thresholds{
				H1: cfg.headingThresholds.h1,
				H2: cfg.headingThresholds.h2,
				H3: cfg.headingThresholds.h3,
			}
		}
		engOpts = append(engOpts, engine.WithHeadingDetector(detector))
	}

	if cfg.listDetection {
		engOpts = append(engOpts, engine.WithListDetection(true))
	}

	eng := engine.NewEngine(engOpts...)

	engResult, err := eng.ConvertFile(ctx, source.Path)
	if err != nil {
		return nil, err
	}

	return &Result{
		Markdown:   engResult.Markdown,
		ImageFiles: engResult.ImageFiles,
		Warnings:   engResult.Warnings,
	}, nil
}

// ConvertFile is a convenience function for converting a file by path.
func ConvertFile(ctx context.Context, path string, opts ...Option) (*Result, error) {
	return Convert(ctx, FromFile(path), opts...)
}
