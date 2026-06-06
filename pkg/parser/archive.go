package parser

import (
	"archive/zip"
	"fmt"
	"io"
)

// Archive provides access to a DOCX ZIP file.
type Archive struct {
	reader *zip.ReadCloser
	files  map[string]*zip.File
}

// OpenArchive opens a DOCX file as a ZIP archive.
func OpenArchive(path string) (*Archive, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open archive %s: %w", path, err)
	}

	files := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		files[f.Name] = f
	}

	return &Archive{reader: r, files: files}, nil
}

// ReadFile returns the contents of a file inside the archive.
func (a *Archive) ReadFile(name string) ([]byte, error) {
	f, ok := a.files[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found in archive", name)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", name, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", name, err)
	}

	return data, nil
}

// FileExists checks if a file exists in the archive.
func (a *Archive) FileExists(name string) bool {
	_, ok := a.files[name]
	return ok
}

// Close closes the archive.
func (a *Archive) Close() error {
	if a.reader != nil {
		return a.reader.Close()
	}
	return nil
}

// File returns a zip.File for a given name (for streaming extraction).
func (a *Archive) File(name string) (*zip.File, bool) {
	f, ok := a.files[name]
	return f, ok
}
