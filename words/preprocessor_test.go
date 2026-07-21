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
	if footnotesXML != "" {
		addZipFile("word/footnotes.xml", footnotesXML)
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
