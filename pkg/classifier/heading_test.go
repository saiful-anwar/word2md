package classifier

import (
	"testing"

	"github.com/saiful-anwar/word2md/pkg/models"
)

func TestNewDetector(t *testing.T) {
	d := NewDetector()
	if d == nil {
		t.Fatal("expected non-nil detector")
	}
	if !d.UseFontSize {
		t.Error("expected UseFontSize to be true by default")
	}
}

func TestDetectByFontSizeH1(t *testing.T) {
	d := NewDetector()
	para := &models.Paragraph{
		Runs: []models.TextRun{
			{Text: "Big Heading", FontSize: 48},
		},
	}
	level, ok := d.DetectHeading(para, nil)
	if !ok || level != 1 {
		t.Errorf("expected heading level 1, got %d (ok=%v)", level, ok)
	}
}

func TestDetectByFontSizeH2(t *testing.T) {
	d := NewDetector()
	para := &models.Paragraph{
		Runs: []models.TextRun{
			{Text: "Medium Heading", FontSize: 40},
		},
	}
	_, ok := d.DetectHeading(para, nil)
	if !ok {
		t.Error("expected heading detection")
	}
}

func TestDetectByFontSizeH3(t *testing.T) {
	d := NewDetector()
	// Short text with font size in h3 range
	para := &models.Paragraph{
		Runs: []models.TextRun{
			{Text: "Short", FontSize: 30},
		},
	}
	level, ok := d.DetectHeading(para, nil)
	if !ok || level != 3 {
		t.Errorf("expected heading level 3, got %d (ok=%v)", level, ok)
	}
}

func TestDetectByFontSizeH3TooLong(t *testing.T) {
	d := NewDetector()
	// Long text should not be detected as h3
	para := &models.Paragraph{
		Runs: []models.TextRun{
			{Text: "This Is A Very Long Paragraph That Should Not Be A Heading", FontSize: 28},
		},
	}
	_, ok := d.DetectHeading(para, nil)
	if ok {
		t.Error("long text should not be detected as heading without prefix")
	}
}

func TestDetectByFontSizeH3WithPrefix(t *testing.T) {
	d := NewDetector()
	// Long text with number prefix should be heading
	para := &models.Paragraph{
		Runs: []models.TextRun{
			{Text: "12345 Section with Longer Title Allowed", FontSize: 28},
		},
	}
	level, ok := d.DetectHeading(para, nil)
	if !ok || level != 3 {
		t.Errorf("expected heading level 3 with prefix, got %d (ok=%v)", level, ok)
	}
}

func TestDetectByStyleName(t *testing.T) {
	d := NewDetector()
	styles := &models.StyleSheet{
		Styles: map[string]models.Style{
			"heading1": {
				StyleID:  "heading1",
				Name:     "heading 1",
				FontSize: 48,
			},
		},
	}
	para := &models.Paragraph{
		StyleID: "heading1",
		Runs:    []models.TextRun{{Text: "Styled Header", FontSize: 24}},
	}
	level, ok := d.DetectHeading(para, styles)
	if !ok || level != 1 {
		t.Errorf("expected heading level 1 from style, got %d (ok=%v)", level, ok)
	}
}

func TestDetectByStyleNamePrecedence(t *testing.T) {
	d := NewDetector()
	// Style name should take precedence over font size
	styles := &models.StyleSheet{
		Styles: map[string]models.Style{
			"heading1": {
				StyleID:  "heading1",
				Name:     "heading 1",
				FontSize: 48,
			},
		},
	}
	para := &models.Paragraph{
		StyleID: "heading1",
		Runs:    []models.TextRun{{Text: "Still Heading 1", FontSize: 20}},
	}
	level, ok := d.DetectHeading(para, styles)
	if !ok || level != 1 {
		t.Errorf("expected heading level 1 despite small font, got %d (ok=%v)", level, ok)
	}
}

func TestNotHeading(t *testing.T) {
	d := NewDetector()
	para := &models.Paragraph{
		Runs: []models.TextRun{
			{Text: "This is a normal paragraph of body text", FontSize: 24},
		},
	}
	_, ok := d.DetectHeading(para, nil)
	if ok {
		t.Error("normal body text should not be detected as heading")
	}
}

func TestHasPrefixMarker(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"12345 Section Title", true},
		{"Chapter 1 Introduction", false}, // doesn't start with "第" + number
		{"第1章 Introduction", true},
		{"一、Introduction", true},
		{"第2节 Example", true},
		{"Normal Text", false},
		{"", false},
	}
	for _, tt := range tests {
		got := HasPrefixMarker(tt.input)
		if got != tt.want {
			t.Errorf("HasPrefixMarker(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestEmptyParagraph(t *testing.T) {
	d := NewDetector()
	para := &models.Paragraph{}
	_, ok := d.DetectHeading(para, nil)
	if ok {
		t.Error("empty paragraph should not be a heading")
	}
}
