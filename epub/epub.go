package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Epub struct {
	reader     *zip.ReadCloser
	entries    []*zipEntry
	entryIndex map[string]*zipEntry
	opfPath    string
	opfDir     string
	opfDoc     *opfPackage
	idCounter  int
}

type zipEntry struct {
	header       zip.FileHeader
	originalFile *zip.File
	modifiedData []byte
	isDir        bool
	removed      bool
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
	defer rc.Close()
	return io.ReadAll(rc)
}

func Open(inputPath string) (*Epub, error) {
	reader, err := zip.OpenReader(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	p := &Epub{reader: reader, entryIndex: make(map[string]*zipEntry)}
	for _, f := range reader.File {
		normName := normalizeZipPath(f.Name)
		entry := &zipEntry{header: f.FileHeader, originalFile: f, isDir: f.FileInfo().IsDir()}
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

func (p *Epub) Save(outputPath string) error {
	if err := p.flushOPF(); err != nil {
		return err
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	writer := zip.NewWriter(outFile)
	defer writer.Close()
	for _, entry := range p.entries {
		if entry.removed {
			continue
		}
		if entry.isDir {
			if err := writeDirEntry(writer, &entry.header); err != nil {
				return err
			}
			continue
		}
		w, err := writer.CreateHeader(&entry.header)
		if err != nil {
			return err
		}
		if entry.modifiedData != nil {
			if _, err := w.Write(entry.modifiedData); err != nil {
				return err
			}
		} else if entry.originalFile != nil {
			rc, err := entry.originalFile.Open()
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, rc); err != nil {
				_ = rc.Close()
				return err
			}
			_ = rc.Close()
		}
	}
	return nil
}

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
	entry.modifiedData = serialized
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
	_ = encoder.Flush()
	return buf.Bytes(), nil
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
