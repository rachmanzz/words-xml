package docx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

type coreEmitter struct {
	doc           *WordsXML
	docx          domain.Document
	paragraph     domain.Paragraph
	run           domain.Run
	listDepth     int
	pendingMods   []pendingRunMod
	insTexts      []string
	delTexts      []string
	numIdMap      map[string]int
}

func newCoreEmitter(doc *WordsXML) *coreEmitter {
	return &coreEmitter{doc: doc}
}

func (e *coreEmitter) addParagraph() (domain.Paragraph, error) {
	return e.docx.AddParagraph()
}

func (e *coreEmitter) emitContent(data []byte) error {
	var block BlockXML
	if err := xml.Unmarshal(data, &block); err != nil {
		return nil
	}

	switch block.XMLName.Local {
	case "p":
		return e.emitPara(data, "")
	case "h1", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9":
		return e.emitPara(data, block.XMLName.Local)
	case "blockquote":
		return e.emitBlockquote(data)
	case "pre":
		return e.emitPre(data)
	case "ul":
		return e.emitUl(block.Content)
	case "ol":
		return e.emitOl(block.Content)
	case "table":
		return e.emitTable(data)
	case "hr":
		return e.emitHr()
	case "img":
		return e.emitImg(block.Content)
	case "br":
		return e.emitBr()
	default:
		return e.emitGenericBlock(block.Content, block.XMLName.Local)
	}
}

func (e *coreEmitter) emitPara(data []byte, heading string) error {
	var para ParaXML
	if err := xml.Unmarshal(data, &para); err != nil {
		return nil
	}

	p, err := e.addParagraph()
	if err != nil {
		return err
	}

	if heading != "" {
		p.SetStyle(heading)
	}

	if para.Style != "" {
		p.SetStyle(para.Style)
	}

	if para.Lang != "" {
		e.injectLang(p, para.Lang)
	}
	if para.Dir == "rtl" {
		e.injectBidi(p)
	}
	if para.At != "" {
		borders := parseBorderAttribute(para.At)
		applyBorderToParagraph(p, borders)
	}
	if para.VAlign != "" {
		e.injectParaValign(p, para.VAlign)
	}

	if e.doc.Style != nil {
		styleName := ""
		for _, item := range para.Content {
			itemData, _ := json.Marshal(item)
			var block BlockXML
			if err := xml.Unmarshal(itemData, &block); err == nil {
				if block.XMLName.Local == "c" {
					styleName = extractText(block.Content)
					break
				}
			}
		}
		applyStyleRules(p, e.doc.Style, heading, styleName)

		tabs := collectMatchingTabs(e.doc.Style, heading, styleName)
		if len(tabs) > 0 {
			e.pendingMods = append(e.pendingMods, pendingRunMod{
				text:    "",
				modType: "tabstops",
				tabs:    tabs,
			})
		}
	}

	for _, item := range para.Content {
		itemData, _ := json.Marshal(item)
		var block BlockXML
		if err := xml.Unmarshal(itemData, &block); err != nil {
			continue
		}
		switch block.XMLName.Local {
		case "c":
			p.SetStyle(extractText(block.Content))
		case "align":
			p.SetAlignment(mapAlignment(extractText(block.Content)))
		case "indentLeft":
			if v, err := strconv.ParseFloat(extractText(block.Content), 64); err == nil {
				indent := p.Indent()
				indent.Left = int(v)
				p.SetIndent(indent)
			}
		case "indentRight":
			if v, err := strconv.ParseFloat(extractText(block.Content), 64); err == nil {
				indent := p.Indent()
				indent.Right = int(v)
				p.SetIndent(indent)
			}
		case "indentFirst":
			if v, err := strconv.ParseFloat(extractText(block.Content), 64); err == nil {
				indent := p.Indent()
				indent.FirstLine = int(v)
				p.SetIndent(indent)
			}
		case "indentHanging":
			if v, err := strconv.ParseFloat(extractText(block.Content), 64); err == nil {
				indent := p.Indent()
				indent.Hanging = int(v)
				p.SetIndent(indent)
			}
		case "gap":
			if v, err := strconv.ParseFloat(extractText(block.Content), 64); err == nil {
				p.SetSpacingBefore(int(v))
			}
		case "line":
			if v, err := strconv.ParseFloat(extractText(block.Content), 64); err == nil {
				p.SetLineSpacing(domain.LineSpacing{
					Rule:  domain.LineSpacingAuto,
					Value: int(v),
				})
			}
		default:
			e.emitInline(item, p)
		}
	}
	return nil
}

func (e *coreEmitter) injectLang(p domain.Paragraph, lang string) {
	e.pendingMods = append(e.pendingMods, pendingRunMod{
		text:    "",
		modType: "lang:" + lang,
	})
}

func (e *coreEmitter) injectBidi(p domain.Paragraph) {
	e.pendingMods = append(e.pendingMods, pendingRunMod{
		text:    "",
		modType: "bidi",
	})
}

func (e *coreEmitter) injectParaValign(p domain.Paragraph, valign string) {
	e.pendingMods = append(e.pendingMods, pendingRunMod{
		text:    "",
		modType: "pvalign:" + valign,
	})
}

func (e *coreEmitter) emitBlockquote(data []byte) error {
	var bq BlockquoteXML
	if err := xml.Unmarshal(data, &bq); err != nil {
		return e.emitGenericBlock(nil, "blockquote")
	}

	p, err := e.addParagraph()
	if err != nil {
		return err
	}

	if bq.Style != "" {
		p.SetStyle(bq.Style)
	} else {
		p.SetStyle("Quote")
	}

	if bq.Lang != "" {
		e.injectLang(p, bq.Lang)
	}
	if bq.Dir == "rtl" {
		e.injectBidi(p)
	}
	if bq.At != "" {
		borders := parseBorderAttribute(bq.At)
		applyBorderToParagraph(p, borders)
	}
	if bq.Align != "" {
		p.SetAlignment(mapAlignment(bq.Align))
	}
	if bq.VAlign != "" {
		e.injectParaValign(p, bq.VAlign)
	}

	if bq.IndentLeft > 0 || bq.IndentRight > 0 || bq.IndentFirst > 0 || bq.IndentHanging > 0 {
		ind := p.Indent()
		if bq.IndentLeft > 0 {
			ind.Left = int(bq.IndentLeft)
		}
		if bq.IndentRight > 0 {
			ind.Right = int(bq.IndentRight)
		}
		if bq.IndentFirst > 0 {
			ind.FirstLine = int(bq.IndentFirst)
		}
		if bq.IndentHanging > 0 {
			ind.Hanging = int(bq.IndentHanging)
		}
		p.SetIndent(ind)
	}

	for _, item := range bq.Content {
		e.emitInline(item, p)
	}
	return nil
}

func (e *coreEmitter) emitPre(data []byte) error {
	var pre PreXML
	if err := xml.Unmarshal(data, &pre); err != nil {
		return e.emitGenericBlock(nil, "pre")
	}

	p, err := e.addParagraph()
	if err != nil {
		return err
	}

	if pre.Style != "" {
		p.SetStyle(pre.Style)
	} else {
		p.SetStyle("NoSpacing")
	}

	if pre.Align != "" {
		p.SetAlignment(mapAlignment(pre.Align))
	}
	if pre.IndentLeft > 0 || pre.IndentRight > 0 || pre.IndentFirst > 0 || pre.IndentHanging > 0 {
		ind := p.Indent()
		if pre.IndentLeft > 0 {
			ind.Left = int(pre.IndentLeft)
		}
		if pre.IndentRight > 0 {
			ind.Right = int(pre.IndentRight)
		}
		if pre.IndentFirst > 0 {
			ind.FirstLine = int(pre.IndentFirst)
		}
		if pre.IndentHanging > 0 {
			ind.Hanging = int(pre.IndentHanging)
		}
		p.SetIndent(ind)
	}

	for _, item := range pre.Content {
		e.emitInline(item, p)
	}
	return nil
}

func (e *coreEmitter) emitUl(content []interface{}) error {
	e.listDepth++
	defer func() { e.listDepth-- }()

	var ul UlXML
	if len(content) > 0 {
		data, _ := json.Marshal(content[0])
		if len(data) > 0 {
			xml.Unmarshal(data, &ul)
		}
	}

	numFmt := "bullet"
	if ul.Type != "" {
		numFmt = mapListFormat(ul.Type)
	}
	numId := e.numIdMap[numFmt+":1"]

	for _, item := range content {
		data, _ := json.Marshal(item)
		var li LiXML
		if err := xml.Unmarshal(data, &li); err != nil {
			continue
		}
		e.emitLi(li.Content, numId, e.listDepth-1)
	}
	return nil
}

func (e *coreEmitter) emitOl(content []interface{}) error {
	e.listDepth++
	defer func() { e.listDepth-- }()

	var ol OlXML
	if len(content) > 0 {
		data, _ := json.Marshal(content[0])
		if len(data) > 0 {
			xml.Unmarshal(data, &ol)
		}
	}

	numFmt := mapListFormat(ol.Type)
	start := ol.Start
	if start <= 0 {
		start = 1
	}
	key := fmt.Sprintf("%s:%d", numFmt, start)
	numId := e.numIdMap[key]
	if numId == 0 {
		numId = e.numIdMap["decimal:1"]
	}

	for _, item := range content {
		data, _ := json.Marshal(item)
		var li LiXML
		if err := xml.Unmarshal(data, &li); err != nil {
			continue
		}
		e.emitLi(li.Content, numId, e.listDepth-1)
	}
	return nil
}

func (e *coreEmitter) emitLi(content []interface{}, numId, level int) error {
	p, err := e.addParagraph()
	if err != nil {
		return err
	}
	p.SetStyle("ListParagraph")

	if numId > 0 {
		p.SetNumbering(domain.NumberingReference{
			ID:    numId,
			Level: level,
		})
	}

	for _, item := range content {
		data, _ := json.Marshal(item)
		var block BlockXML
		if err := xml.Unmarshal(data, &block); err != nil {
			continue
		}
		switch block.XMLName.Local {
		case "br":
			run, _ := p.AddRun()
			run.AddBreak(domain.BreakTypeLine)
		case "ul":
			e.emitUl(block.Content)
		case "ol":
			e.emitOl(block.Content)
		default:
			e.emitInline(item, p)
		}
	}
	return nil
}

func (e *coreEmitter) emitTable(data []byte) error {
	var tblXML TableXML
	if err := xml.Unmarshal(data, &tblXML); err != nil {
		return err
	}

	if len(tblXML.Rows) == 0 {
		return nil
	}

	cols := 0
	for _, row := range tblXML.Rows {
		n := len(row.Cells) + len(row.ThCells)
		if n > cols {
			cols = n
		}
	}
	if cols == 0 {
		return nil
	}

	tbl, err := e.docx.AddTable(len(tblXML.Rows), cols)
	if err != nil {
		return err
	}

	if tblXML.Width > 0 {
		tbl.SetWidth(domain.TableWidth{
			Type:  domain.WidthPct,
			Value: int(tblXML.Width * 50),
		})
	}

	if tblXML.Align != "" {
		tbl.SetAlignment(mapTableAlignment(tblXML.Align))
	}

	if tblXML.At != "" {
		for i := range tblXML.Rows {
			row, err := tbl.Row(i)
			if err != nil {
				continue
			}
			allCells := append(tblXML.Rows[i].Cells, tblXML.Rows[i].ThCells...)
			for j := range allCells {
				if j >= cols {
					break
				}
				cell, err := row.Cell(j)
				if err != nil {
					continue
				}
				cell.SetBorders(parseTableBorders(tblXML.At))
			}
		}
	}

	if tblXML.CellSpace > 0 {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    "",
			modType: "cellSpacing:" + strconv.FormatFloat(tblXML.CellSpace, 'f', -1, 64),
		})
	}
	if tblXML.Indent > 0 {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    "",
			modType: "tableIndent:" + strconv.FormatFloat(tblXML.Indent, 'f', -1, 64),
		})
	}
	if tblXML.ID > 0 {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    "",
			modType: "tableBookmark:" + strconv.Itoa(tblXML.ID),
		})
	}
	if tblXML.Style != "" {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    tblXML.Style,
			modType: "tableStyle",
		})
	}
	if tblXML.Caption != "" {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    tblXML.Caption,
			modType: "tableCaption",
		})
	}
	if tblXML.Summary != "" {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    tblXML.Summary,
			modType: "tableSummary",
		})
	}

	for i, rowXML := range tblXML.Rows {
		row, err := tbl.Row(i)
		if err != nil {
			continue
		}

		allCells := append(rowXML.Cells, rowXML.ThCells...)
		for j, cellXML := range allCells {
			if j >= cols {
				break
			}
			cell, err := row.Cell(j)
			if err != nil {
				continue
			}

			if cellXML.ColSpan > 1 || cellXML.RowSpan > 1 {
				cell.Merge(cellXML.ColSpan, cellXML.RowSpan)
			}

			if cellXML.VAlign != "" {
				cell.SetVerticalAlignment(mapVerticalAlignment(cellXML.VAlign))
			}

			if cellXML.At != "" {
				cell.SetBorders(parseTableBorders(cellXML.At))
			}

			if cellXML.Lang != "" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    cellXML.Lang,
					modType: "cellLang:" + cellXML.Lang,
				})
			}
			if cellXML.NoWrap != "" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    cellXML.NoWrap,
					modType: "cellNoWrap:" + cellXML.NoWrap,
				})
			}
			if cellXML.TextDir != "" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    cellXML.TextDir,
					modType: "cellTextDir:" + cellXML.TextDir,
				})
			}

			for _, child := range cellXML.Content {
				data, _ := json.Marshal(child)
				var pBlock BlockXML
				if err := xml.Unmarshal(data, &pBlock); err != nil {
					continue
				}
				cellP, err := cell.AddParagraph()
				if err != nil {
					continue
				}
				if pBlock.XMLName.Local == "p" || strings.HasPrefix(pBlock.XMLName.Local, "h") {
					e.emitCellPara(data, cellP)
				} else {
					e.emitInline(child, cellP)
				}
			}
		}
	}

	return nil
}

func (e *coreEmitter) emitCellPara(data []byte, p domain.Paragraph) {
	var block BlockXML
	if err := xml.Unmarshal(data, &block); err != nil {
		return
	}
	if block.XMLName.Local != "p" && !strings.HasPrefix(block.XMLName.Local, "h") {
		return
	}

	if strings.HasPrefix(block.XMLName.Local, "h") {
		p.SetStyle(block.XMLName.Local)
	}

	for _, item := range block.Content {
		itemData, _ := json.Marshal(item)
		var inner BlockXML
		if err := xml.Unmarshal(itemData, &inner); err != nil {
			continue
		}
		switch inner.XMLName.Local {
		case "c":
			p.SetStyle(extractText(inner.Content))
		case "align":
			p.SetAlignment(mapAlignment(extractText(inner.Content)))
		case "indentLeft":
			if v, err := strconv.ParseFloat(extractText(inner.Content), 64); err == nil {
				indent := p.Indent()
				indent.Left = int(v)
				p.SetIndent(indent)
			}
		case "indentRight":
			if v, err := strconv.ParseFloat(extractText(inner.Content), 64); err == nil {
				indent := p.Indent()
				indent.Right = int(v)
				p.SetIndent(indent)
			}
		case "gapBefore":
			if v, err := strconv.ParseFloat(extractText(inner.Content), 64); err == nil {
				p.SetSpacingBefore(int(v))
			}
		case "gapAfter":
			if v, err := strconv.ParseFloat(extractText(inner.Content), 64); err == nil {
				p.SetSpacingAfter(int(v))
			}
		default:
			e.emitInline(item, p)
		}
	}
}

func (e *coreEmitter) emitHr() error {
	p, err := e.addParagraph()
	if err != nil {
		return err
	}
	p.SetStyle("NoSpacing")
	borders := domain.ParagraphBorders{
		Bottom: domain.BorderStyle{
			Style: domain.BorderSingle,
			Width: 4,
			Color: domain.Color{R: 0, G: 0, B: 0},
		},
	}
	p.SetBorders(borders)
	return nil
}

func (e *coreEmitter) emitImg(content []interface{}) error {
	p, err := e.addParagraph()
	if err != nil {
		return err
	}
	for _, item := range content {
		data, _ := json.Marshal(item)
		var inline InlineXML
		if err := xml.Unmarshal(data, &inline); err != nil {
			continue
		}
		alt := inline.Alt
		if alt == "" {
			alt = "image"
		}
		run, _ := p.AddRun()
		run.SetText("[Image: " + alt + "]")
		run.SetItalic(true)
		run.SetColor(domain.Color{R: 128, G: 128, B: 128})
	}
	return nil
}

func (e *coreEmitter) emitBr() error {
	if e.paragraph != nil {
		run, _ := e.paragraph.AddRun()
		run.AddBreak(domain.BreakTypeLine)
	}
	return nil
}

func (e *coreEmitter) emitGenericBlock(content []interface{}, tag string) error {
	p, err := e.addParagraph()
	if err != nil {
		return err
	}
	for _, item := range content {
		e.emitInline(item, p)
	}
	return nil
}

func (e *coreEmitter) emitInline(item interface{}, p domain.Paragraph) {
	data, err := json.Marshal(item)
	if err != nil {
		return
	}

	var inline InlineXML
	if err := json.Unmarshal(data, &inline); err != nil {
		return
	}

	tagName := inline.XMLName.Local

	switch tagName {
	case "br":
		run, _ := p.AddRun()
		breakType := domain.BreakTypeLine
		switch strings.ToLower(inline.Type) {
		case "page":
			breakType = domain.BreakTypePage
		case "column":
			breakType = domain.BreakTypeColumn
		}
		run.AddBreak(breakType)
	case "tab":
		run, _ := p.AddRun()
		run.SetText("")
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    "",
			modType: "tab",
		})
	case "t", "":
		run, _ := p.AddRun()
		text := inline.Text
		if text == "" {
			for _, child := range inline.Content {
				childData, _ := json.Marshal(child)
				var childInline InlineXML
				if err := json.Unmarshal(childData, &childInline); err == nil {
					if childInline.Text != "" {
						text = childInline.Text
						break
					}
				}
			}
		}
		if text != "" {
			run.SetText(text)
		}
	case "b":
		run, _ := p.AddRun()
		run.SetBold(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "i":
		run, _ := p.AddRun()
		run.SetItalic(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "u":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
		run.SetUnderline(domain.UnderlineSingle)
	case "s":
		run, _ := p.AddRun()
		run.SetStrike(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "em":
		run, _ := p.AddRun()
		run.SetItalic(true)
		run.SetBold(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "strong":
		run, _ := p.AddRun()
		run.SetBold(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "a":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
		if inline.Href != "" {
			if inline.Title != "" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    inline.Text,
					modType: "hyperlinkTitle:" + inline.Href + "|" + inline.Title,
				})
			}
			p.AddHyperlink(inline.Href, inline.Text)
		}
	case "img":
		run, _ := p.AddRun()
		run.SetText("[Image: " + inline.Alt + "]")
		run.SetItalic(true)
		run.SetColor(domain.Color{R: 128, G: 128, B: 128})
	case "sup":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{text: inline.Text, modType: "superscript"})
		}
	case "sub":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{text: inline.Text, modType: "subscript"})
		}
	case "code", "kbd", "samp":
		run, _ := p.AddRun()
		run.SetFont(domain.Font{Name: "Courier New"})
		run.SetColor(domain.Color{R: 199, G: 37, B: 78})
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "var":
		run, _ := p.AddRun()
		run.SetFont(domain.Font{Name: "Courier New"})
		run.SetItalic(true)
		run.SetColor(domain.Color{R: 199, G: 37, B: 78})
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "cite", "dfn":
		run, _ := p.AddRun()
		run.SetItalic(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "q":
		run, _ := p.AddRun()
		run.SetText("`" + inline.Text + "`")
	case "abbr":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "small":
		run, _ := p.AddRun()
		run.SetSize(16)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "smallcaps":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{text: inline.Text, modType: "smallcaps"})
		}
	case "uppercase":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{text: inline.Text, modType: "uppercase"})
		}
	case "bcs":
		run, _ := p.AddRun()
		run.SetBold(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{text: inline.Text, modType: "boldCs"})
		}
	case "ics":
		run, _ := p.AddRun()
		run.SetItalic(true)
		if inline.Text != "" {
			run.SetText(inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{text: inline.Text, modType: "italicCs"})
		}
	case "mark":
		run, _ := p.AddRun()
		run.SetHighlight(domain.HighlightYellow)
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
	case "ins":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
			author := getAttr(inline.Attr, "author")
			date := getAttr(inline.Attr, "date")
			e.insTexts = append(e.insTexts, inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{
				text:    inline.Text,
				modType: "insAuthor:" + author + "|date:" + date,
			})
		}
	case "del":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
			author := getAttr(inline.Attr, "author")
			date := getAttr(inline.Attr, "date")
			e.delTexts = append(e.delTexts, inline.Text)
			e.pendingMods = append(e.pendingMods, pendingRunMod{
				text:    inline.Text,
				modType: "delAuthor:" + author + "|date:" + date,
			})
		}
	case "span":
		e.emitSpanInline(inline, p)
	case "font":
		e.emitFontInline(inline, p)
	case "sym":
		run, _ := p.AddRun()
		symChar := getAttr(inline.Attr, "char")
		if symChar == "" {
			symChar = inline.Text
		}
		if symChar != "" {
			if r, err := strconv.ParseInt(symChar, 16, 32); err == nil {
				run.SetText(string(rune(r)))
			} else {
				run.SetText(symChar)
			}
		}
	case "fn":
		run, _ := p.AddRun()
		text := inline.Text
		if text == "" {
			text = inline.ID
		}
		run.SetText(text)
	case "fn-ref":
		run, _ := p.AddRun()
		text := inline.Text
		if text == "" {
			text = inline.ID
		}
		run.SetText(text)
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    text,
			modType: "fnref:" + inline.ID,
		})
	case "bm":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    inline.Text,
			modType: "bookmark:" + inline.ID,
		})
	case "comment":
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    inline.Text,
			modType: "comment:" + inline.ID,
		})
	default:
		run, _ := p.AddRun()
		if inline.Text != "" {
			run.SetText(inline.Text)
		}
		for _, child := range inline.Content {
			e.emitInline(child, p)
		}
	}
}

func (e *coreEmitter) emitSpanInline(inline InlineXML, p domain.Paragraph) {
	run, _ := p.AddRun()
	if inline.Text != "" {
		run.SetText(inline.Text)
	}

	font := domain.Font{}
	size := 0
	sizeCS := 0
	color := domain.Color{}
	highlight := ""

	for _, attr := range inline.Attr {
		switch attr.Name.Local {
		case "font":
			font.Name = attr.Value
		case "fontEA":
			font.EastAsia = attr.Value
		case "fontCS":
			font.CS = attr.Value
		case "size":
			if v, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				size = int(v)
			}
		case "sizeCS":
			if v, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				sizeCS = int(v)
			}
		case "color":
			c := strings.TrimPrefix(attr.Value, "#")
			if len(c) == 6 {
				r, _ := strconv.ParseUint(c[0:2], 16, 8)
				g, _ := strconv.ParseUint(c[2:4], 16, 8)
				b, _ := strconv.ParseUint(c[4:6], 16, 8)
				color = domain.Color{R: uint8(r), G: uint8(g), B: uint8(b)}
			}
		case "highlight":
			highlight = attr.Value
		case "hidden":
			if strings.ToLower(attr.Value) == "true" || attr.Value == "1" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    inline.Text,
					modType: "hidden",
				})
			}
		case "dir":
			if strings.ToLower(attr.Value) == "rtl" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    inline.Text,
					modType: "runBidi",
				})
			}
		case "lang":
			if attr.Value != "" {
				e.pendingMods = append(e.pendingMods, pendingRunMod{
					text:    inline.Text,
					modType: "runlang:" + attr.Value,
				})
			}
		}
	}

	if font.Name != "" || font.EastAsia != "" || font.CS != "" {
		run.SetFont(font)
	}
	if size > 0 {
		run.SetSize(size)
	}
	if sizeCS > 0 {
		e.pendingMods = append(e.pendingMods, pendingRunMod{
			text:    inline.Text,
			modType: "sizeCS:" + strconv.Itoa(sizeCS),
		})
	}
	if highlight != "" {
		run.SetHighlight(mapHighlightColor(highlight))
	}
	if color.R != 0 || color.G != 0 || color.B != 0 {
		run.SetColor(color)
	}

	for _, child := range inline.Content {
		childData, _ := json.Marshal(child)
		var childInline InlineXML
		if err := json.Unmarshal(childData, &childInline); err != nil {
			continue
		}
		if childInline.Text != "" {
			run.SetText(childInline.Text)
		}
	}
}

func (e *coreEmitter) emitFontInline(inline InlineXML, p domain.Paragraph) {
	run, _ := p.AddRun()
	if inline.Text != "" {
		run.SetText(inline.Text)
	}

	font := domain.Font{}
	size := 0
	color := domain.Color{}

	for _, attr := range inline.Attr {
		switch attr.Name.Local {
		case "face":
			font.Name = attr.Value
		case "size":
			if v, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				size = int(v)
			}
		case "color":
			c := strings.TrimPrefix(attr.Value, "#")
			if len(c) == 6 {
				r, _ := strconv.ParseUint(c[0:2], 16, 8)
				g, _ := strconv.ParseUint(c[2:4], 16, 8)
				b, _ := strconv.ParseUint(c[4:6], 16, 8)
				color = domain.Color{R: uint8(r), G: uint8(g), B: uint8(b)}
			}
		}
	}

	if font.Name != "" {
		run.SetFont(font)
	}
	if size > 0 {
		run.SetSize(size)
	}
	if color.R != 0 || color.G != 0 || color.B != 0 {
		run.SetColor(color)
	}

	for _, child := range inline.Content {
		childData, _ := json.Marshal(child)
		var childInline InlineXML
		if err := json.Unmarshal(childData, &childInline); err != nil {
			continue
		}
		if childInline.Text != "" {
			run.SetText(childInline.Text)
		}
	}
}

func extractText(content []interface{}) string {
	if len(content) == 0 {
		return ""
	}
	if s, ok := content[0].(string); ok {
		return s
	}
	data, _ := json.Marshal(content[0])
	var block BlockXML
	if err := xml.Unmarshal(data, &block); err == nil {
		if block.XMLName.Local == "" && len(block.Content) > 0 {
			return extractText(block.Content)
		}
	}
	return ""
}

func countTableDimensions(content []interface{}) (rows, cols int) {
	for _, item := range content {
		data, _ := json.Marshal(item)
		var block BlockXML
		if err := xml.Unmarshal(data, &block); err != nil {
			continue
		}
		if block.XMLName.Local == "tr" {
			rows++
			rowCols := 0
			for _, cell := range block.Content {
				cellData, _ := json.Marshal(cell)
				var cellBlock BlockXML
				if err := xml.Unmarshal(cellData, &cellBlock); err == nil {
					if cellBlock.XMLName.Local == "td" || cellBlock.XMLName.Local == "th" {
						rowCols++
					}
				}
			}
			if rowCols > cols {
				cols = rowCols
			}
		}
	}
	if rows == 0 {
		rows = 1
	}
	if cols == 0 {
		cols = 1
	}
	return
}
