package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type zipProcessor struct {
	docxData      []byte
	contentTypes  *contentTypesXML
	relationships *relationshipsXML
	newParts      map[string][]byte
}

type contentTypesXML struct {
	XMLName  xml.Name              `xml:"Types"`
	Xmlns    string                `xml:"xmlns,attr"`
	Defaults []contentTypeDefault  `xml:"Default"`
	Overrides []contentTypeOverride `xml:"Override"`
}

type contentTypeDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

type contentTypeOverride struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

type relationshipsXML struct {
	XMLName       xml.Name           `xml:"Relationships"`
	Xmlns         string             `xml:"xmlns,attr"`
	Relationships []relationshipXML  `xml:"Relationship"`
}

type relationshipXML struct {
	ID      string `xml:"Id,attr"`
	Type    string `xml:"Type,attr"`
	Target  string `xml:"Target,attr"`
}

func newZipProcessor(docxData []byte) *zipProcessor {
	return &zipProcessor{
		docxData: docxData,
		newParts: make(map[string][]byte),
	}
}

func (z *zipProcessor) addPart(path string, data []byte) {
	z.newParts[path] = data
}

func (z *zipProcessor) process() ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(z.docxData), int64(len(z.docxData)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	files := make(map[string]*zipFileEntry)
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open file %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", f.Name, err)
		}
		files[f.Name] = &zipFileEntry{
			data:   data,
			method: f.Method,
		}
	}

	for path, data := range z.newParts {
		files[path] = &zipFileEntry{
			data:   data,
			method: zip.Deflate,
		}
	}

	if err := z.updateContentTypes(files); err != nil {
		return nil, err
	}
	if err := z.updateRelationships(files); err != nil {
		return nil, err
	}

	return z.repackage(files)
}

type zipFileEntry struct {
	data   []byte
	method uint16
}

func (z *zipProcessor) updateContentTypes(files map[string]*zipFileEntry) error {
	entry, ok := files["[Content_Types].xml"]
	if !ok {
		return nil
	}

	var ct contentTypesXML
	if err := xml.Unmarshal(entry.data, &ct); err != nil {
		return fmt.Errorf("parse content types: %w", err)
	}

	neededParts := map[string]string{
		"word/footnotes.xml": "application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml",
		"word/comments.xml":  "application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml",
		"word/numbering.xml": "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml",
		"word/styles.xml":   "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml",
	}

	for partPath, contentType := range neededParts {
		if _, hasFile := files[partPath]; !hasFile {
			continue
		}
		exists := false
		for _, o := range ct.Overrides {
			if o.PartName == "/"+partPath {
				exists = true
				break
			}
		}
		if !exists {
			ct.Overrides = append(ct.Overrides, contentTypeOverride{
				PartName:    "/" + partPath,
				ContentType: contentType,
			})
		}
	}

	data, err := xml.MarshalIndent(&ct, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal content types: %w", err)
	}
	entry.data = append([]byte(xml.Header), data...)
	return nil
}

func (z *zipProcessor) updateRelationships(files map[string]*zipFileEntry) error {
	entry, ok := files["word/_rels/document.xml.rels"]
	if !ok {
		return nil
	}

	var rels relationshipsXML
	if err := xml.Unmarshal(entry.data, &rels); err != nil {
		return fmt.Errorf("parse relationships: %w", err)
	}

	neededRels := []struct {
		target string
		typ    string
	}{
		{"footnotes.xml", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes"},
		{"comments.xml", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"},
		{"numbering.xml", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"},
		{"styles.xml", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"},
	}

	existingTargets := make(map[string]bool)
	for _, r := range rels.Relationships {
		existingTargets[r.Target] = true
	}

	relID := len(rels.Relationships) + 1
	for _, nr := range neededRels {
		if _, hasFile := files["word/"+nr.target]; !hasFile {
			continue
		}
		if existingTargets[nr.target] {
			continue
		}
		rels.Relationships = append(rels.Relationships, relationshipXML{
			ID:     fmt.Sprintf("rId%d", relID),
			Type:   nr.typ,
			Target: nr.target,
		})
		relID++
	}

	data, err := xml.MarshalIndent(&rels, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal relationships: %w", err)
	}
	entry.data = append([]byte(xml.Header), data...)
	return nil
}

func (z *zipProcessor) repackage(files map[string]*zipFileEntry) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for path, entry := range files {
		f, err := writer.Create(path)
		if err != nil {
			return nil, fmt.Errorf("create zip entry %s: %w", path, err)
		}
		if _, err := f.Write(entry.data); err != nil {
			return nil, fmt.Errorf("write zip entry %s: %w", path, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func buildFootnotesXML(notes []NoteItemXML) []byte {
	if len(notes) == 0 {
		return nil
	}

	xf := &FootnotesXML{
		Xmlns:  "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR: "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
	}

	for _, note := range notes {
		fn := &FootnoteXML{
			ID:   fmt.Sprintf("%d", note.ID),
			Type: note.Type,
		}

		for _, body := range note.Body {
			para := &FootnoteParaXML{
				Properties: &FootnoteParaPropsXML{
					Style: &FootnoteStyleXML{Val: "FootnoteText"},
				},
			}

			for _, inline := range body.Inlines {
				text := inline.Text
				if text != "" {
					run := &FootnoteRunXML{
						Properties: &FootnoteRPropsXML{
							RStyle: &FootnoteStyleXML{Val: "FootnoteReference"},
							Font:   &FootnoteFontXML{Ascii: "Calibri", HAnsi: "Calibri"},
							Size:   &FootnoteSizeXML{Val: "18"},
						},
						Text: &FootnoteTextXML{Text: text, Space: "preserve"},
					}
					para.Run = append(para.Run, run)
				}
			}

			if len(para.Run) > 0 {
				fn.Paragraphs = append(fn.Paragraphs, para)
			}
		}

		xf.Items = append(xf.Items, fn)
	}

	data, _ := xml.MarshalIndent(xf, "", "  ")
	return append([]byte(xml.Header), data...)
}

type NoteItemXML struct {
	ID     int
	Type   string
	Author string
	Date   string
	Body   []noteBodyXML
}

type noteBodyXML struct {
	Inlines []noteInlineXML
}

type noteInlineXML struct {
	Text string
}

func buildCommentsXML(notes []NoteItemXML) []byte {
	if len(notes) == 0 {
		return nil
	}

	xf := &CommentsXML{
		Xmlns:  "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsR: "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
	}

	for _, note := range notes {
		if note.Type != "comment" {
			continue
		}
		comment := &CommentXML{
			ID:     fmt.Sprintf("%d", note.ID),
			Author: note.Author,
			Date:   note.Date,
		}

		para := &CommentParaXML{}
		for _, body := range note.Body {
			for _, inline := range body.Inlines {
				text := inline.Text
				if text != "" {
					run := &CommentRunXML{
						Text: &CommentTextXML{Text: text, Space: "preserve"},
					}
					para.Run = append(para.Run, run)
				}
			}
		}
		if len(para.Run) > 0 {
			comment.Paragraphs = append(comment.Paragraphs, para)
		}

		xf.Items = append(xf.Items, comment)
	}

	data, _ := xml.MarshalIndent(xf, "", "  ")
	return append([]byte(xml.Header), data...)
}

func buildNumberingXMLFromWords(numDefs []numDef) []byte {
	xf := buildNumberingXML(numDefs)
	if xf == nil {
		return nil
	}
	data, _ := xml.MarshalIndent(xf, "", "  ")
	return append([]byte(xml.Header), data...)
}

func stripXMLHeader(data []byte) []byte {
	s := string(data)
	if idx := strings.Index(s, "?>"); idx != -1 {
		s = s[idx+2:]
	}
	return []byte(strings.TrimSpace(s))
}

type StylesXML struct {
	XMLName xml.Name      `xml:"w:styles"`
	Xmlns   string        `xml:"xmlns:w,attr"`
	DocDefaults *DocDefaultsXML `xml:"w:docDefaults,omitempty"`
	Styles   []*StyleDefXML `xml:"w:style"`
}

type DocDefaultsXML struct {
	RunPropsDefault *RunPropsDefaultXML `xml:"w:rPrDefault"`
}

type RunPropsDefaultXML struct {
	XMLName xml.Name         `xml:"w:rPrDefault"`
	Props   *DefaultRunPropsXML `xml:"w:rPr"`
}

type DefaultRunPropsXML struct {
	XMLName xml.Name          `xml:"w:rPr"`
	RFonts  *DefaultFontXML   `xml:"w:rFonts,omitempty"`
	Size    *DefaultSizeXML   `xml:"w:sz,omitempty"`
	SizeCS  *DefaultSizeXML   `xml:"w:szCs,omitempty"`
	Color   *StyleColorXML    `xml:"w:color,omitempty"`
	Shading *StyleShadingXML  `xml:"w:shd,omitempty"`
}

type StyleShadingXML struct {
	Val   string `xml:"w:val,attr"`
	Color string `xml:"w:color,attr"`
	Fill  string `xml:"w:fill,attr"`
}

type DefaultFontXML struct {
	Ascii    string `xml:"w:ascii,attr,omitempty"`
	HAnsi    string `xml:"w:hAnsi,attr,omitempty"`
	CS       string `xml:"w:cs,attr,omitempty"`
	EastAsia string `xml:"w:eastAsia,attr,omitempty"`
}

type DefaultSizeXML struct {
	Val string `xml:"w:val,attr"`
}

type StyleDefXML struct {
	XMLName     xml.Name     `xml:"w:style"`
	Type        string       `xml:"w:type,attr"`
	StyleID     string       `xml:"w:styleId,attr"`
	BasedOn     *StyleValXML `xml:"w:basedOn,omitempty"`
	Name        *StyleNameXML `xml:"w:name,omitempty"`
	RunProps     *StyleRunPropsXML `xml:"w:rPr,omitempty"`
	ParaProps    *StyleParaPropsXML `xml:"w:pPr,omitempty"`
}

type StyleValXML struct {
	Val string `xml:"w:val,attr"`
}

type StyleNameXML struct {
	Val string `xml:"w:val,attr"`
}

type StyleRunPropsXML struct {
	XMLName   xml.Name      `xml:"w:rPr"`
	RFonts    *StyleFontXML `xml:"w:rFonts,omitempty"`
	Bold      *StyleBoolXML `xml:"w:b,omitempty"`
	BoldCS    *StyleBoolXML `xml:"w:bCs,omitempty"`
	Italic    *StyleBoolXML `xml:"w:i,omitempty"`
	ItalicCS  *StyleBoolXML `xml:"w:iCs,omitempty"`
	Underline *StyleULXML   `xml:"w:u,omitempty"`
	Strike    *StyleBoolXML `xml:"w:strike,omitempty"`
	SmallCaps *StyleBoolXML `xml:"w:smallCaps,omitempty"`
	Caps      *StyleBoolXML `xml:"w:caps,omitempty"`
	Color     *StyleColorXML `xml:"w:color,omitempty"`
	Size      *StyleSizeXML `xml:"w:sz,omitempty"`
	SizeCS    *StyleSizeXML `xml:"w:szCs,omitempty"`
}

type StyleParaPropsXML struct {
	XMLName     xml.Name             `xml:"w:pPr"`
	Alignment   *StyleAlignXML       `xml:"w:jc,omitempty"`
	Spacing     *StyleSpacingXML     `xml:"w:spacing,omitempty"`
	Ind         *StyleIndentXML      `xml:"w:ind,omitempty"`
	Borders     *StyleBordersXML     `xml:"w:pBdr,omitempty"`
	CellSpacing *StyleCellSpacingXML `xml:"w:tblCellSpacing,omitempty"`
	Width       *StyleWidthXML       `xml:"w:tblW,omitempty"`
}

type StyleBordersXML struct {
	Top    *StyleBorderXML `xml:"w:top,omitempty"`
	Left   *StyleBorderXML `xml:"w:left,omitempty"`
	Bottom *StyleBorderXML `xml:"w:bottom,omitempty"`
	Right  *StyleBorderXML `xml:"w:right,omitempty"`
}

type StyleBorderXML struct {
	Val   string `xml:"w:val,attr"`
	Sz    string `xml:"w:sz,attr"`
	Color string `xml:"w:color,attr"`
	Space string `xml:"w:space,attr"`
}

type StyleCellSpacingXML struct {
	W    string `xml:"w:w,attr"`
	Type string `xml:"w:type,attr"`
}

type StyleWidthXML struct {
	W    string `xml:"w:w,attr"`
	Type string `xml:"w:type,attr"`
}

type StyleAlignXML struct {
	Val string `xml:"w:val,attr"`
}

type StyleSpacingXML struct {
	Before string `xml:"w:before,attr,omitempty"`
	After  string `xml:"w:after,attr,omitempty"`
	Line   string `xml:"w:line,attr,omitempty"`
	LineRule string `xml:"w:lineRule,attr,omitempty"`
}

type StyleIndentXML struct {
	Left     string `xml:"w:left,attr,omitempty"`
	Right    string `xml:"w:right,attr,omitempty"`
	FirstLine string `xml:"w:firstLine,attr,omitempty"`
	Hanging  string `xml:"w:hanging,attr,omitempty"`
}

type StyleFontXML struct {
	Ascii    string `xml:"w:ascii,attr,omitempty"`
	HAnsi    string `xml:"w:hAnsi,attr,omitempty"`
	CS       string `xml:"w:cs,attr,omitempty"`
	EastAsia string `xml:"w:eastAsia,attr,omitempty"`
}

type StyleBoolXML struct{}

type StyleULXML struct {
	Val string `xml:"w:val,attr"`
}

type StyleColorXML struct {
	Val string `xml:"w:val,attr"`
}

type StyleSizeXML struct {
	Val string `xml:"w:val,attr"`
}

func buildStylesXML(theme *StyleTheme, customs []StyleCustom) []byte {
	if theme == nil && len(customs) == 0 {
		return nil
	}

	xf := &StylesXML{
		Xmlns: "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
	}

	if theme != nil && theme.Font != "" {
		xf.DocDefaults = &DocDefaultsXML{
			RunPropsDefault: &RunPropsDefaultXML{
				Props: &DefaultRunPropsXML{
					RFonts: &DefaultFontXML{
						Ascii:    theme.Font,
						HAnsi:    theme.Font,
						CS:       theme.FontCS,
						EastAsia: theme.FontEA,
					},
				},
			},
		}
	}

	if theme != nil && (theme.Fg != "" || theme.Bg != "") {
		if xf.DocDefaults == nil {
			xf.DocDefaults = &DocDefaultsXML{
				RunPropsDefault: &RunPropsDefaultXML{
					Props: &DefaultRunPropsXML{},
				},
			}
		}
		if theme.Fg != "" {
			xf.DocDefaults.RunPropsDefault.Props.Color = &StyleColorXML{Val: strings.TrimPrefix(theme.Fg, "#")}
		}
		if theme.Bg != "" {
			xf.DocDefaults.RunPropsDefault.Props.Shading = &StyleShadingXML{
				Val:    "clear",
				Color:  "auto",
				Fill:   strings.TrimPrefix(theme.Bg, "#"),
			}
		}
	}

	for _, custom := range customs {
		styleID := custom.Name
		if styleID == "" {
			continue
		}

		sd := &StyleDefXML{
			Type:    custom.Type,
			StyleID: styleID,
			Name:    &StyleNameXML{Val: custom.Name},
		}

		if custom.BasedOn != "" {
			sd.BasedOn = &StyleValXML{Val: custom.BasedOn}
		}

		rPr := &StyleRunPropsXML{}
		hasRunProps := false

		if custom.Font != "" || custom.FontEA != "" || custom.FontCS != "" {
			rPr.RFonts = &StyleFontXML{
				Ascii:    custom.Font,
				HAnsi:    custom.Font,
				CS:       custom.FontCS,
				EastAsia: custom.FontEA,
			}
			hasRunProps = true
		}
		if strings.ToLower(custom.Bold) == "true" || custom.Bold == "1" {
			rPr.Bold = &StyleBoolXML{}
			rPr.BoldCS = &StyleBoolXML{}
			hasRunProps = true
		}
		if strings.ToLower(custom.Italic) == "true" || custom.Italic == "1" {
			rPr.Italic = &StyleBoolXML{}
			rPr.ItalicCS = &StyleBoolXML{}
			hasRunProps = true
		}
		if custom.Underline != "" {
			rPr.Underline = &StyleULXML{Val: custom.Underline}
			hasRunProps = true
		}
		if strings.ToLower(custom.Strikethrough) == "true" || custom.Strikethrough == "1" {
			rPr.Strike = &StyleBoolXML{}
			hasRunProps = true
		}
		if strings.ToLower(custom.SmallCaps) == "true" || custom.SmallCaps == "1" {
			rPr.SmallCaps = &StyleBoolXML{}
			hasRunProps = true
		}
		if strings.ToLower(custom.Uppercase) == "true" || custom.Uppercase == "1" {
			rPr.Caps = &StyleBoolXML{}
			hasRunProps = true
		}
		if custom.Color != "" {
			color := strings.TrimPrefix(custom.Color, "#")
			rPr.Color = &StyleColorXML{Val: color}
			hasRunProps = true
		}
		if custom.Size > 0 {
			rPr.Size = &StyleSizeXML{Val: fmt.Sprintf("%d", int(custom.Size*2))}
			hasRunProps = true
		}
		if custom.SizeCS > 0 {
			rPr.SizeCS = &StyleSizeXML{Val: fmt.Sprintf("%d", int(custom.SizeCS*2))}
			hasRunProps = true
		}

		if hasRunProps {
			sd.RunProps = rPr
		}

		pPr := &StyleParaPropsXML{}
		hasParaProps := false

		if custom.Alignment != "" {
			pPr.Alignment = &StyleAlignXML{Val: custom.Alignment}
			hasParaProps = true
		}
		if custom.SpacingBefore > 0 || custom.SpacingAfter > 0 {
			sp := &StyleSpacingXML{}
			if custom.SpacingBefore > 0 {
				sp.Before = fmt.Sprintf("%d", int(custom.SpacingBefore))
			}
			if custom.SpacingAfter > 0 {
				sp.After = fmt.Sprintf("%d", int(custom.SpacingAfter))
			}
			if custom.LineSpacing > 0 {
				sp.Line = fmt.Sprintf("%d", int(custom.LineSpacing))
			}
			if custom.LineRule != "" {
				sp.LineRule = custom.LineRule
			}
			pPr.Spacing = sp
			hasParaProps = true
		}
		if custom.IndentLeft > 0 || custom.IndentRight > 0 || custom.IndentFirst > 0 || custom.IndentHanging > 0 {
			ind := &StyleIndentXML{}
			if custom.IndentLeft > 0 {
				ind.Left = fmt.Sprintf("%d", int(custom.IndentLeft))
			}
			if custom.IndentRight > 0 {
				ind.Right = fmt.Sprintf("%d", int(custom.IndentRight))
			}
			if custom.IndentFirst > 0 {
				ind.FirstLine = fmt.Sprintf("%d", int(custom.IndentFirst))
			}
			if custom.IndentHanging > 0 {
				ind.Hanging = fmt.Sprintf("%d", int(custom.IndentHanging))
			}
			pPr.Ind = ind
			hasParaProps = true
		}
		if custom.BorderWidth > 0 || custom.BorderColor != "" || custom.BorderStyle != "" {
			pPr.Borders = &StyleBordersXML{}
			if custom.BorderWidth > 0 {
				pPr.Borders.Top = &StyleBorderXML{
					Val:    mapBorderStyle(custom.BorderStyle),
					Sz:     fmt.Sprintf("%d", int(custom.BorderWidth)),
					Color:  strings.TrimPrefix(custom.BorderColor, "#"),
					Space:  "4",
				}
				pPr.Borders.Bottom = pPr.Borders.Top
				pPr.Borders.Left = pPr.Borders.Top
				pPr.Borders.Right = pPr.Borders.Top
			}
			hasParaProps = true
		}
		if custom.CellSpacing > 0 {
			pPr.CellSpacing = &StyleCellSpacingXML{
				W:    fmt.Sprintf("%d", int(custom.CellSpacing)),
				Type: "dxa",
			}
			hasParaProps = true
		}
		if custom.Width > 0 {
			pPr.Width = &StyleWidthXML{
				W:    fmt.Sprintf("%d", int(custom.Width)),
				Type: "dxa",
			}
			hasParaProps = true
		}

		if hasParaProps {
			sd.ParaProps = pPr
		}

		xf.Styles = append(xf.Styles, sd)
	}

	if len(xf.Styles) == 0 && xf.DocDefaults == nil {
		return nil
	}

	data, _ := xml.MarshalIndent(xf, "", "  ")
	return append([]byte(xml.Header), data...)
}
