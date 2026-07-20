# Features

Feature tracker for the DOCX → words-XML compiler, ordered by processing pipeline stage.

---

## Implemented

_(No features implemented yet)_

## In Progress

_(No features in progress)_

## Planned

---

### [FEATURE] ZIP Extraction
- **Status**: Planned
- **Description**: Open `.docx` file (ZIP archive) and extract all required parts: `document.xml`, `styles.xml`, `numbering.xml`, `footnotes.xml`, `endnotes.xml`, `document.xml.rels`, `core.xml`, `theme1.xml`, `header*.xml`, `footer*.xml`
- **Package**: `internal/ooxml/`
- **Library**: `archive/zip` (stdlib)
- **Notes**: File is read entirely into memory as `[]byte`, no streaming.

---

### [FEATURE] OOXML Struct Parsing
- **Status**: Planned
- **Description**: Define all Go structs with `xml:` tags to represent OOXML elements: `w:document`, `w:p`, `w:pPr`, `w:r`, `w:rPr`, `w:tbl`, `w:tblPr`, `w:tr`, `w:tc`, `w:sectPr`, `w:pgSz`, `w:pgMar`, `w:numPr`, `w:abstractNum`, `w:num`, `w:drawing`, `w:txbxContent`, etc.
- **Package**: `internal/ooxml/`
- **Library**: `encoding/xml` (stdlib)
- **Notes**: Namespace: `http://schemas.openxmlformats.org/wordprocessingml/2006/main`

---

### [FEATURE] Style Resolution
- **Status**: Planned
- **Description**: Parse `styles.xml` → map `[styleID]StyleDef`. Resolve inheritance chain (`w:basedOn`) to determine semantic role (Heading1-9, Title, Quote, Normal, ListParagraph, Code). Infer heading level from style name.
- **Package**: `internal/preprocessor/`
- **Input**: `styles.xml`, paragraph `w:pStyle`
- **Output**: `map[styleID]StyleDef`, resolved heading level
- **Notes**: Must walk `w:basedOn` recursively until a known style is found.

---

### [FEATURE] Numbering Resolution
- **Status**: Planned
- **Description**: Parse `numbering.xml` → map `(numID, ilvl)` → `numFmt`. Detect restart via `w:lvlOverride/w:startOverride`.
- **Package**: `internal/preprocessor/`
- **Input**: `numbering.xml`
- **Output**: `map[numID_ilvl]numFmt`, `map[numID]map[ilvl]startOverride`
- **Notes**: Critical for list grouping and restart detection.

---

### [FEATURE] Paragraph Processing
- **Status**: Planned
- **Description**: Map `w:p` to target element based on style:
  - Heading1-9 → `<h1>`-`<h9>`
  - Normal → `<p>`
  - Quote/IntenseQuote → `<blockquote>`
  - ListParagraph + numPr → `<li>` (inside `<ul>`/`<ol>`)
  - Code-like style / monospace font → `<pre>`
- **Package**: `internal/preprocessor/`
- **Input**: `[]ParsedParagraph` from `parseContentItems()`
- **Output**: `[]ContentItem`
- **Notes**: Custom styles emitted with `c="..."` + `<s:custom>` in `<style>`.

---

### [FEATURE] Run Extraction (Inline Formatting)
- **Status**: Planned
- **Description**: Process each `w:r` inside a paragraph → `TextRun` with all inline formatting:
  - `<b>`, `<i>`, `<u>`, `<s>` (strikethrough)
  - `<smallcaps>`, `<uppercase>`, `<sup>`, `<sub>`
  - `<span font size color highlight lang hidden fontEA fontCS sizeCS>`
  - `<bcs>`, `<ics>` (Complex Script)
  - `<a href>` — resolved from `r:id` or `instrText HYPERLINK`
  - `<br type="textWrapping|page|column|clear"/>`
  - `<tab/>`
  - `<fn-ref id="..." type="footnote|endnote"/>`
  - `<img alt="..."/>` placeholder
  - `<ins>`, `<del>` (lossless mode only)
- **Package**: `internal/preprocessor/`
- **Notes**: Stateful processor — handles field codes (`fldChar begin/separate/end`), VML pictures, drawings, symbols, tracked changes.

---

### [FEATURE] Table Parsing
- **Status**: Planned
- **Description**: Parse `w:tbl` → `<table>` with:
  - `<s:col ref="n" w="..."/>` from `w:tblGrid`/`w:gridCol`
  - `<th>` from `w:trPr/w:tblHeader` flag (not position)
  - `colspan` from `w:gridSpan`
  - `rowspan` from `w:vMerge` (grid reconstruction — restart/continue)
  - `at="..."` for borders
  - `width`, `align`, `indent`, `cellSpacing`, `caption`, `summary`
- **Package**: `internal/preprocessor/`
- **Notes**: Recursive nested table handling. Table IDs are 1-based pre-order traversal.

---

### [FEATURE] List Grouping
- **Status**: Planned
- **Description**: Group consecutive `ListParagraph` paragraphs into `<ul>`/`<ol>`. Nesting via child `<ul>`/`<ol>` inside `<li>`. Detect restart to split into `<ol start="n">`.
- **Package**: `internal/preprocessor/`
- **Input**: `[]ContentItem` (paragraphs with `numPr`)
- **Output**: `[]ContentItem` with grouped list items
- **Notes**: Consider `numId`, `ilvl`, `abstractNumId`, and restart state.

---

### [FEATURE] Textbox Content Extraction
- **Status**: Planned
- **Description**: Unwrap `w:txbxContent` — paragraphs/runs/tables inside a textbox are processed by normal rules. Only text is kept.
- **Package**: `internal/preprocessor/`
- **Notes**: CRIT-1 from spec — textboxes are NOT excluded. Images inside textboxes remain excluded.

---

### [FEATURE] Style Block Emission
- **Status**: Planned
- **Description**: Generate `<style unit="in">` with:
  - `<s:page>` — page size + margins from `w:sectPr`
  - `<s:gap>` — spacing rules per element
  - `<s:line>` — line spacing
  - `<s:indent>` — paragraph indentation
  - `<s:align>` — paragraph alignment
  - `<s:cols>` — multi-column layout
  - `<s:col>` — column widths per table
  - `<s:tab>` — tab stop definitions
  - `<s:theme>` — global font defaults
  - `<s:custom>` — custom style definitions
- **Package**: `internal/preprocessor/`
- **Notes**: `<style>` is required. Must contain at least `<s:page>`.

---

### [FEATURE] Notes Block Emission
- **Status**: Planned
- **Description**: Generate `<notes>` block after `</write>`:
  - `<fn id="n" type="footnote|endnote">` — body text
  - `<bm id="name"/>` — bookmark positions
  - `<comment id="n" author="..." date="...">` — comment text
- **Package**: `internal/preprocessor/`
- **Notes**: Footnote/endnote body from `word/footnotes.xml` / `word/endnotes.xml`.

---

### [FEATURE] Header/Footer Emission
- **Status**: Planned
- **Description**: Parse `w:hdrReference`/`w:ftrReference` → `<header id="n">` / `<footer id="n">`. Content processed with the same transformation rules as `<write>`.
- **Package**: `internal/preprocessor/`
- **Notes**: One per section. Omitted if empty.

---

### [FEATURE] Metadata Emission
- **Status**: Planned
- **Description**: Extract document metadata from `docProps/core.xml` → `<meta>` block:
  - `<title>`, `<author>`, `<created>`, `<modified>`, `<keywords>`
- **Package**: `internal/preprocessor/`
- **Notes**: `<meta>` is optional — omitted if absent or empty.

---

### [FEATURE] Text Cleanup
- **Status**: Planned
- **Description**: XML escaping (`&`, `<`, `>`, `"`), strip forbidden XML 1.0 control characters (0x00-0x08, 0x0B-0x0C, 0x0E-0x1F, 0x7F-0x84). Whitespace normalization in semantic mode (collapse repeated spaces). `<pre>` content is exempt from normalization.
- **Package**: `internal/preprocessor/`
- **Notes**: `xml:space="preserve"` on `<w:t>` must be honored.

---

### [FEATURE] Dual Mode (Semantic & Lossless)
- **Status**: Planned
- **Description**: Support two processing modes:
  - `mode="semantic"` (default) — whitespace normalized, tracked changes dropped
  - `mode="lossless"` — whitespace preserved, tracked changes emitted as `<ins>`/`<del>`
- **Package**: `internal/preprocessor/`
- **Notes**: CLI flag: `--mode semantic|lossless`

---

### [FEATURE] CLI Interface
- **Status**: Planned
- **Description**: Command-line tool to run the compiler:
  - `words-xml <input.docx>` — output to stdout
  - `--output <file>` — write to file
  - `--mode semantic|lossless` — select mode
  - `--validate` — validate output XML
- **Package**: `cmd/words-xml/`

---

### [FEATURE] Image Placeholder
- **Status**: Planned
- **Description**: `w:drawing` image blip → `<img alt="..."/>` placeholder. No pixel/vector data extracted.
- **Package**: `internal/preprocessor/`
- **Notes**: VML (`w:pict`) is dropped. Images are EXCLUDE by policy.

---

### [FEATURE] Bidi / RTL Support
- **Status**: Planned
- **Description**: Paragraph `w:bidi`, run `w:rPr/w:rtl`, inline `w:dir`/`w:bdo` → `dir="rtl"` on the affected element.
- **Package**: `internal/preprocessor/`
- **Notes**: MOD-7 from spec.

---

---

### Template (copy per feature)

```markdown
#### `[FEATURE] <feature name>`
- **Status**: Planned / In Progress / Done
- **Description**: Short summary of what the feature does
- **Package**: Go package path
- **Input**: ...
- **Output**: ...
- **Notes**: Edge cases, dependencies, open questions
- **Related**: Link to spec or decision log
```
