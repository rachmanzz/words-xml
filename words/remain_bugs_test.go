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
	if strings.Contains(x, "deleted") {
		t.Errorf("para-level w:del text 'deleted' should be dropped in semantic mode: %s", x)
	}
}

// TestParaLevelInsDelLossless verifies that w:del text is kept (wrapped in <del>)
// when the document is processed in lossless mode.
func TestParaLevelInsDelLossless(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t xml:space="preserve">normal </w:t></w:r>
    <w:del w:id="2" w:author="Bob" w:date="2024-01-02T00:00:00Z">
      <w:r><w:delText xml:space="preserve">deleted</w:delText></w:r>
    </w:del>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytesMode(data, "lossless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<del") {
		t.Errorf("expected <del> in lossless mode, got: %s", x)
	}
	if !strings.Contains(x, "deleted") {
		t.Errorf("para-level w:del text 'deleted' missing in lossless mode: %s", x)
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

// TestHeadingBreaksList verifies that a heading between list items
// causes the heading to be emitted as <hN> rather than absorbed into a <li>.
func TestHeadingBreaksList(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Heading3">
    <w:name w:val="heading 3"/>
    <w:pPr><w:outlineLvl w:val="2"/></w:pPr>
  </w:style>
</w:styles>`

	numbering := `<?xml version="1.0" encoding="UTF-8"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0"><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="1"/></w:num>
</w:numbering>`

	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr>
    <w:r><w:t>item one</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr><w:pStyle w:val="Heading3"/></w:pPr>
    <w:r><w:t>A Section Heading</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr>
    <w:r><w:t>item two</w:t></w:r>
  </w:p>
</w:body></w:document>`

	data := makeDocxWithExtras(body, styles, numbering, "", "", "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	// Heading should appear as <h3>, not absorbed into <li>
	if !strings.Contains(x, "<h3") {
		t.Errorf("heading not emitted as <h3>: %s", x)
	}
	headingPos := strings.Index(x, "A Section Heading")
	liOnePos := strings.Index(x, "item one")
	liTwoPos := strings.Index(x, "item two")
	if headingPos < 0 || liOnePos < 0 || liTwoPos < 0 {
		t.Fatalf("missing content: heading=%d liOne=%d liTwo=%d\n%s", headingPos, liOnePos, liTwoPos, x)
	}
	// Order: item one < heading < item two
	if !(liOnePos < headingPos && headingPos < liTwoPos) {
		t.Errorf("order wrong: liOne=%d heading=%d liTwo=%d", liOnePos, headingPos, liTwoPos)
	}
	// Heading must NOT be inside a <li> — check no <br type="textWrapping"> precedes it
	brPos := strings.LastIndex(x[:headingPos], `<br type="textWrapping"/>`)
	liClosePos := strings.LastIndex(x[:headingPos], `</li>`)
	if brPos > liClosePos {
		t.Errorf("heading appears to be inside a <li> (preceded by <br type=textWrapping>): %s", x)
	}
}

// TestSectionBreakBetweenListItems verifies that a horizontal-rule paragraph
// (and following non-list text) sandwiched between two list items sharing the
// same numId is preserved as continuation <p> children of the preceding <li>
// instead of being silently dropped. Both items must stay in the same <ol> so
// the consumer's automatic numbering (1., 2.) is not reset.
func TestSectionBreakBetweenListItems(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="1"/></w:num>
</w:numbering>`

	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr>
    <w:r><w:t>item one</w:t></w:r>
  </w:p>
  <w:p>
    <w:r><w:t>-------------- "............"</w:t></w:r>
  </w:p>
  <w:p>
    <w:r><w:t>(selanjutnya cukup disingkat)</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr>
    <w:r><w:t>item two</w:t></w:r>
  </w:p>
</w:body></w:document>`

	data := makeDocxWithExtras(body, "", numbering, "", "", "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	// Both separator paragraphs must be preserved somewhere in the output.
	sepPos := strings.Index(x, `-------------- &quot;............&quot;`)
	contPos := strings.Index(x, "selanjutnya cukup disingkat")
	if sepPos < 0 || contPos < 0 {
		t.Fatalf("separator/continuation content missing: sep=%d cont=%d\n%s", sepPos, contPos, x)
	}
	onePos := strings.Index(x, "item one")
	twoPos := strings.Index(x, "item two")
	if onePos < 0 || twoPos < 0 {
		t.Fatalf("list item content missing: one=%d two=%d", onePos, twoPos)
	}
	// Both items stay in one <ol>; separator sits between them inside the <li>.
	if strings.Count(x, "<ol ") != 1 {
		t.Errorf("expected a single <ol>, got %d\n%s", strings.Count(x, "<ol "), x)
	}
	if strings.Count(x, "<li>") != 2 {
		t.Errorf("expected two <li> items, got %d\n%s", strings.Count(x, "<li>"), x)
	}
	// Separator and continuation text must be inside the first <li> (before </li>).
	firstLiClose := strings.Index(x, "</li>")
	if firstLiClose < 0 || !(onePos < sepPos && sepPos < firstLiClose) {
		t.Errorf("separator not inside the first <li>: one=%d sep=%d liClose=%d\n%s", onePos, sepPos, firstLiClose, x)
	}
	if !(firstLiClose < twoPos) {
		t.Errorf("item two must follow the first </li>: liClose=%d two=%d", firstLiClose, twoPos)
	}
}

// TestSectionBreakAfterLastListItem verifies that a horizontal-rule paragraph
// after the LAST item of a list is NOT absorbed into the <li> — it terminates
// the list and stays a standalone <p> after </ol>.
func TestSectionBreakAfterLastListItem(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="1"/></w:num>
</w:numbering>`

	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr>
    <w:r><w:t>only item</w:t></w:r>
  </w:p>
  <w:p>
    <w:r><w:t>------------------- DEMIKIAN AKTA INI -----------------</w:t></w:r>
  </w:p>
</w:body></w:document>`

	data := makeDocxWithExtras(body, "", numbering, "", "", "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	itemPos := strings.Index(x, "only item")
	hrPos := strings.Index(x, "DEMIKIAN AKTA INI")
	if itemPos < 0 || hrPos < 0 {
		t.Fatalf("content missing: item=%d hr=%d\n%s", itemPos, hrPos, x)
	}
	olEnd := strings.Index(x, "</ol>")
	liEnd := strings.Index(x, "</li>")
	// The hr must be outside the list, after both </li> and </ol>.
	if !(hrPos > olEnd && olEnd > liEnd) {
		t.Errorf("hr must be standalone after </ol>: liEnd=%d olEnd=%d hr=%d\n%s", liEnd, olEnd, hrPos, x)
	}
}

// TestListResumeNumberingAfterInterleavedSubList verifies that a list group
// which resumes a previously emitted numbering sequence (same numId, split by
// an interleaved different-numId sub-list) carries the continuation number in
// a start attribute — Word numbers the decimal items 1,2 then a,b then 3.
func TestListResumeNumberingAfterInterleavedSubList(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="10">
    <w:lvl w:ilvl="0"><w:numFmt w:val="lowerLetter"/><w:lvlText w:val="%1."/></w:lvl>
  </w:abstractNum>
  <w:abstractNum w:abstractNumId="32">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="22"><w:abstractNumId w:val="32"/></w:num>
  <w:num w:numId="23"><w:abstractNumId w:val="10"/></w:num>
</w:numbering>`

	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="22"/></w:numPr></w:pPr>
    <w:r><w:t>item one</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="22"/></w:numPr></w:pPr>
    <w:r><w:t>item two</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="23"/></w:numPr></w:pPr>
    <w:r><w:t>sub item a</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="22"/></w:numPr></w:pPr>
    <w:r><w:t>item three</w:t></w:r>
  </w:p>
</w:body></w:document>`

	data := makeDocxWithExtras(body, "", numbering, "", "", "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML

	if !strings.Contains(x, `<ol type="decimal" start="3">`) {
		t.Errorf("resumed decimal list missing start=\"3\":\n%s", x)
	}
	// The interleaved sub-list keeps its own base numbering (no start).
	if !strings.Contains(x, `<ol type="lowerLetter">`) {
		t.Errorf("interleaved sub-list missing:\n%s", x)
	}
	// Items stay ordered: 1,2 in the first group, a. then 3. in the resumed group.
	onePos := strings.Index(x, "item one")
	threePos := strings.Index(x, "item three")
	startPos := strings.Index(x, `start="3"`)
	if !(onePos < startPos && startPos < threePos) {
		t.Errorf("resumed group must follow the first group and precede item three: one=%d start=%d three=%d\n%s", onePos, startPos, threePos, x)
	}
	if strings.Count(x, `<ol type="decimal">`) != 1 {
		t.Errorf("expected exactly one base decimal group, got %d\n%s", strings.Count(x, `<ol type="decimal">`), x)
	}
}
