package epub

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const minimalOPF = `<?xml version="1.0" encoding="UTF-8"?><package><metadata></metadata><manifest><item id="item1" href="page.html" media-type="application/xhtml+xml"/></manifest><spine><itemref idref="item1"/></spine></package>`

func TestEpubOpenApplyAndSave(t *testing.T) {
	inputPath := writeTestEpub(t, map[string]string{
		"content.opf": minimalOPF,
		"page.html":   "<html><body>Hello World</body></html>",
	})

	book, err := Open(inputPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer book.Close()

	modified, err := book.ApplyHTML(func(name, html string) (string, error) {
		return "<html><body>Hello Go</body></html>", nil
	})
	if err != nil {
		t.Fatalf("ApplyHTML() error = %v", err)
	}
	if modified != 1 {
		t.Fatalf("ApplyHTML() modified %d entries, want 1", modified)
	}

	outputPath := filepath.Join(t.TempDir(), "saved.epub")
	if err := book.Save(outputPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	saved, err := Open(outputPath)
	if err != nil {
		t.Fatalf("saved EPUB cannot be reopened: %v", err)
	}
	defer saved.Close()
}

func TestOpenWithOptionsRejectsOversizedEntry(t *testing.T) {
	path := writeTestEpub(t, map[string]string{
		"content.opf": minimalOPF,
		"page.html":   "oversized",
	})

	_, err := OpenWithOptions(path, Options{MaxEntryBytes: 4, MaxArchiveBytes: 1024})
	if !errors.Is(err, ErrEntryTooLarge) {
		t.Fatalf("OpenWithOptions() error = %v, want ErrEntryTooLarge", err)
	}
}

func TestOpenWithOptionsRejectsOversizedArchive(t *testing.T) {
	path := writeTestEpub(t, map[string]string{
		"content.opf": minimalOPF,
		"page.html":   "page",
	})

	_, err := OpenWithOptions(path, Options{MaxEntryBytes: 1024, MaxArchiveBytes: 4})
	if !errors.Is(err, ErrArchiveTooLarge) {
		t.Fatalf("OpenWithOptions() error = %v, want ErrArchiveTooLarge", err)
	}
}

func TestApplyHTMLRejectsArchiveGrowthBeyondLimit(t *testing.T) {
	path := writeTestEpub(t, map[string]string{
		"content.opf": minimalOPF,
		"page.html":   strings.Repeat("x", 100),
	})

	book, err := OpenWithOptions(path, Options{MaxEntryBytes: 1024, MaxArchiveBytes: 512})
	if err != nil {
		t.Fatalf("OpenWithOptions() error = %v", err)
	}
	defer book.Close()

	_, err = book.ApplyHTML(func(name, html string) (string, error) {
		return strings.Repeat("x", 400), nil
	})
	if !errors.Is(err, ErrArchiveTooLarge) {
		t.Fatalf("ApplyHTML() error = %v, want ErrArchiveTooLarge", err)
	}
}

func TestCopyEntryLimitedRejectsEntryAndArchiveOverflow(t *testing.T) {
	for _, tc := range []struct {
		name      string
		entryMax  int64
		remaining int64
		wantErr   error
	}{
		{name: "entry", entryMax: 3, remaining: -1, wantErr: ErrEntryTooLarge},
		{name: "archive", entryMax: 10, remaining: 3, wantErr: ErrArchiveTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dst bytes.Buffer
			_, err := copyEntryLimited(&dst, strings.NewReader("oversized"), tc.entryMax, tc.remaining)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("copyEntryLimited() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func writeTestEpub(t *testing.T, entries map[string]string) string {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, value := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := entry.Write([]byte(value)); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("zip Close() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "test.epub")
	if err := os.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
