# Graph Report - words-xml  (2026-08-04)

## Corpus Check
- 28 files · ~62,531 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 786 nodes · 1909 edges · 35 communities (30 shown, 5 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 263 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c700ad93`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Verify
- preprocessor.go
- buildStyleMap
- preprocessor_test.go
- verify.go
- 2. Target Format: `words` (semantic mode)
- GAP Analysis: words-xml v1.1.0
- ooxml.go
- ProcessDOCXBytes
- DOCX → Words: Gap Analysis
- ROUND 2 — MEDIUM IMPACT
- Planned
- Problem: LLM Understanding of `words` Format
- Plan: Paragraph Feature Coverage
- DocPara
- 4. Milestones
- IntVal
- bash
- unmarshalOrderedContent
- extractTheme
- Decisions
- DocOrderedContent
- RunProps
- words-xml
- DocDrawing
- TestComment
- DOCX Preprocessor — Limitations
- Unit Declaration (words-xml)
- AGENTS.md
- TestBrInListItem
- graphify.js
- github.com/rachmanzz/words-xml
- hard_audit_test.go
- TestDocumentOrderNotes
- TestMetdata

## God Nodes (most connected - your core abstractions)
1. `ProcessDOCXBytes()` - 128 edges
2. `Verify()` - 90 edges
3. `makeMinimalDocx()` - 86 edges
4. `makeDocxWithParts()` - 42 edges
5. `VerifyResult` - 32 edges
6. `GAP Analysis: words-xml v1.1.0` - 29 edges
7. `ParsedDocument` - 25 edges
8. `ProcessDOCXBytesMode()` - 24 edges
9. `ParaProps` - 22 edges
10. `verifyBlockContent()` - 21 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ProcessDOCXFile()`  [INFERRED]
  cmd/debug/main.go → words/preprocessor.go
- `main()` --calls--> `ProcessDOCXFileMode()`  [INFERRED]
  cmd/proc/main.go → words/preprocessor.go
- `TestBrInListItem()` --calls--> `ProcessDOCXBytes()`  [INFERRED]
  words/br_test.go → words/preprocessor.go
- `TestBrInListItem()` --calls--> `makeDocxWithParts()`  [INFERRED]
  words/br_test.go → words/preprocessor_test.go
- `TestBrSameRunAsText()` --calls--> `ProcessDOCXBytes()`  [INFERRED]
  words/br_test.go → words/preprocessor.go

## Import Cycles
- None detected.

## Communities (35 total, 5 thin omitted)

### Community 0 - "Verify"
Cohesion: 0.07
Nodes (86): T, TestVerifyBadAlign(), TestVerifyBadBreakType(), TestVerifyBadColspan(), TestVerifyBadColSpec(), TestVerifyBadDir(), TestVerifyBadLiType(), TestVerifyBadMode() (+78 more)

### Community 1 - "preprocessor.go"
Cohesion: 0.07
Nodes (85): Builder, main(), main(), File, BorderInfo, ContentItem, DocFooter, DocHeader (+77 more)

### Community 2 - "buildStyleMap"
Cohesion: 0.18
Nodes (11): buildStyleMap(), TestBuildStyleMapBasedOn(), TestBuildStyleMapHAnsiFallback(), TestBuildStyleMapInvalidXML(), TestBuildStyleMapLineRuleExact(), TestBuildStyleMapNoRPr(), TestBuildStyleMapParaProps(), TestBuildStyleMapStyleFontHAnsi() (+3 more)

### Community 3 - "preprocessor_test.go"
Cohesion: 0.06
Nodes (94): T, makeMinimalDocx(), TestAllCaps(), TestBoldCSItalicCS(), TestBoldCSItalicCSSuppressedWhenSameFont(), TestBoldExplicitOff(), TestBoldExplicitOffOverridesParaDefault(), TestBookmark() (+86 more)

### Community 4 - "verify.go"
Cohesion: 0.30
Nodes (33): Token, findMatchingEnd(), StartElement, isPtSuffixValid(), verifyAnchor(), verifyBlockContent(), verifyBlockquote(), verifyBreak() (+25 more)

### Community 5 - "2. Target Format: `words` (semantic mode)"
Cohesion: 0.06
Nodes (31): 1.1 Problem (semantic mode), 1.2 Goal (semantic mode), 1. Scope and Mode, 2.1 Grammar (v1.1.0), 2.2 Minimal Style Block, 2.3 `<meta>` — Document Metadata Block (optional), 2.4 `<style>` — Layout Block (required), 2.5 Units (+23 more)

### Community 6 - "GAP Analysis: words-xml v1.1.0"
Cohesion: 0.07
Nodes (29): GAP-01: Per-Paragraph Indent/Hanging Attributes, GAP-02: List Continuation Across Non-List Paragraphs, GAP-03: `<s:indent>` Not Emitted for Normal/Heading Styles, GAP-04: Per-Paragraph Align Only for Direct `jc` Values, GAP-05: `<h1>`-`<h9>`, `<li>`, `<blockquote>` Missing Indent, GAP-06: Tab Before Text in Same Run Discards Text, GAP-07: Centering Lost for Mixed-Alignment Documents, GAP-08: Per-Paragraph Spacing Not Emitted (+21 more)

### Community 7 - "ooxml.go"
Cohesion: 0.10
Nodes (37): AClrScheme, AClrSchemeEntry, AFontScheme, AFontSchemeFace, ASrgbClr, ATheme, AThemeElements, ATypeface (+29 more)

### Community 8 - "ProcessDOCXBytes"
Cohesion: 0.10
Nodes (41): ProcessDOCXBytes(), makeDocxWithParts(), TestBulletList(), TestCodeBlockDetection(), TestCustomStyleSizePtSuffix(), TestEmitStyleBlockHeadingIndent(), TestEmitStyleBlockNormalIndent(), TestFootnote() (+33 more)

### Community 9 - "DOCX → Words: Gap Analysis"
Cohesion: 0.07
Nodes (26): 1.1 Sample: `file-sample_1MB.docx`, 1.2 Compression Ratios, 1.3 Why the Reduction is So Large, 1.4 What the 27% Overhead Buys You, 1. Token Comparison, 2.1 What is Preserved (1:1), 2.2 What is Reduced (Compact), 2.3 What is Dropped (Noise) (+18 more)

### Community 10 - "ROUND 2 — MEDIUM IMPACT"
Cohesion: 0.08
Nodes (25): FIX-01: Resolve theme font references, FIX-02: Border width unit conversion, FIX-03: Remove `<colspec>` from inside `<table>`, FIX-04: Table cell content emit lists, FIX-05: Note bodies emit non-paragraph content, FIX-06: Add `lang` to table cells, FIX-07: Merge theme data into defaultFont, FIX-08: Parse paragraph-level line spacing (+17 more)

### Community 11 - "Planned"
Cohesion: 0.08
Nodes (23): [FEATURE] Bidi / RTL Support, [FEATURE] CLI Interface, [FEATURE] Dual Mode (Semantic & Lossless), [FEATURE] Header/Footer Emission, [FEATURE] Image Placeholder, [FEATURE] List Grouping, [FEATURE] Metadata Emission, [FEATURE] Notes Block Emission (+15 more)

### Community 12 - "Problem: LLM Understanding of `words` Format"
Cohesion: 0.09
Nodes (21): 1.1 HTML-like Format, 1.2 What LLM Already Knows vs What It Must Learn, 1. Why `words` is a Natural Fit for LLMs, 2.1 Score Table, 2.2 Zero-shot Analysis (55-60/100), 2.3 With-skill Analysis (88-92/100), 2.4 Fine-tuning Analysis (95-97/100), 2. Current LLM Readiness Scores (+13 more)

### Community 13 - "Plan: Paragraph Feature Coverage"
Cohesion: 0.10
Nodes (19): Current State, Files Modified (Phase 2), Files Modified (Phase 3), Files Modified (Phase 4), Files Modified (Phase 5), Files Modified (Phase 6), Fully Covered (tested), Gap Analysis (+11 more)

### Community 14 - "DocPara"
Cohesion: 0.14
Nodes (17): Name, BrVal, DirBdo, DocComment, DocComments, DocDel, DocDocument, DocHyperlink (+9 more)

### Community 15 - "4. Milestones"
Cohesion: 0.11
Nodes (17): 1. Summary, 2. Goals, 3. Pipeline Architecture, 4. Milestones, 5. Package Layout, 6. Core Libraries, 7. References, Milestone 1 — Foundation (+9 more)

### Community 16 - "IntVal"
Cohesion: 0.32
Nodes (8): AbstractLvl, AbstractNum, DocNumbering, FmtVal, IntVal, LvlOverride, NumDef, NumPr

### Community 17 - "bash"
Cohesion: 0.12
Nodes (15): git add *, git commit *, git push *, git tag *, AGENTS.md, docx-preprosessor.md, instructions, permission (+7 more)

### Community 18 - "unmarshalOrderedContent"
Cohesion: 0.30
Nodes (5): Decoder, StartElement, unmarshalOrderedContent(), Decoder, StartElement

### Community 19 - "extractTheme"
Cohesion: 0.50
Nodes (4): extractTheme(), TestExtractThemeEmptyFont(), TestExtractThemeInvalid(), TestExtractThemeValid()

### Community 20 - "Decisions"
Cohesion: 0.15
Nodes (12): `[DECISION] CLI tool with stdout output`, `[DECISION] Custom XML serialization, not Marshal`, `[DECISION] Idempotent output — same .docx → same words`, `[DECISION] In-memory processing, not streaming`, Decision Logs, `[DECISION] Manual XML output via strings.Builder, not encoding/xml encoder`, `[DECISION] No external dependencies for parsing`, `[DECISION] OOXML structs with namespace-qualified XML tags` (+4 more)

### Community 21 - "DocOrderedContent"
Cohesion: 0.16
Nodes (13): BodyChild, BodyChildType, DocBody, DocBookmark, DocOrderedContent, DocSdt, DocSdtContent, DocTbl (+5 more)

### Community 22 - "RunProps"
Cohesion: 0.13
Nodes (19): ColorVal, DocDefaults, DocStyleDef, DocStyles, HighlightVal, JCVal, LangVal, OnOffVal (+11 more)

### Community 23 - "words-xml"
Cohesion: 0.17
Nodes (11): CLI, Features, Install, Library, License, Output Example, Scripts, Testing (+3 more)

### Community 24 - "DocDrawing"
Cohesion: 0.21
Nodes (12): AGraphic, AGraphicData, DocDrawing, DocTextbox, DocTxbxContent, PicBlip, PicBlipFill, PicPic (+4 more)

### Community 26 - "DOCX Preprocessor — Limitations"
Cohesion: 0.20
Nodes (9): 1. Excluded by Policy (binary / too complex), 2. Dropped as Noise (presentation / revision), 2a. Now Preserved (previously dropped), 3. Partially Handled (kept but lossy), 4. Formatting Constraints, 5. Correctness Boundaries, 6. Planned Follow-ups, DOCX Preprocessor — Limitations (+1 more)

### Community 27 - "Unit Declaration (words-xml)"
Cohesion: 0.20
Nodes (9): Consistency Guarantee, Conversion Rules, Convertible vs Non-Convertible Values (authoritative), Declared Unit, Example, Exceptions — values that are NOT converted, Purpose, Unit Declaration (words-xml) (+1 more)

### Community 28 - "AGENTS.md"
Cohesion: 0.40
Nodes (3): Knowledge Graph (graphify), Reference Projects, Versioning Rule

### Community 29 - "TestBrInListItem"
Cohesion: 0.60
Nodes (4): T, TestBrInListItem(), TestBrSameRunAsText(), TestBrSeparateRuns()

### Community 33 - "hard_audit_test.go"
Cohesion: 0.20
Nodes (15): T, makeMinimalDocxWithNumbering(), TestDocumentOrderParaAfterTable(), TestDocumentOrderSdtBeforePara(), TestDocumentOrderTableBeforePara(), TestHasSameNumIDAheadLimit(), TestPostProcessRunOrderWithTableBefore(), makeDocxWithExtras() (+7 more)

## Knowledge Gaps
- **209 isolated node(s):** `github.com/rachmanzz/words-xml`, `$schema`, `AGENTS.md`, `docx-preprosessor.md`, `AGENTS.md` (+204 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `ProcessDOCXBytes()` connect `ProcessDOCXBytes` to `hard_audit_test.go`, `preprocessor.go`, `preprocessor_test.go`, `TestDocumentOrderNotes`, `TestMetdata`, `TestComment`, `TestBrInListItem`?**
  _High betweenness centrality (0.102) - this node is a cross-community bridge._
- **Why does `ProcessDOCXBytesMode()` connect `preprocessor.go` to `hard_audit_test.go`, `buildStyleMap`, `preprocessor_test.go`, `ProcessDOCXBytes`, `extractTheme`, `DocOrderedContent`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Why does `DocPara` connect `DocPara` to `DocDrawing`, `preprocessor.go`, `DocOrderedContent`, `ooxml.go`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Are the 125 inferred relationships involving `ProcessDOCXBytes()` (e.g. with `TestBrInListItem()` and `TestBrSameRunAsText()`) actually correct?**
  _`ProcessDOCXBytes()` has 125 INFERRED edges - model-reasoned connections that need verification._
- **Are the 84 inferred relationships involving `Verify()` (e.g. with `TestVerifyBadAlign()` and `TestVerifyBadBreakType()`) actually correct?**
  _`Verify()` has 84 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `makeMinimalDocx()` (e.g. with `TestBrSameRunAsText()` and `TestBrSeparateRuns()`) actually correct?**
  _`makeMinimalDocx()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/rachmanzz/words-xml`, `$schema`, `AGENTS.md` to the rest of the system?**
  _209 weakly-connected nodes found - possible documentation gaps or missing edges._