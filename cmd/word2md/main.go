package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/saiful-anwar/word2md"
)

func main() {
	var (
		outputFile       = flag.String("o", "", "Output markdown file path")
		imagesDir        = flag.String("images-dir", "", "Directory for extracted images")
		inlineFormatting = flag.Bool("inline-formatting", true, "Enable bold, italic, code formatting")
		hyperlinks       = flag.Bool("hyperlinks", true, "Enable hyperlink conversion")
		headingDetection = flag.Bool("heading-detection", true, "Enable heading detection")
		listDetection    = flag.Bool("list-detection", true, "Enable list detection")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input.docx>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Convert Microsoft Word (.docx) documents to Markdown.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -o output.md --images-dir ./images document.docx\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)
	outputPath := *outputFile
	if outputPath == "" {
		outputPath = inputPath + ".md"
	}

	ctx := context.Background()

	opts := []word2md.Option{
		word2md.WithImageDir(*imagesDir),
		word2md.WithInlineFormatting(*inlineFormatting),
		word2md.WithHyperlinks(*hyperlinks),
		word2md.WithHeadingDetection(*headingDetection),
		word2md.WithListDetection(*listDetection),
	}

	result, err := word2md.ConvertFile(ctx, inputPath, opts...)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	if err := os.WriteFile(outputPath, []byte(result.Markdown), 0644); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	fmt.Printf("Successfully converted to %s\n", outputPath)
	if len(result.ImageFiles) > 0 {
		fmt.Printf("Extracted %d image(s)\n", len(result.ImageFiles))
	}
	if result.HasWarnings() {
		fmt.Printf("Warnings (%d):\n", result.WarningCount())
		for _, w := range result.Warnings {
			fmt.Printf("  - %v\n", w)
		}
	}
}
