package docx

import (
	"strconv"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

const (
	twipsPerInch  = 1440
	twipsPerPt    = 20
	halfPtsPerPt  = 2
	lineSpaceBase = 240
)

func inchesToTwips(inches float64) int {
	return int(inches * twipsPerInch)
}

func pointsToTwips(points float64) int {
	return int(points * twipsPerPt)
}

func pointsToHalfPoints(points float64) int {
	return int(points * halfPtsPerPt)
}

func lineMultiplierToValue(multiplier float64) int {
	return int(multiplier * lineSpaceBase)
}

func parseColor(hex string) domain.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return domain.ColorBlack
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return domain.Color{R: uint8(r), G: uint8(g), B: uint8(b)}
}

func mapAlignment(align string) domain.Alignment {
	switch strings.ToLower(align) {
	case "center":
		return domain.AlignmentCenter
	case "right":
		return domain.AlignmentRight
	case "both", "justify":
		return domain.AlignmentJustify
	case "distribute":
		return domain.AlignmentDistribute
	default:
		return domain.AlignmentLeft
	}
}

func mapBreakType(brk string) domain.BreakType {
	switch strings.ToLower(brk) {
	case "page":
		return domain.BreakTypePage
	case "column":
		return domain.BreakTypeColumn
	default:
		return domain.BreakTypeLine
	}
}

func mapUnderlineStyle(style string) domain.UnderlineStyle {
	switch strings.ToLower(style) {
	case "single":
		return domain.UnderlineSingle
	case "double":
		return domain.UnderlineDouble
	case "thick":
		return domain.UnderlineThick
	case "dotted":
		return domain.UnderlineDotted
	case "dashed":
		return domain.UnderlineDashed
	case "wave":
		return domain.UnderlineWave
	default:
		return domain.UnderlineNone
	}
}

func mapHighlightColor(name string) domain.HighlightColor {
	switch strings.ToLower(name) {
	case "yellow":
		return domain.HighlightYellow
	case "green":
		return domain.HighlightGreen
	case "cyan":
		return domain.HighlightCyan
	case "magenta":
		return domain.HighlightMagenta
	case "blue":
		return domain.HighlightBlue
	case "red":
		return domain.HighlightRed
	case "darkblue":
		return domain.HighlightDarkBlue
	case "darkcyan":
		return domain.HighlightDarkCyan
	case "darkgreen":
		return domain.HighlightDarkGreen
	case "darkmagenta":
		return domain.HighlightDarkMagenta
	case "darkred":
		return domain.HighlightDarkRed
	case "darkyellow":
		return domain.HighlightDarkYellow
	case "darkgray":
		return domain.HighlightDarkGray
	case "lightgray":
		return domain.HighlightLightGray
	default:
		return domain.HighlightNone
	}
}

func mapVerticalAlignment(align string) domain.VerticalAlignment {
	switch strings.ToLower(align) {
	case "center", "middle":
		return domain.VerticalAlignCenter
	case "bottom":
		return domain.VerticalAlignBottom
	default:
		return domain.VerticalAlignTop
	}
}

func mapPageSize(size string) domain.PageSize {
	switch strings.ToUpper(size) {
	case "A3":
		return domain.PageSize{Width: 16838, Height: 23811}
	case "A4":
		return domain.PageSizeA4
	case "A5":
		return domain.PageSize{Width: 8395, Height: 11906}
	case "A6":
		return domain.PageSize{Width: 5957, Height: 8395}
	case "B5":
		return domain.PageSize{Width: 10120, Height: 14335}
	case "LETTER":
		return domain.PageSizeLetter
	case "LEGAL":
		return domain.PageSizeLegal
	case "TABLOID":
		return domain.PageSize{Width: 15840, Height: 24480}
	case "EXECUTIVE":
		return domain.PageSize{Width: 10440, Height: 13680}
	case "STATEMENT":
		return domain.PageSize{Width: 7920, Height: 12240}
	case "FOLIO":
		return domain.PageSize{Width: 8395, Height: 12240}
	default:
		return domain.PageSizeA4
	}
}

func mapTableStyle(name string) domain.TableStyle {
	switch name {
	case "TableGrid":
		return domain.TableStyleGrid
	case "NormalTable", "Normal":
		return domain.TableStyleNormal
	default:
		return domain.TableStyleNormal
	}
}

func mapListFormat(format string) string {
	switch strings.ToLower(format) {
	case "decimal":
		return "decimal"
	case "lowerletter":
		return "lowerLetter"
	case "upperletter":
		return "upperLetter"
	case "lowerroman":
		return "lowerRoman"
	case "upperroman":
		return "upperRoman"
	case "bullet":
		return "bullet"
	default:
		return "decimal"
	}
}
