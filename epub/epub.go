package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	// DefaultMaxEntryBytes limits the expanded size of a single EPUB entry.
	DefaultMaxEntryBytes int64 = 64 << 20
	// DefaultMaxArchiveBytes limits the total declared expanded EPUB size.
	DefaultMaxArchiveBytes int64 = 512 << 20
)

var (
	// ErrEntryTooLarge is returned when one EPUB entry exceeds its configured limit.
	ErrEntryTooLarge = errors.New("epub: entry exceeds configured size limit")
	// ErrArchiveTooLarge is returned when the declared total EPUB size exceeds its limit.
	ErrArchiveTooLarge = errors.New("epub: archive exceeds configured size limit")
)

// Options controls resource limits for opening and processing EPUB files.
// A zero value uses the safe defaults; a negative limit explicitly disables it.
type Options struct {
	MaxEntryBytes   int64
	MaxArchiveBytes int64
}

func (opts Options) withDefaults() Options {
	if opts.MaxEntryBytes == 0 {
		opts.MaxEntryBytes = DefaultMaxEntryBytes
	}
	if opts.MaxArchiveBytes == 0 {
		opts.MaxArchiveBytes = DefaultMaxArchiveBytes
	}
	return opts
}

// Epub 表示一个 EPUB 电子书文档。
type Epub struct {
	reader          *zip.ReadCloser
	entries         []*zipEntry
	entryIndex      map[string]*zipEntry
	opfPath         string
	opfDir          string
	opfDoc          *opfPackage
	idCounter       int
	maxArchiveBytes int64
}

type zipEntry struct {
	header       zip.FileHeader
	originalFile *zip.File
	modifiedData []byte
	isDir        bool
	removed      bool
	maxBytes     int64
}

func (e *zipEntry) GetData() ([]byte, error) {
	if e.removed {
		return nil, fmt.Errorf("entry removed")
	}
	if e.modifiedData != nil {
		return e.modifiedData, nil
	}
	if e.originalFile == nil {
		return nil, nil
	}
	rc, err := e.originalFile.Open()
	if err != nil {
		return nil, err
	}
	data, readErr := readAllLimited(rc, e.maxBytes)
	closeErr := rc.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func Open(inputPath string) (*Epub, error) {
	return OpenWithOptions(inputPath, Options{})
}

// OpenWithOptions opens an EPUB while enforcing the supplied resource limits.
func OpenWithOptions(inputPath string, opts Options) (*Epub, error) {
	opts = opts.withDefaults()
	reader, err := zip.OpenReader(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	p := &Epub{
		reader:          reader,
		entryIndex:      make(map[string]*zipEntry),
		maxArchiveBytes: opts.MaxArchiveBytes,
	}
	var archiveBytes uint64
	for _, f := range reader.File {
		if opts.MaxEntryBytes > 0 && f.UncompressedSize64 > uint64(opts.MaxEntryBytes) {
			_ = reader.Close()
			return nil, fmt.Errorf("%w: %s", ErrEntryTooLarge, f.Name)
		}
		if ^uint64(0)-archiveBytes < f.UncompressedSize64 {
			_ = reader.Close()
			return nil, ErrArchiveTooLarge
		}
		archiveBytes += f.UncompressedSize64
		if opts.MaxArchiveBytes > 0 && archiveBytes > uint64(opts.MaxArchiveBytes) {
			_ = reader.Close()
			return nil, ErrArchiveTooLarge
		}
		normName := normalizeZipPath(f.Name)
		entry := &zipEntry{
			header:       f.FileHeader,
			originalFile: f,
			isDir:        f.FileInfo().IsDir(),
			maxBytes:     opts.MaxEntryBytes,
		}
		if !entry.isDir && isOPFFile(f.Name) {
			p.opfPath = normName
			p.opfDir = normalizeZipPath(path.Dir(normName))
			if p.opfDir == "." {
				p.opfDir = ""
			}
			data, err := entry.GetData()
			if err != nil {
				_ = reader.Close()
				return nil, err
			}
			doc := &opfPackage{}
			if err := xml.Unmarshal(data, doc); err != nil {
				_ = reader.Close()
				return nil, err
			}
			p.opfDoc = doc
			p.idCounter = len(doc.Manifest.Items)
		}
		p.entries = append(p.entries, entry)
		p.entryIndex[normName] = entry
	}
	if p.opfDoc == nil {
		_ = reader.Close()
		return nil, fmt.Errorf("content.opf not found")
	}
	return p, nil
}

func (p *Epub) Close() error {
	if p.reader != nil {
		return p.reader.Close()
	}
	return nil
}

// Save 将修改后的内容保存到指定的 EPUB 文件。
func (p *Epub) Save(outputPath string) error {
	if err := p.flushOPF(); err != nil {
		return err
	}
	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	outFile, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := outFile.Name()
	committed := false
	closed := false
	defer func() {
		if !closed {
			_ = outFile.Close()
		}
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()
	writer := zip.NewWriter(outFile)
	var archiveBytes int64
	for _, entry := range p.entries {
		if entry.removed {
			continue
		}
		if entry.isDir {
			if err := writeDirEntry(writer, &entry.header); err != nil {
				_ = writer.Close()
				return err
			}
			continue
		}
		w, err := writer.CreateHeader(&entry.header)
		if err != nil {
			_ = writer.Close()
			return err
		}
		if entry.modifiedData != nil {
			if entry.maxBytes > 0 && int64(len(entry.modifiedData)) > entry.maxBytes {
				_ = writer.Close()
				return fmt.Errorf("%w: %s", ErrEntryTooLarge, entry.header.Name)
			}
			if err := p.addArchiveBytes(&archiveBytes, int64(len(entry.modifiedData))); err != nil {
				_ = writer.Close()
				return err
			}
			if _, err := w.Write(entry.modifiedData); err != nil {
				_ = writer.Close()
				return err
			}
		} else if entry.originalFile != nil {
			rc, err := entry.originalFile.Open()
			if err != nil {
				_ = writer.Close()
				return err
			}
			written, err := copyEntryLimited(w, rc, entry.maxBytes, p.archiveBytesRemaining(archiveBytes))
			if err != nil {
				_ = rc.Close()
				_ = writer.Close()
				return err
			}
			if err := rc.Close(); err != nil {
				_ = writer.Close()
				return err
			}
			if err := p.addArchiveBytes(&archiveBytes, written); err != nil {
				_ = writer.Close()
				return err
			}
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if err := outFile.Close(); err != nil {
		return err
	}
	closed = true
	if err := os.Rename(tempPath, outputPath); err != nil {
		return err
	}
	committed = true
	return nil
}

// ApplyHTML 对 EPUB 中的所有 HTML 页面应用回调函数进行批量修改。
func (p *Epub) ApplyHTML(fn func(name string, html string) (string, error)) (int, error) {
	if fn == nil {
		return 0, nil
	}
	modified := 0
	for _, entry := range p.entryIndex {
		if entry.removed || !isHTMLEntry(entry) {
			continue
		}
		data, err := entry.GetData()
		if err != nil {
			return 0, err
		}
		original := string(data)
		updated, err := fn(entry.header.Name, original)
		if err != nil {
			return 0, err
		}
		if entry.maxBytes > 0 && int64(len(updated)) > entry.maxBytes {
			return 0, fmt.Errorf("%w: %s", ErrEntryTooLarge, entry.header.Name)
		}
		if err := p.checkArchiveSize(entry, int64(len(updated))); err != nil {
			return 0, err
		}
		if updated != original {
			entry.modifiedData = []byte(updated)
			modified++
		}
	}
	return modified, nil
}

func (p *Epub) FindHTMLByText(text string) ([]string, error) {
	var matches []string
	for _, entry := range p.entryIndex {
		if entry.removed || !isHTMLEntry(entry) {
			continue
		}
		data, err := entry.GetData()
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(data), text) {
			matches = append(matches, entry.header.Name)
		}
	}
	return matches, nil
}

func (p *Epub) removeEntry(norm string) error {
	entry, ok := p.entryIndex[norm]
	if !ok {
		return fmt.Errorf("not found: %s", norm)
	}
	entry.removed = true
	if p.opfDoc != nil {
		href, _ := p.hrefForOPF(norm)
		p.removeFromOPF(href)
	}
	return nil
}

func (p *Epub) flushOPF() error {
	if p.opfDoc == nil {
		return nil
	}
	serialized, err := serializeOPF(p.opfDoc)
	if err != nil {
		return err
	}
	entry := p.entryIndex[p.opfPath]
	previous := entry.modifiedData
	entry.modifiedData = serialized
	if err := p.checkArchiveSize(nil, 0); err != nil {
		entry.modifiedData = previous
		return err
	}
	return nil
}

func (p *Epub) hrefForOPF(norm string) (string, error) {
	if p.opfDir == "" {
		return norm, nil
	}
	if !strings.HasPrefix(norm, p.opfDir+"/") {
		return "", fmt.Errorf("file not under OPF dir")
	}
	return strings.TrimPrefix(norm[len(p.opfDir)+1:], ""), nil
}

func (p *Epub) removeFromOPF(href string) {
	removedIDs := map[string]struct{}{}
	items := make([]opfManifestItem, 0, len(p.opfDoc.Manifest.Items))
	for _, item := range p.opfDoc.Manifest.Items {
		if normalizeZipPath(item.Href) == normalizeZipPath(href) {
			removedIDs[item.ID] = struct{}{}
			continue
		}
		items = append(items, item)
	}
	p.opfDoc.Manifest.Items = items
	spine := make([]opfSpineItem, 0, len(p.opfDoc.Spine.Items))
	for _, item := range p.opfDoc.Spine.Items {
		if _, ok := removedIDs[item.IDRef]; !ok {
			spine = append(spine, item)
		}
	}
	p.opfDoc.Spine.Items = spine
}

// 辅助函数
func normalizeZipPath(pth string) string { return path.Clean(strings.ReplaceAll(pth, "\\", "/")) }
func isHTMLEntry(entry *zipEntry) bool {
	n := strings.ToLower(entry.header.Name)
	return strings.HasSuffix(n, ".html") || strings.HasSuffix(n, ".xhtml") || strings.HasSuffix(n, ".htm")
}
func isOPFFile(name string) bool { return strings.HasSuffix(strings.ToLower(name), ".opf") }

func writeDirEntry(writer *zip.Writer, header *zip.FileHeader) error {
	h := *header
	h.Method = zip.Store
	if !strings.HasSuffix(h.Name, "/") {
		h.Name += "/"
	}
	_, err := writer.CreateHeader(&h)
	return err
}

func serializeOPF(doc *opfPackage) ([]byte, error) {
	buf := bytes.NewBufferString(xml.Header)
	encoder := xml.NewEncoder(buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readAllLimited(reader io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 || maxBytes == math.MaxInt64 {
		return io.ReadAll(reader)
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrEntryTooLarge
	}
	return data, nil
}

func (p *Epub) archiveBytesRemaining(used int64) int64 {
	if p.maxArchiveBytes <= 0 {
		return -1
	}
	return p.maxArchiveBytes - used
}

func (p *Epub) addArchiveBytes(used *int64, count int64) error {
	if count < 0 || *used > math.MaxInt64-count {
		return ErrArchiveTooLarge
	}
	*used += count
	if p.maxArchiveBytes > 0 && *used > p.maxArchiveBytes {
		return ErrArchiveTooLarge
	}
	return nil
}

func (p *Epub) checkArchiveSize(replacement *zipEntry, replacementSize int64) error {
	if p.maxArchiveBytes <= 0 {
		return nil
	}
	var total int64
	for _, entry := range p.entries {
		if entry.removed {
			continue
		}

		size := int64(0)
		switch {
		case entry == replacement:
			size = replacementSize
		case entry.modifiedData != nil:
			size = int64(len(entry.modifiedData))
		case entry.originalFile != nil:
			if entry.originalFile.UncompressedSize64 > math.MaxInt64 {
				return ErrArchiveTooLarge
			}
			size = int64(entry.originalFile.UncompressedSize64)
		}
		if err := p.addArchiveBytes(&total, size); err != nil {
			return err
		}
	}
	return nil
}

func copyEntryLimited(dst io.Writer, src io.Reader, entryMaxBytes, archiveBytesRemaining int64) (int64, error) {
	limit := entryMaxBytes
	archiveIsLimit := false
	if archiveBytesRemaining >= 0 && (limit <= 0 || archiveBytesRemaining < limit) {
		limit = archiveBytesRemaining
		archiveIsLimit = true
	}
	if limit < 0 || limit == math.MaxInt64 {
		return io.Copy(dst, src)
	}
	if limit == 0 {
		written, err := io.Copy(dst, io.LimitReader(src, 1))
		if err != nil {
			return written, err
		}
		if written > 0 {
			return written, ErrArchiveTooLarge
		}
		return written, nil
	}

	written, err := io.Copy(dst, io.LimitReader(src, limit+1))
	if err != nil {
		return written, err
	}
	if written > limit {
		if archiveIsLimit {
			return written, ErrArchiveTooLarge
		}
		return written, ErrEntryTooLarge
	}
	return written, nil
}

// 结构体定义
type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
	Manifest opfManifest `xml:"manifest"`
	Spine    opfSpine    `xml:"spine"`
}
type opfMetadata struct {
	InnerXML []byte `xml:",innerxml"`
}
type opfManifest struct {
	Items []opfManifestItem `xml:"item"`
}
type opfManifestItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}
type opfSpine struct {
	Items []opfSpineItem `xml:"itemref"`
}
type opfSpineItem struct {
	IDRef string `xml:"idref,attr"`
}
