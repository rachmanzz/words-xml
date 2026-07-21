package words

import "encoding/xml"

type DocDocument struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main document"`
	Body    DocBody  `xml:"body"`
}

type DocBody struct {
	Paras     []DocPara     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables    []DocTbl      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	Sdts      []DocSdt      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sdt"`
	Sections  []DocSection  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sectPr"`
	Bookmarks []DocBookmark `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bookmarkStart"`
}

type DocSection struct {
	PageSz  *DocPageSize `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgSz"`
	PageMar *DocPageMar  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pgMar"`
	Cols    *DocCols     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cols"`
}

type DocCols struct {
	Num   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num,attr"`
	Space int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main space,attr"`
}

type DocPageSize struct {
	W int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr"`
	H int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main h,attr"`
}

type DocPageMar struct {
	Top    int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top,attr"`
	Right  int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right,attr"`
	Bottom int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom,attr"`
	Left   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr"`
	Header int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main header,attr"`
	Footer int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footer,attr"`
}

type DocBookmark struct {
	ID   int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Name string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name,attr"`
}

type DocPara struct {
	PPr        *ParaProps      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	Runs       []DocRun        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
	Hyperlinks []DocHyperlink  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`
	Textboxes  []DocTextbox    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
	Dir        []DirBdo        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main dir"`
	Bdo        []DirBdo        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bdo"`
}

type DirBdo struct {
	Val   string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	Runs  []DocRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type ParaProps struct {
	PStyle          *StyleRef    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pStyle"`
	NumPr           *NumPr       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numPr"`
	JC              *JCVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	Spacing         *SpacingVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main spacing"`
	Ind             *IndVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ind"`
	Bidi            *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bidi"`
	PBdr            *PBdrProps   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pBdr"`
	Tabs            *TabsVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tabs"`
	PageBreakBefore *struct{}    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pageBreakBefore"`
	TextAlign       *JCVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main textAlignment"`
	Lang            *LangVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lang"`
	OutlineLvl      *IntVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main outlineLvl"`
}

type TabsVal struct {
	Tabs []TabVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
}

type TabVal struct {
	Val    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	Pos    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pos,attr"`
	Leader string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main leader,attr"`
}

type StyleRef struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type JCVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type NumPr struct {
	Ilvl  *IntVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl"`
	NumID *IntVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId"`
}

type IntVal struct {
	Val int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type SpacingVal struct {
	Before   int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main before,attr"`
	After    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main after,attr"`
	Line     int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main line,attr"`
	LineRule string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lineRule,attr"`
}

type IndVal struct {
	Left      int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left,attr"`
	Right     int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right,attr"`
	FirstLine int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main firstLine,attr"`
	Hanging   int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hanging,attr"`
}

type PBdrProps struct {
	Top    *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top"`
	Bottom *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom"`
	Left   *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left"`
	Right  *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right"`
}

type BorderVal struct {
	Val   string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
	Sz    int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz,attr"`
	Space int    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main space,attr"`
	Color string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color,attr"`
}

type LangVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type DocRun struct {
	RPr          *RunProps       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	Text         []DocText       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
	FldChar      *FldCharVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fldChar"`
	InstrText    []DocInstrText  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main instrText"`
	Pict         *DocPict        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pict"`
	Drawing      *DocDrawing     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main drawing"`
	Sym          *DocSym         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sym"`
	Break        *BrVal          `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main br"`
	Tab          *struct{}       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tab"`
	FootnoteRef  *NoteRef        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footnoteReference"`
	EndnoteRef   *NoteRef        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main endnoteReference"`
	NoBreakHyphen *struct{}       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main noBreakHyphen"`
	SoftHyphen    *struct{}       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main softHyphen"`
	Ins           *DocIns         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ins"`
	Del           *DocDel         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main del"`
}

type RunProps struct {
	B          *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main b"`
	BCs        *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bCs"`
	I          *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main i"`
	ICs        *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main iCs"`
	U          *UVal         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main u"`
	Strike     *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main strike"`
	DStrike    *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main dstrike"`
	SmallCaps  *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main smallCaps"`
	Caps       *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main caps"`
	VertAlign  *VertAlignVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vertAlign"`
	RFonts     *RFontsVal    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rFonts"`
	Sz         *IntVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sz"`
	SzCs       *IntVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main szCs"`
	Color      *ColorVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main color"`
	Highlight  *HighlightVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main highlight"`
	Vanish     *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vanish"`
	Lang       *LangVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lang"`
	Spacing    *IntVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main spacing"`
	Rtl        *struct{}     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rtl"`
}

type UVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type VertAlignVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type RFontsVal struct {
	Ascii         string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ascii,attr"`
	HAnsi         string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hAnsi,attr"`
	EastAsia      string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsia,attr"`
	CS            string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cs,attr"`
	AsciiTheme    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main asciiTheme,attr"`
	HAnsiTheme    string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hAnsiTheme,attr"`
	EastAsiaTheme string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main eastAsiaTheme,attr"`
	CSTheme       string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main cstheme,attr"`
}

type ColorVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type HighlightVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type FldCharVal struct {
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
}

type DocInstrText struct {
	Text string `xml:",chardata"`
}

type DocText struct {
	Text  string `xml:",chardata"`
	Space string `xml:"http://www.w3.org/XML/1998/namespace space,attr"`
}

type BrVal struct {
	Type string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
}

type NoteRef struct {
	ID int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
}

type DocSym struct {
	Char string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main char,attr"`
	Font string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main font,attr"`
}

type DocIns struct {
	Runs []DocRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type DocDel struct {
	Runs []DocRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type DocPict struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

type DocHyperlink struct {
	ID    string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	RID   string  `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	Runs  []DocRun `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type DocTextbox struct {
	TxbxContent *DocTxbxContent `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main txbxContent"`
}

type DocTxbxContent struct {
	Paras  []DocPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables []DocTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	Sdts   []DocSdt  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sdt"`
}

type DocDrawing struct {
	Inline       *WpInline       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main inline"`
	Anchor       *WpAnchor       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main anchor"`
	TxbxContent  *DocTxbxContent `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main txbxContent"`
}

type WpInline struct {
	Extent  *WpExtent  `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
	DocPr   *WpDocPr   `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing docPr"`
	Graphic *AGraphic  `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
}

type WpAnchor struct {
	Extent  *WpExtent  `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing extent"`
	DocPr   *WpDocPr   `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing docPr"`
	Graphic *AGraphic  `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphic"`
}

type WpExtent struct {
	Cx int64 `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing cx,attr"`
	Cy int64 `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing cy,attr"`
}

type WpDocPr struct {
	ID   int    `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing id,attr"`
	Name string `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing name,attr"`
	Desc string `xml:"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing descr,attr"`
}

type AGraphic struct {
	GraphicData *AGraphicData `xml:"http://schemas.openxmlformats.org/drawingml/2006/main graphicData"`
}

type AGraphicData struct {
	Pic *PicPic `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture pic"`
}

type PicPic struct {
	BlipFill *PicBlipFill `xml:"http://schemas.openxmlformats.org/drawingml/2006/picture blipFill"`
}

type PicBlipFill struct {
	Blip *PicBlip `xml:"http://schemas.openxmlformats.org/drawingml/2006/main blip"`
}

type PicBlip struct {
	Embed string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships embed,attr"`
}

type DocTbl struct {
	TblPr   *TblPr      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblPr"`
	TblGrid *TblGrid    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblGrid"`
	Rows    []DocTblRow `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tr"`
}

type TblPr struct {
	TblStyle *StringVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblStyle"`
	TblW     *TblWidth   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblW"`
	JC       *JCVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main jc"`
	Ind      *TblWidth   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblInd"`
	Spacing  *TblWidth   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblCellSpacing"`
	Borders  *TblBorders `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblBorders"`
	Caption  *StringVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblCaption"`
	Desc     *StringVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblDescription"`
}

type TblWidth struct {
	W int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr"`
}

type StringVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type TblBorders struct {
	Top    *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main top"`
	Bottom *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bottom"`
	Left   *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main left"`
	Right  *BorderVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main right"`
}

type TblGrid struct {
	Cols []GridCol `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridCol"`
}

type GridCol struct {
	W int `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main w,attr"`
}

type DocTblRow struct {
	TrPr  *TrPr        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main trPr"`
	Cells []DocTblCell `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tc"`
}

type TrPr struct {
	TblHeader *struct{} `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblHeader"`
}

type DocTblCell struct {
	TcPr   *TcPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcPr"`
	Paras  []DocPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables []DocTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	Sdts   []DocSdt  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sdt"`
}

type TcPr struct {
	GridSpan *IntVal     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main gridSpan"`
	VMerge   *VMergeVal  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vMerge"`
	VAlign   *JCVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main vAlign"`
	TextDir  *JCVal      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main textDirection"`
	NoWrap   *struct{}   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main noWrap"`
	Borders  *TblBorders `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tcBorders"`
}

type VMergeVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type DocSdt struct {
	Content DocSdtContent `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sdtContent"`
}

type DocSdtContent struct {
	Paras  []DocPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Tables []DocTbl  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tbl"`
	Sdts   []DocSdt  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main sdt"`
}

type DocNumbering struct {
	AbstractNums []AbstractNum `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNum"`
	Nums         []NumDef      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main num"`
}

type AbstractNum struct {
	ID     int           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId,attr"`
	Levels []AbstractLvl `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvl"`
}

type AbstractLvl struct {
	Ilvl   int     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl,attr"`
	NumFmt *FmtVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numFmt"`
}

type FmtVal struct {
	Val string `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr"`
}

type NumDef struct {
	NumID         int           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main numId,attr"`
	AbstractNumID *IntVal       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main abstractNumId"`
	LvlOverrides  []LvlOverride `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main lvlOverride"`
}

type LvlOverride struct {
	Ilvl          int     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main ilvl,attr"`
	StartOverride *IntVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main startOverride"`
}

type DocFootnotes struct {
	Footnotes []DocNote `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main footnote"`
}

type DocEndnotes struct {
	Endnotes []DocNote `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main endnote"`
}

type DocComments struct {
	Comments []DocComment `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main comment"`
}

type DocNote struct {
	ID    int       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Type  string    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	Paras []DocPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
}

type DocComment struct {
	ID     int       `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr"`
	Author string    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main author,attr"`
	Date   string    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main date,attr"`
	Paras  []DocPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
}

type RelsDoc struct {
	Items []RelsItem `xml:"Relationship"`
}

type RelsItem struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
	Type   string `xml:"Type,attr"`
}

type CoreProps struct {
	Title       string `xml:"http://purl.org/dc/elements/1.1/ title"`
	Creator     string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Subject     string `xml:"http://purl.org/dc/elements/1.1/ subject"`
	Description string `xml:"http://purl.org/dc/elements/1.1/ description"`
	Created     string `xml:"http://purl.org/dc/terms/ created"`
	Modified    string `xml:"http://purl.org/dc/terms/ modified"`
	Keywords    string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties keywords"`
}

type DocStyles struct {
	DocDefaults *DocDefaults `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main docDefaults"`
	Styles      []DocStyleDef `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main style"`
}

type DocDefaults struct {
	RunPropsDefault *RunPropsDefault `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPrDefault"`
}

type RunPropsDefault struct {
	RunProps *RunProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
}

type DocStyleDef struct {
	ID        string     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main styleId,attr"`
	Type      string     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main type,attr"`
	Name      *StringVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name"`
	BasedOn   *StringVal `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main basedOn"`
	RunProps   *RunProps  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main rPr"`
	ParaProps  *ParaProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	TblPr      *TblPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main tblPr"`
}

func (s *DocStyles) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias DocStyles
	var a Alias
	if err := d.DecodeElement(&a, &start); err != nil {
		return err
	}
	*s = DocStyles(a)
	return nil
}

type ATheme struct {
	ThemeElements *AThemeElements `xml:"http://schemas.openxmlformats.org/drawingml/2006/main themeElements"`
}

type AThemeElements struct {
	FontScheme *AFontScheme `xml:"http://schemas.openxmlformats.org/drawingml/2006/main fontScheme"`
	ClrScheme  *AClrScheme  `xml:"http://schemas.openxmlformats.org/drawingml/2006/main clrScheme"`
}

type AFontScheme struct {
	Minor *AFontSchemeFace `xml:"http://schemas.openxmlformats.org/drawingml/2006/main minorFont"`
	Major *AFontSchemeFace `xml:"http://schemas.openxmlformats.org/drawingml/2006/main majorFont"`
}

type AFontSchemeFace struct {
	Latin *ATypeface `xml:"http://schemas.openxmlformats.org/drawingml/2006/main latin"`
	Ea   *ATypeface `xml:"http://schemas.openxmlformats.org/drawingml/2006/main ea"`
	Cs   *ATypeface `xml:"http://schemas.openxmlformats.org/drawingml/2006/main cs"`
}

type ATypeface struct {
	Typeface string `xml:"typeface,attr"`
}

type AClrScheme struct {
	Dk1 *AClrSchemeEntry `xml:"http://schemas.openxmlformats.org/drawingml/2006/main dk1"`
	Lt1 *AClrSchemeEntry `xml:"http://schemas.openxmlformats.org/drawingml/2006/main lt1"`
}

type AClrSchemeEntry struct {
	SrgbClr *ASrgbClr `xml:"http://schemas.openxmlformats.org/drawingml/2006/main srgbClr"`
}

type ASrgbClr struct {
	Val string `xml:"http://schemas.openxmlformats.org/drawingml/2006/main val,attr"`
}
