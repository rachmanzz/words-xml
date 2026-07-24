package docx

import "encoding/xml"

type FootnotesXML struct {
	XMLName  xml.Name       `xml:"w:footnotes"`
	Xmlns    string         `xml:"xmlns:w,attr"`
	XmlnsR   string         `xml:"xmlns:r,attr,omitempty"`
	Items    []*FootnoteXML `xml:"w:footnote"`
}

type FootnoteXML struct {
	XMLName     xml.Name           `xml:"w:footnote"`
	ID          string             `xml:"w:id,attr"`
	Type        string             `xml:"w:type,attr,omitempty"`
	Paragraphs  []*FootnoteParaXML `xml:"w:p"`
}

type FootnoteParaXML struct {
	XMLName     xml.Name                `xml:"w:p"`
	Properties  *FootnoteParaPropsXML   `xml:"w:pPr,omitempty"`
	Run         []*FootnoteRunXML       `xml:"w:r"`
}

type FootnoteParaPropsXML struct {
	XMLName   xml.Name            `xml:"w:pPr"`
	Style     *FootnoteStyleXML   `xml:"w:pStyle,omitempty"`
	RPr       *FootnoteRPropsXML  `xml:"w:rPr,omitempty"`
}

type FootnoteStyleXML struct {
	Val string `xml:"w:val,attr"`
}

type FootnoteRunXML struct {
	XMLName    xml.Name           `xml:"w:r"`
	Properties *FootnoteRPropsXML `xml:"w:rPr,omitempty"`
	Text       *FootnoteTextXML   `xml:"w:t,omitempty"`
}

type FootnoteRPropsXML struct {
	XMLName  xml.Name `xml:"w:rPr"`
	RStyle   *FootnoteStyleXML `xml:"w:rStyle,omitempty"`
	Font     *FootnoteFontXML  `xml:"w:rFonts,omitempty"`
	Size     *FootnoteSizeXML  `xml:"w:sz,omitempty"`
	SizeCS   *FootnoteSizeXML  `xml:"w:szCs,omitempty"`
}

type FootnoteFontXML struct {
	Ascii    string `xml:"w:ascii,attr,omitempty"`
	HAnsi    string `xml:"w:hAnsi,attr,omitempty"`
}

type FootnoteSizeXML struct {
	Val string `xml:"w:val,attr"`
}

type FootnoteTextXML struct {
	Space string `xml:"xml:space,attr,omitempty"`
	Text  string `xml:",chardata"`
}

type CommentsXML struct {
	XMLName  xml.Name        `xml:"w:comments"`
	Xmlns    string          `xml:"xmlns:w,attr"`
	XmlnsR   string          `xml:"xmlns:r,attr,omitempty"`
	Items    []*CommentXML   `xml:"w:comment"`
}

type CommentXML struct {
	XMLName     xml.Name             `xml:"w:comment"`
	ID          string               `xml:"w:id,attr"`
	Author      string               `xml:"w:author,attr"`
	Date        string               `xml:"w:date,attr,omitempty"`
	Initials    string               `xml:"w:initials,attr,omitempty"`
	Paragraphs  []*CommentParaXML    `xml:"w:p"`
}

type CommentParaXML struct {
	XMLName    xml.Name              `xml:"w:p"`
	Properties *CommentParaPropsXML  `xml:"w:pPr,omitempty"`
	Run        []*CommentRunXML      `xml:"w:r"`
}

type CommentParaPropsXML struct {
	XMLName xml.Name `xml:"w:pPr"`
}

type CommentRunXML struct {
	XMLName    xml.Name           `xml:"w:r"`
	Properties *CommentRPropsXML  `xml:"w:rPr,omitempty"`
	Text       *CommentTextXML    `xml:"w:t,omitempty"`
}

type CommentRPropsXML struct {
	XMLName xml.Name `xml:"w:rPr"`
}

type CommentTextXML struct {
	Space string `xml:"xml:space,attr,omitempty"`
	Text  string `xml:",chardata"`
}

type CommentRangeStartXML struct {
	XMLName xml.Name `xml:"w:commentRangeStart"`
	ID      string   `xml:"w:id,attr"`
}

type CommentRangeEndXML struct {
	XMLName xml.Name `xml:"w:commentRangeEnd"`
	ID      string   `xml:"w:id,attr"`
}

type CommentReferenceRunXML struct {
	XMLName    xml.Name            `xml:"w:r"`
	Properties *CommentRefPropsXML `xml:"w:rPr,omitempty"`
	Ref        *CommentRefXML      `xml:"w:commentReference"`
}

type CommentRefPropsXML struct {
	XMLName xml.Name `xml:"w:rPr"`
	RStyle  *CommentRefStyleXML `xml:"w:rStyle,omitempty"`
}

type CommentRefStyleXML struct {
	Val string `xml:"w:val,attr"`
}

type CommentRefXML struct {
	XMLName xml.Name `xml:"w:commentReference"`
	ID      string   `xml:"w:id,attr"`
}

type BookmarkStartXML struct {
	XMLName xml.Name `xml:"w:bookmarkStart"`
	ID      string   `xml:"w:id,attr"`
	Name    string   `xml:"w:name,attr"`
}

type BookmarkEndXML struct {
	XMLName xml.Name `xml:"w:bookmarkEnd"`
	ID      string   `xml:"w:id,attr"`
}
