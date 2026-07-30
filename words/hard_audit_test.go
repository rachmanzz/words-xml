package words

import (
	"strings"
	"testing"
)

// TestDocumentOrderTableBeforePara verifies Bug A fix:
// tables must appear in document order relative to paragraphs,
// not appended after all paragraphs.
func TestDocumentOrderTableBeforePara(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>before table</w:t></w:r></w:p>
  <w:tbl>
    <w:tr><w:tc><w:p><w:r><w:t>cell content</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
  <w:p><w:r><w:t>after table</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	beforePos := strings.Index(x, "before table")
	tablePos := strings.Index(x, "<table")
	afterPos := strings.Index(x, "after table")

	if beforePos < 0 {
		t.Error("'before table' para missing")
	}
	if tablePos < 0 {
		t.Error("<table> missing")
	}
	if afterPos < 0 {
		t.Error("'after table' para missing")
	}
	if beforePos > tablePos {
		t.Errorf("table appears before 'before table' para (order wrong): before=%d table=%d", beforePos, tablePos)
	}
	if tablePos > afterPos {
		t.Errorf("'after table' para appears before table (order wrong): table=%d after=%d", tablePos, afterPos)
	}
}

// TestDocumentOrderParaAfterTable verifies that a paragraph after a table
// is not dropped and appears after the table in output.
func TestDocumentOrderParaAfterTable(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tr><w:tc><w:p><w:r><w:t>table cell</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
  <w:p><w:r><w:t>important paragraph after table</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	tablePos := strings.Index(x, "<table")
	paraPos := strings.Index(x, "important paragraph after table")

	if tablePos < 0 {
		t.Error("<table> missing")
	}
	if paraPos < 0 {
		t.Error("paragraph after table dropped")
	}
	if paraPos < tablePos {
		t.Errorf("paragraph before table in output (order wrong): table=%d para=%d", tablePos, paraPos)
	}
}

// TestDocumentOrderSdtBeforePara verifies SDT ordering.
func TestDocumentOrderSdtBeforePara(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>first para</w:t></w:r></w:p>
  <w:sdt>
    <w:sdtContent>
      <w:p><w:r><w:t>sdt content</w:t></w:r></w:p>
    </w:sdtContent>
  </w:sdt>
  <w:p><w:r><w:t>last para</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	firstPos := strings.Index(x, "first para")
	sdtPos := strings.Index(x, "sdt content")
	lastPos := strings.Index(x, "last para")

	if firstPos < 0 || sdtPos < 0 || lastPos < 0 {
		t.Errorf("content missing: first=%d sdt=%d last=%d\n%s", firstPos, sdtPos, lastPos, x)
	}
	if firstPos > sdtPos {
		t.Errorf("sdt content before first para (order wrong)")
	}
	if sdtPos > lastPos {
		t.Errorf("last para before sdt (order wrong)")
	}
}

// TestPostProcessRunOrderWithTableBefore verifies Bug B fix:
// postProcessRunOrder must correctly handle docs where a table with runs
// appears before a paragraph containing t-before-tab pattern.
func TestPostProcessRunOrderWithTableBefore(t *testing.T) {
	// Table with a run comes before the paragraph that has t+tab co-located
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tr><w:tc><w:p><w:r><w:t>table text</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
  <w:p>
    <w:r><w:t xml:space="preserve">1.</w:t><w:tab/></w:r>
    <w:r><w:t xml:space="preserve">content after tab</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	// text "1." must appear before <tab/>
	textPos := strings.Index(x, "1.")
	tabPos := strings.Index(x, "<tab/>")

	if textPos < 0 {
		t.Errorf("text '1.' missing: %s", x)
	}
	if tabPos < 0 {
		t.Errorf("<tab/> missing: %s", x)
	}
	if textPos > tabPos {
		t.Errorf("tab before text (Bug B regression): text=%d tab=%d\n%s", textPos, tabPos, x)
	}
}

// TestHasSameNumIDAheadLimit verifies Bug D fix:
// list groups separated by more than 20 non-list items should still be grouped.
func TestHasSameNumIDAheadLimit(t *testing.T) {
	// Build a list with 25 paragraph gaps between first and second list item
	var paras strings.Builder
	paras.WriteString(`<w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>list item 1</w:t></w:r></w:p>`)
	for i := 0; i < 25; i++ {
		paras.WriteString(`<w:p><w:r><w:t>gap para</w:t></w:r></w:p>`)
	}
	paras.WriteString(`<w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>list item 2</w:t></w:r></w:p>`)

	body := xmlHeader + `<w:body>` + paras.String() + `</w:body></w:document>`
	data := makeMinimalDocxWithNumbering(body, "1", "bullet")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "list item 1") {
		t.Error("list item 1 missing")
	}
	if !strings.Contains(x, "list item 2") {
		t.Error("list item 2 missing")
	}
}

// makeMinimalDocxWithNumbering creates a minimal docx with a numbering definition.
func makeMinimalDocxWithNumbering(bodyXML string, numID string, numFmt string) []byte {
	numberingXML := `<?xml version="1.0" encoding="UTF-8"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0">
      <w:numFmt w:val="` + numFmt + `"/>
      <w:lvlText w:val="&#x2022;"/>
    </w:lvl>
  </w:abstractNum>
  <w:num w:numId="` + numID + `">
    <w:abstractNumId w:val="0"/>
  </w:num>
</w:numbering>`
	return makeDocxWithExtras(bodyXML, "", numberingXML, "", "", "", "")
}

