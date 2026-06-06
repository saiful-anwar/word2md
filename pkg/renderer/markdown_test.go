package renderer

import (
	"strings"
	"testing"

	"github.com/saiful-anwar/word2md/pkg/models"
)

func TestNewMarkdownRenderer(t *testing.T) {
	r := NewMarkdownRenderer()
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
	if !r.EnableInlineFormatting {
		t.Error("expected inline formatting enabled by default")
	}
}

func TestRenderEmptyDocument(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{}
	result, err := r.Render(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderParagraph(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{
						{Text: "Hello, world!"},
					},
				},
			},
		},
	}
	result, err := r.Render(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Hello, world!\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderParagraphWithInlineFormatting(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{
						{Text: "bold text", Bold: true},
						{Text: " and "},
						{Text: "italic text", Italic: true},
						{Text: " and "},
						{Text: "code", Code: true},
					},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	if !strings.Contains(result, "**bold text**") {
		t.Errorf("expected bold text, got %q", result)
	}
	if !strings.Contains(result, "*italic text*") {
		t.Errorf("expected italic text, got %q", result)
	}
	if !strings.Contains(result, "`code`") {
		t.Errorf("expected code, got %q", result)
	}
}

func TestRenderParagraphWithStrikethrough(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{
						{Text: "deleted", Strikethrough: true},
					},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	if !strings.Contains(result, "~~deleted~~") {
		t.Errorf("expected strikethrough, got %q", result)
	}
}

func TestRenderInlineFormattingDisabled(t *testing.T) {
	r := NewMarkdownRenderer()
	r.EnableInlineFormatting = false
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{
						{Text: "bold text", Bold: true},
					},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	if strings.Contains(result, "**") {
		t.Errorf("expected no bold formatting, got %q", result)
	}
}

func TestRenderTable(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindTable,
				Value: &models.Table{
					Rows: []models.TableRow{
						{
							Cells: []models.TableCell{
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "Name"}}}}}},
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "Age"}}}}}},
							},
						},
						{
							Cells: []models.TableCell{
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "Alice"}}}}}},
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "30"}}}}}},
							},
						},
					},
				},
			},
		},
	}
	result, err := r.Render(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "| Name") {
		t.Errorf("expected table header, got %q", result)
	}
	if !strings.Contains(result, "| ----") {
		t.Errorf("expected separator row, got %q", result)
	}
	if !strings.Contains(result, "| Alice") {
		t.Errorf("expected table content, got %q", result)
	}
}

func TestRenderImage(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindImage,
				Value: &models.Image{
					Name: "test.png",
					Path: "images/test.png",
				},
			},
		},
	}
	result, _ := r.Render(doc)
	expected := "![test.png](images/test.png)\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderImageWithAltText(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindImage,
				Value: &models.Image{
					Name:    "test.png",
					Path:    "images/test.png",
					AltText: "My Image",
				},
			},
		},
	}
	result, _ := r.Render(doc)
	expected := "![My Image](images/test.png)\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderListNumbered(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindList,
				Value: &models.List{
					Items: []models.Paragraph{
						{Runs: []models.TextRun{{Text: "First item"}}},
						{Runs: []models.TextRun{{Text: "Second item"}}},
						{Runs: []models.TextRun{{Text: "Third item"}}},
					},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	if !strings.Contains(result, "1. First item") {
		t.Errorf("expected numbered list, got %q", result)
	}
	if !strings.Contains(result, "2. Second item") {
		t.Errorf("expected numbered list, got %q", result)
	}
}

func TestRenderListBullet(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindList,
				Value: &models.List{
					IsBullet: true,
					Items: []models.Paragraph{
						{Runs: []models.TextRun{{Text: "Apple"}}},
						{Runs: []models.TextRun{{Text: "Banana"}}},
					},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	if !strings.Contains(result, "- Apple") {
		t.Errorf("expected bullet list, got %q", result)
	}
	if !strings.Contains(result, "- Banana") {
		t.Errorf("expected bullet list, got %q", result)
	}
}

func TestRenderHyperlinks(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{
						{Text: "example", Hyperlink: &models.Hyperlink{URL: "https://example.com"}},
					},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	expected := "[example](https://example.com)\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderMultipleBlocks(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{{Text: "First para"}},
				},
			},
			{
				Kind: models.KindParagraph,
				Value: &models.Paragraph{
					Runs: []models.TextRun{{Text: "Second para"}},
				},
			},
		},
	}
	result, _ := r.Render(doc)
	expected := "First para\n\nSecond para\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 5},
		{"hello world", 11},
		{"你好", 4},
		{"日本語", 6},
		{"한글", 4},
		{" café ", 6},
		{"A\nB", 2},
		{"A\tB", 2},
	}
	for _, tt := range tests {
		got := stringWidth(tt.input)
		if got != tt.want {
			t.Errorf("stringWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestRenderTableCJK(t *testing.T) {
	r := NewMarkdownRenderer()
	doc := &models.Document{
		Body: []models.Block{
			{
				Kind: models.KindTable,
				Value: &models.Table{
					Rows: []models.TableRow{
						{
							Cells: []models.TableCell{
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "名前"}}}}}},
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "年齢"}}}}}},
							},
						},
						{
							Cells: []models.TableCell{
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "太郎"}}}}}},
								{Blocks: []models.Block{{Kind: models.KindParagraph, Value: &models.Paragraph{Runs: []models.TextRun{{Text: "30"}}}}}},
							},
						},
					},
				},
			},
		},
	}
	result, err := r.Render(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}
	// Verify header and separator are aligned
	if !strings.Contains(lines[0], "| 名前 | 年齢 |") {
		t.Errorf("expected aligned CJK header, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "| ---- | ---- |") {
		t.Errorf("expected aligned separator, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "| 太郎 | 30   |") {
		t.Errorf("expected aligned CJK row, got %q", lines[2])
	}
}
