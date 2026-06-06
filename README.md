# word2md

[![Go Reference](https://img.shields.io/badge/go-reference-blue)](https://pkg.go.dev/github.com/saiful-anwar/word2md)
[![Go Report Card](https://goreportcard.com/badge/github.com/saiful-anwar/word2md)](https://goreportcard.com/report/github.com/saiful-anwar/word2md)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

> **word2md** — Convert Microsoft Word (.docx) documents to clean Markdown with smart heading detection, table support, image extraction, and rich formatting.

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start (Library)](#quick-start-library)
- [CLI Usage](#cli-usage)
- [Advanced Configuration](#advanced-configuration)
- [Supported Elements](#supported-elements)
- [Heading Detection](#heading-detection)
- [Architecture](#architecture)
- [API Reference](#api-reference)
- [Examples](#examples)
- [License](#license)

---

## Features

- **Library-first** — Clean, idiomatic Go API. Import and use in any project.
- **Smart heading detection** — Combines Word outline levels, style names, and font size heuristics for accurate heading recognition.
- **Rich formatting** — Bold, italic, strikethrough, inline code, and hyperlinks.
- **Lists** — Ordered (numbered) and unordered (bullet) lists with proper indentation.
- **Tables** — Properly aligned Markdown tables with CJK character support via `go-runewidth`.
- **Images** — Automatic image extraction and Markdown reference generation.
- **Context-aware** — Supports `context.Context` for cancellation and timeouts.
- **Extensible** — Pluggable renderer architecture allows custom output formats (HTML, JSON, etc.).
- **No global state** — Pure functions and composable components.
- **Warning reporting** — Non-fatal issues (missing styles, corrupt images) are reported as warnings without aborting conversion.

---

## Installation

### As a Library

```bash
go get github.com/saiful-anwar/word2md
```

### CLI Tool

```bash
go install github.com/saiful-anwar/word2md/cmd/word2md@latest
```

---

## Quick Start (Library)

### Minimal Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/saiful-anwar/word2md"
)

func main() {
    ctx := context.Background()

    result, err := word2md.ConvertFile(ctx, "document.docx")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Markdown)
}
```

### With Options

```go
result, err := word2md.ConvertFile(
    ctx,
    "document.docx",
    word2md.WithImageDir("./output/images"),
    word2md.WithHeadingDetection(true),
    word2md.WithListDetection(true),
    word2md.WithInlineFormatting(true),
    word2md.WithHyperlinks(true),
)
```

---

## CLI Usage

```bash
# Basic conversion
word2md -o output.md document.docx

# With image extraction
word2md -o output.md --images-dir ./images document.docx

# All options
word2md -o output.md \
    --images-dir ./images \
    --heading-detection \
    --list-detection \
    --inline-formatting \
    --hyperlinks \
    document.docx
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o` | `input.md` | Output file path |
| `--images-dir` | (none) | Directory for extracted images |
| `--inline-formatting` | `true` | Enable bold, italic, code formatting |
| `--hyperlinks` | `true` | Enable hyperlink conversion |
| `--heading-detection` | `true` | Enable heading detection |
| `--list-detection` | `true` | Enable list detection |

---

## Advanced Configuration

### Heading Detection

By default, `word2md` uses a three-tier approach:

1. **Style names** — Checks for "Heading 1", "Heading 2", etc. in Word's style definitions.
2. **Outline levels** — Reads Word's built-in outline level settings.
3. **Font size heuristics** — Falls back to font size thresholds (customizable).

```go
word2md.WithHeadingDetection(true)

// Custom font size thresholds (in half-points: 48 = 24pt)
word2md.WithHeadingThresholds(48, 36, 28) // h1, h2, h3
```

### Custom Renderer

Implement the `renderer.Renderer` interface to output any format:

```go
package main

import (
    "context"
    "fmt"
    "github.com/saiful-anwar/word2md"
    "github.com/saiful-anwar/word2md/pkg/models"
)

type JSONRenderer struct{}

func (r *JSONRenderer) Render(doc *models.Document) (string, error) {
    // Custom JSON serialization logic
    return "", nil
}
```

---

## Supported Elements

| Word Element | Markdown Output | Status |
|---|---|---|
| Paragraph | Plain text | ✅ |
| Heading 1–6 | `#` to `######` | ✅ |
| Bold | `**text**` | ✅ |
| Italic | `*text*` | ✅ |
| Strikethrough | `~~text~~` | ✅ |
| Inline Code | `` `code` `` | ✅ (via font detection) |
| Hyperlinks | `[text](url)` | ✅ |
| Numbered Lists | `1. item` | ✅ |
| Bullet Lists | `- item` | ✅ |
| Tables | `\| a \| b \|` | ✅ |
| Images | `![alt](path)` | ✅ |
| Footnotes | `[^1]` | 🚧 Planned |
| Comments | _(silently ignored)_ | ✅ |
| Page breaks | _(newline)_ | ✅ |

---

## Heading Detection

`word2md` prioritizes accuracy over guesswork. The detection algorithm works in this order:

### 1. Style Name Matching

If Word's styles.xml defines a paragraph style named "Heading 1", "Heading 2", etc., the system uses that directly regardless of font size.

### 2. Outline Levels

If the style or paragraph has an explicit outline level attribute, that takes priority.

### 3. Font Size Heuristics (Fallback)

When no style or outline information exists, `word2md` uses font size analysis:

| Heading | Font Size Threshold (half-points) | Notes |
|---|---|---|
| H1 | >= 48 (24pt) | No length limit |
| H2 | >= 36 (18pt) | No length limit |
| H3 | >= 28 (14pt) | Max 15 chars, relaxed to 45 with numbering prefix |
| H4 | >= 24 (12pt) | Only with numbering prefix |

> **Note:** Font sizes in DOCX are stored as half-points (e.g., `w:val="24"` = 12pt). The thresholds use the same unit system.

---

## Architecture

```
word2md/
├── api.go              # PUBLIC API: Convert(), ConvertFile()
├── options.go           # PUBLIC API: Option functions
├── result.go            # PUBLIC API: Result type
├── cmd/
│   └── word2md/         # CLI (thin wrapper around API)
├── pkg/
│   ├── engine/          # Orchestration pipeline
│   ├── parser/          # DOCX ZIP + XML parsing
│   ├── classifier/      # Heading & list detection
│   ├── renderer/        # Markdown generation
│   ├── extractor/       # Image/media extraction
│   └── models/          # Domain types
├── internal/
│   ├── config/          # Default configurations
│   └── utils/           # Internal utilities

```

### Pipeline

```
DOCX file
    │
    ▼
pkg/parser/ —— ZIP → XML —→ models.Document
    │
    ▼
pkg/classifier/ — heading detection —→ list detection
    │
    ▼
pkg/extractor/ — images → models.Image
    │
    ▼
pkg/renderer/ — MarkdownRenderer —→ string
    │
    ▼
Markdown output
```

---

## API Reference

### `Convert(ctx, source, opts...) (*Result, error)`

The core conversion function. Accepts a `Source` (file path) and optional configuration.

### `ConvertFile(ctx, path, opts...) (*Result, error)`

Convenience wrapper that takes a file path directly.

### `Result`

| Field | Type | Description |
|---|---|---|
| `Markdown` | `string` | The generated Markdown content |
| `ImageFiles` | `[]string` | Paths to extracted image files |
| `Warnings` | `[]error` | Non-fatal issues encountered during conversion |

### Options

| Function | Description |
|---|---|
| `WithImageDir(dir)` | Set directory for extracted images |
| `WithInlineFormatting(bool)` | Enable bold, italic, code, strikethrough |
| `WithHyperlinks(bool)` | Enable hyperlink conversion |
| `WithHeadingDetection(bool)` | Enable heading detection |
| `WithListDetection(bool)` | Enable list detection |
| `WithHeadingThresholds(h1, h2, h3)` | Custom font size thresholds |

---

## Examples

### Basic

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/saiful-anwar/word2md"
)

func main() {
	ctx := context.Background()

	result, err := word2md.ConvertFile(ctx, "document.docx")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Markdown)
}
```

### Custom Renderer

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/saiful-anwar/word2md"
)

func main() {
	ctx := context.Background()

	result, err := word2md.ConvertFile(
		ctx,
		"document.docx",
		word2md.WithImageDir("./output/images"),
		word2md.WithHeadingDetection(true),
		word2md.WithListDetection(true),
		word2md.WithInlineFormatting(true),
		word2md.WithHyperlinks(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Markdown)
	fmt.Printf("Images: %v\n", result.ImageFiles)
	fmt.Printf("Warnings: %v\n", result.Warnings)
}
```

---

## License

MIT License — see [LICENSE](./LICENSE)
