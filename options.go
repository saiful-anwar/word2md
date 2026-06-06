package word2md

// config holds all conversion options.
type config struct {
	imageDir          string
	inlineFormatting  bool
	hyperlinks        bool
	headingDetection  bool
	listDetection     bool
	headingThresholds struct {
		h1, h2, h3 int
	}
}

func defaultConfig() *config {
	return &config{
		imageDir:         "",
		inlineFormatting: true,
		hyperlinks:       true,
		headingDetection: true,
		listDetection:    true,
		headingThresholds: struct {
			h1, h2, h3 int
		}{
			h1: 48,
			h2: 36,
			h3: 28,
		},
	}
}

// Option is a function that configures conversion behavior.
type Option func(*config)

// WithImageDir sets the directory for extracted images.
func WithImageDir(dir string) Option {
	return func(c *config) {
		c.imageDir = dir
	}
}

// WithInlineFormatting enables or disables bold, italic, code, and strikethrough formatting.
func WithInlineFormatting(enabled bool) Option {
	return func(c *config) {
		c.inlineFormatting = enabled
	}
}

// WithHyperlinks enables or disables hyperlink conversion.
func WithHyperlinks(enabled bool) Option {
	return func(c *config) {
		c.hyperlinks = enabled
	}
}

// WithHeadingDetection enables or disables heading detection.
func WithHeadingDetection(enabled bool) Option {
	return func(c *config) {
		c.headingDetection = enabled
	}
}

// WithListDetection enables or disables list detection.
func WithListDetection(enabled bool) Option {
	return func(c *config) {
		c.listDetection = enabled
	}
}

// WithHeadingThresholds sets custom font size thresholds for heading detection.
// Values are in half-points (e.g., 48 = 24pt).
func WithHeadingThresholds(h1, h2, h3 int) Option {
	return func(c *config) {
		c.headingThresholds.h1 = h1
		c.headingThresholds.h2 = h2
		c.headingThresholds.h3 = h3
	}
}
