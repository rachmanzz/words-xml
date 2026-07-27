# DOCX Preprocessor — Limitations

This document consolidates the **limitations and exclusions** of the DOCX Preprocessor
(`docx-preprosessor.md`, `words` v1.0.1). It is the authoritative reference for what the
preprocessor intentionally does **not** handle.

The preprocessor is a **deterministic, LLM-free, noise-removing transformer**. Its job is to
emit a semantic, token-efficient `words` document — not to faithfully reproduce every
binary or presentation detail of the source `.docx`.

---

## 1. Excluded by Policy (binary / too complex)

These are **out of scope** for v1.0.1. They are emitted as placeholders or dropped entirely.

| Excluded construct | OOXML source | Behavior in `words` | Impact |
|--------------------|--------------|--------------------|--------|
| **Images (non-textbox)** | `w:drawing` image blip, `w:pict` (VML) | `<img alt="..."/>` placeholder; no pixel/vector data | Visual content (photos, diagrams, scanned specs) lost |
| **OLE objects** | `w:object` | dropped | Embedded spreadsheets/Visio lost |
| **Charts** | chart drawing parts | dropped | Data visualizations lost |
| **SmartArt / diagrams** | smartart drawing parts | dropped | Flowcharts/org-charts lost |
| **Office Math** | `w:Math` (OMML) | dropped (not converted to LaTeX/MathML) | Equations absent |
| **External chunks** | `w:altChunk` | dropped | Embedded HTML lost |
| ~~**Headers/footers content**~~ | ~~`w:hdrReference`/`w:ftrReference`~~ | ~~dropped~~ | **RESOLVED**: kept in `<header>`/`<footer>` blocks (see spec §2.6) |

> Rationale: these require specialized renderers (image decoding, OLE/Visio parsers, math
> layout engines) and/or carry binary payloads. Converting them deterministically is
> out of scope and would bloat the token budget.

> **Textboxes are NOT excluded** (resolved, CRIT-1): `w:txbxContent` text is extracted
> into `<write>` by the normal rules. Only the image *inside* a textbox remains excluded.

**Workarounds**
- Images: describe manually, or run a separate OCR/Vision pass and inject captions before curation.
- Math: convert OMML → LaTeX in a dedicated preprocessing step if equations are required.
- OLE/charts: extract source data (e.g., the embedded workbook) and re-represent as text/table.

---

## 2. Dropped as Noise (presentation / revision)

Removed on purpose to keep the semantic body clean. They do **not** appear in `words`.

- **Paragraph presentation**: `pageBreakBefore`, `outlineLvl`.
  Note: `w:jc` (justification) is **preserved** as `<s:align>` in `<style>` (LOSSLESS_METADATA).
  Note: `w:pBdr` (borders) is **preserved** via compact `at` attribute on `<p>`/`<h1>`-`<h9>`.
  Note: `w:pPr/w:spacing/@w:line` is **preserved** as `<s:line>` in `<style>` (P1, LOSSLESS_METADATA).
  Note: `w:pPr/w:textAlignment` is **preserved** as `<p valign="...">` (P8, LOSSLESS_METADATA).
  Note: `w:pPr/w:tabs` is **preserved** as `<s:tab>` in `<style>` (LOSSLESS_METADATA).
  Note: `w:tab` is **preserved** as `<tab/>` inline element (LOSSLESS_METADATA).
  Note: `w:pPr/w:shd` is **preserved** as `<p shd="...">` (see §2a).
  Note: `w:pPr/w:framePr` is **preserved** as `<p frame="...">` (see §2a).
  Note: `w:pPr/w:keepNext` is **preserved** as `<p keepNext="true">` (see §2a).
  Note: `w:pPr/w:keepLines` is **preserved** as `<p keepLines="true">` (see §2a).
  Note: `w:pPr/w:widowControl` is **preserved** as `<p widowControl="true">` (see §2a).
- **Run presentation**: `w:rPr/w:spacing`,
  `w:shadow`, `w:emboss`, `w:imprint`, `w:effect`, `w:border` (run), `w:shd` (run).
  Note: `w:rFonts`, `w:sz`, `w:color`, `w:highlight` are **preserved** via `<span>` (see spec §3.2).
  `w:rPr/w:lang` is **preserved** as `<span lang="...">` (P2). `w:rPr/w:vanish` is **preserved** as `<span hidden="true">` (P9).
- **Fields**: `w:fldSimple`, `w:fldChar`, `w:instrText` (TOC, PAGE, REF, …) and their
  cached results — except HYPERLINK (resolved via `r:id` **or** `instrText`), kept as `<a>`.
- **Anchors**: `w:bookmarkStart` / `w:bookmarkEnd` — **preserved** in `<notes>` as `<bm id="name"/>`.
- **Review**: `w:commentRange*`, `w:commentReference` — **preserved** in `<notes>` as `<comment>`.
  `w:proofError` — dropped.
- **Track changes (semantic mode)**: `w:ins`, `w:del`, `w:delText` — dropped in `mode="semantic"` (default); preserved as `<ins>`/`<del>` in `mode="lossless"`.
- **Wrappers**: `w:sdt`, `w:smartTag`, `w:customXml` — unwrapped to children.

> These removals are intentional noise reduction, not defects.
> Note: strikethrough (`w:strike`/`w:dstrike`), small/ALL caps, and RTL direction are
> **preserved** (`<s>`, `<smallcaps>`, `<uppercase>`, `dir="rtl"`).

## 2a. Now Preserved (previously dropped)

The following items were previously listed as dropped but are now **preserved** in `words`:

| Construct | OOXML source | `words` representation | Notes |
|-----------|--------------|------------------------|-------|
| **Paragraph shading** | `w:pPr/w:shd` | `<p shd="...">` | hex color or pattern |
| **Frame properties** | `w:pPr/w:framePr` | `<p frame="...">` | compound attribute with drop-cap, lines, width, etc. |
| **Keep next paragraph** | `w:pPr/w:keepNext` | `<p keepNext="true">` | paragraph formatting |
| **Keep lines together** | `w:pPr/w:keepLines` | `<p keepLines="true">` | paragraph formatting |
| **Widow/orphan control** | `w:pPr/w:widowControl` | `<p widowControl="true">` | widow/orphan control |
| **Section breaks (inline)** | `w:pPr/w:sectPr` | `<p sectionBreak="...">` | `nextPage`, `continuous`, `evenPage`, `oddPage` |
| **Paragraph revision metadata** | `w:pPr/w:pPrChange` | `<p revisionAuthor="..." revisionDate="...">` | tracked change author/date |
| **Suppress auto-hyphens** | `w:pPr/w:suppressAutoHyphens` | `<p suppressAutoHyph="true">` | East Asian layout |
| **Snap to grid** | `w:pPr/w:snapToGrid` | `<p snapToGrid="true">` | layout grid |
| **Kinsoku** | `w:pPr/w:kinsoku` | `<p kinsoku="true">` | East Asian line-break control |
| **Word wrap** | `w:pPr/w:wordWrap` | `<p wordWrap="true">` | word wrapping |
| **Overflow punctuation** | `w:pPr/w:overflowPunct` | `<p overflowPunct="true">` | punctuation overflow |
| **Top line punctuation** | `w:pPr/w:topLinePunct` | `<p topLinePunct="true">` | top-line punctuation |
| **Auto-space DE** | `w:pPr/w:autoSpaceDE` | `<p autoSpaceDE="true">` | auto-spacing D/E |
| **Auto-space DN** | `w:pPr/w:autoSpaceDN` | `<p autoSpaceDN="true">` | auto-spacing D/N |
| **Text direction** | `w:pPr/w:textDirection` | `<p textDirection="...">` | `lr`, `rl`, `lrV`, `rlV` |
| **Suppress overlap** | `w:pPr/w:suppressOverlap` | `<p suppressOverlap="true">` | suppress floating object overlap |
| **HTML div association** | `w:pPr/w:divId` | `<p divID="...">` | HTML div ID |
| **Conditional table style** | `w:pPr/w:cnfStyle` | `<p cnfStyle="...">` | conditional formatting |
| **Paragraph run defaults** | `w:pPr/w:rPr` | inherited by runs | paragraph-level default run properties |

---

## 3. Partially Handled (kept but lossy)

| Construct | Handling | Lost detail |
|-----------|----------|-------------|
| **Table styling** (`w:tblPr`/`w:tcPr`) | borders preserved via `at` attribute; shading dropped; `w:tblGrid`/`w:gridCol` → `<s:col ref="n">` linked to `<table id="n">`; `colspan`/`rowspan` preserved; `w:tblW` → `<table width>` (P5); `w:jc` → `<table align>` (P6); `w:tblInd` → `<table indent>` (P10); `w:tblCellSpacing` → `<table cellSpacing>` (P11); `w:tblCaption` → `<table caption>` (P3); `w:tblDescription` → `<table summary>` (P4); `w:tcPr/w:vAlign` → `<td valign>` / `<th valign>` (P7); `w:tcPr/w:textDirection` → `<td textDir>` / `<th textDir>` (P12); `w:tcPr/w:noWrap` → `<td noWrap>` / `<th noWrap>` (P13) | shading/color lost |
| **Section layout** (`w:sectPr`) | page size + margins → `<s:page>`; **multiple sections** (`<s:page>` per section); headers/footers → `<header>`/`<footer>`; `w:cols` → `<s:cols n=".." space=".."/>` (P19) | — |
| **Vertical alignment** (`w:vertAlign`) | mapped to `<sup>`/`<sub>` | only sup/sub; other vertAlign values dropped |
| **Multilingual** (`w:lang`) | propagated to `lang` attribute per element (BCP 47) — block-level and `<span lang="...">` (P2) | resolved in spec v1.0.1 |
| **Code detection** (style/font heuristics) | pattern-matched to `<pre>`; monospace font OR style name | false positives/negatives possible (custom styles named "Code" for non-code content) |

---

## 4. Formatting Constraints

- **Units**: all layout values resolved to the declared `unit` (default `in`); twips
  converted. Inconsistent source units are normalized — original unit labels are not kept.
- **Processing mode**: `mode="semantic"` (default) normalizes whitespace; `mode="lossless"` preserves more layout detail for round-tripping.
- **Whitespace**: in `semantic` mode, repeated spaces collapsed, `w:tab` → `<tab/>`,
  `w:br`/`w:cr` → `<br/>`. Honors `xml:space="preserve"` when present on `<w:t>`.
  **Exception**: content inside `<pre>` preserves original spacing verbatim in both modes.
- **Global font**: `w:docDefaults/w:rPrDefault/w:rPr/w:rFonts` and theme fontScheme resolved to `<s:theme font=".." fontEA=".." fontCS="..">`. Priority: inline > style > global.
- **Text**: kept verbatim (no translation/summarization). Tracked-deletion text is wrapped
  in `<del>` in `lossless` mode; dropped in `semantic` mode.
- **XML escaping**: `&`, `<`, `>`, `"` are escaped per XML 1.0. Forbidden control characters
  (0x00–0x08, 0x0B–0x0C, 0x0E–0x1F, 0x7F–0x84) are stripped. No CDATA sections emitted.

---

## 5. Correctness Boundaries

- The transform is **idempotent** (same `.docx` → identical `words`) only when source
  relationships (`document.xml.rels`, `numbering.xml`, `styles.xml`) are present and valid.
- Output validity depends on a well-formed `.docx` package; corrupt or partial packages may
  produce partial `words` and are not auto-repaired.
- The preprocessor performs **no semantic understanding** — it maps structure, not meaning.

---

## 6. Planned Follow-ups

- Optionally preserve rendered field results (TOC/PAGE snapshot) instead of dropping.
- Revisit image handling: pluggable extractor interface (OCR/Vision) feeding `<img>`.
- Consider extracting textbox images via OCR (currently the textbox *text* is kept, the
  embedded image is still `<img>`).

### Resolved in v1.0.1

The following gaps were identified and addressed during specification review:

| Issue | Resolution |
|-------|-----------|
| **Column identity ambiguity** in multiple/nested tables | `<s:col ref="n">` paired with `<table id="n">` (1-based index) |
| **Grid integrity loss** from flatten `gridSpan`/`vMerge` | `colspan`/`rowspan` preserved as attributes on `<td>`/`<th>`, content not concatenated |
| **Whitespace contradiction** with `<pre>` | Content inside `<pre>` excluded from whitespace normalization; literal tab preserved |
| **Permanent footnote content loss** | `<notes>` container added after `</write>` to store footnote/endnote body text |
| **Blind to list numbering restart** (`w:lvlOverride`) | `w:lvlOverride` detected; list split into separate `<ol>` elements with `start="n"` attribute |
| **`<pre>` trigger undefined** | Style name matching (`Code`, `Source`, `Output`) + fallback monospace font detection; inline formatting suppressed inside `<pre>` |
| **Nested list structure unclear** | `<ul>`/`<ol>` placed inside parent `<li>`; arbitrary depth, mixed types allowed |
| **Sentence fragmentation** from inline textbox anchors | Runs before/after textbox anchor merged into single `<p>`; textbox content emitted as sibling elements afterward |
| **Malformed XML** from verbatim text & URLs | All text and attributes escaped per XML 1.0 (`&`, `<`, `>`, `"`); forbidden control chars (0x00–0x08, etc.) stripped |
| **Namespace XML undefined** | Two namespaces declared: `xmlns="urn:words:v1"` (all elements), `xmlns:s="urn:words:v1:style"` (layout); block and inline elements use no prefix |
| **DROP too aggressive** (justification, tracked changes) | `w:jc` → `<s:align>` in `<style>` (LOSSLESS_METADATA); `w:ins`/`w:del` → optional `<ins>`/`<del>` |
| **Header/footer dropped** | Now KEPT in `<header>`/`<footer>` blocks per section, processed with normal transformation rules |
| **Ambiguity marker vs body footnote** | Split: `<fn-ref id="n" type="footnote|endnote"/>` (marker) and `<fn id="n" type="...">` (body), `id` as connector |
| **vMerge restart/continue unresolved** | Grid reconstruction algorithm: restart → rowspan, continue → omit |
| **Inaccurate list grouping** | Grouping considers `numId` + `ilvl` + `abstractNumId` + restart state |
| **Whitespace normalization violates 1:1** | `mode="semantic"` (default) for AI; `mode="lossless"` for minimal transformation; `xml:space="preserve"` honored |
| **Style inheritance not resolved** | Inheritance chain (`w:basedOn`) walked until finding known semantic style |
| **Missing document metadata** | `<meta>` block added with title/author/created/modified/keywords from `docProps/core.xml` |
| **Language preservation optional** | `lang` attribute (BCP 47) now REQUIRED on all block elements |
| **Font/size/color/highlight dropped** | Now preserved via `<span font=".." size=".." color=".." highlight="..">` inline element |
| **Run language dropped** | Now preserved via `<span lang="...">` (P2) |
| **Hidden text dropped** | Now preserved via `<span hidden="true">` (P9) |
| **Line spacing dropped** | Now preserved via `<s:line>` in `<style>` (P1) |
| **Text alignment dropped** | Now preserved via `<p valign="...">` (P8) |
| **Table width dropped** | Now preserved via `<table width="...">` (P5) |
| **Table alignment dropped** | Now preserved via `<table align="...">` (P6) |
| **Table caption/summary dropped** | Now preserved via `<table caption="..." summary="...">` (P3/P4) |
| **Cell vertical alignment dropped** | Now preserved via `<td valign="...">` / `<th valign="...">` (P7) |
| **Table indentation dropped** | Now preserved via `<table indent="...">` (P10) |
| **Cell spacing dropped** | Now preserved via `<table cellSpacing="...">` (P11) |
| **Text direction in cell dropped** | Now preserved via `<td textDir="...">` / `<th textDir="...">` (P12) |
| **No-wrap flag dropped** | Now preserved via `<td noWrap="true">` / `<th noWrap="true">` (P13) |
| **Custom styles dropped** | Now preserved via `<s:custom>` in `<style>` + `c="..."` attribute on element |
| **`<style>` optional** | Now REQUIRED with at minimum `<s:page>` (size + margins); default unit changed to `in` |
| **Borders dropped** | Now preserved via compact `at` attribute on `<p>`, `<h1>`-`<h9>`, `<td>`, `<th>`, `<table>` |
| **Bookmarks dropped** | Now preserved in `<notes>` as `<bm id="name"/>` |
| **Comments dropped** | Now preserved in `<notes>` as `<comment id="n" author="..." date="...">text</comment>` |
| **Shading dropped** | Now preserved via `<p shd="...">` (paragraph shading) |
| **Frame properties dropped** | Now preserved via `<p frame="...">` (compound attribute with drop-cap, lines, width, etc.) |
| **keepNext/keepLines dropped** | Now preserved via `<p keepNext="true">` and `<p keepLines="true">` |
| **widowControl dropped** | Now preserved via `<p widowControl="true">` |
| **Section breaks in paragraphs lost** | Now preserved via `<p sectionBreak="...">` (nextPage, continuous, evenPage, oddPage) |
| **Paragraph revision marks lost** | Now preserved via `<p revisionAuthor="..." revisionDate="...">` |
| **Typographic niceties dropped** | Now preserved: suppressAutoHyph, snapToGrid, kinsoku, wordWrap, overflowPunct, topLinePunct, autoSpaceDE/DN, textDirection, suppressOverlap, divID, cnfStyle |
| **Paragraph-level run defaults ignored** | Now propagated to runs via `applyParaRunDefaults` |
| **List continuation across non-list paragraphs** | Non-list paragraphs between same-numId items included in preceding `<li>` with `<br>` separator (GAP-02) |
| **Heading/blockquote indent missing** | Now emitted on all block elements (GAP-05) |

---

*See also:* `docx-preprosessor.md` (§2.1, §3.0, §8).
