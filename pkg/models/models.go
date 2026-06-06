package models

// Block represents a top-level element in a document body.
type Block struct {
	Kind  BlockKind
	Value interface{}
}

// BlockKind distinguishes the type of a document block.
type BlockKind int

const (
	// KindParagraph is a text paragraph.
	KindParagraph BlockKind = iota
	// KindTable is a table element.
	KindTable
	// KindImage is an image element.
	KindImage
	// KindList is a grouped list of items.
	KindList
)

// Document is the root model after parsing.
type Document struct {
	Body     []Block
	Metadata DocumentMetadata
}

// DocumentMetadata holds extracted metadata from the docx.
type DocumentMetadata struct {
	Title    string
	Author   string
	Subject  string
	Keywords string
}

// Paragraph is a block of text runs.
type Paragraph struct {
	Runs         []TextRun
	StyleID      string
	NumPr        *NumberingProperties
	Alignment    string
	Indent       *ParagraphIndent
	HeadingLevel int // 0 = not a heading, 1-6 = heading level
}

// NumberingProperties indicates a list item.
type NumberingProperties struct {
	NumID    int
	Ilvl     int
	IsBullet bool
}

// ParagraphIndent holds indentation info.
type ParagraphIndent struct {
	Left    int
	Right   int
	First   int
	Hanging int
}

// TextRun is a span of text with uniform formatting.
type TextRun struct {
	Text          string
	Bold          bool
	Italic        bool
	Strikethrough bool
	Code          bool
	FontSize      int // in half-points (e.g., 24 = 12pt)
	FontName      string
	Color         string
	Hyperlink     *Hyperlink
	DrawingRID    string // rId for inline images (empty if no image)
}

// Hyperlink holds link info.
type Hyperlink struct {
	URL   string
	Title string
}

// Table is a grid of rows and cells.
type Table struct {
	Rows []TableRow
}

// TableRow is a row of cells.
type TableRow struct {
	Cells []TableCell
}

// TableCell is a cell containing blocks.
type TableCell struct {
	Blocks []Block
}

// Image is a reference to an extracted image.
type Image struct {
	Name    string
	Path    string
	AltText string
}

// List is a container for list items (paragraphs with NumPr).
type List struct {
	Items    []Paragraph
	NumID    int
	IsBullet bool
}

// Style represents a Word style definition.
type Style struct {
	StyleID  string
	Name     string
	Type     string
	FontSize int
	Bold     bool
	Italic   bool
	BasedOn  string
	Next     string
}

// StyleSheet is the collection of styles from styles.xml.
type StyleSheet struct {
	Styles map[string]Style
}

// Relationship maps a resource ID to a target.
type Relationship struct {
	ID     string
	Type   string
	Target string
}

// Relationships is the collection from .rels files.
type Relationships struct {
	Items map[string]Relationship
}
