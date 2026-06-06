package word2md

// Result holds the output of a conversion.
type Result struct {
	Markdown   string
	ImageFiles []string
	Warnings   []error
}

// IsWarning returns true if there are non-fatal warnings.
func (r *Result) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// WarningCount returns the number of warnings.
func (r *Result) WarningCount() int {
	return len(r.Warnings)
}
