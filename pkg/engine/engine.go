package engine

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/saiful-anwar/word2md/pkg/classifier"
	"github.com/saiful-anwar/word2md/pkg/extractor"
	"github.com/saiful-anwar/word2md/pkg/models"
	"github.com/saiful-anwar/word2md/pkg/parser"
	"github.com/saiful-anwar/word2md/pkg/renderer"
)

// Engine orchestrates the conversion pipeline.
type Engine struct {
	reader          *parser.DefaultReader
	extractor       *extractor.ImageExtractor
	headingDetector *classifier.Detector
	listDetector    *classifier.ListDetector
	renderer        renderer.Renderer
}

// NewEngine creates a new conversion engine with default components.
func NewEngine(opts ...EngineOption) *Engine {
	eng := &Engine{
		reader:          parser.NewReader(),
		extractor:       extractor.NewImageExtractor(),
		headingDetector: classifier.NewDetector(),
		listDetector:    classifier.NewListDetector(),
		renderer:        renderer.NewMarkdownRenderer(),
	}
	for _, opt := range opts {
		opt(eng)
	}
	return eng
}

// EngineOption configures the engine.
type EngineOption func(*Engine)

// WithImageDir sets the image extraction directory.
func WithImageDir(dir string) EngineOption {
	return func(e *Engine) {
		e.reader.WithImageDir(dir)
		e.extractor.WithOutputDir(dir)
	}
}

// WithHeadingDetector sets a custom heading detector.
func WithHeadingDetector(d *classifier.Detector) EngineOption {
	return func(e *Engine) {
		e.headingDetector = d
	}
}

// WithListDetection enables or disables list detection.
func WithListDetection(enabled bool) EngineOption {
	return func(e *Engine) {
		if !enabled {
			e.listDetector = nil
		}
	}
}

// WithRenderer sets a custom renderer.
func WithRenderer(r renderer.Renderer) EngineOption {
	return func(e *Engine) {
		e.renderer = r
	}
}

// WithInlineFormatting enables inline formatting in the renderer.
func WithInlineFormatting(enabled bool) EngineOption {
	return func(e *Engine) {
		if mr, ok := e.renderer.(*renderer.MarkdownRenderer); ok {
			mr.EnableInlineFormatting = enabled
		}
	}
}

// WithHyperlinks enables hyperlink rendering.
func WithHyperlinks(enabled bool) EngineOption {
	return func(e *Engine) {
		if mr, ok := e.renderer.(*renderer.MarkdownRenderer); ok {
			mr.EnableHyperlinks = enabled
		}
	}
}

// Result holds the conversion output.
type Result struct {
	Markdown   string
	ImageFiles []string
	Warnings   []error
}

// ConvertFile converts a DOCX file to Markdown.
func (e *Engine) ConvertFile(ctx context.Context, path string) (*Result, error) {
	parsed, err := e.reader.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse document: %w", err)
	}

	// Extract images
	archive, err := parser.OpenArchive(path)
	if err != nil {
		return nil, fmt.Errorf("open archive for extraction: %w", err)
	}
	defer archive.Close()

	images, imgWarnings := e.extractor.Extract(archive, parsed.Relationships, parsed.Document)
	parsed.Warnings = append(parsed.Warnings, imgWarnings...)

	// Build image lookup by rId
	imageMap := make(map[string]models.Image)
	for i, img := range images {
		imageMap[img.Name] = images[i]
	}

	// Apply classification
	processed := e.processDocument(parsed.Document, parsed.Styles)

	// Resolve inline image placeholders
	processed.Body = e.resolveImagePlaceholders(processed.Body, imageMap, parsed.Relationships)

	// Render
	markdown, err := e.renderer.Render(processed)
	if err != nil {
		return nil, fmt.Errorf("render document: %w", err)
	}

	imagePaths := make([]string, len(images))
	for i, img := range images {
		imagePaths[i] = img.Path
	}

	return &Result{
		Markdown:   markdown,
		ImageFiles: imagePaths,
		Warnings:   parsed.Warnings,
	}, nil
}

func (e *Engine) resolveImagePlaceholders(blocks []models.Block, imageMap map[string]models.Image, rels *models.Relationships) []models.Block {
	var result []models.Block
	for _, block := range blocks {
		if block.Kind == models.KindImage {
			img := block.Value.(*models.Image)
			if img.Name == "" {
				continue
			}
			// Try direct lookup by name
			if resolved, ok := imageMap[img.Name]; ok {
				result = append(result, models.Block{Kind: models.KindImage, Value: &resolved})
				continue
			}
			// Try lookup by rId through relationships
			if rels != nil {
				if rel, ok := rels.Items[img.Name]; ok {
					imageName := filepath.Base(rel.Target)
					if resolved, ok := imageMap[imageName]; ok {
						result = append(result, models.Block{Kind: models.KindImage, Value: &resolved})
						continue
					}
				}
			}
			// Could not resolve - skip
			continue
		}
		result = append(result, block)
	}
	return result
}

func (e *Engine) processDocument(doc *models.Document, styles *models.StyleSheet) *models.Document {
	// Step 0: Apply style font sizes to runs that don't have explicit sizes
	if styles != nil {
		e.applyStyleFontSizes(doc, styles)
	}

	// Step 1: Detect and group lists FIRST (before heading detection)
	// so that paragraphs with NumPr are preserved as list items
	if e.listDetector != nil {
		doc.Body = e.listDetector.DetectListItems(doc.Body)
		e.listDetector.ApplyStyleListInfo(doc.Body)
	}

	// Step 2: Apply heading detection to paragraphs
	// Set HeadingLevel instead of replacing the paragraph
	for i := range doc.Body {
		if doc.Body[i].Kind == models.KindParagraph {
			para := doc.Body[i].Value.(*models.Paragraph)
			if level, isHeading := e.headingDetector.DetectHeading(para, styles); isHeading {
				para.HeadingLevel = level
			}
		}
	}

	// Step 3: Handle inline images (DrawingRID)
	doc.Body = e.processInlineImages(doc.Body)

	return doc
}

func (e *Engine) applyStyleFontSizes(doc *models.Document, styles *models.StyleSheet) {
	for i := range doc.Body {
		if doc.Body[i].Kind == models.KindParagraph {
			para := doc.Body[i].Value.(*models.Paragraph)
			if para.StyleID == "" {
				continue
			}
			style, ok := styles.Styles[para.StyleID]
			if !ok {
				continue
			}
			if style.FontSize == 0 {
				continue
			}
			for j := range para.Runs {
				if para.Runs[j].FontSize == 0 {
					para.Runs[j].FontSize = style.FontSize
				}
			}
		}
	}
}

func (e *Engine) processInlineImages(blocks []models.Block) []models.Block {
	var result []models.Block
	for _, block := range blocks {
		if block.Kind == models.KindParagraph {
			para := block.Value.(*models.Paragraph)
			var textRuns []models.TextRun
			var hasImages bool
			for _, run := range para.Runs {
				if run.DrawingRID != "" {
					// Flush any pending text runs
					if len(textRuns) > 0 {
						newPara := *para
						newPara.Runs = textRuns
						result = append(result, models.Block{Kind: models.KindParagraph, Value: &newPara})
						textRuns = nil
					}
					// Add placeholder image (will be resolved later)
					result = append(result, models.Block{
						Kind:  models.KindImage,
						Value: &models.Image{Name: run.DrawingRID},
					})
					hasImages = true
					continue
				}
				textRuns = append(textRuns, run)
			}
			if len(textRuns) > 0 || !hasImages {
				newPara := *para
				newPara.Runs = textRuns
				result = append(result, models.Block{Kind: models.KindParagraph, Value: &newPara})
			}
		} else {
			result = append(result, block)
		}
	}
	return result
}
