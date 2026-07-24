package docx

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"

	docs "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"
)

func EmitDocx(wordsXML string) ([]byte, error) {
	var doc WordsXML
	if err := xml.Unmarshal([]byte(wordsXML), &doc); err != nil {
		return nil, fmt.Errorf("parse words-xml: %w", err)
	}

	docx := docs.NewDocument()

	applyStylesToDocument(&doc, docx)
	applyHeadersFooters(&doc, docx)

	emitter := &coreEmitter{doc: &doc, docx: docx, numIdMap: make(map[string]int)}

	applyColumnStyles(&doc, emitter)
	numDefs := extractNumberingDefs(&doc)
	for i, def := range numDefs {
		key := fmt.Sprintf("%s:%d", def.numFmt, def.start)
		emitter.numIdMap[key] = i + 1
	}

	if doc.Write != nil {
		for _, item := range doc.Write.Content {
			data, err := json.Marshal(item)
			if err != nil {
				continue
			}
			if err := emitter.emitContent(data); err != nil {
				return nil, err
			}
		}
	}

	var docxBuf bytes.Buffer
	if _, err := docx.WriteTo(&docxBuf); err != nil {
		return nil, fmt.Errorf("write docx: %w", err)
	}

	processor := newZipProcessor(docxBuf.Bytes())

	if err := addNumbering(processor, &doc); err != nil {
		return nil, fmt.Errorf("add numbering: %w", err)
	}
	if err := addFootnotes(processor, &doc); err != nil {
		return nil, fmt.Errorf("add footnotes: %w", err)
	}
	if err := addComments(processor, &doc); err != nil {
		return nil, fmt.Errorf("add comments: %w", err)
	}
	if err := addStyles(processor, &doc); err != nil {
		return nil, fmt.Errorf("add styles: %w", err)
	}

	result, err := processor.process()
	if err != nil {
		return nil, fmt.Errorf("process zip: %w", err)
	}

	result, err = postProcessDocx(result, emitter.pendingMods)
	if err != nil {
		return nil, fmt.Errorf("post-process run props: %w", err)
	}

	result, err = postProcessInsDel(result, emitter.insTexts, emitter.delTexts, emitter.pendingMods)
	if err != nil {
		return nil, fmt.Errorf("post-process ins/del: %w", err)
	}

	return result, nil
}

func EmitDocxToFile(wordsXML, outputPath string) error {
	data, err := EmitDocx(wordsXML)
	if err != nil {
		return err
	}
	return writeFile(outputPath, data)
}

func EmitDocxToBytes(wordsXML string) ([]byte, error) {
	return EmitDocx(wordsXML)
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func addNumbering(processor *zipProcessor, doc *WordsXML) error {
	numDefs := extractNumberingDefs(doc)
	if len(numDefs) == 0 {
		return nil
	}
	xf := buildNumberingXML(numDefs)
	if xf == nil {
		return nil
	}
	data, err := xml.MarshalIndent(xf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal numbering: %w", err)
	}
	data = append([]byte(xml.Header), data...)
	processor.addPart("word/numbering.xml", data)
	return nil
}

func addFootnotes(processor *zipProcessor, doc *WordsXML) error {
	if doc.Notes == nil {
		return nil
	}
	var notes []NoteItemXML
	for _, item := range doc.Notes.Items {
		if item.Type == "footnote" || item.Type == "endnote" {
			notes = append(notes, NoteItemXML{
				ID:   item.ID,
				Type: item.Type,
			})
		}
	}
	if len(notes) == 0 {
		return nil
	}
	data := buildFootnotesXML(notes)
	if data != nil {
		processor.addPart("word/footnotes.xml", data)
	}
	return nil
}

func addComments(processor *zipProcessor, doc *WordsXML) error {
	if doc.Notes == nil {
		return nil
	}
	var notes []NoteItemXML
	for _, item := range doc.Notes.Items {
		if item.Type == "comment" {
			notes = append(notes, NoteItemXML{
				ID:     item.ID,
				Type:   item.Type,
				Author: item.Author,
				Date:   item.Date,
			})
		}
	}
	if len(notes) == 0 {
		return nil
	}
	data := buildCommentsXML(notes)
	if data != nil {
		processor.addPart("word/comments.xml", data)
	}
	return nil
}

func addStyles(processor *zipProcessor, doc *WordsXML) error {
	if doc.Style == nil {
		return nil
	}
	var theme *StyleTheme
	if doc.Style.Theme != nil {
		theme = doc.Style.Theme
	}
	data := buildStylesXML(theme, doc.Style.Customs)
	if data != nil {
		processor.addPart("word/styles.xml", data)
	}
	return nil
}

func extractNumberingDefs(doc *WordsXML) []numDef {
	defs := make(map[string]numDef)
	var walk func(content []interface{})
	walk = func(content []interface{}) {
		for _, item := range content {
			data, _ := xml.Marshal(item)
			var block BlockXML
			if err := xml.Unmarshal(data, &block); err != nil {
				continue
			}
			switch block.XMLName.Local {
			case "ul":
				key := "bullet:1"
				if _, ok := defs[key]; !ok {
					defs[key] = numDef{numFmt: "bullet", start: 1, levelText: "\u2022"}
				}
				var ul UlXML
				xml.Unmarshal(data, &ul)
				for _, item := range ul.Items {
					walk(item.Content)
				}
			case "ol":
				var ol OlXML
				xml.Unmarshal(data, &ol)
				numFmt := mapListFormat(ol.Type)
				start := ol.Start
				if start <= 0 {
					start = 1
				}
				key := fmt.Sprintf("%s:%d", numFmt, start)
				if _, ok := defs[key]; !ok {
					defs[key] = numDef{numFmt: numFmt, start: start, levelText: "%1."}
				}
				for _, item := range ol.Items {
					walk(item.Content)
				}
			}
		}
	}
	if doc.Write != nil {
		walk(doc.Write.Content)
	}
	result := make([]numDef, 0, len(defs))
	for _, d := range defs {
		result = append(result, d)
	}
	return result
}

type borderInfo struct {
	style string
	sz    int
	color string
	space int
}

func parseBorderAttribute(at string) map[string]borderInfo {
	result := make(map[string]borderInfo)
	if at == "" {
		return result
	}
	parts := strings.Split(at, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) < 2 {
			continue
		}
		side := fields[0]
		widthStr := fields[1]
		style := "single"
		color := "000000"
		space := 4

		if len(fields) >= 3 {
			s := fields[2]
			if idx := strings.IndexAny(s, "0123456789"); idx >= 0 {
				end := idx
				for end < len(s) && s[end] >= '0' && s[end] <= '9' {
					end++
				}
				style = s[:idx]
				if style == "" {
					style = "single"
				}
			} else {
				style = s
			}
		}
		if len(fields) >= 4 {
			c := fields[3]
			c = strings.TrimPrefix(c, "#")
			if len(c) == 6 {
				color = c
			}
		}

		w, _ := strconv.Atoi(widthStr)
		if w == 0 {
			w = 4
		}

		result[side] = borderInfo{
			style: style,
			sz:    w,
			color: color,
			space: space,
		}
	}
	return result
}

func applyStylesToDocument(doc *WordsXML, docx domain.Document) {
	if doc.Meta != nil {
		meta := &domain.Metadata{
			Title:       doc.Meta.Title,
			Subject:     doc.Meta.Subject,
			Creator:     doc.Meta.Author,
			Keywords:    strings.Split(doc.Meta.Keywords, ","),
			Description: doc.Meta.Description,
		}
		if doc.Meta.Created != "" {
			meta.Created = doc.Meta.Created
		}
		if doc.Meta.Modified != "" {
			meta.Modified = doc.Meta.Modified
		}
		docx.SetMetadata(meta)
	}

	if doc.Style == nil {
		return
	}

	section, err := docx.DefaultSection()
	if err != nil {
		return
	}

	for _, page := range doc.Style.Pages {
		if page.Size != "" {
			section.SetPageSize(mapPageSize(page.Size))
		} else if page.W > 0 && page.H > 0 {
			section.SetPageSize(domain.PageSize{
				Width:  int(page.W),
				Height: int(page.H),
			})
		}
		if page.MT > 0 || page.MB > 0 || page.ML > 0 || page.MR > 0 || page.MH > 0 || page.MF > 0 {
			section.SetMargins(domain.Margins{
				Top:    int(page.MT),
				Bottom: int(page.MB),
				Left:   int(page.ML),
				Right:  int(page.MR),
				Header: int(page.MH),
				Footer: int(page.MF),
			})
		}
	}

	for _, cols := range doc.Style.Cols {
		if cols.N > 1 {
			section.SetColumns(cols.N)
		}
	}
}

func applyColumnStyles(doc *WordsXML, emitter *coreEmitter) {
	if doc.Style == nil {
		return
	}

	for _, cols := range doc.Style.Cols {
		if cols.Space > 0 {
			emitter.pendingMods = append(emitter.pendingMods, pendingRunMod{
				text:    "",
				modType: "colSpace:" + strconv.FormatFloat(cols.Space, 'f', -1, 64),
			})
		}
	}

	for _, col := range doc.Style.Cols2 {
		if col.W > 0 {
			emitter.pendingMods = append(emitter.pendingMods, pendingRunMod{
				text:    "",
				modType: "colWidth:" + strconv.Itoa(col.Ref) + ":" + strconv.FormatFloat(col.W, 'f', -1, 64),
			})
		}
	}
}

func mapHeaderType(id int) domain.HeaderType {
	switch id {
	case 1:
		return domain.HeaderFirst
	case 2:
		return domain.HeaderEven
	default:
		return domain.HeaderDefault
	}
}

func mapFooterType(id int) domain.FooterType {
	switch id {
	case 1:
		return domain.FooterFirst
	case 2:
		return domain.FooterEven
	default:
		return domain.FooterDefault
	}
}

func applyHeadersFooters(doc *WordsXML, docx domain.Document) {
	if len(doc.Headers) == 0 && len(doc.Footers) == 0 {
		return
	}

	section, err := docx.DefaultSection()
	if err != nil {
		return
	}

	for _, hdr := range doc.Headers {
		headerType := mapHeaderType(hdr.ID)
		header, err := section.Header(headerType)
		if err != nil {
			continue
		}
		p, err := header.AddParagraph()
		if err != nil {
			continue
		}
		for _, item := range hdr.Content {
			data, _ := json.Marshal(item)
			var block BlockXML
			if err := xml.Unmarshal(data, &block); err != nil {
				continue
			}
			if block.XMLName.Local == "p" {
				for _, child := range block.Content {
					childData, _ := json.Marshal(child)
					var childBlock BlockXML
					if err := xml.Unmarshal(childData, &childBlock); err != nil {
						continue
					}
					if childBlock.XMLName.Local == "c" {
						p.SetStyle(extractText(childBlock.Content))
					} else {
						e := &coreEmitter{doc: doc, docx: docx}
						e.emitInline(child, p)
					}
				}
			} else {
				e := &coreEmitter{doc: doc, docx: docx}
				e.emitInline(item, p)
			}
		}
	}

	for _, ftr := range doc.Footers {
		footerType := mapFooterType(ftr.ID)
		footer, err := section.Footer(footerType)
		if err != nil {
			continue
		}
		p, err := footer.AddParagraph()
		if err != nil {
			continue
		}
		for _, item := range ftr.Content {
			data, _ := json.Marshal(item)
			var block BlockXML
			if err := xml.Unmarshal(data, &block); err != nil {
				continue
			}
			if block.XMLName.Local == "p" {
				for _, child := range block.Content {
					childData, _ := json.Marshal(child)
					var childBlock BlockXML
					if err := xml.Unmarshal(childData, &childBlock); err != nil {
						continue
					}
					if childBlock.XMLName.Local == "c" {
						p.SetStyle(extractText(childBlock.Content))
					} else {
						e := &coreEmitter{doc: doc, docx: docx}
						e.emitInline(child, p)
					}
				}
			} else {
				e := &coreEmitter{doc: doc, docx: docx}
				e.emitInline(item, p)
			}
		}
	}
}

func applyBorderToParagraph(p domain.Paragraph, borders map[string]borderInfo) {
	if len(borders) == 0 {
		return
	}
	b := domain.ParagraphBorders{}
	if top, ok := borders["top"]; ok {
		b.Top = domain.BorderStyle{
			Style: mapBorderStyleEnum(top.style),
			Width: top.sz,
			Color: parseColorToDomain(top.color),
		}
	}
	if bottom, ok := borders["bottom"]; ok {
		b.Bottom = domain.BorderStyle{
			Style: mapBorderStyleEnum(bottom.style),
			Width: bottom.sz,
			Color: parseColorToDomain(bottom.color),
		}
	}
	if left, ok := borders["left"]; ok {
		b.Left = domain.BorderStyle{
			Style: mapBorderStyleEnum(left.style),
			Width: left.sz,
			Color: parseColorToDomain(left.color),
		}
	}
	if right, ok := borders["right"]; ok {
		b.Right = domain.BorderStyle{
			Style: mapBorderStyleEnum(right.style),
			Width: right.sz,
			Color: parseColorToDomain(right.color),
		}
	}
	p.SetBorders(b)
}

func parseColorToDomain(color string) domain.Color {
	color = strings.TrimPrefix(color, "#")
	if len(color) == 6 {
		r, _ := strconv.ParseUint(color[0:2], 16, 8)
		g, _ := strconv.ParseUint(color[2:4], 16, 8)
		b, _ := strconv.ParseUint(color[4:6], 16, 8)
		return domain.Color{R: uint8(r), G: uint8(g), B: uint8(b)}
	}
	return domain.Color{}
}

func applyStyleRules(p domain.Paragraph, style *StyleXML, tagName, styleName string) {
	if style == nil {
		return
	}

	for _, gap := range style.Gaps {
		if matchStyleSelector(gap.EL, gap.C, tagName, styleName) {
			if gap.Before > 0 {
				p.SetSpacingBefore(int(gap.Before))
			}
			if gap.After > 0 {
				p.SetSpacingAfter(int(gap.After))
			}
		}
	}

	for _, line := range style.Lines {
		if matchStyleSelector(line.EL, line.C, tagName, styleName) {
			if line.Value > 0 {
				p.SetLineSpacing(domain.LineSpacing{
					Rule:  mapLineRule(line.Rule),
					Value: int(line.Value),
				})
			}
		}
	}

	for _, indent := range style.Indents {
		if matchStyleSelector(indent.EL, indent.C, tagName, styleName) {
			ind := p.Indent()
			if indent.Left > 0 {
				ind.Left = int(indent.Left)
			}
			if indent.Right > 0 {
				ind.Right = int(indent.Right)
			}
			if indent.First > 0 {
				ind.FirstLine = int(indent.First)
			}
			if indent.Hanging > 0 {
				ind.Hanging = int(indent.Hanging)
			}
			p.SetIndent(ind)
		}
	}

	for _, align := range style.Aligns {
		if matchStyleSelector(align.EL, align.C, tagName, styleName) {
			p.SetAlignment(mapAlignment(align.Value))
		}
	}
}

func collectMatchingTabs(style *StyleXML, tagName, styleName string) []StyleTab {
	if style == nil {
		return nil
	}
	var result []StyleTab
	for _, tab := range style.Tabs {
		if matchStyleSelector(tab.EL, tab.C, tagName, styleName) {
			result = append(result, tab)
		}
	}
	return result
}

func matchStyleSelector(el, class, tagName, styleName string) bool {
	if el != "" && el != tagName {
		return false
	}
	if class != "" && styleName != class {
		return false
	}
	return true
}

func mapLineRule(rule string) domain.LineSpacingRule {
	switch strings.ToLower(rule) {
	case "auto":
		return domain.LineSpacingAuto
	case "exact":
		return domain.LineSpacingExact
	case "atleast":
		return domain.LineSpacingAtLeast
	default:
		return domain.LineSpacingAuto
	}
}

func mapBorderStyleEnum(style string) domain.BorderLineStyle {
	switch strings.ToLower(style) {
	case "single":
		return domain.BorderSingle
	case "double":
		return domain.BorderDouble
	case "dashed":
		return domain.BorderDashed
	case "dotted":
		return domain.BorderDotted
	default:
		return domain.BorderSingle
	}
}

func mapBorderStyle(style string) string {
	switch strings.ToLower(style) {
	case "single", "s":
		return "single"
	case "double", "d":
		return "double"
	case "dashed", "ds":
		return "dashed"
	case "dotted", "dt":
		return "dotted"
	case "none", "n":
		return "none"
	default:
		return "single"
	}
}

func parseUnderlineStyle(style string) string {
	switch strings.ToLower(style) {
	case "single":
		return "single"
	case "double":
		return "double"
	case "thick":
		return "thick"
	case "dotted":
		return "dotted"
	case "wave":
		return "wave"
	default:
		return ""
	}
}

func mapTabAlign(align string) string {
	switch strings.ToLower(align) {
	case "left":
		return "left"
	case "center":
		return "center"
	case "right":
		return "right"
	case "decimal":
		return "decimal"
	default:
		return "left"
	}
}

func mapLeader(leader string) string {
	switch strings.ToLower(leader) {
	case "dot":
		return "dot"
	case "hyphen":
		return "hyphen"
	case "underscore":
		return "underscore"
	case "heavy":
		return "heavy"
	case "underline":
		return "underline"
	default:
		return ""
	}
}

func parseSelector(elements string, classes string) []string {
	var rules []string
	if elements != "" {
		for _, el := range strings.Split(elements, ",") {
			rules = append(rules, strings.TrimSpace(el))
		}
	}
	if classes != "" {
		for _, c := range strings.Split(classes, ",") {
			rules = append(rules, "."+strings.TrimSpace(c))
		}
	}
	return rules
}

func matchSelector(rules []string, p domain.Paragraph) bool {
	return len(rules) > 0
}

func getAttr(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func parseTableBorders(at string) domain.TableBorders {
	attrs := parseBorderAttribute(at)
	borders := domain.TableBorders{}
	if top, ok := attrs["top"]; ok {
		borders.Top = domain.BorderStyle{
			Style: mapBorderStyleEnum(top.style),
			Width: top.sz,
			Color: parseColor("#" + top.color),
		}
	}
	if left, ok := attrs["left"]; ok {
		borders.Left = domain.BorderStyle{
			Style: mapBorderStyleEnum(left.style),
			Width: left.sz,
			Color: parseColor("#" + left.color),
		}
	}
	if bottom, ok := attrs["bottom"]; ok {
		borders.Bottom = domain.BorderStyle{
			Style: mapBorderStyleEnum(bottom.style),
			Width: bottom.sz,
			Color: parseColor("#" + bottom.color),
		}
	}
	if right, ok := attrs["right"]; ok {
		borders.Right = domain.BorderStyle{
			Style: mapBorderStyleEnum(right.style),
			Width: right.sz,
			Color: parseColor("#" + right.color),
		}
	}
	return borders
}

func mapTableAlignment(align string) domain.Alignment {
	switch strings.ToLower(align) {
	case "center":
		return domain.AlignmentCenter
	case "right":
		return domain.AlignmentRight
	case "justify", "distribute":
		return domain.AlignmentDistribute
	default:
		return domain.AlignmentLeft
	}
}
