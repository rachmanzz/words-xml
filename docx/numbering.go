package docx

import "encoding/xml"

type NumberingXML struct {
	XMLName      xml.Name          `xml:"w:numbering"`
	Xmlns        string            `xml:"xmlns:w,attr"`
	AbstractNums []*AbstractNumXML `xml:"w:abstractNum"`
	Nums         []*NumXML         `xml:"w:num"`
}

type AbstractNumXML struct {
	XMLName xml.Name       `xml:"w:abstractNum"`
	ID      int            `xml:"w:abstractNumId,attr"`
	Levels  []*NumLevelXML `xml:"w:lvl"`
}

type NumLevelXML struct {
	XMLName     xml.Name                   `xml:"w:lvl"`
	Level       int                        `xml:"w:ilvl,attr"`
	Start       *DecimalNumberXML          `xml:"w:start,omitempty"`
	NumFmt      *NumFmtXML                 `xml:"w:numFmt,omitempty"`
	LevelText   *LevelTextXML              `xml:"w:lvlText,omitempty"`
	LevelJc     *LevelJcXML                `xml:"w:lvlJc,omitempty"`
	ParagraphPr *NumberingParaPropsXML     `xml:"w:pPr,omitempty"`
}

type DecimalNumberXML struct {
	Val int `xml:"w:val,attr"`
}

type NumFmtXML struct {
	Val string `xml:"w:val,attr"`
}

type LevelTextXML struct {
	Val string `xml:"w:val,attr"`
}

type LevelJcXML struct {
	Val string `xml:"w:val,attr"`
}

type NumberingParaPropsXML struct {
	XMLName xml.Name              `xml:"w:pPr"`
	Indent  *NumberingIndentXML   `xml:"w:ind,omitempty"`
}

type NumberingIndentXML struct {
	Left    int `xml:"w:left,attr"`
	Hanging int `xml:"w:hanging,attr"`
}

type NumXML struct {
	XMLName      xml.Name          `xml:"w:num"`
	ID           int               `xml:"w:numId,attr"`
	AbstractNumID *DecimalNumberXML `xml:"w:abstractNumId"`
}

func buildNumberingXML(numDefs []numDef) *NumberingXML {
	if len(numDefs) == 0 {
		return nil
	}

	xf := &NumberingXML{
		Xmlns: "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
	}

	for i, def := range numDefs {
		abstractNum := &AbstractNumXML{
			ID: i + 1,
			Levels: []*NumLevelXML{
				{
					Level: 0,
					Start: &DecimalNumberXML{Val: def.start},
					NumFmt: &NumFmtXML{Val: def.numFmt},
					LevelText: &LevelTextXML{Val: def.levelText},
					LevelJc: &LevelJcXML{Val: "left"},
					ParagraphPr: &NumberingParaPropsXML{
						Indent: &NumberingIndentXML{
							Left:    720,
							Hanging: 360,
						},
					},
				},
			},
		}
		xf.AbstractNums = append(xf.AbstractNums, abstractNum)

		xf.Nums = append(xf.Nums, &NumXML{
			ID:            i + 1,
			AbstractNumID: &DecimalNumberXML{Val: i + 1},
		})
	}

	return xf
}

type numDef struct {
	numFmt    string
	start     int
	levelText string
}

func defaultNumDefs() []numDef {
	return []numDef{
		{numFmt: "bullet", start: 1, levelText: "\u2022"},
		{numFmt: "decimal", start: 1, levelText: "%1."},
		{numFmt: "lowerLetter", start: 1, levelText: "%1."},
		{numFmt: "upperLetter", start: 1, levelText: "%1."},
		{numFmt: "lowerRoman", start: 1, levelText: "%1."},
		{numFmt: "upperRoman", start: 1, levelText: "%1."},
	}
}
