package words

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const twipsPerInch = 1440.0

var vmlImageRe = regexp.MustCompile(`r:id="([^"]+)"|r:embed="([^"]+)"|o:href="([^"]+)"`)

type pageSize struct {
	name string
	w    int
	h    int
}

var pagePresets = []pageSize{
	{"A3", 842, 1191},
	{"A4", 595, 842},
	{"A5", 420, 595},
	{"A6", 298, 420},
	{"B5", 516, 729},
	{"Letter", 612, 792},
	{"Legal", 612, 1008},
	{"Tabloid", 792, 1224},
	{"Executive", 540, 720},
	{"Statement", 396, 612},
	{"Folio", 612, 936},
}

func matchPageSize(wPt, hPt int) string {
	for _, p := range pagePresets {
		if (wPt == p.w && hPt == p.h) || (wPt == p.h && hPt == p.w) {
			return p.name
		}
	}
	return ""
}

var monospaceFonts = map[string]bool{
	"courier new": true, "courier": true, "consolas": true,
	"lucida console": true, "menlo": true, "monaco": true, "monospace": true,
}

func isMonospaceFont(f string) bool {
	f = strings.ToLower(strings.TrimSpace(f))
	if monospaceFonts[f] {
		return true
	}
	if strings.Contains(f, "mono") || strings.Contains(f, "courier") {
		return true
	}
	return false
}

func hasWordBoundary(s, word string) bool {
	idx := strings.Index(s, word)
	if idx < 0 {
		return false
	}
	if idx > 0 && s[idx-1] != ' ' && s[idx-1] != '-' && s[idx-1] != '_' {
		return false
	}
	end := idx + len(word)
	if end < len(s) && s[end] != ' ' && s[end] != '-' && s[end] != '_' {
		return false
	}
	return true
}

func ProcessDOCXFile(filePath string) (*ProcessedDocument, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return ProcessDOCXBytesMode(data, "semantic")
}

func ProcessDOCXFileMode(filePath string, mode string) (*ProcessedDocument, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return ProcessDOCXBytesMode(data, mode)
}

func ProcessDOCXBytes(data []byte) (*ProcessedDocument, error) {
	return ProcessDOCXBytesMode(data, "semantic")
}

func ProcessDOCXBytesMode(data []byte, mode string) (*ProcessedDocument, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx as zip: %w", err)
	}

	var docXML, stylesXML, numberingXML, relsXML, themeXML, coreXML []byte
	var footnotesXML, endnotesXML, commentsXML []byte
	headerFiles := make(map[string][]byte)
	footerFiles := make(map[string][]byte)

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		content, err := readZipFile(f)
		if err != nil {
			continue
		}
		name := filepath.ToSlash(f.Name)

		switch {
		case name == "word/document.xml":
			docXML = content
		case name == "word/styles.xml":
			stylesXML = content
		case name == "word/numbering.xml":
			numberingXML = content
		case name == "word/_rels/document.xml.rels":
			relsXML = content
		case name == "word/theme/theme1.xml":
			themeXML = content
		case name == "docProps/core.xml":
			coreXML = content
		case name == "word/footnotes.xml":
			footnotesXML = content
		case name == "word/endnotes.xml":
			endnotesXML = content
		case name == "word/comments.xml":
			commentsXML = content
		default:
			base := filepath.Base(name)
			dir := filepath.Dir(name)
			if strings.HasPrefix(base, "header") && strings.HasSuffix(base, ".xml") && dir == "word" {
				headerFiles[name] = content
			} else if strings.HasPrefix(base, "footer") && strings.HasSuffix(base, ".xml") && dir == "word" {
				footerFiles[name] = content
			}
		}
	}

	if docXML == nil {
		return nil, fmt.Errorf("document.xml not found in docx")
	}

	doc := &ParsedDocument{
		DefaultFont:  StyleDef{Family: "Times New Roman", SizePt: 11},
		StyleMap:     make(map[string]StyleDef),
		StyleNameMap: make(map[string]string),
		Mode:         mode,
	}

	if coreXML != nil {
		var cp CoreProps
		if err := xml.Unmarshal(coreXML, &cp); err == nil {
			doc.Meta = Meta{
				Title: cp.Title, Author: cp.Creator, Created: cp.Created,
				Modified: cp.Modified, Keywords: cp.Keywords, Subject: cp.Subject,
				Description: cp.Description,
			}
		}
	}

	if themeXML != nil {
		doc.Theme = extractTheme(themeXML)
	}

	themeFontMap := map[string]string{}
	if doc.Theme != nil && doc.Theme.FontMap != nil {
		themeFontMap = doc.Theme.FontMap
	}

	if stylesXML != nil {
		doc.StyleMap, doc.StyleNameMap, doc.DefaultFont = buildStyleMap(stylesXML, themeFontMap)
	}

	var numFmtMap map[string]string
	var numLvlTextMap map[string]string
	var numberingStartMap map[int]map[int]int
	var numToAbstract map[int]int
	var numLvlIndMap map[string]NumLvlInd
	if numberingXML != nil {
		numFmtMap, numLvlTextMap, numberingStartMap, numToAbstract, numLvlIndMap = buildNumberingMap(numberingXML)
	}

	relMap := make(map[string]string)
	if relsXML != nil {
		var rels RelsDoc
		if err := xml.Unmarshal(relsXML, &rels); err == nil {
			for _, item := range rels.Items {
				if strings.HasPrefix(item.Target, "http://") || strings.HasPrefix(item.Target, "https://") || strings.HasPrefix(item.Target, "mailto:") {
					relMap[item.ID] = item.Target
				} else {
					relMap[item.ID] = "word/" + item.Target
				}
			}
		}
	}

	var document DocDocument
	if err := xml.Unmarshal(docXML, &document); err != nil {
		return nil, fmt.Errorf("failed to parse document.xml: %w", err)
	}

	// Post-process runs to track child element order (tab/br before/after text)
	postProcessRunOrder(docXML, &document.Body)

	body := document.Body
	if len(body.Sections) > 0 {
		doc.PageSections = make([]PageLayout, 0, len(body.Sections))
		for _, sec := range body.Sections {
			pl := PageLayout{}
			if sec.PageSz != nil {
				pl.WidthInch = float64(sec.PageSz.W) / twipsPerInch
				pl.HeightInch = float64(sec.PageSz.H) / twipsPerInch
			}
			if sec.PageMar != nil {
				pl.MarginTop = float64(sec.PageMar.Top) / twipsPerInch
				pl.MarginRight = float64(sec.PageMar.Right) / twipsPerInch
				pl.MarginBottom = float64(sec.PageMar.Bottom) / twipsPerInch
				pl.MarginLeft = float64(sec.PageMar.Left) / twipsPerInch
				pl.HeaderMargin = float64(sec.PageMar.Header) / twipsPerInch
				pl.FooterMargin = float64(sec.PageMar.Footer) / twipsPerInch
			}
			if sec.Cols != nil && sec.Cols.Num > 1 {
				pl.Cols = sec.Cols.Num
				pl.ColsSpace = float64(sec.Cols.Space) / twipsPerInch
			}
			doc.PageSections = append(doc.PageSections, pl)
		}
		last := body.Sections[len(body.Sections)-1]
		if last.PageSz != nil {
			doc.PageLayout.WidthInch = float64(last.PageSz.W) / twipsPerInch
			doc.PageLayout.HeightInch = float64(last.PageSz.H) / twipsPerInch
		}
		if last.PageMar != nil {
			doc.PageLayout.MarginTop = float64(last.PageMar.Top) / twipsPerInch
			doc.PageLayout.MarginRight = float64(last.PageMar.Right) / twipsPerInch
			doc.PageLayout.MarginBottom = float64(last.PageMar.Bottom) / twipsPerInch
			doc.PageLayout.MarginLeft = float64(last.PageMar.Left) / twipsPerInch
			doc.PageLayout.HeaderMargin = float64(last.PageMar.Header) / twipsPerInch
			doc.PageLayout.FooterMargin = float64(last.PageMar.Footer) / twipsPerInch
		}
		if last.Cols != nil && last.Cols.Num > 1 {
			doc.Cols = last.Cols.Num
			doc.ColsSpace = float64(last.Cols.Space) / twipsPerInch
		}
	}

	headerContent := make(map[string][]ContentItem)
	for path, hdrData := range headerFiles {
		hdr, err := unmarshalHeader(hdrData)
		if err == nil {
			headerContent[path] = parseContentItemsOrdered(&hdr.DocOrderedContent, hdr.Children, relMap, doc.StyleMap, doc.StyleNameMap, numFmtMap, numberingStartMap, mode, themeFontMap)
		}
	}
	footerContent := make(map[string][]ContentItem)
	for path, ftrData := range footerFiles {
		ftr, err := unmarshalFooter(ftrData)
		if err == nil {
			footerContent[path] = parseContentItemsOrdered(&ftr.DocOrderedContent, ftr.Children, relMap, doc.StyleMap, doc.StyleNameMap, numFmtMap, numberingStartMap, mode, themeFontMap)
		}
	}

	doc.Headers = make([]HeaderFooter, 0, len(headerContent))
	for _, path := range sortedKeys(headerContent) {
		doc.Headers = append(doc.Headers, HeaderFooter{ID: len(doc.Headers) + 1, Content: headerContent[path]})
	}
	doc.Footers = make([]HeaderFooter, 0, len(footerContent))
	for _, path := range sortedKeys(footerContent) {
		doc.Footers = append(doc.Footers, HeaderFooter{ID: len(doc.Footers) + 1, Content: footerContent[path]})
	}
	doc.NumStartMap = numberingStartMap
	doc.NumToAbstract = numToAbstract
	doc.NumFmtMap = numFmtMap
	doc.NumLvlTextMap = numLvlTextMap
	doc.NumLvlIndMap = numLvlIndMap

	if footnotesXML != nil {
		var fndoc DocFootnotes
		if err := xml.Unmarshal(footnotesXML, &fndoc); err == nil {
			for _, fn := range fndoc.Footnotes {
				if fn.Type == "normal" || fn.Type == "" {
					items := parseContentItems(fn.Paras, nil, nil, relMap, doc.StyleMap, doc.StyleNameMap, numFmtMap, numberingStartMap, mode, themeFontMap)
					doc.Notes = append(doc.Notes, NoteItem{Type: "footnote", ID: fn.ID, Body: items})
				}
			}
		}
	}

	if endnotesXML != nil {
		var endoc DocEndnotes
		if err := xml.Unmarshal(endnotesXML, &endoc); err == nil {
			for _, en := range endoc.Endnotes {
				if en.Type == "normal" || en.Type == "" {
					items := parseContentItems(en.Paras, nil, nil, relMap, doc.StyleMap, doc.StyleNameMap, numFmtMap, numberingStartMap, mode, themeFontMap)
					doc.Notes = append(doc.Notes, NoteItem{Type: "endnote", ID: en.ID, Body: items})
				}
			}
		}
	}

	if commentsXML != nil {
		var cmdoc DocComments
		if err := xml.Unmarshal(commentsXML, &cmdoc); err == nil {
			for _, cm := range cmdoc.Comments {
				items := parseContentItems(cm.Paras, nil, nil, relMap, doc.StyleMap, doc.StyleNameMap, numFmtMap, numberingStartMap, mode, themeFontMap)
				doc.Notes = append(doc.Notes, NoteItem{Type: "comment", ID: cm.ID, Author: cm.Author, Date: cm.Date, Body: items})
			}
		}
	}

	doc.Content = parseContentItemsOrdered(&body, body.Children, relMap, doc.StyleMap, doc.StyleNameMap, numFmtMap, numberingStartMap, mode, themeFontMap)

	bmOrder, noteRefOrder := buildBodyNoteOrder(docXML)
	for i := range doc.Notes {
		n := &doc.Notes[i]
		key := fmt.Sprintf("%s_%d", n.Type, n.ID)
		if pos, ok := noteRefOrder[key]; ok {
			n.DocOrder = pos
		}
	}
	for _, bm := range body.Bookmarks {
		if bm.Name != "" {
			docOrder := -1
			if pos, ok := bmOrder[bm.ID]; ok {
				docOrder = pos
			}
			doc.Notes = append(doc.Notes, NoteItem{Type: "bm", ID: bm.ID, Name: bm.Name, Body: nil, DocOrder: docOrder})
		}
	}
	sort.SliceStable(doc.Notes, func(i, j int) bool {
		return doc.Notes[i].DocOrder < doc.Notes[j].DocOrder
	})

	tableID := 0
	var assignTableIDs func(items []ContentItem)
	assignTableIDs = func(items []ContentItem) {
		for i := range items {
			if items[i].Type == "table" && items[i].Table != nil {
				tableID++
				items[i].Table.ID = tableID
				doc.AllTables = append(doc.AllTables, items[i].Table)
				for _, row := range items[i].Table.Rows {
					for _, cell := range row.Cells {
						assignTableIDs(cell.Content)
					}
				}
			}
		}
	}
	assignTableIDs(doc.Content)
	// Also assign IDs to tables in headers and footers
	for i := range doc.Headers {
		assignTableIDs(doc.Headers[i].Content)
	}
	for i := range doc.Footers {
		assignTableIDs(doc.Footers[i].Content)
	}

	wordsXML := formatForLLM(doc)

	return &ProcessedDocument{
		WordsXML: wordsXML,
	}, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sortedKeys(m map[string][]ContentItem) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

var (
	wmlNS    = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	bookTag  = wmlNS + " bookmarkStart"
	pTag     = wmlNS + " p"
	rTag     = wmlNS + " r"
	fnRefTag = wmlNS + " footnoteReference"
	enRefTag = wmlNS + " endnoteReference"
	cmRefTag = wmlNS + " commentReference"
	insTag   = wmlNS + " ins"
	delTag   = wmlNS + " del"
)

// postProcessRunOrder scans the raw document XML with a token decoder to detect
// the relative order of w:tab/w:br/w:cr vs w:t within each w:r element.
// Go xml.Unmarshal loses child element order, so we need this secondary pass
// to correctly set TabBeforeText and BreakBeforeText on each DocRun.
//
// Only direct-body paragraph runs are indexed (body.Paras[*].Runs[*]).
// Runs inside w:tbl, w:sdt at body level are skipped during token scan
// to keep the allRuns index in sync with the collected run pointers.
func postProcessRunOrder(docXML []byte, body *DocBody) {
	const wml = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	dec := xml.NewDecoder(bytes.NewReader(docXML))

	// Collect pointers to all runs in direct-body paragraphs, in document order.
	type runPtr struct{ r *DocRun }
	var allRuns []runPtr
	for pi := range body.Paras {
		for ri := range body.Paras[pi].Runs {
			allRuns = append(allRuns, runPtr{r: &body.Paras[pi].Runs[ri]})
		}
	}
	if len(allRuns) == 0 {
		return
	}

	// Token scan: walk body tokens and for each w:r that belongs to a direct-body
	// paragraph (not inside w:tbl or w:sdt), detect child element order.
	inBody := false
	// skipDepth > 0 means we are inside a body-level w:tbl or w:sdt — skip all w:r.
	skipDepth := 0
	inPara := false // inside a direct-body w:p
	inRun := false  // inside a w:r within a direct-body paragraph
	runIdx := 0
	seenText := false
	seenTab := false
	seenBr := false
	tabBeforeText := false
	brBeforeText := false
	runDepth := 0 // depth inside w:r (to handle self-closing sub-elements)

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := t.Name.Local
			ns := t.Name.Space
			if ns != wml {
				if inRun {
					runDepth++
				}
				break
			}
			if local == "body" {
				inBody = true
				break
			}
			if !inBody {
				break
			}
			// Track body-level elements that contain their own runs (skip them)
			if !inPara && !inRun && (local == "tbl" || local == "sdt") {
				skipDepth++
				break
			}
			if skipDepth > 0 {
				if local == "tbl" || local == "sdt" {
					skipDepth++
				}
				break
			}
			if local == "p" && !inPara && !inRun {
				inPara = true
				break
			}
			// Skip runs inside w:hyperlink, w:dir, w:bdo — they are not in allRuns
			// (allRuns only has direct p.Runs, not p.Hyperlinks[].Runs etc.)
			if inPara && (local == "hyperlink" || local == "dir" || local == "bdo") {
				skipDepth++
				break
			}
			if local == "r" && inPara && !inRun && skipDepth == 0 {
				inRun = true
				runDepth = 1
				seenText = false
				seenTab = false
				seenBr = false
				tabBeforeText = false
				brBeforeText = false
				break
			}
			if inRun {
				runDepth++
				switch local {
				case "t", "delText":
					seenText = true
				case "tab":
					seenTab = true
					if !seenText {
						tabBeforeText = true
					}
				case "br", "cr":
					seenBr = true
					if !seenText {
						brBeforeText = true
					}
				}
			}
		case xml.EndElement:
			if !inBody {
				break
			}
			local := t.Name.Local
			ns := t.Name.Space
			if ns != wml {
				if inRun {
					runDepth--
				}
				break
			}
			if skipDepth > 0 {
				if local == "tbl" || local == "sdt" {
					skipDepth--
				}
				break
			}
			if inRun {
				runDepth--
				if runDepth == 0 {
					// End of w:r — apply order flags to matching DocRun
					if runIdx < len(allRuns) {
						if seenTab && seenText {
							allRuns[runIdx].r.TabBeforeText = tabBeforeText
						}
						if seenBr && seenText {
							allRuns[runIdx].r.BreakBeforeText = brBeforeText
						}
					}
					runIdx++
					inRun = false
				}
				break
			}
			if inPara && (local == "hyperlink" || local == "dir" || local == "bdo") && skipDepth > 0 {
				skipDepth--
				break
			}
			if local == "p" && inPara {
				inPara = false
				break
			}
			if local == "body" {
				inBody = false
			}
		}
	}
}

func buildBodyNoteOrder(docXML []byte) (bmOrder map[int]int, noteRefOrder map[string]int) {
	bmOrder = make(map[int]int)
	noteRefOrder = make(map[string]int)
	decoder := xml.NewDecoder(bytes.NewReader(docXML))
	inBody := false
	bodyPos := 0
	inIns := false
	inDel := false
	inP := false
	pPos := -1
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			fullName := t.Name.Space + " " + t.Name.Local
			switch {
			case !inBody && fullName == wmlNS+" body":
				inBody = true
			case inBody && fullName == bookTag:
				id := -1
				for _, a := range t.Attr {
					if a.Name.Local == "id" {
						if v, err := strconv.Atoi(a.Value); err == nil {
							id = v
						}
					}
				}
				if id >= 0 {
					bmOrder[id] = bodyPos
				}
				bodyPos++
			case inBody && fullName == pTag:
				inP = true
				pPos = bodyPos
				bodyPos++
			case inP && fullName == rTag:
			case inP && fullName == insTag:
				inIns = true
			case inP && fullName == delTag:
				inDel = true
			case inP && (fullName == fnRefTag || fullName == enRefTag || fullName == cmRefTag) && !inIns && !inDel:
				id := -1
				ntype := ""
				switch fullName {
				case fnRefTag:
					ntype = "footnote"
				case enRefTag:
					ntype = "endnote"
				case cmRefTag:
					ntype = "comment"
				}
				for _, a := range t.Attr {
					if a.Name.Local == "id" {
						if v, err := strconv.Atoi(a.Value); err == nil {
							id = v
						}
					}
				}
				if id >= 0 && ntype != "" {
					key := fmt.Sprintf("%s_%d", ntype, id)
					if _, exists := noteRefOrder[key]; !exists {
						noteRefOrder[key] = pPos
					}
				}
			case inBody && (fullName == wmlNS+" tbl" || fullName == wmlNS+" sdt"):
				bodyPos++
			case inBody && fullName == wmlNS+" sectPr":
				bodyPos++
			}
		case xml.EndElement:
			fullName := t.Name.Space + " " + t.Name.Local
			switch {
			case inP && fullName == pTag:
				inP = false
				pPos = -1
				inIns = false
				inDel = false
			case inP && fullName == insTag:
				inIns = false
			case inP && fullName == delTag:
				inDel = false
			}
		}
	}
	return bmOrder, noteRefOrder
}

func unmarshalHeader(data []byte) (*DocHeader, error) {
	var h DocHeader
	if err := xml.Unmarshal(data, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func unmarshalFooter(data []byte) (*DocFooter, error) {
	var f DocFooter
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

type DocHeader struct {
	DocOrderedContent
}

func (h *DocHeader) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshalOrderedContent(&h.DocOrderedContent, d, start)
}

type DocFooter struct {
	DocOrderedContent
}

func (f *DocFooter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshalOrderedContent(&f.DocOrderedContent, d, start)
}

// --- Style Resolver ---

func buildStyleMap(stylesXML []byte, themeFontMap map[string]string) (map[string]StyleDef, map[string]string, StyleDef) {
	var styles DocStyles
	if err := xml.Unmarshal(stylesXML, &styles); err != nil {
		return nil, nil, StyleDef{Family: "Times New Roman", SizePt: 11}
	}

	m := make(map[string]StyleDef)
	nameMap := make(map[string]string)
	defaultFont := StyleDef{Family: "Times New Roman", SizePt: 11}

	if styles.DocDefaults != nil && styles.DocDefaults.RunPropsDefault != nil {
		rp := styles.DocDefaults.RunPropsDefault.RunProps
		if rp != nil {
			if rp.RFonts != nil {
				if rp.RFonts.Ascii != "" {
					defaultFont.Family = rp.RFonts.Ascii
				} else if rp.RFonts.HAnsi != "" {
					defaultFont.Family = rp.RFonts.HAnsi
				} else if rp.RFonts.AsciiTheme != "" {
					if f, ok := themeFontMap[rp.RFonts.AsciiTheme]; ok {
						defaultFont.Family = f
					}
				} else if rp.RFonts.HAnsiTheme != "" {
					if f, ok := themeFontMap[rp.RFonts.HAnsiTheme]; ok {
						defaultFont.Family = f
					}
				}
				if rp.RFonts.EastAsia != "" {
					defaultFont.FontEA = rp.RFonts.EastAsia
				} else if rp.RFonts.EastAsiaTheme != "" {
					if f, ok := themeFontMap[rp.RFonts.EastAsiaTheme]; ok {
						defaultFont.FontEA = f
					}
				}
				if rp.RFonts.CS != "" {
					defaultFont.FontCS = rp.RFonts.CS
				} else if rp.RFonts.CSTheme != "" {
					if f, ok := themeFontMap[rp.RFonts.CSTheme]; ok {
						defaultFont.FontCS = f
					}
				}
			}
			if rp.Sz != nil {
				defaultFont.SizePt = float64(rp.Sz.Val) / 2.0
			}
			if rp.Color != nil && rp.Color.Val != "" {
				defaultFont.Color = rp.Color.Val
			}
		}
	}

	for _, s := range styles.Styles {
		sd := StyleDef{}
		sd.Type = s.Type
		if s.Name != nil {
			nameMap[s.ID] = s.Name.Val
			sd.Name = s.Name.Val
		}
		if s.BasedOn != nil {
			sd.BasedOn = s.BasedOn.Val
		}
		if s.RunProps != nil {
			if s.RunProps.RFonts != nil {
				if s.RunProps.RFonts.Ascii != "" {
					sd.Family = s.RunProps.RFonts.Ascii
				} else if s.RunProps.RFonts.HAnsi != "" {
					sd.Family = s.RunProps.RFonts.HAnsi
				} else if s.RunProps.RFonts.AsciiTheme != "" {
					if f, ok := themeFontMap[s.RunProps.RFonts.AsciiTheme]; ok {
						sd.Family = f
					}
				} else if s.RunProps.RFonts.HAnsiTheme != "" {
					if f, ok := themeFontMap[s.RunProps.RFonts.HAnsiTheme]; ok {
						sd.Family = f
					}
				}
				if s.RunProps.RFonts.EastAsia != "" {
					sd.FontEA = s.RunProps.RFonts.EastAsia
				} else if s.RunProps.RFonts.EastAsiaTheme != "" {
					if f, ok := themeFontMap[s.RunProps.RFonts.EastAsiaTheme]; ok {
						sd.FontEA = f
					}
				}
				if s.RunProps.RFonts.CS != "" {
					sd.FontCS = s.RunProps.RFonts.CS
				} else if s.RunProps.RFonts.CSTheme != "" {
					if f, ok := themeFontMap[s.RunProps.RFonts.CSTheme]; ok {
						sd.FontCS = f
					}
				}
			}
			if s.RunProps.Sz != nil {
				sd.SizePt = float64(s.RunProps.Sz.Val) / 2.0
			}
			if s.RunProps.SzCs != nil {
				sd.SizeCS = float64(s.RunProps.SzCs.Val) / 2.0
			}
			if s.RunProps.Color != nil && s.RunProps.Color.Val != "" {
				sd.Color = s.RunProps.Color.Val
			}
			if s.RunProps.B.IsOn() {
				sd.Bold = true
			}
			if s.RunProps.I.IsOn() {
				sd.Italic = true
			}
			if s.RunProps.U != nil {
				sd.Underline = s.RunProps.U.Val
			}
			if s.RunProps.Strike.IsOn() {
				sd.Strikethrough = true
			}
			if s.RunProps.SmallCaps.IsOn() {
				sd.SmallCaps = true
			}
			if s.RunProps.Caps.IsOn() {
				sd.Uppercase = true
			}
		}
		if s.ParaProps != nil {
			if s.ParaProps.OutlineLvl != nil {
				sd.HeadingLevel = s.ParaProps.OutlineLvl.Val + 1
			}
			if s.ParaProps.JC != nil {
				sd.Align = s.ParaProps.JC.Val
			}
			if s.ParaProps.Spacing != nil {
				sd.SpacingBefore = float64(s.ParaProps.Spacing.Before) / twipsPerInch
				sd.SpacingAfter = float64(s.ParaProps.Spacing.After) / twipsPerInch
				if s.ParaProps.Spacing.Line > 0 {
					switch s.ParaProps.Spacing.LineRule {
					case "auto":
						sd.LineSpacing = float64(s.ParaProps.Spacing.Line) / 240.0
					default:
						// exact/atLeast: twips → inches
						sd.LineSpacing = float64(s.ParaProps.Spacing.Line) / twipsPerInch
					}
					sd.LineRule = s.ParaProps.Spacing.LineRule
				}
			}
			if s.ParaProps.Ind != nil {
				sd.IndentLeft = float64(s.ParaProps.Ind.Left) / twipsPerInch
				sd.IndentRight = float64(s.ParaProps.Ind.Right) / twipsPerInch
				sd.IndentFirst = float64(s.ParaProps.Ind.FirstLine) / twipsPerInch
				sd.IndentHanging = float64(s.ParaProps.Ind.Hanging) / twipsPerInch
			}
			if s.ParaProps.Tabs != nil {
				for _, t := range s.ParaProps.Tabs.Tabs {
					pt := ParsedTab{Pos: float64(t.Pos) / twipsPerInch}
					if t.Val != "" {
						pt.Align = t.Val
					} else {
						pt.Align = "left"
					}
					pt.Leader = t.Leader
					if pt.Leader == "" {
						pt.Leader = "none"
					}
					sd.Tabs = append(sd.Tabs, pt)
				}
			}
		}
		if s.TblPr != nil {
			if s.TblPr.Borders != nil {
				if s.TblPr.Borders.Top != nil {
					sd.BorderWidth = float64(s.TblPr.Borders.Top.Sz) / 576.0
					sd.BorderColor = s.TblPr.Borders.Top.Color
					sd.BorderStyle = s.TblPr.Borders.Top.Val
				}
			}
			if s.TblPr.Spacing != nil {
				sd.CellSpacing = float64(s.TblPr.Spacing.W) / twipsPerInch
			}
			if s.TblPr.TblW != nil {
				sd.Width = float64(s.TblPr.TblW.W) / twipsPerInch
			}
		}
		if sd.HeadingLevel == 0 {
			sd.HeadingLevel = inferHeadingLevel(s.ID)
		}
		m[s.ID] = sd
	}

	return m, nameMap, defaultFont
}

func inferHeadingLevel(styleID string) int {
	id := strings.ToLower(styleID)
	if strings.HasPrefix(id, "heading") {
		numStr := strings.TrimPrefix(id, "heading")
		if num, err := strconv.Atoi(numStr); err == nil && num >= 1 && num <= 9 {
			return num
		}
	}
	if id == "title" || id == "heading" {
		return 1
	}
	if id == "subtitle" {
		return 2
	}
	return 0
}

func resolveHeadingLevel(styleID, styleName string, styleMap map[string]StyleDef) int {
	level := inferHeadingLevel(styleID)
	if level > 0 {
		return level
	}
	visited := make(map[string]bool)
	current := styleID
	for current != "" && !visited[current] {
		visited[current] = true
		sd, ok := styleMap[current]
		if !ok {
			break
		}
		if sd.HeadingLevel > 0 {
			return sd.HeadingLevel
		}
		current = sd.BasedOn
	}
	return 0
}

// --- Numbering Resolver ---

func buildNumberingMap(numberingXML []byte) (map[string]string, map[string]string, map[int]map[int]int, map[int]int, map[string]NumLvlInd) {
	var numbering DocNumbering
	if err := xml.Unmarshal(numberingXML, &numbering); err != nil {
		return nil, nil, nil, nil, nil
	}

	abstractFmt := make(map[int]map[int]string)
	abstractLvlText := make(map[int]map[int]string)
	abstractLvlInd := make(map[int]map[int]NumLvlInd)
	abstractLvlStart := make(map[int]map[int]int)
	for _, an := range numbering.AbstractNums {
		levels := make(map[int]string)
		lvlTexts := make(map[int]string)
		lvlInds := make(map[int]NumLvlInd)
		lvlStarts := make(map[int]int)
		for _, lvl := range an.Levels {
			if lvl.NumFmt != nil {
				levels[lvl.Ilvl] = lvl.NumFmt.Val
			}
			if lvl.LvlText != nil {
				lvlTexts[lvl.Ilvl] = lvl.LvlText.Val
			}
			if lvl.Start != nil {
				lvlStarts[lvl.Ilvl] = lvl.Start.Val
			}
			if lvl.PPr != nil && lvl.PPr.Ind != nil {
				lvlInds[lvl.Ilvl] = NumLvlInd{
					Left:      float64(lvl.PPr.Ind.Left) / twipsPerInch,
					Hanging:   float64(lvl.PPr.Ind.Hanging) / twipsPerInch,
					FirstLine: float64(lvl.PPr.Ind.FirstLine) / twipsPerInch,
				}
			}
		}
		abstractFmt[an.ID] = levels
		abstractLvlText[an.ID] = lvlTexts
		abstractLvlInd[an.ID] = lvlInds
		abstractLvlStart[an.ID] = lvlStarts
	}

	numFmtMap := make(map[string]string)
	numLvlTextMap := make(map[string]string)
	numLvlIndMap := make(map[string]NumLvlInd)
	numStartMap := make(map[int]map[int]int)
	numToAbstract := make(map[int]int)
	for _, num := range numbering.Nums {
		aid := -1
		if num.AbstractNumID != nil {
			aid = num.AbstractNumID.Val
		}
		numToAbstract[num.NumID] = aid
		if af, ok := abstractFmt[aid]; ok {
			for ilvl, nf := range af {
				key := fmt.Sprintf("%d_%d", num.NumID, ilvl)
				numFmtMap[key] = nf
			}
		}
		if lt, ok := abstractLvlText[aid]; ok {
			for ilvl, txt := range lt {
				key := fmt.Sprintf("%d_%d", num.NumID, ilvl)
				numLvlTextMap[key] = txt
			}
		}
		if li, ok := abstractLvlInd[aid]; ok {
			for ilvl, ind := range li {
				key := fmt.Sprintf("%d_%d", num.NumID, ilvl)
				numLvlIndMap[key] = ind
			}
		}
		// Seed the start value from the abstract level's base w:start so the
		// emitted list reflects Word's actual numbering; a num-level
		// startOverride wins over the base value.
		if ls, ok := abstractLvlStart[aid]; ok {
			for ilvl, st := range ls {
				if _, ok := numStartMap[num.NumID]; !ok {
					numStartMap[num.NumID] = make(map[int]int)
				}
				numStartMap[num.NumID][ilvl] = st
			}
		}
		for _, ov := range num.LvlOverrides {
			if ov.StartOverride != nil {
				if _, ok := numStartMap[num.NumID]; !ok {
					numStartMap[num.NumID] = make(map[int]int)
				}
				numStartMap[num.NumID][ov.Ilvl] = ov.StartOverride.Val
			}
		}
	}

	return numFmtMap, numLvlTextMap, numStartMap, numToAbstract, numLvlIndMap
}

// --- Parser ---

// parseContentItemsOrdered processes body children in document order,
// preserving the interleaving of paragraphs, tables, and SDTs.
func parseContentItemsOrdered(content OrderedContentAccessor, children []BodyChild, relMap map[string]string, styleMap map[string]StyleDef, styleNameMap map[string]string, numFmtMap map[string]string, numStartMap map[int]map[int]int, mode string, themeFontMap map[string]string) []ContentItem {
	var items []ContentItem
	for _, child := range children {
		switch child.Type {
		case BodyChildPara:
			p := *content.ParaAt(child.Idx)
			if p.PPr != nil && p.PPr.NumPr != nil {
				numID := p.PPr.NumPr.NumID.Val
				if numID != 0 {
					ilvl := 0
					if p.PPr.NumPr.Ilvl != nil {
						ilvl = p.PPr.NumPr.Ilvl.Val
					}
					lp := &ParsedParagraph{
						IsList:     true,
						ListLevel:  ilvl,
						NumID:      numID,
						ListFormat: "ul",
					}
					if p.PPr.PStyle != nil {
						lp.StyleID = p.PPr.PStyle.Val
						if name, ok := styleNameMap[lp.StyleID]; ok {
							lp.StyleName = name
						}
					}
					nKey := fmt.Sprintf("%d_%d", numID, ilvl)
					if nf, ok := numFmtMap[nKey]; ok && nf != "bullet" {
						lp.ListFormat = "ol"
					}
					fillParaLayout(lp, p.PPr, styleMap)
					var tbItems []ContentItem
					lp.Runs, tbItems = extractRuns(p, styleMap, styleNameMap, relMap, numFmtMap, numStartMap, mode, false, themeFontMap, styleFontDefaults(p, styleMap, themeFontMap), nil)
					items = append(items, tbItems...)
					items = append(items, ContentItem{Type: "list", Paragraph: lp})
					// Handle textboxes in list paragraphs
					for _, tb := range p.Textboxes {
						if tb.TxbxContent == nil {
							continue
						}
						tbI := parseContentItemsOrdered(&tb.TxbxContent.DocOrderedContent, tb.TxbxContent.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
						items = append(items, tbI...)
					}
					continue
				}
			}
			item, paraTbItems := parseParagraph(p, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
			items = append(items, item)
			items = append(items, paraTbItems...)
			for _, tb := range p.Textboxes {
				if tb.TxbxContent == nil {
					continue
				}
				tbI := parseContentItemsOrdered(&tb.TxbxContent.DocOrderedContent, tb.TxbxContent.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
				items = append(items, tbI...)
			}
		case BodyChildTable:
			t := parseTable(*content.TableAt(child.Idx), relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
			items = append(items, ContentItem{Type: "table", Table: t})
		case BodyChildSdt:
			sdt := content.SdtAt(child.Idx)
			items = append(items, parseContentItemsOrdered(&sdt.Content.DocOrderedContent, sdt.Content.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)...)
		}
	}
	return items
}

func parseContentItems(paras []DocPara, tables []DocTbl, sdts []DocSdt, relMap map[string]string, styleMap map[string]StyleDef, styleNameMap map[string]string, numFmtMap map[string]string, numStartMap map[int]map[int]int, mode string, themeFontMap map[string]string) []ContentItem {
	var items []ContentItem

	i := 0
	for i < len(paras) {
		p := paras[i]

		if p.PPr != nil && p.PPr.NumPr != nil {
			numID := p.PPr.NumPr.NumID.Val
			// w:numId=0 means "remove list formatting" — treat as regular paragraph
			if numID == 0 {
				goto parsePara
			}
			ilvl := 0
			if p.PPr.NumPr.Ilvl != nil {
				ilvl = p.PPr.NumPr.Ilvl.Val
			}

			lp := &ParsedParagraph{
				IsList:     true,
				ListLevel:  ilvl,
				NumID:      numID,
				ListFormat: "ul",
			}

			if p.PPr.PStyle != nil {
				lp.StyleID = p.PPr.PStyle.Val
				if name, ok := styleNameMap[lp.StyleID]; ok {
					lp.StyleName = name
				}
			}

			nKey := fmt.Sprintf("%d_%d", numID, ilvl)
			if nf, ok := numFmtMap[nKey]; ok {
				if nf != "bullet" {
					lp.ListFormat = "ol"
				}
			}
			fillParaLayout(lp, p.PPr, styleMap)

			var tbItems []ContentItem
			lp.Runs, tbItems = extractRuns(p, styleMap, styleNameMap, relMap, numFmtMap, numStartMap, mode, false, themeFontMap, styleFontDefaults(p, styleMap, themeFontMap), nil)
			items = append(items, tbItems...)
			items = append(items, ContentItem{Type: "list", Paragraph: lp})
			i++
			continue
		}

	parsePara:
		item, paraTbItems := parseParagraph(p, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
		items = append(items, item)
		items = append(items, paraTbItems...)

		for _, tb := range p.Textboxes {
			if tb.TxbxContent == nil {
				continue
			}
			tbItems := parseContentItemsOrdered(&tb.TxbxContent.DocOrderedContent, tb.TxbxContent.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
			items = append(items, tbItems...)
		}
		i++
	}

	for ti := range tables {
		t := parseTable(tables[ti], relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
		items = append(items, ContentItem{Type: "table", Table: t})
	}

	for _, sdt := range sdts {
		items = append(items, parseContentItemsOrdered(&sdt.Content.DocOrderedContent, sdt.Content.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)...)
	}

	return items
}

// fillParaLayout copies the paragraph-level layout properties (indent, spacing,
// tabs) from pPr onto lp. It is shared by the plain-paragraph path and both
// list-item paths so they can never diverge on which w:pPr properties survive.
//
// Indent and spacing follow Word's block semantics: a paragraph-direct w:ind or
// w:spacing block replaces the style's block entirely, and when the paragraph
// has none, the nearest style in the basedOn chain that defines one is inherited
// (the nearest definition wins). List items are excluded from indent/spacing
// inheritance — their marker geometry comes from the numbering level and their
// spacing from the direct paragraph properties.
//
// Tabs are resolved to the paragraph's effective stop set: custom stops inherit
// from the paragraph style chain and merge with paragraph-direct stops, where a
// direct stop overrides an inherited one at the same position and a direct
// w:val="clear" removes the inherited stop at that position. This mirrors how
// Word keeps style stops until a paragraph explicitly clears them.
func fillParaLayout(lp *ParsedParagraph, pPr *ParaProps, styleMap map[string]StyleDef) {
	if pPr == nil {
		return
	}
	if pPr.Spacing != nil {
		lp.SpacingBefore = float64(pPr.Spacing.Before) / twipsPerInch
		lp.SpacingAfter = float64(pPr.Spacing.After) / twipsPerInch
		if pPr.Spacing.Line > 0 {
			switch pPr.Spacing.LineRule {
			case "auto":
				lp.LineSpacing = float64(pPr.Spacing.Line) / 240.0
			default:
				// exact/atLeast: twips → inches
				lp.LineSpacing = float64(pPr.Spacing.Line) / twipsPerInch
			}
			lp.LineRule = pPr.Spacing.LineRule
		}
	} else if !lp.IsList {
		if st, ok := inheritedStyleParaProps(lp.StyleID, styleMap, styleDefinesSpacing); ok {
			lp.SpacingBefore = st.SpacingBefore
			lp.SpacingAfter = st.SpacingAfter
			lp.LineSpacing = st.LineSpacing
			lp.LineRule = st.LineRule
		}
	}
	if pPr.Ind != nil {
		lp.IndentLeft = float64(pPr.Ind.Left) / twipsPerInch
		lp.IndentRight = float64(pPr.Ind.Right) / twipsPerInch
		lp.IndentFirst = float64(pPr.Ind.FirstLine) / twipsPerInch
		lp.IndentHanging = float64(pPr.Ind.Hanging) / twipsPerInch
	} else if !lp.IsList {
		if st, ok := inheritedStyleParaProps(lp.StyleID, styleMap, styleDefinesInd); ok {
			lp.IndentLeft = st.IndentLeft
			lp.IndentRight = st.IndentRight
			lp.IndentFirst = st.IndentFirst
			lp.IndentHanging = st.IndentHanging
		}
	}
	inherited := inheritedStyleTabs(lp.StyleID, styleMap)
	if pPr.Tabs != nil {
		lp.Tabs = resolveEffectiveTabs(parseTabs(pPr.Tabs), inherited)
	} else {
		lp.Tabs = inherited
	}
}

// styleDefinesInd reports whether the style carries its own w:ind block.
func styleDefinesInd(sd StyleDef) bool {
	return sd.IndentLeft > 0 || sd.IndentRight > 0 || sd.IndentFirst > 0 || sd.IndentHanging > 0
}

// styleDefinesSpacing reports whether the style carries its own w:spacing block.
func styleDefinesSpacing(sd StyleDef) bool {
	return sd.SpacingBefore > 0 || sd.SpacingAfter > 0 || sd.LineSpacing > 0
}

// inheritedStyleParaProps walks the paragraph style's basedOn chain and returns
// the nearest style for which want() is true (the nearest definition wins).
func inheritedStyleParaProps(styleID string, styleMap map[string]StyleDef, want func(StyleDef) bool) (StyleDef, bool) {
	visited := make(map[string]bool)
	current := styleID
	for current != "" && !visited[current] {
		visited[current] = true
		sd, ok := styleMap[current]
		if !ok {
			break
		}
		if want(sd) {
			return sd, true
		}
		current = sd.BasedOn
	}
	return StyleDef{}, false
}

// parseTabs converts the raw OOXML w:tabs entries into ParsedTab values.
func parseTabs(tabs *TabsVal) []ParsedTab {
	var out []ParsedTab
	if tabs == nil {
		return nil
	}
	for _, t := range tabs.Tabs {
		pt := ParsedTab{Pos: float64(t.Pos) / twipsPerInch}
		if t.Val != "" {
			pt.Align = t.Val
		} else {
			pt.Align = "left"
		}
		pt.Leader = t.Leader
		if pt.Leader == "" {
			pt.Leader = "none"
		}
		out = append(out, pt)
	}
	return out
}

// inheritedStyleTabs walks the paragraph style's basedOn chain and returns the
// tabs of the nearest style that defines its own w:tabs (the nearest wins).
func inheritedStyleTabs(styleID string, styleMap map[string]StyleDef) []ParsedTab {
	visited := make(map[string]bool)
	current := styleID
	for current != "" && !visited[current] {
		visited[current] = true
		sd, ok := styleMap[current]
		if !ok {
			break
		}
		if len(sd.Tabs) > 0 {
			return sd.Tabs
		}
		current = sd.BasedOn
	}
	return nil
}

// resolveEffectiveTabs merges paragraph-direct stops over style-inherited ones.
// A direct stop at the same position replaces the inherited stop; a direct
// w:val="clear" removes the stop at that position entirely. The result is sorted
// by position (the order the tab-stop dialog and the attribute contract use).
func resolveEffectiveTabs(direct, inherited []ParsedTab) []ParsedTab {
	byPos := make(map[float64]ParsedTab)
	var order []float64
	drop := func(pos float64) {
		if _, ok := byPos[pos]; ok {
			delete(byPos, pos)
			for i, p := range order {
				if p == pos {
					order = append(order[:i], order[i+1:]...)
					break
				}
			}
		}
	}
	for _, t := range inherited {
		byPos[t.Pos] = t
		order = append(order, t.Pos)
	}
	for _, t := range direct {
		if t.Align == "clear" {
			drop(t.Pos)
			continue
		}
		if _, ok := byPos[t.Pos]; !ok {
			order = append(order, t.Pos)
		}
		byPos[t.Pos] = t
	}
	sort.Float64s(order)
	out := make([]ParsedTab, 0, len(order))
	for _, pos := range order {
		out = append(out, byPos[pos])
	}
	return out
}

func parseParagraph(p DocPara, relMap map[string]string, styleMap map[string]StyleDef, styleNameMap map[string]string, numFmtMap map[string]string, numStartMap map[int]map[int]int, mode string, themeFontMap map[string]string) (ContentItem, []ContentItem) {
	lp := &ParsedParagraph{}

	if p.PPr != nil {
		if p.PPr.PStyle != nil {
			lp.StyleID = p.PPr.PStyle.Val
			if name, ok := styleNameMap[lp.StyleID]; ok {
				lp.StyleName = name
			}
			lp.HeadingLevel = resolveHeadingLevel(lp.StyleID, lp.StyleName, styleMap)
		}
		// Para-level w:outlineLvl overrides style-derived heading level
		if p.PPr.OutlineLvl != nil && lp.HeadingLevel == 0 {
			lvl := p.PPr.OutlineLvl.Val + 1
			if lvl >= 1 && lvl <= 9 {
				lp.HeadingLevel = lvl
			}
		}
		if p.PPr.Bidi.IsOn() {
			lp.Bidi = true
		}
		if p.PPr.Lang != nil && p.PPr.Lang.Val != "" {
			lp.Lang = p.PPr.Lang.Val
		}
		if p.PPr.TextAlign != nil && p.PPr.TextAlign.Val != "" {
			lp.VAlign = p.PPr.TextAlign.Val
		}
		if p.PPr.JC != nil && p.PPr.JC.Val != "" {
			lp.Align = p.PPr.JC.Val
		}
		fillParaLayout(lp, p.PPr, styleMap)
		if p.PPr.PBdr != nil {
			if p.PPr.PBdr.Top != nil {
				lp.BorderTop = &BorderInfo{Val: p.PPr.PBdr.Top.Val, Sz: p.PPr.PBdr.Top.Sz, Space: p.PPr.PBdr.Top.Space, Color: p.PPr.PBdr.Top.Color}
			}
			if p.PPr.PBdr.Bottom != nil {
				lp.BorderBottom = &BorderInfo{Val: p.PPr.PBdr.Bottom.Val, Sz: p.PPr.PBdr.Bottom.Sz, Space: p.PPr.PBdr.Bottom.Space, Color: p.PPr.PBdr.Bottom.Color}
			}
			if p.PPr.PBdr.Left != nil {
				lp.BorderLeft = &BorderInfo{Val: p.PPr.PBdr.Left.Val, Sz: p.PPr.PBdr.Left.Sz, Space: p.PPr.PBdr.Left.Space, Color: p.PPr.PBdr.Left.Color}
			}
			if p.PPr.PBdr.Right != nil {
				lp.BorderRight = &BorderInfo{Val: p.PPr.PBdr.Right.Val, Sz: p.PPr.PBdr.Right.Sz, Space: p.PPr.PBdr.Right.Space, Color: p.PPr.PBdr.Right.Color}
			}
		}
		if p.PPr.Shd != nil {
			lp.Shading = p.PPr.Shd.Fill
		}
		if p.PPr.KeepNext.IsOn() {
			lp.KeepNext = true
		}
		if p.PPr.KeepLines.IsOn() {
			lp.KeepLines = true
		}
		if p.PPr.WidowControl.IsOn() {
			lp.WidowControl = true
		}
		if p.PPr.RPr != nil {
			var defaults TextRun
			// Seed defaults from the paragraph's style rPr first (style inheritance),
			// then override with the explicit paragraph-level rPr.
			if lp.StyleID != "" {
				if sd, ok := styleMap[lp.StyleID]; ok {
					if sd.Bold {
						defaults.Bold = true
					}
					if sd.Italic {
						defaults.Italic = true
					}
					if sd.Underline != "" {
						defaults.Underline = sd.Underline
					}
					if sd.Strikethrough {
						defaults.Strike = true
					}
					if sd.SmallCaps {
						defaults.SmallCaps = true
					}
					if sd.Uppercase {
						defaults.AllCaps = true
					}
					if sd.FontEA != "" {
						defaults.FontEA = sd.FontEA
					}
					if sd.FontCS != "" {
						defaults.FontCS = sd.FontCS
					}
					if sd.Family != "" {
						defaults.FontFamily = sd.Family
					}
					if sd.SizePt > 0 {
						defaults.FontSizePt = sd.SizePt
					}
					if sd.SizeCS > 0 {
						defaults.FontSizeCS = sd.SizeCS
					}
					if sd.Color != "" {
						defaults.FontColor = sd.Color
					}
				}
			}
			applyParaRunDefaults(p.PPr.RPr, &defaults, themeFontMap)
			lp.ParaDefaults = &defaults
		}
		if p.PPr.SectPr != nil {
			if p.PPr.SectPr.Type != nil {
				lp.SectionBreak = p.PPr.SectPr.Type.Val
			} else {
				lp.SectionBreak = "nextPage"
			}
		}
		if p.PPr.PPrChange != nil {
			lp.RevisionAuthor = p.PPr.PPrChange.Author
			lp.RevisionDate = p.PPr.PPrChange.Date
		}
		if p.PPr.SuppressAutoHyph.IsOn() {
			lp.SuppressAutoHyph = true
		}
		if p.PPr.SnapToGrid.IsOn() {
			lp.SnapToGrid = true
		}
		if p.PPr.Kinsoku.IsOn() {
			lp.Kinsoku = true
		}
		if p.PPr.WordWrap.IsOn() {
			lp.WordWrap = true
		}
		if p.PPr.OverflowPunct.IsOn() {
			lp.OverflowPunct = true
		}
		if p.PPr.TopLinePunct.IsOn() {
			lp.TopLinePunct = true
		}
		if p.PPr.AutoSpaceDE.IsOn() {
			lp.AutoSpaceDE = true
		}
		if p.PPr.AutoSpaceDN.IsOn() {
			lp.AutoSpaceDN = true
		}
		if p.PPr.TextDirection != nil {
			lp.TextDirection = p.PPr.TextDirection.Val
		}
		if p.PPr.SuppressOverlap.IsOn() {
			lp.SuppressOverlap = true
		}
		if p.PPr.DivID != nil {
			lp.DivID = p.PPr.DivID.Val
		}
		if p.PPr.CnfStyle != nil {
			lp.CnfStyle = p.PPr.CnfStyle.Val
		}
		if p.PPr.FramePr != nil {
			fp := &FrameProps{
				DropCap: p.PPr.FramePr.DropCap,
				Lines:   p.PPr.FramePr.Lines,
				Width:   p.PPr.FramePr.W,
				Height:  p.PPr.FramePr.H,
				VSpace:  p.PPr.FramePr.VSpace,
				HSpace:  p.PPr.FramePr.HSpace,
				Wrap:    p.PPr.FramePr.Wrap,
				HAnchor: p.PPr.FramePr.HAnchor,
				VAnchor: p.PPr.FramePr.VAnchor,
				X:       p.PPr.FramePr.X,
				XAlign:  p.PPr.FramePr.XAlign,
				Y:       p.PPr.FramePr.Y,
				YAlign:  p.PPr.FramePr.YAlign,
				HRule:   p.PPr.FramePr.HRule,
			}
			if p.PPr.FramePr.AnchorLock != nil {
				fp.AnchorLock = true
			}
			lp.FramePr = fp
		}
	}

	isQuote := lp.StyleName == "Quote" || lp.StyleName == "IntenseQuote" || lp.StyleName == "BlockText"
	// Walk the basedOn chain to detect custom styles based on Quote/IntenseQuote/BlockText
	if !isQuote && lp.StyleID != "" {
		quoteAncestors := map[string]bool{"Quote": true, "IntenseQuote": true, "BlockText": true}
		visited := make(map[string]bool)
		current := lp.StyleID
		for current != "" && !visited[current] {
			visited[current] = true
			sd, ok := styleMap[current]
			if !ok {
				break
			}
			if quoteAncestors[sd.Name] {
				isQuote = true
				break
			}
			current = sd.BasedOn
		}
	}
	lp.IsQuote = isQuote

	isCode := false
	knownCodeStyles := map[string]bool{
		"code": true, "code block": true, "codeblock": true,
		"plain text": true, "plaintext": true,
		"source code": true, "sourcecode": true,
		"preformatted": true, "preformatted text": true,
		"source": true, "output": true,
	}
	knownCodeIDs := map[string]bool{
		"code": true, "codeblock": true, "codeblock1": true, "codeblock2": true,
		"plaintext": true, "sourcecode": true, "preformatted": true,
		"source": true, "output": true,
	}
	codeWords := []string{"code", "source", "output"}

	isKnownCode := false
	if lp.StyleName != "" {
		if knownCodeStyles[strings.ToLower(lp.StyleName)] {
			isKnownCode = true
		}
	}
	if !isKnownCode && lp.StyleID != "" {
		if knownCodeIDs[strings.ToLower(lp.StyleID)] {
			isKnownCode = true
		}
	}

	hasCodeWord := false
	if lp.StyleName != "" {
		sn := strings.ToLower(lp.StyleName)
		for _, w := range codeWords {
			if hasWordBoundary(sn, w) {
				hasCodeWord = true
				break
			}
		}
	}
	if !hasCodeWord && lp.StyleID != "" {
		sid := strings.ToLower(lp.StyleID)
		for _, w := range codeWords {
			if hasWordBoundary(sid, w) {
				hasCodeWord = true
				break
			}
		}
	}

	allMonospace := len(p.Runs) > 0
	if allMonospace {
		for _, r := range p.Runs {
			if r.RPr == nil || r.RPr.RFonts == nil {
				allMonospace = false
				break
			}
			if !isMonospaceFont(r.RPr.RFonts.Ascii) && !isMonospaceFont(r.RPr.RFonts.HAnsi) {
				allMonospace = false
				break
			}
		}
	}

	isCode = isKnownCode || (hasCodeWord && allMonospace)
	lp.IsCode = isCode

	runs, tbItems := extractRuns(p, styleMap, styleNameMap, relMap, numFmtMap, numStartMap, mode, isCode, themeFontMap, styleFontDefaults(p, styleMap, themeFontMap), lp.ParaDefaults)
	lp.Runs = runs

	// First-run lang fallback: if paragraph has no explicit lang, use the first run's lang
	if lp.Lang == "" {
		for _, r := range runs {
			if r.Lang != "" {
				lp.Lang = r.Lang
				break
			}
		}
	}

	if lp.HeadingLevel > 0 && !isCode {
		var textLen int
		for _, r := range runs {
			textLen += len(r.Text)
		}
		if textLen > 60 || lp.Align == "both" {
			lp.HeadingLevel = 0
		}
	}

	return ContentItem{Type: "paragraph", Paragraph: lp}, tbItems
}

// styleFontDefaults returns the merged font defaults for a paragraph:
// the paragraph style's ascii/hAnsi/eastAsia/cs fonts overridden by the
// paragraph-level rPr. It is used to seed font attributes onto runs whose
// own rPr does not fully specify the fonts (style -> paragraph rPr -> run rPr).
func styleFontDefaults(p DocPara, styleMap map[string]StyleDef, themeFontMap map[string]string) *TextRun {
	var fd TextRun
	if p.PPr != nil {
		if p.PPr.PStyle != nil {
			if sd, ok := styleMap[p.PPr.PStyle.Val]; ok {
				fd.FontFamily = sd.Family
				fd.FontEA = sd.FontEA
				fd.FontCS = sd.FontCS
			}
		}
		if p.PPr.RPr != nil {
			applyParaRunDefaults(p.PPr.RPr, &fd, themeFontMap)
		}
	}
	return &fd
}

func extractRuns(p DocPara, styleMap map[string]StyleDef, styleNameMap map[string]string, relMap map[string]string, numFmtMap map[string]string, numStartMap map[int]map[int]int, mode string, isCode bool, themeFontMap map[string]string, fontDefaults *TextRun, paraDefaults *TextRun) ([]TextRun, []ContentItem) {
	var runs []TextRun
	var tbItems []ContentItem

	var fieldURL string
	inFieldDisplay := false

	proc := func(r DocRun) ([]TextRun, []ContentItem) {
		var out []TextRun
		var items []ContentItem

		if r.FldChar != nil {
			switch r.FldChar.Type {
			case "begin":
				fieldURL = ""
				inFieldDisplay = false
			case "separate":
				inFieldDisplay = true
			case "end":
				fieldURL = ""
				inFieldDisplay = false
			}
			return out, items
		}
		if len(r.InstrText) > 0 {
			// Concatenate all instrText segments (field instruction can span multiple elements)
			var sb strings.Builder
			for _, it := range r.InstrText {
				sb.WriteString(it.Text)
			}
			instr := strings.TrimSpace(sb.String())
			if instr != "" {
				// Accumulate: a field instruction may arrive in pieces across runs
				if fieldURL == "" {
					if u, ok := extractHyperlinkURL(instr); ok {
						fieldURL = u
					}
				}
			}
			return out, items
		}

		if r.Pict != nil {
			if img := pictImage(r.Pict, relMap); img != nil {
				out = append(out, *img)
			}
			return out, items
		}

		if r.Drawing != nil {
			if r.Drawing.TxbxContent != nil {
				items = append(items, parseContentItemsOrdered(&r.Drawing.TxbxContent.DocOrderedContent, r.Drawing.TxbxContent.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)...)
			} else if img := drawingImage(r.Drawing, relMap); img != nil {
				out = append(out, *img)
			}
			return out, items
		}

		if r.Sym != nil {
			out = append(out, TextRun{IsSym: true, SymChar: r.Sym.Char})
			return out, items
		}

		text := ""
		preserveSpace := false
		// Use w:t for normal runs, fall back to w:delText for deleted runs
		textSrc := r.Text
		if len(textSrc) == 0 && len(r.DelText) > 0 {
			textSrc = r.DelText
		}
		for _, t := range textSrc {
			text += t.Text
			if t.Space == "preserve" {
				preserveSpace = true
			}
		}

		if mode == "semantic" && !preserveSpace && !isCode {
			text = normalizeWhitespace(text)
		}

		tr := TextRun{}
		// Seed the ascii/hAnsi/eastAsia/cs fonts inherited from the paragraph
		// style (and paragraph-level rPr) so that runs with a partial rPr still
		// carry the correct font. applyRunProps below only overrides the fields
		// the run's own rPr actually specifies.
		if fontDefaults != nil {
			tr.FontFamily = fontDefaults.FontFamily
			tr.FontEA = fontDefaults.FontEA
			tr.FontCS = fontDefaults.FontCS
		}
		if paraDefaults != nil && r.RPr == nil {
			tr = *paraDefaults
		}

		// Determine if this run has co-located text
		hasCoText := len(r.Text) > 0 || len(r.DelText) > 0

		// Tab: if no co-located text, emit immediately (run is tab-only).
		// If co-located text exists, emit tab before or after text per TabBeforeText.
		if r.Tab != nil && (!hasCoText || r.TabBeforeText) {
			tabRun := tr
			tabRun.IsTab = true
			out = append(out, tabRun)
			tr = TextRun{}
		}

		// Break: same logic as tab above
		if r.Break != nil && (!hasCoText || r.BreakBeforeText) {
			brRun := tr
			brRun.IsLineBreak = true
			brRun.BreakType = r.Break.Type
			if brRun.BreakType == "" {
				brRun.BreakType = "textWrapping"
			}
			out = append(out, brRun)
			tr = TextRun{}
		}

		// Cr: same logic
		if r.Cr != nil && (!hasCoText || r.BreakBeforeText) {
			out = append(out, TextRun{IsLineBreak: true, BreakType: "textWrapping"})
		}

		if r.NoBreakHyphen != nil {
			// Emit any co-located w:t text first, then the hyphen character
			if text != "" {
				textRun := tr
				textRun.Text = text
				applyRunProps(r.RPr, &textRun, themeFontMap)
				out = append(out, textRun)
			}
			hyphenRun := TextRun{Text: "\u00AC"}
			applyRunProps(r.RPr, &hyphenRun, themeFontMap)
			out = append(out, hyphenRun)
			return out, items
		}

		if r.SoftHyphen != nil {
			// Emit any co-located w:t text first, then the soft hyphen character
			if text != "" {
				textRun := tr
				textRun.Text = text
				applyRunProps(r.RPr, &textRun, themeFontMap)
				out = append(out, textRun)
			}
			hyphenRun := TextRun{Text: "\u00AD"}
			applyRunProps(r.RPr, &hyphenRun, themeFontMap)
			out = append(out, hyphenRun)
			return out, items
		}

		if r.FootnoteRef != nil {
			// Emit any co-located w:t text first, then the footnote ref
			if text != "" {
				textRun := tr
				textRun.Text = text
				applyRunProps(r.RPr, &textRun, themeFontMap)
				out = append(out, textRun)
			}
			refRun := TextRun{IsFootnoteRef: true, NoteID: r.FootnoteRef.ID, NoteType: "footnote"}
			out = append(out, refRun)
			return out, items
		}

		if r.EndnoteRef != nil {
			// Emit any co-located w:t text first, then the endnote ref
			if text != "" {
				textRun := tr
				textRun.Text = text
				applyRunProps(r.RPr, &textRun, themeFontMap)
				out = append(out, textRun)
			}
			refRun := TextRun{IsFootnoteRef: true, NoteID: r.EndnoteRef.ID, NoteType: "endnote"}
			out = append(out, refRun)
			return out, items
		}

		if text == "" {
			return out, items
		}

		tr.Text = text
		if inFieldDisplay && fieldURL != "" {
			tr.IsHyperlink = true
			tr.HyperlinkURL = fieldURL
		}
		applyRunProps(r.RPr, &tr, themeFontMap)
		out = append(out, tr)

		// Emit tab/break that appears AFTER text in source XML
		if r.Tab != nil && !r.TabBeforeText {
			tabRun := TextRun{IsTab: true}
			applyRunProps(r.RPr, &tabRun, themeFontMap)
			out = append(out, tabRun)
		}
		if r.Break != nil && !r.BreakBeforeText {
			brRun := TextRun{IsLineBreak: true, BreakType: r.Break.Type}
			if brRun.BreakType == "" {
				brRun.BreakType = "textWrapping"
			}
			out = append(out, brRun)
		}
		if r.Cr != nil && !r.BreakBeforeText {
			out = append(out, TextRun{IsLineBreak: true, BreakType: "textWrapping"})
		}
		return out, items
	}

	if p.PPr != nil && p.PPr.PageBreakBefore.IsOn() {
		runs = append(runs, TextRun{IsLineBreak: true, BreakType: "page"})
	}

	for _, r := range p.Runs {
		got, items := proc(r)
		runs = append(runs, got...)
		tbItems = append(tbItems, items...)

		if r.Ins != nil {
			for _, ir := range r.Ins.Runs {
				insRuns, _ := proc(ir)
				for i := range insRuns {
					insRuns[i].IsInserted = true
					insRuns[i].InsAuthor = r.Ins.Author
					insRuns[i].InsDate = r.Ins.Date
				}
				runs = append(runs, insRuns...)
			}
		}
		if r.Del != nil {
			for _, dr := range r.Del.Runs {
				delRuns, _ := proc(dr)
				for i := range delRuns {
					delRuns[i].IsDeleted = true
					delRuns[i].DelAuthor = r.Del.Author
					delRuns[i].DelDate = r.Del.Date
				}
				runs = append(runs, delRuns...)
			}
		}
	}

	// Process para-level w:ins (w:ins directly under w:p wrapping w:r)
	for _, ins := range p.ParaIns {
		for _, ir := range ins.Runs {
			insRuns, _ := proc(ir)
			for i := range insRuns {
				insRuns[i].IsInserted = true
				insRuns[i].InsAuthor = ins.Author
				insRuns[i].InsDate = ins.Date
			}
			runs = append(runs, insRuns...)
		}
	}

	// Process para-level w:del (w:del directly under w:p wrapping w:r)
	for _, del := range p.ParaDel {
		for _, dr := range del.Runs {
			delRuns, _ := proc(dr)
			for i := range delRuns {
				delRuns[i].IsDeleted = true
				delRuns[i].DelAuthor = del.Author
				delRuns[i].DelDate = del.Date
			}
			runs = append(runs, delRuns...)
		}
	}

	for _, hl := range p.Hyperlinks {
		// Resolve hyperlink URL once for all runs in this hyperlink
		hlURL := ""
		if hl.RID != "" {
			if u := relMap[hl.RID]; u != "" {
				hlURL = u
			}
		}
		if hlURL == "" && hl.Anchor != "" {
			hlURL = "#" + hl.Anchor
		}
		if hlURL == "" && hl.ID != "" {
			hlURL = hl.ID
		}
		for _, r := range hl.Runs {
			text := ""
			for _, t := range r.Text {
				text += t.Text
			}
			// Emit tab/break/cr even when there is no text
			hasCoText := text != ""
			if !hasCoText && r.Tab == nil && r.Break == nil && r.Cr == nil {
				continue
			}
			applyFormatting := func(tr *TextRun) {
				tr.IsHyperlink = true
				tr.HyperlinkURL = hlURL
				if r.RPr != nil {
					tr.Bold = r.RPr.B.IsOn()
					tr.Italic = r.RPr.I.IsOn()
					if r.RPr.U != nil && r.RPr.U.Val != "none" && r.RPr.U.Val != "" {
						tr.Underline = r.RPr.U.Val
					}
					tr.Strike = r.RPr.Strike.IsOn() || r.RPr.DStrike.IsOn()
				}
			}
			// Emit tab if present (before or after text depending on order)
			if r.Tab != nil && (!hasCoText || r.TabBeforeText) {
				tabRun := TextRun{IsTab: true}
				applyFormatting(&tabRun)
				runs = append(runs, tabRun)
			}
			// Emit break/cr if present before text
			if r.Break != nil && (!hasCoText || r.BreakBeforeText) {
				bt := r.Break.Type
				if bt == "" {
					bt = "textWrapping"
				}
				brRun := TextRun{IsLineBreak: true, BreakType: bt}
				applyFormatting(&brRun)
				runs = append(runs, brRun)
			}
			if r.Cr != nil && (!hasCoText || r.BreakBeforeText) {
				crRun := TextRun{IsLineBreak: true, BreakType: "textWrapping"}
				applyFormatting(&crRun)
				runs = append(runs, crRun)
			}
			// Emit text
			if hasCoText {
				tr := TextRun{Text: text}
				applyFormatting(&tr)
				runs = append(runs, tr)
			}
			// Emit tab/break/cr after text
			if r.Tab != nil && hasCoText && !r.TabBeforeText {
				tabRun := TextRun{IsTab: true}
				applyFormatting(&tabRun)
				runs = append(runs, tabRun)
			}
			if r.Break != nil && hasCoText && !r.BreakBeforeText {
				bt := r.Break.Type
				if bt == "" {
					bt = "textWrapping"
				}
				brRun := TextRun{IsLineBreak: true, BreakType: bt}
				applyFormatting(&brRun)
				runs = append(runs, brRun)
			}
			if r.Cr != nil && hasCoText && !r.BreakBeforeText {
				crRun := TextRun{IsLineBreak: true, BreakType: "textWrapping"}
				applyFormatting(&crRun)
				runs = append(runs, crRun)
			}
		}
	}

	for _, d := range p.Dir {
		isRTL := d.Val == "rtl"
		for _, r := range d.Runs {
			got, items := proc(r)
			if isRTL {
				for i := range got {
					got[i].IsRTL = true
				}
			}
			runs = append(runs, got...)
			tbItems = append(tbItems, items...)
			// Also handle w:ins/w:del inside w:dir runs
			if r.Ins != nil {
				for _, ir := range r.Ins.Runs {
					insRuns, _ := proc(ir)
					for i := range insRuns {
						insRuns[i].IsInserted = true
						insRuns[i].InsAuthor = r.Ins.Author
						insRuns[i].InsDate = r.Ins.Date
						if isRTL {
							insRuns[i].IsRTL = true
						}
					}
					runs = append(runs, insRuns...)
				}
			}
			if r.Del != nil {
				for _, dr := range r.Del.Runs {
					delRuns, _ := proc(dr)
					for i := range delRuns {
						delRuns[i].IsDeleted = true
						delRuns[i].DelAuthor = r.Del.Author
						delRuns[i].DelDate = r.Del.Date
						if isRTL {
							delRuns[i].IsRTL = true
						}
					}
					runs = append(runs, delRuns...)
				}
			}
		}
	}
	for _, d := range p.Bdo {
		isRTL := d.Val == "rtl"
		for _, r := range d.Runs {
			got, items := proc(r)
			if isRTL {
				for i := range got {
					got[i].IsRTL = true
				}
			}
			runs = append(runs, got...)
			tbItems = append(tbItems, items...)
			// Also handle w:ins/w:del inside w:bdo runs
			if r.Ins != nil {
				for _, ir := range r.Ins.Runs {
					insRuns, _ := proc(ir)
					for i := range insRuns {
						insRuns[i].IsInserted = true
						insRuns[i].InsAuthor = r.Ins.Author
						insRuns[i].InsDate = r.Ins.Date
						if isRTL {
							insRuns[i].IsRTL = true
						}
					}
					runs = append(runs, insRuns...)
				}
			}
			if r.Del != nil {
				for _, dr := range r.Del.Runs {
					delRuns, _ := proc(dr)
					for i := range delRuns {
						delRuns[i].IsDeleted = true
						delRuns[i].DelAuthor = r.Del.Author
						delRuns[i].DelDate = r.Del.Date
						if isRTL {
							delRuns[i].IsRTL = true
						}
					}
					runs = append(runs, delRuns...)
				}
			}
		}
	}

	return runs, tbItems
}

func parseTable(tbl DocTbl, relMap map[string]string, styleMap map[string]StyleDef, styleNameMap map[string]string, numFmtMap map[string]string, numStartMap map[int]map[int]int, mode string, themeFontMap map[string]string) *ParsedTable {
	t := &ParsedTable{}

	if tbl.TblPr != nil {
		if tbl.TblPr.TblStyle != nil {
			styleID := tbl.TblPr.TblStyle.Val
			if resolved, ok := styleNameMap[styleID]; ok {
				t.StyleName = resolved
			} else {
				t.StyleName = styleID
			}
		}
		if tbl.TblPr.TblW != nil {
			t.Width = float64(tbl.TblPr.TblW.W) / twipsPerInch
		}
		if tbl.TblPr.JC != nil {
			t.Alignment = tbl.TblPr.JC.Val
		}
		if tbl.TblPr.Ind != nil {
			t.Indent = float64(tbl.TblPr.Ind.W) / twipsPerInch
		}
		if tbl.TblPr.Spacing != nil {
			t.CellSpace = float64(tbl.TblPr.Spacing.W) / twipsPerInch
		}
		if tbl.TblPr.Caption != nil {
			t.Caption = tbl.TblPr.Caption.Val
		}
		if tbl.TblPr.Desc != nil {
			t.Summary = tbl.TblPr.Desc.Val
		}
		if tbl.TblPr.Borders != nil {
			if tbl.TblPr.Borders.Top != nil {
				t.BorderTop = &BorderInfo{Val: tbl.TblPr.Borders.Top.Val, Sz: tbl.TblPr.Borders.Top.Sz, Space: tbl.TblPr.Borders.Top.Space, Color: tbl.TblPr.Borders.Top.Color}
			}
			if tbl.TblPr.Borders.Bottom != nil {
				t.BorderBottom = &BorderInfo{Val: tbl.TblPr.Borders.Bottom.Val, Sz: tbl.TblPr.Borders.Bottom.Sz, Space: tbl.TblPr.Borders.Bottom.Space, Color: tbl.TblPr.Borders.Bottom.Color}
			}
			if tbl.TblPr.Borders.Left != nil {
				t.BorderLeft = &BorderInfo{Val: tbl.TblPr.Borders.Left.Val, Sz: tbl.TblPr.Borders.Left.Sz, Space: tbl.TblPr.Borders.Left.Space, Color: tbl.TblPr.Borders.Left.Color}
			}
			if tbl.TblPr.Borders.Right != nil {
				t.BorderRight = &BorderInfo{Val: tbl.TblPr.Borders.Right.Val, Sz: tbl.TblPr.Borders.Right.Sz, Space: tbl.TblPr.Borders.Right.Space, Color: tbl.TblPr.Borders.Right.Color}
			}
		}
	}

	if tbl.TblGrid != nil {
		for _, col := range tbl.TblGrid.Cols {
			t.Grid = append(t.Grid, float64(col.W)/twipsPerInch)
		}
	}

	for _, row := range tbl.Rows {
		pr := ParsedTableRow{}
		if row.TrPr != nil && row.TrPr.TblHeader.IsOn() {
			pr.IsHeader = true
		}
		for _, cell := range row.Cells {
			pc := ParsedTableCell{}
			if cell.TcPr != nil {
				if cell.TcPr.GridSpan != nil {
					pc.GridSpan = cell.TcPr.GridSpan.Val
				}
				if cell.TcPr.VMerge != nil {
					if cell.TcPr.VMerge.Val == "restart" {
						pc.VMerge = 1
					} else {
						pc.VMerge = 2
					}
				}
				if cell.TcPr.VAlign != nil {
					pc.VAlign = cell.TcPr.VAlign.Val
				}
				if cell.TcPr.TextDir != nil {
					pc.TextDir = cell.TcPr.TextDir.Val
				}
				pc.NoWrap = cell.TcPr.NoWrap.IsOn()
				if cell.TcPr.TcW != nil && cell.TcPr.TcW.W > 0 {
					pc.Width = float64(cell.TcPr.TcW.W) / twipsPerInch
				}
				if cell.TcPr.Borders != nil {
					if cell.TcPr.Borders.Top != nil {
						pc.BorderTop = &BorderInfo{Val: cell.TcPr.Borders.Top.Val, Sz: cell.TcPr.Borders.Top.Sz, Space: cell.TcPr.Borders.Top.Space, Color: cell.TcPr.Borders.Top.Color}
					}
					if cell.TcPr.Borders.Bottom != nil {
						pc.BorderBottom = &BorderInfo{Val: cell.TcPr.Borders.Bottom.Val, Sz: cell.TcPr.Borders.Bottom.Sz, Space: cell.TcPr.Borders.Bottom.Space, Color: cell.TcPr.Borders.Bottom.Color}
					}
					if cell.TcPr.Borders.Left != nil {
						pc.BorderLeft = &BorderInfo{Val: cell.TcPr.Borders.Left.Val, Sz: cell.TcPr.Borders.Left.Sz, Space: cell.TcPr.Borders.Left.Space, Color: cell.TcPr.Borders.Left.Color}
					}
					if cell.TcPr.Borders.Right != nil {
						pc.BorderRight = &BorderInfo{Val: cell.TcPr.Borders.Right.Val, Sz: cell.TcPr.Borders.Right.Sz, Space: cell.TcPr.Borders.Right.Space, Color: cell.TcPr.Borders.Right.Color}
					}
				}
			}
			pc.Content = parseContentItemsOrdered(&cell.DocOrderedContent, cell.Children, relMap, styleMap, styleNameMap, numFmtMap, numStartMap, mode, themeFontMap)
			for _, ci := range pc.Content {
				if ci.Type == "paragraph" && ci.Paragraph != nil {
					for _, r := range ci.Paragraph.Runs {
						if r.Lang != "" {
							pc.Lang = r.Lang
							break
						}
					}
					if pc.Lang != "" {
						break
					}
				}
			}
			pr.Cells = append(pr.Cells, pc)
		}
		t.Rows = append(t.Rows, pr)
	}

	type vmRestart struct {
		col     int
		spanCnt int
	}
	restarts := make(map[int]*vmRestart)
	for ri := range t.Rows {
		colIdx := 0
		for ci := range t.Rows[ri].Cells {
			cell := &t.Rows[ri].Cells[ci]
			gs := cell.GridSpan
			if gs < 1 {
				gs = 1
			}
			col := colIdx
			colIdx += gs
			if cell.VMerge == 1 {
				restarts[ri] = &vmRestart{col: col, spanCnt: 0}
			}
		}
	}
	for ri := range t.Rows {
		colIdx := 0
		for ci := range t.Rows[ri].Cells {
			cell := &t.Rows[ri].Cells[ci]
			gs := cell.GridSpan
			if gs < 1 {
				gs = 1
			}
			col := colIdx
			colIdx += gs
			if cell.VMerge == 2 {
				for prev := ri - 1; prev >= 0; prev-- {
					if r, ok := restarts[prev]; ok && r.col == col {
						r.spanCnt++
						cell.Omitted = true
						break
					}
				}
				if !cell.Omitted {
					cell.RowSpan = 2
				}
			}
		}
	}
	for ri, r := range restarts {
		totalSpan := r.spanCnt + 1
		if totalSpan > 1 {
			colIdx := 0
			for ci := range t.Rows[ri].Cells {
				gs := t.Rows[ri].Cells[ci].GridSpan
				if gs < 1 {
					gs = 1
				}
				if colIdx == r.col {
					t.Rows[ri].Cells[ci].RowSpan = totalSpan
					break
				}
				colIdx += gs
			}
		}
	}

	return t
}

// --- Run Processor ---

func extractTheme(themeXML []byte) *ThemeData {
	var theme ATheme
	if err := xml.Unmarshal(themeXML, &theme); err != nil {
		return nil
	}
	td := &ThemeData{FontMap: make(map[string]string)}
	if theme.ThemeElements != nil {
		if fs := theme.ThemeElements.FontScheme; fs != nil {
			if fs.Minor != nil {
				if fs.Minor.Latin != nil && fs.Minor.Latin.Typeface != "" {
					td.Font = fs.Minor.Latin.Typeface
				}
				if fs.Minor.Ea != nil && fs.Minor.Ea.Typeface != "" {
					td.FontEA = fs.Minor.Ea.Typeface
				}
				if fs.Minor.Cs != nil && fs.Minor.Cs.Typeface != "" {
					td.FontCS = fs.Minor.Cs.Typeface
				}
				if fs.Minor.Latin != nil && fs.Minor.Latin.Typeface != "" {
					td.FontMap["minorHAnsi"] = fs.Minor.Latin.Typeface
					td.FontMap["minorAscii"] = fs.Minor.Latin.Typeface
				}
				if fs.Minor.Ea != nil && fs.Minor.Ea.Typeface != "" {
					td.FontMap["minorEastAsia"] = fs.Minor.Ea.Typeface
				}
				if fs.Minor.Cs != nil && fs.Minor.Cs.Typeface != "" {
					td.FontMap["minorBidi"] = fs.Minor.Cs.Typeface
				}
			}
			if fs.Major != nil {
				if fs.Major.Latin != nil && fs.Major.Latin.Typeface != "" {
					td.FontMap["majorHAnsi"] = fs.Major.Latin.Typeface
					td.FontMap["majorAscii"] = fs.Major.Latin.Typeface
				}
				if fs.Major.Ea != nil && fs.Major.Ea.Typeface != "" {
					td.FontMap["majorEastAsia"] = fs.Major.Ea.Typeface
				}
				if fs.Major.Cs != nil && fs.Major.Cs.Typeface != "" {
					td.FontMap["majorBidi"] = fs.Major.Cs.Typeface
				}
			}
		}
		if cs := theme.ThemeElements.ClrScheme; cs != nil {
			if cs.Dk1 != nil && cs.Dk1.SrgbClr != nil {
				td.Fg = cs.Dk1.SrgbClr.Val
			}
			if cs.Lt1 != nil && cs.Lt1.SrgbClr != nil {
				td.Bg = cs.Lt1.SrgbClr.Val
			}
		}
	}
	if td.Font == "" && td.Fg == "" {
		return nil
	}
	return td
}

func drawingImage(d *DocDrawing, relMap map[string]string) *TextRun {
	if d == nil {
		return nil
	}
	var inline *WpInline
	if d.Inline != nil {
		inline = d.Inline
	} else if d.Anchor != nil {
		inline = &WpInline{
			Extent:  d.Anchor.Extent,
			DocPr:   d.Anchor.DocPr,
			Graphic: d.Anchor.Graphic,
		}
	}
	if inline == nil || inline.Graphic == nil || inline.Graphic.GraphicData == nil || inline.Graphic.GraphicData.Pic == nil {
		return nil
	}
	embed := inline.Graphic.GraphicData.Pic.BlipFill.Blip.Embed
	src := relMap[embed]
	if src == "" {
		src = embed
	}
	tr := &TextRun{
		IsImage:  true,
		ImageSrc: src,
		ImageAlt: inline.DocPr.Desc,
	}
	if inline.Extent != nil {
		tr.ImageWidth = float64(inline.Extent.Cx) / 914400.0
		tr.ImageHeight = float64(inline.Extent.Cy) / 914400.0
	}
	return tr
}

func pictImage(pict *DocPict, relMap map[string]string) *TextRun {
	if pict == nil || pict.Content == "" {
		return nil
	}
	matches := vmlImageRe.FindStringSubmatch(pict.Content)
	if matches == nil {
		return nil
	}
	rid := ""
	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			rid = matches[i]
			break
		}
	}
	if rid == "" {
		return nil
	}
	src := relMap[rid]
	if src == "" {
		src = rid
	}
	return &TextRun{
		IsImage:  true,
		ImageSrc: src,
	}
}

func extractHyperlinkURL(instr string) (string, bool) {
	instr = strings.TrimSpace(instr)
	if !strings.HasPrefix(instr, "HYPERLINK") {
		return "", false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(instr, "HYPERLINK"))

	// Handle \l "anchor" — local bookmark link (no URL before \l)
	// e.g. HYPERLINK \l "section1"  -> #section1
	if strings.HasPrefix(rest, `\l`) {
		anchor := strings.TrimSpace(strings.TrimPrefix(rest, `\l`))
		anchor = strings.Trim(anchor, `"`)
		if anchor != "" {
			return "#" + anchor, true
		}
		return "", false
	}

	// Extract quoted URL (may be followed by options like \o "tooltip" \t "_blank")
	// e.g. HYPERLINK "https://example.com" \o "tooltip"
	if len(rest) > 0 && rest[0] == '"' {
		end := strings.Index(rest[1:], `"`)
		if end >= 0 {
			url := rest[1 : end+1]
			// Check for \l anchor appended: HYPERLINK "url" \l "anchor"
			after := strings.TrimSpace(rest[end+2:])
			if strings.HasPrefix(after, `\l`) {
				anchor := strings.TrimSpace(strings.TrimPrefix(after, `\l`))
				anchor = strings.Trim(anchor, `"`)
				if anchor != "" {
					return url + "#" + anchor, true
				}
			}
			return url, true
		}
	}

	// Bare URL (no quotes)
	if rest != "" {
		// Strip any trailing \x options
		if idx := strings.Index(rest, ` \`); idx >= 0 {
			rest = strings.TrimSpace(rest[:idx])
		}
		return rest, true
	}
	return "", false
}

func applyRunProps(rPr *RunProps, tr *TextRun, themeFontMap map[string]string) {
	if rPr == nil {
		return
	}
	if rPr.B.IsOn() {
		tr.Bold = true
	}
	if rPr.BCs.IsOn() {
		tr.IsBoldCS = true
	}
	if rPr.I.IsOn() {
		tr.Italic = true
	}
	if rPr.ICs.IsOn() {
		tr.IsItalicCS = true
	}
	if rPr.U != nil {
		tr.Underline = rPr.U.Val
	}
	if rPr.Strike.IsOn() || rPr.DStrike.IsOn() {
		tr.Strike = true
	}
	if rPr.SmallCaps.IsOn() {
		tr.SmallCaps = true
	}
	if rPr.Caps.IsOn() {
		tr.AllCaps = true
	}
	if rPr.VertAlign != nil {
		if rPr.VertAlign.Val == "superscript" {
			tr.SuperScript = true
		} else if rPr.VertAlign.Val == "subscript" {
			tr.SubScript = true
		}
	}
	if rPr.RFonts != nil {
		if rPr.RFonts.Ascii != "" {
			tr.FontFamily = rPr.RFonts.Ascii
		} else if rPr.RFonts.HAnsi != "" {
			tr.FontFamily = rPr.RFonts.HAnsi
		} else if rPr.RFonts.AsciiTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.AsciiTheme]; ok {
				tr.FontFamily = f
			}
		} else if rPr.RFonts.HAnsiTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.HAnsiTheme]; ok {
				tr.FontFamily = f
			}
		}
		if rPr.RFonts.EastAsia != "" {
			tr.FontEA = rPr.RFonts.EastAsia
		} else if rPr.RFonts.EastAsiaTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.EastAsiaTheme]; ok {
				tr.FontEA = f
			}
		}
		if rPr.RFonts.CS != "" {
			tr.FontCS = rPr.RFonts.CS
		} else if rPr.RFonts.CSTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.CSTheme]; ok {
				tr.FontCS = f
			}
		}
	}
	if rPr.Sz != nil {
		tr.FontSizePt = float64(rPr.Sz.Val) / 2.0
	}
	if rPr.SzCs != nil {
		tr.FontSizeCS = float64(rPr.SzCs.Val) / 2.0
	}
	if rPr.Color != nil {
		tr.FontColor = rPr.Color.Val
	}
	if rPr.Highlight != nil {
		tr.Highlight = rPr.Highlight.Val
	}
	tr.Hidden = rPr.Vanish.IsOn()
	if rPr.Lang != nil {
		tr.Lang = rPr.Lang.Val
	}
	tr.IsRTL = rPr.Rtl.IsOn()
}

func applyParaRunDefaults(rPr *RunProps, tr *TextRun, themeFontMap map[string]string) {
	if rPr == nil {
		return
	}
	if rPr.B.IsOn() {
		tr.Bold = true
	}
	if rPr.BCs.IsOn() {
		tr.IsBoldCS = true
	}
	if rPr.I.IsOn() {
		tr.Italic = true
	}
	if rPr.ICs.IsOn() {
		tr.IsItalicCS = true
	}
	if rPr.U != nil {
		tr.Underline = rPr.U.Val
	}
	if rPr.Strike.IsOn() || rPr.DStrike.IsOn() {
		tr.Strike = true
	}
	if rPr.SmallCaps.IsOn() {
		tr.SmallCaps = true
	}
	if rPr.Caps.IsOn() {
		tr.AllCaps = true
	}
	if rPr.VertAlign != nil {
		if rPr.VertAlign.Val == "superscript" {
			tr.SuperScript = true
		} else if rPr.VertAlign.Val == "subscript" {
			tr.SubScript = true
		}
	}
	if rPr.RFonts != nil {
		if rPr.RFonts.Ascii != "" {
			tr.FontFamily = rPr.RFonts.Ascii
		} else if rPr.RFonts.HAnsi != "" {
			tr.FontFamily = rPr.RFonts.HAnsi
		} else if rPr.RFonts.AsciiTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.AsciiTheme]; ok {
				tr.FontFamily = f
			}
		} else if rPr.RFonts.HAnsiTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.HAnsiTheme]; ok {
				tr.FontFamily = f
			}
		}
		if rPr.RFonts.EastAsia != "" {
			tr.FontEA = rPr.RFonts.EastAsia
		} else if rPr.RFonts.EastAsiaTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.EastAsiaTheme]; ok {
				tr.FontEA = f
			}
		}
		if rPr.RFonts.CS != "" {
			tr.FontCS = rPr.RFonts.CS
		} else if rPr.RFonts.CSTheme != "" {
			if f, ok := themeFontMap[rPr.RFonts.CSTheme]; ok {
				tr.FontCS = f
			}
		}
	}
	if rPr.Sz != nil {
		tr.FontSizePt = float64(rPr.Sz.Val) / 2.0
	}
	if rPr.SzCs != nil {
		tr.FontSizeCS = float64(rPr.SzCs.Val) / 2.0
	}
	if rPr.Color != nil {
		tr.FontColor = rPr.Color.Val
	}
	if rPr.Highlight != nil {
		tr.Highlight = rPr.Highlight.Val
	}
	if rPr.Vanish.IsOn() {
		tr.Hidden = true
	}
	if rPr.Lang != nil {
		tr.Lang = rPr.Lang.Val
	}
	tr.IsRTL = rPr.Rtl.IsOn()
}

func formatPt(v float64) string {
	return strings.TrimSuffix(fmt.Sprintf("%.1f", v), ".0")
}

func formatBorderStyle(val string) string {
	switch val {
	case "single":
		return "s"
	case "double":
		return "d"
	case "dashed", "dashSmallGap":
		return "ds"
	case "dotted":
		return "dt"
	case "none", "nil":
		return "n"
	default:
		return "s"
	}
}

func buildBorderAttr(top, bot, left, right *BorderInfo) string {
	var parts []string
	if top != nil && top.Val != "none" && top.Val != "" {
		w := float64(top.Sz) / 576.0
		parts = append(parts, fmt.Sprintf("bt %.3f %s%d #%s", w, formatBorderStyle(top.Val), top.Space, top.Color))
	}
	if bot != nil && bot.Val != "none" && bot.Val != "" {
		w := float64(bot.Sz) / 576.0
		parts = append(parts, fmt.Sprintf("bb %.3f %s%d #%s", w, formatBorderStyle(bot.Val), bot.Space, bot.Color))
	}
	if left != nil && left.Val != "none" && left.Val != "" {
		w := float64(left.Sz) / 576.0
		parts = append(parts, fmt.Sprintf("bl %.3f %s%d #%s", w, formatBorderStyle(left.Val), left.Space, left.Color))
	}
	if right != nil && right.Val != "none" && right.Val != "" {
		w := float64(right.Sz) / 576.0
		parts = append(parts, fmt.Sprintf("br %.3f %s%d #%s", w, formatBorderStyle(right.Val), right.Space, right.Color))
	}
	return strings.Join(parts, "; ")
}

// --- Text Cleanup ---

func xmlEscape(s string) string {
	s = stripControlChars(s)
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func stripControlChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' || r == '\n' || r == '\r' {
			b.WriteRune(r)
			continue
		}
		if r < 0x20 || (r >= 0x7F && r <= 0x84) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func normalizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\t", " ")
	var prev string
	for prev != s {
		prev = s
		s = strings.ReplaceAll(s, "  ", " ")
	}
	s = strings.TrimLeft(s, "\n")
	s = strings.TrimRight(s, "\n ")
	return s
}

// --- Emitter ---

func formatForLLM(doc *ParsedDocument) string {
	var b strings.Builder

	mode := doc.Mode
	if mode == "" {
		mode = "semantic"
	}
	b.WriteString(fmt.Sprintf("<words xmlns=\"urn:words:v1\" xmlns:s=\"urn:words:v1:style\" version=\"1.1.0\" mode=\"%s\">\n", mode))

	if doc.Meta.Title != "" || doc.Meta.Author != "" || doc.Meta.Created != "" || doc.Meta.Modified != "" || doc.Meta.Keywords != "" {
		b.WriteString("  <meta>\n")
		if doc.Meta.Title != "" {
			b.WriteString(fmt.Sprintf("    <title>%s</title>\n", xmlEscape(doc.Meta.Title)))
		}
		if doc.Meta.Author != "" {
			b.WriteString(fmt.Sprintf("    <author>%s</author>\n", xmlEscape(doc.Meta.Author)))
		}
		if doc.Meta.Created != "" {
			b.WriteString(fmt.Sprintf("    <created>%s</created>\n", xmlEscape(doc.Meta.Created)))
		}
		if doc.Meta.Modified != "" {
			b.WriteString(fmt.Sprintf("    <modified>%s</modified>\n", xmlEscape(doc.Meta.Modified)))
		}
		if doc.Meta.Keywords != "" {
			b.WriteString(fmt.Sprintf("    <keywords>%s</keywords>\n", xmlEscape(doc.Meta.Keywords)))
		}
		b.WriteString("  </meta>\n")
	}

	emitStyleBlock(&b, doc)

	for _, h := range doc.Headers {
		fmt.Fprintf(&b, "  <header id=\"%d\">\n", h.ID)
		for _, ci := range h.Content {
			emitContentItem(&b, ci, doc, "    ")
		}
		b.WriteString("  </header>\n")
	}
	for _, f := range doc.Footers {
		fmt.Fprintf(&b, "  <footer id=\"%d\">\n", f.ID)
		for _, ci := range f.Content {
			emitContentItem(&b, ci, doc, "    ")
		}
		b.WriteString("  </footer>\n")
	}

	b.WriteString("  <write>\n")
	emitContent(&b, doc)
	b.WriteString("  </write>\n")

	if len(doc.Notes) > 0 {
		b.WriteString("  <notes>\n")
		for _, n := range doc.Notes {
			switch n.Type {
			case "bm":
				fmt.Fprintf(&b, "    <bm id=\"%s\"/>\n", xmlEscape(n.Name))
			case "footnote", "endnote":
				fmt.Fprintf(&b, "    <fn id=\"%d\" type=\"%s\"", n.ID, n.Type)
				if len(n.Body) == 0 {
					b.WriteString("/>\n")
				} else {
					b.WriteString(">\n")
					for _, ci := range n.Body {
						switch ci.Type {
						case "paragraph":
							formatParagraph(&b, ci.Paragraph, doc)
							b.WriteString("\n")
						case "list":
							emitContentItem(&b, ci, doc, "    ")
						case "table":
							writeTableIndent(&b, ci.Table, doc, "    ")
						}
					}
					b.WriteString("    </fn>\n")
				}
			case "comment":
				fmt.Fprintf(&b, "    <comment id=\"%d\"", n.ID)
				if n.Author != "" {
					fmt.Fprintf(&b, " author=\"%s\"", xmlEscape(n.Author))
				}
				if n.Date != "" {
					fmt.Fprintf(&b, " date=\"%s\"", xmlEscape(n.Date))
				}
				b.WriteString(">")
				for _, ci := range n.Body {
					switch ci.Type {
					case "paragraph":
						formatParagraph(&b, ci.Paragraph, doc)
					case "list":
						emitContentItem(&b, ci, doc, "    ")
					case "table":
						writeTableIndent(&b, ci.Table, doc, "    ")
					}
				}
				b.WriteString("</comment>\n")
			}
		}
		b.WriteString("  </notes>\n")
	}

	b.WriteString("</words>\n")
	return b.String()
}

func emitStyleBlock(b *strings.Builder, doc *ParsedDocument) {
	b.WriteString("  <style unit=\"in\">\n")

	for _, sec := range doc.PageSections {
		if sec.WidthInch <= 0 && sec.HeightInch <= 0 {
			continue
		}
		wPt := int(sec.WidthInch * 72)
		hPt := int(sec.HeightInch * 72)
		if preset := matchPageSize(wPt, hPt); preset != "" {
			fmt.Fprintf(b, "    <s:page size=\"%s\" mt=\"%.2f\" mb=\"%.2f\" ml=\"%.2f\" mr=\"%.2f\" mh=\"%.2f\" mf=\"%.2f\"/>\n",
				preset,
				sec.MarginTop, sec.MarginBottom, sec.MarginLeft, sec.MarginRight,
				sec.HeaderMargin, sec.FooterMargin)
		} else {
			fmt.Fprintf(b, "    <s:page w=\"%.2f\" h=\"%.2f\" mt=\"%.2f\" mb=\"%.2f\" ml=\"%.2f\" mr=\"%.2f\" mh=\"%.2f\" mf=\"%.2f\"/>\n",
				sec.WidthInch, sec.HeightInch,
				sec.MarginTop, sec.MarginBottom, sec.MarginLeft, sec.MarginRight,
				sec.HeaderMargin, sec.FooterMargin)
		}
		if sec.Cols > 1 {
			fmt.Fprintf(b, "    <s:cols n=\"%d\"", sec.Cols)
			if sec.ColsSpace > 0 {
				fmt.Fprintf(b, " space=\"%.2f\"", sec.ColsSpace)
			}
			b.WriteString("/>\n")
		}
	}

	for level := 1; level <= 9; level++ {
		headingName := fmt.Sprintf("Heading%d", level)
		if sd, ok := doc.StyleMap[headingName]; ok && (sd.SpacingBefore > 0 || sd.SpacingAfter > 0) {
			fmt.Fprintf(b, "    <s:gap el=\"h\" c=\"%s\"", headingName)
			if sd.SpacingBefore > 0 {
				fmt.Fprintf(b, " before=\"%.2f\"", sd.SpacingBefore)
			}
			if sd.SpacingAfter > 0 {
				fmt.Fprintf(b, " after=\"%.2f\"", sd.SpacingAfter)
			}
			b.WriteString("/>\n")
		}
	}

	if def, ok := doc.StyleMap["Normal"]; ok && (def.SpacingBefore > 0 || def.SpacingAfter > 0) {
		fmt.Fprintf(b, "    <s:gap el=\"p\"")
		if def.SpacingBefore > 0 {
			fmt.Fprintf(b, " before=\"%.2f\"", def.SpacingBefore)
		}
		if def.SpacingAfter > 0 {
			fmt.Fprintf(b, " after=\"%.2f\"", def.SpacingAfter)
		}
		b.WriteString("/>\n")
	}

	if def, ok := doc.StyleMap["Normal"]; ok {
		if def.IndentLeft > 0 || def.IndentRight > 0 || def.IndentFirst > 0 || def.IndentHanging > 0 {
			b.WriteString("    <s:indent el=\"p\"")
			if def.IndentLeft > 0 {
				fmt.Fprintf(b, " left=\"%.2f\"", def.IndentLeft)
			}
			if def.IndentRight > 0 {
				fmt.Fprintf(b, " right=\"%.2f\"", def.IndentRight)
			}
			if def.IndentFirst > 0 {
				fmt.Fprintf(b, " firstLine=\"%.2f\"", def.IndentFirst)
			}
			if def.IndentHanging > 0 {
				fmt.Fprintf(b, " hanging=\"%.2f\"", def.IndentHanging)
			}
			b.WriteString("/>\n")
		}
		if def.Align != "" {
			fmt.Fprintf(b, "    <s:align el=\"p\" value=\"%s\"/>\n", def.Align)
		}
		if def.LineSpacing > 0 {
			fmt.Fprintf(b, "    <s:line el=\"p\" value=\"%.2f\"", def.LineSpacing)
			if def.LineRule != "" {
				fmt.Fprintf(b, " rule=\"%s\"", def.LineRule)
			}
			b.WriteString("/>\n")
		}
	}

	for level := 1; level <= 9; level++ {
		headingName := fmt.Sprintf("Heading%d", level)
		if sd, ok := doc.StyleMap[headingName]; ok {
			if sd.IndentLeft > 0 || sd.IndentRight > 0 || sd.IndentFirst > 0 || sd.IndentHanging > 0 {
				b.WriteString("    <s:indent el=\"p\"")
				fmt.Fprintf(b, " c=\"%s\"", headingName)
				if sd.IndentLeft > 0 {
					fmt.Fprintf(b, " left=\"%.2f\"", sd.IndentLeft)
				}
				if sd.IndentRight > 0 {
					fmt.Fprintf(b, " right=\"%.2f\"", sd.IndentRight)
				}
				if sd.IndentFirst > 0 {
					fmt.Fprintf(b, " firstLine=\"%.2f\"", sd.IndentFirst)
				}
				if sd.IndentHanging > 0 {
					fmt.Fprintf(b, " hanging=\"%.2f\"", sd.IndentHanging)
				}
				b.WriteString("/>\n")
			}
			if sd.Align != "" {
				fmt.Fprintf(b, "    <s:align el=\"p\" c=\"%s\" value=\"%s\"/>\n", headingName, sd.Align)
			}
			if sd.LineSpacing > 0 {
				fmt.Fprintf(b, "    <s:line el=\"p\" c=\"%s\" value=\"%.2f\"", headingName, sd.LineSpacing)
				if sd.LineRule != "" {
					fmt.Fprintf(b, " rule=\"%s\"", sd.LineRule)
				}
				b.WriteString("/>\n")
			}
		}
	}

	for _, tbl := range doc.AllTables {
		for _, w := range tbl.Grid {
			fmt.Fprintf(b, "    <s:col ref=\"%d\" w=\"%.2f\"/>\n", tbl.ID, w)
		}
	}

	tabSeen := make(map[string]bool)
	emitTabs := func(items []ContentItem) {
		for _, item := range items {
			if item.Type == "paragraph" && item.Paragraph != nil && len(item.Paragraph.Tabs) > 0 {
				var el string
				if item.Paragraph.HeadingLevel > 0 {
					el = fmt.Sprintf("h%d", item.Paragraph.HeadingLevel)
				} else {
					el = "p"
				}
				for _, t := range item.Paragraph.Tabs {
					key := fmt.Sprintf("%s_%.2f_%s_%s", el, t.Pos, t.Align, t.Leader)
					if tabSeen[key] {
						continue
					}
					tabSeen[key] = true
					fmt.Fprintf(b, "    <s:tab el=\"%s\" pos=\"%.2f\" align=\"%s\" leader=\"%s\"/>\n",
						el, t.Pos, t.Align, t.Leader)
				}
			}
		}
	}
	emitTabs(doc.Content)
	for _, hdr := range doc.Headers {
		emitTabs(hdr.Content)
	}
	for _, ftr := range doc.Footers {
		emitTabs(ftr.Content)
	}

	builtinIDs := map[string]bool{
		"Normal": true, "DefaultParagraphFont": true, "Heading1": true, "Heading2": true,
		"Heading3": true, "Heading4": true, "Heading5": true, "Heading6": true,
		"Heading7": true, "Heading8": true, "Heading9": true, "Title": true, "Subtitle": true,
		"Quote": true, "IntenseQuote": true, "BlockText": true, "ListParagraph": true,
		"ListBullet": true, "ListNumber": true, "Caption": true, "TOCHeading": true,
		"Hyperlink": true, "FootnoteText": true, "EndnoteText": true, "FootnoteReference": true,
		"EndnoteReference": true, "CommentText": true, "Header": true, "Footer": true,
	}
	ids := make([]string, 0, len(doc.StyleMap))
	for id := range doc.StyleMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		sd := doc.StyleMap[id]
		for _, t := range sd.Tabs {
			var el string
			if sd.HeadingLevel > 0 {
				el = fmt.Sprintf("h%d", sd.HeadingLevel)
			} else {
				el = "p"
			}
			key := fmt.Sprintf("%s_%.2f_%s_%s", el, t.Pos, t.Align, t.Leader)
			if tabSeen[key] {
				continue
			}
			tabSeen[key] = true
			fmt.Fprintf(b, "    <s:tab el=\"%s\" pos=\"%.2f\" align=\"%s\" leader=\"%s\"/>\n",
				el, t.Pos, t.Align, t.Leader)
		}
	}
	for _, id := range ids {
		sd := doc.StyleMap[id]
		if builtinIDs[id] {
			continue
		}
		name := sd.Name
		if name == "" {
			name = id
		}
		if name == "Normal" || name == "Default Paragraph Font" {
			continue
		}
		b.WriteString("    <s:custom")
		fmt.Fprintf(b, " name=\"%s\"", xmlEscape(name))
		if sd.Type != "" {
			fmt.Fprintf(b, " type=\"%s\"", sd.Type)
		}
		if sd.BasedOn != "" {
			fmt.Fprintf(b, " basedOn=\"%s\"", xmlEscape(sd.BasedOn))
		}
		if sd.Family != "" {
			fmt.Fprintf(b, " font=\"%s\"", xmlEscape(sd.Family))
		}
		if sd.FontEA != "" {
			fmt.Fprintf(b, " fontEA=\"%s\"", xmlEscape(sd.FontEA))
		}
		if sd.FontCS != "" {
			fmt.Fprintf(b, " fontCS=\"%s\"", xmlEscape(sd.FontCS))
		}
		if sd.SizePt > 0 {
			fmt.Fprintf(b, " size=\"%spt\"", formatPt(sd.SizePt))
		}
		if sd.SizeCS > 0 {
			fmt.Fprintf(b, " sizeCS=\"%spt\"", formatPt(sd.SizeCS))
		}
		if sd.Color != "" {
			fmt.Fprintf(b, " color=\"%s\"", strings.TrimPrefix(sd.Color, "#"))
		}
		if sd.Bold {
			b.WriteString(" bold=\"true\"")
		}
		if sd.Italic {
			b.WriteString(" italic=\"true\"")
		}
		if sd.Underline != "" {
			fmt.Fprintf(b, " underline=\"%s\"", sd.Underline)
		}
		if sd.Strikethrough {
			b.WriteString(" strikethrough=\"true\"")
		}
		if sd.SmallCaps {
			b.WriteString(" smallCaps=\"true\"")
		}
		if sd.Uppercase {
			b.WriteString(" uppercase=\"true\"")
		}
		if sd.Align != "" {
			fmt.Fprintf(b, " alignment=\"%s\"", sd.Align)
		}
		if sd.SpacingBefore > 0 {
			fmt.Fprintf(b, " spacingBefore=\"%.2f\"", sd.SpacingBefore)
		}
		if sd.SpacingAfter > 0 {
			fmt.Fprintf(b, " spacingAfter=\"%.2f\"", sd.SpacingAfter)
		}
		if sd.LineSpacing > 0 {
			fmt.Fprintf(b, " lineSpacing=\"%.2f\"", sd.LineSpacing)
		}
		if sd.LineRule != "" {
			fmt.Fprintf(b, " lineRule=\"%s\"", sd.LineRule)
		}
		if sd.IndentLeft > 0 {
			fmt.Fprintf(b, " indentLeft=\"%.2f\"", sd.IndentLeft)
		}
		if sd.IndentRight > 0 {
			fmt.Fprintf(b, " indentRight=\"%.2f\"", sd.IndentRight)
		}
		if sd.IndentFirst > 0 {
			fmt.Fprintf(b, " indentFirst=\"%.2f\"", sd.IndentFirst)
		}
		if sd.IndentHanging > 0 {
			fmt.Fprintf(b, " indentHanging=\"%.2f\"", sd.IndentHanging)
		}
		if sd.BorderWidth > 0 {
			fmt.Fprintf(b, " borderWidth=\"%.2f\"", sd.BorderWidth)
		}
		if sd.BorderColor != "" {
			fmt.Fprintf(b, " borderColor=\"%s\"", strings.TrimPrefix(sd.BorderColor, "#"))
		}
		if sd.BorderStyle != "" {
			fmt.Fprintf(b, " borderStyle=\"%s\"", sd.BorderStyle)
		}
		if sd.CellSpacing > 0 {
			fmt.Fprintf(b, " cellSpacing=\"%.2f\"", sd.CellSpacing)
		}
		if sd.Width > 0 {
			fmt.Fprintf(b, " width=\"%.2f\"", sd.Width)
		}
		b.WriteString("/>\n")
	}

	if doc.Theme != nil {
		b.WriteString("    <s:theme")
		if doc.Theme.Font != "" {
			fmt.Fprintf(b, " font=\"%s\"", xmlEscape(doc.Theme.Font))
		}
		if doc.Theme.FontEA != "" {
			fmt.Fprintf(b, " fontEA=\"%s\"", xmlEscape(doc.Theme.FontEA))
		}
		if doc.Theme.FontCS != "" {
			fmt.Fprintf(b, " fontCS=\"%s\"", xmlEscape(doc.Theme.FontCS))
		}
		if doc.Theme.Fg != "" {
			fmt.Fprintf(b, " fg=\"%s\"", doc.Theme.Fg)
		}
		if doc.Theme.Bg != "" {
			fmt.Fprintf(b, " bg=\"%s\"", doc.Theme.Bg)
		}
		b.WriteString("/>\n")
	}

	alignSeen := make(map[string]bool)
	for _, item := range doc.Content {
		if item.Type != "paragraph" || item.Paragraph == nil {
			continue
		}
		p := item.Paragraph
		if p.Align == "" {
			continue
		}
		if p.StyleID != "" {
			if sd, ok := doc.StyleMap[p.StyleID]; ok && sd.Align == p.Align {
				continue
			}
		}
		var el string
		if p.HeadingLevel > 0 {
			el = fmt.Sprintf("h%d", p.HeadingLevel)
		} else {
			el = "p"
		}
		key := fmt.Sprintf("%s_%s", el, p.Align)
		if alignSeen[key] {
			continue
		}
		alignSeen[key] = true
		fmt.Fprintf(b, "    <s:align el=\"%s\" value=\"%s\"/>\n", el, p.Align)
	}

	b.WriteString("  </style>\n")
}

func emitContent(b *strings.Builder, doc *ParsedDocument) {
	i := 0
	for i < len(doc.Content) {
		item := doc.Content[i]
		if item.Type == "list" {
			i = emitListGroup(b, i, doc)
			continue
		}
		emitContentItem(b, item, doc, "    ")
		i++
	}
}

func emitListGroup(b *strings.Builder, start int, doc *ParsedDocument) int {
	first := doc.Content[start].Paragraph
	tag, typeAttr := listTagAndType(first, doc)
	indent := "    "
	fmt.Fprintf(b, "%s<%s type=\"%s\"", indent, tag, typeAttr)
	if tag == "ol" {
		if st := listStart(first, doc); st > 1 {
			fmt.Fprintf(b, " start=\"%d\"", st)
		}
	}
	b.WriteString(">\n")
	end := emitListItems(b, start, first.NumID, first.ListLevel, indent+"  ", doc)
	fmt.Fprintf(b, "%s</%s>\n", indent, tag)
	return end
}

func emitListItems(b *strings.Builder, idx, numID, level int, indent string, doc *ParsedDocument) int {
	startIdx := idx
	startAbstractID := -1
	if doc.NumToAbstract != nil {
		startAbstractID = doc.NumToAbstract[numID]
	}
	for idx < len(doc.Content) {
		item := doc.Content[idx]
		if item.Type != "list" || item.Paragraph.NumID != numID {
			if !hasSameNumIDAhead(doc.Content, idx, numID, startAbstractID) {
				break
			}
			// Emit table items that appear between list items
			if item.Type == "table" {
				writeTableIndent(b, item.Table, doc, strings.TrimRight(indent, " "))
			}
			idx++
			continue
		}
		if doc.NumToAbstract != nil && startAbstractID >= 0 {
			itemAbstractID := doc.NumToAbstract[item.Paragraph.NumID]
			if itemAbstractID != startAbstractID {
				break
			}
		}
		ilvl := item.Paragraph.ListLevel
		if ilvl < level {
			break
		}
		if ilvl == level {
			if doc.NumStartMap != nil {
				if levels, ok := doc.NumStartMap[numID]; ok {
					if _, hasOverride := levels[ilvl]; hasOverride && idx != startIdx {
						break
					}
				}
			}
			content := buildInlineText(item.Paragraph.Runs, doc.DefaultFont, doc.Mode)
			idx++
			conts := collectListContinuations(doc, idx, numID, startAbstractID)
			fmt.Fprintf(b, "%s<li>\n", indent)
			// The first <p> is the geometry owner: resolve marker geometry from
			// the numbering level when the item paragraph has no w:ind of its own.
			writeListFirstParagraph(b, item.Paragraph, content, indent+"  ", doc)
			for _, ci := range conts {
				cc := buildInlineText(doc.Content[ci].Paragraph.Runs, doc.DefaultFont, doc.Mode)
				writeListParagraph(b, doc.Content[ci].Paragraph, cc, indent+"  ")
			}
			idx += len(conts)
			idx = emitListNested(b, idx, numID, level, startAbstractID, indent, doc)
			fmt.Fprintf(b, "%s</li>\n", indent)
			continue
		}
		break
	}
	return idx
}

// collectListContinuations returns the content indices of consecutive paragraphs
// that continue the current list item (no numPr of their own). Per GAP-02, any
// non-list paragraph that is not a section break is treated as continuation
// content when it appears immediately after a list item — including after the
// last item of a list. Such paragraphs are emitted as <p> children of the <li>.
func collectListContinuations(doc *ParsedDocument, from int, numID int, abstractID int) []int {
	var conts []int
	for i := from; i < len(doc.Content); i++ {
		it := doc.Content[i]
		if it.Type != "paragraph" || it.Paragraph == nil {
			break
		}
		if isSectionBreak(it.Paragraph) {
			break
		}
		conts = append(conts, i)
	}
	return conts
}

// emitListNested emits nested sub-lists that belong to the current <li> and
// returns the index past the nested list items.
func emitListNested(b *strings.Builder, idx, numID, level int, abstractID int, indent string, doc *ParsedDocument) int {
	for idx < len(doc.Content) {
		next := doc.Content[idx]
		if next.Type != "list" || next.Paragraph.NumID != numID {
			break
		}
		if abstractID >= 0 && doc.NumToAbstract != nil {
			if doc.NumToAbstract[next.Paragraph.NumID] != abstractID {
				break
			}
		}
		nl := next.Paragraph.ListLevel
		if nl <= level {
			break
		}
		nestedTag, nestedTypeAttr := listTagAndType(next.Paragraph, doc)
		fmt.Fprintf(b, "<%s type=\"%s\"", nestedTag, nestedTypeAttr)
		if nestedTag == "ol" {
			if st := listStart(next.Paragraph, doc); st > 1 {
				fmt.Fprintf(b, " start=\"%d\"", st)
			}
		}
		b.WriteString(">\n")
		idx = emitListItems(b, idx, numID, nl, indent+"    ", doc)
		fmt.Fprintf(b, "</%s>", nestedTag)
	}
	return idx
}

// writeListFirstParagraph writes the geometry-owner <p> of a list item, resolving
// the marker geometry (indentLeft/indentHanging) from the numbering level when
// the item paragraph itself carries no w:ind. Word renders a numbered item with
// a hanging-only w:ind as if left==hanging, so the left indent mirrors the
// hanging value. Continuation paragraphs are NOT passed through here: they emit
// only their own geometry.
func writeListFirstParagraph(b *strings.Builder, p *ParsedParagraph, content string, indent string, doc *ParsedDocument) {
	firstP := *p
	if firstP.IndentLeft <= 0 && firstP.IndentHanging <= 0 {
		if lvlLeft, lvlHanging := resolveLevelInd(p, doc); lvlLeft > 0 || lvlHanging > 0 {
			firstP.IndentLeft = lvlLeft
			firstP.IndentHanging = lvlHanging
		}
	}
	if firstP.IndentLeft <= 0 && firstP.IndentHanging > 0 {
		firstP.IndentLeft = firstP.IndentHanging
	}
	writeListParagraph(b, &firstP, content, indent)
}

// writeListParagraph writes a <p> child inside a <li>.
func writeListParagraph(b *strings.Builder, p *ParsedParagraph, content string, indent string) {
	fmt.Fprintf(b, "%s<p", indent)
	writeParagraphAttrs(b, p)
	b.WriteString(">")
	b.WriteString(content)
	b.WriteString("</p>\n")
}

// resolveLevelInd returns the numbering level's indentation (in inches) for a
// list item, from the level's pPr/w:ind. Returns (0, 0) when absent.
func resolveLevelInd(p *ParsedParagraph, doc *ParsedDocument) (float64, float64) {
	if p == nil || doc == nil || doc.NumLvlIndMap == nil {
		return 0, 0
	}
	key := fmt.Sprintf("%d_%d", p.NumID, p.ListLevel)
	if ind, ok := doc.NumLvlIndMap[key]; ok {
		return ind.Left, ind.Hanging
	}
	return 0, 0
}

func hasSameNumIDAhead(items []ContentItem, from int, numID int, abstractID int) bool {
	for i := from; i < len(items); i++ {
		item := items[i]
		if item.Type == "list" && item.Paragraph.NumID == numID {
			return true
		}
		// Stop looking if we encounter a different list (not a gap paragraph/table)
		if item.Type == "list" && item.Paragraph.NumID != numID {
			return false
		}
		// Stop looking if we encounter a heading — it is a structural break
		if item.Type == "paragraph" && item.Paragraph != nil && item.Paragraph.HeadingLevel > 0 {
			return false
		}
		// Allow up to 50 gap items (paragraphs, tables) between list items
		if i-from > 50 {
			return false
		}
	}
	return false
}

func isSectionBreak(p *ParsedParagraph) bool {
	if p == nil || len(p.Runs) == 0 {
		return false
	}
	text := ""
	for _, r := range p.Runs {
		text += r.Text
	}
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return true
	}
	if strings.HasPrefix(text, "--------") || strings.HasPrefix(text, "------ ") {
		return true
	}
	if p.HeadingLevel > 0 {
		return true
	}
	return false
}

func listTagAndType(p *ParsedParagraph, doc *ParsedDocument) (string, string) {
	numFmt := ""
	if doc.NumFmtMap != nil {
		key := fmt.Sprintf("%d_%d", p.NumID, p.ListLevel)
		if nf, ok := doc.NumFmtMap[key]; ok {
			numFmt = nf
		}
	}
	if numFmt == "bullet" {
		if doc.NumLvlTextMap != nil {
			key := fmt.Sprintf("%d_%d", p.NumID, p.ListLevel)
			if lt, ok := doc.NumLvlTextMap[key]; ok && lt != "" {
				return "ul", lt
			}
		}
		return "ul", "bullet"
	}
	if numFmt == "" {
		return "ul", "bullet"
	}
	return "ol", numFmt
}

func listStart(p *ParsedParagraph, doc *ParsedDocument) int {
	if doc.NumStartMap != nil {
		if levels, ok := doc.NumStartMap[p.NumID]; ok {
			if v, ok := levels[p.ListLevel]; ok && v > 0 {
				return v
			}
		}
	}
	return 1
}

func emitContentItem(b *strings.Builder, item ContentItem, doc *ParsedDocument, indent string) {
	switch item.Type {
	case "paragraph":
		formatParagraphIndent(b, item.Paragraph, doc, indent)
	case "list":
		tag, typeAttr := listTagAndType(item.Paragraph, doc)
		content := buildInlineText(item.Paragraph.Runs, doc.DefaultFont, doc.Mode)
		fmt.Fprintf(b, "%s<%s type=\"%s\">\n", indent, tag, typeAttr)
		fmt.Fprintf(b, "%s  <li>\n", indent)
		writeListFirstParagraph(b, item.Paragraph, content, indent+"    ", doc)
		fmt.Fprintf(b, "%s  </li>\n", indent)
		fmt.Fprintf(b, "%s</%s>\n", indent, tag)
	case "table":
		writeTableIndent(b, item.Table, doc, indent)
	}
}

func formatParagraph(b *strings.Builder, p *ParsedParagraph, doc *ParsedDocument) {
	formatParagraphIndent(b, p, doc, "    ")
}

func formatParagraphIndent(b *strings.Builder, p *ParsedParagraph, doc *ParsedDocument, indent string) {
	if p.IsCode {
		content := buildInlineText(p.Runs, doc.DefaultFont, doc.Mode, true)
		fmt.Fprintf(b, "%s<pre>%s</pre>\n", indent, content)
		return
	}
	if p.HeadingLevel > 0 && p.HeadingLevel <= 9 {
		content := buildInlineText(p.Runs, doc.DefaultFont, doc.Mode)
		fmt.Fprintf(b, "%s<h%d", indent, p.HeadingLevel)
		writeParagraphAttrs(b, p)
		b.WriteString(">")
		b.WriteString(content)
		fmt.Fprintf(b, "</h%d>\n", p.HeadingLevel)
		return
	}
	if p.IsQuote {
		content := buildInlineText(p.Runs, doc.DefaultFont, doc.Mode)
		b.WriteString(indent)
		b.WriteString("<blockquote")
		writeParagraphAttrs(b, p)
		b.WriteString(">")
		b.WriteString(content)
		b.WriteString("</blockquote>\n")
		return
	}
	content := buildInlineText(p.Runs, doc.DefaultFont, doc.Mode)
	b.WriteString(indent)
	b.WriteString("<p")
	writeParagraphAttrs(b, p)
	b.WriteString(">")
	b.WriteString(content)
	b.WriteString("</p>\n")
}

func customStyleName(p *ParsedParagraph) string {
	if p.StyleName == "" {
		return ""
	}
	sn := strings.ToLower(strings.TrimSpace(p.StyleName))
	if p.HeadingLevel > 0 {
		expected := fmt.Sprintf("heading %d", p.HeadingLevel)
		if sn == expected || sn == "title" || sn == "subtitle" {
			return ""
		}
		return p.StyleName
	}
	if p.IsQuote {
		if sn == "quote" || sn == "intense quote" || sn == "blocktext" || sn == "block text" {
			return ""
		}
		return p.StyleName
	}
	if sn == "normal" || sn == "default paragraph font" || sn == "list paragraph" {
		return ""
	}
	return p.StyleName
}

func writeParagraphAttrs(b *strings.Builder, p *ParsedParagraph) {
	if c := customStyleName(p); c != "" {
		fmt.Fprintf(b, " c=\"%s\"", xmlEscape(c))
	}
	if p.Align != "" && p.Align != "left" {
		fmt.Fprintf(b, " align=\"%s\"", p.Align)
	}
	if p.IndentLeft > 0 {
		fmt.Fprintf(b, " indentLeft=\"%.2f\"", p.IndentLeft)
	}
	if p.IndentHanging > 0 {
		fmt.Fprintf(b, " indentHanging=\"%.2f\"", p.IndentHanging)
	}
	if p.IndentRight > 0 {
		fmt.Fprintf(b, " indentRight=\"%.2f\"", p.IndentRight)
	}
	if p.IndentFirst > 0 {
		fmt.Fprintf(b, " indentFirst=\"%.2f\"", p.IndentFirst)
	}
	if len(p.Tabs) > 0 {
		b.WriteString(" tabs=\"")
		for i, t := range p.Tabs {
			if i > 0 {
				b.WriteByte(' ')
			}
			fmt.Fprintf(b, "%.2f", t.Pos)
			// A stop is compact when it uses the defaults (left/none). Anything
			// else — including "clear", which is NOT a landing position — must be
			// carried explicitly so consumers never mistake it for a plain stop.
			if t.Align != "left" || t.Leader != "none" {
				fmt.Fprintf(b, "@%s", t.Align)
				if t.Leader != "none" {
					fmt.Fprintf(b, ":%s", t.Leader)
				}
			}
		}
		b.WriteByte('"')
	}
	if p.SpacingBefore > 0 {
		fmt.Fprintf(b, " spacingBefore=\"%.2f\"", p.SpacingBefore)
	}
	if p.SpacingAfter > 0 {
		fmt.Fprintf(b, " spacingAfter=\"%.2f\"", p.SpacingAfter)
	}
	if p.LineSpacing > 0 {
		fmt.Fprintf(b, " lineSpacing=\"%.2f\"", p.LineSpacing)
	}
	if p.LineRule != "" && p.LineRule != "auto" {
		fmt.Fprintf(b, " lineRule=\"%s\"", p.LineRule)
	}
	if p.Bidi {
		b.WriteString(" dir=\"rtl\"")
	}
	if p.Lang != "" {
		fmt.Fprintf(b, " lang=\"%s\"", p.Lang)
	}
	if p.VAlign != "" {
		fmt.Fprintf(b, " valign=\"%s\"", p.VAlign)
	}
	if p.Shading != "" {
		fmt.Fprintf(b, " shd=\"%s\"", p.Shading)
	}
	if p.KeepNext {
		b.WriteString(" keepNext=\"true\"")
	}
	if p.KeepLines {
		b.WriteString(" keepLines=\"true\"")
	}
	if p.WidowControl {
		b.WriteString(" widowControl=\"true\"")
	}
	if p.SectionBreak != "" {
		fmt.Fprintf(b, " sectionBreak=\"%s\"", p.SectionBreak)
	}
	if p.RevisionAuthor != "" {
		fmt.Fprintf(b, " revisionAuthor=\"%s\"", xmlEscape(p.RevisionAuthor))
	}
	if p.RevisionDate != "" {
		fmt.Fprintf(b, " revisionDate=\"%s\"", p.RevisionDate)
	}
	if p.SuppressAutoHyph {
		b.WriteString(" suppressAutoHyph=\"true\"")
	}
	if p.SnapToGrid {
		b.WriteString(" snapToGrid=\"true\"")
	}
	if p.Kinsoku {
		b.WriteString(" kinsoku=\"true\"")
	}
	if p.WordWrap {
		b.WriteString(" wordWrap=\"true\"")
	}
	if p.OverflowPunct {
		b.WriteString(" overflowPunct=\"true\"")
	}
	if p.TopLinePunct {
		b.WriteString(" topLinePunct=\"true\"")
	}
	if p.AutoSpaceDE {
		b.WriteString(" autoSpaceDE=\"true\"")
	}
	if p.AutoSpaceDN {
		b.WriteString(" autoSpaceDN=\"true\"")
	}
	if p.TextDirection != "" {
		fmt.Fprintf(b, " textDirection=\"%s\"", p.TextDirection)
	}
	if p.SuppressOverlap {
		b.WriteString(" suppressOverlap=\"true\"")
	}
	if p.DivID > 0 {
		fmt.Fprintf(b, " divID=\"%d\"", p.DivID)
	}
	if p.CnfStyle != "" {
		fmt.Fprintf(b, " cnfStyle=\"%s\"", xmlEscape(p.CnfStyle))
	}
	if p.FramePr != nil {
		fp := p.FramePr
		frameAttrs := ""
		if fp.DropCap != "" {
			frameAttrs += fmt.Sprintf(" dropCap='%s'", fp.DropCap)
		}
		if fp.Lines > 0 {
			frameAttrs += fmt.Sprintf(" lines='%d'", fp.Lines)
		}
		if fp.Width > 0 {
			frameAttrs += fmt.Sprintf(" width='%.2f'", float64(fp.Width)/twipsPerInch)
		}
		if fp.Height > 0 {
			frameAttrs += fmt.Sprintf(" height='%.2f'", float64(fp.Height)/twipsPerInch)
		}
		if fp.VSpace > 0 {
			frameAttrs += fmt.Sprintf(" vSpace='%.2f'", float64(fp.VSpace)/twipsPerInch)
		}
		if fp.HSpace > 0 {
			frameAttrs += fmt.Sprintf(" hSpace='%.2f'", float64(fp.HSpace)/twipsPerInch)
		}
		if fp.Wrap != "" {
			frameAttrs += fmt.Sprintf(" wrap='%s'", fp.Wrap)
		}
		if fp.HAnchor != "" {
			frameAttrs += fmt.Sprintf(" hAnchor='%s'", fp.HAnchor)
		}
		if fp.VAnchor != "" {
			frameAttrs += fmt.Sprintf(" vAnchor='%s'", fp.VAnchor)
		}
		if fp.X > 0 {
			frameAttrs += fmt.Sprintf(" x='%.2f'", float64(fp.X)/twipsPerInch)
		}
		if fp.XAlign != "" {
			frameAttrs += fmt.Sprintf(" xAlign='%s'", fp.XAlign)
		}
		if fp.Y > 0 {
			frameAttrs += fmt.Sprintf(" y='%.2f'", float64(fp.Y)/twipsPerInch)
		}
		if fp.YAlign != "" {
			frameAttrs += fmt.Sprintf(" yAlign='%s'", fp.YAlign)
		}
		if fp.HRule != "" {
			frameAttrs += fmt.Sprintf(" hRule='%s'", fp.HRule)
		}
		if fp.AnchorLock {
			frameAttrs += " anchorLock='true'"
		}
		if frameAttrs != "" {
			fmt.Fprintf(b, " frame=\"%s\"", frameAttrs[1:])
		}
	}
	at := buildBorderAttr(p.BorderTop, p.BorderBottom, p.BorderLeft, p.BorderRight)
	if at != "" {
		fmt.Fprintf(b, " at=\"%s\"", at)
	}
}

func buildInlineText(runs []TextRun, defaultFont StyleDef, mode string, isCode ...bool) string {
	codeMode := len(isCode) > 0 && isCode[0]
	var b strings.Builder
	for _, r := range runs {
		if r.IsDeleted && mode != "lossless" {
			continue
		}
		var core string
		switch {
		case r.IsTab:
			if codeMode {
				core = "\t"
			} else {
				core = "<tab/>"
			}
		case r.IsLineBreak:
			if codeMode {
				core = "\n"
			} else {
				core = fmt.Sprintf("<br type=\"%s\"/>", r.BreakType)
			}
		case r.IsFootnoteRef:
			noteType := r.NoteType
			if noteType == "" {
				noteType = "footnote"
			}
			core = fmt.Sprintf("<fn-ref id=\"%d\" type=\"%s\"/>", r.NoteID, noteType)
		case r.IsHyperlink:
			content := xmlEscape(r.Text)
			if r.Bold {
				content = "<b>" + content + "</b>"
			}
			if r.Italic {
				content = "<i>" + content + "</i>"
			}
			if r.Underline != "" {
				content = "<u>" + content + "</u>"
			}
			if r.Strike {
				content = "<s>" + content + "</s>"
			}
			core = fmt.Sprintf("<a href=\"%s\">%s</a>", xmlEscape(r.HyperlinkURL), content)
		case r.IsImage:
			if r.ImageAlt != "" {
				core = fmt.Sprintf("<img alt=\"%s\"/>", xmlEscape(r.ImageAlt))
			} else {
				core = "<img/>"
			}
		case r.IsSym:
			if r.SymChar != "" {
				char, err := strconv.ParseInt(r.SymChar, 16, 32)
				if err == nil {
					core = string(rune(char))
				}
			}
		case r.Text != "":
			core = xmlEscape(r.Text)
			if !codeMode {
				spanAttrs := ""
				if r.FontFamily != "" && r.FontFamily != defaultFont.Family {
					spanAttrs += fmt.Sprintf(" font=\"%s\"", r.FontFamily)
				}
				if r.FontEA != "" && r.FontEA != defaultFont.FontEA {
					spanAttrs += fmt.Sprintf(" fontEA=\"%s\"", r.FontEA)
				}
				if r.FontCS != "" && r.FontCS != defaultFont.FontCS {
					spanAttrs += fmt.Sprintf(" fontCS=\"%s\"", r.FontCS)
				}
				if r.FontSizePt > 0 && r.FontSizePt != defaultFont.SizePt {
					spanAttrs += fmt.Sprintf(" size=\"%spt\"", formatPt(r.FontSizePt))
				}
				if r.FontSizeCS > 0 && r.FontSizeCS != defaultFont.SizeCS {
					spanAttrs += fmt.Sprintf(" sizeCS=\"%spt\"", formatPt(r.FontSizeCS))
				}
				if r.FontColor != "" && r.FontColor != defaultFont.Color {
					spanAttrs += fmt.Sprintf(" color=\"%s\"", strings.TrimPrefix(r.FontColor, "#"))
				}
				if r.Highlight != "" && r.Highlight != "none" {
					spanAttrs += fmt.Sprintf(" highlight=\"%s\"", r.Highlight)
				}
				if r.Lang != "" {
					spanAttrs += fmt.Sprintf(" lang=\"%s\"", r.Lang)
				}
				if r.Hidden {
					spanAttrs += " hidden=\"true\""
				}
				if r.IsRTL {
					spanAttrs += " dir=\"rtl\""
				}
				if spanAttrs != "" {
					core = "<span" + spanAttrs + ">" + core + "</span>"
				}
				if r.SmallCaps {
					core = "<smallcaps>" + core + "</smallcaps>"
				}
				if r.AllCaps {
					core = "<uppercase>" + core + "</uppercase>"
				}
				if r.Bold {
					core = "<b>" + core + "</b>"
				}
				// Only emit <bcs> when:
				// 1. w:b is not already set (avoid double-wrapping), AND
				// 2. the CS font differs from the main font (real complex script font in use).
				// When CS font == main font (e.g. both "Courier New"), w:bCs is a
				// Latin-font artifact set by Word with no visual effect on Latin text.
				if r.IsBoldCS && !r.Bold && r.FontCS != "" && r.FontCS != r.FontFamily {
					core = "<bcs>" + core + "</bcs>"
				}
				if r.Italic {
					core = "<i>" + core + "</i>"
				}
				// Same rule as <bcs>: only emit <ics> when CS font differs from main font.
				if r.IsItalicCS && !r.Italic && r.FontCS != "" && r.FontCS != r.FontFamily {
					core = "<ics>" + core + "</ics>"
				}
				if r.Underline != "" {
					core = "<u>" + core + "</u>"
				}
				if r.Strike {
					core = "<s>" + core + "</s>"
				}
				if r.SuperScript {
					core = "<sup>" + core + "</sup>"
				}
				if r.SubScript {
					core = "<sub>" + core + "</sub>"
				}
			}
		default:
			continue
		}
		if r.IsInserted && mode == "lossless" {
			tag := "<ins"
			if r.InsAuthor != "" {
				tag += fmt.Sprintf(" author=\"%s\"", xmlEscape(r.InsAuthor))
			}
			if r.InsDate != "" {
				tag += fmt.Sprintf(" date=\"%s\"", xmlEscape(r.InsDate))
			}
			tag += ">"
			core = tag + core + "</ins>"
		}
		if r.IsDeleted && mode == "lossless" {
			tag := "<del"
			if r.DelAuthor != "" {
				tag += fmt.Sprintf(" author=\"%s\"", xmlEscape(r.DelAuthor))
			}
			if r.DelDate != "" {
				tag += fmt.Sprintf(" date=\"%s\"", xmlEscape(r.DelDate))
			}
			tag += ">"
			core = tag + core + "</del>"
		}
		b.WriteString(core)
	}
	return b.String()
}

func writeTableIndent(b *strings.Builder, t *ParsedTable, doc *ParsedDocument, indent string) {
	fmt.Fprintf(b, "%s<table id=\"%d\"", indent, t.ID)
	if t.StyleName != "" {
		fmt.Fprintf(b, " c=\"%s\"", xmlEscape(t.StyleName))
	}
	if t.Width > 0 {
		fmt.Fprintf(b, " width=\"%.2f\"", t.Width)
	}
	if t.Alignment != "" {
		fmt.Fprintf(b, " align=\"%s\"", t.Alignment)
	}
	if t.Indent > 0 {
		fmt.Fprintf(b, " indent=\"%.2f\"", t.Indent)
	}
	if t.CellSpace > 0 {
		fmt.Fprintf(b, " cellSpacing=\"%.2f\"", t.CellSpace)
	}
	if t.Caption != "" {
		fmt.Fprintf(b, " caption=\"%s\"", xmlEscape(t.Caption))
	}
	if t.Summary != "" {
		fmt.Fprintf(b, " summary=\"%s\"", xmlEscape(t.Summary))
	}
	at := buildBorderAttr(t.BorderTop, t.BorderBottom, t.BorderLeft, t.BorderRight)
	if at != "" {
		fmt.Fprintf(b, " at=\"%s\"", at)
	}
	b.WriteString(">\n")

	for _, row := range t.Rows {
		tag := "td"
		if row.IsHeader {
			tag = "th"
		}
		fmt.Fprintf(b, "%s  <tr>\n", indent)
		for _, cell := range row.Cells {
			if cell.Omitted {
				continue
			}
			fmt.Fprintf(b, "%s    <%s", indent, tag)
			if cell.GridSpan > 1 {
				fmt.Fprintf(b, " colspan=\"%d\"", cell.GridSpan)
			}
			if cell.RowSpan > 1 {
				fmt.Fprintf(b, " rowspan=\"%d\"", cell.RowSpan)
			}
			if cell.VAlign != "" {
				fmt.Fprintf(b, " valign=\"%s\"", cell.VAlign)
			}
			if cell.TextDir != "" {
				fmt.Fprintf(b, " textDir=\"%s\"", cell.TextDir)
			}
			if cell.Lang != "" {
				fmt.Fprintf(b, " lang=\"%s\"", cell.Lang)
			}
			if cell.NoWrap {
				b.WriteString(" noWrap=\"true\"")
			}
			cellAt := buildBorderAttr(cell.BorderTop, cell.BorderBottom, cell.BorderLeft, cell.BorderRight)
			if cellAt != "" {
				fmt.Fprintf(b, " at=\"%s\"", cellAt)
			}
			b.WriteString(">")
			for _, ci := range cell.Content {
				switch ci.Type {
				case "paragraph":
					b.WriteString(buildInlineText(ci.Paragraph.Runs, doc.DefaultFont, doc.Mode))
				case "list":
					tag, typeAttr := listTagAndType(ci.Paragraph, doc)
					content := buildInlineText(ci.Paragraph.Runs, doc.DefaultFont, doc.Mode)
					firstP := *ci.Paragraph
					if firstP.IndentLeft <= 0 && firstP.IndentHanging <= 0 {
						if lvlLeft, lvlHanging := resolveLevelInd(ci.Paragraph, doc); lvlLeft > 0 || lvlHanging > 0 {
							firstP.IndentLeft = lvlLeft
							firstP.IndentHanging = lvlHanging
						}
					}
					if firstP.IndentLeft <= 0 && firstP.IndentHanging > 0 {
						firstP.IndentLeft = firstP.IndentHanging
					}
					fmt.Fprintf(b, "<%s type=\"%s\"><li><p", tag, typeAttr)
					writeParagraphAttrs(b, &firstP)
					fmt.Fprintf(b, ">%s</p></li></%s>", content, tag)
				case "table":
					writeTableIndent(b, ci.Table, doc, indent+"    ")
				}
			}
			fmt.Fprintf(b, "</%s>\n", tag)
		}
		fmt.Fprintf(b, "%s  </tr>\n", indent)
	}
	fmt.Fprintf(b, "%s</table>\n", indent)
}
