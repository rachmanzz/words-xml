package words

import (
	"fmt"
	"strings"
	"testing"
)

func TestVerifyValidMinimal(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p>Hello</p></write>` +
		`</words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyEmptyInput(t *testing.T) {
	r := Verify("")
	if r.Valid {
		t.Error("expected invalid for empty input")
	}
}

func TestVerifyNoRoot(t *testing.T) {
	r := Verify(`not xml at all`)
	if r.Valid {
		t.Error("expected invalid for non-xml")
	}
}

func TestVerifyWrongRoot(t *testing.T) {
	input := `<document xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<write><p>x</p></write></document>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for wrong root")
	}
	if !strings.Contains(r.Errors[0], "<words>") {
		t.Errorf("expected error about <words>, got: %s", r.Errors[0])
	}
}

func TestVerifyMissingNamespace(t *testing.T) {
	input := `<words version="1.0.1" mode="semantic">` +
		`<write><p>x</p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing namespace")
	}
}

func TestVerifyBadVersion(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="2.0.0" mode="semantic">` +
		`<write><p>x</p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad version")
	}
}

func TestVerifyBadMode(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="invalid">` +
		`<write><p>x</p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad mode")
	}
}

func TestVerifyMissingWrite(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing <write>")
	}
}

func TestVerifyMissingStyle(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<write><p>x</p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid (style is warn), got errors: %v", r.Errors)
	}
	if len(r.Warns) == 0 {
		t.Error("expected warn for missing style")
	}
}

func TestVerifyDuplicateElement(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<style unit="in"></style>` +
		`<write><p>x</p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for duplicate <style>")
	}
}

func TestVerifyBadTableID(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><table id="abc">` +
		`<tr><td>x</td></tr>` +
		`</table></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for non-integer table id")
	}
}

func TestVerifyBadColSpec(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><table id="1"><colspec/>` +
		`<tr><td>x</td></tr></table></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for colspec inside table (unexpected element)")
	}
}

func TestVerifyBadAlign(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><table id="1" align="wrong">` +
		`<tr><td>x</td></tr>` +
		`</table></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad table align")
	}
}

func TestVerifyBadColspan(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><table id="1">` +
		`<tr><td colspan="abc">x</td></tr></table></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for non-integer colspan")
	}
}

func TestVerifyBadValign(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><table id="1">` +
		`<tr><td valign="wrong">x</td></tr></table></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad valign")
	}
}

func TestVerifyBadDir(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p dir="center">x</p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad dir")
	}
}

func TestVerifyMissingLi(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><ul></ul></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for empty <ul>")
	}
}

func TestVerifyBadLiType(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><ol><li type="wrong">x</li></ol></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad li type")
	}
}

func TestVerifyMissingAnchorHref(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p><a>link</a></p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing <a> href")
	}
}

func TestVerifyEmptyHref(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p><a href="">link</a></p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for empty href")
	}
}

func TestVerifyImgWithSrc(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p><img src="foo.png" alt="pic"/></p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for <img> with src")
	}
}

func TestVerifyMissingImgAlt(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p><img/></p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing <img> alt")
	}
}

func TestVerifyBadBreakType(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p><br type="wrong"/></p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for bad br type")
	}
}

func TestVerifyMissingBreakType(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p><br/></p></write></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing br type")
	}
}

func TestVerifyMissingNoteId(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p>x</p></write>` +
		`<notes><fn>text</fn></notes></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing note id")
	}
}

func TestVerifyMissingHeaderId(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style>` +
		`<write><p>x</p></write>` +
		`<header><p>hdr</p></header></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing header id")
	}
}

func TestVerifyValidFullDoc(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<meta><title>T</title><author>A</author></meta>` +
		`<style unit="in">` +
		`<s:page size="Letter" mt="1.00" mb="1.00" ml="1.00" mr="1.00" mh="0.50" mf="0.50"/>` +
		`</style>` +
		`<write>` +
		`<p dir="rtl" lang="ar">text</p>` +
		`<h1>Title</h1>` +
		`<table id="1" align="center" caption="C" summary="S">` +
		`<tr><th colspan="2" valign="center">H</th></tr>` +
		`<tr><td rowspan="2" noWrap="true">C</td><td>x</td></tr>` +
		`</table>` +
		`<ul><li tag="1" type="decimal">item</li></ul>` +
		`<ol><li type="bullet"><ul><li>nested</li></ul></li></ol>` +
		`<blockquote>q</blockquote>` +
		`<pre>code</pre>` +
		`<p><b><i><u><s><sup><sub><smallcaps><uppercase><hidden>fmt</hidden></uppercase></smallcaps></sub></sup></s></u></i></b></p>` +
		`<p><a href="https://x.com">link</a></p>` +
		`<p><img alt="pic"/></p>` +
		`<p><br type="page"/><br type="textWrapping"/></p>` +
		`<p><fn-ref id="1" type="footnote">1</fn-ref></p>` +
		`<p><bm id="b1"/>text</p>` +
		`</write>` +
		`<notes>` +
		`<fn id="1"><p>footnote</p></fn>` +
		`<en id="1"><p>endnote</p></en>` +
		`<bm id="b1"><p>bookmark</p></bm>` +
		`<comment id="1"><p>comment</p></comment>` +
		`</notes>` +
		`<header id="1"><p>header</p></header>` +
		`<footer id="1"><p>footer</p></footer>` +
		`</words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWarnings(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<write><p><custom>unknown</custom></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid with warns, got errors: %v", r.Errors)
	}
	if len(r.Warns) == 0 {
		t.Error("expected warnings for unknown elements")
	}
}

func TestVerifyBreakClearValid(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<write><p><br type="clear"/></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyEnInvalid(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<notes><en id="1"><p>x</p></en></notes></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for <en> in notes")
	}
}

func TestVerifyEnRefInvalid(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<write><p><en-ref id="1"/></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid with warning, got errors: %v", r.Errors)
	}
	hasWarn := false
	for _, w := range r.Warns {
		if strings.Contains(w, "en-ref") {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("expected warning for unknown <en-ref> element")
	}
}

func TestVerifyTableWidth(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1" width="abc"><tr><td>x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "width must be number") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected width error, got: %v", r.Errors)
	}
}

func TestVerifyTableCellSpacing(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1" cellSpacing="abc"><tr><td>x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "cellSpacing must be number") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected cellSpacing error, got: %v", r.Errors)
	}
}

func TestVerifyTableInvalidAlign(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1" align="invalid"><tr><td>x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "align must be") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected align error, got: %v", r.Errors)
	}
}

func TestVerifyTableColSpec(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><colspec/><tr><td>x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "colspec") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected colspec error, got: %v", r.Errors)
	}
}

func TestVerifyTableUnexpectedChild(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><foo/><tr><td>x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unexpected") && strings.Contains(w, "foo") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for unexpected element, got warns: %v", r.Warns)
	}
}

func TestVerifyTableRowNoCell(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr><p>x</p></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "must contain at least one") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for row without cell, got: %v", r.Errors)
	}
}

func TestVerifyTrSpan(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr span="abc"><td>x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "span must be integer") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected span error, got: %v", r.Errors)
	}
}

func TestVerifyCellBadColspan(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr><td colspan="abc">x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "colspan must be integer") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected colspan error, got: %v", r.Errors)
	}
}

func TestVerifyCellBadValign(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr><td valign="bad">x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "valign must be") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected valign error, got: %v", r.Errors)
	}
}

func TestVerifyCellBadNoWrap(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr><td noWrap="yes">x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "noWrap must be") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected noWrap error, got: %v", r.Errors)
	}
}

func TestVerifyUnknownInlineElement(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<p><foo>x</foo></p>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unknown inline element") && strings.Contains(w, "foo") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for unknown inline, got warns: %v", r.Warns)
	}
}

func TestVerifyMetaUnknownChild(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<meta><foo>bar</foo></meta>` +
		`<style unit="in"></style><write><p>x</p></write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unknown") && strings.Contains(w, "<foo>") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for unknown meta child, got warns: %v", r.Warns)
	}
}

func TestVerifyStyleBadUnit(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="bad"></style><write><p>x</p></write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unit must be") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for bad unit, got warns: %v", r.Warns)
	}
}

func TestVerifyStyleNoUnit(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style></style><write><p>x</p></write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "missing unit") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for missing unit, got warns: %v", r.Warns)
	}
}

func TestVerifyNotesUnknownElement(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><foo id="1">x</foo></notes></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unknown element") && strings.Contains(w, "foo") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for unknown notes element, got warns: %v", r.Warns)
	}
}

func TestVerifyNoteItemMissingID(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><fn><p>x</p></fn></notes></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "missing required id") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for missing id, got: %v", r.Errors)
	}
}

func TestVerifyHeaderFooterBadID(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<header id="abc"><p>hdr</p></header></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "id must be integer") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for bad header id, got: %v", r.Errors)
	}
}

func TestVerifyFooterBadID(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<footer id="abc"><p>ftr</p></footer></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "id must be integer") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for bad footer id, got: %v", r.Errors)
	}
}

func TestVerifyHeadingLevels(t *testing.T) {
	for level := 4; level <= 9; level++ {
		tag := fmt.Sprintf("h%d", level)
		input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
			`<style unit="in"></style><write>` +
			`<` + tag + `>Title</` + tag + `>` +
			`</write></words>`
		r := Verify(input)
		if !r.Valid {
			t.Errorf("expected valid for <%s>, got errors: %v", tag, r.Errors)
		}
	}
}

func TestVerifyWriteTable(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr><th>x</th></tr><tr><td>y</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWritePre(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<pre>code block</pre>` +
		`</write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWritePreWithElement(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<pre><b>bad</b></pre>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unexpected element") && strings.Contains(w, "b") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for element inside pre, got warns: %v", r.Warns)
	}
}

func TestVerifyWriteBlockquote(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<blockquote><p>quoted text</p></blockquote>` +
		`</write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyLiInvalidType(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<ol><li type="bad">x</li></ol>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "type must be") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected li type error, got: %v", r.Errors)
	}
}

func TestVerifyComment(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><comment id="1" author="A" date="2024-01-01"><p>text</p></comment></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyCellBadRowspan(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<table id="1"><tr><td rowspan="abc">x</td></tr></table>` +
		`</write></words>`
	r := Verify(input)
	found := false
	for _, e := range r.Errors {
		if strings.Contains(e, "rowspan must be integer") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rowspan error, got: %v", r.Errors)
	}
}

func TestVerifyMissingWriteElement(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style></words>`
	r := Verify(input)
	if r.Valid {
		t.Error("expected invalid for missing write")
	}
}

func TestVerifyMissingStyleElement(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<write><p>x</p></write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "missing") && strings.Contains(w, "style") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for missing style, got warns: %v", r.Warns)
	}
}

func TestVerifyBlockContentFigure(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<figure><figcaption>Caption</figcaption></figure>` +
		`</write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyBlockContentHr(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write>` +
		`<p>x</p><hr/><p>y</p>` +
		`</write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyBlockContentTableInNotes(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><fn id="1"><table id="1"><tr><td>a</td></tr></table></fn></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyBlockContentListInNotes(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><fn id="1"><ul type="bullet"><li>a</li></ul></fn></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyBlockContentPreInNotes(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><fn id="1"><pre>code</pre></fn></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyBlockContentBlockquoteInNotes(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><fn id="1"><blockquote><p>q</p></blockquote></fn></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteWithH4(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><h4>Title</h4></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyParagraphValign(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p valign="center">x</p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyParagraphAlign(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p align="both">x</p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyChildrenUnknownBlock(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><unknown>x</unknown></write></words>`
	r := Verify(input)
	found := false
	for _, w := range r.Warns {
		if strings.Contains(w, "unknown element") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning for unknown block, got warns: %v", r.Warns)
	}
}

func TestVerifyWriteHeading(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><h1>Title</h1><h2>Sub</h2><h3>Sub2</h3></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteList(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><ul type="bullet"><li>x</li><li>y</li></ul></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteImg(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><img alt="test"/></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteFnRef(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>text<fn-ref id="1" type="footnote"/></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteSpan(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p><span font="Arial" size="12">text</span></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteAnchor(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p><a href="https://example.com">link</a></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteTab(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x<tab/>y</p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteSym(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x<sym char="F0B7"/>y</p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyWriteBm(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x<bm id="b1"/></p></write></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyCommentAttrs(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><comment id="1" author="John" date="2024-01-01T00:00:00Z"><p>comment</p></comment></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyNotesFootnote(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><fn id="1" type="footnote"><p>fn text</p></fn></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestVerifyNotesBookmark(t *testing.T) {
	input := `<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">` +
		`<style unit="in"></style><write><p>x</p></write>` +
		`<notes><bm id="b1"/></notes></words>`
	r := Verify(input)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}
