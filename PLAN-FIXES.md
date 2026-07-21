# words-xml: Comprehensive Fix Plan

Audit date: 2026-07-21
Status: PLANNING

---

## Priority Order (HIGH → MEDIUM → LOW)

### ROUND 1 — HIGH IMPACT (Fix first)

---

#### FIX-01: Resolve theme font references
**Problem**: `RFontsVal` struct missing `asciiTheme`, `hAnsiTheme`, `eastAsiaTheme`, `cstheme`. Modern Word docs use theme refs (e.g. `w:asciiTheme="minorHAnsi"`) instead of explicit font names. Result: `font=""` on most `<span>` elements.

**Files**:
- `words/ooxml.go:184-189` — add 4 theme attrs to `RFontsVal`
- `words/types.go` — add `ThemeFonts` to `TextRun` and `ThemeData` to `ParsedDocument` (already has `ThemeData`)
- `words/preprocessor.go:1327-1335` — in `applyRunProps`, if `Ascii/HAnsi` empty but `AsciiTheme`/`HAnsiTheme` set, resolve from theme map
- `words/preprocessor.go:2040-2056` — in `buildInlineText`, use resolved font name for comparison

**Theme resolution map**:
```
minorHAnsi → theme.MajorHAnsi (or MinorHAnsi)
majorHAnsi → theme.MajorHAnsi
minorAscii → theme.MajorHAnsi (or MinorHAnsi)
majorAscii → theme.MajorHAnsi
minorEastAsia → theme.MajorEastAsia
majorEastAsia → theme.MajorEastAsia
minorBidi → theme.MajorBidi
majorBidi → theme.MajorBidi
```

**Test**: Add test with docx using theme fonts, verify `<span font="Calibri">` appears.

---

#### FIX-02: Border width unit conversion
**Problem**: Border widths in `at` attributes and `<s:custom>` use raw OOXML values (eighths-of-point) instead of declared unit (inches). A 1pt border emits `bt 8` instead of `bt 0.014`.

**Files**:
- `words/preprocessor.go:1372-1387` — `buildBorderAttr`: convert `Sz` from eighths-of-point to inches (`/ 576.0`)
- `words/preprocessor.go:505` — `buildStyleMap`: change `/8.0` to `/576.0` for `BorderWidth`

**Formula**: `inches = raw_sz / 8.0 / 72.0 = raw_sz / 576.0`

**Test**: Add test verifying border width values are in inches.

---

#### FIX-03: Remove `<colspec>` from inside `<table>`
**Problem**: `<colspec ref="n" w=".."/>` is emitted inside `<table>` but is NOT defined in the spec grammar. Column widths belong only in `<s:col>` inside `<style>` (already emitted correctly).

**Files**:
- `words/preprocessor.go:2147-2149` — remove `<colspec>` emission loop
- `words/verify.go:320-341` — remove `<colspec>` validation from table verification

**Test**: Run verify on all outputs, confirm no regressions.

---

#### FIX-04: Table cell content emit lists
**Problem**: `writeTableIndent` only handles `ci.Type == "paragraph"` and `"table"`, not `"list"`. Lists inside `<td>` are silently dropped.

**Files**:
- `words/preprocessor.go:2182-2188` — add `case "list"` handler that calls `emitListGroup`

**Test**: Create test docx with list inside table cell, verify `<li>` appears in output.

---

#### FIX-05: Note bodies emit non-paragraph content
**Problem**: Footnote/endnote/comment bodies only emit `ci.Type == "paragraph"`. Lists, tables inside notes are lost.

**Files**:
- `words/preprocessor.go:1494-1500` — footnote/endnote body loop: add list/table handling
- `words/preprocessor.go:1511-1515` — comment body loop: same fix

**Test**: Create test with list inside footnote, verify output.

---

#### FIX-06: Add `lang` to table cells
**Problem**: Spec defines `lang=".."` on `<th>`/`<td>` but emitter never produces it. `ParsedTableCell` has no `Lang` field.

**Files**:
- `words/types.go` — add `Lang string` to `ParsedTableCell`
- `words/preprocessor.go` — in table cell parsing, extract `lang` from cell's first run or cell properties
- `words/preprocessor.go:2157-2180` — emit `lang=".."` on `<td>`/`<th>` if non-empty

**Test**: Add test with multilingual table, verify `lang` attribute.

---

### ROUND 2 — MEDIUM IMPACT

---

#### FIX-07: Merge theme data into defaultFont
**Problem**: `defaultFont` uses hardcoded `"Times New Roman"` fallback instead of actual theme default (often `"Calibri"`). Wrong suppression/emission of `font=` attributes.

**Files**:
- `words/preprocessor.go:404-424` — after extracting theme, resolve theme refs in docDefaults and merge into `defaultFont`

**Test**: Verify font suppression works correctly with theme-based defaults.

---

#### FIX-08: Parse paragraph-level line spacing
**Problem**: `parseParagraph` ignores `Line`/`LineRule` from `p.PPr.Spacing` even though OOXML struct has them.

**Files**:
- `words/types.go` — add `LineSpacing float64` and `LineRule string` to `ParsedParagraph`
- `words/preprocessor.go:698-701` — read `Spacing.Line` and `Spacing.LineRule`

**Test**: Add test with paragraph having custom line spacing.

---

#### FIX-09: Emit `<s:line>`/`<s:indent>`/`<s:align>` for all styles
**Problem**: Only emitted for Normal style. Heading and other styles with custom layout lose those in style block.

**Files**:
- `words/preprocessor.go:1548-1601` — extend style emission to emit `<s:line>`, `<s:indent>`, `<s:align>` for heading levels and other styles with non-zero values using `c` attribute

**Test**: Verify heading styles emit their own layout elements.

---

#### FIX-10: `<meta>` guard include Modified
**Problem**: If only `Modified` is set, entire `<meta>` block is skipped.

**Files**:
- `words/preprocessor.go:1441` — add `|| doc.Meta.Modified != ""` to condition

**Test**: Test with docx having only modified date.

---

#### FIX-11: Bookmarks in document order within `<notes>`
**Problem**: All bookmarks appended at end, not interleaved with footnotes/endnotes per spec.

**Files**:
- `words/preprocessor.go:271-311` — collect all notes items (fn, en, bm, comment) into a single slice during parsing, sort by document order, then assign to `doc.Notes`

**Test**: Verify ordering matches source document.

---

#### FIX-12: List restart forces `<ol>` split
**Problem**: Same `numId` with `w:lvlOverride/w:startOverride` stays merged into single `<ol>`.

**Files**:
- `words/preprocessor.go:1800-1858` — in `emitListItems`, check if current item has `startOverride` and force split

**Test**: Create test with restarting list.

---

#### FIX-13: Add `<s:indent>`/`<s:theme>` to verifier
**Problem**: Emitted by preprocessor but verifier's `validStyle` map doesn't list them → false warnings.

**Files**:
- `words/verify.go:210-213` — add `"indent": true, "theme": true` to `validStyle` map

**Test**: Run verify, confirm zero warnings.

---

#### FIX-14: Deterministic heading `<s:gap>` selection
**Problem**: Go map iteration means random style per heading level per run.

**Files**:
- `words/preprocessor.go:1548-1562` — prefer the built-in HeadingN style (styleId matches `Heading[1-9]`) over random selection

**Test**: Run multiple times, verify consistent output.

---

#### FIX-15: Collect tabs from headers/footers/styles
**Problem**: Only body content scanned for `<s:tab>`.

**Files**:
- `words/preprocessor.go:1618-1637` — extend tab collection to scan headers, footers, and style definitions

**Test**: Create test with header tab stops.

---

### ROUND 3 — LOW IMPACT (Nice to have)

---

#### FIX-16: Inconsistent default filtering for fontEA/fontCS
**Problem**: `fontEA`/`fontCS` not compared to defaults (always emitted if present), while `font`/`size`/`color` are.

**Files**:
- `words/preprocessor.go:2043-2047` — add default comparison for `fontEA` and `fontCS`

---

#### FIX-17: `<s:cols>` per-section storage
**Problem**: Only captures last section's column info.

**Files**:
- `words/types.go` — change `Cols`/`ColsSpace` to per-section storage
- `words/preprocessor.go` — emit `<s:cols>` alongside each `<s:page>`

---

#### FIX-18: Verifier cleanup
**Problem**: Verifier accepts `<en>`/`<en-ref>` (not in spec), rejects `<br type="clear">` (in spec).

**Files**:
- `words/verify.go` — remove `<en>`, `<en-ref>` from valid elements; add `"clear"` to valid break types

---

## Implementation Order

```
Phase 1 (HIGH):  FIX-01 → FIX-02 → FIX-03 → FIX-04 → FIX-05 → FIX-06
Phase 2 (MED):   FIX-07 → FIX-08 → FIX-09 → FIX-10 → FIX-11 → FIX-12 → FIX-13 → FIX-14 → FIX-15
Phase 3 (LOW):   FIX-16 → FIX-17 → FIX-18
```

## Verification

After each phase:
1. `go test ./words/ -v` — all unit tests pass
2. `go run ./examples/scripts/generate/main.go` — all outputs generated + verified clean
3. `go run ./examples/scripts/verify/main.go` — standalone verify pass
4. Manual check on `default-akta-pendirian-pt_source.docx` for heading suppression
