package epub

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"
)

func TestEpub_Open(t *testing.T) {
	// 创建一个内存中的最小 EPUB 结构用于测试
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// 添加 container.xml (虽然 Open 里没校验，但为了合规)
	f, _ := zw.Create("META-INF/container.xml")
	f.Write([]byte(`<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>`))

	// 添加 content.opf
	f, _ = zw.Create("content.opf")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="uid"><metadata></metadata><manifest><item id="item1" href="page1.html" media-type="application/xhtml+xml"/></manifest><spine><itemref idref="item1"/></spine></package>`))

	// 添加 page1.html
	f, _ = zw.Create("page1.html")
	f.Write([]byte(`<html><body>Hello World</body></html>`))

	zw.Close()

	tmpFile := "test_sample.epub"
	os.WriteFile(tmpFile, buf.Bytes(), 0644)
	defer os.Remove(tmpFile)

	e, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open epub: %v", err)
	}
	defer e.Close()

	if e.opfPath != "content.opf" {
		t.Errorf("expected opfPath content.opf, got %s", e.opfPath)
	}

	// 测试 ApplyHTML
	mod, err := e.ApplyHTML(func(name, html string) (string, error) {
		return "<html><body>Hello Go</body></html>", nil
	})
	if err != nil {
		t.Errorf("applyHTML failed: %v", err)
	}
	if mod != 1 {
		t.Errorf("expected 1 modification, got %d", mod)
	}

	// 测试 Save
	savePath := "test_saved.epub"
	if err := e.Save(savePath); err != nil {
		t.Errorf("save failed: %v", err)
	}
	defer os.Remove(savePath)

	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("saved file does not exist")
	}
}
