# Unit Declaration (words-xml)

This document is the **authoritative reference for all unit matters** in words-xml:
which values are converted to the declared unit, which values are not (and why), and
how the verifier treats them. The main specification (`docx-preprosessor.md`, §2.5)
summarizes this policy and defers to this document for the full rules.

## Purpose

This document explains how the words-xml preprocessor expresses the numeric layout
values it extracts from a DOCX file. Every number emitted by the preprocessor carries
an implicit unit, and that unit must be unambiguous for downstream consumers
(rendering engines, LLM prompts, validators).

OOXML does **not** store a document-wide unit. Word writes every length as an integer
number of **twips** (`1 pt = 20 twips`, `1 in = 1440 twips`), regardless of the unit
the user sees in the Word UI (the UI unit is an application preference, not a file
property). The words-xml preprocessor therefore declares a unit itself, in the
`<style>` block, and converts all twips-based lengths to that unit before emitting.

Because no default unit is explicitly declared in the DOCX, the preprocessor
**recommends `in` (inch)** as the default declared unit — the same default the spec
declares (see §2.5 Units). It converts every twips-based *physical length* to this
declared unit. Values that are inherently point-based (font sizes) are **never**
converted: they always stay in `pt` and are written with an explicit `pt` suffix.

## Declared Unit

The `<style>` block declares the default unit for all numeric layout values:

```xml
<style unit="in">
```

Allowed values: `in` (inch, default — **recommended**), `pt` (point), `px` (pixel),
`cm`, `mm`.

The preprocessor currently emits `unit="in"` (the recommended default) and converts
every twips-based physical length to inches. It does not yet auto-detect a different
unit from the DOCX (the DOCX does not declare one) nor convert to another declared
unit; those remain future extensions. Inline per-value overrides (e.g., `ml="2cm"`)
are declared by the spec but are not emitted by the preprocessor today.

**Point-only values**: some values are inherently point-based and are NEVER converted
to the declared unit — most notably **font sizes** (`size`, `sizeCS`). To remove any
ambiguity against the declared unit, every point-based value MUST carry an explicit
`pt` suffix after the number (e.g., `size="11pt"`, `sizeCS="10pt"`), regardless of the
declared `unit`.

Rules:

1. A bare number is interpreted in the declared unit (e.g., `ml="54"` with
   `unit="in"` means 54 inches — the preprocessor always converts, see below).
2. A value may override the default inline by suffixing its own unit
   (e.g., `ml="2cm"` even when `unit="pt"`).
3. The preprocessor MUST convert twips → the declared unit before emitting.
4. Point-based values (font sizes `size`, `sizeCS`) MUST stay in `pt` and MUST be
   suffixed with `pt` after the number (e.g., `size="11pt"`) — never bare, never
   converted to the declared unit.

## Convertible vs Non-Convertible Values (authoritative)

The table below is the **authoritative decision** for every numeric value the
preprocessor can emit: whether it is converted to the declared unit, and why.

| Emitted value                                          | OOXML source                        | Converted?                              | Why |
|--------------------------------------------------------|-------------------------------------|-----------------------------------------|-----|
| `<s:page>` `w h mt mb ml mr mh mf` (page + margins)    | `w:pgSz` / `w:pgMar` (twips)        | ✅ yes — twips ÷ 1440                    | physical length |
| `s:cols space` (column spacing)                        | `w:cols/@w:space` (twips)           | ✅ yes — twips ÷ 1440                    | physical length |
| `indentLeft/Right/First/Hanging` (`<p>`, `<s:custom>`, `<s:indent>`) | `w:pPr/w:ind` (twips) | ✅ yes — twips ÷ 1440                    | physical length |
| `spacingBefore/After` (`<p>`, `<s:custom>`, `<s:gap>`) | `w:pPr/w:spacing` (twips)           | ✅ yes — twips ÷ 1440                    | physical length |
| `lineSpacing` + `lineRule="exact\|atLeast"`            | `w:pPr/w:spacing/@w:line` (twips)   | ✅ yes — twips ÷ 1440                    | physical length |
| `<s:tab pos>`                                          | `w:pPr/w:tabs/@w:pos` (twips)       | ✅ yes — twips ÷ 1440                    | physical length |
| table `width` / `indent` / `cellSpacing`, `<s:col>` grid `w` | `w:tblPr`, `w:gridCol` (twips) | ✅ yes — twips ÷ 1440                | physical length |
| frame `width` / `height` / `vSpace` / `hSpace` / `x` / `y` | `w:pPr/w:framePr` (twips)       | ✅ yes — twips ÷ 1440                    | physical length |
| `at` border `width`, `borderWidth`                     | `w:sz` (eighths of a point)         | ✅ yes — ÷ 576                           | physical length |
| `lineSpacing` + `lineRule="auto"`                      | `w:pPr/w:spacing/@w:line`           | ❌ no — `w:line ÷ 240`                   | dimensionless multiplier |
| `size`, `sizeCS` (font sizes)                          | `w:sz` / `w:szCs` (half-points)     | ❌ no — kept in `pt`, always `pt`-suffixed | point-based font value |
| frame `lines` (drop-cap lines)                         | `w:framePr/@w:lines`                | ❌ no                                    | count |
| table `colspan`/`rowspan`, `s:cols n`, list `start`, heading level, `divID`, table `id` | various | ❌ no | count / identifier |
| `lineRule`, `alignment`, tab `leader`, border style, frame enums, colors, style/font names | various | ❌ no | enumeration |
| `w:leftChars` / `rightChars` / `firstLineChars` / `hangingChars` | `w:pPr/w:ind/@w:*Chars` | ❌ no | character-relative (depends on font metrics) |

The remaining sections expand on each row. The conversion table below lists the exact
sources and factors; the Exceptions section explains each non-converted group in detail.

## Conversion Rules

The preprocessor converts OOXML twips-based *physical lengths* to the declared unit:

| OOXML source                        | Conversion | Emitted as                                   |
|-------------------------------------|------------|----------------------------------------------|
| `w:pgSz`, `w:pgMar` (page geometry) | twips ÷ 1440 (→ in) | `<s:page w h mt mb ml mr mh mf>` |
| `w:cols/@w:space`                   | twips ÷ 1440 (→ in) | `<s:cols n space>`                           |
| `w:pPr/w:ind` (indents)             | twips ÷ 1440 (→ in) | `indentLeft`, `indentRight`, `indentFirst`, `indentHanging` |
| `w:pPr/w:spacing/@w:before`         | twips ÷ 1440 (→ in) | `spacingBefore`                             |
| `w:pPr/w:spacing/@w:after`          | twips ÷ 1440 (→ in) | `spacingAfter`                              |
| `w:pPr/w:spacing/@w:line` (exact/atLeast) | twips ÷ 1440 (→ in) | `lineSpacing` + `lineRule="exact|atLeast"`  |
| `w:pPr/w:spacing/@w:line` (auto)    | twips ÷ 240 (multiplier, dimensionless) | `lineSpacing` (bare multiplier; `lineRule="auto"` is suppressed on `<p>`, only shown in `<s:line rule="auto">`) |
| `w:pPr/w:tabs/@w:pos`               | twips ÷ 1440 (→ in) | `<s:tab pos>`                               |
| `w:pPr/w:framePr` `w`/`h`/`vSpace`/`hSpace`/`x`/`y` | twips ÷ 1440 (→ in) | `frame` `width`/`height`/`vSpace`/`hSpace`/`x`/`y` |
| `w:tblPr` widths / `w:gridCol` / `w:tcW` | twips ÷ 1440 (→ in) | table `width`, `indent`, `cellSpacing`, cell `width` |
| border width (`w:sz`, eighths of a point) | ÷ 576 (→ in) | `at`, `borderWidth`                         |

### Exceptions — values that are NOT converted

Everything below is deliberately **never** converted to the declared unit. They fall
into five groups. For each, the reason it cannot be converted and what it is emitted
as is given.

**Group 1 — point-based values (fonts).** These are real lengths, but they are
inherently *point*-based: typography defines type sizes in points, never in inches,
millimeters, or pixels. Converting them to the declared unit would destroy the font
semantics (a "11pt" typeface has no natural meaning as "0.15in" to a font engine).
They always stay in `pt` and MUST carry an explicit `pt` suffix so they are never
mistaken for the declared unit:

- `size`, `sizeCS` — font size (and Complex Script size) on `<span>` and `<s:custom>`.
  The DocDefaults font size is **never emitted** as an attribute — it only serves as
  the comparison baseline for attribute suppression. Source: OOXML half-points
  (`w:sz ÷ 2`, e.g., `w:val="22"` → `11pt`).

**Group 2 — dimensionless multipliers.** These have **no physical dimension at all**,
so there is nothing to convert: a ratio is a ratio regardless of unit.

- `lineSpacing` with `lineRule="auto"` is a **multiplier**: `w:line ÷ 240` produces
  `1.5` for one-and-a-half spacing, `2` for double. It is emitted as a bare number
  (`lineSpacing="1.5"`). Only `exact`/`atLeast` line spacing is a length and is
  converted to the declared unit (see conversion table).

**Group 3 — counts and identifiers.** These are whole-number counts, not lengths.
Converting a count "to inches" is meaningless:

- `frame` `lines` — number of drop-cap lines.
- Table `colspan`/`rowspan`, `w:cols/@w:num`, `s:cols n` — column counts.
- List `start`, heading level, `divID`, table `id` — integer identifiers/ordinals.

**Group 4 — enumerations (enums).** Values that select an option rather than measure
a distance. They are not numeric and are never converted:

- `lineRule="auto|exact|atLeast"`, paragraph/style `alignment`, tab `leader`,
  border style values.
- `frame` positional enums (`wrap`, `hAnchor`, `vAnchor`, `xAlign`, `yAlign`,
  `dropCap`, `hRule`) — enumerations. Only the frame's numeric *length* attributes
  (`width`, `height`, `vSpace`, `hSpace`, `x`, `y`) are lengths and ARE converted
  (see conversion table).
- Colors (`color`, `borderColor`), style names, font family names.

**Group 5 — character-relative values.** These are measured in *characters*, not in
any length unit. They depend on the font metrics of the text they apply to, so they
cannot be mapped to an absolute physical length without knowing the rendered width of
the characters:

- `w:leftChars`, `w:rightChars`, `w:firstLineChars`, `w:hangingChars` (hundredths of
  a character) are parsed but not converted; when present, the preprocessor uses the
  paired twips fallback (`w:left`, `w:right`, `w:firstLine`, `w:hanging`) instead.

## Example

Source paragraph with `w:spacing w:before="240" w:after="120"` and `unit="in"`:

```xml
<p spacingBefore="0.17" spacingAfter="0.08">…</p>
```

- `240 twips ÷ 1440 = 0.1667 in ≈ 0.17`
- `120 twips ÷ 1440 = 0.0833 in ≈ 0.08`

(For reference, the same paragraph expressed in points would be `spacingBefore="12.00"
spacingAfter="6.00"` — `240 ÷ 20`, `120 ÷ 20` — but the preprocessor emits `in`
today, not `pt`.)

Font sizes always carry their own `pt` suffix and are never converted, even with
`unit="in"`. A run with `w:sz w:val="22"` (22 half-points = 11 pt) emits:

```xml
<span size="11pt">…</span>
```

## Consistency Guarantee

All twips-based physical lengths within one document are converted with the **same**
conversion factor (the declared unit), so the values are mutually consistent: a
paragraph indent, its spacing, and the page margins are all expressed in the same
unit. Only the documented exceptions above deviate (font points — always `pt`-suffixed
— the auto line-spacing multiplier, counts/identifiers, enums, and character-width
indents are all excluded because they are not absolute physical lengths).

## Validation

The verifier accepts `in`, `cm`, `pt`, `px`, `mm` for `<style unit=…>` and warns if the
attribute is missing or unknown. Numeric layout attributes (`spacingBefore`,
`spacingAfter`, `lineSpacing`, `indentLeft`, `indentRight`, `indentFirst`,
`indentHanging`) must parse as numbers; `lineRule`
must be one of `auto`, `exact`, `atLeast`. Point-based values (`size`, `sizeCS`) must
be numbers followed by `pt` (e.g., `11pt`).
