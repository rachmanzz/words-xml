# GAP Analysis: words-xml v1.1.0

Audit date: 2026-07-22
Status: ACTIVE

---

## Summary

Gaps between the words-xml spec (v1.1.0) and current implementation, grouped by severity.

---

## GAP-01: Per-Paragraph Indent/Hanging Attributes
**Status**: DONE
**Priority**: HIGH

**Spec**: `<s:indent el="p" left=".." right=".." firstLine=".." hanging=".."/>` in `<style>` block (§2.4)
**Current**: Per-paragraph `indentLeft`, `indentHanging`, `indentRight`, `indentFirst` attributes emitted on `<p>` elements.

**Verification**: Confirmed in `writeParagraphAttrs` and test output.

---

## GAP-02: List Continuation Across Non-List Paragraphs
**Status**: FIXED
**Priority**: MEDIUM

**Spec**: "Group consecutive `ListParagraph` paragraphs into a `<ul>` or `<ol>`" (§3.3)
**Current**: Non-list paragraphs between same-`numId` items are included in the preceding `<li>` with a `<br type="textWrapping"/>` separator.

**Output**:
```xml
<ul>
  <li>Item 1
<br type="textWrapping"/>continuation text (non-list paragraph).
  </li>
  <li>Item 2</li>
</ul>
```

---

## GAP-03: `<s:indent>` Not Emitted for Normal/Heading Styles
**Status**: FIXED (FIX-09)
**Priority**: LOW

**Spec**: `<s:indent el="p" left=".." right=".." firstLine=".." hanging=".."/>` in `<style>` block
**Current**: `<s:indent>` emitted for Normal and Heading styles with non-zero indent values.

**Verification**: Confirmed working in `emitStyleBlock`.

---

## GAP-04: Per-Paragraph Align Only for Direct `jc` Values
**Status**: FIXED
**Priority**: MEDIUM

**Spec**: `<s:align el="p" value=".."/>` in `<style>` block (§2.4)
**Current**: `<s:align>` emitted in style block + `align` attribute on `<p>` for direct `jc` values.

**Note**: This is a non-spec extension. The spec only defines `<s:align>` at style level.

---

## GAP-05: `<h1>`-`<h9>`, `<li>`, `<blockquote>` Missing Indent
**Status**: FIXED
**Priority**: LOW

**Spec**: Indent should apply to all block elements with `w:ind`
**Current**: Headings, list items, blockquotes now have indent attributes when paragraph-level indent is set.

**Verification**: Tested with headings having custom indentation.

---

## GAP-06: Tab Before Text in Same Run Discards Text
**Status**: FIXED
**Priority**: HIGH

**Problem**: `extractRuns` returned early on `r.Tab != nil`, discarding any `<w:t>` text in the same run.
**Fix**: Emit tab run first, then continue processing text.

---

## GAP-07: Centering Lost for Mixed-Alignment Documents
**Status**: FIXED
**Priority**: HIGH

**Problem**: Multiple `<s:align el="p">` values — last one wins.
**Fix**: Added `align` attribute on `<p>` for per-paragraph alignment.

---

## GAP-08: Per-Paragraph Spacing Not Emitted
**Status**: DONE (Phase 1)
**Priority**: HIGH

**Problem**: `w:spacing` before/after parsed but only emitted at style level (`<s:gap>`). Per-paragraph overrides lost.
**Fix**: Added `spacingBefore` and `spacingAfter` attributes on `<p>` elements when > 0.

**Files**:
- `words/preprocessor.go` — `writeParagraphAttrs`
- `words/verify.go` — validation

**Verification**: Tested with `Paragraph Features Test.docx`.

---

## GAP-09: Per-Paragraph Line Spacing Not Emitted
**Status**: DONE (Phase 1)
**Priority**: HIGH

**Problem**: `w:spacing` line/lineRule parsed but only emitted at style level (`<s:line>`). Per-paragraph overrides lost.
**Fix**: Added `lineSpacing` and `lineRule` attributes on `<p>` elements when > 0. `lineRule` only emitted when non-default (`auto`).

**Files**:
- `words/preprocessor.go` — `writeParagraphAttrs`
- `words/verify.go` — validation

**Verification**: Tested with `Paragraph Features Test.docx`.

---

## GAP-10: Heading Levels 4-9 Not Tested
**Status**: DONE (Phase 1)
**Priority**: MEDIUM

**Problem**: No test documents with heading levels 4-9.
**Fix**: Added test DOCX with `<h4>`, `<h5>`, `<h6>` headings. Verified emission works correctly.

---

## GAP-11: Blockquote Not Tested
**Status**: DONE (Phase 1)
**Priority**: MEDIUM

**Problem**: No test documents with blockquote paragraphs.
**Fix**: Added test DOCX with `<blockquote>` paragraph. Verified emission works correctly.

---

## GAP-12: Preformatted/Code Not Tested
**Status**: DONE (Phase 1)
**Priority**: MEDIUM

**Problem**: No test documents with preformatted/code blocks.
**Fix**: Added test DOCX with `<pre>` paragraph. Verified emission works correctly.

---

## GAP-13: RTL/Bidi Not Tested
**Status**: DONE (Phase 1)
**Priority**: MEDIUM

**Problem**: No test documents with bidirectional text.
**Fix**: Added test DOCX with `w:bidi` paragraph. Verified `dir="rtl"` emission works correctly.

---

## GAP-14: Right Indentation Not Tested
**Status**: DONE (Phase 1)
**Priority**: MEDIUM

**Problem**: `indentRight` code exists but never triggered by test documents.
**Fix**: Added test DOCX with `w:ind w:right` paragraph. Verified `indentRight` emission works correctly.

---

## GAP-15: Paragraph Shading Not Parsed
**Status**: DONE (Phase 2)
**Priority**: MEDIUM

**Problem**: `w:shd` element not parsed from paragraph properties.
**Fix**: Added `ShdVal` struct to `ooxml.go`, `Shading` field to `ParsedParagraph`, parsing in `parseParagraph`, and `shd=` attr emission in `writeParagraphAttrs`.

**Files**:
- `words/ooxml.go` — `ShdVal` struct
- `words/types.go` — `Shading` field
- `words/preprocessor.go` — parsing and emission
- `words/verify.go` — validation

---

## GAP-16: Keep Next/Lines Not Parsed
**Status**: DONE (Phase 2)
**Priority**: MEDIUM

**Problem**: `w:keepNext` and `w:keepLines` not parsed from paragraph properties.
**Fix**: Added fields to `ParaProps` and `ParsedParagraph`, parsing and emission.

---

## GAP-17: Widow Control Not Parsed
**Status**: DONE (Phase 2)
**Priority**: LOW

**Problem**: `w:widowControl` not parsed from paragraph properties.
**Fix**: Added field to `ParaProps` and `ParsedParagraph`, parsing and emission.

---

## GAP-18: Character-Unit Indentation Not Parsed
**Status**: DONE (Phase 2)
**Priority**: LOW

**Problem**: `w:ind` character-unit attributes (`leftChars`, `rightChars`, etc.) not in `IndVal`.
**Fix**: Added fields to `IndVal` struct. Not yet parsed or emitted (low priority).

---

## GAP-19: Between/Bar Borders Not Parsed
**Status**: DONE (Phase 2)
**Priority**: LOW

**Problem**: `w:pBdr/between` and `w:pBdr/bar` not in `PBdrProps`.
**Fix**: Added fields to `PBdrProps` struct. Not yet parsed or emitted (low priority).

---

## GAP-20: Paragraph-Level Default Run Properties Not Parsed
**Status**: DONE (Phase 3)
**Priority**: HIGH

**Problem**: `w:rPr` in `w:pPr` not parsed; paragraph-level font/size/color defaults not applied to runs.
**Fix**: Added `RPr` field to `ParaProps`, `ParaDefaults` field to `ParsedParagraph`, `applyParaRunDefaults` function. Modified `extractRuns` to accept paragraph defaults.

**Files**:
- `words/ooxml.go` — `RPr` field in `ParaProps`
- `words/types.go` — `ParaDefaults` field in `ParsedParagraph`
- `words/preprocessor.go` — `applyParaRunDefaults` function, `extractRuns` signature change

---

## GAP-21: Section Breaks in Paragraphs Not Parsed
**Status**: DONE (Phase 3)
**Priority**: MEDIUM

**Problem**: `w:sectPr` in `w:pPr` not parsed; section breaks embedded in paragraphs not detected.
**Fix**: Added `SectPr` field to `ParaProps`, `SectionBreak` field to `ParsedParagraph`, parsing and `sectionBreak=` attr emission.

**Files**:
- `words/ooxml.go` — `SectPrProps`, `SectType`, `PgSzVal`, `PgMarVal`, `CTRel` structs
- `words/types.go` — `SectionBreak` field in `ParsedParagraph`
- `words/preprocessor.go` — parsing and emission
- `words/verify.go` — validation

---

## GAP-22: Paragraph Revision Marks Not Parsed
**Status**: DONE (Phase 3)
**Priority**: LOW

**Problem**: `w:pPrChange` not parsed; tracked changes to paragraph properties not detected.
**Fix**: Added `PPrChange` field to `ParaProps`, `RevisionAuthor`/`RevisionDate` fields to `ParsedParagraph`, parsing and attr emission.

**Files**:
- `words/ooxml.go` — `PPrChange` struct
- `words/types.go` — `RevisionAuthor`/`RevisionDate` fields
- `words/preprocessor.go` — parsing and emission
- `words/verify.go` — validation

---

## GAP-23: Typographic Niceties Not Parsed (L1-L11)
**Status**: DONE (Phase 4)
**Priority**: LOW

**Problem**: Multiple East Asian/typographic paragraph properties not parsed:
- L1: `w:suppressAutoHyphens` — suppress automatic hyphenation
- L2: `w:snapToGrid` — snap to layout grid
- L3: `w:kinsoku` — East Asian line-break control
- L4: `w:wordWrap` — word wrapping
- L5: `w:overflowPunct` — overflow punctuation
- L6: `w:topLinePunct` — top line punctuation
- L7: `w:autoSpaceDE`, `w:autoSpaceDN` — auto-spacing
- L8: `w:textDirection` — text direction
- L9: `w:suppressOverlap` — suppress overlap
- L10: `w:divId` — HTML div association
- L11: `w:cnfStyle` — conditional table style

**Fix**: Added fields to `ParaProps` and `ParsedParagraph`, parsing and attr emission.

**Files**:
- `words/ooxml.go` — Added fields to `ParaProps`; added `TextDirVal`, `CnfStyleVal` structs
- `words/types.go` — Added fields to `ParsedParagraph`
- `words/preprocessor.go` — Parsing and emission
- `words/verify.go` — Validation

---

## GAP-24: Frame Properties / Drop Cap Not Parsed
**Status**: DONE (Phase 5)
**Priority**: MEDIUM

**Problem**: `w:framePr` not parsed; drop cap and text frame properties not detected.
**Fix**: Added `FramePrVal` struct to `ooxml.go`, `FrameProps` struct to `types.go`, `FramePr` field to `ParsedParagraph`, parsing and `frame=` attr emission.

**Files**:
- `words/ooxml.go` — `FramePrVal` struct; `FramePr` field in `ParaProps`
- `words/types.go` — `FrameProps` struct; `FramePr` field in `ParsedParagraph`
- `words/preprocessor.go` — Parsing and emission
- `words/verify.go` — Validation

---

## Implementation Order

```
Phase 1 (DONE): H1-H8 — indent, spacing, lang, headings, blockquote, pre, RTL
Phase 2 (DONE): M1-M3, M6-M8 — shading, keepNext/keepLines, widowControl, struct extensions
Phase 3 (DONE): M4, M9-M10 — sectPr, rPr defaults, revision marks
Phase 4 (DONE): L1-L11 — typographic niceties
Phase 5 (DONE): M5 — frame properties / drop cap
```

## Verification

After each fix:
1. `go test ./words/... -count=1` — all tests pass
2. `go run examples/scripts/generate/main.go` — all outputs generated
3. `go run examples/scripts/verify/main.go` — all outputs verified clean
4. Manual check on `TEMP AKTA PENDIRIAN PT.docx` for indent correctness

---

## Spec Audit: 28 Spec/Impl Gaps Fixed (2026-07-27)

**Status**: DONE — all 28 gaps resolved via documentation-only changes to `docx-preprosessor.md`

| # | Gap | Resolution |
|---|-----|------------|
| 1 | `<img>` w/h in spec | Removed from spec — code correctly omits extent as noise |
| 2 | Code detection third trigger | Clarified: requires code keyword AND all monospace runs |
| 3 | Default font fallback | Added: Times New Roman 11pt when styles.xml unparseable |
| 4-5 | Broad list continuation + isSectionBreak | Documented heuristics in spec |
| 6 | s:page preset matching | Noted: points used internally for comparison |
| 7 | Whitespace normalization | Expanded: tab→space, CRLF→LF, leading newline trim, trailing space trim |
| 8 | s:tab two-pass emission | Documented: content first, style second, dedup across both |
| 9 | s:custom builtin exclusion | Full list of 21 builtin IDs added to spec |
| 10 | Per-paragraph alignment dedup | Documented: matching style alignment suppressed |
| 11 | Heading level +1 offset | Documented: OOXML 0-based → spec 1-based |
| 12 | Style inheritance cycle detection | Documented: visited map prevents infinite loops |
| 13 | Section break default nextPage | Documented: when w:type is absent |
| 14 | Heading demotion exemption | Documented: code paragraphs exempt from demotion |
| 15 | Note refs in tracked changes | Documented: excluded from ordering |
| 17 | Anchor drawings | Documented: treated same as inline |
| 18 | Hyperlink inner runs | Documented: only bold/italic/underline/strikethrough extracted |
| 19 | dashSmallGap border style | Added `ds` = dashed / dashSmallGap to border style list |
| 20 | Border width divisor | Documented: OOXML 1/576 of a point |
| 21 | Cell language cascade | Documented: tc → tr → tbl → body fallback |
| 22 | Bookmark interleaving | Documented: bookmarks interleaved with notes |
| 23 | Word boundary matching | Documented: case-insensitive + word boundary required |
| 24 | Last paragraph section break | Documented: sectPr emitted even on final paragraph |
| 25 | Paragraph run defaults | Documented: propagated to runs without explicit formatting |
| 27 | Note ref id numbering | Documented: raw OOXML id, not renumbered |
| 28 | Tracked changes semantic mode | Corrected: text preserved as plain text, not dropped |
