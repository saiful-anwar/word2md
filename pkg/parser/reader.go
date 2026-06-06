package parser

import (
	"fmt"

	"github.com/saiful-anwar/word2md/pkg/models"
)

// Reader is the unified DOCX parser interface.
type Reader interface {
	Parse(path string) (*ParsedDocument, error)
}

// ParsedDocument contains all extracted data from a DOCX file.
type ParsedDocument struct {
	Document      *models.Document
	Styles        *models.StyleSheet
	Relationships *models.Relationships
	Images        []models.Image
	Warnings      []error
}

// DefaultReader implements Reader using the internal parser packages.
type DefaultReader struct {
	imageDir string
}

// NewReader creates a new DefaultReader.
func NewReader() *DefaultReader {
	return &DefaultReader{}
}

// WithImageDir sets the directory for extracted images.
func (r *DefaultReader) WithImageDir(dir string) *DefaultReader {
	r.imageDir = dir
	return r
}

// Parse parses a DOCX file at the given path.
func (r *DefaultReader) Parse(path string) (*ParsedDocument, error) {
	archive, err := OpenArchive(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	parsed := &ParsedDocument{
		Warnings: make([]error, 0),
	}

	// Read document.xml
	docData, err := archive.ReadFile("word/document.xml")
	if err != nil {
		return nil, fmt.Errorf("read document.xml: %w", err)
	}

	docReader := NewDocumentReader()
	parsed.Document, err = docReader.Read(docData)
	if err != nil {
		return nil, fmt.Errorf("parse document: %w", err)
	}

	// Read styles.xml
	if archive.FileExists("word/styles.xml") {
		stylesData, err := archive.ReadFile("word/styles.xml")
		if err != nil {
			parsed.Warnings = append(parsed.Warnings, fmt.Errorf("read styles.xml: %w", err))
		} else {
			stylesReader := NewStylesReader()
			parsed.Styles, err = stylesReader.Read(stylesData)
			if err != nil {
				parsed.Warnings = append(parsed.Warnings, fmt.Errorf("parse styles: %w", err))
			}
		}
	}

	// Read relationships
	if archive.FileExists("word/_rels/document.xml.rels") {
		relsData, err := archive.ReadFile("word/_rels/document.xml.rels")
		if err != nil {
			parsed.Warnings = append(parsed.Warnings, fmt.Errorf("read relationships: %w", err))
		} else {
			relsReader := NewRelationshipsReader()
			parsed.Relationships, err = relsReader.Read(relsData)
			if err != nil {
				parsed.Warnings = append(parsed.Warnings, fmt.Errorf("parse relationships: %w", err))
			}
		}
	}

	// Resolve hyperlinks
	if parsed.Relationships != nil {
		for i := range parsed.Document.Body {
			if parsed.Document.Body[i].Kind == models.KindParagraph {
				para := parsed.Document.Body[i].Value.(*models.Paragraph)
				for j := range para.Runs {
					if para.Runs[j].Hyperlink != nil {
						if rel, ok := parsed.Relationships.Items[para.Runs[j].Hyperlink.URL]; ok {
							para.Runs[j].Hyperlink.URL = rel.Target
						}
					}
				}
			}
		}
	}

	return parsed, nil
}
