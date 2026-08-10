package words

import (
	"strings"
	"testing"
)

func TestParseLvlTextTemplate(t *testing.T) {
	cases := []struct {
		in   string
		want []LvlTextSegment
	}{
		{"%1.", []LvlTextSegment{{Level: 1}, {Text: "."}}},
		{"(%1)", []LvlTextSegment{{Text: "("}, {Level: 1}, {Text: ")"}}},
		{"(%1.)", []LvlTextSegment{{Text: "("}, {Level: 1}, {Text: ".)"}}},
		{"%1.%2", []LvlTextSegment{{Level: 1}, {Text: "."}, {Level: 2}}},
		{"%2.", []LvlTextSegment{{Level: 2}, {Text: "."}}},
		{"•", []LvlTextSegment{{Text: "•"}}},
		{"-", []LvlTextSegment{{Text: "-"}}},
		{"o", []LvlTextSegment{{Text: "o"}}},
		{"", []LvlTextSegment{{Text: ""}}},
	}
	for _, c := range cases {
		got := parseLvlTextTemplate(c.in)
		if len(got) != len(c.want) {
			t.Errorf("parseLvlTextTemplate(%q) len = %d, want %d (got %+v)", c.in, len(got), len(c.want), got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("parseLvlTextTemplate(%q)[%d] = %+v, want %+v", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestFormatListCounter(t *testing.T) {
	cases := []struct {
		counter int
		numFmt  string
		want    string
	}{
		{1, "decimal", "1"},
		{9, "decimal", "9"},
		{1, "lowerLetter", "a"},
		{2, "lowerLetter", "b"},
		{27, "lowerLetter", "aa"},
		{1, "upperLetter", "A"},
		{4, "upperLetter", "D"},
		{1, "lowerRoman", "i"},
		{4, "lowerRoman", "iv"},
		{9, "lowerRoman", "ix"},
		{1, "upperRoman", "I"},
		{4, "upperRoman", "IV"},
		{1, "decimalZero", "01"},
		{12, "decimalZero", "12"},
		{3, "bullet", ""},
		{3, "unknownFmt", "3"},
	}
	for _, c := range cases {
		got := formatListCounter(c.counter, c.numFmt)
		if got != c.want {
			t.Errorf("formatListCounter(%d, %q) = %q, want %q", c.counter, c.numFmt, got, c.want)
		}
	}
}

func TestLvlTextMarkerExpansion(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="(%1)"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>First</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Second</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Third</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<ol type="decimal" format="(%1)">`) {
		t.Errorf("expected <ol> with format attribute, got: %s", x)
	}
	for _, want := range []string{
		`<li marker="(1)">`,
		`<li marker="(2)">`,
		`<li marker="(3)">`,
	} {
		if !strings.Contains(x, want) {
			t.Errorf("expected %s in output", want)
		}
	}
}

func TestLvlTextPeriodSuffix(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>One</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Two</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<ol type="decimal" format="%1.">`) {
		t.Errorf("expected format attr %q, got: %s", "%1.", x)
	}
	if !strings.Contains(x, `<li marker="1.">`) || !strings.Contains(x, `<li marker="2.">`) {
		t.Errorf("expected markers 1. and 2., got: %s", x)
	}
}

func TestLvlTextBulletLiteralNoFormatAttr(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="bullet"/><w:lvlText w:val="&#8226;"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Item</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if strings.Contains(x, `format=`) {
		t.Errorf("literal bullet must not carry a format attribute, got: %s", x)
	}
	if !strings.Contains(x, `<ul type="•">`) {
		t.Errorf("expected literal bullet in type, got: %s", x)
	}
}

func TestLvlTextMultiLevelMarker(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1"/></w:lvl>
    <w:lvl w:ilvl="1"><w:numFmt w:val="lowerLetter"/><w:lvlText w:val="%1.%2"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Parent</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="1"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Child A</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="1"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Child B</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, `<ol type="lowerLetter" format="%1.%2">`) {
		t.Errorf("expected nested <ol> with format, got: %s", x)
	}
	if !strings.Contains(x, `<li marker="1.a">`) {
		t.Errorf("expected marker 1.a, got: %s", x)
	}
	if !strings.Contains(x, `<li marker="1.b">`) {
		t.Errorf("expected marker 1.b, got: %s", x)
	}
}

func TestLvlTextStartFromBase(t *testing.T) {
	numbering := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="0">
    <w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="(%1)"/><w:start w:val="8"/></w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
	body := xmlHeader + `<w:body>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Eight</w:t></w:r></w:p>
  <w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>Nine</w:t></w:r></w:p>
</w:body></w:document>`
	data := makeDocxWithParts(body, "", numbering, "")
	doc, err := ProcessDOCXBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	x := doc.WordsXML
	if !strings.Contains(x, ` start="8"`) {
		t.Errorf("expected start=8, got: %s", x)
	}
	if !strings.Contains(x, `<li marker="(8)">`) {
		t.Errorf("expected marker (8), got: %s", x)
	}
	if !strings.Contains(x, `<li marker="(9)">`) {
		t.Errorf("expected marker (9), got: %s", x)
	}
}
