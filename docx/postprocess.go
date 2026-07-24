package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type pendingRunMod struct {
	text    string
	modType string
	tabs    []StyleTab
}

func postProcessDocx(docxData []byte, mods []pendingRunMod) ([]byte, error) {
	if len(mods) == 0 {
		return docxData, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(docxData), int64(len(docxData)))
	if err != nil {
		return docxData, nil
	}

	files := make(map[string]*zipFileEntry)
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		files[f.Name] = &zipFileEntry{data: data, method: f.Method}
	}

	docEntry, ok := files["word/document.xml"]
	if !ok {
		return docxData, nil
	}

	xmlStr := string(docEntry.data)
	for _, mod := range mods {
		switch {
		case mod.modType == "smallcaps":
			xmlStr = injectRunProp(xmlStr, mod.text, "<w:smallCaps/>")
		case mod.modType == "uppercase":
			xmlStr = injectRunProp(xmlStr, mod.text, "<w:caps/>")
		case mod.modType == "superscript":
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:vertAlign w:val="superscript"/>`)
		case mod.modType == "subscript":
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:vertAlign w:val="subscript"/>`)
		case mod.modType == "tab":
			xmlStr = injectTab(xmlStr)
		case mod.modType == "boldCs":
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:bCs/>`)
		case mod.modType == "italicCs":
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:iCs/>`)
		case mod.modType == "hidden":
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:vanish/>`)
		case strings.HasPrefix(mod.modType, "sizeCS:"):
			val := strings.TrimPrefix(mod.modType, "sizeCS:")
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:szCs w:val="`+val+`"/>`)
		case mod.modType == "runBidi":
			xmlStr = injectRunProp(xmlStr, mod.text, `<w:bidi/>`)
		case strings.HasPrefix(mod.modType, "pvalign:"):
			valign := strings.TrimPrefix(mod.modType, "pvalign:")
			xmlStr = injectParaValign(xmlStr, valign)
		case strings.HasPrefix(mod.modType, "lang:"):
			lang := strings.TrimPrefix(mod.modType, "lang:")
			xmlStr = injectParaLang(xmlStr, lang)
		case strings.HasPrefix(mod.modType, "runlang:"):
			lang := strings.TrimPrefix(mod.modType, "runlang:")
			xmlStr = injectRunLang(xmlStr, mod.text, lang)
		case mod.modType == "bidi":
			xmlStr = injectParaBidi(xmlStr)
		case strings.HasPrefix(mod.modType, "bookmark:"):
			id := strings.TrimPrefix(mod.modType, "bookmark:")
			xmlStr = injectBookmark(xmlStr, mod.text, id)
		case strings.HasPrefix(mod.modType, "comment:"):
			id := strings.TrimPrefix(mod.modType, "comment:")
			xmlStr = injectComment(xmlStr, mod.text, id)
		case strings.HasPrefix(mod.modType, "fnref:"):
			id := strings.TrimPrefix(mod.modType, "fnref:")
			xmlStr = injectFnRef(xmlStr, mod.text, id)
		case strings.HasPrefix(mod.modType, "cellLang:"):
			lang := strings.TrimPrefix(mod.modType, "cellLang:")
			xmlStr = injectCellLang(xmlStr, lang)
		case strings.HasPrefix(mod.modType, "cellNoWrap:"):
			val := strings.TrimPrefix(mod.modType, "cellNoWrap:")
			xmlStr = injectCellNoWrap(xmlStr, val)
		case strings.HasPrefix(mod.modType, "cellTextDir:"):
			dir := strings.TrimPrefix(mod.modType, "cellTextDir:")
			xmlStr = injectCellTextDir(xmlStr, dir)
		case strings.HasPrefix(mod.modType, "cellSpacing:"):
			spacing := strings.TrimPrefix(mod.modType, "cellSpacing:")
			xmlStr = injectCellSpacing(xmlStr, spacing)
		case strings.HasPrefix(mod.modType, "tableIndent:"):
			indent := strings.TrimPrefix(mod.modType, "tableIndent:")
			xmlStr = injectTableIndent(xmlStr, indent)
		case mod.modType == "tabstops":
			xmlStr = injectTabStops(xmlStr, mod.tabs)
		case strings.HasPrefix(mod.modType, "tableBookmark:"):
			id := strings.TrimPrefix(mod.modType, "tableBookmark:")
			xmlStr = injectTableBookmark(xmlStr, id)
		case mod.modType == "tableStyle":
			xmlStr = injectTableStyle(xmlStr, mod.text)
		case mod.modType == "tableCaption":
			xmlStr = injectTableCaption(xmlStr, mod.text)
		case mod.modType == "tableSummary":
			xmlStr = injectTableSummary(xmlStr, mod.text)
		case strings.HasPrefix(mod.modType, "hyperlinkTitle:"):
			val := strings.TrimPrefix(mod.modType, "hyperlinkTitle:")
			parts := strings.SplitN(val, "|", 2)
			if len(parts) == 2 {
				xmlStr = injectHyperlinkTitle(xmlStr, parts[0], parts[1])
			}
		case strings.HasPrefix(mod.modType, "colSpace:"):
			val := strings.TrimPrefix(mod.modType, "colSpace:")
			xmlStr = injectColumnSpacing(xmlStr, val)
		case strings.HasPrefix(mod.modType, "colWidth:"):
			val := strings.TrimPrefix(mod.modType, "colWidth:")
			parts := strings.SplitN(val, ":", 2)
			if len(parts) == 2 {
				xmlStr = injectColumnWidth(xmlStr, parts[0], parts[1])
			}
		}
	}
	docEntry.data = []byte(xmlStr)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for path, entry := range files {
		f, err := writer.Create(path)
		if err != nil {
			continue
		}
		f.Write(entry.data)
	}
	writer.Close()

	return buf.Bytes(), nil
}

func injectRunProp(xmlStr, text, propXML string) string {
	if text == "" {
		return xmlStr
	}

	escaped := escapeXMLText(text)

	idx := strings.Index(xmlStr, escaped)
	if idx < 0 {
		return xmlStr
	}

	runStart := strings.LastIndex(xmlStr[:idx], "<w:r>")
	if runStart < 0 {
		return xmlStr
	}

	rPrStart := strings.Index(xmlStr[runStart:idx], "<w:rPr>")
	if rPrStart >= 0 {
		insertPos := runStart + rPrStart + len("<w:rPr>")
		return xmlStr[:insertPos] + propXML + xmlStr[insertPos:]
	}

	insertPos := runStart + len("<w:r>")
	return xmlStr[:insertPos] + "<w:rPr>" + propXML + "</w:rPr>" + xmlStr[insertPos:]
}

func injectParaLang(xmlStr, lang string) string {
	idx := strings.LastIndex(xmlStr, "<w:pPr>")
	if idx < 0 {
		idx = strings.Index(xmlStr, "<w:p>")
		if idx < 0 {
			return xmlStr
		}
		return xmlStr[:idx] + "<w:pPr><w:rPr><w:lang w:val=\"" + lang + "\"/></w:rPr></w:pPr>" + xmlStr[idx:]
	}
	insertPos := idx + len("<w:pPr>")
	return xmlStr[:insertPos] + "<w:rPr><w:lang w:val=\"" + lang + "\"/></w:rPr>" + xmlStr[insertPos:]
}

func injectParaBidi(xmlStr string) string {
	idx := strings.LastIndex(xmlStr, "<w:pPr>")
	if idx < 0 {
		idx = strings.Index(xmlStr, "<w:p>")
		if idx < 0 {
			return xmlStr
		}
		return xmlStr[:idx] + "<w:pPr><w:bidi/></w:pPr>" + xmlStr[idx:]
	}
	insertPos := idx + len("<w:pPr>")
	return xmlStr[:insertPos] + "<w:bidi/>" + xmlStr[insertPos:]
}

func injectRunLang(xmlStr, text, lang string) string {
	if text == "" {
		return xmlStr
	}

	escaped := escapeXMLText(text)
	idx := strings.Index(xmlStr, escaped)
	if idx < 0 {
		return xmlStr
	}

	runStart := strings.LastIndex(xmlStr[:idx], "<w:r>")
	if runStart < 0 {
		return xmlStr
	}

	rPrStart := strings.Index(xmlStr[runStart:idx], "<w:rPr>")
	if rPrStart >= 0 {
		insertPos := runStart + rPrStart + len("<w:rPr>")
		return xmlStr[:insertPos] + "<w:lang w:val=\"" + lang + "\"/>" + xmlStr[insertPos:]
	}

	insertPos := runStart + len("<w:r>")
	return xmlStr[:insertPos] + "<w:rPr><w:lang w:val=\"" + lang + "\"/></w:rPr>" + xmlStr[insertPos:]
}

func escapeXMLText(text string) string {
	var buf bytes.Buffer
	xml.Escape(&buf, []byte(text))
	return buf.String()
}

func postProcessInsDel(docxData []byte, insTexts, delTexts []string, mods []pendingRunMod) ([]byte, error) {
	if len(insTexts) == 0 && len(delTexts) == 0 {
		return docxData, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(docxData), int64(len(docxData)))
	if err != nil {
		return docxData, nil
	}

	files := make(map[string]*zipFileEntry)
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		files[f.Name] = &zipFileEntry{data: data, method: f.Method}
	}

	docEntry, ok := files["word/document.xml"]
	if !ok {
		return docxData, nil
	}

	xmlStr := string(docEntry.data)
	idCounter := 100

	insModIdx := 0
	delModIdx := 0
	for _, mod := range mods {
		if strings.HasPrefix(mod.modType, "insAuthor:") {
			if insModIdx < len(insTexts) {
				author, date := parseInsDelMod(mod.modType)
				if author == "" {
					author = "words-xml"
				}
				if date == "" {
					date = "2024-01-01T00:00:00Z"
				}
				xmlStr = wrapRunInTrackedChange(xmlStr, insTexts[insModIdx], "ins", author, date, &idCounter)
				insModIdx++
			}
		} else if strings.HasPrefix(mod.modType, "delAuthor:") {
			if delModIdx < len(delTexts) {
				author, date := parseInsDelMod(mod.modType)
				if author == "" {
					author = "words-xml"
				}
				if date == "" {
					date = "2024-01-01T00:00:00Z"
				}
				xmlStr = wrapRunInTrackedChange(xmlStr, delTexts[delModIdx], "del", author, date, &idCounter)
				delModIdx++
			}
		}
	}

	docEntry.data = []byte(xmlStr)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for path, entry := range files {
		f, err := writer.Create(path)
		if err != nil {
			continue
		}
		f.Write(entry.data)
	}
	writer.Close()

	return buf.Bytes(), nil
}

func parseInsDelMod(modType string) (author, date string) {
	parts := strings.SplitN(strings.TrimPrefix(modType, "insAuthor:"), "|date:", 2)
	if len(parts) >= 1 {
		author = parts[0]
	}
	if len(parts) >= 2 {
		date = parts[1]
	}
	return
}

func wrapRunInTrackedChange(xmlStr, text, changeType, author, date string, idCounter *int) string {
	if text == "" {
		return xmlStr
	}

	escaped := escapeXMLText(text)
	idx := strings.Index(xmlStr, escaped)
	if idx < 0 {
		return xmlStr
	}

	runStart := strings.LastIndex(xmlStr[:idx], "<w:r>")
	if runStart < 0 {
		return xmlStr
	}

	runEnd := strings.Index(xmlStr[idx:], "</w:r>")
	if runEnd < 0 {
		return xmlStr
	}
	runEnd = idx + runEnd + len("</w:r>")

	runXML := xmlStr[runStart:runEnd]
	*idCounter++

	tag := "w:ins"
	if changeType == "del" {
		tag = "w:del"
	}

	wrapper := fmt.Sprintf("<%s w:id=\"%d\" w:author=\"%s\" w:date=\"%s\">%s</%s>",
		tag, *idCounter, author, date, runXML, tag)

	return xmlStr[:runStart] + wrapper + xmlStr[runEnd:]
}

func injectBookmark(xmlStr, text, id string) string {
	if text == "" {
		return xmlStr
	}

	escaped := escapeXMLText(text)
	idx := strings.Index(xmlStr, escaped)
	if idx < 0 {
		return xmlStr
	}

	runStart := strings.LastIndex(xmlStr[:idx], "<w:r>")
	if runStart < 0 {
		return xmlStr
	}

	runEnd := strings.Index(xmlStr[idx:], "</w:r>")
	if runEnd < 0 {
		return xmlStr
	}
	runEnd = idx + runEnd + len("</w:r>")

	startTag := fmt.Sprintf("<w:bookmarkStart w:id=\"%s\" w:name=\"%s\"/>", id, id)
	endTag := fmt.Sprintf("<w:bookmarkEnd w:id=\"%s\"/>", id)

	return xmlStr[:runStart] + startTag + xmlStr[runStart:runEnd] + endTag + xmlStr[runEnd:]
}

func injectComment(xmlStr, text, id string) string {
	if text == "" {
		return xmlStr
	}

	escaped := escapeXMLText(text)
	idx := strings.Index(xmlStr, escaped)
	if idx < 0 {
		return xmlStr
	}

	runStart := strings.LastIndex(xmlStr[:idx], "<w:r>")
	if runStart < 0 {
		return xmlStr
	}

	runEnd := strings.Index(xmlStr[idx:], "</w:r>")
	if runEnd < 0 {
		return xmlStr
	}
	runEnd = idx + runEnd + len("</w:r>")

	rangeStart := fmt.Sprintf("<w:commentRangeStart w:id=\"%s\"/>", id)
	rangeEnd := fmt.Sprintf("<w:commentRangeEnd w:id=\"%s\"/>", id)
	refRun := fmt.Sprintf("<w:r><w:rPr><w:rStyle w:val=\"CommentReference\"/></w:rPr><w:commentReference w:id=\"%s\"/></w:r>", id)

	return xmlStr[:runStart] + rangeStart + xmlStr[runStart:runEnd] + rangeEnd + refRun + xmlStr[runEnd:]
}

func injectFnRef(xmlStr, text, id string) string {
	if text == "" {
		return xmlStr
	}

	escaped := escapeXMLText(text)
	idx := strings.Index(xmlStr, escaped)
	if idx < 0 {
		return xmlStr
	}

	runStart := strings.LastIndex(xmlStr[:idx], "<w:r>")
	if runStart < 0 {
		return xmlStr
	}

	runEnd := strings.Index(xmlStr[idx:], "</w:r>")
	if runEnd < 0 {
		return xmlStr
	}
	runEnd = idx + runEnd + len("</w:r>")

	fnRef := fmt.Sprintf("<w:r><w:rPr><w:rStyle w:val=\"FootnoteReference\"/></w:rPr><w:footnoteReference w:id=\"%s\"/></w:r>", id)

	return xmlStr[:runStart] + fnRef + xmlStr[runEnd:]
}

func injectTab(xmlStr string) string {
	idx := strings.LastIndex(xmlStr, "<w:r>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:r>")
	return xmlStr[:insertPos] + "<w:tab/>" + xmlStr[insertPos:]
}

func injectParaValign(xmlStr, valign string) string {
	val := "top"
	switch strings.ToLower(valign) {
	case "center", "middle":
		val = "center"
	case "bottom":
		val = "bottom"
	}
	idx := strings.LastIndex(xmlStr, "<w:pPr>")
	if idx < 0 {
		idx = strings.Index(xmlStr, "<w:p>")
		if idx < 0 {
			return xmlStr
		}
		return xmlStr[:idx] + "<w:pPr><w:vAlign w:val=\"" + val + "\"/></w:pPr>" + xmlStr[idx:]
	}
	insertPos := idx + len("<w:pPr>")
	return xmlStr[:insertPos] + "<w:vAlign w:val=\"" + val + "\"/>" + xmlStr[insertPos:]
}

func injectCellLang(xmlStr, lang string) string {
	idx := strings.LastIndex(xmlStr, "<w:tcPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tcPr>")
	return xmlStr[:insertPos] + "<w:tcBorders><w:lang w:val=\"" + lang + "\"/></w:tcBorders>" + xmlStr[insertPos:]
}

func injectCellNoWrap(xmlStr, val string) string {
	idx := strings.LastIndex(xmlStr, "<w:tcPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tcPr>")
	return xmlStr[:insertPos] + "<w:noWrap/>" + xmlStr[insertPos:]
}

func injectCellTextDir(xmlStr, dir string) string {
	val := "ltr"
	if strings.ToLower(dir) == "rtl" {
		val = "rtl"
	}
	idx := strings.LastIndex(xmlStr, "<w:tcPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tcPr>")
	return xmlStr[:insertPos] + "<w:textDirection w:val=\"" + val + "\"/>" + xmlStr[insertPos:]
}

func injectCellSpacing(xmlStr, spacing string) string {
	idx := strings.LastIndex(xmlStr, "<w:tblPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tblPr>")
	return xmlStr[:insertPos] + "<w:tblCellSpacing w:w=\"" + spacing + "\" w:type=\"dxa\"/>" + xmlStr[insertPos:]
}

func injectTableIndent(xmlStr, indent string) string {
	idx := strings.LastIndex(xmlStr, "<w:tblPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tblPr>")
	return xmlStr[:insertPos] + "<w:tblInd w:w=\"" + indent + "\" w:type=\"dxa\"/>" + xmlStr[insertPos:]
}

func injectTableBookmark(xmlStr, id string) string {
	idx := strings.LastIndex(xmlStr, "<w:tblPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tblPr>")
	return xmlStr[:insertPos] + "<w:tblStyle w:val=\"Table" + id + "\"/>" + xmlStr[insertPos:]
}

func injectTableStyle(xmlStr, style string) string {
	idx := strings.LastIndex(xmlStr, "<w:tblPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tblPr>")
	return xmlStr[:insertPos] + "<w:tblStyle w:val=\"" + style + "\"/>" + xmlStr[insertPos:]
}

func injectTableCaption(xmlStr, caption string) string {
	idx := strings.LastIndex(xmlStr, "<w:tblPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tblPr>")
	return xmlStr[:insertPos] + "<w tblCaption=\"" + caption + "\"/>" + xmlStr[insertPos:]
}

func injectTableSummary(xmlStr, summary string) string {
	idx := strings.LastIndex(xmlStr, "<w:tblPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:tblPr>")
	return xmlStr[:insertPos] + "<w tblSummary=\"" + summary + "\"/>" + xmlStr[insertPos:]
}

func injectTabStops(xmlStr string, tabs []StyleTab) string {
	if len(tabs) == 0 {
		return xmlStr
	}
	var tabXML string
	for _, tab := range tabs {
		pos := int(tab.Pos)
		if pos <= 0 {
			continue
		}
		val := mapTabAlign(tab.Align)
		leader := mapLeader(tab.Leader)
		tabXML += fmt.Sprintf("<w:tab w:pos=\"%d\" w:val=\"%s\"", pos, val)
		if leader != "" {
			tabXML += fmt.Sprintf(" w:leader=\"%s\"", leader)
		}
		tabXML += "/>"
	}
	if tabXML == "" {
		return xmlStr
	}
	tabsXML := "<w:tabs>" + tabXML + "</w:tabs>"
	idx := strings.LastIndex(xmlStr, "<w:pPr>")
	if idx < 0 {
		idx = strings.Index(xmlStr, "<w:p>")
		if idx < 0 {
			return xmlStr
		}
		return xmlStr[:idx] + "<w:pPr>" + tabsXML + "</w:pPr>" + xmlStr[idx:]
	}
	insertPos := idx + len("<w:pPr>")
	return xmlStr[:insertPos] + tabsXML + xmlStr[insertPos:]
}

func injectThemeFont(xmlStr string, font string) string {
	idx := strings.LastIndex(xmlStr, "<w:rPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:rPr>")
	fontsXML := fmt.Sprintf(`<w:rFonts w:ascii="%s" w:hAnsi="%s" w:cs="%s"/>`, font, font, font)
	return xmlStr[:insertPos] + fontsXML + xmlStr[insertPos:]
}

func injectHyperlinkTitle(xmlStr, href, title string) string {
	escaped := escapeXMLText(title)
	idx := strings.Index(xmlStr, "<w:hyperlink")
	if idx < 0 {
		return xmlStr
	}
	endIdx := strings.Index(xmlStr[idx:], ">")
	if endIdx < 0 {
		return xmlStr
	}
	insertPos := idx + endIdx
	return xmlStr[:insertPos] + " w:tooltip=\"" + escaped + "\"" + xmlStr[insertPos:]
}

func injectColumnSpacing(xmlStr, space string) string {
	idx := strings.LastIndex(xmlStr, "<w:sectPr>")
	if idx < 0 {
		return xmlStr
	}
	insertPos := idx + len("<w:sectPr>")
	return xmlStr[:insertPos] + "<w:cols w:space=\"" + space + "\"/>" + xmlStr[insertPos:]
}

func injectColumnWidth(xmlStr, ref, width string) string {
	idx := strings.LastIndex(xmlStr, "<w:cols")
	if idx < 0 {
		return xmlStr
	}
	endIdx := strings.Index(xmlStr[idx:], "/>")
	if endIdx < 0 {
		return xmlStr
	}
	insertPos := idx + endIdx
	colXML := fmt.Sprintf("<w:col w:w=\"%s\"/>", width)
	return xmlStr[:insertPos] + colXML + xmlStr[insertPos:]
}
