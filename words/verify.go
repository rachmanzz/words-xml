package words

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type VerifyResult struct {
	Valid  bool
	Errors []string
	Warns  []string
}

func (v *VerifyResult) addError(format string, args ...interface{}) {
	v.Errors = append(v.Errors, fmt.Sprintf(format, args...))
	v.Valid = false
}

func (v *VerifyResult) addWarn(format string, args ...interface{}) {
	v.Warns = append(v.Warns, fmt.Sprintf(format, args...))
}

func Verify(wordsXML string) VerifyResult {
	r := VerifyResult{Valid: true}

	if strings.TrimSpace(wordsXML) == "" {
		r.addError("empty input")
		return r
	}

	decoder := xml.NewDecoder(strings.NewReader(wordsXML))
	decoder.Strict = false

	var tokens []xml.Token
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err != io.EOF {
				r.addError("XML parse error: %v", err)
			}
			break
		}
		tokens = append(tokens, tok)
	}

	if len(tokens) == 0 {
		r.addError("no XML tokens found")
		return r
	}

	verifyRoot(tokens, &r)
	verifyChildren(tokens, &r)
	return r
}

func verifyRoot(tokens []xml.Token, r *VerifyResult) {
	start, ok := tokens[0].(xml.StartElement)
	if !ok {
		r.addError("root element must be <words>, got %T", tokens[0])
		return
	}

	if start.Name.Local != "words" {
		r.addError("root element must be <words>, got <%s>", start.Name.Local)
	}

	nsOK := false
	for _, a := range start.Attr {
		if a.Name.Local == "xmlns" && a.Value == "urn:words:v1" {
			nsOK = true
		}
	}
	if !nsOK {
		r.addError("missing xmlns=\"urn:words:v1\"")
	}

	version := ""
	mode := ""
	for _, a := range start.Attr {
		if a.Name.Local == "version" {
			version = a.Value
		}
		if a.Name.Local == "mode" {
			mode = a.Value
		}
	}

	if version == "" {
		r.addError("missing version attribute on <words>")
	} else if version != "1.0.1" {
		r.addError("version must be \"1.0.1\", got %q", version)
	}

	if mode == "" {
		r.addWarn("missing mode attribute on <words>")
	} else if mode != "semantic" && mode != "lossless" {
		r.addError("mode must be \"semantic\" or \"lossless\", got %q", mode)
	}
}

func verifyChildren(tokens []xml.Token, r *VerifyResult) {
	if len(tokens) < 2 {
		return
	}

	root := tokens[1 : len(tokens)-1]

	hasWrite := false
	hasStyle := false
	seenUnique := map[string]bool{}

	for i := 0; i < len(root); i++ {
		tok := root[i]
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		local := start.Name.Local

		switch local {
		case "meta":
			if seenUnique["meta"] {
				r.addError("duplicate <%s> element", local)
			}
			seenUnique["meta"] = true
			verifyMeta(root, &i, r)
		case "style":
			if seenUnique["style"] {
				r.addError("duplicate <%s> element", local)
			}
			seenUnique["style"] = true
			hasStyle = true
			verifyStyle(root, start, &i, r)
		case "write":
			if seenUnique["write"] {
				r.addError("duplicate <%s> element", local)
			}
			seenUnique["write"] = true
			hasWrite = true
			verifyWrite(root, &i, r)
		case "header":
			verifyHeaderFooter(root, &i, r)
		case "footer":
			verifyHeaderFooter(root, &i, r)
		case "notes":
			if seenUnique["notes"] {
				r.addError("duplicate <%s> element", local)
			}
			seenUnique["notes"] = true
			verifyNotes(root, &i, r)
		default:
			r.addWarn("unexpected top-level element <%s>", local)
		}
	}

	if !hasWrite {
		r.addError("missing required <write> element")
	}
	if !hasStyle {
		r.addWarn("missing <style> element")
	}
}

func verifyMeta(tokens []xml.Token, idx *int, r *VerifyResult) {
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<meta> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	validMeta := map[string]bool{"title": true, "author": true, "created": true, "modified": true, "keywords": true, "subject": true, "description": true}

	for _, tok := range section {
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := start.Name.Local
		if !validMeta[local] {
			r.addWarn("unknown <meta> child <%s>", local)
		}
	}
	*idx = endIdx
}

func verifyStyle(tokens []xml.Token, start xml.StartElement, idx *int, r *VerifyResult) {
	unit := ""
	for _, a := range start.Attr {
		if a.Name.Local == "unit" {
			unit = a.Value
		}
	}
	if unit == "" {
		r.addWarn("<style> missing unit attribute")
	} else if unit != "in" && unit != "cm" && unit != "pt" && unit != "px" {
		r.addWarn("<style> unit must be in|cm|pt|px, got %q", unit)
	}

	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<style> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]

	validStyle := map[string]bool{
		"page": true, "gap": true, "custom": true, "col": true,
		"cols": true, "align": true, "line": true, "tab": true,
		"theme": true, "indent": true,
	}

	for _, tok := range section {
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := start.Name.Local
		if !validStyle[local] {
			r.addWarn("unknown element <%s> inside <style>", local)
		}
	}

	*idx = endIdx
}

func verifyWrite(tokens []xml.Token, idx *int, r *VerifyResult) {
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<write> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyBlockContent(section, r, "write")
	*idx = endIdx
}

func verifyBlockContent(tokens []xml.Token, r *VerifyResult, context string) {
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := start.Name.Local

		switch local {
		case "p":
			verifyParagraph(tokens, &i, r)
		case "h1", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9":
			verifyParagraph(tokens, &i, r)
		case "table":
			verifyTable(start, tokens, &i, r)
		case "ul", "ol":
			verifyList(start, tokens, &i, r)
		case "pre":
			verifyPre(tokens, &i, r)
		case "blockquote":
			verifyBlockquote(tokens, &i, r)
		case "figure":
			figEnd := findMatchingEnd(tokens, i)
			if figEnd < 0 {
				r.addError("<figure> has no matching end tag")
			} else {
				verifyBlockContent(tokens[i+1:figEnd], r, "figure")
				i = figEnd
			}
		case "figcaption":
			// inline only, ok
		case "img":
			verifyImg(start, r)
		case "br":
			verifyBreak(start, r)
		case "hr":
			// ok
		case "b", "i", "u", "s", "sup", "sub", "smallcaps", "uppercase", "hidden", "bcs", "ics":
			verifyInlineAttrs(start, local, r)
		case "a":
			verifyAnchor(start, r)
		case "fn-ref", "bm":
			// ok
		case "span":
			verifySpan(start, r)
		case "sym", "tab":
			// ok
		default:
			r.addWarn("unknown element <%s> in %s", local, context)
		}
	}
}

func verifyParagraph(tokens []xml.Token, idx *int, r *VerifyResult) {
	start := tokens[*idx].(xml.StartElement)
	verifyInlineAttrs(start, "p", r)

	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<%s> has no matching end tag", start.Name.Local)
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyInlineContent(section, r, start.Name.Local)
	*idx = endIdx
}

func verifyTable(start xml.StartElement, tokens []xml.Token, idx *int, r *VerifyResult) {
	verifyTableAttrs(start, r)
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<table> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]

	for i := 0; i < len(section); i++ {
		tok := section[i]
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if start.Name.Local == "tr" {
			verifyTableRow(start, section, &i, r)
		} else if start.Name.Local == "colspec" {
			r.addError("<colspec> is not valid inside <table>; use <s:col> in <style> instead")
		} else {
			r.addWarn("unexpected element <%s> inside <table>", start.Name.Local)
		}
	}

	*idx = endIdx
}

func verifyTableAttrs(start xml.StartElement, r *VerifyResult) {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "id":
			if _, err := strconv.Atoi(a.Value); err != nil {
				r.addError("<table> id must be integer, got %q", a.Value)
			}
		case "width":
			if _, err := strconv.ParseFloat(a.Value, 64); err != nil {
				r.addError("<table> width must be number, got %q", a.Value)
			}
		case "align":
			valid := map[string]bool{"left": true, "center": true, "right": true}
			if !valid[a.Value] {
				r.addError("<table> align must be left|center|right, got %q", a.Value)
			}
		case "cellSpacing":
			if _, err := strconv.ParseFloat(a.Value, 64); err != nil {
				r.addError("<table> cellSpacing must be number, got %q", a.Value)
			}
		}
	}
}

func verifyTableRow(start xml.StartElement, tokens []xml.Token, idx *int, r *VerifyResult) {
	verifyTrAttrs(start, r)
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<tr> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	hasCell := false
	for i := 0; i < len(section); i++ {
		tok := section[i]
		s, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if s.Name.Local == "td" || s.Name.Local == "th" {
			hasCell = true
			verifyCell(s, section, &i, r)
		}
	}
	if !hasCell {
		r.addError("<tr> must contain at least one <td> or <th>")
	}
	*idx = endIdx
}

func verifyTrAttrs(start xml.StartElement, r *VerifyResult) {
	for _, a := range start.Attr {
		if a.Name.Local == "span" {
			if _, err := strconv.Atoi(a.Value); err != nil {
				r.addError("<tr> span must be integer, got %q", a.Value)
			}
		}
	}
}

func verifyCell(start xml.StartElement, tokens []xml.Token, idx *int, r *VerifyResult) {
	verifyCellAttrs(start, r)
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<%s> has no matching end tag", start.Name.Local)
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyBlockContent(section, r, start.Name.Local)
	*idx = endIdx
}

func verifyCellAttrs(start xml.StartElement, r *VerifyResult) {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "colspan", "rowspan":
			if _, err := strconv.Atoi(a.Value); err != nil {
				r.addError("<%s> %s must be integer, got %q", start.Name.Local, a.Name.Local, a.Value)
			}
		case "valign":
			valid := map[string]bool{"top": true, "center": true, "bottom": true}
			if !valid[a.Value] {
				r.addError("<%s> valign must be top|center|bottom, got %q", start.Name.Local, a.Value)
			}
		case "noWrap":
			if a.Value != "true" && a.Value != "false" {
				r.addError("<%s> noWrap must be true|false, got %q", start.Name.Local, a.Value)
			}
		}
	}
}

func verifyList(start xml.StartElement, tokens []xml.Token, idx *int, r *VerifyResult) {
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<%s> has no matching end tag", start.Name.Local)
		return
	}
	section := tokens[*idx+1 : endIdx]
	hasLi := false
	for i := 0; i < len(section); i++ {
		tok := section[i]
		s, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if s.Name.Local == "li" {
			hasLi = true
			verifyListItem(s, section, &i, r)
		} else {
			r.addWarn("unexpected element <%s> inside <%s>", s.Name.Local, start.Name.Local)
		}
	}
	if !hasLi {
		r.addError("<%s> must contain at least one <li>", start.Name.Local)
	}
	*idx = endIdx
}

func verifyListItem(start xml.StartElement, tokens []xml.Token, idx *int, r *VerifyResult) {
	verifyLiAttrs(start, r)
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<li> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyBlockContent(section, r, "li")
	*idx = endIdx
}

func verifyLiAttrs(start xml.StartElement, r *VerifyResult) {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "tag":
			// ok, freeform
		case "type":
			valid := map[string]bool{
				"decimal": true, "lowerLetter": true, "upperLetter": true,
				"lowerRoman": true, "upperRoman": true, "bullet": true,
			}
			if !valid[a.Value] {
				r.addError("<li> type must be decimal|lowerLetter|upperLetter|lowerRoman|upperRoman|bullet, got %q", a.Value)
			}
		}
	}
}

func verifyPre(tokens []xml.Token, idx *int, r *VerifyResult) {
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<pre> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	for _, tok := range section {
		if s, ok := tok.(xml.StartElement); ok {
			r.addWarn("unexpected element <%s> inside <pre>", s.Name.Local)
		}
	}
	*idx = endIdx
}

func verifyBlockquote(tokens []xml.Token, idx *int, r *VerifyResult) {
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<blockquote> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyBlockContent(section, r, "blockquote")
	*idx = endIdx
}

func verifyImg(start xml.StartElement, r *VerifyResult) {
	hasAlt := false
	for _, a := range start.Attr {
		if a.Name.Local == "alt" {
			hasAlt = true
		}
		if a.Name.Local == "src" {
			r.addError("<img> must not have src attribute (only alt is allowed)")
		}
		if a.Name.Local == "width" || a.Name.Local == "height" {
			r.addError("<img> must not have %s attribute (only alt is allowed)", a.Name.Local)
		}
	}
	if !hasAlt {
		r.addError("<img> missing required alt attribute")
	}
}

func verifyBreak(start xml.StartElement, r *VerifyResult) {
	hasType := false
	for _, a := range start.Attr {
		if a.Name.Local == "type" {
			hasType = true
			valid := map[string]bool{"page": true, "textWrapping": true, "column": true, "clear": true}
			if !valid[a.Value] {
				r.addError("<br> type must be page|textWrapping|column, got %q", a.Value)
			}
		}
	}
	if !hasType {
		r.addError("<br> missing required type attribute")
	}
}

func verifyInlineAttrs(start xml.StartElement, elemType string, r *VerifyResult) {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "id":
			// ok
		case "lang":
			// ok, BCP 47
		case "dir":
			if a.Value != "ltr" && a.Value != "rtl" {
				r.addError("<%s> dir must be ltr|rtl, got %q", elemType, a.Value)
			}
		case "class":
			// ok
		case "align":
			valid := map[string]bool{"center": true, "both": true, "right": true}
			if !valid[a.Value] {
				r.addError("<%s> align must be center|both|right, got %q", elemType, a.Value)
			}
		case "indentLeft", "indentHanging", "indentRight", "indentFirst":
			if _, err := strconv.ParseFloat(a.Value, 64); err != nil {
				r.addError("<%s> %s must be number, got %q", elemType, a.Name.Local, a.Value)
			}
		case "valign":
			valid := map[string]bool{"top": true, "center": true, "bottom": true, "baseline": true}
			if !valid[a.Value] {
				r.addError("<%s> valign must be top|center|bottom|baseline, got %q", elemType, a.Value)
			}
		case "c":
			// ok, custom style name
		case "at":
			// ok, border attribute
		}
	}
}

func verifyInlineContent(tokens []xml.Token, r *VerifyResult, context string) {
	for _, tok := range tokens {
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := start.Name.Local
		switch local {
		case "b", "i", "u", "s", "sup", "sub", "smallcaps", "uppercase", "hidden", "bcs", "ics":
			verifyInlineElement(tokens, &start, r)
		case "a":
			verifyAnchor(start, r)
		case "fn-ref", "bm":
			// ok
		case "img":
			verifyImg(start, r)
		case "br":
			verifyBreak(start, r)
		case "tab":
			// ok
		case "span":
			verifySpan(start, r)
		case "sym":
			// ok
		default:
			r.addWarn("unknown inline element <%s> in <%s>", local, context)
		}
	}
}

func verifyInlineElement(tokens []xml.Token, start *xml.StartElement, r *VerifyResult) {
	verifyInlineAttrs(*start, start.Name.Local, r)
}

func verifyAnchor(start xml.StartElement, r *VerifyResult) {
	hasHref := false
	for _, a := range start.Attr {
		if a.Name.Local == "href" {
			hasHref = true
			if a.Value == "" {
				r.addError("<a> href must not be empty")
			}
		}
	}
	if !hasHref {
		r.addError("<a> missing required href attribute")
	}
}

func verifySpan(start xml.StartElement, r *VerifyResult) {
	verifyInlineAttrs(start, "span", r)
}

func verifyHeaderFooter(tokens []xml.Token, idx *int, r *VerifyResult) {
	start := tokens[*idx].(xml.StartElement)
	hasID := false
	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			hasID = true
			if _, err := strconv.Atoi(a.Value); err != nil {
				r.addError("<%s> id must be integer, got %q", start.Name.Local, a.Value)
			}
		}
	}
	if !hasID {
		r.addError("<%s> missing required id attribute", start.Name.Local)
	}
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<%s> has no matching end tag", start.Name.Local)
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyBlockContent(section, r, start.Name.Local)
	*idx = endIdx
}

func verifyNotes(tokens []xml.Token, idx *int, r *VerifyResult) {
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<notes> has no matching end tag")
		return
	}
	section := tokens[*idx+1 : endIdx]
	for i := 0; i < len(section); i++ {
		tok := section[i]
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := start.Name.Local
		switch local {
		case "fn", "bm", "comment":
			verifyNoteItem(start, section, &i, r)
		default:
			r.addWarn("unknown element <%s> inside <notes>", local)
		}
	}
	*idx = endIdx
}

func verifyNoteItem(start xml.StartElement, tokens []xml.Token, idx *int, r *VerifyResult) {
	hasID := false
	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			hasID = true
		}
	}
	if !hasID {
		r.addError("<%s> missing required id attribute", start.Name.Local)
	}
	endIdx := findMatchingEnd(tokens, *idx)
	if endIdx < 0 {
		r.addError("<%s> has no matching end tag", start.Name.Local)
		return
	}
	section := tokens[*idx+1 : endIdx]
	verifyBlockContent(section, r, start.Name.Local)
	*idx = endIdx
}

func findMatchingEnd(tokens []xml.Token, startIdx int) int {
	depth := 1
	for i := startIdx + 1; i < len(tokens); i++ {
		switch tokens[i].(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
