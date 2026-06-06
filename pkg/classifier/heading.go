package classifier

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/saiful-anwar/word2md/pkg/models"
)

// Detector holds configuration for heading detection.
type Detector struct {
	UseFontSize       bool
	UseStyleNames     bool
	UseOutlineLevels  bool
	HeadingThresholds Thresholds
}

// Thresholds define font size boundaries for heading levels.
// Values are in half-points (e.g., 24 = 12pt).
type Thresholds struct {
	H1 int
	H2 int
	H3 int
	H4 int
}

// DefaultThresholds returns sensible defaults.
func DefaultThresholds() Thresholds {
	return Thresholds{
		H1: 48,
		H2: 36,
		H3: 28,
		H4: 24,
	}
}

// NewDetector creates a detector with default settings.
func NewDetector() *Detector {
	return &Detector{
		UseFontSize:       true,
		UseStyleNames:     true,
		UseOutlineLevels:  true,
		HeadingThresholds: DefaultThresholds(),
	}
}

// DetectHeading determines if a paragraph is a heading and its level.
// Returns (level, true) if it's a heading, (0, false) otherwise.
// Level is 1-6.
func (d *Detector) DetectHeading(para *models.Paragraph, styles *models.StyleSheet) (int, bool) {
	if d.UseStyleNames {
		level := d.detectByStyleName(para, styles)
		if level > 0 {
			return level, true
		}
	}

	if d.UseOutlineLevels {
		level := d.detectByOutlineLevel(para, styles)
		if level > 0 {
			return level, true
		}
	}

	if d.UseFontSize {
		level := d.detectByFontSize(para)
		if level > 0 {
			return level, true
		}
	}

	return 0, false
}

// headingPatterns maps known heading style name fragments to their level.
var headingPatterns = []struct {
	patterns []string
	level    int
}{
	{[]string{"heading 1", "heading1", "标题 1", "标题1"}, 1},
	{[]string{"heading 2", "heading2", "标题 2", "标题2"}, 2},
	{[]string{"heading 3", "heading3", "标题 3", "标题3"}, 3},
	{[]string{"heading 4", "heading4", "标题 4", "标题4"}, 4},
	{[]string{"heading 5", "heading5", "标题 5", "标题5"}, 5},
	{[]string{"heading 6", "heading6", "标题 6", "标题6"}, 6},
}

func (d *Detector) detectByStyleName(para *models.Paragraph, styles *models.StyleSheet) int {
	if styles == nil || para.StyleID == "" {
		return 0
	}

	style, ok := styles.Styles[para.StyleID]
	if !ok {
		return 0
	}

	name := strings.ToLower(style.Name)
	for _, hp := range headingPatterns {
		for _, p := range hp.patterns {
			if strings.Contains(name, p) {
				return hp.level
			}
		}
	}
	return 0
}

func (d *Detector) detectByOutlineLevel(para *models.Paragraph, styles *models.StyleSheet) int {
	if styles == nil || para.StyleID == "" {
		return 0
	}

	style, ok := styles.Styles[para.StyleID]
	if !ok {
		return 0
	}

	// Check outline level in style name
	name := strings.ToLower(style.Name)
	level := extractOutlineLevel(name)
	if level > 0 {
		return level
	}

	return 0
}

func (d *Detector) detectByFontSize(para *models.Paragraph) int {
	text := d.extractText(para)
	if text == "" {
		return 0
	}

	minSize := d.minFontSize(para)
	if minSize == 0 {
		return 0
	}

	// Check for numbered/chapter prefixes
	isPrefixed := HasPrefixMarker(text)
	maxLen := 15
	if isPrefixed || para.NumPr != nil {
		maxLen = 45
	}

	th := d.HeadingThresholds

	if th.H1 <= minSize {
		return 1
	}
	if th.H2 <= minSize {
		return 2
	}
	if th.H3 <= minSize && len(text) < maxLen {
		return 3
	}
	if th.H4 <= minSize && len(text) < maxLen && isPrefixed {
		return 4
	}

	return 0
}

func (d *Detector) extractText(para *models.Paragraph) string {
	var sb strings.Builder
	for _, run := range para.Runs {
		sb.WriteString(run.Text)
	}
	return strings.TrimSpace(sb.String())
}

func (d *Detector) minFontSize(para *models.Paragraph) int {
	minSize := 0
	for _, run := range para.Runs {
		if run.FontSize > 0 && (minSize == 0 || run.FontSize < minSize) {
			minSize = run.FontSize
		}
	}
	return minSize
}

func extractOutlineLevel(name string) int {
	re := regexp.MustCompile(`outlinelevel(\d+)`)
	matches := re.FindStringSubmatch(name)
	if len(matches) > 1 {
		level := 0
		fmt.Sscanf(matches[1], "%d", &level)
		if level >= 1 && level <= 6 {
			return level
		}
	}
	return 0
}

var (
	arabicNumberRe    = regexp.MustCompile(`^\d`)
	chineseNumberRe   = regexp.MustCompile(`^[一二三四五六七八九十]`)
	diArabicNumberRe  = regexp.MustCompile(`^第\d+`)
	diChineseNumberRe = regexp.MustCompile(`^第[一二三四五六七八九十]`)
)

// HasPrefixMarker checks if text starts with a heading-like marker.
func HasPrefixMarker(s string) bool {
	if s == "" {
		return false
	}
	return arabicNumberRe.MatchString(s) ||
		chineseNumberRe.MatchString(s) ||
		diArabicNumberRe.MatchString(s) ||
		diChineseNumberRe.MatchString(s)
}
