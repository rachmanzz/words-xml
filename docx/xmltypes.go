package docx

import "encoding/xml"

type WordsXML struct {
	XMLName xml.Name     `xml:"words"`
	Version string       `xml:"version,attr"`
	Mode    string       `xml:"mode,attr"`
	Meta    *MetaXML     `xml:"meta"`
	Style   *StyleXML    `xml:"style"`
	Headers []HeaderXML  `xml:"header"`
	Footers []FooterXML  `xml:"footer"`
	Write   *WriteXML    `xml:"write"`
	Notes   *NotesXML    `xml:"notes"`
}

type MetaXML struct {
	Title       string `xml:"title"`
	Author      string `xml:"author"`
	Created     string `xml:"created"`
	Modified    string `xml:"modified"`
	Subject     string `xml:"subject"`
	Keywords    string `xml:"keywords"`
	Description string `xml:"description"`
}

type StyleXML struct {
	Unit    string        `xml:"unit,attr"`
	Pages   []StylePage   `xml:"page"`
	Cols    []StyleCols   `xml:"cols"`
	Gaps    []StyleGap    `xml:"gap"`
	Indents []StyleIndent `xml:"indent"`
	Aligns  []StyleAlign  `xml:"align"`
	Lines   []StyleLine   `xml:"line"`
	Cols2   []StyleCol    `xml:"col"`
	Tabs    []StyleTab    `xml:"tab"`
	Theme   *StyleTheme   `xml:"theme"`
	Customs []StyleCustom `xml:"custom"`
}

type StylePage struct {
	Size string  `xml:"size,attr"`
	W    float64 `xml:"w,attr"`
	H    float64 `xml:"h,attr"`
	MT   float64 `xml:"mt,attr"`
	MB   float64 `xml:"mb,attr"`
	ML   float64 `xml:"ml,attr"`
	MR   float64 `xml:"mr,attr"`
	MH   float64 `xml:"mh,attr"`
	MF   float64 `xml:"mf,attr"`
}

type StyleCols struct {
	N     int     `xml:"n,attr"`
	Space float64 `xml:"space,attr"`
}

type StyleGap struct {
	EL     string  `xml:"el,attr"`
	C      string  `xml:"c,attr"`
	Before float64 `xml:"before,attr"`
	After  float64 `xml:"after,attr"`
}

type StyleIndent struct {
	EL      string  `xml:"el,attr"`
	C       string  `xml:"c,attr"`
	Left    float64 `xml:"left,attr"`
	Right   float64 `xml:"right,attr"`
	First   float64 `xml:"firstLine,attr"`
	Hanging float64 `xml:"hanging,attr"`
}

type StyleAlign struct {
	EL    string `xml:"el,attr"`
	C     string `xml:"c,attr"`
	Value string `xml:"value,attr"`
}

type StyleLine struct {
	EL    string  `xml:"el,attr"`
	C     string  `xml:"c,attr"`
	Value float64 `xml:"value,attr"`
	Rule  string  `xml:"rule,attr"`
}

type StyleCol struct {
	Ref int     `xml:"ref,attr"`
	W   float64 `xml:"w,attr"`
}

type StyleTab struct {
	EL     string  `xml:"el,attr"`
	C      string  `xml:"c,attr"`
	Pos    float64 `xml:"pos,attr"`
	Align  string  `xml:"align,attr"`
	Leader string  `xml:"leader,attr"`
}

type StyleTheme struct {
	Font   string `xml:"font,attr"`
	FontEA string `xml:"fontEA,attr"`
	FontCS string `xml:"fontCS,attr"`
	Fg     string `xml:"fg,attr"`
	Bg     string `xml:"bg,attr"`
}

type StyleCustom struct {
	Name           string  `xml:"name,attr"`
	Type           string  `xml:"type,attr"`
	BasedOn        string  `xml:"basedOn,attr"`
	Font           string  `xml:"font,attr"`
	FontEA         string  `xml:"fontEA,attr"`
	FontCS         string  `xml:"fontCS,attr"`
	Size           float64 `xml:"size,attr"`
	SizeCS         float64 `xml:"sizeCS,attr"`
	Color          string  `xml:"color,attr"`
	Bold           string  `xml:"bold,attr"`
	Italic         string  `xml:"italic,attr"`
	Underline      string  `xml:"underline,attr"`
	Strikethrough  string  `xml:"strikethrough,attr"`
	SmallCaps      string  `xml:"smallCaps,attr"`
	Uppercase      string  `xml:"uppercase,attr"`
	Alignment      string  `xml:"alignment,attr"`
	SpacingBefore  float64 `xml:"spacingBefore,attr"`
	SpacingAfter   float64 `xml:"spacingAfter,attr"`
	LineSpacing    float64 `xml:"lineSpacing,attr"`
	LineRule       string  `xml:"lineRule,attr"`
	IndentLeft     float64 `xml:"indentLeft,attr"`
	IndentRight    float64 `xml:"indentRight,attr"`
	IndentFirst    float64 `xml:"indentFirst,attr"`
	IndentHanging  float64 `xml:"indentHanging,attr"`
	BorderWidth    float64 `xml:"borderWidth,attr"`
	BorderColor    string  `xml:"borderColor,attr"`
	BorderStyle    string  `xml:"borderStyle,attr"`
	CellSpacing    float64 `xml:"cellSpacing,attr"`
	Width          float64 `xml:"width,attr"`
}

type HeaderXML struct {
	ID      int           `xml:"id,attr"`
	Content []interface{} `xml:",any"`
}

type FooterXML struct {
	ID      int           `xml:"id,attr"`
	Content []interface{} `xml:",any"`
}

type WriteXML struct {
	Content []interface{} `xml:",any"`
}

type NotesXML struct {
	Items []NoteXML `xml:",any"`
}

type NoteXML struct {
	XMLName xml.Name
	ID      int    `xml:"id,attr"`
	Type    string `xml:"type,attr"`
	Author  string `xml:"author,attr"`
	Date    string `xml:"date,attr"`
	Name    string `xml:"name,attr"`
}

type BlockXML struct {
	XMLName xml.Name
	Content []interface{} `xml:",any"`
}

type ParaXML struct {
	XMLName       xml.Name      `xml:"p"`
	Style         string        `xml:"c,attr"`
	Align         string        `xml:"align,attr"`
	IndentLeft    float64       `xml:"indentLeft,attr"`
	IndentRight   float64       `xml:"indentRight,attr"`
	IndentFirst   float64       `xml:"indentFirst,attr"`
	IndentHanging float64       `xml:"indentHanging,attr"`
	Dir           string        `xml:"dir,attr"`
	Lang          string        `xml:"lang,attr"`
	VAlign        string        `xml:"valign,attr"`
	At            string        `xml:"at,attr"`
	Content       []interface{} `xml:",any"`
}

type HeadingXML struct {
	XMLName       xml.Name
	Level         int
	Style         string        `xml:"c,attr"`
	Align         string        `xml:"align,attr"`
	IndentLeft    float64       `xml:"indentLeft,attr"`
	IndentRight   float64       `xml:"indentRight,attr"`
	IndentFirst   float64       `xml:"indentFirst,attr"`
	IndentHanging float64       `xml:"indentHanging,attr"`
	Dir           string        `xml:"dir,attr"`
	Lang          string        `xml:"lang,attr"`
	VAlign        string        `xml:"valign,attr"`
	At            string        `xml:"at,attr"`
	Content       []interface{} `xml:",any"`
}

type BlockquoteXML struct {
	XMLName       xml.Name      `xml:"blockquote"`
	Style         string        `xml:"c,attr"`
	Align         string        `xml:"align,attr"`
	IndentLeft    float64       `xml:"indentLeft,attr"`
	IndentRight   float64       `xml:"indentRight,attr"`
	IndentFirst   float64       `xml:"indentFirst,attr"`
	IndentHanging float64       `xml:"indentHanging,attr"`
	Dir           string        `xml:"dir,attr"`
	Lang          string        `xml:"lang,attr"`
	VAlign        string        `xml:"valign,attr"`
	At            string        `xml:"at,attr"`
	Content       []interface{} `xml:",any"`
}

type PreXML struct {
	XMLName       xml.Name      `xml:"pre"`
	Style         string        `xml:"c,attr"`
	Align         string        `xml:"align,attr"`
	IndentLeft    float64       `xml:"indentLeft,attr"`
	IndentRight   float64       `xml:"indentRight,attr"`
	IndentFirst   float64       `xml:"indentFirst,attr"`
	IndentHanging float64       `xml:"indentHanging,attr"`
	Content       []interface{} `xml:",any"`
}

type UlXML struct {
	XMLName xml.Name `xml:"ul"`
	Type    string   `xml:"type,attr"`
	Items   []LiXML  `xml:"li"`
}

type OlXML struct {
	XMLName xml.Name `xml:"ol"`
	Type    string   `xml:"type,attr"`
	Start   int      `xml:"start,attr"`
	Items   []LiXML  `xml:"li"`
}

type LiXML struct {
	XMLName xml.Name     `xml:"li"`
	Tag     string       `xml:"tag,attr"`
	Content []interface{} `xml:",any"`
}

type TableXML struct {
	XMLName   xml.Name         `xml:"table"`
	ID        int              `xml:"id,attr"`
	Style     string           `xml:"c,attr"`
	Width     float64          `xml:"width,attr"`
	Align     string           `xml:"align,attr"`
	Indent    float64          `xml:"indent,attr"`
	CellSpace float64          `xml:"cellSpacing,attr"`
	Caption   string           `xml:"caption,attr"`
	Summary   string           `xml:"summary,attr"`
	At        string           `xml:"at,attr"`
	Rows      []TableRowXML    `xml:"tr"`
}

type TableRowXML struct {
	XMLName  xml.Name        `xml:"tr"`
	Cells    []TableCellXML  `xml:"td"`
	ThCells  []TableCellXML  `xml:"th"`
}

type TableCellXML struct {
	XMLName    xml.Name      `xml:"td"`
	ColSpan    int           `xml:"colspan,attr"`
	RowSpan    int           `xml:"rowspan,attr"`
	VAlign     string        `xml:"valign,attr"`
	TextDir    string        `xml:"textDir,attr"`
	Lang       string        `xml:"lang,attr"`
	NoWrap     string        `xml:"noWrap,attr"`
	At         string        `xml:"at,attr"`
	Content    []interface{} `xml:",any"`
}

type InlineXML struct {
	XMLName xml.Name
	Attr    []xml.Attr
	Text    string       `xml:",chardata"`
	Href    string       `xml:"href,attr"`
	Title   string       `xml:"title,attr"`
	Type    string       `xml:"type,attr"`
	ID      string       `xml:"id,attr"`
	Alt     string       `xml:"alt,attr"`
	Content []InlineXML  `xml:",any"`
}

type SpanXML struct {
	XMLName  xml.Name `xml:"span"`
	Font     string   `xml:"font,attr"`
	FontEA   string   `xml:"fontEA,attr"`
	FontCS   string   `xml:"fontCS,attr"`
	Size     float64  `xml:"size,attr"`
	SizeCS   float64  `xml:"sizeCS,attr"`
	Color    string   `xml:"color,attr"`
	Highlight string `xml:"highlight,attr"`
	Lang     string   `xml:"lang,attr"`
	Hidden   string   `xml:"hidden,attr"`
	Dir      string   `xml:"dir,attr"`
	Content  string   `xml:",chardata"`
}
