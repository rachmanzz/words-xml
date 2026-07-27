# DOCX Preprocessor

This document specifies how raw Microsoft Word (`.docx`) OOXML is preprocessed into a
compact, LLM-friendly intermediate markup called **`words`**.

The goal of the preprocessor is to strip Word's verbose, presentation-oriented XML down
to a semantic, token-efficient representation that is easy to parse and validate.

---

## 1. Scope and Mode

This specification defines the **DOCX Preprocessor**, which transforms raw Microsoft Word (`.docx`) OOXML into a compact, LLM-friendly intermediate markup called **`words`**.

The preprocessor operates in one of two modes:

- `mode="semantic"` (default): stripped-down representation for **AI training** and **downstream consumption**.
- `mode="lossless"`: preserves additional metadata for **round‑tripping** or **document‑reconstruction**. Differences from semantic mode:
  - Whitespace is NOT normalized (original spacing preserved).
  - Tracked changes (`w:ins`/`w:del`) are emitted as `<ins>`/`<del>` elements.
  - All other transformation rules remain the same.

### 1.1 Problem (semantic mode)

Raw `.docx` body XML is noisy and redundant for LLM consumption. A single heading looks like:

```xml
<w:p w14:paraId="7A3B2" w:rsidR="007X">
  <w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
  <w:r><w:t>Specifications for Data Center Racks</w:t></w:r>
</w:p>
```

Issues:
- Presentation attributes (`w14:paraId`, `w:rsidR`, `w:pPr`) carry no semantic value.
- Deep nesting (`w:p` → `w:pPr` → `w:pStyle` → `w:val`) wastes tokens.
- Style info is buried and inconsistent across documents.

### 1.2 Goal (semantic mode)

Emit a flat, semantic, versioned XML (`words`) that is:
- **deterministic** (identical DOCX → identical output),
- **token‑efficient** (no presentation junk),
- **parser‑friendly** (strict XML, well‑formed, schema‑compatible).

---

## 2. Target Format: `words` (semantic mode)

A flat, semantic, versioned XML:

```xml
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.1.0" mode="semantic">
  <style unit="in">
    <s:page size="A4" mt="0.75" mb="0.75" ml="0.75" mr="0.75" mh="0.5" mf="0.5"/>
  </style>
  <write>
    <h1>Specifications for Data Center Racks</h1>
  </write>
</words>
```

### 2.1 Grammar (v1.1.0)

**Namespace declarations**: The root `<words>` element MUST declare two namespaces:
- `xmlns="urn:words:v1"` — default namespace for all elements (`<meta>`, `<style>`, `<write>`, `<notes>`, `<header>`, `<footer>`, `<p>`, `<h1>`-`<h9>`, `<ul>`, `<ol>`, `<table>`, `<pre>`, `<blockquote>`, `<tr>`, `<th>`, `<td>`, `<li>`, `<b>`, `<i>`, `<u>`, `<s>`, `<span>`, `<a>`, `<br>`, etc.)
- `xmlns:s="urn:words:v1:style"` — prefix for style/layout elements (`<s:page>`, `<s:gap>`, etc.)
- Inline elements (`<b>`, `<i>`, `<u>`, `<s>`, `<span>`, `<a>`, `<br>`, etc.) use **no prefix**

```text
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.1.0" mode="semantic">
  <meta>                         # optional document metadata (after root)
    <title>...</title>           # dc:title from docProps/core.xml
    <author>...</author>         # dc:creator
    <created>...</created>       # dcterms:created (ISO 8601)
    <modified>...</modified>     # dcterms:modified
    <keywords>...</keywords>     # cp:keywords
  </meta>
  <style unit="in">             # layout config (required, before <write>)
                                # unit = default unit for all numeric layout values
    <s:page size="A4"           # named preset (A3/A4/A5/Letter/Legal/Tabloid/B5/
                                #   A6/Executive/Statement/Folio); resolves w/h
            w=".." h=".."       # OR explicit page size (overrides size)
            mt=".." mb=".." ml=".." mr=".."   # margins (top/bottom/left/right)
            mh=".." mf=".."/>    # header / footer margins
    # <s:page> MAY repeat once per document section (see MOD-6 / §2.4)
    <s:gap el="h" c="Heading1" before=".." after=".."/>
    <s:gap el="p" before=".." after=".."/>
    <s:line el="p" value="1.5" rule="auto"/>  # line spacing (LOSSLESS_METADATA)
    <s:indent el="p"          # paragraph indentation (MOD-5)
              left=".." right=".." firstLine=".." hanging=".."/>
    <s:align el="p" value="left|center|right|both"/> # paragraph alignment (LOSSLESS_METADATA)
    <s:cols n=".." space=".."/>  # multi-column layout — P19
    <s:col ref="n" w=".."/>     # column widths / grid; ref = table id (1-based index)
    <s:tab el="p|h1" pos="1.0" align="left|center|right|decimal" leader="none|dot|dash|underscore|bar"/>  # tab stop definition
    <s:theme font=".." fontEA=".." fontCS=".." bg=".." fg=".."/>  # optional global defaults (font + color tokens)
    <s:custom name=".." basedOn=".." type="paragraph|character|table"
              font=".." fontEA=".." fontCS=".." size=".." sizeCS=".."
              color=".." bold="true" italic="true" underline=".."
              strikethrough="true" smallCaps="true" uppercase="true"
              alignment=".." spacingBefore=".." spacingAfter=".."
              lineSpacing=".." lineRule="auto|exact|atLeast"
              indentLeft=".." indentRight=".." indentFirst=".." indentHanging=".."
              borderWidth=".." borderColor=".." borderStyle=".."
              cellSpacing=".." width=".."/>  # custom style definition
  </style>
  <header id="n">                # header content (per section, optional)
    <p>...</p>               # same block elements as <write>
  </header>
  <footer id="n">                # footer content (per section, optional)
    <p>...</p>
  </footer>
  <write>                       # one document / one logical write session
                                # any block element may carry dir="rtl"|"ltr" (MOD-7)
                                # and lang=".." (BCP 47 language tag)
    <h1 lang="..">...</h1>
    <h2 lang="..">...</h2>
    <h3 lang="..">...</h3>
    <p lang=".." valign="top|center|baseline">...</p>
    <p at="bb 12 s1 #000000" lang=".." valign="top|center|baseline">...</p>  # paragraph with border
    <p shd="..." keepNext="true" keepLines="true" widowControl="true" spacingBefore=".." spacingAfter=".." lineSpacing=".." lineRule="auto|exact|atLeast" indentLeft=".." indentHanging=".." indentRight=".." indentFirst=".." sectionBreak="nextPage|continuous|evenPage|oddPage" revisionAuthor=".." revisionDate="..">...</p>  # paragraph with extended attributes
    <blockquote lang="..">...</blockquote>
    <b>...</b>              # bold (inline)
    <i>...</i>              # italic (inline)
    <u>...</u>              # underline (inline)
    <s>...</s>              # strikethrough (inline) — CRIT-3
    <smallcaps>...</smallcaps> # small caps (inline) — MOD-4
    <uppercase>...</uppercase> # all caps (inline) — MOD-4
    <sub>...</sub>          # subscript (inline)
    <sup>...</sup>          # superscript (inline)
    <bcs>...</bcs>          # Complex Script bold (inline) — P16
    <ics>...</ics>          # Complex Script italic (inline) — P17
    <span font=".." size=".." color=".." highlight=".." lang=".." hidden="true" fontEA=".." fontCS=".." sizeCS="..">...</span>  # font/style span (inline)
    <a href="...">...</a>   # hyperlink (r:id or instrText HYPERLINK) — MOD-3
    <br type="textWrapping|page|column|clear"/> # line break — MIN-1
    <tab/>                  # tab character (moves to next tab stop)
    <fn-ref id="n" type="footnote|endnote"/>  # marker with type attribute
    <ins>...</ins> / <del>...</del> # tracked change (optional, LOSSLESS_METADATA)
    <ul type="bullet|...">    # unordered list (type from numFmt)
      <li>                    # list item; nesting via child <ul>/<ol>
        ...                     # text + inline elements
        <ul type="...">       # nested sub-list (one level deeper)
          <li>...</li>
        </ul>
      </li>
    </ul>
    <ol type="decimal|lowerLetter|..." start="n">  # ordered list (type from numFmt)
      <li>
        ...
        <ol type="...">       # nested ordered sub-list
          <li>...</li>
        </ol>
      </li>
    </ol>
    <pre>...</pre>        # code / monospace block (whitespace preserved verbatim)
    <table id="n" at="..." c=".." width=".." align="left|center|right" indent=".." cellSpacing=".." caption=".." summary="..">  # table
      <tr><th colspan="n" rowspan="n" lang=".." valign="top|center|bottom" textDir=".." noWrap="true">..</th></tr>
      <tr><td colspan="n" rowspan="n" at="..." lang=".." valign="top|center|bottom" textDir=".." noWrap="true">..</td></tr>
    </table>
    <img alt="..."/>         # PLACEHOLDER ONLY (images excluded — §3.0)
  </write>
  <notes>                         # footnote/endnote bodies, bookmarks, comments
    <fn id="1" type="footnote">This is the footnote text for reference 1.</fn>
    <bm id="bookmark1"/>        # bookmark position
    <comment id="1" author="..." date="...">...</comment>
  </notes>
</words>
```

### 2.2 Minimal Style Block

The minimal required `<style>` block with page size A4 in inches:

```xml
<style unit="in">
  <s:page size="A4" mt="0.75" mb="0.75" ml="0.75" mr="0.75" mh="0.5" mf="0.5"/>
</style>
```

**Unit conversions (pt → in):**
- 54pt = 0.75in (standard 0.75 inch margins)
- 36pt = 0.5in (standard 0.5 inch header/footer margins)

### 2.3 `<meta>` — Document Metadata Block (optional)

The `<meta>` element appears **once, immediately after the root `<words>` open tag and
before `<style>`**. It carries document-level metadata extracted from `docProps/core.xml`
so provenance info is preserved for downstream filtering.

- `<title>` — document title (`dc:title`).
- `<author>` — creator name (`dc:creator`).
- `<created>` — creation timestamp (`dcterms:created`, ISO 8601).
- `<modified>` — last-modified timestamp (`dcterms:modified`, ISO 8601).
- `<keywords>` — tag/keywords (`cp:keywords`).

All `<meta>` children are optional. If `docProps/core.xml` is absent or empty, the
`<meta>` block is omitted entirely.

### 2.4 `<style>` — Layout Block (required)

The `<style>` element is **required** and appears **once, immediately after `<meta>` (or the root `<words>`
open tag if `<meta>` is absent) and before `<write>`**. It carries presentation/layout
metadata extracted from the source document so layout intent is preserved without
polluting the semantic `<write>` body. At minimum, `<style>` MUST contain a `<s:page>` element
with page size and margins.

- `<s:page>` — page geometry (from `w:sectPr` / `pgSz` / `pgMar`):
  - `size` — named preset; resolves `w`/`h` automatically. Allowed: `A3`, `A4`, `A5`, `A6`,
    `B5`, `Letter`, `Legal`, `Tabloid`, `Executive`, `Statement`, `Folio`. If `w`/`h` are also given, they **override** the preset.
  - `w`, `h` — page width/height (explicit; overrides `size`).
  - `mt`, `mb`, `ml`, `mr` — margins mapped from `w:pgMar` `@w:top` / `@w:bottom` /
    `@w:left` / `@w:right`.
  - `mh`, `mf` — header/footer margins from `w:pgMar` `@w:header` / `@w:footer`.

  - **Multiple sections (MOD-6)**: `<s:page>` MAY appear more than once, once per
    document section (each `w:sectPr` in the body). The first `<s:page>` is the default;
    subsequent entries describe later sections (e.g., a landscape section inside a
    portrait document). Renderers use the entry matching the section in question.

  - `<s:indent>` — paragraph indentation (MOD-5), from `w:pPr/w:ind`:
    - `el` — target element (usually `p`).
    - `left`, `right` — left/right indent.
    - `firstLine` — first-line indent (positive) ; `hanging` — hanging indent (positive).
    - Sign convention follows Word: `w:ind/@w:firstLine` positive → `firstLine`;
      `w:ind/@w:hanging` positive → `hanging`.

  - `<s:align>` — paragraph alignment (LOSSLESS_METADATA), from `w:pPr/w:jc`:
    - `el` — target element (usually `p`).
    - `value` — alignment: `left`, `center`, `right`, `both` (justify).
    - Mapped from `w:jc/@w:val`: `left` → `left`, `center` → `center`, `right` → `right`,
      `both` → `both`.

  **Page-size presets** (resolved in the declared `unit`; values shown in `pt`):

  | `size`     | `w` (pt) | `h` (pt) | notes                |
  |------------|----------|----------|----------------------|
  | `A3`       | 842      | 1191     |                      |
  | `A4`       | 595      | 842      | default document size |
  | `A5`       | 420      | 595      |                      |
  | `A6`       | 298      | 420      |                      |
  | `B5`       | 516      | 729      | ISO B5               |
  | `Letter`   | 612      | 792      | 8.5 × 11 in          |
  | `Legal`    | 612      | 1008     | 8.5 × 14 in          |
  | `Tabloid`  | 792      | 1224     | 11 × 17 in (≈ B4)    |
  | `Executive`| 540      | 720      | 7.5 × 10 in          |
  | `Statement`| 396      | 612      | 5.5 × 8.5 in         |
  | `Folio`    | 612      | 936      | 8.5 × 13 in          |

  > **Fallback (MIN-4)**: if `w:pgSz/@w:w`/`@w:h` does not match a known preset, emit
  > explicit `w`/`h` (in the declared `unit`) with **no** `size` attribute. Never
  > silently coerce to the nearest preset.

  > Resolution must use the declared `unit` (e.g., `unit="mm"` → A4 = `w="210" h="297"`).
  > Conversion: `pt ÷ 2.834645669` → `mm`; `pt ÷ 72` → `in`.

### 2.5 Units

All numeric layout values use the **`unit`** declared on `<style>` (default `in`).
Allowed units: `in` (inch, default), `pt` (point), `px` (pixel), `cm`, `mm`.

- A bare number (`mt="54"`) is interpreted in the declared `unit`.
- A value may override the default inline by suffixing its own unit
  (e.g., `ml="2cm"` even when `unit="pt"`).
- Word OOXML stores sizes in **twips** (`1pt = 20 twips`); the preprocessor MUST convert
  twips → the declared unit before emitting.
- `<s:gap>` — spacing rules keyed by element (`el`) and optional style (`c`):
  `before`/`after` gaps in the declared unit. Lets downstream renderers reproduce vertical rhythm.
- `<s:line>` — line spacing keyed by element (`el`) and optional style (`c`):
  `value` = line spacing multiplier (e.g., `1.5` for 1.5 line spacing, `2` for double);
  `rule` = `auto` (proportional), `exact` (fixed), `atLeast` (minimum). From `w:spacing/@w:line`
  and `@w:lineRule`. Example: `<s:line el="p" value="1.5" rule="auto"/>`.
- `<s:col>` — column/grid widths (from `w:tblGrid` / `w:gridCol`). Each `<s:col>` carries a
  `ref="n"` attribute that matches the `<table id="n">` it belongs to (1-based document
  order). Tables without a `w:tblGrid` emit no `<s:col>`.
- `<s:theme>` — optional global defaults (font, fontEA, fontCS, bg, fg) from theme part + docDefaults.
- `<s:custom>` — custom style definition (from `w:style` in `styles.xml`):
  `name` = style name (REQUIRED); `basedOn` = parent style name (optional);
  `type` = paragraph|character|table (optional); formatting properties as attributes
  (font, size, color, bold, italic, alignment, spacing, indentation, borders, etc.).
  Only emitted for custom styles (not standard Heading1-9, Normal, etc.).
  Example: `<s:custom name="MyHeading" basedOn="Heading1" font="Arial" color="FF0000"/>`.

### `c` Attribute — Original Style Name

The `c` attribute preserves the original style name from DOCX for round-tripping.
- **Standard styles** (Heading1-9, Normal, Title, Quote, ListParagraph, etc.): `c` is NOT emitted
  (redundant — element name already implies the style).
- **Custom styles**: `c` IS emitted to preserve the original style name.
- Example: `<h1 c="MyCustomHeading">` for custom style, `<h1>` for standard Heading1.

### Block Element Attributes

All block elements (`<p>`, `<h1>`-`<h9>`, `<li>`, `<blockquote>`, `<pre>`) can carry the following attributes:

| Attribute | Type | Source | Description |
|-----------|------|--------|-------------|
| `lang` | string | `w:pPr/w:pStyle/lang`, `w:lang`, or `w:rPr/w:rLang` | BCP 47 language tag |
| `dir` | string | `w:bidi` or `w:rPr/w:rtl` | Text direction (`rtl` or `ltr`) |
| `c` | string | `w:pPr/w:pStyle` (custom styles only) | Original style name |
| `align` | string | `w:pPr/w:jc` | Paragraph alignment (`left`, `center`, `right`, `both`) |
| `valign` | string | `w:pPr/w:textAlignment` | Vertical text alignment (`top`, `center`, `baseline`) |
| `shd` | string | `w:pPr/w:shd` | Paragraph shading (hex color or pattern) |
| `at` | string | `w:pPr/w:pBdr` | Compact border representation (see §`at` Attribute) |
| `spacingBefore` | float | `w:pPr/w:spacing/@w:before` | Space before paragraph (in declared unit) |
| `spacingAfter` | float | `w:pPr/w:spacing/@w:after` | Space after paragraph (in declared unit) |
| `lineSpacing` | float | `w:pPr/w:spacing/@w:line` | Line spacing (in declared unit or multiplier) |
| `lineRule` | string | `w:pPr/w:spacing/@w:lineRule` | Line spacing rule (`auto`, `exact`, `atLeast`) |
| `indentLeft` | float | `w:pPr/w:ind/@w:left` | Left indent (in declared unit) |
| `indentRight` | float | `w:pPr/w:ind/@w:right` | Right indent (in declared unit) |
| `indentFirst` | float | `w:pPr/w:ind/@w:firstLine` | First-line indent (in declared unit) |
| `indentHanging` | float | `w:pPr/w:ind/@w:hanging` | Hanging indent (in declared unit) |
| `keepNext` | bool | `w:pPr/w:keepNext` | Keep paragraph with next paragraph |
| `keepLines` | bool | `w:pPr/w:keepLines` | Keep all lines of paragraph together |
| `widowControl` | bool | `w:pPr/w:widowControl` | Allow widow/orphan lines |
| `sectionBreak` | string | `w:pPr/w:sectPr/w:type/@w:val` | Section break type (`nextPage`, `continuous`, `evenPage`, `oddPage`) |
| `revisionAuthor` | string | `w:pPr/w:pPrChange/@w:author` | Tracked change author name |
| `revisionDate` | string | `w:pPr/w:pPrChange/@w:date` | Tracked change date (ISO 8601) |
| `suppressAutoHyph` | bool | `w:pPr/w:suppressAutoHyphens` | Suppress automatic hyphenation |
| `snapToGrid` | bool | `w:pPr/w:snapToGrid` | Snap to document grid |
| `kinsoku` | bool | `w:pPr/w:kinsoku` | East Asian line-break control |
| `wordWrap` | bool | `w:pPr/w:wordWrap` | Enable word wrapping |
| `overflowPunct` | bool | `w:pPr/w:overflowPunct` | Allow punctuation to overflow margin |
| `topLinePunct` | bool | `w:pPr/w:topLinePunct` | Compress punctuation at top of line |
| `autoSpaceDE` | bool | `w:pPr/w:autoSpaceDE` | Auto-space between D/E (Latin/East Asian) |
| `autoSpaceDN` | bool | `w:pPr/w:autoSpaceDN` | Auto-space between D/N (numbers/East Asian) |
| `textDirection` | string | `w:pPr/w:textDirection` | Text direction (`lr`, `rl`, `lrV`, `rlV`) |
| `suppressOverlap` | bool | `w:pPr/w:suppressOverlap` | Suppress floating object overlap |
| `divID` | int | `w:pPr/w:divId` | HTML div association ID |
| `cnfStyle` | string | `w:pPr/w:cnfStyle` | Conditional table style formatting |
| `frame` | compound | `w:pPr/w:framePr` | Frame/drop-cap properties (see below) |

**Frame Properties (`frame` attribute):**

The `frame` attribute is a compound attribute containing frame/drop-cap properties. Format:
```
frame="dropCap='drop' lines='3' width='1440' wrap='around'"
```

Sub-attributes:
- `dropCap` — drop cap style (`drop`, `margin`, `none`)
- `lines` — number of lines for drop cap
- `width` — frame width (in twips, converted to declared unit)
- `height` — frame height
- `vSpace` — vertical space
- `hSpace` — horizontal space
- `wrap` — text wrapping (`around`, `none`, `through`, `topAndBottom`)
- `hAnchor` — horizontal anchor (`text`, `page`, `margin`, `column`)
- `vAnchor` — vertical anchor (`text`, `page`, `paragraph`, `margin`)
- `x` — x-position
- `xAlign` — x-alignment (`left`, `center`, `right`, `inside`, `outside`)
- `y` — y-position
- `yAlign` — y-alignment (`top`, `center`, `bottom`, `inside`, `outside`)
- `hRule` — height rule (`atLeast`, `exact`, `auto`)
- `anchorLock` — anchor lock (boolean)

**Precedence**: Per-paragraph attributes override style-level defaults (`<s:gap>`, `<s:line>`, `<s:indent>`, `<s:align>`). When a per-paragraph attribute is present, it takes precedence over the corresponding style-level value.

### `at` Attribute — Compact Border Representation

The `at` attribute provides a compact syntax for borders on block elements and table cells.
Format: `at="[side] [width] [style][space] [color]; ..."` where:
- `side`: `bt` (top), `bb` (bottom), `bl` (left), `br` (right)
- `width`: border width in the declared unit
- `style`: `s` (single), `d` (double), `ds` (dashed), `dt` (dotted), `n` (none)
- `space`: spacing value (appended to style code, e.g., `s1` = single, space 1)
- `color`: hex color (e.g., `#000000`)
- Multiple borders separated by `;`

Examples:
```xml
<p at="bb 12 s1 #000000"/>                     <!-- bottom border only -->
<p at="bt 8 d2 #FF0000; bb 4 s1 #000000"/>     <!-- top double + bottom single -->
<td at="bb 4 s1 #000000"/>                        <!-- cell bottom border -->
<table at="bb 4 s1 #000000"/>                     <!-- table default border -->
```

The `<style>` block is **required**: it MUST appear in every `words` document with at minimum a `<s:page>` element specifying page size and margins. The `<write>` body depends on `<style>` for layout context.

### 2.6 `<header>` / `<footer>` — Header & Footer Content

Each `<header>` or `<footer>` element appears **after `<style>` and before `<write>`**,
one per document section (matching `<s:page>` entries). They carry the text content
extracted from the corresponding `w:hdrReference`/`w:ftrReference` parts.

- `<header id="n">` — header content for section `n` (1-based, matches `<s:page>` order).
- `<footer id="n">` — footer content for section `n`.
- Content inside `<header>`/`<footer>` uses the same block elements as `<write>`
  (`<p>`, `<h1>`-`<h9>`, `<table>`, etc.) processed through the full transformation rules.
- If a header/footer part is empty or missing, the corresponding element is omitted.
- Headers/footers are **NOT** excluded — only their presentation chrome is dropped.

### 2.7 `<notes>` — Notes Container

The `<notes>` element appears **once, immediately after the closing `</write>` tag and
before the root `</words>`**. It carries footnote/endnote bodies, bookmarks, and comments.

- `<fn id="n" type="footnote|endnote">` — body with matching type attribute
  The marker in `<write>` is an empty element `<fn-ref id="n" type="footnote|endnote"/>`; the body lives here.
- `<bm id="name"/>` — bookmark position marker (self-closing, `id` = bookmark name from `w:bookmarkStart/@w:name`).
- `<comment id="n" author="..." date="...">text</comment>` — comment text with author and date metadata.
  `id` is a 1-based index; `author` from `w:comment/@w:author`; `date` from `w:comment/@w:date` (ISO 8601).
- If the footnote/endnote has no text content, `<fn id="n" type="footnote|endnote"/>` is self-closing (marker-only).
- The text body is extracted from `word/footnotes.xml` or `word/endnotes.xml` in the `.docx`
  package, processed through the same paragraph/run transformation rules as the main body,
  but wrapped in the footnote container.
- Bookmarks and comments are placed in document order within the `<notes>` block.
- Footnotes, endnotes, bookmarks, and comments are all placed in document order within a single `<notes>` block.

---

## 3. Transformation Rules

### 3.0 DOCX Feature Coverage & Noise Matrix

Every OOXML construct the preprocessor encounters is classified into one of four categories:

- **KEEP** — mapped to a semantic `words` element (e.g., text, headings, bold, tables).
- **LOSSLESS_METADATA** — presentation/layout info that does not affect semantic meaning
  but is preserved as non-lossy metadata in `<style>` or as attributes (e.g., alignment,
  tracked changes). These are useful for round-tripping or AI tasks that care about layout.
- **DROP** — presentation noise or renderer hints safely removed (e.g., shading, tabs,
  page break before, frame properties).
- **EXCLUDE** — out of scope / too complex / binary (e.g., images, OLE, Math).

| DOCX element | Category | Action / Target | Rationale |
|--------------|----------|-----------------|-----------|
| `w:body` | container | unwrap | structural only |
| `w:p` | struct | `<h1>`-`<h9>`/`<p>`/`<li>`/`<blockquote>` | semantic block |
| `w:pPr/w:pStyle` | style | `c="..."` attr (only when style name ≠ element name) + `<s:custom>` in `<style>` | keep style name + custom style definition |
| `w:pPr/w:numPr` | list | drives `<ul>`/`<ol>` | list structure |
| `w:pPr/w:spacing` | layout | `<s:gap before/after>` + `<s:line>` | vertical rhythm + line spacing |
| `w:pPr/w:ind` | layout | `<s:indent>` (in `<style>`) | indentation preserved (MOD-5) |
| `w:pPr/w:jc` | layout | `<s:align>` in `<style>` | justification preserved as LOSSLESS_METADATA |
| `w:pPr/w:textAlignment` | keep | `valign="top|center|baseline"` on `<p>` | vertical text alignment |
| `w:bidi` (p), `w:rPr/w:rtl` (r), `w:dir`/`w:bdo` | direction | `dir="rtl"` attribute on element | RTL/bidi support (MOD-7) |
| `w:pPr/w:outlineLvl` | style | DROP (inferred from heading) | redundant |
| `w:pPr/w:suppressLineNumbers` | misc | DROP | renderer hint |
| `w:pPr/w:keepNext` | layout | `keepNext="true"` on `<p>` | paragraph formatting |
| `w:pPr/w:keepLines` | layout | `keepLines="true"` on `<p>` | paragraph formatting |
| `w:pPr/w:widowControl` | layout | `widowControl="true"` on `<p>` | widow/orphan control |
| `w:pPr/w:shd` | layout | `shd="..."` on `<p>` | paragraph shading preserved |
| `w:pPr/w:sectPr` (inline) | layout | `sectionBreak="..."` on `<p>` | section break type |
| `w:pPr/w:pPrChange` | revision | `revisionAuthor="..."` `revisionDate="..."` on `<p>` | tracked change metadata |
| `w:pPr/w:suppressAutoHyphens` | layout | `suppressAutoHyph="true"` on `<p>` | suppress auto hyphens |
| `w:pPr/w:snapToGrid` | layout | `snapToGrid="true"` on `<p>` | snap to grid |
| `w:pPr/w:kinsoku` | layout | `kinsoku="true"` on `<p>` | East Asian line-break control |
| `w:pPr/w:wordWrap` | layout | `wordWrap="true"` on `<p>` | word wrapping |
| `w:pPr/w:overflowPunct` | layout | `overflowPunct="true"` on `<p>` | overflow punctuation |
| `w:pPr/w:topLinePunct` | layout | `topLinePunct="true"` on `<p>` | top line punctuation |
| `w:pPr/w:autoSpaceDE` | layout | `autoSpaceDE="true"` on `<p>` | auto-space D/E |
| `w:pPr/w:autoSpaceDN` | layout | `autoSpaceDN="true"` on `<p>` | auto-space D/N |
| `w:pPr/w:textDirection` | layout | `textDirection="..."` on `<p>` | text direction |
| `w:pPr/w:suppressOverlap` | layout | `suppressOverlap="true"` on `<p>` | suppress overlap |
| `w:pPr/w:divId` | layout | `divID="..."` on `<p>` | HTML div association |
| `w:pPr/w:cnfStyle` | layout | `cnfStyle="..."` on `<p>` | conditional table style |
| `w:pPr/w:framePr` | layout | `frame="..."` on `<p>` | frame/drop-cap properties |
| `w:pPr/w:pBdr` | present | `at="bb ..."` on `<p>` | paragraph borders preserved compact |
| `w:pPr/w:tabs` | tab stops | `<s:tab el=".." pos=".." align=".." leader=".."/>` in `<style>` | tab stop definition preserved |
| `w:tab` | break | `<tab/>` | tab character preserved |
| `w:r` | run | text content | — |
| `w:t` | text | element text | — |
| `w:rPr/w:b` | fmt | `<b>` | bold |
| `w:rPr/w:i` | fmt | `<i>` | italic |
| `w:rPr/w:u` | fmt | `<u>` | underline |
| `w:rPr/w:strike`,`w:rPr/w:dstrike` | fmt | `<s>` | strikethrough (CRIT-3) |
| `w:rPr/w:smallCaps` | fmt | `<smallcaps>` | small caps (MOD-4) |
| `w:rPr/w:caps` | fmt | `<uppercase>` | all caps (MOD-4) |
| `w:rPr/w:vertAlign` (sup/sub) | fmt | `<sup>`/`<sub>` | semantics |
| `w:rPr/w:rFonts` | fmt | `<span font="..">` | font family (KEEP) |
| `w:rPr/w:sz` | fmt | `<span size="..">` | font size in pt (KEEP) |
| `w:rPr/w:color` | fmt | `<span color="..">` | text color hex (KEEP) |
| `w:rPr/w:highlight` | fmt | `<span highlight="..">` | highlight color (KEEP) |
| `w:rPr/w:spacing` | present | DROP | character spacing noise |
| `w:rPr/w:lang` | keep | `lang="..."` on `<span>` | run-level language |
| `w:rPr/w:vanish` | keep | `hidden="true"` on `<span>` | hidden text |
| `w:rPr/w:rFonts/@w:eastAsia` | keep | `fontEA="..."` on `<span>` — P14 | East Asian font family |
| `w:rPr/w:rFonts/@w:cs` | keep | `fontCS="..."` on `<span>` — P15 | Complex Script font family |
| `w:rPr/w:bCs` | keep | `<bcs>...</bcs>` — P16 | Complex Script bold |
| `w:rPr/w:iCs` | keep | `<ics>...</ics>` — P17 | Complex Script italic |
| `w:rPr/w:szCs` | keep | `sizeCS="..."` on `<span>` — P18 | Complex Script font size |
| `w:br`,`w:cr` | break | `<br type="textWrapping|page|column|clear"/>` | explicit break w/ kind (MIN-1) |
| `w:tab` | break | `<tab/>` | tab character preserved |
| `w:noBreakHyphen`,`w:softHyphen`,`w:sym` | text | keep as char | literal |
| `w:hyperlink` | link | `<a href>` (r:id or instrText HYPERLINK) | link (MOD-3) |
| `w:instrText` (field code) | field | DROP unless HYPERLINK | field codes are noise |
| `w:fldSimple`,`w:fldChar` | field | DROP | TOC/PAGE/etc. noise |
| `w:bookmarkStart/End` | anchor | KEEP in `<notes>` as `<bm id="name"/>` | bookmark position preserved |
| `w:commentRange*`,`w:commentReference` | comment | KEEP in `<notes>` as `<comment>` | comment text preserved |
| `w:proofError` | proof | DROP | spelling/grammar noise |
| `w:ins`,`w:del` (track changes) | change | `<ins>/<del>` (optional, LOSSLESS_METADATA) | revision tracking preserved for legal docs |
| `w:sdt`,`w:smartTag`,`w:customXml` | wrapper | unwrap children | tag wrappers |
| `w:sectPr` | section | feed `<s:page>` in `<style>` | page layout |
| `w:sectPr/w:cols` | keep | `<s:cols n=".." space=".."/>` in `<style>` — P19 | multi-column layout |
| `w:tbl` | struct | `<table>` | table |
| `w:tblGrid`/`w:gridCol` | table layout | `<s:col ref="n">` widths (in `<style>`) | column widths linked by `ref` (MIN-3) |
| `w:tblPr` (borders) | present | `at="..."` on `<table>` | table borders preserved compact |
| `w:tblPr` (shading) | present | DROP | table shading noise |
| `w:tblPr/w:tblStyle` | style | `c="..."` attr on `<table>` + `<s:custom>` in `<style>` | keep style name + custom style definition |
| `w:tblPr/w:tblCaption` | keep | `caption="..."` attr on `<table>` | accessibility caption |
| `w:tblPr/w:tblDescription` | keep | `summary="..."` attr on `<table>` | accessibility description |
| `w:tblPr/w:tblW` | keep | `width="..."` attr on `<table>` | table width |
| `w:tblPr/w:jc` | keep | `align="left|center|right"` attr on `<table>` | table alignment |
| `w:tblPr/w:tblInd` | keep | `indent="..."` attr on `<table>` — P10 | table indentation (in declared unit) |
| `w:tblPr/w:tblCellSpacing` | keep | `cellSpacing="..."` attr on `<table>` — P11 | cell spacing (in declared unit) |
| `w:tcPr` (borders) | present | `at="..."` on `<td>`/`<th>` | cell borders preserved compact |
| `w:tcPr` (shading) | present | DROP | cell shading noise |
| `w:tcPr/w:vAlign` | keep | `valign="top|center|bottom"` on `<td>`/`<th>` | vertical alignment |
| `w:tcPr/w:textDirection` | keep | `textDir="..."` on `<td>`/`<th>` — P12 | text direction in cell |
| `w:tcPr/w:noWrap` | keep | `noWrap="true"` on `<td>`/`<th>` — P13 | no-wrap flag |
| `w:tr`/`w:tc` | struct | `<tr>`/`<th>`/`<td>` | table cells |
| `w:gridSpan`/`w:vMerge` | merge | `colspan`/`rowspan` on `<td>`/`<th>`; continue cells omitted | grid integrity preserved |
| `w:footnoteReference`/`w:endnoteReference` | note | `<fn-ref id="n" type="...">` marker; `<fn id="n" type="...">` body in `<notes>` | note marker + body, type distinguishes footnote/endnote |
| `w:drawing` (image blip) | **EXCLUDE** | `<img alt>` placeholder | images excluded |
| `w:pict` (VML) | **EXCLUDE** | DROP | legacy VML, no text extracted |
| `w:txbxContent` (textbox body) | KEEP | unwrap paragraphs/runs/tables into `<write>` | textbox text extracted (CRIT-1) |
| `w:hdrReference`,`w:ftrReference` | section | KEEP in `<header>`/`<footer>` blocks; margins via `<s:page mh/mf>` | header/footer content preserved |
| `w:object` (OLE) | **EXCLUDE** | DROP | complex object |
| charts / SmartArt / diagrams | **EXCLUDE** | DROP | complex object |
| `w:Math` (OMML) | **EXCLUDE** | DROP | math too complex |
| `w:altChunk` | EXCLUDE | DROP | external html chunk |

> **EXCLUDED by policy**: images (non-textbox), OLE objects, charts, SmartArt/diagrams,
> and Office Math. These are either binary or require specialized renderers, so the
> preprocessor emits a placeholder or drops them.
> **Textboxes (`w:txbxContent`) are NOT excluded** — their text content is extracted into
> `<write>` (CRIT-1). Images embedded *inside* a textbox are still excluded as `<img>`.
> **Headers/footers content** is now KEPT — see §2.6. Only the presentation chrome
> (shading) around headers/footers is dropped.
> **Bookmarks and comments** are now KEPT in `<notes>` — see §2.7.

### 3.1 Paragraph → element mapping

| Source (`w:pStyle w:val`) | Target element              |
|----------------------------|-----------------------------|
| `Heading1`                 | `<h1>`        |
| `Heading2`                 | `<h2>`        |
| `Heading3`                 | `<h3>`        |
| `Title`                    | `<h1>`           |
| `ListParagraph` (+ numPr)  | `<li>` (inside `<ul>`/`<ol>`) |
| `Quote`/`IntenseQuote`/`BlockText` | `<blockquote>` (MIN-2) |
| Code-like styles (see below) | `<pre>`               |
| (none / `Normal`)          | `<p>`                     |

- **Code block detection**: a paragraph maps to `<pre>` when either:
  - Its `w:pStyle w:val` resolved via `styles.xml` matches a code-like style name
    (`Code`, `HTML`, `XML`, `PlainText`, `SourceCode`, `Example`, `Output`, or any style
    whose `w:name` contains `"Code"`, `"Source"`, or `"Output"` as a word), OR
  - The **first run** in the paragraph uses a monospace font family
    (`w:rPr/w:rFonts/@w:ascii` or `@w:hAnsi` matching `Courier New`, `Consolas`,
    `Lucida Console`, `Menlo`, `Monaco`, `monospace`, or any font containing `"Mono"` or
    `"Courier"`, case-insensitive).
  - The entire paragraph content (including all runs) becomes the text content of `<pre>`,
    with original spacing preserved per §3.5.
  - If a paragraph qualifies as `<pre>`, all inline formatting tags (`<b>`, `<i>`,
    etc.) inside it are suppressed — only the raw text is kept.
- Drop `w:pPr` presentation noise (`w14:paraId`, `w:rsidR`, shading, tabs, etc.); layout attributes (spacing, indent, alignment) are preserved in `<style>` per §3.0; borders preserved via `at` attribute.
- The `c` attribute preserves the **original style name** for downstream semantic tagging.
- **Style resolution**: if `w:pStyle w:val` is a custom name, resolve it via `styles.xml`
  (`w:styleId` → `w:name`) to a semantic role (Heading/Quote/etc.) when possible; otherwise
  fall back to `c="<customName>"` and treat as `<p>`.
- **Style inheritance chain**: Word styles inherit from a parent via `w:basedOn`. The
  preprocessor MUST walk the inheritance chain to determine the final semantic role.
  For example, if `MyHeading` has `<w:basedOn w:val="Heading1"/>`, it is treated as
   `<h1 c="MyHeading">`. Resolution algorithm:
  1. Start with the paragraph's `w:pStyle w:val`.
  2. Look up its `w:style` entry in `styles.xml`.
  3. If it has a `w:basedOn` reference, recurse into the parent style.
  4. Stop at a known semantic style (Heading1-9, Title, Quote, Normal, ListParagraph, Code)
     or when no `w:basedOn` exists.
  5. Map the deepest known ancestor to the appropriate target element.

### 3.2 Runs → inline formatting

- `w:r` → `w:t` text becomes the element's text content.
- **Language (BCP 47)**: The `lang` attribute is set on the **target block element**
  (`<p>`, `<h1>`-`<h9>`, `<li>`, `<td>`, etc.) based on:
  1. Paragraph-level `w:pPr/w:pStyle/lang` if present
  2. Section default `w:lang` if paragraph-level is absent
  3. The first run's `w:rPr/w:rLang` as fallback
  If runs within the same paragraph have different languages, the first run's language
  takes precedence. Inline language changes are lost (`<span>` supports font/size/color/highlight
  but not `lang`).
- `w:rPr` with `w:b` → wrap run in `<b>`.
- `w:rPr` with `w:i` → wrap run in `<i>`.
- `w:rPr` with `w:u` → wrap run in `<u>`.
- `w:rPr` with `w:strike`/`w:dstrike` → wrap run in `<s>` (CRIT-3).
- `w:rPr` with `w:smallCaps` → wrap run in `<smallcaps>`; `w:caps` → `<uppercase>` (MOD-4).
- `w:rPr` with `w:vertAlign w:val="superscript"` → `<sup>`; `"subscript"` → `<sub>`.
- `w:rPr` with `w:rFonts` → wrap run in `<span font="..">` (font family name from `@w:ascii` or `@w:hAnsi`).
- `w:rPr` with `w:sz` → wrap run in `<span size="..">` (font size in half-points, converted to pt: `w:val ÷ 2`).
- `w:rPr` with `w:color` → wrap run in `<span color="..">` (hex color from `@w:val`, e.g., `"FF0000"`).
- `w:rPr` with `w:highlight` → wrap run in `<span highlight="..">` (highlight color name from `@w:val`).
- Multiple font properties on the same run are combined into a single `<span>` element with all applicable attributes.
- **Direction (MOD-7)**: paragraph `w:bidi`, run `w:rPr/w:rtl`, or inline `w:dir`/`w:bdo`
  → emit `dir="rtl"` on the affected element. Mixed LTR/RTL runs each carry their own `dir`.
- Hyperlinks (MOD-3) — resolve target in this order:
  1. `w:hyperlink/@r:id` → look up `document.xml.rels` → `<a href="...">`.
  2. Else if `w:hyperlink` contains `w:instrText` with `HYPERLINK "..."` → extract the
     URL from the field code → `<a href="...">`.
  3. Internal/bookmark targets (no URL) → `<a href="#bookmarkName">` when resolvable.
- **Textboxes (CRIT-1)**: `w:txbxContent` (inside `w:drawing` shapes or
  `mc:AlternateContent`) is unwrapped — its child paragraphs/runs/tables are processed by
  the normal rules. Only the *text* is kept; the shape/frame chrome is dropped.
  - **Inline anchor handling**: when a textbox is anchored inside a `<w:r>` run (inline),
    the host paragraph's runs *before* and *after* the textbox anchor are merged into a
    single `<p>` element. The textbox's own paragraphs are NOT spliced mid-sentence;
    instead they are emitted as **sibling elements immediately after** the host `<p>`.
    This prevents sentence fragmentation while keeping document order intact.
  - If the textbox is the sole content of its host paragraph (no surrounding runs), its
    paragraphs replace the host `<p>` directly.
- **Footnotes/endnotes**: `w:footnoteReference`/`w:endnoteReference` → `<fn-ref id="n" type="footnote|endnote"/>`
  marker in `<write>`. The body is extracted from `word/footnotes.xml` or `word/endnotes.xml`,
  processed through normal paragraph/run rules, and placed in the `<notes>` block as
  `<fn id="n" type="footnote|endnote">...</fn>` (see §2.7).
- **Tracked changes (LOSSLESS_METADATA)**: `w:ins` → `<ins>...</ins>`,
  `w:del` → `<del>...</del>` (deleted text included for context).
  Only emitted when the preprocessor is in `mode="lossless"`; in `mode="semantic"` (default),
  they are dropped per §3.0.

### 3.3 Lists

- Group consecutive `ListParagraph` paragraphs into a `<ul>` or `<ol>` with `<li>`
  children. Grouping MUST consider all of the following:
  - `w:numId` — numeric ID of the numbering definition.
  - `w:ilvl` — indent level (drives nesting).
  - `w:abstractNumId` — resolved from `numbering.xml` via the `w:num` entry.
  - **Restart state**: a `w:lvlOverride` with `w:startOverride` resets the numbering;
    this forces a SPLIT into a new `<ol>` element even if `w:numId` is unchanged.
  - A change in `w:abstractNumId` (different numbering scheme) also forces a split.
- Paragraphs with the same `w:numId` but different `w:ilvl` are parent/child within the
  same list structure (see nesting rules below).
- **Numbering restart**: detect `w:lvlOverride` in `numbering.xml` (under the matching
  `w:num` definition). When a `w:lvlOverride/w:startOverride/@w:val` resets numbering to 1
  (or another value), the preprocessor MUST split the list into a new `<ol>` element at
  the restart point. The new `<ol>` carries `start="n"` where `n` is the restart value
  (default `1`). Absent `w:lvlOverride`, no `start` attribute is emitted.
- **List type**: resolve `w:numId` → `w:abstractNumId` → `w:numFmt` in `numbering.xml`:
  - `bullet` → `<ul type="bullet">`
  - `decimal` → `<ol type="decimal">`
  - `lowerLetter`/`upperLetter` → `<ol type="lowerLetter|upperLetter">`
  - `lowerRoman`/`upperRoman` → `<ol type="lowerRoman|upperRoman">`
  - **Fallback (MOD-2)**: any other `w:numFmt` value (e.g., `hebrew1`, `arabicAlpha`,
    `thaiNumbers`, `chicago`, `ideographDigital`, …) → `<ol type="...">` with the
    **raw `w:numFmt` value preserved verbatim** as the `type` attribute. Never coerce to
    `decimal`. This keeps the numbering scheme discoverable downstream.
- **Nesting structure**: nested `<ul>`/`<ol>` elements are placed **inside** the
  `<li>` of the parent item. A level-N item becomes a direct child of the level-(N−1)
  `<li>`. Example:
  ```xml
  <ul type="bullet">
    <li>Parent item
      <ul type="bullet">
        <li>Child item (level 1)</li>
      </ul>
    </li>
  </ul>
  ```
  This structure allows arbitrary nesting depth and mixed list types (e.g., `<ol>` inside
  `<li>` inside `<ul>`).

### 3.4 Tables

- `w:tbl` → `<table id="n">` where `n` is a 1-based index across all tables in the
  document (in **pre-order traversal** of the XML tree). This `id` links back to
  `<s:col ref="n">` in `<style>`. Nested tables receive IDs in pre-order sequence
  (parent table gets its ID before its child tables).
- `w:tr` → `<tr>`.
- **Header rows (CRIT-2)**: a row is a header **iff** its `w:trPr/w:tblHeader` flag is set
  (not by position). Rows with `w:tblHeader` → `<tr><th>…</th></tr>`; all other
  rows → `<tr><td>…</td></tr>`. A table may have zero, one, or several successive
  header rows — each flagged row becomes a `<th>` row.
- **Merge cells — grid reconstruction**:
  - `w:gridSpan` → `colspan="n"` on `<td>`/`<th>`.
  - `w:vMerge` (vertical merge) requires reconstructing the grid to determine the final
    `rowspan`. Algorithm:
    1. Parse all rows of the table to build a virtual grid.
    2. A cell with `<w:vMerge w:val="restart"/>` starts a vertical merge group.
    3. Subsequent cells in the same column with `<w:vMerge/>` (no val, i.e., "continue")
       are part of that group. These continue-cells are **omitted** from output.
    4. The restart cell gets `rowspan="n"` where `n` = count of continues + 1.
       (For example: 1 restart + 2 continues = `rowspan="3"`)
    5. If `w:vMerge` appears without a prior restart in the column, treat as `rowspan="2"`
       (Word default behavior).
  - This preserves grid integrity: downstream parsers can reconstruct the exact cell grid
    without needing to resolve merge states.
  - Nested tables handled recursively.
- **Column widths (MIN-3)**: the authoritative source is `w:tblGrid` → `w:gridCol/@w:w`.
  Emit one `<s:col ref="n" w=".."/>` per `w:gridCol` into `<style>`, where `n` is the
  `<table id>` of the owning table. Per‑cell `w:tcW` is treated as a secondary override
  only when a `w:gridCol` is absent.
- Drop `w:tblPr`/`w:tcPr` shading; borders preserved via `at` attribute; `w:tblW` is informational.

### 3.5 Text cleanup

- **Processing mode** (controlled by `mode` attribute on root `<words>`):
  - `mode="semantic"` (default) — whitespace normalized for AI/training efficiency.
  - `mode="lossless"` — minimal transformation; whitespace and tracked changes preserved
    for round-tripping or legal/document-reconstruction scenarios.
- **Whitespace normalization** (applies in `semantic` mode only):
  - Collapse repeated spaces to single space, trim line breaks inside a run.
  - `w:tab` → `<tab/>` (tab character preserved); `w:br`/`w:cr` → `<br type="…"/>`.
  - **Exception**: content inside `<pre>` blocks is exempt — all original spacing,
    indentation, and line breaks are preserved verbatim regardless of mode.
  - **`xml:space` preservation**: if a `<w:t>` element carries `xml:space="preserve"`,
  the preprocessor MUST honor it and NOT collapse whitespace within that run,
  regardless of mode. This prevents data loss in poetry, code snippets, and
  ASCII diagrams where spacing is intentional.
- `w:tab` → `<tab/>` (tab character preserved); inside `<pre>` it remains literal tab (`\t`);
  `w:br`/`w:cr` → `<br type="…"/>` preserving `@w:type`
  (`textWrapping` default, `page`, `column`, `clear`) (MIN-1).
- `dir="rtl"` attributes are preserved on the relevant elements (MOD-7).
- Preserve intentional paragraph breaks as separate elements.
- Keep all original text verbatim (no translation/summarization at this stage).
- **XML escaping**: all text content and attribute values MUST be valid XML 1.0.
  The preprocessor MUST escape the following characters before emitting:
  - In text content: `&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`
  - In attribute values (href, c, type, etc.): additionally `"` → `&quot;`
  - CDATA sections are NOT used — all content is escaped inline.
  - This applies to all text from `w:t`, `w:instrText` (when kept), and resolved URLs
    from `w:hyperlink/@r:id` and `document.xml.rels` targets.
- **Forbidden XML 1.0 control characters**: the preprocessor MUST strip any character
  in the ranges `0x00–0x08`, `0x0B–0x0C`, `0x0E–0x1F`, and `0x7F–0x84` (except
  `0x09` tab, `0x0A` LF, `0x0D` CR which are valid). These characters are illegal
  in XML 1.0 and would produce malformed output.
- Drop tracked-change, proofing, and field-code noise per §3.0. Bookmarks and comments are preserved in `<notes>` (see §2.7).

---

## 4. Worked Example

**Input (raw docx XML):**

```xml
<w:p w14:paraId="7A3B2" w:rsidR="007X">
  <w:pPr><w:pStyle w:val="Heading1"/><w:pBdr><w:bottom w:val="single" w:sz="12" w:space="1" w:color="000000"/></w:pBdr></w:pPr>
  <w:r><w:t>Specifications for Data Center Racks</w:t></w:r>
</w:p>
<w:p>
  <w:pPr><w:pStyle w:val="Normal"/></w:pPr>
<w:r><w:t>Rack </w:t></w:r>
<w:r><w:rPr><w:b/></w:rPr><w:t>42U</w:t></w:r>
<w:r><w:t> houses servers.</w:t></w:r>
  <w:r><w:footnoteReference w:id="1"/></w:r>
</w:p>
<w:p>
  <w:pPr><w:pStyle w:val="ListParagraph"/><w:numPr><w:ilvl w:val="0"/><w:numId w:val="2"/></w:numPr></w:pPr>
  <w:r><w:t>Rack mount standard</w:t></w:r>
</w:p>
<w:p>
  <w:pPr><w:pStyle w:val="ListParagraph"/><w:numPr><w:ilvl w:val="0"/><w:numId w:val="2"/></w:numPr></w:pPr>
  <w:r><w:t>Cold aisle containment</w:t></w:r>
</w:p>
<w:p>
  <w:pPr><w:pStyle w:val="Normal"/></w:pPr>
  <w:hyperlink r:id="rId7"><w:r><w:t>See official guide</w:t></w:r></w:hyperlink>
</w:p>
<w:p>
  <w:pPr><w:pStyle w:val="Normal"/></w:pPr>
  <w:r><w:t>Note: </w:t></w:r>
  <w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="24"/><w:color w:val="FF0000"/></w:rPr><w:t>red text in Arial 12pt</w:t></w:r>
  <w:r><w:t>. This is an inline drawing with a textbox (see output note below).</w:t></w:r>
</w:p>
<w:p>
  <w:pPr><w:pStyle w:val="Normal"/><w:shd w:val="clear" w:color="auto" w:fill="FFFF00"/><w:spacing w:before="240" w:after="120"/><w:keepNext/></w:pPr>
  <w:r><w:t>Important note with yellow background.</w:t></w:r>
</w:p>
<w:bookmarkStart w:id="0" w:name="Section1"/>
<w:p>
  <w:pPr><w:pStyle w:val="Normal"/></w:pPr>
  <w:r><w:t>Textbox content extracted: Use C13 category bolt.</w:t></w:r>
</w:p>
<w:bookmarkEnd w:id="0"/>
```

**Output (`words` v1.1.0):**

```xml
<words xmlns="urn:words:v1" xmlns:s="urn:words:v1:style" version="1.1.0" mode="semantic">
  <style unit="in">
    <s:page size="A4" mt="0.75" mb="0.75" ml="0.75" mr="0.75" mh="0.5" mf="0.5"/>
    <s:gap el="h" c="Heading1" before="0.22" after="0.11"/>
    <s:gap el="p" before="0" after="0.11"/>
  </style>
  <write>
    <h1 at="bb 12 s1 #000000">Specifications for Data Center Racks</h1>
    <p>Rack <b>42U</b> houses servers.<fn-ref id="1" type="footnote"/></p>
    <ul type="bullet">
      <li>Rack mount standard</li>
      <li>Cold aisle containment</li>
    </ul>
    <p><a href="https://example.com/guide">See official guide</a></p>
    <p>Note: <span font="Arial" size="12" color="FF0000">red text in Arial 12pt</span>. This is an inline drawing with a textbox (see output note below).</p>
    <p shd="FFFF00" keepNext="true" spacingBefore="0.17" spacingAfter="0.08">Important note with yellow background.</p>
    <p>Textbox content extracted: Use C13 category bolt.</p>
  </write>
  <notes>
    <fn id="1" type="footnote">This is the footnote text for reference 1.</fn>
    <bm id="Section1"/>
  </notes>
</words>
```

---

## 5. Pipeline Position

```text
.docx  ──▶  [DOCX Preprocessor]  ──▶  words (v1.1.0)
 (OOXML)        (this module)         (semantic markup)
```

- The preprocessor is **pure transformation**: no LLM calls, deterministic, reproducible.
- Output `words` is the contract between DOCX ingestion and downstream processing.
- Version the format (`version="1.1.0"`) so downstream prompts/parsers can branch on schema.

---

## 6. Implementation Notes

- Parse the `word/document.xml` part inside the `.docx` (zip) package.
- Extract `styles.xml` only when style names need resolution beyond `w:pStyle w:val`.
- Emit strict, well-formed XML; validate against a schema.
- Idempotent: same `.docx` → identical `words` output.

---

## 7. Open Questions

None. All design decisions for v1.1.0 are finalized.

## 8. Explicitly Excluded (Policy)

The following are **out of scope** for v1.1.0 and are emitted as placeholders or dropped.

- **Images (non-textbox)** (`w:drawing` image blip, `w:pict` VML) → `<img alt="..."/>`
  placeholder, no pixels extracted. Images *inside* a textbox are also excluded.
- **Textboxes** are **NOT** excluded — `w:txbxContent` text is extracted (CRIT-1).
- **OLE objects** (`w:object`) → dropped.
- **Charts / SmartArt / diagrams** → dropped.
- **Office Math** (OMML, `w:Math`) → dropped.
- **External chunks** (`w:altChunk`) → dropped.
