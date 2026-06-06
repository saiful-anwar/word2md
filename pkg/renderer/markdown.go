package renderer

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/saiful-anwar/word2md/pkg/classifier"
	"github.com/saiful-anwar/word2md/pkg/models"
)

// Renderer is the interface for output format generation.
type Renderer interface {
	Render(doc *models.Document) (string, error)
}

// MarkdownRenderer generates Markdown output.
type MarkdownRenderer struct {
	EnableInlineFormatting bool
	EnableHyperlinks       bool
}

// NewMarkdownRenderer creates a MarkdownRenderer with all formatting enabled.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		EnableInlineFormatting: true,
		EnableHyperlinks:       true,
	}
}

// Render converts a Document model to Markdown string.
func (mr *MarkdownRenderer) Render(doc *models.Document) (string, error) {
	var buf bytes.Buffer

	for i, block := range doc.Body {
		if i > 0 {
			buf.WriteString("\n")
		}

		rendered, err := mr.renderBlock(block)
		if err != nil {
			return "", fmt.Errorf("render block %d: %w", i, err)
		}
		buf.WriteString(rendered)
	}

	return buf.String(), nil
}

func (mr *MarkdownRenderer) renderBlock(block models.Block) (string, error) {
	switch block.Kind {
	case models.KindParagraph:
		para := block.Value.(*models.Paragraph)
		return mr.renderParagraph(para), nil
	case models.KindTable:
		tbl := block.Value.(*models.Table)
		return mr.renderTable(tbl), nil
	case models.KindImage:
		img := block.Value.(*models.Image)
		return mr.renderImage(img), nil
	case models.KindList:
		list := block.Value.(*models.List)
		return mr.renderList(list), nil
	default:
		return "", nil
	}
}

func (mr *MarkdownRenderer) renderParagraph(para *models.Paragraph) string {
	if para == nil {
		return ""
	}

	var buf bytes.Buffer
	for _, run := range para.Runs {
		buf.WriteString(mr.renderRun(&run))
	}

	text := buf.String()
	text = strings.TrimSpace(text)

	if text == "" {
		return ""
	}

	// If heading level is set, render as heading
	if para.HeadingLevel > 0 && para.HeadingLevel <= 6 {
		prefix := strings.Repeat("#", para.HeadingLevel) + " "
		return prefix + text + "\n"
	}

	return text + "\n"
}

func (mr *MarkdownRenderer) renderRun(run *models.TextRun) string {
	if run == nil {
		return ""
	}

	text := run.Text

	if mr.EnableInlineFormatting {
		if run.Code && text != "" {
			return "`" + text + "`"
		}
		if run.Bold && text != "" {
			text = "**" + text + "**"
		}
		if run.Italic && text != "" {
			text = "*" + text + "*"
		}
		if run.Strikethrough && text != "" {
			text = "~~" + text + "~~"
		}
	}

	if mr.EnableHyperlinks && run.Hyperlink != nil && run.Hyperlink.URL != "" {
		text = fmt.Sprintf("[%s](%s)", text, run.Hyperlink.URL)
	}

	return text
}

func (mr *MarkdownRenderer) renderTable(tbl *models.Table) string {
	if len(tbl.Rows) == 0 {
		return ""
	}

	// Calculate max column count
	maxCols := 0
	for _, row := range tbl.Rows {
		if len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}

	if maxCols == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, maxCols)
	for _, row := range tbl.Rows {
		for j, cell := range row.Cells {
			cellText := mr.extractCellText(&cell)
			w := stringWidth(cellText)
			if w > colWidths[j] {
				colWidths[j] = w
			}
		}
	}

	// Ensure minimum width
	for j := range colWidths {
		if colWidths[j] < 1 {
			colWidths[j] = 1
		}
	}

	var buf bytes.Buffer

	for i, row := range tbl.Rows {
		// Render cells
		for j, cell := range row.Cells {
			cellText := mr.extractCellText(&cell)
			w := stringWidth(cellText)
			pad := colWidths[j] - w
			buf.WriteString("| ")
			buf.WriteString(cellText)
			buf.WriteString(strings.Repeat(" ", pad))
			buf.WriteString(" ")
		}
		// Fill empty columns
		for j := len(row.Cells); j < maxCols; j++ {
			buf.WriteString("| ")
			buf.WriteString(strings.Repeat(" ", colWidths[j]))
			buf.WriteString(" ")
		}
		buf.WriteString("|\n")

		// Render separator after header row
		if i == 0 {
			for j := 0; j < maxCols; j++ {
				buf.WriteString("| ")
				buf.WriteString(strings.Repeat("-", colWidths[j]))
				buf.WriteString(" ")
			}
			buf.WriteString("|\n")
		}
	}

	return buf.String()
}

func (mr *MarkdownRenderer) extractCellText(cell *models.TableCell) string {
	var buf bytes.Buffer
	for _, block := range cell.Blocks {
		if block.Kind == models.KindParagraph {
			para := block.Value.(*models.Paragraph)
			for _, run := range para.Runs {
				buf.WriteString(run.Text)
			}
		}
	}
	return strings.TrimSpace(buf.String())
}

func (mr *MarkdownRenderer) renderImage(img *models.Image) string {
	alt := img.AltText
	if alt == "" {
		alt = img.Name
	}
	return fmt.Sprintf("![%s](%s)\n", alt, img.Path)
}

func (mr *MarkdownRenderer) renderList(list *models.List) string {
	var buf bytes.Buffer

	for i, item := range list.Items {
		var text string
		if list.IsBullet {
			text = "- " + mr.extractParagraphText(&item)
		} else {
			text = fmt.Sprintf("%d. %s", i+1, mr.extractParagraphText(&item))
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String()
}

func (mr *MarkdownRenderer) extractParagraphText(para *models.Paragraph) string {
	var buf bytes.Buffer
	for _, run := range para.Runs {
		buf.WriteString(mr.renderRun(&run))
	}
	return strings.TrimSpace(buf.String())
}

// HeadingRenderer wraps a MarkdownRenderer and adds heading detection.
type HeadingRenderer struct {
	*MarkdownRenderer
	detector *classifier.Detector
}

func NewHeadingRenderer(detector *classifier.Detector) *HeadingRenderer {
	return &HeadingRenderer{
		MarkdownRenderer: NewMarkdownRenderer(),
		detector:         detector,
	}
}

func (hr *HeadingRenderer) RenderParagraphWithHeading(para *models.Paragraph, styles *models.StyleSheet) string {
	if level, ok := hr.detector.DetectHeading(para, styles); ok {
		text := hr.extractParagraphText(para)
		if text != "" {
			prefix := strings.Repeat("#", level) + " "
			return prefix + text + "\n"
		}
	}
	return hr.MarkdownRenderer.renderParagraph(para)
}

// wideRanges defines Unicode blocks whose runes occupy two display columns.
var wideRanges = []struct{ start, end rune }{
	{0x1100, 0x115F},   // Hangul Jamo
	{0x2E80, 0x2FDF},   // CJK Radicals / Kangxi
	{0x3000, 0x303F},   // CJK Symbols & Punctuation
	{0x3040, 0x309F},   // Hiragana
	{0x30A0, 0x30FF},   // Katakana
	{0x3100, 0x312F},   // Bopomofo
	{0x3130, 0x318F},   // Hangul Compatibility Jamo
	{0x31A0, 0x31BF},   // Bopomofo Extended
	{0x31C0, 0x31EF},   // CJK Strokes
	{0x3200, 0x33FF},   // Enclosed CJK / CJK Compatibility
	{0x3400, 0x4DBF},   // CJK Extension A
	{0x4E00, 0x9FFF},   // CJK Unified Ideographs
	{0xAC00, 0xD7AF},   // Hangul Syllables
	{0xF900, 0xFAFF},   // CJK Compatibility Ideographs
	{0xFE10, 0xFE19},   // Vertical Forms
	{0xFE30, 0xFE6F},   // CJK Compatibility Forms
	{0xFF01, 0xFF60},   // Fullwidth Forms
	{0xFFE0, 0xFFE6},   // Fullwidth Signs
	{0x1B000, 0x1B0FF}, // Kana Supplement
	{0x1F200, 0x1F2FF}, // Enclosed Ideographic Supplement
	{0x20000, 0x2A6DF}, // CJK Extension B
	{0x2A700, 0x2B73F}, // CJK Extension C
	{0x2B740, 0x2B81F}, // CJK Extension D
	{0x2B820, 0x2CEAF}, // CJK Extension E
	{0x2F800, 0x2FA1F}, // CJK Compatibility Supplement
	{0x30000, 0x3134F}, // CJK Extension G
}

// isWideRune reports whether r is a rune that occupies two display columns.
func isWideRune(r rune) bool {
	for _, rng := range wideRanges {
		if r >= rng.start && r <= rng.end {
			return true
		}
	}
	return false
}

// stringWidth returns the visual display width of s.
// CJK and Hangul characters count as 2 columns; printable ASCII and Latin as 1.
// Non-printable and control characters are ignored.
func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if isWideRune(r) {
			w += 2
		} else if unicode.IsPrint(r) {
			w++
		}
	}
	return w
}
