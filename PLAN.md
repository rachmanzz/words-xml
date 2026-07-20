# PLAN — words-xml compiler

**Status:** Draft v0.1
**Date:** 2026-07-20
**Language:** Go

---

## 1. Summary

A Go compiler that transforms Microsoft Word (`.docx`) OOXML into **words-XML** (v1.0.1) —
a compact, deterministic, LLM-friendly XML representation.

Spec: [`docx-preprosessor.md`](https://github.com/rachmanzz/docx-preprocessor/blob/main/docx-preprosessor.md)
Reference implementation: [`dcdtunning/backend/internal/preprocess/docx.go`](../dcdtunning/backend/internal/preprocess/docx.go)

---

## 2. Goals

1. Build a CLI tool (and importable Go library) that reads `.docx` and emits `words` XML.
2. Strictly follow the `words` v1.0.1 spec — output must be valid, well-formed, and idempotent.
3. Support two modes: `semantic` (default) and `lossless`.
4. Zero LLM dependencies — transformation is purely deterministic.

---

## 3. Pipeline Architecture

```
.docx file (ZIP archive)
    │
    ▼
[1] Open ZIP ── archive/zip
    │
    ▼
[2] Extract named parts:
    ├── word/document.xml
    ├── word/styles.xml
    ├── word/numbering.xml
    ├── word/footnotes.xml
    ├── word/endnotes.xml
    ├── word/_rels/document.xml.rels
    ├── docProps/core.xml
    ├── word/theme/theme1.xml
    └── word/header*.xml / word/footer*.xml
    │
    ▼
[3] Unmarshal each part into Go structs ── encoding/xml
    │
    ▼
[4] Build support maps:
    ├── buildStyleMap()       → map[styleID]StyleDef
    ├── buildNumberingMap()   → map[numID_ilvl]numFmt
    ├── extractTheme()        → ThemeData (font, fg, bg)
    └── extractRels()         → map[rID]target (for hyperlinks)
    │
    ▼
[5] Parse content:
    ├── parseContentItems()   → []ContentItem
    │   ├── parseParagraph()  → {Type:"paragraph", Paragraph:*ParsedParagraph}
    │   │   └── extractRuns() → []TextRun
    │   └── parseTable()      → *ParsedTable
    ├── parseFootnotes()      → []NoteItem
    ├── parseEndnotes()       → []NoteItem
    └── parseBookmarks()      → []NoteItem
    │
    ▼
[6] Emit output XML ── strings.Builder + fmt.Fprintf
    │
    ▼
    words-XML (v1.0.1)
```

---

## 4. Milestones

### Milestone 1 — Foundation
- [ ] Initialize Go module (`go.mod`, package structure)
- [ ] Open `.docx` as ZIP, extract required parts
- [ ] Define all OOXML structs (p, r, tbl, sectPr, etc.)
- [ ] Parse `document.xml` → raw tree

### Milestone 2 — Style & Numbering
- [ ] Parse `styles.xml` → `buildStyleMap()`
- [ ] Parse `numbering.xml` → `buildNumberingMap()`
- [ ] Resolve style inheritance (`w:basedOn` chain)
- [ ] Infer heading level from style name

### Milestone 3 — Block Elements
- [ ] `parseParagraph()` → heading, paragraph, blockquote, code block
- [ ] `parseTable()` → table with colspan/rowspan
- [ ] List grouping → `<ul>` / `<ol>` with nesting
- [ ] List restart detection (`w:lvlOverride`)

### Milestone 4 — Inline Formatting
- [ ] `extractRuns()` → bold, italic, underline, strikethrough
- [ ] `<span>` — font, size, color, highlight, hidden
- [ ] `<sup>`, `<sub>`, `<smallcaps>`, `<uppercase>`
- [ ] `<tab/>`, `<br type="..."/>`
- [ ] `<a href>` — resolve hyperlinks from `r:id` and `instrText HYPERLINK`

### Milestone 5 — Notes & Metadata
- [ ] Parse footnotes/endnotes → `<notes>` block
- [ ] Bookmark → `<bm id="..."/>` in `<notes>`
- [ ] Comment → `<comment>` in `<notes>`
- [ ] `<fn-ref>` marker in `<write>`
- [ ] `<meta>` from `docProps/core.xml`

### Milestone 6 — Style Output
- [ ] `<style>` block — `<s:page>`, `<s:gap>`, `<s:indent>`, `<s:align>`, `<s:line>`
- [ ] `<s:col>` — column widths per table
- [ ] `<s:tab>` — tab stop definitions
- [ ] `<s:theme>` — global font defaults
- [ ] `<s:custom>` — custom style definitions
- [ ] `<header>` / `<footer>` per section

### Milestone 7 — Textbox & Special
- [ ] `w:txbxContent` — unwrap textbox text into `<write>`
- [ ] `w:drawing` — `<img alt="..."/>` placeholder
- [ ] `w:pict` (VML) — drop
- [ ] Bidi / RTL → `dir="rtl"`

### Milestone 8 — Lossless Mode
- [ ] `<ins>` / `<del>` for tracked changes
- [ ] Whitespace preservation (`xml:space="preserve"`)
- [ ] `mode="lossless"` flag on root element

### Milestone 9 — CLI & Finishing
- [ ] CLI interface (`cmd/words-xml/main.go`)
- [ ] Flags: `--mode semantic|lossless`, `--output`, `--validate`
- [ ] Token counting (optional, tiktoken-go)
- [ ] XML escaping & forbidden character stripping
- [ ] Unit test coverage ≥ 80%

---

## 5. Package Layout

```
words-xml/
├── cmd/
│   └── words-xml/
│       └── main.go              # CLI entry point
├── internal/
│   ├── ooxml/                   # OOXML struct definitions
│   │   ├── document.go          # w:document, w:body
│   │   ├── paragraph.go         # w:p, w:pPr, w:r
│   │   ├── table.go             # w:tbl, w:tr, w:tc
│   │   ├── style.go             # w:style, w:styles
│   │   ├── numbering.go         # w:numbering, w:abstractNum
│   │   ├── section.go           # w:sectPr, w:pgSz, w:pgMar
│   │   ├── notes.go             # w:footnote, w:endnote
│   │   ├── rels.go              # w:document.xml.rels
│   │   ├── theme.go             # w:theme
│   │   ├── drawing.go           # w:drawing, w:txbxContent
│   │   └── coreprops.go         # docProps/core.xml
│   ├── preprocessor/            # Core pipeline
│   │   ├── preprocessor.go      # ProcessDOCXFile, ProcessDOCXBytes
│   │   ├── style_resolver.go    # resolveStyle(), inferHeadingLevel()
│   │   ├── numbering_resolver.go # resolveNumbering()
│   │   ├── run_processor.go     # extractRuns()
│   │   ├── list_grouper.go      # groupListItems()
│   │   ├── table_parser.go      # parseTable(), resolveMerge()
│   │   ├── text_cleanup.go      # xmlEscape(), stripControlChars()
│   │   └── emitter.go           # FormatForLLM() — emit output XML
│   └── types/                   # Domain types
│       ├── document.go          # ParsedDocument
│       ├── paragraph.go         # ParsedParagraph, TextRun
│       ├── table.go             # ParsedTable, ParsedTableCell
│       ├── style.go             # StyleDef, ThemeData
│       └── content.go           # ContentItem, NoteItem
├── go.mod
├── go.sum
├── PLAN.md
├── FEATURES.md
├── DECISION-LOGS.md
└── README.md
```

---

## 6. Core Libraries

| Library | Purpose |
|---------|---------|
| `archive/zip` (stdlib) | Open `.docx` ZIP archive |
| `encoding/xml` (stdlib) | Unmarshal OOXML into Go structs |
| `strings` / `bytes` (stdlib) | `strings.Builder` for XML output generation |
| `regexp` (stdlib) | URL extraction from `instrText`, VML image extraction |
| `strconv` (stdlib) | Numeric conversion (twips → inches) |
| `path/filepath` (stdlib) | File path handling |

**Note:** The reference implementation in dcdtunning uses `tiktoken-go` for token
counting. This is optional and can be added in a later milestone.

---

## 7. References

| Document | Location |
|----------|----------|
| Words spec v1.0.1 | `https://github.com/rachmanzz/docx-preprocessor/blob/main/docx-preprosessor.md` |
| Preprocessor limitations | `https://github.com/rachmanzz/docx-preprocessor/blob/main/docx-preprosessor-limitation.md` |
| DOCX→words gap analysis | `https://github.com/rachmanzz/docx-preprocessor/blob/main/docx-words-gap.md` |
| Reference implementation | `../dcdtunning/backend/internal/preprocess/docx.go` |
