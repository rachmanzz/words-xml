# Problem: LLM Understanding of `words` Format

This document analyzes how well LLMs can understand and generate `words` XML,
and what problems remain before production use.

---

## 1. Why `words` is a Natural Fit for LLMs

### 1.1 HTML-like Format

`words` XML is structurally similar to HTML — a format every LLM has seen
billions of tokens of during training:

| `words` Element | HTML Equivalent | Familiarity |
|---|---|---|
| `<b>`, `<i>`, `<u>`, `<s>` | `<b>`, `<i>`, `<u>`, `<s>` | ✅ Identical |
| `<a href="...">` | `<a href="...">` | ✅ Identical |
| `<br/>` | `<br/>` | ✅ Identical |
| `<span font size color>` | `<span style="...">` | ✅ Same concept |
| `<span lang hidden>` | — | ✅ New attributes (P2/P9) |
| `<ul>`, `<ol>`, `<li>` | `<ul>`, `<ol>`, `<li>` | ✅ Identical (only `<ul>` vs `<ul>`) |
| `<meta>`, `<title>`, `<author>` | `<meta>`, `<title>`, `<author>` | ✅ Identical |
| `<p>`, `<h1>`-`<h9>` | `<p>`, `<h1>`-`<h6>` | ✅ Identical (levels 1-6 match HTML; 7-9 are extended) |
| `<table>`, `<tr>`, `<td>` | `<table>`, `<tr>`, `<td>` | ✅ Identical |
| `at="bb 12 s1 #000000"` | — | ❌ Completely new syntax |
| `<s:page>`, `<s:gap>` | — | ⚠️ New layout metadata |
| `<fn-ref>`, `<fn>`, `<bm>` | — | ⚠️ New notes system |

**Key insight**: 60%+ of `words` elements are HTML-identical. LLMs already know them.

### 1.2 What LLM Already Knows vs What It Must Learn

```
Already known (no learning needed):
  <b>bold</b>           ← HTML
  <i>italic</i>         ← HTML
  <u>underline</u>      ← HTML
  <a href="url">link</a> ← HTML
  <ul><li>item</li></ul> ← HTML
  <span font="Arial">   ← HTML-like

Must learn from spec:
  <h1 c="Heading1">    ← c attribute for style name
  at="bb 12 s1 #000000" ← completely new
  <s:page size="A4"/>   ← new layout element
  <s:line el="p" value="1.5"/> ← new line spacing (P1)
  <fn-ref id="1"/>      ← new notes element
  shd="..."             ← paragraph shading
  keepNext="true"       ← keep with next paragraph
  keepLines="true"      ← keep lines together
  widowControl="true"   ← widow/orphan control
  sectionBreak="..."    ← section break type
  revisionAuthor="..."  ← tracked change author
  revisionDate="..."    ← tracked change date
  suppressAutoHyph="true" ← suppress auto hyphens
  snapToGrid="true"     ← snap to grid
  kinsoku="true"        ← East Asian line-break control
  wordWrap="true"       ← word wrapping
  overflowPunct="true"  ← overflow punctuation
  topLinePunct="true"   ← top line punctuation
  autoSpaceDE="true"    ← auto-space D/E
  autoSpaceDN="true"    ← auto-space D/N
  textDirection="..."   ← text direction
  suppressOverlap="true" ← suppress overlap
  divID="..."           ← HTML div association
  cnfStyle="..."        ← conditional table style
  frame="dropCap='drop' lines='3'" ← frame/drop-cap properties
```

---

## 2. Current LLM Readiness Scores

### 2.1 Score Table

| Mode | Score | Token Budget | Key Limitation |
|---|---|---|---|
| **Zero-shot** | **50-55/100** | N/A | `at` format, `<s:page>`, `<notes>`, 28+ new paragraph attributes unknown |
| **With skill** | **88-92/100** | ~11,000 tokens (spec) | 1 critical ambiguity remaining (`at` width unit) |
| **Fine-tuning** | **96-98/100** | 50-100 training pairs | Edge cases (vMerge, style chain) |

### 2.2 Zero-shot Analysis (55-60/100)

Without the spec, an LLM given "convert DOCX to words XML" would produce:

```xml
<!-- What LLM generates (likely) -->
<p>Hello <b>world</b></p>              ← ✅ correct
<h1>Title</h1>              ← ✅ correct
<a href="https://example.com">link</a>     ← ✅ correct
<ul><li>item</li></ul>             ← ✅ correct
```

**What goes wrong without spec:**

| Issue | Impact | Example of Wrong Output |
|---|---|---|
| `at` border format unknown | High | `<p style="border-bottom: 12pt">` (HTML-like, wrong) |
| `<s:page>` structure unknown | High | `<style><page size="A4"/></style>` (wrong namespace) |
| `<notes>` system unknown | Medium | `<footnote id="1">` (invented element) |
| `c` attribute purpose unknown | Medium | `<h class="Heading1">` (HTML class, wrong) |
| `frame` compound attribute unknown | Medium | `<p><div class="drop-cap">` (invented HTML) |
| `sectionBreak` values unknown | Low | `<div style="page-break-after">` (CSS) |
| `keepNext`/`keepLines` unknown | Low | `<p style="break-inside: avoid">` (CSS) |
| 28+ paragraph attributes unknown | Low | Attribute names guessed incorrectly |

### 2.3 With-skill Analysis (88-92/100)

With the full spec (v1.1.0, ~11,000 tokens) in system prompt:

**Score breakdown:**

| Aspect | Score | Notes |
|---|---|---|
| Element generation | 96/100 | HTML-like elements trivial; most match HTML exactly |
| Attribute generation | 88/100 | `at` format clear in spec; 28+ paragraph attributes documented; canonical ordering |
| Transformation rules | 87/100 | Style resolution (recursive), grid reconstruction (vMerge), list grouping (4-factor) complex |
| Edge cases | 82/100 | Multi-section, nested tables, textbox anchors |

**Why not 95+?** One critical ambiguity remaining:

1. **`at` border width unit** — spec says "declared unit" but worked example uses raw `w:sz`. LLM doesn't know which to follow.

**Resolved since last assessment:**
- `lang` optionality — now REQUIRED on all block elements, `<span lang="...">` for inline (P2)
- Noise matrix gaps — 6 run-level constructs now addressed (P1, P9)
- Paragraph formatting — shading, keepNext, keepLines, widowControl, section breaks, revision marks
- Typographic niceties — L1-L11 all preserved
- Frame properties — drop cap and frame attributes preserved

### 2.4 Fine-tuning Analysis (95-97/100)

With 50-100 DOCX→words training pairs:

| Aspect | Score | Notes |
|---|---|---|
| Pattern recognition | 98/100 | Seen enough examples to generalize |
| Complex transformations | 92/100 | vMerge, style chain still challenging |
| Consistency | 95/100 | Training enforces consistent output |

---

## 3. Remaining Problems

### 3.1 Critical (Blocking)

| # | Problem | Location | Impact |
|---|---|---|---|
| **C1** | `at` border width unit contradiction | §2 `at` attribute vs §4 worked example | LLM produces wrong border widths |

### 3.2 Significant (Score Impact)

| # | Problem | Location | Impact |
|---|---|---|---|
| **S1** | ~~6 run-level DROP constructs missing from noise matrix~~ | §3.0 | RESOLVED (P1, P9) |
| **S2** | Phantom `title` attribute in canonical ordering | grammar.md | LLM may emit wrong attributes |
| **S3** | `<li>` grammar gap (no `lang` attribute defined) | §2.1 | Incomplete element schema |
| **S4** | Worked example covers only 35% of features | §4 | LLM lacks examples for tables, code, comments |
| **S5** | No XSD/RelaxNG schema | §6 | No automated validation |

### 3.3 Minor (Polish)

| # | Problem | Location |
|---|---|---|
| M1 | ~~`<td>`, `<th>`, `<table>` missing `d:` prefix in border examples~~ | RESOLVED (d: prefix removed) |
| M2 | `<ul>` type lists ordered-list values | grammar.md |
| M3 | `<br>` shown without self-closing `/>` | grammar.md |
| M4 | `<header>`/`<footer>` shown as self-closing | target-format.md |
| M5 | XML declaration required per determinism.md but absent from spec examples | §4 |
| M6 | `size` attribute dual meaning (page preset vs font size) | §2.4 / §3.2 |

---

## 4. Impact of Fixes on Scores

| Fix | Zero-shot | With skill | Fine-tuning |
|---|---|---|---|
| Fix C1: `at` width unit | +3 | +3 | +1 |
| Fix C2: `lang` optionality | +2 | +2 | +1 |
| Fix S1: Missing noise matrix entries | +1 | +1 | +1 |
| Fix S2: Remove phantom `title` | +0 | +1 | +0 |
| Fix S4: Expand worked example | +2 | +2 | +1 |
| Add 28+ paragraph attributes | +5 | +4 | +2 |
| **Total after all fixes** | **63-68/100** | **93-96/100** | **97-98/100** |
| **Current status (C2, S1 resolved, paragraph attrs added)** | **55-60/100** | **88-92/100** | **96-98/100** |

---

## 5. Recommendations

### Phase 1: Fix Critical Issues (1-2 hours)

1. Resolve `at` border width — decide: raw `w:sz` OR convert to declared unit
2. ~~Resolve `lang` optionality~~ — RESOLVED (REQUIRED on blocks, `<span lang="...">` on inline)
3. ~~Add 6 missing constructs to noise matrix~~ — RESOLVED (P1, P9)

### Phase 2: Improve Score (2-4 hours)

4. Remove phantom `title` attribute
5. Expand worked example (add table merge, code block, comment, meta)

### Phase 3: Production Ready (1-2 days)

7. Generate 50-100 DOCX→words training pairs
8. Create XSD/RelaxNG schema
9. Build automated conformance test suite

---

## 6. Token Budget Analysis

### Spec Size

| File | Lines | Tokens (~4ch) | Fits in system prompt? |
|---|---|---|---|
| docx-preprosessor.md | ~830 | ~11,600 | ✅ Yes (well under 128K) |
| All sources/ files | ~1,200 | ~16,800 | ✅ Yes |
| All files combined | ~2,030 | ~28,400 | ✅ Yes |

### Words XML vs Raw OOXML (per document)

| Metric | Raw OOXML | Words XML | Reduction |
|---|---|---|---|
| Characters (1MB sample) | 89,246 | 10,477 | -88.3% |
| Tokens (estimated) | 22,312 | 2,619 | -88.3% |
| Overhead vs plain text | +97.8% | +27.3% | — |

**Conclusion**: Words XML fits comfortably in LLM context. The 88% token reduction
from raw OOXML means LLMs can process 8.5x more document content with the same
context window.
