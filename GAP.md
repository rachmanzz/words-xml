# GAP Analysis: words-xml v1.0.1

Audit date: 2026-07-22
Status: ACTIVE

---

## Summary

Gaps between the words-xml spec (v1.0.1) and current implementation, grouped by severity.

---

## GAP-01: Per-Paragraph Indent/Hanging Attributes
**Status**: FIXING NOW
**Priority**: HIGH

**Spec**: `<s:indent el="p" left=".." right=".." firstLine=".." hanging=".."/>` in `<style>` block (§2.4)
**Current**: Indent only emitted in `<s:custom>` style definitions. No per-paragraph indent on `<p>`.

**Impact**: Documents with direct `w:ind` on paragraphs (not inherited from style) lose indentation info.

**Fix**: Add `indentLeft`, `indentHanging`, `indentRight`, `indentFirst` attributes on `<p>` elements when they have non-zero indent that differs from their style's indent. Similar to how `align` was added.

**Files**:
- `words/preprocessor.go` — `writeParagraphAttrs`
- `words/verify.go` — add valid attributes

---

## GAP-02: List Continuation Across Non-List Paragraphs
**Status**: PARTIALLY FIXED
**Priority**: MEDIUM

**Spec**: "Group consecutive `ListParagraph` paragraphs into a `<ul>` or `<ol>`" (§3.3)
**Current**: Non-list paragraphs between same-`numId` items are skipped (items grouped, but continuation text lost).

**Impact**: In Word, list items with non-list continuation paragraphs between them are rendered as:
```
1. Item 1
   continuation text (non-list paragraph)
2. Item 2
```
Current output skips continuation text:
```xml
<ol><li>Item 1</li><li>Item 2</li></ol>
```

**Fix needed**: Wrap non-list continuation paragraphs inside the preceding `<li>` content.

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
**Status**: NOT FIXED
**Priority**: LOW

**Spec**: Indent should apply to all block elements with `w:ind`
**Current**: Only `<p>` has indent attributes. Headings, list items, blockquotes inherit from style.

**Impact**: Minimal — most documents use style-level indent for these elements.

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

## Implementation Order

```
GAP-01 (indent) → GAP-02 (list continuation) → GAP-05 (h1/li/blockquote indent)
```

## Verification

After each fix:
1. `go test ./words/... -count=1` — all tests pass
2. `go run examples/scripts/generate/main.go` — all outputs generated
3. `go run examples/scripts/verify/main.go` — all outputs verified clean
4. Manual check on `TEMP AKTA PENDIRIAN PT.docx` for indent correctness
