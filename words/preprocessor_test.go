package words

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
	"testing"
)

func makeMinimalDocx(bodyXML string) []byte {
	return makeDocxWithParts(bodyXML, "", "", "")
}

func makeDocxWithFootnotes(bodyXML, footnotesXML string) []byte {
	return makeDocxWithExtras(bodyXML, "", "", "", footnotesXML, "", "")
}

func makeDocxWithComments(bodyXML, commentsXML string) []byte {
	return makeDocxWithExtras(bodyXML, "", "", "", "", "", commentsXML)
}

func makeDocxWithExtras(bodyXML, stylesXML, numberingXML, relsXML, footnotesXML, endnotesXML, commentsXML string) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	addZipFile := func(name, content string) {
		if content == "" {
			return
		}
		f, _ := w.Create(name)
		f.Write([]byte(content))
	}

	addZipFile("word/document.xml", bodyXML)
	if stylesXML != "" {
		addZipFile("word/styles.xml", stylesXML)
	}
	if numberingXML != "" {
		addZipFile("word/numbering.xml", numberingXML)
	}
	if relsXML != "" {
		addZipFile("word/_rels/document.xml.rels", relsXML)
	}
	if footnotesXML != "" {
		addZipFile("word/footnotes.xml", footnotesXML)
	}
	if endnotesXML != "" {
		addZipFile("word/endnotes.xml", endnotesXML)
	}
	if commentsXML != "" {
		addZipFile("word/comments.xml", commentsXML)
	}

	w.Close()
	return buf.Bytes()
}

func makeDocxWithParts(bodyXML, stylesXML, numberingXML, relsXML string) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	addZipFile := func(name, content string) {
		if content == "" {
			return
		}
		f, _ := w.Create(name)
		f.Write([]byte(content))
	}

	addZipFile("word/document.xml", bodyXML)
	if stylesXML != "" {
		addZipFile("word/styles.xml", stylesXML)
	}
	if numberingXML != "" {
		addZipFile("word/numbering.xml", numberingXML)
	}
	if relsXML != "" {
		addZipFile("word/_rels/document.xml.rels", relsXML)
	}

	w.Close()
	return buf.Bytes()
}

func makeFullDocx(bodyXML, stylesXML, numberingXML, relsXML, coreXML, sectPr string) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	addZipFile := func(name, content string) {
		if content == "" {
			return
		}
		f, _ := w.Create(name)
		f.Write([]byte(content))
	}

	if sectPr != "" {
		if strings.Contains(bodyXML, "<w:body>") {
			bodyXML = strings.Replace(bodyXML, "</w:body>", sectPr+"</w:body>", 1)
		} else {
			bodyXML = strings.Replace(bodyXML, "</w:document>", "<w:body>"+sectPr+"</w:body></w:document>", 1)
		}
	}

	addZipFile("word/document.xml", bodyXML)
	addZipFile("[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`)
	addZipFile("_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`)
	addZipFile("word/_rels/document.xml.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`)

	if stylesXML != "" {
		addZipFile("word/styles.xml", stylesXML)
	}
	if numberingXML != "" {
		addZipFile("word/numbering.xml", numberingXML)
	}
	if relsXML != "" {
		addZipFile("word/_rels/document.xml.rels", relsXML)
	}
	if coreXML != "" {
		addZipFile("docProps/core.xml", coreXML)
	}

	w.Close()
	return buf.Bytes()
}

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
            xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006"
            xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing">`

func TestMinimalDocx(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>Hello World</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "Hello World") {
		t.Error("expected 'Hello World' in output")
	}
	if !strings.Contains(doc.WordsXML, "<p>") {
		t.Error("expected <p> element")
	}
}

func TestHeadingMapping(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Title</w:t></w:r></w:p>
  <w:p><w:pPr><w:pStyle w:val="Heading2"/></w:pPr><w:r><w:t>Subtitle</w:t></w:r></w:p>
  <w:p><w:pPr><w:pStyle w:val="Heading3"/></w:pPr><w:r><w:t>Section</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<h1>") {
		t.Error("expected <h1> for Heading1")
	}
	if !strings.Contains(doc.WordsXML, "<h2>") {
		t.Error("expected <h2> for Heading2")
	}
	if !strings.Contains(doc.WordsXML, "<h3>") {
		t.Error("expected <h3> for Heading3")
	}
}

func TestFalseHeadingSuppression(t *testing.T) {
	longText := "On this very long day, in the year of our Lord two thousand and twenty-six, the parties gathered"
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:pStyle w:val="Heading1"/><w:jc w:val="both"/></w:pPr><w:r><w:t>` + longText + `</w:t></w:r></w:p>
  <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>This is a short heading</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(doc.WordsXML, "<h1>"+longText[:20]) {
		t.Error("long paragraph with jc=both should NOT be h1")
	}
	if !strings.Contains(doc.WordsXML, "<h1>This is a short heading") {
		t.Error("short paragraph should still be h1")
	}
}

func TestInlineFormatting(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:b/></w:rPr><w:t>Bold</w:t></w:r>
    <w:r><w:rPr><w:i/></w:rPr><w:t>Italic</w:t></w:r>
    <w:r><w:rPr><w:u w:val="single"/></w:rPr><w:t>Underline</w:t></w:r>
    <w:r><w:rPr><w:strike/></w:rPr><w:t>Strike</w:t></w:r>
    <w:r><w:rPr><w:smallCaps/></w:rPr><w:t>SmallCaps</w:t></w:r>
    <w:r><w:rPr><w:vertAlign w:val="superscript"/></w:rPr><w:t>Sup</w:t></w:r>
    <w:r><w:rPr><w:vertAlign w:val="subscript"/></w:rPr><w:t>Sub</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	for _, tc := range []struct {
		name, tag string
	}{
		{"bold", "<b>Bold</b>"},
		{"italic", "<i>Italic</i>"},
		{"underline", "<u>Underline</u>"},
		{"strike", "<s>Strike</s>"},
		{"smallcaps", "<smallcaps>SmallCaps</smallcaps>"},
		{"superscript", "<sup>Sup</sup>"},
		{"subscript", "<sub>Sub</sub>"},
	} {
		if !strings.Contains(x, tc.tag) {
			t.Errorf("expected %s: %s", tc.name, tc.tag)
		}
	}
}

func TestFontSpan(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="24"/><w:color w:val="FF0000"/></w:rPr><w:t>styled</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `font="Arial"`) {
		t.Error("expected font=Arial")
	}
	if !strings.Contains(x, `size="12"`) {
		t.Error("expected size=12 (24 half-points)")
	}
	if !strings.Contains(x, `color="FF0000"`) {
		t.Error("expected color=FF0000")
	}
}

func TestBoldCSItalicCS(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:bCs/></w:rPr><w:t>BoldCS</w:t></w:r>
    <w:r><w:rPr><w:iCs/></w:rPr><w:t>ItalicCS</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<bcs>BoldCS</bcs>") {
		t.Error("expected <bcs>BoldCS</bcs>")
	}
	if !strings.Contains(x, "<ics>ItalicCS</ics>") {
		t.Error("expected <ics>ItalicCS</ics>")
	}
}

func TestLineBreak(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t>before</w:t></w:r>
    <w:r><w:br w:type="textWrapping"/></w:r>
    <w:r><w:t>after</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<br type="textWrapping"/>`) {
		t.Error("expected <br type=\"textWrapping\"/>")
	}
}

func TestTab(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t>before</w:t></w:r>
    <w:r><w:tab/></w:r>
    <w:r><w:t>after</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<tab/>") {
		t.Error("expected <tab/>")
	}
}

func TestHyperlink(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:hyperlink r:id="rId1"><w:r><w:t>click here</w:t></w:r></w:hyperlink>
  </w:p>
</w:body></w:document>`
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.com" TargetMode="External"/>
</Relationships>`
	data := makeDocxWithParts(body, "", "", rels)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<a href="https://example.com">click here</a>`) {
		t.Errorf("expected hyperlink, got: %s", doc.WordsXML)
	}
}

func TestHyperlinkBold(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:hyperlink r:id="rId1">
      <w:r><w:rPr><w:b/></w:rPr><w:t>bold link</w:t></w:r>
    </w:hyperlink>
  </w:p>
</w:body></w:document>`
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.com" TargetMode="External"/>
</Relationships>`
	data := makeDocxWithParts(body, "", "", rels)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<a href="https://example.com"><b>bold link</b></a>`) {
		t.Errorf("expected bold inside hyperlink, got: %s", doc.WordsXML)
	}
}

func TestQuote(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Quote">
    <w:name w:val="Quote"/>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:pStyle w:val="Quote"/></w:pPr><w:r><w:t>quoted text</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<blockquote>") {
		t.Error("expected <blockquote> for Quote style")
	}
}

func TestBulletList(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="bullet"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Item 1</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Item 2</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<ul type="bullet">`) {
		t.Error("expected <ul type=\"bullet\">")
	}
	if !strings.Contains(x, "<li>Item 1") {
		t.Error("expected <li>Item 1")
	}
	if !strings.Contains(x, "<li>Item 2") {
		t.Error("expected <li>Item 2")
	}
}

func TestOrderedList(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>First</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Second</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<ol type="decimal">`) {
		t.Errorf("expected <ol type=\"decimal\">, got: %s", x)
	}
}

func TestNestedList(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="bullet"/></w:lvl>
    <w:lvl w:ilvl="1"><w:numFmt w:val="bullet"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Parent</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="1"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Child</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<li>Parent") {
		t.Error("expected <li>Parent")
	}
	if !strings.Contains(x, "<ul") {
		t.Error("expected nested <ul>")
	}
	if !strings.Contains(x, "<li>Child") {
		t.Error("expected <li>Child")
	}
}

func TestTable(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr>
      <w:tblW w:w="5000" w:type="dxa"/>
      <w:jc w:val="center"/>
    </w:tblPr>
    <w:tblGrid>
      <w:gridCol w:w="2500"/>
      <w:gridCol w:w="2500"/>
    </w:tblGrid>
    <w:tr>
      <w:tc><w:tcPr><w:gridSpan w:val="2"/></w:tcPr><w:p><w:r><w:t>Header</w:t></w:r></w:p></w:tc>
    </w:tr>
    <w:tr>
      <w:tc><w:p><w:r><w:t>Cell 1</w:t></w:r></w:p></w:tc>
      <w:tc><w:p><w:r><w:t>Cell 2</w:t></w:r></w:p></w:tc>
    </w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<table id="1"`) {
		t.Error("expected <table id=\"1\">")
	}
	if !strings.Contains(x, `colspan="2"`) {
		t.Error("expected colspan=\"2\"")
	}
	if !strings.Contains(x, "Header") {
		t.Error("expected Header cell")
	}
	if !strings.Contains(x, "Cell 1") {
		t.Error("expected Cell 1")
	}
}

func TestTableBorders(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr>
      <w:tblBorders>
        <w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/>
        <w:bottom w:val="single" w:sz="4" w:space="0" w:color="000000"/>
        <w:left w:val="single" w:sz="4" w:space="0" w:color="000000"/>
        <w:right w:val="single" w:sz="4" w:space="0" w:color="000000"/>
      </w:tblBorders>
    </w:tblPr>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "at=") {
		t.Error("expected at= attribute for borders")
	}
	if !strings.Contains(x, "bt") {
		t.Error("expected bt (border top)")
	}
	if !strings.Contains(x, "bb") {
		t.Error("expected bb (border bottom)")
	}
	if !strings.Contains(x, "bl") {
		t.Error("expected bl (border left)")
	}
	if !strings.Contains(x, "br") {
		t.Error("expected br (border right)")
	}
}

func TestCellBorders(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc>
      <w:tcPr>
        <w:tcBorders>
          <w:top w:val="single" w:sz="4" w:space="0" w:color="FF0000"/>
          <w:bottom w:val="single" w:sz="4" w:space="0" w:color="00FF00"/>
        </w:tcBorders>
      </w:tcPr>
      <w:p><w:r><w:t>Bordered Cell</w:t></w:r></w:p>
    </w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "at=") {
		t.Error("expected at= attribute for cell borders")
	}
	if !strings.Contains(x, "FF0000") {
		t.Error("expected FF0000 in border color")
	}
}

func TestVMerge(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc>
      <w:tcPr><w:vMerge w:val="restart"/></w:tcPr>
      <w:p><w:r><w:t>Merged</w:t></w:r></w:p>
    </w:tc></w:tr>
    <w:tr><w:tc>
      <w:tcPr><w:vMerge/></w:tcPr>
      <w:p><w:r><w:t>Continue</w:t></w:r></w:p>
    </w:tc></w:tr>
    <w:tr><w:tc>
      <w:tcPr><w:vMerge/></w:tcPr>
      <w:p><w:r><w:t>Continue2</w:t></w:r></w:p>
    </w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `rowspan="3"`) {
		t.Errorf("expected rowspan=\"3\", got: %s", x)
	}
	if strings.Contains(x, "Continue") {
		t.Errorf("continue cells should be omitted, got: %s", x)
	}
}

func TestFootnote(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t>Text</w:t></w:r>
    <w:r><w:footnoteReference w:id="1"/></w:r>
  </w:p>
</w:body></w:document>`
	footnotes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:footnotes xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:footnote w:type="normal" w:id="1">
    <w:p><w:r><w:t>This is a footnote.</w:t></w:r></w:p>
  </w:footnote>
</w:footnotes>`
	data := makeDocxWithParts(body, "", "", "")
	_ = footnotes

	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<fn-ref id="1" type="footnote"/>`) {
		t.Errorf("expected fn-ref with type footnote, got: %s", x)
	}
}

func TestEndnote(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t>Text</w:t></w:r>
    <w:r><w:endnoteReference w:id="1"/></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<fn-ref id="1" type="endnote"/>`) {
		t.Errorf("expected fn-ref with type endnote, got: %s", x)
	}
}

func TestBookmark(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:bookmarkStart w:id="0" w:name="Section1"/>
  <w:p><w:r><w:t>Content</w:t></w:r></w:p>
  <w:bookmarkEnd w:id="0"/>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<bm id="Section1"/>`) {
		t.Errorf("expected <bm id=\"Section1\"/>, got: %s", x)
	}
}

func TestBuildBodyNoteOrder(t *testing.T) {
	docXML := []byte(xmlHeader + `<w:body>
  <w:bookmarkStart w:id="0" w:name="Bm1"/>
  <w:p><w:r><w:t>Para 1</w:t></w:r></w:p>
  <w:bookmarkStart w:id="1" w:name="Bm2"/>
  <w:p><w:r><w:t>Para 2</w:t></w:r></w:p>
  <w:p><w:r><w:t>Para 3</w:t></w:r></w:p>
</w:body></w:document>`)
	bmOrder, _ := buildBodyNoteOrder(docXML)
	if bmOrder[0] != 0 {
		t.Errorf("expected Bm1 at pos 0, got %d", bmOrder[0])
	}
	if bmOrder[1] != 2 {
		t.Errorf("expected Bm2 at pos 2, got %d", bmOrder[1])
	}
}

func TestBuildBodyNoteOrderWithFootnoteRef(t *testing.T) {
	docXML := []byte(xmlHeader + `<w:body>
  <w:p><w:r><w:t>Para 1</w:t></w:r></w:p>
  <w:p><w:r><w:footnoteReference w:id="1"/></w:r><w:r><w:t>Para 2</w:t></w:r></w:p>
  <w:p><w:r><w:t>Para 3</w:t></w:r></w:p>
</w:body></w:document>`)
	_, noteRefOrder := buildBodyNoteOrder(docXML)
	key := "footnote_1"
	if pos, ok := noteRefOrder[key]; !ok {
		t.Errorf("expected footnote_1 in noteRefOrder")
	} else if pos != 1 {
		t.Errorf("expected footnote_1 at pos 1, got %d", pos)
	}
}

func TestDocumentOrderNotes(t *testing.T) {
	notesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:footnotes xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:footnote w:type="normal" w:id="1"><w:p><w:r><w:t>Fn1</w:t></w:r></w:p></w:footnote>
  <w:footnote w:type="normal" w:id="2"><w:p><w:r><w:t>Fn2</w:t></w:r></w:p></w:footnote>
</w:footnotes>`
	body := xmlHeader + `<w:body>
  <w:bookmarkStart w:id="0" w:name="Bm1"/>
  <w:p><w:r><w:footnoteReference w:id="1"/></w:r><w:r><w:t>Text</w:t></w:r></w:p>
  <w:bookmarkStart w:id="1" w:name="Bm2"/>
  <w:p><w:r><w:footnoteReference w:id="2"/></w:r><w:r><w:t>Text</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithFootnotes(body, notesXML)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	bm1Idx := strings.Index(x, `<bm id="Bm1"/>`)
	bm2Idx := strings.Index(x, `<bm id="Bm2"/>`)
	fn1Idx := strings.Index(x, `<fn id="1"`)
	fn2Idx := strings.Index(x, `<fn id="2"`)
	if bm1Idx < 0 {
		t.Fatal("Bm1 not found")
	}
	if bm2Idx < 0 {
		t.Fatal("Bm2 not found")
	}
	if fn1Idx < 0 {
		t.Fatal("fn id=1 not found")
	}
	if fn2Idx < 0 {
		t.Fatal("fn id=2 not found")
	}
	if bm1Idx > fn1Idx {
		t.Errorf("Bm1 should come before fn1: bm1=%d fn1=%d", bm1Idx, fn1Idx)
	}
	if fn1Idx > bm2Idx {
		t.Errorf("fn1 should come before Bm2: fn1=%d bm2=%d", fn1Idx, bm2Idx)
	}
	if bm2Idx > fn2Idx {
		t.Errorf("Bm2 should come before fn2: bm2=%d fn2=%d", bm2Idx, fn2Idx)
	}
}

func TestXMLControlCharStripping(t *testing.T) {
	input := "before\x01after"
	result := stripControlChars(input)
	if strings.Contains(result, "\x01") {
		t.Error("control char 0x01 should be stripped")
	}
	if !strings.Contains(result, "before") || !strings.Contains(result, "after") {
		t.Errorf("expected 'before' and 'after' preserved, got: %q", result)
	}
}

func TestWhitespaceNormalization(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>  multiple   spaces  </w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	idx := strings.Index(x, "<p>")
	end := strings.Index(x, "</p>")
	if idx >= 0 && end > idx {
		pContent := x[idx : end+4]
		if strings.Contains(pContent, "  ") {
			t.Errorf("expected whitespace normalized in paragraph, got: %s", pContent)
		}
	}
}

func TestPageBreakBefore(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:pageBreakBefore/></w:pPr><w:r><w:t>Page</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<br type="page"/>`) {
		t.Errorf("expected <br type=\"page\"/>, got: %s", doc.WordsXML)
	}
}

func TestImageAltOnly(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r>
      <w:drawing>
        <wp:inline>
          <a:graphic>
            <a:graphicData>
              <pic:pic>
                <pic:blipFill>
                  <a:blip r:embed="rId1"/>
                </pic:blipFill>
                <pic:spPr>
                  <a:xfrm>
                    <a:ext cx="609600" cy="406400"/>
                  </a:xfrm>
                </pic:spPr>
              </pic:pic>
            </a:graphicData>
          </a:graphic>
          <wp:docPr name="Image1" descr="Test Image"/>
        </wp:inline>
      </w:drawing>
    </w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if strings.Contains(x, "src=") {
		t.Error("src= should not be emitted")
	}
	if strings.Contains(x, "width=") && strings.Contains(x, "<img") {
		t.Error("width= should not be emitted on <img>")
	}
}

func TestTrackedChangesLossless(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r>
      <w:ins w:author="Author" w:date="2024-01-01T00:00:00Z">
        <w:r><w:t>inserted text</w:t></w:r>
      </w:ins>
    </w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)

	// Semantic mode: should NOT have <ins>
	doc, err := ProcessDOCXBytesMode(data, "semantic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(doc.WordsXML, "<ins>") {
		t.Error("<ins> should not appear in semantic mode")
	}

	// Lossless mode: should have <ins>
	doc, err = ProcessDOCXBytesMode(data, "lossless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<ins>") {
		t.Error("<ins> should appear in lossless mode")
	}
}

func TestModeAttribute(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>Text</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)

	doc, err := ProcessDOCXBytesMode(data, "semantic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `mode="semantic"`) {
		t.Error("expected mode=\"semantic\"")
	}

	doc, err = ProcessDOCXBytesMode(data, "lossless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `mode="lossless"`) {
		t.Error("expected mode=\"lossless\"")
	}
}

func TestRootElement(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>Text</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `xmlns="urn:words:v1"`) {
		t.Error("expected xmlns=\"urn:words:v1\"")
	}
	if !strings.Contains(x, `xmlns:s="urn:words:v1:style"`) {
		t.Error("expected xmlns:s=\"urn:words:v1:style\"")
	}
	if !strings.Contains(x, `version="1.0.1"`) {
		t.Error("expected version=\"1.0.1\"")
	}
}

func TestStyleBlock(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>Text</w:t></w:r></w:p>
  <w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<style unit=\"in\">") {
		t.Error("expected <style unit=\"in\">")
	}
	if !strings.Contains(x, "<s:page") {
		t.Error("expected <s:page>")
	}
}

func TestMetdata(t *testing.T) {
	coreXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
                    xmlns:dc="http://purl.org/dc/elements/1.1/"
                    xmlns:dcterms="http://purl.org/dc/terms/"
                    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>My Document</dc:title>
  <dc:creator>Author Name</dc:creator>
  <dcterms:created xsi:type="dcterms:W3CDTF">2024-01-15T10:30:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2024-01-20T14:00:00Z</dcterms:modified>
  <cp:keywords>test, docx, preprocessor</cp:keywords>
</cp:coreProperties>`
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>Text</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeFullDocx(body, "", "", "", coreXML, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<meta>") {
		t.Error("expected <meta>")
	}
	if !strings.Contains(x, "<title>My Document</title>") {
		t.Error("expected <title>My Document</title>")
	}
	if !strings.Contains(x, "<author>Author Name</author>") {
		t.Error("expected <author>Author Name</author>")
	}
	if !strings.Contains(x, "<keywords>test, docx, preprocessor</keywords>") {
		t.Error("expected <keywords>")
	}
}

func TestRTL(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:rtl/></w:rPr><w:t>Arabic text</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `dir="rtl"`) {
		t.Errorf("expected dir=\"rtl\", got: %s", doc.WordsXML)
	}
}

func TestNoBreakHyphen(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:noBreakHyphen/></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "\u00AC") {
		t.Errorf("expected non-breaking hyphen character, got: %s", doc.WordsXML)
	}
}

func TestSoftHyphen(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:softHyphen/></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "\u00AD") {
		t.Errorf("expected soft hyphen character, got: %s", doc.WordsXML)
	}
}

func TestParagraphBidi(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:bidi/></w:pPr><w:r><w:t>RTL paragraph</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `dir="rtl"`) {
		t.Errorf("expected dir=\"rtl\" for bidi paragraph, got: %s", doc.WordsXML)
	}
}

func TestInvalidZip(t *testing.T) {
	_, err := ProcessDOCXBytes([]byte("not a zip"))
	if err == nil {
		t.Error("expected error for invalid zip")
	}
}

func TestMissingDocumentXML(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("word/styles.xml")
	f.Write([]byte("<styles/>"))
	w.Close()

	_, err := ProcessDOCXBytes(buf.Bytes())
	if err == nil {
		t.Error("expected error for missing document.xml")
	}
}

func TestImagePlaceholder(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r>
      <w:pict>
        <v:shape>
          <v:imagedata r:id="rId1"/>
        </v:shape>
      </w:pict>
    </w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = doc
}

func TestSDTUnwrap(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:sdt>
    <w:sdtContent>
      <w:p><w:r><w:t>SDT Content</w:t></w:r></w:p>
    </w:sdtContent>
  </w:sdt>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "SDT Content") {
		t.Error("expected SDT content to be unwrapped")
	}
}

func TestEmptyDocument(t *testing.T) {
	body := xmlHeader + `<w:body></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<words`) {
		t.Error("expected root <words> element")
	}
}

func TestProcessDOCXFile(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>File Test</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)

	tmpDir := t.TempDir()
	path := tmpDir + "/test.docx"
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	doc, err := ProcessDOCXFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "File Test") {
		t.Error("expected 'File Test' in output")
	}
}

func TestProcessDOCXFileMode(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>Mode Test</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)

	tmpDir := t.TempDir()
	path := tmpDir + "/test.docx"
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	doc, err := ProcessDOCXFileMode(path, "lossless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `mode="lossless"`) {
		t.Error("expected mode=\"lossless\"")
	}
}

func TestTableStyleName(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="table" w:styleId="CustomTable">
    <w:name w:val="Custom Table Style"/>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr><w:tblStyle w:val="CustomTable"/></w:tblPr>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `c="Custom Table Style"`) {
		t.Errorf("expected c=\"Custom Table Style\", got: %s", doc.WordsXML)
	}
}

func TestHiddenText(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:vanish/></w:rPr><w:t>hidden</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `hidden="true"`) {
		t.Errorf("expected hidden=\"true\", got: %s", doc.WordsXML)
	}
}

func TestLanguage(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:lang w:val="id-ID"/></w:rPr><w:t>Bahasa</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `lang="id-ID"`) {
		t.Errorf("expected lang=\"id-ID\", got: %s", doc.WordsXML)
	}
}

func TestXMLEscaping(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t>1 &lt; 2 &amp; 3 &gt; 1</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "&lt;") {
		t.Error("expected &lt; in output")
	}
	if !strings.Contains(doc.WordsXML, "&amp;") {
		t.Error("expected &amp; in output")
	}
	if !strings.Contains(doc.WordsXML, "&gt;") {
		t.Error("expected &gt; in output")
	}
}

func TestPagePreset(t *testing.T) {
	presets := []struct {
		name string
		w, h int
	}{
		{"A4", 595, 842},
		{"Letter", 612, 792},
		{"Legal", 612, 1008},
		{"A3", 842, 1191},
	}
	for _, p := range presets {
		result := matchPageSize(p.w, p.h)
		if result != p.name {
			t.Errorf("matchPageSize(%d, %d) = %q, want %q", p.w, p.h, result, p.name)
		}
		// Test rotation
		result = matchPageSize(p.h, p.w)
		if result != p.name {
			t.Errorf("matchPageSize(%d, %d) = %q, want %q (rotated)", p.h, p.w, result, p.name)
		}
	}

	if matchPageSize(100, 200) != "" {
		t.Error("unknown size should return empty")
	}
}

func TestIsMonospaceFont(t *testing.T) {
	monospace := []string{"Courier New", "Consolas", "Menlo", "Monaco", "monospace", "Courier", "Lucida Console"}
	for _, f := range monospace {
		if !isMonospaceFont(f) {
			t.Errorf("isMonospaceFont(%q) = false, want true", f)
		}
	}
	nonMonospace := []string{"Arial", "Times New Roman", "Helvetica", "Calibri"}
	for _, f := range nonMonospace {
		if isMonospaceFont(f) {
			t.Errorf("isMonospaceFont(%q) = true, want false", f)
		}
	}
}

func TestWhitespacePreserve(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p><w:r><w:t xml:space="preserve">  two  spaces  </w:t></w:r></w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "  two  spaces  ") {
		t.Errorf("expected preserved whitespace, got: %s", doc.WordsXML)
	}
}

func TestTableStyleProperties(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="table" w:styleId="BorderedTable">
    <w:name w:val="Bordered Table"/>
    <w:tblPr>
      <w:tblBorders>
        <w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/>
      </w:tblBorders>
      <w:tblCellSpacing w:w="100"/>
      <w:tblW w:w="5000"/>
    </w:tblPr>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr><w:tblStyle w:val="BorderedTable"/></w:tblPr>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<s:custom`) {
		t.Error("expected <s:custom> for BorderedTable style")
	}
	if !strings.Contains(x, `borderWidth=`) {
		t.Error("expected borderWidth in s:custom")
	}
	if !strings.Contains(x, `borderColor=`) {
		t.Error("expected borderColor in s:custom")
	}
	if !strings.Contains(x, `borderStyle=`) {
		t.Error("expected borderStyle in s:custom")
	}
}

func TestColWidths(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid>
      <w:gridCol w:w="3000"/>
      <w:gridCol w:w="4000"/>
    </w:tblGrid>
    <w:tr>
      <w:tc><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc>
      <w:tc><w:p><w:r><w:t>B</w:t></w:r></w:p></w:tc>
    </w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<s:col ref="1"`) {
		t.Error("expected <s:col ref=\"1\">")
	}
	if strings.Count(x, `<s:col ref="1"`) != 2 {
		t.Errorf("expected 2 s:col elements for table 1, got: %s", x)
	}
}

func TestAllCaps(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:caps/></w:rPr><w:t>UPPERCASE</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<uppercase>UPPERCASE</uppercase>") {
		t.Errorf("expected <uppercase>, got: %s", doc.WordsXML)
	}
}

func TestStrikeAndDoubleStrike(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:strike/></w:rPr><w:t>Strike</w:t></w:r>
    <w:r><w:rPr><w:dstrike/></w:rPr><w:t>DoubleStrike</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<s>Strike</s>") {
		t.Error("expected <s>Strike</s>")
	}
	if !strings.Contains(x, "<s>DoubleStrike</s>") {
		t.Error("expected <s>DoubleStrike</s>")
	}
}

func TestHighlight(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:highlight w:val="yellow"/></w:rPr><w:t>Highlighted</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `highlight="yellow"`) {
		t.Errorf("expected highlight=\"yellow\", got: %s", doc.WordsXML)
	}
}

func TestNestedTable(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc>
      <w:tbl>
        <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
        <w:tr><w:tc><w:p><w:r><w:t>Nested</w:t></w:r></w:p></w:tc></w:tr>
      </w:tbl>
    </w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<table id="1"`) {
		t.Error("expected outer table id=1")
	}
	if !strings.Contains(x, `<table id="2"`) {
		t.Error("expected nested table id=2")
	}
}

func TestColSpan(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid>
      <w:gridCol w:w="2500"/>
      <w:gridCol w:w="2500"/>
    </w:tblGrid>
    <w:tr>
      <w:tc><w:tcPr><w:gridSpan w:val="2"/></w:tcPr><w:p><w:r><w:t>Spanned</w:t></w:r></w:p></w:tc>
    </w:tr>
    <w:tr>
      <w:tc><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc>
      <w:tc><w:p><w:r><w:t>B</w:t></w:r></w:p></w:tc>
    </w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `colspan="2"`) {
		t.Errorf("expected colspan=\"2\", got: %s", doc.WordsXML)
	}
}

func TestHeaderRow(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:trPr><w:tblHeader/></w:trPr>
      <w:tc><w:p><w:r><w:t>Header</w:t></w:r></w:p></w:tc>
    </w:tr>
    <w:tr>
      <w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc>
    </w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "<th>") {
		t.Error("expected <th> for header row")
	}
	if !strings.Contains(x, "<td>") {
		t.Error("expected <td> for data row")
	}
}

func TestCellValign(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc>
      <w:tcPr><w:vAlign w:val="center"/></w:tcPr>
      <w:p><w:r><w:t>Centered</w:t></w:r></w:p>
    </w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `valign="center"`) {
		t.Errorf("expected valign=\"center\", got: %s", doc.WordsXML)
	}
}

func TestCellNoWrap(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc>
      <w:tcPr><w:noWrap/></w:tcPr>
      <w:p><w:r><w:t>NoWrap</w:t></w:r></w:p>
    </w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `noWrap="true"`) {
		t.Errorf("expected noWrap=\"true\", got: %s", doc.WordsXML)
	}
}

func TestTableCaptionAndSummary(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr>
      <w:tblCaption w:val="Table Caption"/>
      <w:tblDescription w:val="Table Summary"/>
    </w:tblPr>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `caption="Table Caption"`) {
		t.Error("expected caption attribute")
	}
	if !strings.Contains(x, `summary="Table Summary"`) {
		t.Error("expected summary attribute")
	}
}

func TestTableAlignment(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr><w:jc w:val="center"/></w:tblPr>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `align="center"`) {
		t.Errorf("expected align=\"center\", got: %s", doc.WordsXML)
	}
}

func TestBreakTypes(t *testing.T) {
	breakTypes := []struct {
		attr, expected string
	}{
		{`w:type="page"`, `<br type="page"/>`},
		{`w:type="column"`, `<br type="column"/>`},
		{`w:type="clear"`, `<br type="clear"/>`},
	}
	for _, tc := range breakTypes {
		body := xmlHeader + `<w:body>
  <w:p><w:r><w:br ` + tc.attr + `/></w:r></w:p>
</w:body></w:document>`
		data := makeMinimalDocx(body)
		doc, err := ProcessDOCXBytes(data)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.attr, err)
		}
		if !strings.Contains(doc.WordsXML, tc.expected) {
			t.Errorf("expected %s, got: %s", tc.expected, doc.WordsXML)
		}
	}
}

const stylesXMLHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`

func TestCodeBlockDetection(t *testing.T) {
	tests := []struct {
		name         string
		styles       string
		body         string
		expected     string
		notExpected  string
	}{
		{
			name: "known code style exact match",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="Code">
				<w:name w:val="Code"/>
			</w:style></w:styles>`,
			body:     `<w:p><w:pPr><w:pStyle w:val="Code"/></w:pPr><w:r><w:t>hello</w:t></w:r></w:p>`,
			expected: "<pre>hello</pre>",
		},
		{
			name: "code with word boundary + monospace = code",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="MyCodeBlock">
				<w:name w:val="My Code Block"/>
			</w:style></w:styles>`,
			body: `<w:p>
				<w:pPr><w:pStyle w:val="MyCodeBlock"/></w:pPr>
				<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/></w:rPr><w:t>code here</w:t></w:r>
			</w:p>`,
			expected: "<pre>code here</pre>",
		},
		{
			name: "code word without monospace = NOT code",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="MyCodeBlock">
				<w:name w:val="My Code Block"/>
			</w:style></w:styles>`,
			body:        `<w:p><w:pPr><w:pStyle w:val="MyCodeBlock"/></w:pPr><w:r><w:t>code here</w:t></w:r></w:p>`,
			notExpected: "<pre>",
		},
		{
			name: "example style is NOT code",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="ExampleHeading">
				<w:name w:val="Example Heading"/>
			</w:style></w:styles>`,
			body:        `<w:p><w:pPr><w:pStyle w:val="ExampleHeading"/></w:pPr><w:r><w:t>not code</w:t></w:r></w:p>`,
			notExpected: "<pre>",
		},
		{
			name: "output summary style without monospace is NOT code",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="OutputSummary">
				<w:name w:val="Output Summary"/>
			</w:style></w:styles>`,
			body:        `<w:p><w:pPr><w:pStyle w:val="OutputSummary"/></w:pPr><w:r><w:t>not code</w:t></w:r></w:p>`,
			notExpected: "<pre>",
		},
		{
			name:   "heading with monospace font stays heading",
			styles: ``,
			body: `<w:p>
				<w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
				<w:r><w:rPr><w:rFonts w:ascii="Courier New" w:hAnsi="Courier New"/></w:rPr><w:t>title</w:t></w:r>
			</w:p>`,
			expected: "<h1",
		},
		{
			name:   "title with monospace font stays title",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/></w:style></w:styles>`,
			body: `<w:p>
				<w:pPr><w:pStyle w:val="Title"/></w:pPr>
				<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/></w:rPr><w:t>title</w:t></w:r>
			</w:p>`,
			expected: "<h1",
		},
		{
			name:   "list with monospace font stays list",
			styles: ``,
			body: `<w:p>
				<w:pPr>
					<w:pStyle w:val="ListParagraph"/>
					<w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr>
				</w:pPr>
				<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/></w:rPr><w:t>item</w:t></w:r>
			</w:p>`,
			expected: "<li",
		},
		{
			name:   "quote with monospace font stays quote",
			styles: stylesXMLHeader + `<w:style w:type="paragraph" w:styleId="Quote"><w:name w:val="Quote"/></w:style></w:styles>`,
			body: `<w:p>
				<w:pPr><w:pStyle w:val="Quote"/></w:pPr>
				<w:r><w:rPr><w:rFonts w:ascii="Courier New" w:hAnsi="Courier New"/></w:rPr><w:t>quoted</w:t></w:r>
			</w:p>`,
			expected: "<blockquote>",
		},
		{
			name:   "monospace font alone on normal paragraph is NOT code",
			styles: ``,
			body: `<w:p>
				<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/></w:rPr><w:t>fmt.Println("hello")</w:t></w:r>
			</w:p>`,
			notExpected: "<pre>",
		},
		{
			name:   "mixed monospace and non-monospace stays paragraph",
			styles: ``,
			body: `<w:p>
				<w:r><w:rPr><w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/></w:rPr><w:t>code</w:t></w:r>
				<w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/></w:rPr><w:t>text</w:t></w:r>
			</w:p>`,
			notExpected: "<pre>",
		},
	}


	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := xmlHeader + `<w:body>` + tc.body + `</w:body></w:document>`
			data := makeDocxWithParts(body, tc.styles, "", "")
			doc, err := ProcessDOCXBytes(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.expected != "" && !strings.Contains(doc.WordsXML, tc.expected) {
				t.Errorf("expected output to contain %q, got:\n%s", tc.expected, doc.WordsXML)
			}
			if tc.notExpected != "" && strings.Contains(doc.WordsXML, tc.notExpected) {
				t.Errorf("expected output to NOT contain %q, got:\n%s", tc.notExpected, doc.WordsXML)
			}
		})
	}
}

func TestHasWordBoundary(t *testing.T) {
	tests := []struct {
		s, word string
		want    bool
	}{
		{"code block", "code", true},
		{"mycode", "code", false},
		{"encoding", "code", false},
		{"source code", "source", true},
		{"sourcelanguage", "source", false},
		{"output log", "output", true},
		{"outputlog", "output", false},
		{"code-source", "code", true},
		{"code_source", "code", true},
	}
	for _, tc := range tests {
		got := hasWordBoundary(tc.s, tc.word)
		if got != tc.want {
			t.Errorf("hasWordBoundary(%q, %q) = %v, want %v", tc.s, tc.word, got, tc.want)
		}
	}
}

func TestBuildStyleMapThemeFonts(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault><w:rPr>
      <w:rFonts w:asciiTheme="minorHAnsi" w:hAnsiTheme="minorHAnsi" w:eastAsiaTheme="minorEastAsia" w:cstheme="minorBidi"/>
      <w:sz w:val="22"/>
      <w:color w:val="333333"/>
    </w:rPr></w:rPrDefault>
  </w:docDefaults>
  <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`
	themeMap := map[string]string{"minorHAnsi": "Calibri", "minorEastAsia": "MS Mincho", "minorBidi": "Arial"}
	m, _, def := buildStyleMap([]byte(styles), themeMap)
	if def.Family != "Calibri" {
		t.Errorf("expected default font Calibri from theme, got %q", def.Family)
	}
	if def.FontEA != "MS Mincho" {
		t.Errorf("expected fontEA MS Mincho from theme, got %q", def.FontEA)
	}
	if def.FontCS != "Arial" {
		t.Errorf("expected fontCS Arial from theme, got %q", def.FontCS)
	}
	if def.SizePt != 11 {
		t.Errorf("expected size 11pt, got %v", def.SizePt)
	}
	if def.Color != "333333" {
		t.Errorf("expected color 333333, got %q", def.Color)
	}
	if _, ok := m["Normal"]; !ok {
		t.Error("expected Normal style in map")
	}
}

func TestBuildStyleMapHAnsiFallback(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault><w:rPr>
      <w:rFonts w:hAnsi="Arial"/>
    </w:rPr></w:rPrDefault>
  </w:docDefaults>
  <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`
	_, _, def := buildStyleMap([]byte(styles), nil)
	if def.Family != "Arial" {
		t.Errorf("expected Arial from HAnsi fallback, got %q", def.Family)
	}
}

func TestBuildStyleMapStyleFontHAnsi(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Custom">
    <w:name w:val="Custom"/>
    <w:rPr><w:rFonts w:hAnsi="Verdana"/></w:rPr>
  </w:style>
</w:styles>`
	m, _, _ := buildStyleMap([]byte(styles), nil)
	sd := m["Custom"]
	if sd.Family != "Verdana" {
		t.Errorf("expected Verdana from HAnsi, got %q", sd.Family)
	}
}

func TestBuildStyleMapStyleThemeFonts(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Themed">
    <w:name w:val="Themed"/>
    <w:rPr>
      <w:rFonts w:asciiTheme="majorHAnsi" w:eastAsiaTheme="majorEastAsia" w:cstheme="majorBidi"/>
      <w:szCs w:val="28"/>
      <w:u w:val="single"/>
      <w:strike/>
      <w:smallCaps/>
      <w:caps/>
    </w:rPr>
  </w:style>
</w:styles>`
	themeMap := map[string]string{"majorHAnsi": "Cambria", "majorEastAsia": "SimSun", "majorBidi": "Times New Roman"}
	m, _, _ := buildStyleMap([]byte(styles), themeMap)
	sd := m["Themed"]
	if sd.Family != "Cambria" {
		t.Errorf("expected Cambria from theme, got %q", sd.Family)
	}
	if sd.FontEA != "SimSun" {
		t.Errorf("expected SimSun from EA theme, got %q", sd.FontEA)
	}
	if sd.FontCS != "Times New Roman" {
		t.Errorf("expected Times New Roman from CS theme, got %q", sd.FontCS)
	}
	if sd.SizeCS != 14 {
		t.Errorf("expected sizeCS 14, got %v", sd.SizeCS)
	}
	if !sd.Strikethrough {
		t.Error("expected strikethrough")
	}
	if !sd.SmallCaps {
		t.Error("expected smallCaps")
	}
	if !sd.Uppercase {
		t.Error("expected uppercase")
	}
	if sd.Underline != "single" {
		t.Errorf("expected underline single, got %q", sd.Underline)
	}
}

func TestBuildStyleMapParaProps(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Full">
    <w:name w:val="Full"/>
    <w:pPr>
      <w:outlineLvl w:val="2"/>
      <w:jc w:val="center"/>
      <w:spacing w:before="200" w:after="100" w:line="480" w:lineRule="auto"/>
      <w:ind w:left="720" w:right="360" w:firstLine="480" w:hanging="240"/>
      <w:tabs>
        <w:tab w:val="right" w:pos="7200" w:leader="dot"/>
        <w:tab w:pos="3600"/>
      </w:tabs>
    </w:pPr>
  </w:style>
</w:styles>`
	m, _, _ := buildStyleMap([]byte(styles), nil)
	sd := m["Full"]
	if sd.HeadingLevel != 3 {
		t.Errorf("expected HeadingLevel 3 (outlineLvl+1), got %d", sd.HeadingLevel)
	}
	if sd.Align != "center" {
		t.Errorf("expected align center, got %q", sd.Align)
	}
	if sd.SpacingBefore != 10 {
		t.Errorf("expected spacingBefore 10, got %v", sd.SpacingBefore)
	}
	if sd.SpacingAfter != 5 {
		t.Errorf("expected spacingAfter 5, got %v", sd.SpacingAfter)
	}
	if sd.LineSpacing != 2 {
		t.Errorf("expected lineSpacing 2 (480/240), got %v", sd.LineSpacing)
	}
	if sd.LineRule != "auto" {
		t.Errorf("expected lineRule auto, got %q", sd.LineRule)
	}
	if sd.IndentLeft != 0.5 {
		t.Errorf("expected indentLeft 0.5, got %v", sd.IndentLeft)
	}
	if sd.IndentRight != 0.25 {
		t.Errorf("expected indentRight 0.25, got %v", sd.IndentRight)
	}
	if len(sd.Tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(sd.Tabs))
	}
	if sd.Tabs[0].Align != "right" || sd.Tabs[0].Leader != "dot" {
		t.Errorf("expected tab right+dot, got align=%s leader=%s", sd.Tabs[0].Align, sd.Tabs[0].Leader)
	}
	if sd.Tabs[1].Align != "left" || sd.Tabs[1].Leader != "none" {
		t.Errorf("expected tab left+none default, got align=%s leader=%s", sd.Tabs[1].Align, sd.Tabs[1].Leader)
	}
}

func TestBuildStyleMapLineRuleExact(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Exact">
    <w:name w:val="Exact"/>
    <w:pPr><w:spacing w:line="360" w:lineRule="exact"/></w:pPr>
  </w:style>
</w:styles>`
	m, _, _ := buildStyleMap([]byte(styles), nil)
	sd := m["Exact"]
	if sd.LineRule != "exact" {
		t.Errorf("expected lineRule exact, got %q", sd.LineRule)
	}
	if sd.LineSpacing != 18 {
		t.Errorf("expected lineSpacing 18 (360/20), got %v", sd.LineSpacing)
	}
}

func TestBuildStyleMapTableProps(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="table" w:styleId="TblStyle">
    <w:name w:val="TblStyle"/>
    <w:tblPr>
      <w:tblBorders><w:top w:val="single" w:sz="8" w:space="0" w:color="FF0000"/></w:tblBorders>
      <w:tblCellSpacing w:w="144"/>
      <w:tblW w:w="5000" w:type="dxa"/>
    </w:tblPr>
  </w:style>
</w:styles>`
	m, _, _ := buildStyleMap([]byte(styles), nil)
	sd := m["TblStyle"]
	if sd.BorderWidth <= 0 {
		t.Errorf("expected borderWidth > 0, got %v", sd.BorderWidth)
	}
	if sd.BorderColor != "FF0000" {
		t.Errorf("expected borderColor FF0000, got %q", sd.BorderColor)
	}
	if sd.BorderStyle != "single" {
		t.Errorf("expected borderStyle single, got %q", sd.BorderStyle)
	}
	if sd.CellSpacing <= 0 {
		t.Errorf("expected cellSpacing > 0, got %v", sd.CellSpacing)
	}
	if sd.Width <= 0 {
		t.Errorf("expected width > 0, got %v", sd.Width)
	}
}

func TestBuildStyleMapBasedOn(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Parent"><w:name w:val="Parent"/></w:style>
  <w:style w:type="paragraph" w:styleId="Child">
    <w:name w:val="Child"/>
    <w:basedOn w:val="Parent"/>
  </w:style>
</w:styles>`
	m, _, _ := buildStyleMap([]byte(styles), nil)
	sd := m["Child"]
	if sd.BasedOn != "Parent" {
		t.Errorf("expected basedOn Parent, got %q", sd.BasedOn)
	}
}

func TestBuildStyleMapInvalidXML(t *testing.T) {
	_, _, def := buildStyleMap([]byte("not xml at all"), nil)
	if def.Family != "Times New Roman" {
		t.Errorf("expected default fallback font, got %q", def.Family)
	}
}

func TestBuildStyleMapNoRPr(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault><w:rPr/></w:rPrDefault>
  </w:docDefaults>
  <w:style w:type="paragraph" w:styleId="Plain"><w:name w:val="Plain"/></w:style>
</w:styles>`
	m, _, def := buildStyleMap([]byte(styles), nil)
	if def.Family != "Times New Roman" {
		t.Errorf("expected default font, got %q", def.Family)
	}
	sd := m["Plain"]
	if sd.Name != "Plain" {
		t.Errorf("expected name Plain, got %q", sd.Name)
	}
}

func TestInferHeadingLevel(t *testing.T) {
	tests := []struct {
		id    string
		level int
	}{
		{"Heading1", 1}, {"Heading5", 5}, {"Heading9", 9},
		{"heading3", 3}, {"Title", 1}, {"Normal", 0},
		{"Heading0", 0}, {"Heading10", 0}, {"Heading", 1},
	}
	for _, tc := range tests {
		got := inferHeadingLevel(tc.id)
		if got != tc.level {
			t.Errorf("inferHeadingLevel(%q) = %d, want %d", tc.id, got, tc.level)
		}
	}
}

func TestProcessDOCXFileNonexistent(t *testing.T) {
	_, err := ProcessDOCXFile("/nonexistent/file.docx")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestProcessDOCXFileModeNonexistent(t *testing.T) {
	_, err := ProcessDOCXFileMode("/nonexistent/file.docx", "lossless")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExtractThemeValid(t *testing.T) {
	themeXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Test">
  <a:themeElements>
    <a:fontScheme name="Test">
      <a:majorFont>
        <a:latin typeface="Cambria"/>
        <a:ea typeface="SimSun"/>
        <a:cs typeface="Times New Roman"/>
      </a:majorFont>
      <a:minorFont>
        <a:latin typeface="Calibri"/>
        <a:ea typeface="MS Mincho"/>
        <a:cs typeface="Arial"/>
      </a:minorFont>
    </a:fontScheme>
    <a:clrScheme name="Test">
      <a:dk1><a:srgbClr val="000000"/></a:dk1>
      <a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="44546A"/></a:dk2>
      <a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
      <a:accent1><a:srgbClr val="4472C4"/></a:accent1>
      <a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
      <a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
      <a:accent4><a:srgbClr val="FFC000"/></a:accent4>
      <a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>
      <a:accent6><a:srgbClr val="70AD47"/></a:accent6>
      <a:hlink><a:srgbClr val="0563C1"/></a:hlink>
      <a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
    </a:clrScheme>
  </a:themeElements>
</a:theme>`)
	td := extractTheme(themeXML)
	if td == nil {
		t.Fatal("expected non-nil theme data")
	}
	if td.Font != "Calibri" {
		t.Errorf("expected Font Calibri, got %q", td.Font)
	}
	if td.FontEA != "MS Mincho" {
		t.Errorf("expected FontEA MS Mincho, got %q", td.FontEA)
	}
	if td.FontCS != "Arial" {
		t.Errorf("expected FontCS Arial, got %q", td.FontCS)
	}
	if td.Fg != "000000" {
		t.Errorf("expected Fg 000000, got %q", td.Fg)
	}
	if td.Bg != "FFFFFF" {
		t.Errorf("expected Bg FFFFFFF, got %q", td.Bg)
	}
	if td.FontMap["minorHAnsi"] != "Calibri" {
		t.Errorf("expected minorHAnsi Calibri, got %q", td.FontMap["minorHAnsi"])
	}
	if td.FontMap["minorEastAsia"] != "MS Mincho" {
		t.Errorf("expected minorEastAsia MS Mincho, got %q", td.FontMap["minorEastAsia"])
	}
	if td.FontMap["minorBidi"] != "Arial" {
		t.Errorf("expected minorBidi Arial, got %q", td.FontMap["minorBidi"])
	}
	if td.FontMap["majorHAnsi"] != "Cambria" {
		t.Errorf("expected majorHAnsi Cambria, got %q", td.FontMap["majorHAnsi"])
	}
	if td.FontMap["majorEastAsia"] != "SimSun" {
		t.Errorf("expected majorEastAsia SimSun, got %q", td.FontMap["majorEastAsia"])
	}
	if td.FontMap["majorBidi"] != "Times New Roman" {
		t.Errorf("expected majorBidi Times New Roman, got %q", td.FontMap["majorBidi"])
	}
}

func TestExtractThemeInvalid(t *testing.T) {
	td := extractTheme([]byte("not xml"))
	if td != nil {
		t.Error("expected nil for invalid theme XML")
	}
}

func TestExtractThemeEmptyFont(t *testing.T) {
	themeXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Test">
  <a:themeElements>
    <a:fontScheme name="Test">
      <a:majorFont><a:latin typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:clrScheme name="Test">
      <a:dk1><a:srgbClr val="000000"/></a:dk1>
      <a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="44546A"/></a:dk2>
      <a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
      <a:accent1><a:srgbClr val="4472C4"/></a:accent1>
      <a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
      <a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
      <a:accent4><a:srgbClr val="FFC000"/></a:accent4>
      <a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>
      <a:accent6><a:srgbClr val="70AD47"/></a:accent6>
      <a:hlink><a:srgbClr val="0563C1"/></a:hlink>
      <a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
    </a:clrScheme>
  </a:themeElements>
</a:theme>`)
	td := extractTheme(themeXML)
	if td == nil {
		t.Fatal("expected non-nil theme data from colors")
	}
	if td.Fg != "000000" {
		t.Errorf("expected Fg, got %q", td.Fg)
	}
}

func TestExtractHyperlinkURL(t *testing.T) {
	tests := []struct {
		input string
		url   string
		ok    bool
	}{
		{`HYPERLINK "https://example.com"`, "https://example.com", true},
		{`HYPERLINK https://example.com`, "https://example.com", true},
		{`HYPERLINK "mailto:test@test.com"`, "mailto:test@test.com", true},
		{`PAGE`, "", false},
		{``, "", false},
	}
	for _, tc := range tests {
		url, ok := extractHyperlinkURL(tc.input)
		if ok != tc.ok || url != tc.url {
			t.Errorf("extractHyperlinkURL(%q) = (%q, %v), want (%q, %v)", tc.input, url, ok, tc.url, tc.ok)
		}
	}
}

func TestFormatBorderStyle(t *testing.T) {
	tests := []struct {
		input string
		output string
	}{
		{"single", "s"}, {"double", "d"}, {"dashed", "ds"}, {"dotted", "dt"}, {"none", "n"},
		{"thick", "s"}, {"unknown", "s"},
	}
	for _, tc := range tests {
		got := formatBorderStyle(tc.input)
		if got != tc.output {
			t.Errorf("formatBorderStyle(%q) = %q, want %q", tc.input, got, tc.output)
		}
	}
}

func TestResolveHeadingLevel(t *testing.T) {
	sm := map[string]StyleDef{
		"Normal":   {HeadingLevel: 0},
		"Heading1": {HeadingLevel: 1},
		"Quote":    {HeadingLevel: 0},
	}
	tests := []struct {
		styleID string
		level   int
	}{
		{"Heading1", 1},
		{"Normal", 0},
		{"Quote", 0},
		{"UnknownStyle", 0},
		{"", 0},
	}
	for _, tc := range tests {
		got := resolveHeadingLevel(tc.styleID, tc.styleID, sm)
		if got != tc.level {
			t.Errorf("resolveHeadingLevel(%q) = %d, want %d", tc.styleID, got, tc.level)
		}
	}
}

func TestListStartOverride(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
  <w:num w:numId="2">
    <w:abstractNumId w:val="0"/>
    <w:lvlOverride w:ilvl="0"><w:startOverride w:val="5"/></w:lvlOverride>
  </w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="2"/></w:numPr></w:pPr><w:r><w:t>Restart at 5</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="2"/></w:numPr></w:pPr><w:r><w:t>Continue</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `start="5"`) {
		t.Errorf("expected start=\"5\", got: %s", doc.WordsXML)
	}
}

func TestParagraphLineSpacingExact(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Spacing">
    <w:name w:val="Spacing"/>
    <w:pPr><w:spacing w:line="360" w:lineRule="exact"/></w:pPr>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:pStyle w:val="Spacing"/></w:pPr>
    <w:r><w:t>Exact spacing</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "lineSpacing") {
		t.Errorf("expected lineSpacing in output, got: %s", x)
	}
	if !strings.Contains(x, "lineRule") {
		t.Errorf("expected lineRule in output, got: %s", x)
	}
}

func TestParagraphAlignBoth(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Justified">
    <w:name w:val="Justified"/>
    <w:pPr><w:jc w:val="both"/></w:pPr>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:pStyle w:val="Justified"/></w:pPr>
    <w:r><w:t>Justified</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `alignment="both"`) {
		t.Errorf("expected alignment=\"both\" in custom style, got: %s", x)
	}
}

func TestParagraphIndent(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Indented">
    <w:name w:val="Indented"/>
    <w:pPr><w:ind w:left="720" w:right="360" w:firstLine="480" w:hanging="240"/></w:pPr>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:pStyle w:val="Indented"/></w:pPr>
    <w:r><w:t>Indented</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "indentLeft") {
		t.Errorf("expected indentLeft in custom style, got: %s", x)
	}
	if !strings.Contains(x, "indentRight") {
		t.Errorf("expected indentRight in custom style, got: %s", x)
	}
}

func TestSuperscriptSubscript(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:rPr><w:vertAlign w:val="superscript"/></w:rPr><w:t>sup</w:t></w:r>
    <w:r><w:rPr><w:vertAlign w:val="subscript"/></w:rPr><w:t>sub</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<sup>") {
		t.Errorf("expected <sup>, got: %s", doc.WordsXML)
	}
	if !strings.Contains(doc.WordsXML, "<sub>") {
		t.Errorf("expected <sub>, got: %s", doc.WordsXML)
	}
}

func TestParagraphLang(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/></w:rPr><w:lang w:val="en-US"/></w:pPr>
    <w:r><w:t>English</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `lang="en-US"`) {
		t.Errorf("expected lang attr, got: %s", doc.WordsXML)
	}
}

func TestTableWidthAndIndent(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblPr>
      <w:tblW w:w="5000" w:type="dxa"/>
      <w:tblInd w:w="360"/>
      <w:tblCellSpacing w:w="144"/>
    </w:tblPr>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc><w:p><w:r><w:t>Data</w:t></w:r></w:p></w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, "width=") {
		t.Errorf("expected width attr, got: %s", x)
	}
	if !strings.Contains(x, "indent=") {
		t.Errorf("expected indent attr, got: %s", x)
	}
	if !strings.Contains(x, "cellSpacing=") {
		t.Errorf("expected cellSpacing attr, got: %s", x)
	}
}

func TestRTLDir(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr><w:bidi/></w:pPr>
    <w:r><w:t>RTL text</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `dir="rtl"`) {
		t.Errorf("expected dir=\"rtl\", got: %s", doc.WordsXML)
	}
}

func TestComment(t *testing.T) {
	comments := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:comments xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:comment w:id="0" w:author="John" w:date="2024-01-01T00:00:00Z">
    <w:p><w:r><w:t>This is a comment</w:t></w:r></w:p>
  </w:comment>
</w:comments>`
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t>Text</w:t></w:r>
    <w:r><w:commentReference w:id="0"/></w:r>
  </w:p>
</w:body></w:document>`
	data := makeDocxWithComments(body, comments)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<comment id="0" author="John"`) {
		t.Errorf("expected comment in notes, got: %s", doc.WordsXML)
	}
}

func TestTrackedChangesSemantic(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:ins w:id="1" w:author="Author" w:date="2024-01-01T00:00:00Z">
      <w:r><w:t>inserted text</w:t></w:r>
    </w:ins>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytesMode(data, "semantic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(doc.WordsXML, "<ins>") {
		t.Errorf("expected no <ins> in semantic mode, got: %s", doc.WordsXML)
	}
}

func TestTrackedChangesLossless2(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r><w:t>inserted text</w:t></w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytesMode(data, "lossless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "mode=\"lossless\"") {
		t.Errorf("expected lossless mode, got: %s", doc.WordsXML)
	}
}

func TestImageWidthHeight(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r>
      <w:pict>
        <v:shape type="#_x0000_t75" style="width:100pt;height:50pt">
          <v:imagedata r:id="rId1"/>
        </v:shape>
      </w:pict>
    </w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "<img") {
		t.Errorf("expected <img>, got: %s", doc.WordsXML)
	}
}

func TestNestedTableIDs(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:tbl>
    <w:tblGrid><w:gridCol w:w="5000"/></w:tblGrid>
    <w:tr><w:tc>
      <w:tbl>
        <w:tblGrid><w:gridCol w:w="3000"/></w:tblGrid>
        <w:tr><w:tc><w:p><w:r><w:t>Nested</w:t></w:r></w:p></w:tc></w:tr>
      </w:tbl>
    </w:tc></w:tr>
  </w:tbl>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `id="1"`) {
		t.Errorf("expected outer table id=1, got: %s", doc.WordsXML)
	}
	if !strings.Contains(doc.WordsXML, `id="2"`) {
		t.Errorf("expected nested table id=2, got: %s", doc.WordsXML)
	}
}

func TestSdtUnwrap(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:sdt>
    <w:sdtContent>
      <w:p><w:r><w:t>SDT content</w:t></w:r></w:p>
    </w:sdtContent>
  </w:sdt>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, "SDT content") {
		t.Errorf("expected SDT content, got: %s", doc.WordsXML)
	}
	if strings.Contains(doc.WordsXML, "<sdt>") || strings.Contains(doc.WordsXML, "<sdt>") {
		t.Errorf("expected no <sdt> in output, got: %s", doc.WordsXML)
	}
}

func TestEmitStyleBlockNormalIndent(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Normal" w:default="1">
    <w:name w:val="Normal"/>
    <w:pPr>
      <w:ind w:left="720" w:right="360" w:firstLine="480" w:hanging="240"/>
      <w:jc w:val="both"/>
      <w:spacing w:before="200" w:after="100" w:line="480" w:lineRule="auto"/>
    </w:pPr>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body><w:p><w:r><w:t>x</w:t></w:r></w:p></w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<s:indent el="p"`) {
		t.Errorf("expected s:indent, got: %s", x)
	}
	if !strings.Contains(x, `<s:align`) || !strings.Contains(x, `value="both"`) {
		t.Errorf("expected s:align value=both, got: %s", x)
	}
	if !strings.Contains(x, `<s:line el="p"`) {
		t.Errorf("expected s:line, got: %s", x)
	}
	if !strings.Contains(x, `<s:gap el="p"`) {
		t.Errorf("expected s:gap, got: %s", x)
	}
}

func TestEmitStyleBlockHeadingIndent(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Heading1">
    <w:name w:val="heading 1"/>
    <w:pPr>
      <w:ind w:left="360"/>
      <w:jc w:val="center"/>
      <w:spacing w:before="400" w:after="200" w:line="360" w:lineRule="exact"/>
      <w:outlineLvl w:val="0"/>
    </w:pPr>
  </w:style>
</w:styles>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Title</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, styles, "", "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<s:indent`) || !strings.Contains(x, `c="Heading1"`) {
		t.Errorf("expected heading indent, got: %s", x)
	}
	if !strings.Contains(x, `<s:align`) || !strings.Contains(x, `c="Heading1"`) {
		t.Errorf("expected heading align, got: %s", x)
	}
	if !strings.Contains(x, `<s:line`) || !strings.Contains(x, `c="Heading1"`) {
		t.Errorf("expected heading line, got: %s", x)
	}
	if !strings.Contains(x, `<s:gap`) || !strings.Contains(x, `c="Heading1"`) {
		t.Errorf("expected heading gap, got: %s", x)
	}
}

func TestRunThemeFonts(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/>
    <w:rPr><w:rFonts w:asciiTheme="minorHAnsi" w:hAnsiTheme="minorHAnsi"/></w:rPr>
  </w:style>
</w:styles>`
	themeXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Test">
  <a:themeElements>
    <a:fontScheme name="Test">
      <a:minorFont><a:latin typeface="Calibri"/></a:minorFont>
      <a:majorFont><a:latin typeface="Cambria"/></a:majorFont>
    </a:fontScheme>
    <a:clrScheme name="Test">
      <a:dk1><a:srgbClr val="000000"/></a:dk1>
      <a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="44546A"/></a:dk2>
      <a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
      <a:accent1><a:srgbClr val="4472C4"/></a:accent1>
      <a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
      <a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
      <a:accent4><a:srgbClr val="FFC000"/></a:accent4>
      <a:accent5><a:srgbClr val="5B9BD5"/></a:accent5>
      <a:accent6><a:srgbClr val="70AD47"/></a:accent6>
      <a:hlink><a:srgbClr val="0563C1"/></a:hlink>
      <a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
    </a:clrScheme>
  </a:themeElements>
</a:theme>`)
	body := xmlHeader + `<w:body><w:p><w:r><w:t>Text</w:t></w:r></w:p></w:body></w:document>`

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f1, _ := w.Create("word/document.xml")
	f1.Write([]byte(body))
	f2, _ := w.Create("word/styles.xml")
	f2.Write([]byte(styles))
	f3, _ := w.Create("word/theme/theme1.xml")
	f3.Write(themeXML)
	w.Close()
	data := buf.Bytes()

	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `<s:theme font="Calibri"`) {
		t.Errorf("expected theme font, got: %s", doc.WordsXML)
	}
}

func TestRunThemeFontOnRun(t *testing.T) {
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`
	body := xmlHeader + `<w:body><w:p><w:r><w:rPr><w:rFonts w:asciiTheme="majorHAnsi" w:hAnsiTheme="majorHAnsi"/></w:rPr><w:t>Themed</w:t></w:r></w:p></w:body></w:document>`

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f1, _ := w.Create("word/document.xml")
	f1.Write([]byte(body))
	f2, _ := w.Create("word/styles.xml")
	f2.Write([]byte(styles))
	themeXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="T">
  <a:themeElements>
    <a:fontScheme name="T">
      <a:majorFont><a:latin typeface="Cambria"/></a:majorFont>
      <a:minorFont><a:latin typeface="Calibri"/></a:minorFont>
    </a:fontScheme>
    <a:clrScheme name="T">
      <a:dk1><a:srgbClr val="000000"/></a:dk1><a:lt1><a:srgbClr val="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="44546A"/></a:dk2><a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
      <a:accent1><a:srgbClr val="4472C4"/></a:accent1><a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
      <a:accent3><a:srgbClr val="A5A5A5"/></a:accent3><a:accent4><a:srgbClr val="FFC000"/></a:accent4>
      <a:accent5><a:srgbClr val="5B9BD5"/></a:accent5><a:accent6><a:srgbClr val="70AD47"/></a:accent6>
      <a:hlink><a:srgbClr val="0563C1"/></a:hlink><a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
    </a:clrScheme>
  </a:themeElements>
</a:theme>`)
	f3, _ := w.Create("word/theme/theme1.xml")
	f3.Write(themeXML)
	w.Close()
	data := buf.Bytes()

	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `font="Cambria"`) {
		t.Errorf("expected Cambria from theme on run, got: %s", doc.WordsXML)
	}
}

func TestRunDirectionRTL(t *testing.T) {
	body := xmlHeader + `<w:body><w:p><w:r><w:rPr><w:rtl/></w:rPr><w:t>RTL</w:t></w:r></w:p></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `dir="rtl"`) {
		t.Errorf("expected dir=rtl on run, got: %s", doc.WordsXML)
	}
}

func TestHighlightOnRun(t *testing.T) {
	body := xmlHeader + `<w:body><w:p><w:r><w:rPr><w:highlight w:val="yellow"/></w:rPr><w:t>highlighted</w:t></w:r></w:p></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `highlight="yellow"`) {
		t.Errorf("expected highlight, got: %s", doc.WordsXML)
	}
}

func TestColorOnRun(t *testing.T) {
	body := xmlHeader + `<w:body><w:p><w:r><w:rPr><w:color w:val="FF0000"/></w:rPr><w:t>red</w:t></w:r></w:p></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `color="FF0000"`) {
		t.Errorf("expected color, got: %s", doc.WordsXML)
	}
}

func TestSizeOnRun(t *testing.T) {
	body := xmlHeader + `<w:body><w:p><w:r><w:rPr><w:sz w:val="24"/><w:szCs w:val="28"/></w:rPr><w:t>sized</w:t></w:r></w:p></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `size="12"`) {
		t.Errorf("expected size, got: %s", doc.WordsXML)
	}
	if !strings.Contains(doc.WordsXML, `sizeCS="14"`) {
		t.Errorf("expected sizeCS, got: %s", doc.WordsXML)
	}
}

func TestFontOnRun(t *testing.T) {
	body := xmlHeader + `<w:body><w:p><w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:eastAsia="MS Gothic" w:cs="Arial"/></w:rPr><w:t>fonted</w:t></w:r></w:p></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `font="Arial"`) {
		t.Errorf("expected font, got: %s", x)
	}
	if !strings.Contains(x, `fontEA="MS Gothic"`) {
		t.Errorf("expected fontEA, got: %s", x)
	}
	if !strings.Contains(x, `fontCS="Arial"`) {
		t.Errorf("expected fontCS, got: %s", x)
	}
}

func TestEmptySortedKeys(t *testing.T) {
	m := map[string][]ContentItem{}
	keys := sortedKeys(m)
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %d", len(keys))
	}
}

func TestSortedKeysMultiple(t *testing.T) {
	m := map[string][]ContentItem{"c": nil, "a": nil, "b": nil}
	keys := sortedKeys(m)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("expected [a b c], got %v", keys)
	}
}

func TestParseParagraphTextAlign(t *testing.T) {
	body := xmlHeader + `<w:body><w:p><w:pPr><w:textAlignment w:val="center"/></w:pPr><w:r><w:t>x</w:t></w:r></w:p></w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(doc.WordsXML, `valign="center"`) {
		t.Errorf("expected valign, got: %s", doc.WordsXML)
	}
}
