package docx

import (
	"encoding/xml"
	"strings"
	"testing"

	docs "github.com/mmonterroca/docxgo/v2"
)

func TestParseBorderAttribute(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]borderInfo
	}{
		{
			name:     "empty",
			input:    "",
			expected: map[string]borderInfo{},
		},
		{
			name:  "single top border",
			input: "top 4 single #000000",
			expected: map[string]borderInfo{
				"top": {style: "single", sz: 4, color: "000000", space: 4},
			},
		},
		{
			name:  "multiple borders",
			input: "top 4 single #000000; bottom 4 double #FF0000",
			expected: map[string]borderInfo{
				"top":    {style: "single", sz: 4, color: "000000", space: 4},
				"bottom": {style: "double", sz: 4, color: "FF0000", space: 4},
			},
		},
		{
			name:  "all four borders",
			input: "top 4 single #000000; bottom 4 single #000000; left 4 single #000000; right 4 single #000000",
			expected: map[string]borderInfo{
				"top":    {style: "single", sz: 4, color: "000000", space: 4},
				"bottom": {style: "single", sz: 4, color: "000000", space: 4},
				"left":   {style: "single", sz: 4, color: "000000", space: 4},
				"right":  {style: "single", sz: 4, color: "000000", space: 4},
			},
		},
		{
			name:  "dashed border",
			input: "top 2 dashed #FF0000",
			expected: map[string]borderInfo{
				"top": {style: "dashed", sz: 2, color: "FF0000", space: 4},
			},
		},
		{
			name:  "thick border",
			input: "bottom 8 thick #0000FF",
			expected: map[string]borderInfo{
				"bottom": {style: "thick", sz: 8, color: "0000FF", space: 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBorderAttribute(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseBorderAttribute(%q) returned %d borders, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for side, expected := range tt.expected {
				if actual, ok := result[side]; !ok {
					t.Errorf("parseBorderAttribute(%q) missing border for %s", tt.input, side)
				} else {
					if actual.style != expected.style {
						t.Errorf("parseBorderAttribute(%q)[%s].style = %q, want %q", tt.input, side, actual.style, expected.style)
					}
					if actual.sz != expected.sz {
						t.Errorf("parseBorderAttribute(%q)[%s].sz = %d, want %d", tt.input, side, actual.sz, expected.sz)
					}
					if actual.color != expected.color {
						t.Errorf("parseBorderAttribute(%q)[%s].color = %q, want %q", tt.input, side, actual.color, expected.color)
					}
				}
			}
		})
	}
}

func TestMapBorderStyle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"single", "single"},
		{"s", "single"},
		{"double", "double"},
		{"d", "double"},
		{"dashed", "dashed"},
		{"ds", "dashed"},
		{"dotted", "dotted"},
		{"dt", "dotted"},
		{"none", "none"},
		{"n", "none"},
		{"unknown", "single"},
		{"", "single"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapBorderStyle(tt.input)
			if result != tt.expected {
				t.Errorf("mapBorderStyle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseUnderlineStyle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"single", "single"},
		{"double", "double"},
		{"thick", "thick"},
		{"dotted", "dotted"},
		{"wave", "wave"},
		{"", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseUnderlineStyle(tt.input)
			if result != tt.expected {
				t.Errorf("parseUnderlineStyle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapTabAlign(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"left", "left"},
		{"center", "center"},
		{"right", "right"},
		{"decimal", "decimal"},
		{"unknown", "left"},
		{"", "left"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapTabAlign(tt.input)
			if result != tt.expected {
				t.Errorf("mapTabAlign(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapLeader(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"dot", "dot"},
		{"hyphen", "hyphen"},
		{"underscore", "underscore"},
		{"heavy", "heavy"},
		{"underline", "underline"},
		{"", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapLeader(tt.input)
			if result != tt.expected {
				t.Errorf("mapLeader(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name     string
		elements string
		classes  string
		expected []string
	}{
		{
			name:     "empty",
			elements: "",
			classes:  "",
			expected: nil,
		},
		{
			name:     "elements only",
			elements: "p,h1",
			classes:  "",
			expected: []string{"p", "h1"},
		},
		{
			name:     "classes only",
			elements: "",
			classes:  "custom,style",
			expected: []string{".custom", ".style"},
		},
		{
			name:     "both",
			elements: "p",
			classes:  "custom",
			expected: []string{"p", ".custom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSelector(tt.elements, tt.classes)
			if len(result) != len(tt.expected) {
				t.Errorf("parseSelector(%q, %q) returned %d items, want %d", tt.elements, tt.classes, len(result), len(tt.expected))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("parseSelector(%q, %q)[%d] = %q, want %q", tt.elements, tt.classes, i, result[i], expected)
				}
			}
		})
	}
}

func TestExtractNumberingDefs(t *testing.T) {
	doc := &WordsXML{
		Write: &WriteXML{
			Content: []interface{}{
				UlXML{
					Items: []LiXML{
						{Content: []interface{}{"item"}},
					},
				},
			},
		},
	}

	defs := extractNumberingDefs(doc)
	if len(defs) == 0 {
		t.Error("extractNumberingDefs returned no defs for document with ul")
	}
}

func TestApplyBorderToParagraph(t *testing.T) {
	docx := docs.NewDocument()
	p, _ := docx.AddParagraph()

	borders := map[string]borderInfo{
		"top":    {style: "single", sz: 4, color: "000000", space: 4},
		"bottom": {style: "double", sz: 4, color: "FF0000", space: 4},
	}

	applyBorderToParagraph(p, borders)
}

func TestApplyBorderToParagraphEmpty(t *testing.T) {
	docx := docs.NewDocument()
	p, _ := docx.AddParagraph()

	applyBorderToParagraph(p, map[string]borderInfo{})
}

func TestParseSelectorEmpty(t *testing.T) {
	result := parseSelector("", "")
	if len(result) != 0 {
		t.Errorf("parseSelector empty returned %d items, want 0", len(result))
	}
}

func TestParseSelectorElements(t *testing.T) {
	result := parseSelector("p,h1,h2", "")
	if len(result) != 3 {
		t.Errorf("parseSelector elements returned %d items, want 3", len(result))
	}
}

func TestParseSelectorClasses(t *testing.T) {
	result := parseSelector("", "custom,style")
	if len(result) != 2 {
		t.Errorf("parseSelector classes returned %d items, want 2", len(result))
	}
}

func TestParseSelectorBoth(t *testing.T) {
	result := parseSelector("p", "custom")
	if len(result) != 2 {
		t.Errorf("parseSelector both returned %d items, want 2", len(result))
	}
}

func TestMatchSelector(t *testing.T) {
	docx := docs.NewDocument()
	p, _ := docx.AddParagraph()

	tests := []struct {
		name     string
		rules    []string
		expected bool
	}{
		{"empty", nil, false},
		{"single", []string{"p"}, true},
		{"multiple", []string{"p", ".custom"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchSelector(tt.rules, p)
			if result != tt.expected {
				t.Errorf("matchSelector(%v, p) = %v, want %v", tt.rules, result, tt.expected)
			}
		})
	}
}

func TestGetAttr(t *testing.T) {
	attrs := []xml.Attr{
		{Name: xml.Name{Local: "font"}, Value: "Arial"},
		{Name: xml.Name{Local: "size"}, Value: "12"},
	}

	tests := []struct {
		name     string
		attrName string
		expected string
	}{
		{"found", "font", "Arial"},
		{"found2", "size", "12"},
		{"not found", "color", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAttr(attrs, tt.attrName)
			if result != tt.expected {
				t.Errorf("getAttr(%v, %q) = %q, want %q", attrs, tt.attrName, result, tt.expected)
			}
		})
	}
}

func TestParseTableBorders(t *testing.T) {
	tests := []struct {
		name       string
		at         string
		expectNone bool
	}{
		{"empty", "", true},
		{"top only", "top 4 single #000000", false},
		{"all four", "top 4 single #000000;right 4 single #000000;bottom 4 single #000000;left 4 single #000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			borders := parseTableBorders(tt.at)
			if tt.expectNone {
				if borders.Top.Style != 0 || borders.Left.Style != 0 || borders.Bottom.Style != 0 || borders.Right.Style != 0 {
					t.Errorf("expected all borders to be zero, got Top=%v Left=%v Bottom=%v Right=%v",
						borders.Top.Style, borders.Left.Style, borders.Bottom.Style, borders.Right.Style)
				}
			} else {
				if borders.Top.Style == 0 && borders.Left.Style == 0 && borders.Bottom.Style == 0 && borders.Right.Style == 0 {
					t.Error("expected at least one border to be set")
				}
			}
		})
	}
}

func TestMapTableAlignment(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"center", 1},
		{"right", 2},
		{"justify", 4},
		{"distribute", 4},
		{"left", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapTableAlignment(tt.input)
			if int(result) != tt.expected {
				t.Errorf("mapTableAlignment(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEmitTableIntegration(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}

	tableXML := []byte(`<table>
		<tr>
			<td>
				<p>Hello</p>
			</td>
		</tr>
	</table>`)

	err := emitter.emitTable(tableXML)
	if err != nil {
		t.Fatalf("emitTable failed: %v", err)
	}

	tables := docx.Tables()
	if len(tables) == 0 {
		t.Fatal("expected at least one table")
	}
}

func TestEmitBrTypePage(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "br"},
		Type:    "page",
	}
	emitter.emitInline(inline, p)
	breaks := p.Runs()
	if len(breaks) == 0 {
		t.Fatal("expected at least one run with break")
	}
}

func TestEmitBrTypeColumn(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "br"},
		Type:    "column",
	}
	emitter.emitInline(inline, p)
	breaks := p.Runs()
	if len(breaks) == 0 {
		t.Fatal("expected at least one run with break")
	}
}

func TestEmitInlineTab(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "tab"},
	}
	emitter.emitInline(inline, p)
	runs := p.Runs()
	if len(runs) == 0 {
		t.Fatal("expected at least one run with tab")
	}
}

func TestEmitFnRef(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "fn-ref"},
		ID:      "1",
	}
	emitter.emitInline(inline, p)
	runs := p.Runs()
	if len(runs) == 0 {
		t.Fatal("expected at least one run with fn-ref")
	}
	text := runs[0].Text()
	if text != "1" {
		t.Errorf("expected fn-ref text '1', got %q", text)
	}
}

func TestEmitSmallcaps(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "smallcaps"},
		Text:    "Small Text",
	}
	emitter.emitInline(inline, p)
	if len(emitter.pendingMods) != 1 {
		t.Fatalf("expected 1 pending mod, got %d", len(emitter.pendingMods))
	}
	if emitter.pendingMods[0].modType != "smallcaps" {
		t.Errorf("expected smallcaps mod, got %s", emitter.pendingMods[0].modType)
	}
}

func TestEmitUppercase(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "uppercase"},
		Text:    "Upper Text",
	}
	emitter.emitInline(inline, p)
	if len(emitter.pendingMods) != 1 {
		t.Fatalf("expected 1 pending mod, got %d", len(emitter.pendingMods))
	}
	if emitter.pendingMods[0].modType != "uppercase" {
		t.Errorf("expected uppercase mod, got %s", emitter.pendingMods[0].modType)
	}
}

func TestEmitSup(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "sup"},
		Text:    "2",
	}
	emitter.emitInline(inline, p)
	if len(emitter.pendingMods) != 1 {
		t.Fatalf("expected 1 pending mod, got %d", len(emitter.pendingMods))
	}
	if emitter.pendingMods[0].modType != "superscript" {
		t.Errorf("expected superscript mod, got %s", emitter.pendingMods[0].modType)
	}
}

func TestEmitSub(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "sub"},
		Text:    "2",
	}
	emitter.emitInline(inline, p)
	if len(emitter.pendingMods) != 1 {
		t.Fatalf("expected 1 pending mod, got %d", len(emitter.pendingMods))
	}
	if emitter.pendingMods[0].modType != "subscript" {
		t.Errorf("expected subscript mod, got %s", emitter.pendingMods[0].modType)
	}
}

func TestEmitIns(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:    &WordsXML{},
		docx:   docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "ins"},
		Text:    "inserted text",
	}
	emitter.emitInline(inline, p)
	if len(emitter.insTexts) != 1 {
		t.Fatalf("expected 1 ins text, got %d", len(emitter.insTexts))
	}
	if emitter.insTexts[0] != "inserted text" {
		t.Errorf("expected 'inserted text', got %q", emitter.insTexts[0])
	}
}

func TestEmitDel(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:    &WordsXML{},
		docx:   docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "del"},
		Text:    "deleted text",
	}
	emitter.emitInline(inline, p)
	if len(emitter.delTexts) != 1 {
		t.Fatalf("expected 1 del text, got %d", len(emitter.delTexts))
	}
	if emitter.delTexts[0] != "deleted text" {
		t.Errorf("expected 'deleted text', got %q", emitter.delTexts[0])
	}
}

func TestInjectRunProp(t *testing.T) {
	tests := []struct {
		name     string
		xmlStr   string
		text     string
		prop     string
		contains string
	}{
		{
			name:     "smallcaps",
			xmlStr:   `<w:r><w:t>Hello</w:t></w:r>`,
			text:     "Hello",
			prop:     "<w:smallCaps/>",
			contains: "<w:smallCaps/>",
		},
		{
			name:     "uppercase",
			xmlStr:   `<w:r><w:t>World</w:t></w:r>`,
			text:     "World",
			prop:     "<w:caps/>",
			contains: "<w:caps/>",
		},
		{
			name:     "superscript",
			xmlStr:   `<w:r><w:t>2</w:t></w:r>`,
			text:     "2",
			prop:     `<w:vertAlign w:val="superscript"/>`,
			contains: `<w:vertAlign w:val="superscript"/>`,
		},
		{
			name:     "empty text",
			xmlStr:   `<w:r><w:t>Hello</w:t></w:r>`,
			text:     "",
			prop:     "<w:smallCaps/>",
			contains: "<w:r>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectRunProp(tt.xmlStr, tt.text, tt.prop)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestWrapRunInTrackedChange(t *testing.T) {
	xmlStr := `<w:r><w:t>inserted</w:t></w:r>`
	idCounter := 1
	result := wrapRunInTrackedChange(xmlStr, "inserted", "ins", "test", "2024-01-01T00:00:00Z", &idCounter)
	if !strings.Contains(result, "<w:ins") {
		t.Errorf("expected result to contain <w:ins, got %q", result)
	}
	if !strings.Contains(result, `w:author="test"`) {
		t.Errorf("expected result to contain author, got %q", result)
	}
}

func TestPostProcessDocx(t *testing.T) {
	result, err := postProcessDocx(nil, nil)
	if err != nil {
		t.Fatalf("postProcessDocx failed: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result for nil input with no mods")
	}
}

func TestPostProcessInsDel(t *testing.T) {
	result, err := postProcessInsDel(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("postProcessInsDel failed: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result for nil input with no texts")
	}
}

func TestEmitParaWithLang(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	paraXML := []byte(`<p lang="en-US">Hello world</p>`)
	err := emitter.emitPara(paraXML, "")
	if err != nil {
		t.Fatalf("emitPara failed: %v", err)
	}
	langFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "lang:") {
			langFound = true
			break
		}
	}
	if !langFound {
		t.Error("expected lang mod to be added")
	}
}

func TestEmitParaWithDirRtl(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	paraXML := []byte(`<p dir="rtl">مرحبا بالعالم</p>`)
	err := emitter.emitPara(paraXML, "")
	if err != nil {
		t.Fatalf("emitPara failed: %v", err)
	}
	bidiFound := false
	for _, mod := range emitter.pendingMods {
		if mod.modType == "bidi" {
			bidiFound = true
			break
		}
	}
	if !bidiFound {
		t.Error("expected bidi mod to be added")
	}
}

func TestEmitSpanWithLang(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "span"},
		Text:    "Hello",
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "lang"}, Value: "fr-FR"},
		},
	}
	emitter.emitSpanInline(inline, p)
	langFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "runlang:") {
			langFound = true
			break
		}
	}
	if !langFound {
		t.Error("expected runlang mod to be added")
	}
}

func TestInjectParaLang(t *testing.T) {
	xmlStr := `<w:p><w:r><w:t>Hello</w:t></w:r></w:p>`
	result := injectParaLang(xmlStr, "en-US")
	if !strings.Contains(result, "w:lang") {
		t.Errorf("expected lang in result, got %q", result)
	}
	if !strings.Contains(result, "en-US") {
		t.Errorf("expected en-US in result, got %q", result)
	}
}

func TestInjectParaBidi(t *testing.T) {
	xmlStr := `<w:p><w:r><w:t>Hello</w:t></w:r></w:p>`
	result := injectParaBidi(xmlStr)
	if !strings.Contains(result, "<w:bidi/>") {
		t.Errorf("expected bidi in result, got %q", result)
	}
}

func TestInjectRunLang(t *testing.T) {
	xmlStr := `<w:r><w:t>Hello</w:t></w:r>`
	result := injectRunLang(xmlStr, "Hello", "fr-FR")
	if !strings.Contains(result, "w:lang") {
		t.Errorf("expected lang in result, got %q", result)
	}
	if !strings.Contains(result, "fr-FR") {
		t.Errorf("expected fr-FR in result, got %q", result)
	}
}

func TestMatchStyleSelector(t *testing.T) {
	tests := []struct {
		name      string
		el        string
		class     string
		tagName   string
		styleName string
		expected  bool
	}{
		{"both empty", "", "", "", "", true},
		{"el match", "p", "", "p", "", true},
		{"el no match", "p", "", "h1", "", false},
		{"class match", "", "myclass", "", "myclass", true},
		{"class no match", "", "myclass", "", "other", false},
		{"both match", "p", "myclass", "p", "myclass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchStyleSelector(tt.el, tt.class, tt.tagName, tt.styleName)
			if result != tt.expected {
				t.Errorf("matchStyleSelector(%q, %q, %q, %q) = %v, want %v",
					tt.el, tt.class, tt.tagName, tt.styleName, result, tt.expected)
			}
		})
	}
}

func TestMapLineRule(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"auto", 0},
		{"exact", 1},
		{"atLeast", 2},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapLineRule(tt.input)
			if int(result) != tt.expected {
				t.Errorf("mapLineRule(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestApplyStyleRules(t *testing.T) {
	docx := docs.NewDocument()
	p, _ := docx.AddParagraph()

	style := &StyleXML{
		Gaps: []StyleGap{
			{EL: "p", Before: 100, After: 200},
		},
		Lines: []StyleLine{
			{EL: "p", Value: 240, Rule: "auto"},
		},
		Indents: []StyleIndent{
			{EL: "p", Left: 720},
		},
		Aligns: []StyleAlign{
			{EL: "p", Value: "center"},
		},
	}

	applyStyleRules(p, style, "p", "")
}

func TestApplyHeadersFooters(t *testing.T) {
	docx := docs.NewDocument()
	doc := &WordsXML{
		Headers: []HeaderXML{
			{
				ID: 1,
				Content: []interface{}{
					"Header Text",
				},
			},
		},
		Footers: []FooterXML{
			{
				ID: 1,
				Content: []interface{}{
					"Footer Text",
				},
			},
		},
	}
	applyHeadersFooters(doc, docx)
}

func TestApplyHeadersFootersEmpty(t *testing.T) {
	docx := docs.NewDocument()
	doc := &WordsXML{}
	applyHeadersFooters(doc, docx)
}

func TestApplyStylesToDocumentWithMeta(t *testing.T) {
	docx := docs.NewDocument()
	doc := &WordsXML{
		Meta: &MetaXML{
			Title:    "Test Title",
			Author:   "Test Author",
			Subject:  "Test Subject",
			Keywords: "key1,key2",
			Created:  "2024-01-01T00:00:00Z",
			Modified: "2024-01-02T00:00:00Z",
		},
	}
	applyStylesToDocument(doc, docx)
}

func TestApplyStylesToDocumentWithPages(t *testing.T) {
	docx := docs.NewDocument()
	doc := &WordsXML{
		Style: &StyleXML{
			Pages: []StylePage{
				{W: 11906, H: 16838, MT: 1440, MB: 1440, ML: 1800, MR: 1800},
			},
		},
	}
	applyStylesToDocument(doc, docx)
}

func TestInjectBookmark(t *testing.T) {
	xmlStr := `<w:r><w:t>Hello</w:t></w:r>`
	result := injectBookmark(xmlStr, "Hello", "bm1")
	if !strings.Contains(result, "<w:bookmarkStart") {
		t.Errorf("expected bookmarkStart in result, got %q", result)
	}
	if !strings.Contains(result, "<w:bookmarkEnd") {
		t.Errorf("expected bookmarkEnd in result, got %q", result)
	}
	if !strings.Contains(result, "w:name=\"bm1\"") {
		t.Errorf("expected bookmark name in result, got %q", result)
	}
}

func TestInjectComment(t *testing.T) {
	xmlStr := `<w:r><w:t>Hello</w:t></w:r>`
	result := injectComment(xmlStr, "Hello", "c1")
	if !strings.Contains(result, "<w:commentRangeStart") {
		t.Errorf("expected commentRangeStart in result, got %q", result)
	}
	if !strings.Contains(result, "<w:commentRangeEnd") {
		t.Errorf("expected commentRangeEnd in result, got %q", result)
	}
	if !strings.Contains(result, "<w:commentReference") {
		t.Errorf("expected commentReference in result, got %q", result)
	}
}

func TestInjectFnRef(t *testing.T) {
	xmlStr := `<w:r><w:t>1</w:t></w:r>`
	result := injectFnRef(xmlStr, "1", "fn1")
	if !strings.Contains(result, "<w:footnoteReference") {
		t.Errorf("expected footnoteReference in result, got %q", result)
	}
	if !strings.Contains(result, "w:id=\"fn1\"") {
		t.Errorf("expected footnote id in result, got %q", result)
	}
}

func TestEmitBm(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "bm"},
		ID:      "bookmark1",
		Text:    "Bookmark",
	}
	emitter.emitInline(inline, p)
	bmFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "bookmark:") {
			bmFound = true
			break
		}
	}
	if !bmFound {
		t.Error("expected bookmark mod to be added")
	}
}

func TestEmitComment(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "comment"},
		ID:      "comment1",
		Text:    "Comment text",
	}
	emitter.emitInline(inline, p)
	cmFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "comment:") {
			cmFound = true
			break
		}
	}
	if !cmFound {
		t.Error("expected comment mod to be added")
	}
}

func TestEmitFnRefWithMod(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "fn-ref"},
		ID:      "fn1",
		Text:    "1",
	}
	emitter.emitInline(inline, p)
	fnFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "fnref:") {
			fnFound = true
			break
		}
	}
	if !fnFound {
		t.Error("expected fnref mod to be added")
	}
}

func TestEmitUlWithNumbering(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
		numIdMap:    map[string]int{"bullet:1": 1},
	}
	ulContent := []interface{}{
		map[string]interface{}{
			"XMLName": xml.Name{Local: "li"},
			"Content": []interface{}{"Item 1"},
		},
	}
	err := emitter.emitUl(ulContent)
	if err != nil {
		t.Fatalf("emitUl failed: %v", err)
	}
}

func TestEmitOlWithNumbering(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
		numIdMap:    map[string]int{"decimal:1": 2},
	}
	olContent := []interface{}{
		map[string]interface{}{
			"XMLName": xml.Name{Local: "li"},
			"Content": []interface{}{"Item 1"},
		},
	}
	err := emitter.emitOl(olContent)
	if err != nil {
		t.Fatalf("emitOl failed: %v", err)
	}
}

func TestEmitLiWithNestedList(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
		numIdMap:    map[string]int{"bullet:1": 1, "decimal:1": 2},
	}
	liContent := []interface{}{
		"Parent item",
		map[string]interface{}{
			"XMLName": xml.Name{Local: "ul"},
			"Content": []interface{}{
				map[string]interface{}{
					"XMLName": xml.Name{Local: "li"},
					"Content": []interface{}{"Nested item"},
				},
			},
		},
	}
	err := emitter.emitLi(liContent, 1, 0)
	if err != nil {
		t.Fatalf("emitLi failed: %v", err)
	}
}

func TestEmitH7H8H9(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	for _, tag := range []string{"h7", "h8", "h9"} {
		data := []byte(`<` + tag + `>Heading</` + tag + `>`)
		err := emitter.emitContent(data)
		if err != nil {
			t.Fatalf("emitContent(%s) failed: %v", tag, err)
		}
	}
}

func TestEmitBlockquoteWithAttrs(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<blockquote lang="en" dir="rtl" at="left 4 single #FF0000">quoted text</blockquote>`)
	err := emitter.emitBlockquote(data)
	if err != nil {
		t.Fatalf("emitBlockquote failed: %v", err)
	}
	langFound := false
	bidiFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "lang:") {
			langFound = true
		}
		if mod.modType == "bidi" {
			bidiFound = true
		}
	}
	if !langFound {
		t.Error("expected lang mod from blockquote")
	}
	if !bidiFound {
		t.Error("expected bidi mod from blockquote")
	}
}

func TestEmitOlWithTypeAndStart(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
		numIdMap:    map[string]int{"lowerLetter:5": 3, "decimal:1": 2},
	}
	data := []byte(`<ol type="lowerletter" start="5"><li>Item A</li></ol>`)
	err := emitter.emitContent(data)
	if err != nil {
		t.Fatalf("emitContent(ol) failed: %v", err)
	}
}

func TestEmitTab(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "tab"},
	}
	emitter.emitInline(inline, p)
	tabFound := false
	for _, mod := range emitter.pendingMods {
		if mod.modType == "tab" {
			tabFound = true
			break
		}
	}
	if !tabFound {
		t.Error("expected tab mod to be added")
	}
}

func TestInjectTab(t *testing.T) {
	xmlStr := `<w:r><w:t>Hello</w:t></w:r>`
	result := injectTab(xmlStr)
	if !strings.Contains(result, "<w:tab/>") {
		t.Errorf("expected <w:tab/> in result, got %q", result)
	}
}

func TestEmitImgXmlUnmarshal(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	content := []interface{}{
		InlineXML{
			XMLName: xml.Name{Local: "img"},
			Alt:     "Test image",
		},
	}
	err := emitter.emitImg(content)
	if err != nil {
		t.Fatalf("emitImg failed: %v", err)
	}
}

func TestEmitInsDelWithAuthorDate(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "ins"},
		Text:    "new text",
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "author"}, Value: "John"},
			{Name: xml.Name{Local: "date"}, Value: "2025-01-01T00:00:00Z"},
		},
	}
	emitter.emitInline(inline, p)
	if len(emitter.insTexts) != 1 {
		t.Fatalf("expected 1 ins text, got %d", len(emitter.insTexts))
	}
	insModFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "insAuthor:John") {
			insModFound = true
			break
		}
	}
	if !insModFound {
		t.Error("expected insAuthor:John mod")
	}
}

func TestEmitSpanHighlight(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "span"},
		Text:    "highlighted",
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "highlight"}, Value: "yellow"},
		},
	}
	emitter.emitSpanInline(inline, p)
	runs := p.Runs()
	if len(runs) == 0 {
		t.Fatal("expected at least one run")
	}
}

func TestEmitSpanHidden(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "span"},
		Text:    "hidden text",
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "hidden"}, Value: "true"},
		},
	}
	emitter.emitSpanInline(inline, p)
	hiddenFound := false
	for _, mod := range emitter.pendingMods {
		if mod.modType == "hidden" {
			hiddenFound = true
			break
		}
	}
	if !hiddenFound {
		t.Error("expected hidden mod")
	}
}

func TestEmitSpanDir(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "span"},
		Text:    "RTL text",
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "dir"}, Value: "rtl"},
		},
	}
	emitter.emitSpanInline(inline, p)
	bidiFound := false
	for _, mod := range emitter.pendingMods {
		if mod.modType == "runBidi" {
			bidiFound = true
			break
		}
	}
	if !bidiFound {
		t.Error("expected runBidi mod")
	}
}

func TestEmitSpanSizeCS(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "span"},
		Text:    "text",
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "sizeCS"}, Value: "24"},
		},
	}
	emitter.emitSpanInline(inline, p)
	sizeCSFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "sizeCS:") {
			sizeCSFound = true
			break
		}
	}
	if !sizeCSFound {
		t.Error("expected sizeCS mod")
	}
}

func TestEmitParaValign(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<p valign="center">centered</p>`)
	err := emitter.emitPara(data, "")
	if err != nil {
		t.Fatalf("emitPara failed: %v", err)
	}
	valignFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "pvalign:") {
			valignFound = true
			break
		}
	}
	if !valignFound {
		t.Error("expected pvalign mod")
	}
}

func TestEmitBcsIcs(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	p, _ := docx.AddParagraph()

	bcsInline := InlineXML{
		XMLName: xml.Name{Local: "bcs"},
		Text:    "bold CS",
	}
	emitter.emitInline(bcsInline, p)

	icsInline := InlineXML{
		XMLName: xml.Name{Local: "ics"},
		Text:    "italic CS",
	}
	emitter.emitInline(icsInline, p)

	bcsFound := false
	icsFound := false
	for _, mod := range emitter.pendingMods {
		if mod.modType == "boldCs" {
			bcsFound = true
		}
		if mod.modType == "italicCs" {
			icsFound = true
		}
	}
	if !bcsFound {
		t.Error("expected boldCs mod")
	}
	if !icsFound {
		t.Error("expected italicCs mod")
	}
}

func TestEmitSym(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:  &WordsXML{},
		docx: docx,
	}
	p, _ := docx.AddParagraph()
	inline := InlineXML{
		XMLName: xml.Name{Local: "sym"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "char"}, Value: "00A9"},
		},
	}
	emitter.emitInline(inline, p)
	runs := p.Runs()
	if len(runs) == 0 {
		t.Fatal("expected at least one run for sym")
	}
}

func TestEmitTableWithAttrs(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	tableXML := []byte(`<table width="100" align="center" indent="50" cellSpacing="10" at="top 4 single #000000"><tr><td><p>Cell</p></td></tr></table>`)
	err := emitter.emitTable(tableXML)
	if err != nil {
		t.Fatalf("emitTable failed: %v", err)
	}
}

func TestEmitTableCellAttrs(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	tableXML := []byte(`<table><tr><td lang="en" noWrap="true" textDir="rtl"><p>Cell</p></td></tr></table>`)
	err := emitter.emitTable(tableXML)
	if err != nil {
		t.Fatalf("emitTable with cell attrs failed: %v", err)
	}
	cellLangFound := false
	cellNoWrapFound := false
	cellTextDirFound := false
	for _, mod := range emitter.pendingMods {
		if strings.HasPrefix(mod.modType, "cellLang:") {
			cellLangFound = true
		}
		if strings.HasPrefix(mod.modType, "cellNoWrap:") {
			cellNoWrapFound = true
		}
		if strings.HasPrefix(mod.modType, "cellTextDir:") {
			cellTextDirFound = true
		}
	}
	if !cellLangFound {
		t.Error("expected cellLang mod")
	}
	if !cellNoWrapFound {
		t.Error("expected cellNoWrap mod")
	}
	if !cellTextDirFound {
		t.Error("expected cellTextDir mod")
	}
}

func TestApplyStylesToDocumentWithCols(t *testing.T) {
	docx := docs.NewDocument()
	doc := &WordsXML{
		Style: &StyleXML{
			Cols: []StyleCols{
				{N: 3, Space: 720},
			},
		},
	}
	applyStylesToDocument(doc, docx)
}

func TestParseInsDelMod(t *testing.T) {
	tests := []struct {
		name           string
		modType        string
		expectedAuthor string
		expectedDate   string
	}{
		{"full", "insAuthor:John|date:2025-01-01", "John", "2025-01-01"},
		{"author only", "insAuthor:John|date:", "John", ""},
		{"empty", "insAuthor:|date:", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			author, date := parseInsDelMod(tt.modType)
			if author != tt.expectedAuthor {
				t.Errorf("author = %q, want %q", author, tt.expectedAuthor)
			}
			if date != tt.expectedDate {
				t.Errorf("date = %q, want %q", date, tt.expectedDate)
			}
		})
	}
}

func TestInjectParaValign(t *testing.T) {
	tests := []struct {
		name     string
		valign   string
		expected string
	}{
		{"center", "center", `w:val="center"`},
		{"bottom", "bottom", `w:val="bottom"`},
		{"top", "top", `w:val="top"`},
		{"middle", "middle", `w:val="center"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlStr := `<w:p><w:r><w:t>Hello</w:t></w:r></w:p>`
			result := injectParaValign(xmlStr, tt.valign)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected %q in result, got %q", tt.expected, result)
			}
		})
	}
}

func TestInjectCellNoWrap(t *testing.T) {
	xmlStr := `<w:tc><w:tcPr></w:tcPr><w:p></w:p></w:tc>`
	result := injectCellNoWrap(xmlStr, "true")
	if !strings.Contains(result, "<w:noWrap/>") {
		t.Errorf("expected <w:noWrap/> in result, got %q", result)
	}
}

func TestInjectCellTextDir(t *testing.T) {
	xmlStr := `<w:tc><w:tcPr></w:tcPr><w:p></w:p></w:tc>`
	result := injectCellTextDir(xmlStr, "rtl")
	if !strings.Contains(result, `w:val="rtl"`) {
		t.Errorf("expected rtl text direction in result, got %q", result)
	}
}

func TestInjectCellSpacing(t *testing.T) {
	xmlStr := `<w:tbl><w:tblPr></w:tblPr></w:tbl>`
	result := injectCellSpacing(xmlStr, "100")
	if !strings.Contains(result, "tblCellSpacing") {
		t.Errorf("expected tblCellSpacing in result, got %q", result)
	}
}

func TestInjectTableIndent(t *testing.T) {
	xmlStr := `<w:tbl><w:tblPr></w:tblPr></w:tbl>`
	result := injectTableIndent(xmlStr, "50")
	if !strings.Contains(result, "tblInd") {
		t.Errorf("expected tblInd in result, got %q", result)
	}
}

func TestInjectTabStops(t *testing.T) {
	tabs := []StyleTab{
		{Pos: 720, Align: "left", Leader: "dot"},
		{Pos: 1440, Align: "right", Leader: ""},
	}
	xmlStr := `<w:p><w:pPr></w:pPr></w:p>`
	result := injectTabStops(xmlStr, tabs)
	if !strings.Contains(result, "<w:tabs>") {
		t.Errorf("expected <w:tabs> in result, got %q", result)
	}
	if !strings.Contains(result, `w:pos="720"`) {
		t.Errorf("expected tab pos 720 in result, got %q", result)
	}
	if !strings.Contains(result, `w:leader="dot"`) {
		t.Errorf("expected leader=dot in result, got %q", result)
	}
}

func TestEmitContentWithBcsIcsFull(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<p><bcs>Bold CS</bcs></p>`)
	err := emitter.emitContent(data)
	if err != nil {
		t.Fatalf("emitContent with bcs failed: %v", err)
	}
}

func TestEmitContentWithSym(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<p><sym char="00A9"/></p>`)
	err := emitter.emitContent(data)
	if err != nil {
		t.Fatalf("emitContent with sym failed: %v", err)
	}
}

func TestEmitFullDocument(t *testing.T) {
	wordsXML := `<?xml version="1.0" encoding="UTF-8"?>
<words version="1.0.1" mode="edit">
  <meta>
    <title>Test</title>
    <author>Author</author>
  </meta>
  <write>
    <p>Hello <b>world</b></p>
    <h1>Heading</h1>
    <blockquote>quoted</blockquote>
    <ol type="decimal" start="1"><li>Item 1</li><li>Item 2</li></ol>
    <ul type="bullet"><li>Bullet 1</li></ul>
    <table width="50" align="center" at="top 4 single #000000"><tr><td><p>Cell</p></td></tr></table>
    <pre>code block</pre>
    <hr/>
    <p><br type="page"/></p>
    <p><span font="Arial" size="12" color="#FF0000" highlight="yellow">styled</span></p>
    <p><sup>x</sup> and <sub>y</sub></p>
    <p><smallcaps>Small</smallcaps> <uppercase>Upper</uppercase></p>
    <p><bcs>BoldCS</bcs> <ics>ItalicCS</ics></p>
    <p><sym char="00A9"/></p>
    <p dir="rtl" lang="ar">Arabic</p>
    <p valign="center">centered</p>
    <h7>H7</h7><h8>H8</h8><h9>H9</h9>
  </write>
</words>`

	result, err := EmitDocx(wordsXML)
	if err != nil {
		t.Fatalf("EmitDocx failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestWrapRunInTrackedChangeDel(t *testing.T) {
	xmlStr := `<w:r><w:t>deleted</w:t></w:r>`
	idCounter := 1
	result := wrapRunInTrackedChange(xmlStr, "deleted", "del", "author", "2025-06-01T00:00:00Z", &idCounter)
	if !strings.Contains(result, "<w:del") {
		t.Errorf("expected <w:del in result, got %q", result)
	}
	if !strings.Contains(result, `w:author="author"`) {
		t.Errorf("expected author in result, got %q", result)
	}
	if !strings.Contains(result, `w:date="2025-06-01T00:00:00Z"`) {
		t.Errorf("expected date in result, got %q", result)
	}
}

func TestEmitBlockquoteWithAlign(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<blockquote align="center">quoted</blockquote>`)
	err := emitter.emitBlockquote(data)
	if err != nil {
		t.Fatalf("emitBlockquote failed: %v", err)
	}
}

func TestEmitPreWithAttrs(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<pre align="left" indentLeft="720" indentFirst="360">code</pre>`)
	err := emitter.emitPre(data)
	if err != nil {
		t.Fatalf("emitPre failed: %v", err)
	}
}

func TestEmitTableWithCaptionSummary(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	tableXML := []byte(`<table id="1" c="TableGrid" caption="My Table" summary="Table description"><tr><td><p>Cell</p></td></tr></table>`)
	err := emitter.emitTable(tableXML)
	if err != nil {
		t.Fatalf("emitTable failed: %v", err)
	}
	captionFound := false
	summaryFound := false
	styleFound := false
	for _, mod := range emitter.pendingMods {
		if mod.modType == "tableCaption" {
			captionFound = true
		}
		if mod.modType == "tableSummary" {
			summaryFound = true
		}
		if mod.modType == "tableStyle" {
			styleFound = true
		}
	}
	if !captionFound {
		t.Error("expected tableCaption mod")
	}
	if !summaryFound {
		t.Error("expected tableSummary mod")
	}
	if !styleFound {
		t.Error("expected tableStyle mod")
	}
}

func TestBuildStylesXML(t *testing.T) {
	theme := &StyleTheme{
		Font:   "Arial",
		FontEA: "Arial",
		FontCS: "Arial",
	}
	customs := []StyleCustom{
		{
			Name:    "MyStyle",
			Type:    "paragraph",
			BasedOn: "Normal",
			Font:    "Times New Roman",
			Size:    12,
			Bold:    "true",
			Italic:  "true",
			Alignment: "center",
			SpacingBefore: 200,
			SpacingAfter:  100,
			IndentLeft:    720,
		},
	}
	data := buildStylesXML(theme, customs)
	if data == nil {
		t.Fatal("expected non-nil styles data")
	}
	s := string(data)
	if !strings.Contains(s, "Arial") {
		t.Error("expected Arial theme font in styles")
	}
	if !strings.Contains(s, "MyStyle") {
		t.Error("expected MyStyle in styles")
	}
	if !strings.Contains(s, "Times New Roman") {
		t.Error("expected Times New Roman font in styles")
	}
}

func TestBuildStylesXMLNil(t *testing.T) {
	data := buildStylesXML(nil, nil)
	if data != nil {
		t.Error("expected nil for nil theme and customs")
	}
}

func TestBuildStylesXMLEmpty(t *testing.T) {
	data := buildStylesXML(&StyleTheme{}, nil)
	if data != nil {
		t.Error("expected nil for empty theme")
	}
}

func TestBuildStylesXMLCustomOnly(t *testing.T) {
	customs := []StyleCustom{
		{
			Name:     "TestBold",
			Type:     "character",
			Bold:     "true",
			Underline: "single",
			SmallCaps: "true",
			Color:    "FF0000",
			Size:     14,
		},
	}
	data := buildStylesXML(nil, customs)
	if data == nil {
		t.Fatal("expected non-nil styles data")
	}
	s := string(data)
	if !strings.Contains(s, "TestBold") {
		t.Error("expected TestBold in styles")
	}
}

func TestAddStyles(t *testing.T) {
	processor := newZipProcessor([]byte("dummy"))
	doc := &WordsXML{
		Style: &StyleXML{
			Theme: &StyleTheme{Font: "Calibri"},
			Customs: []StyleCustom{
				{Name: "Custom1", Type: "paragraph", Font: "Arial"},
			},
		},
	}
	err := addStyles(processor, doc)
	if err != nil {
		t.Fatalf("addStyles failed: %v", err)
	}
	if _, ok := processor.newParts["word/styles.xml"]; !ok {
		t.Error("expected styles.xml to be added")
	}
}

func TestAddStylesNil(t *testing.T) {
	processor := newZipProcessor([]byte("dummy"))
	doc := &WordsXML{}
	err := addStyles(processor, doc)
	if err != nil {
		t.Fatalf("addStyles failed: %v", err)
	}
}

func TestInjectTableBookmark(t *testing.T) {
	xmlStr := `<w:tbl><w:tblPr></w:tblPr></w:tbl>`
	result := injectTableBookmark(xmlStr, "1")
	if !strings.Contains(result, "tblStyle") {
		t.Errorf("expected tblStyle in result, got %q", result)
	}
}

func TestInjectTableStyle(t *testing.T) {
	xmlStr := `<w:tbl><w:tblPr></w:tblPr></w:tbl>`
	result := injectTableStyle(xmlStr, "TableGrid")
	if !strings.Contains(result, `w:val="TableGrid"`) {
		t.Errorf("expected TableGrid in result, got %q", result)
	}
}

func TestInjectTableCaption(t *testing.T) {
	xmlStr := `<w:tbl><w:tblPr></w:tblPr></w:tbl>`
	result := injectTableCaption(xmlStr, "My Table")
	if !strings.Contains(result, "tblCaption") {
		t.Errorf("expected tblCaption in result, got %q", result)
	}
}

func TestInjectTableSummary(t *testing.T) {
	xmlStr := `<w:tbl><w:tblPr></w:tblPr></w:tbl>`
	result := injectTableSummary(xmlStr, "Description")
	if !strings.Contains(result, "tblSummary") {
		t.Errorf("expected tblSummary in result, got %q", result)
	}
}

func TestCollectMatchingTabs(t *testing.T) {
	style := &StyleXML{
		Tabs: []StyleTab{
			{EL: "p", Pos: 720, Align: "left"},
			{EL: "h1", Pos: 1440, Align: "right"},
		},
	}
	tabs := collectMatchingTabs(style, "p", "")
	if len(tabs) != 1 {
		t.Errorf("expected 1 matching tab, got %d", len(tabs))
	}
	tabs2 := collectMatchingTabs(style, "h1", "")
	if len(tabs2) != 1 {
		t.Errorf("expected 1 matching tab for h1, got %d", len(tabs2))
	}
}

func TestCollectMatchingTabsEmpty(t *testing.T) {
	tabs := collectMatchingTabs(nil, "p", "")
	if len(tabs) != 0 {
		t.Errorf("expected 0 tabs for nil style, got %d", len(tabs))
	}
}

func TestEmitPreWithInlineContent(t *testing.T) {
	docx := docs.NewDocument()
	emitter := &coreEmitter{
		doc:         &WordsXML{},
		docx:        docx,
		pendingMods: []pendingRunMod{},
	}
	data := []byte(`<pre><b>bold code</b></pre>`)
	err := emitter.emitPre(data)
	if err != nil {
		t.Fatalf("emitPre failed: %v", err)
	}
}

func TestEmitFullDocumentWithTheme(t *testing.T) {
	wordsXML := `<?xml version="1.0" encoding="UTF-8"?>
<words version="1.0.1" mode="edit">
  <style>
    <page w="11906" h="16838" mt="1440" mb="1440" ml="1800" mr="1800"/>
    <theme font="Arial" fontEA="Arial" fontCS="Arial"/>
    <custom name="CustomPara" type="paragraph" basedOn="Normal" font="Times New Roman" size="12" bold="true" alignment="center"/>
    <tab el="p" pos="720" align="left" leader="dot"/>
  </style>
  <meta>
    <title>Full Test</title>
    <author>Author</author>
  </meta>
  <write>
    <p>Hello <b>world</b></p>
    <blockquote align="center">quoted text</blockquote>
    <pre align="left" indentLeft="720">code block</pre>
    <ol type="decimal" start="1"><li>Item 1</li></ol>
    <ul type="bullet"><li>Bullet 1</li></ul>
    <table width="50" align="center" at="top 4 single #000000" id="1" c="TableGrid" caption="Test Table" summary="A test"><tr><td><p>Cell</p></td></tr></table>
    <p><span font="Arial" size="12" color="#FF0000" highlight="yellow" sizeCS="14" hidden="true" dir="rtl" lang="fr">styled</span></p>
    <p><bcs>BoldCS</bcs> <ics>ItalicCS</ics></p>
    <p><sym char="00A9"/></p>
    <p dir="rtl" lang="ar">Arabic</p>
    <p valign="center">centered</p>
    <h7>H7</h7><h8>H8</h8><h9>H9</h9>
    <hr/>
    <p><br type="page"/></p>
    <p><sup>x</sup> and <sub>y</sub></p>
    <p><smallcaps>Small</smallcaps> <uppercase>Upper</uppercase></p>
  </write>
</words>`

	result, err := EmitDocx(wordsXML)
	if err != nil {
		t.Fatalf("EmitDocx failed: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}
