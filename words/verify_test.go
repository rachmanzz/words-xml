package words

import (
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
