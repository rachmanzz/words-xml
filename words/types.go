package words

type ProcessedDocument struct {
	WordsXML string
}

type Meta struct {
	Title       string
	Subject     string
	Author      string
	Created     string
	Modified    string
	Keywords    string
	Description string
}

type PageLayout struct {
	WidthInch    float64
	HeightInch   float64
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	MarginLeft   float64
	HeaderMargin float64
	FooterMargin float64
	Cols         int
	ColsSpace    float64
}

type ParsedDocument struct {
	Meta          Meta
	PageLayout    PageLayout
	PageSections  []PageLayout
	Content       []ContentItem
	Headers       []HeaderFooter
	Footers       []HeaderFooter
	Notes         []NoteItem
	StyleMap      map[string]StyleDef
	StyleNameMap  map[string]string
	Theme         *ThemeData
	DefaultFont   StyleDef
	Mode          string
	AllTables     []*ParsedTable
	NumStartMap   map[int]map[int]int
	NumToAbstract map[int]int
	NumFmtMap     map[string]string
	Cols          int
	ColsSpace     float64
}

type HeaderFooter struct {
	ID      int
	Content []ContentItem
}

type ContentItem struct {
	Type      string
	Paragraph *ParsedParagraph
	Table     *ParsedTable
}

type NoteItem struct {
	Type   string
	ID     int
	Name   string
	Author string
	Date   string
	Body   []ContentItem
}

type StyleDef struct {
	ID             string
	Name           string
	Type           string
	BasedOn        string
	Family         string
	FontEA         string
	FontCS         string
	SizePt         float64
	SizeCS         float64
	Color          string
	Bold           bool
	Italic         bool
	Underline      string
	Strikethrough  bool
	SmallCaps      bool
	Uppercase      bool
	HeadingLevel   int
	Align          string
	SpacingBefore  float64
	SpacingAfter   float64
	LineSpacing    float64
	LineRule       string
	IndentLeft     float64
	IndentRight    float64
	IndentFirst    float64
	IndentHanging  float64
	BorderWidth    float64
	BorderColor    string
	BorderStyle    string
	CellSpacing    float64
	Width          float64
	Tabs           []ParsedTab
}

type ThemeData struct {
	Font    string
	FontEA  string
	FontCS  string
	Fg      string
	Bg      string
	FontMap map[string]string
}

type ParsedParagraph struct {
	StyleID      string
	StyleName    string
	HeadingLevel int
	IsList       bool
	ListLevel    int
	NumID        int
	ListFormat   string
	IsQuote      bool
	IsCode       bool
	Bidi         bool
	Lang         string
	VAlign       string
	Align        string
	SpacingBefore float64
	SpacingAfter  float64
	LineSpacing   float64
	LineRule      string
	IndentLeft    float64
	IndentRight   float64
	IndentFirst   float64
	IndentHanging float64
	Tabs          []ParsedTab
	BorderTop     *BorderInfo
	BorderBottom  *BorderInfo
	BorderLeft    *BorderInfo
	BorderRight   *BorderInfo
	Runs          []TextRun
}

type ParsedTab struct {
	Pos    float64
	Align  string
	Leader string
}

type TextRun struct {
	Text         string
	Bold         bool
	Italic       bool
	Underline    string
	Strike       bool
	SmallCaps    bool
	AllCaps      bool
	SuperScript  bool
	SubScript    bool
	IsBoldCS     bool
	IsItalicCS   bool
	FontFamily   string
	FontEA       string
	FontCS       string
	FontSizePt   float64
	FontSizeCS   float64
	FontColor    string
	Highlight    string
	Hidden       bool
	Lang         string
	IsTab        bool
	IsLineBreak  bool
	BreakType    string
	IsHyperlink  bool
	HyperlinkURL string
	IsImage      bool
	ImageSrc     string
	ImageWidth   float64
	ImageHeight  float64
	ImageAlt     string
	IsFootnoteRef bool
	NoteID       int
	NoteType     string
	IsSym        bool
	SymChar      string
	IsInserted   bool
	IsDeleted    bool
	IsRTL        bool
}

type BorderInfo struct {
	Val   string
	Sz    int
	Space int
	Color string
}

type ParsedTable struct {
	ID          int
	Width       float64
	Alignment   string
	Indent      float64
	CellSpace   float64
	Caption     string
	Summary     string
	StyleName   string
	Grid        []float64
	Rows        []ParsedTableRow
	BorderTop    *BorderInfo
	BorderBottom *BorderInfo
	BorderLeft   *BorderInfo
	BorderRight  *BorderInfo
}

type ParsedTableRow struct {
	IsHeader bool
	Cells    []ParsedTableCell
}

type ParsedTableCell struct {
	GridSpan   int
	RowSpan    int
	VMerge     int
	Omitted    bool
	VAlign     string
	TextDir    string
	NoWrap     bool
	Lang       string
	Content    []ContentItem
	BorderTop    *BorderInfo
	BorderBottom *BorderInfo
	BorderLeft   *BorderInfo
	BorderRight  *BorderInfo
}
