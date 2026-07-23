package words

import (
	"strings"
	"testing"
)

func TestBrInListItem(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:pPr>
      <w:numPr>
        <w:ilvl w:val="0"/>
        <w:numId w:val="1"/>
      </w:numPr>
    </w:pPr>
    <w:r><w:t>item1</w:t></w:r>
  </w:p>
  <w:p>
    <w:pPr>
      <w:numPr>
        <w:ilvl w:val="0"/>
        <w:numId w:val="1"/>
      </w:numPr>
    </w:pPr>
    <w:r><w:t>before</w:t></w:r>
    <w:r><w:br w:type="textWrapping"/></w:r>
    <w:r><w:t>after</w:t></w:r>
  </w:p>
</w:body></w:document>`
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:num w:numId="1">
    <w:abstractNumId w:val="1"/>
  </w:num>
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0">
      <w:numFmt w:val="bullet"/>
      <w:lvlText w:val="\u2022"/>
    </w:lvl>
  </w:abstractNum>
</w:numbering>`
	data := makeDocxWithParts(body, styles, numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("OUTPUT:\n%s", doc.WordsXML)
	if !strings.Contains(doc.WordsXML, `<br type="textWrapping"/>`) {
		t.Error("expected <br type=\"textWrapping\"/> in list item")
	}
	if !strings.Contains(doc.WordsXML, "after") {
		t.Error("expected 'after' text preserved in list item")
	}
}

func TestBrSameRunAsText(t *testing.T) {
	body := xmlHeader + `<w:body>
  <w:p>
    <w:r>
      <w:br w:type="textWrapping"/>
      <w:t>text after br</w:t>
    </w:r>
  </w:p>
</w:body></w:document>`
	data := makeMinimalDocx(body)
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("OUTPUT:\n%s", doc.WordsXML)
	if !strings.Contains(doc.WordsXML, `<br type="textWrapping"/>`) {
		t.Error("expected <br type=\"textWrapping\"/>")
	}
	if !strings.Contains(doc.WordsXML, "text after br") {
		t.Error("expected text after br to be preserved")
	}
}

func TestBrSeparateRuns(t *testing.T) {
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
	t.Logf("OUTPUT:\n%s", doc.WordsXML)
	if !strings.Contains(doc.WordsXML, `<br type="textWrapping"/>`) {
		t.Error("expected <br type=\"textWrapping\"/>")
	}
	if !strings.Contains(doc.WordsXML, "after") {
		t.Error("expected 'after' text preserved")
	}
}
