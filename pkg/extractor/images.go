package extractor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/saiful-anwar/word2md/pkg/models"
	"github.com/saiful-anwar/word2md/pkg/parser"
)

// ImageExtractor handles extraction of images from DOCX archives.
type ImageExtractor struct {
	outputDir string
}

// NewImageExtractor creates a new extractor.
func NewImageExtractor() *ImageExtractor {
	return &ImageExtractor{}
}

// WithOutputDir sets the output directory for extracted images.
func (e *ImageExtractor) WithOutputDir(dir string) *ImageExtractor {
	e.outputDir = dir
	return e
}

// Extract pulls all images referenced in the parsed document.
func (e *ImageExtractor) Extract(archive *parser.Archive, rels *models.Relationships, doc *models.Document) ([]models.Image, []error) {
	var images []models.Image
	var warnings []error

	if e.outputDir == "" {
		return images, warnings
	}

	// Clean and validate output directory
	cleanDir, err := filepath.Abs(e.outputDir)
	if err != nil {
		return images, []error{fmt.Errorf("resolve image directory: %w", err)}
	}

	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return images, []error{fmt.Errorf("create image directory: %w", err)}
	}

	// Extract all media files referenced in relationships
	if rels != nil {
		for _, rel := range rels.Items {
			if isImageRelationship(rel.Type) {
				imageName := filepath.Base(rel.Target)
				imagePath := filepath.Join(cleanDir, imageName)

				if err := e.extractImageFile(archive, rel.Target, imagePath, cleanDir); err != nil {
					warnings = append(warnings, fmt.Errorf("extract image %s: %w", imageName, err))
					continue
				}

				// Store relative path for markdown
				relPath, _ := filepath.Rel(filepath.Dir(cleanDir), imagePath)
				if relPath == "" {
					relPath = imageName
				}

				images = append(images, models.Image{
					Name: imageName,
					Path: relPath,
				})
			}
		}
	}

	return images, warnings
}

func (e *ImageExtractor) extractImageFile(archive *parser.Archive, target string, destPath string, outputDir string) error {
	// Prevent path traversal in target
	cleanTarget := filepath.Clean(target)
	if strings.Contains(cleanTarget, "..") {
		return fmt.Errorf("invalid path traversal in target: %s", target)
	}

	mediaPath := "word/" + target
	if strings.HasPrefix(target, "/") {
		mediaPath = target
	}

	f, ok := archive.File(mediaPath)
	if !ok {
		// Try without word/ prefix
		f, ok = archive.File(target)
		if !ok {
			return fmt.Errorf("image file not found: %s", target)
		}
	}

	// Verify destination path is within output directory
	absDest, err := filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("resolve destination path: %w", err)
	}
	rel, err := filepath.Rel(outputDir, absDest)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("destination path escapes output directory: %s", destPath)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open image file: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read image file: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("write image file: %w", err)
	}

	return nil
}

func isImageRelationship(relType string) bool {
	return strings.Contains(relType, "image")
}
