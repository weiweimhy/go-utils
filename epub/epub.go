package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

// Epub 封装 EPUB 解压、修改与重新打包的能力
type Epub struct {
	entries    []*zipEntry
	entryIndex map[string]*zipEntry

	opfPath string
	opfDir  string
	opfDoc  *opfPackage

	idCounter int
}

type zipEntry struct {
	header  zip.FileHeader
	data    []byte
	isDir   bool
	removed bool
}

// ProcessOptions 用于组合常见的 EPUB 处理操作
type ProcessOptions struct {
	InputPath          string
	OutputPath         string
	RemoveHTMLKeywords []string
	ReplaceHTML        func(name string, html string) (string, error)
	Customize          func(p *Epub) error
}

// Open 从 EPUB 文件构建 Epub，所有数据会被读取到内存中
func Open(inputPath string) (*Epub, error) {
	reader, err := zip.OpenReader(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer func(reader *zip.ReadCloser) {
		_ = reader.Close()
	}(reader)

	p := &Epub{
		entryIndex: make(map[string]*zipEntry),
	}

	for _, f := range reader.File {
		header := f.FileHeader
		entry := &zipEntry{
			header: header,
			isDir:  f.FileInfo().IsDir(),
		}

		normName := normalizeZipPath(header.Name)
		if !entry.isDir {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to read zip entry (%s): %w", header.Name, err)
			}

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read file content (%s): %w", header.Name, err)
			}

			err = rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to close file (%s): %w", header.Name, err)
			}

			entry.data = data

			if isOPFFile(header.Name) {
				p.opfPath = normName
				p.opfDir = normalizeZipPath(path.Dir(normName))
				if p.opfDir == "." {
					p.opfDir = ""
				}

				doc := &opfPackage{}
				if err := xml.Unmarshal(data, doc); err != nil {
					return nil, fmt.Errorf("failed to parse content.opf (%s): %w", header.Name, err)
				}
				p.opfDoc = doc
				p.idCounter = len(doc.Manifest.Items)
			}
		}

		p.entries = append(p.entries, entry)
		p.entryIndex[normName] = entry
	}

	if p.opfDoc == nil {
		return nil, fmt.Errorf("content.opf not found")
	}

	return p, nil
}

// Save 将当前状态写入新的 EPUB 文件
func (p *Epub) Save(outputPath string) error {
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	if err := p.flushOPF(); err != nil {
		return err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	writer := zip.NewWriter(outFile)
	defer func() { _ = writer.Close() }()

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

		if err := writeFileEntry(writer, &entry.header, entry.data); err != nil {
			return err
		}
	}

	return nil
}

// FindHTMLByText 搜索包含指定文案的 HTML 章节，返回其路径
func (p *Epub) FindHTMLByText(text string) ([]string, error) {
	if text == "" {
		return nil, fmt.Errorf("search text cannot be empty")
	}
	var matches []string
	for _, entry := range p.entryIndex {
		if entry.removed || !isHTMLEntry(entry) {
			continue
		}
		if strings.Contains(string(entry.data), text) {
			matches = append(matches, entry.header.Name)
		}
	}
	return matches, nil
}

// ReplaceAllHTML 统一替换所有 HTML 中的文案，返回替换次数
func (p *Epub) ReplaceAllHTML(oldText, newText string) (int, error) {
	if oldText == "" {
		return 0, fmt.Errorf("old text cannot be empty")
	}
	count := 0
	for _, entry := range p.entryIndex {
		if entry.removed || !isHTMLEntry(entry) {
			continue
		}
		html := string(entry.data)
		replaced := strings.Count(html, oldText)
		if replaced == 0 {
			continue
		}
		html = strings.ReplaceAll(html, oldText, newText)
		entry.data = []byte(html)
		count += replaced
	}
	return count, nil
}

func (p *Epub) CountHTML() int {
	totalHTML := 0
	for _, entry := range p.entryIndex {
		if !entry.removed && isHTMLEntry(entry) {
			totalHTML++
		}
	}
	return totalHTML
}

// ApplyHTML 对所有 HTML 执行自定义函数，返回被修改的章节数
func (p *Epub) ApplyHTML(fn func(name string, html string) (string, error)) (int, error) {
	if fn == nil {
		return 0, nil
	}

	modified := 0
	currentIndex := 0
	for _, entry := range p.entryIndex {
		if entry.removed || !isHTMLEntry(entry) {
			continue
		}
		original := string(entry.data)
		// 计算进度：当前处理的 HTML 文件索引 / 总 HTML 文件数
		updated, err := fn(entry.header.Name, original)
		if err != nil {
			return 0, fmt.Errorf("failed to process HTML (%s): %w", entry.header.Name, err)
		}
		if updated != original {
			entry.data = []byte(updated)
			modified++
		}
		currentIndex++
	}
	return modified, nil
}

// RemoveHTMLContaining 删除所有包含关键词的 HTML 章节，返回删除的路径
func (p *Epub) RemoveHTMLContaining(keywords []string) ([]string, error) {
	if len(keywords) == 0 {
		return nil, nil
	}
	var removed []string
	for norm, entry := range p.entryIndex {
		if entry.removed || !isHTMLEntry(entry) {
			continue
		}
		if shouldRemoveHTML(string(entry.data), keywords) {
			if err := p.removeEntry(norm); err != nil {
				return nil, err
			}
			removed = append(removed, entry.header.Name)
		}
	}
	return removed, nil
}

// AddChapter 新增章节
// filePath 为 ZIP 内路径（相对于 EPUB 根目录），spineIndex 为插入到 OPF spine 的位置（-1 表示追加）
func (p *Epub) AddChapter(filePath, html string, spineIndex int) error {
	if filePath == "" {
		return fmt.Errorf("chapter path cannot be empty")
	}
	norm := normalizeZipPath(filePath)
	if _, ok := p.entryIndex[norm]; ok {
		return fmt.Errorf("file already exists: %s", filePath)
	}

	if err := p.ensureDirectories(norm); err != nil {
		return err
	}

	header := zip.FileHeader{
		Name:   norm,
		Method: zip.Deflate,
	}

	entry := &zipEntry{
		header: header,
		data:   []byte(html),
	}
	p.entries = append(p.entries, entry)
	p.entryIndex[norm] = entry

	if err := p.addToOPF(norm, spineIndex); err != nil {
		return err
	}

	return nil
}

// AddChapterFromFile 从本地文件系统路径添加章节
// epubChapterPath 是 EPUB 内的章节路径（相对于 EPUB 根目录）
// htmlFilePath 是本地 HTML 文件路径
// spineIndex 为插入到 OPF spine 的位置（-1 表示追加）
func (p *Epub) AddChapterFromFile(epubChapterPath, htmlFilePath string, spineIndex int) error {
	if epubChapterPath == "" {
		return fmt.Errorf("EPUB chapter path cannot be empty")
	}
	if htmlFilePath == "" {
		return fmt.Errorf("HTML file path cannot be empty")
	}

	// 读取 HTML 文件
	htmlData, err := os.ReadFile(htmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %w", err)
	}
	htmlContent := string(htmlData)

	// 获取 HTML 文件所在目录（用于解析相对路径的图片）
	htmlDir := path.Dir(htmlFilePath)

	// 提取图片路径
	imagePaths := p.extractImagePathsFromHTML(htmlContent, htmlDir)

	// 获取 EPUB 章节所在目录
	epubChapterDir := normalizeZipPath(path.Dir(epubChapterPath))
	if epubChapterDir == "." {
		epubChapterDir = ""
	}

	// 确定图片存放目录：优先放在 OPF 目录下的 static_images 子目录
	var imageBaseDir string
	if p.opfDir != "" {
		imageBaseDir = p.opfDir + "/static_images"
	} else {
		if epubChapterDir == "" {
			imageBaseDir = "static_images"
		} else {
			imageBaseDir = epubChapterDir + "/static_images"
		}
	}
	imageBaseDir = normalizeZipPath(imageBaseDir)

	// 添加图片到 EPUB 并更新 HTML 中的路径
	imageMap := make(map[string]string) // 原路径 -> EPUB 内路径
	for _, imgPath := range imagePaths {
		// 构建 EPUB 内的图片路径
		imgName := path.Base(imgPath)
		epubImgPath := normalizeZipPath(imageBaseDir + "/" + imgName)

		// 检查图片是否已存在
		if _, exists := p.entryIndex[epubImgPath]; exists {
			// 图片已存在，使用现有路径
			imageMap[imgPath] = epubImgPath
			continue
		}

		// 读取图片文件
		imgData, err := os.ReadFile(imgPath)
		if err != nil {
			// 如果图片文件不存在，跳过（可能是相对路径解析问题）
			continue
		}

		// 添加图片到 EPUB
		if err := p.addImageFile(epubImgPath, imgData); err != nil {
			return fmt.Errorf("failed to add image (%s): %w", imgPath, err)
		}

		imageMap[imgPath] = epubImgPath
	}

	// 更新 HTML 中的图片路径
	updatedHTML := p.updateImagePathsInHTML(htmlContent, imageMap, htmlDir, epubChapterDir, imageBaseDir)

	// 添加 HTML 章节
	return p.AddChapter(epubChapterPath, updatedHTML, spineIndex)
}

// RemoveFileByName 删除指定文件，并更新 content.opf
func (p *Epub) RemoveFileByName(filePath string) error {
	norm := normalizeZipPath(filePath)
	return p.removeEntry(norm)
}

// SaveAs 调用 Save 的别名
func (p *Epub) SaveAs(outputPath string) error {
	return p.Save(outputPath)
}

// ---------- 内部工具 ----------

func (p *Epub) removeEntry(norm string) error {
	entry, ok := p.entryIndex[norm]
	if !ok {
		return fmt.Errorf("file does not exist: %s", norm)
	}
	if entry.removed {
		return nil
	}
	entry.removed = true

	if p.opfDoc != nil {
		href, err := p.hrefForOPF(norm)
		if err != nil {
			return err
		}
		p.removeFromOPF(href)
	}
	return nil
}

func (p *Epub) ensureDirectories(norm string) error {
	dir := normalizeZipPath(path.Dir(norm))
	if dir == "." || dir == "" {
		return nil
	}

	parts := strings.Split(dir, "/")
	curr := ""
	for _, part := range parts {
		if curr == "" {
			curr = part
		} else {
			curr = curr + "/" + part
		}
		if _, ok := p.entryIndex[curr]; ok {
			continue
		}
		header := zip.FileHeader{
			Name:   curr + "/",
			Method: zip.Store,
		}
		entry := &zipEntry{
			header: header,
			isDir:  true,
		}
		p.entries = append(p.entries, entry)
		p.entryIndex[curr] = entry
	}
	return nil
}

func (p *Epub) addToOPF(norm string, spineIndex int) error {
	if p.opfDoc == nil {
		return fmt.Errorf("content.opf not loaded")
	}
	href, err := p.hrefForOPF(norm)
	if err != nil {
		return err
	}

	id := p.generateID(path.Base(norm))
	item := opfManifestItem{
		ID:        id,
		Href:      href,
		MediaType: "application/xhtml+xml",
	}
	p.opfDoc.Manifest.Items = append(p.opfDoc.Manifest.Items, item)

	itemRef := opfSpineItem{IDRef: id}
	if spineIndex < 0 || spineIndex >= len(p.opfDoc.Spine.Items) {
		p.opfDoc.Spine.Items = append(p.opfDoc.Spine.Items, itemRef)
	} else {
		items := append([]opfSpineItem{}, p.opfDoc.Spine.Items[:spineIndex]...)
		items = append(items, itemRef)
		items = append(items, p.opfDoc.Spine.Items[spineIndex:]...)
		p.opfDoc.Spine.Items = items
	}
	return nil
}

func (p *Epub) removeFromOPF(href string) {
	if p.opfDoc == nil {
		return
	}

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
		if _, ok := removedIDs[item.IDRef]; ok {
			continue
		}
		spine = append(spine, item)
	}
	p.opfDoc.Spine.Items = spine
}

func (p *Epub) flushOPF() error {
	if p.opfDoc == nil {
		return nil
	}
	serialized, err := serializeOPF(p.opfDoc)
	if err != nil {
		return err
	}
	entry, ok := p.entryIndex[p.opfPath]
	if !ok {
		return fmt.Errorf("content.opf not found in entries")
	}
	entry.data = serialized
	return nil
}

func (p *Epub) hrefForOPF(norm string) (string, error) {
	if p.opfDir == "" {
		return norm, nil
	}
	if !strings.HasPrefix(norm, p.opfDir+"/") {
		return "", fmt.Errorf("file %s is not under OPF directory %s", norm, p.opfDir)
	}
	return strings.TrimPrefix(norm[len(p.opfDir)+1:], ""), nil
}

func (p *Epub) generateID(base string) string {
	base = sanitizeID(base)
	p.idCounter++
	return fmt.Sprintf("item-%s-%d", base, p.idCounter)
}

func sanitizeID(base string) string {
	if base == "" {
		return "chapter"
	}
	base = strings.ToLower(base)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	base = reg.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		return "chapter"
	}
	return base
}

// extractImagePathsFromHTML 从 HTML 内容中提取所有图片路径（基于本地文件系统）
func (p *Epub) extractImagePathsFromHTML(htmlContent, htmlDir string) []string {
	var imagePaths []string
	seen := make(map[string]bool)

	// 匹配 <img> 标签的 src 属性
	imgRegex := regexp.MustCompile(`<img[^>]+src\s*=\s*["']?([^"'\s>]+)["']?[^>]*>`)
	matches := imgRegex.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		src := match[1]

		// 跳过 data URI 和外部 URL
		if strings.HasPrefix(src, "data:") || strings.HasPrefix(src, "http://") ||
			strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "//") {
			continue
		}

		// 处理相对路径和绝对路径
		var fullPath string
		if path.IsAbs(src) {
			fullPath = src
		} else {
			fullPath = path.Join(htmlDir, src)
		}
		fullPath = path.Clean(fullPath)

		if !seen[fullPath] {
			seen[fullPath] = true
			imagePaths = append(imagePaths, fullPath)
		}
	}

	// 也匹配 CSS 中的 url() 引用
	urlRegex := regexp.MustCompile(`url\s*\(\s*["']?([^"')]+)["']?\s*\)`)
	urlMatches := urlRegex.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range urlMatches {
		if len(match) < 2 {
			continue
		}
		urlPath := match[1]

		// 跳过 data URI 和外部 URL
		if strings.HasPrefix(urlPath, "data:") || strings.HasPrefix(urlPath, "http://") ||
			strings.HasPrefix(urlPath, "https://") || strings.HasPrefix(urlPath, "//") {
			continue
		}

		// 处理相对路径和绝对路径
		var fullPath string
		if path.IsAbs(urlPath) {
			fullPath = urlPath
		} else {
			fullPath = path.Join(htmlDir, urlPath)
		}
		fullPath = path.Clean(fullPath)

		// 只处理图片文件
		ext := strings.ToLower(path.Ext(fullPath))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" ||
			ext == ".svg" || ext == ".webp" || ext == ".bmp" {
			if !seen[fullPath] {
				seen[fullPath] = true
				imagePaths = append(imagePaths, fullPath)
			}
		}
	}

	return imagePaths
}

// addImageFile 添加图片文件到 EPUB 并更新 manifest
func (p *Epub) addImageFile(epubImgPath string, imgData []byte) error {
	norm := normalizeZipPath(epubImgPath)
	if _, ok := p.entryIndex[norm]; ok {
		return nil // 已存在，跳过
	}

	// 确保目录存在
	if err := p.ensureDirectories(norm); err != nil {
		return err
	}

	// 创建 zip entry
	header := zip.FileHeader{
		Name:   norm,
		Method: zip.Deflate,
	}

	entry := &zipEntry{
		header: header,
		data:   imgData,
	}
	p.entries = append(p.entries, entry)
	p.entryIndex[norm] = entry

	// 添加到 manifest
	if err := p.addImageToManifest(norm); err != nil {
		return err
	}

	return nil
}

// addImageToManifest 将图片添加到 content.opf 的 manifest
func (p *Epub) addImageToManifest(norm string) error {
	if p.opfDoc == nil {
		return fmt.Errorf("content.opf not loaded")
	}

	href, err := p.hrefForOPF(norm)
	if err != nil {
		return err
	}

	// 确定媒体类型
	ext := strings.ToLower(path.Ext(norm))
	mediaType := "image/jpeg" // 默认
	switch ext {
	case ".png":
		mediaType = "image/png"
	case ".gif":
		mediaType = "image/gif"
	case ".svg":
		mediaType = "image/svg+xml"
	case ".webp":
		mediaType = "image/webp"
	case ".bmp":
		mediaType = "image/bmp"
	case ".jpg", ".jpeg":
		mediaType = "image/jpeg"
	}

	id := p.generateID(path.Base(norm))
	item := opfManifestItem{
		ID:        id,
		Href:      href,
		MediaType: mediaType,
	}
	p.opfDoc.Manifest.Items = append(p.opfDoc.Manifest.Items, item)

	return nil
}

// updateImagePathsInHTML 更新 HTML 中的图片路径
func (p *Epub) updateImagePathsInHTML(htmlContent string, imageMap map[string]string, htmlDir, epubChapterDir, imageBaseDir string) string {
	// 更新 <img> 标签的 src 属性
	imgRegex := regexp.MustCompile(`(<img[^>]+src\s*=\s*["'])([^"']+)(["'][^>]*>)`)
	htmlContent = imgRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
		submatches := imgRegex.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}
		prefix := submatches[1]
		oldSrc := submatches[2]
		suffix := submatches[3]

		// 跳过 data URI 和外部 URL
		if strings.HasPrefix(oldSrc, "data:") || strings.HasPrefix(oldSrc, "http://") ||
			strings.HasPrefix(oldSrc, "https://") || strings.HasPrefix(oldSrc, "//") {
			return match
		}

		// 解析原路径
		var fullPath string
		if path.IsAbs(oldSrc) {
			fullPath = oldSrc
		} else {
			fullPath = path.Join(htmlDir, oldSrc)
		}
		fullPath = path.Clean(fullPath)

		// 查找对应的 EPUB 路径
		if epubPath, ok := imageMap[fullPath]; ok {
			// 计算相对于 HTML 文件的路径
			relPath := calculateRelativePath(epubChapterDir, epubPath)
			return prefix + relPath + suffix
		}

		return match
	})

	// 更新 CSS url() 中的图片路径
	urlRegex := regexp.MustCompile(`(url\s*\(\s*["'])([^"')]+)(["']\s*\))`)
	htmlContent = urlRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
		submatches := urlRegex.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}
		prefix := submatches[1]
		oldUrl := submatches[2]
		suffix := submatches[3]

		// 跳过 data URI 和外部 URL
		if strings.HasPrefix(oldUrl, "data:") || strings.HasPrefix(oldUrl, "http://") ||
			strings.HasPrefix(oldUrl, "https://") || strings.HasPrefix(oldUrl, "//") {
			return match
		}

		// 解析原路径
		var fullPath string
		if path.IsAbs(oldUrl) {
			fullPath = oldUrl
		} else {
			fullPath = path.Join(htmlDir, oldUrl)
		}
		fullPath = path.Clean(fullPath)

		// 查找对应的 EPUB 路径
		if epubPath, ok := imageMap[fullPath]; ok {
			// 计算相对于 HTML 文件的路径
			relPath := calculateRelativePath(epubChapterDir, epubPath)
			return prefix + relPath + suffix
		}

		return match
	})

	return htmlContent
}

// calculateRelativePath 计算从 fromDir 到 toPath 的相对路径
func calculateRelativePath(fromDir, toPath string) string {
	if fromDir == "" || fromDir == "." {
		// HTML 在根目录，图片路径就是相对于根目录的路径
		return toPath
	}

	// 找到共同的前缀
	fromParts := strings.Split(normalizeZipPath(fromDir), "/")
	toParts := strings.Split(normalizeZipPath(toPath), "/")

	// 找到共同前缀的长度
	commonLen := 0
	minLen := len(fromParts)
	if len(toParts) < minLen {
		minLen = len(toParts)
	}
	for i := 0; i < minLen; i++ {
		if fromParts[i] == toParts[i] {
			commonLen++
		} else {
			break
		}
	}

	// 构建相对路径
	var relParts []string
	// 向上回到共同目录
	for i := commonLen; i < len(fromParts); i++ {
		relParts = append(relParts, "..")
	}
	// 向下到目标文件
	relParts = append(relParts, toParts[commonLen:]...)

	if len(relParts) == 0 {
		return path.Base(toPath)
	}

	return strings.Join(relParts, "/")
}

func isHTMLEntry(entry *zipEntry) bool {
	return strings.HasSuffix(strings.ToLower(entry.header.Name), ".html") ||
		strings.HasSuffix(strings.ToLower(entry.header.Name), ".xhtml") ||
		strings.HasSuffix(strings.ToLower(entry.header.Name), ".htm")
}

// ---------- 辅助结构 ----------

type opfPackage struct {
	XMLName          xml.Name    `xml:"package"`
	XMLNS            string      `xml:"xmlns,attr,omitempty"`
	XMLNSDC          string      `xml:"xmlns:dc,attr,omitempty"`
	XMLNSOPF         string      `xml:"xmlns:opf,attr,omitempty"`
	XMLNSDCTerms     string      `xml:"xmlns:dcterms,attr,omitempty"`
	XMLLang          string      `xml:"xml:lang,attr,omitempty"`
	Version          string      `xml:"version,attr,omitempty"`
	UniqueIdentifier string      `xml:"unique-identifier,attr,omitempty"`
	Prefix           string      `xml:"prefix,attr,omitempty"`
	Metadata         opfMetadata `xml:"metadata"`
	Manifest         opfManifest `xml:"manifest"`
	Spine            opfSpine    `xml:"spine"`
	Guide            *opfGuide   `xml:"guide,omitempty"`
}

type opfMetadata struct {
	InnerXML []byte `xml:",innerxml"`
}

type opfManifest struct {
	Items []opfManifestItem `xml:"item"`
}

type opfManifestItem struct {
	ID           string `xml:"id,attr"`
	Href         string `xml:"href,attr"`
	MediaType    string `xml:"media-type,attr"`
	Properties   string `xml:"properties,attr,omitempty"`
	Fallback     string `xml:"fallback,attr,omitempty"`
	MediaOverlay string `xml:"media-overlay,attr,omitempty"`
}

type opfSpine struct {
	Toc                      string         `xml:"toc,attr,omitempty"`
	PageProgressionDirection string         `xml:"page-progression-direction,attr,omitempty"`
	Items                    []opfSpineItem `xml:"itemref"`
}

type opfSpineItem struct {
	IDRef      string `xml:"idref,attr"`
	Linear     string `xml:"linear,attr,omitempty"`
	Properties string `xml:"properties,attr,omitempty"`
}

type opfGuide struct {
	InnerXML []byte `xml:",innerxml"`
}

func writeDirEntry(writer *zip.Writer, header *zip.FileHeader) error {
	dirHeader := *header
	dirHeader.Method = zip.Store
	if !strings.HasSuffix(dirHeader.Name, "/") {
		dirHeader.Name += "/"
	}
	if _, err := writer.CreateHeader(&dirHeader); err != nil {
		return fmt.Errorf("failed to write directory (%s): %w", header.Name, err)
	}
	return nil
}

func writeFileEntry(writer *zip.Writer, header *zip.FileHeader, data []byte) error {
	fileHeader := *header
	if fileHeader.Method == 0 {
		fileHeader.Method = zip.Deflate
	}
	w, err := writer.CreateHeader(&fileHeader)
	if err != nil {
		return fmt.Errorf("failed to write file (%s): %w", header.Name, err)
	}
	if _, err = w.Write(data); err != nil {
		return fmt.Errorf("failed to write file content (%s): %w", header.Name, err)
	}
	return nil
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

func normalizeZipPath(pth string) string {
	pth = strings.ReplaceAll(pth, "\\", "/")
	return path.Clean(pth)
}

func shouldRemoveHTML(htmlContent string, keywords []string) bool {
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(htmlContent, keyword) {
			return true
		}
	}
	return false
}

func isOPFFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".opf")
}
