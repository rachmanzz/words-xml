package words

import (
	"strings"
	"testing"
)

// TestHyperlinkTabRun verifies Bug 5 fix:
// a run inside a hyperlink that contains only a tab (no text) should emit <tab/>.
func TestHyperlinkTabRun(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:hyperlink r:id="rId1" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
      <w:r><w:t xml:space="preserve">click</w:t></w:r>
      <w:r><w:tab/></w:r>
      <w:r><w:t xml:space="preserve">here</w:t></w:r>
    </w:hyperlink>
  </w:p>
</w:body></w:document>`
	rels := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.com" TargetMode="External"/>
</Relationships>`
	data := makeDocxWithExtras(body, "", "", rels, "", "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<tab/>") {
		t.Errorf("tab inside hyperlink run not emitted: %s", x)
	}
	if !strings.Contains(x, "click") {
		t.Errorf("hyperlink text 'click' missing: %s", x)
	}
	if !strings.Contains(x, "here") {
		t.Errorf("hyperlink text 'here' missing: %s", x)
	}
}

// TestParaLevelInsDel verifies Bug 3 fix:
// w:ins/w:del directly wrapping w:r inside w:p (para-level) should be captured.
func TestParaLevelInsDel(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t xml:space="preserve">normal </w:t></w:r>
    <w:ins w:id="1" w:author="Alice" w:date="2024-01-01T00:00:00Z">
      <w:r><w:t xml:space="preserve">inserted</w:t></w:r>
    </w:ins>
    <w:del w:id="2" w:author="Bob" w:date="2024-01-02T00:00:00Z">
      <w:r><w:delText xml:space="preserve">deleted</w:delText></w:r>
    </w:del>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "normal") {
		t.Errorf("normal text missing: %s", x)
	}
	if !strings.Contains(x, "inserted") {
		t.Errorf("para-level w:ins text 'inserted' missing: %s", x)
	}
	if !strings.Contains(x, "deleted") {
		t.Errorf("para-level w:del text 'deleted' missing: %s", x)
	}
}

// TestTblCellDocumentOrder verifies Bug 2 fix (DocTblCell):
// a table cell with a table followed by a paragraph preserves that order.
func TestTblCellDocumentOrder(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tr>
      <w:tc>
        <w:tbl>
          <w:tr><w:tc><w:p><w:r><w:t>inner table</w:t></w:r></w:p></w:tc></w:tr>
        </w:tbl>
        <w:p><w:r><w:t>after inner table</w:t></w:r></w:p>
      </w:tc>
    </w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	innerTablePos := strings.Index(x, "inner table")
	afterPos := strings.Index(x, "after inner table")
	if innerTablePos < 0 {
		t.Errorf("inner table content missing: %s", x)
	}
	if afterPos < 0 {
		t.Errorf("'after inner table' para missing (dropped)")
	}
	if innerTablePos > afterPos {
		t.Errorf("paragraph appears before inner table (order wrong): table=%d para=%d", innerTablePos, afterPos)
	}
}

// TestSdtContentDocumentOrder verifies Bug 2 fix (DocSdtContent):
// an SDT containing a table followed by a paragraph preserves that order.
func TestSdtContentDocumentOrder(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:sdt>
    <w:sdtContent>
      <w:tbl>
        <w:tr><w:tc><w:p><w:r><w:t>sdt table</w:t></w:r></w:p></w:tc></w:tr>
      </w:tbl>
      <w:p><w:r><w:t>after sdt table</w:t></w:r></w:p>
    </w:sdtContent>
  </w:sdt>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	tblPos := strings.Index(x, "sdt table")
	paraPos := strings.Index(x, "after sdt table")
	if tblPos < 0 {
		t.Errorf("sdt table content missing")
	}
	if paraPos < 0 {
		t.Errorf("'after sdt table' para missing (dropped)")
	}
	if tblPos > paraPos {
		t.Errorf("paragraph before table in sdt (order wrong): table=%d para=%d", tblPos, paraPos)
	}
}

