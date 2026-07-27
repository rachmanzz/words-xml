# Plan: Paragraph Feature Coverage

## Current State

The `words/` package converts DOCX paragraphs to words-XML. Core paragraph features
(style, alignment, indentation, tabs, borders, lists, headings 1-3) work correctly.

---

## Gap Analysis

### Fully Covered (tested)

| Feature | Element/Attr | Status |
|---------|-------------|--------|
| Style reference | `c=` on `<p>` | ✅ |
| Heading levels 1-6 | `<h1>`-`<h6>` | ✅ |
| Alignment | `align=` (center, both, right) | ✅ |
| Indent left | `indentLeft=` | ✅ |
| Indent hanging | `indentHanging=` | ✅ |
| Indent first | `indentFirst=` | ✅ |
| Indent right | `indentRight=` | ✅ |
| Tab stops | `<s:tab>` in style block | ✅ |
| Custom styles | `<s:custom>` | ✅ |
| Lists (bullet + ordered) | `<ul>`, `<ol>`, `<li>` | ✅ |
| Inline formatting | `<span>` with bold, italic, etc. | ✅ |
| Page break before | `<br type="page"/>` | ✅ |
| Per-paragraph spacing | `spacingBefore=`, `spacingAfter=` | ✅ |
| Per-paragraph line spacing | `lineSpacing=`, `lineRule=` | ✅ |
| Paragraph language | `lang=` on `<p>` | ✅ |
| Blockquote | `<blockquote>` | ✅ |
| Pre/code | `<pre>` | ✅ |
| RTL/Bidi | `dir="rtl"` on `<p>` | ✅ |

### Partially Covered

| Feature | Gap | Status |
|---------|-----|--------|
| `<s:indent>` / `<s:gap>` / `<s:line>` in style block | Code exists, verified | ✅ Tested |
| Indent on headings | `writeParagraphAttrs` handles it | ✅ Tested |
| List continuation (GAP-02) | Continuation text included in `<li>` | ✅ Tested |

### Not Covered

None - all features implemented and tested.

---

## Test Document Requirements

To validate uncovered features, we need DOCX files containing:

| Feature Needed | Example Content | Status |
|----------------|-----------------|--------|
| Heading levels 4-6 | Deeply nested document structure | ✅ Tested |
| Blockquotes | Quoted text with citation | ✅ Tested |
| Code blocks | Preformatted/monospace text | ✅ Tested |
| RTL/Bidi text | Arabic, Hebrew, or mixed-direction content | ✅ Tested |
| Paragraph-level language | Paragraphs in different languages | ✅ Tested |
| Right indentation | Paragraphs with right indent set | ✅ Tested |
| Paragraph spacing overrides | Paragraphs with spacing different from style | ✅ Tested |
| Line spacing overrides | Paragraphs with custom line spacing | ✅ Tested |
| Paragraph shading | Highlighted/callout paragraphs | ✅ Tested |
| Keep next/lines | Headings with keep-next, long paragraphs with keep-lines | ✅ Tested |
| Indent on headings | Headings with custom indentation | ✅ Tested |
| List continuation | Non-list paragraphs between list items | ✅ Tested |

---

## Implementation Order

### Phase 1 (Completed)
1. ~~**H6**: Add `indentRight=` to `writeParagraphAttrs`~~ ✅
2. ~~**H7-H8**: Add per-paragraph spacing/lineSpacing attrs~~ ✅
3. ~~**H5**: Add `lang=` on `<p>` when paragraph-level~~ ✅
4. ~~**H1-H4**: Verify existing code works; create test documents~~ ✅

### Phase 2 (Completed)
5. ~~**M1**: Add `w:shd` parsing and `shd=` attr emission~~ ✅
6. ~~**M2-M3**: Add `keepNext`/`keepLines` attrs~~ ✅
7. ~~**M6-M7**: Extend `PBdrProps` and `IndVal` structs~~ ✅
8. ~~**M8**: Add `widowControl` attr~~ ✅

### Phase 3 (Completed)
9. ~~**M4**: Handle `w:sectPr` in paragraphs~~ ✅
10. ~~**M9**: Add paragraph-level default run properties~~ ✅
11. ~~**M10**: Add revision marks on paragraph props~~ ✅

### Phase 4 (Completed)
12. ~~**L1**: Add `suppressAutoHyph` attr~~ ✅
13. ~~**L2**: Add `snapToGrid` attr~~ ✅
14. ~~**L3**: Add `kinsoku` attr~~ ✅
15. ~~**L4**: Add `wordWrap` attr~~ ✅
16. ~~**L5-L6**: Add `overflowPunct`/`topLinePunct` attrs~~ ✅
17. ~~**L7**: Add `autoSpaceDE`/`autoSpaceDN` attrs~~ ✅
18. ~~**L8**: Add `textDirection` attr~~ ✅
19. ~~**L9**: Add `suppressOverlap` attr~~ ✅
20. ~~**L10**: Add `divID` attr~~ ✅
21. ~~**L11**: Add `cnfStyle` attr~~ ✅

### Phase 5 (Completed)
22. ~~**M5**: Add `frame` attr (drop cap / text frame)~~ ✅

### Phase 6 (Completed)
23. ~~**GAP-02**: Fix list continuation across non-list paragraphs~~ ✅
24. ~~**GAP-05**: Fix heading indent~~ ✅
25. ~~**Test coverage**: Add test cases for shading, keepNext/keepLines, heading indent, list continuation~~ ✅

---

## Files Modified (Phase 2)

| File | Changes |
|------|---------|
| `words/ooxml.go` | Added `ShdVal` struct, `Shd`/`KeepNext`/`KeepLines`/`WidowControl` to `ParaProps`, `Between`/`Bar` to `PBdrProps`, char-unit fields to `IndVal` |
| `words/types.go` | Added `Shading`/`KeepNext`/`KeepLines`/`WidowControl` fields to `ParsedParagraph` |
| `words/preprocessor.go` | Parse new OOXML elements; emit `shd`/`keepNext`/`keepLines`/`widowControl` attrs |
| `words/verify.go` | Added validation for new attrs |
| `examples/scripts/create-test-docx/main.go` | Added test cases for shading, keepNext, keepLines, widowControl |
| `GAP.md` | Updated with GAP-15 through GAP-19 (all DONE) |

## Files Modified (Phase 3)

| File | Changes |
|------|---------|
| `words/ooxml.go` | Added `RPr`/`SectPr`/`PPrChange` to `ParaProps`; added `SectPrProps`, `SectType`, `PgSzVal`, `PgMarVal`, `CTRel`, `PPrChange` structs |
| `words/types.go` | Added `ParaDefaults`, `SectionBreak`, `RevisionAuthor`, `RevisionDate` fields to `ParsedParagraph` |
| `words/preprocessor.go` | Added `applyParaRunDefaults` function; modified `extractRuns` to accept paragraph defaults; parse `sectPr`/`pPrChange`; emit `sectionBreak`/`revisionAuthor`/`revisionDate` attrs |
| `words/verify.go` | Added validation for `sectionBreak`, `revisionAuthor`, `revisionDate` attrs |
| `GAP.md` | Updated with GAP-20 through GAP-22 (all DONE) |

## Files Modified (Phase 4)

| File | Changes |
|------|---------|
| `words/ooxml.go` | Added L1-L11 fields to `ParaProps`; added `TextDirVal`, `CnfStyleVal` structs |
| `words/types.go` | Added L1-L11 fields to `ParsedParagraph` |
| `words/preprocessor.go` | Parsing and emission for L1-L11 attrs |
| `words/verify.go` | Added validation for L1-L11 attrs |
| `GAP.md` | Updated with GAP-23 (DONE) |

## Files Modified (Phase 5)

| File | Changes |
|------|---------|
| `words/ooxml.go` | Added `FramePrVal` struct; added `FramePr` field to `ParaProps` |
| `words/types.go` | Added `FrameProps` struct; added `FramePr` field to `ParsedParagraph` |
| `words/preprocessor.go` | Parsing and emission for `frame=` attr |
| `words/verify.go` | Added validation for `frame` attr |
| `GAP.md` | Updated with GAP-24 (DONE) |

## Files Modified (Phase 6)

| File | Changes |
|------|---------|
| `examples/scripts/create-test-docx/main.go` | Added test cases for heading indent, list continuation |
| `GAP.md` | Updated GAP-02 (FIXED), GAP-05 (FIXED), test requirements |
