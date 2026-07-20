# Decision Logs

Architectural and technical decisions for the words-xml compiler project.
Every significant choice is recorded here so the rationale is preserved over time.

---

## Decisions

---

#### `[DECISION] Use Go (no other languages)`
- **Date**: 2026-07-20
- **Context**: The compiler needs to parse OOXML and emit custom XML efficiently.
  Go provides `encoding/xml` and `archive/zip` in the stdlib, with strong
  performance for text processing and zero external dependencies required.
- **Decision**: The compiler is built entirely in Go.
- **Alternatives**:
  - Python — richer XML ecosystem but slower, heavier runtime
  - Rust — excellent performance but higher complexity for XML work
- **Consequences**:
  - (+) High performance for XML transformation
  - (+) `encoding/xml` stdlib covers all parsing needs
  - (+) Consistent with the reference implementation in dcdtunning
  - (-) Must define all OOXML structs manually (no mature DOCX library in Go)
- **Owner**: project owner
- **Related**: `../dcdtunning/backend/go.mod`

---

#### `[DECISION] Manual XML output via strings.Builder, not encoding/xml encoder`
- **Date**: 2026-07-20
- **Context**: The `words` XML format has specific requirements — attribute ordering,
  self-closing elements, custom namespace prefixes — that are difficult to achieve
  with `encoding/xml.Encoder`.
- **Decision**: Generate XML manually using `strings.Builder` and `fmt.Fprintf`.
- **Alternatives**:
  - `encoding/xml.Encoder` — automatic but hard to control details
  - Third-party XML templating library
- **Consequences**:
  - (+) Full control over output formatting
  - (+) No namespace prefix handling issues
  - (+) Consistent with the reference implementation
  - (-) Must handle XML escaping manually
  - (-) More code to write and maintain
- **Owner**: project owner
- **Related**: Reference implementation `dcdtunning/backend/internal/preprocess/docx.go`

---

#### `[DECISION] OOXML structs with namespace-qualified XML tags`
- **Date**: 2026-07-20
- **Context**: OOXML uses namespace `http://schemas.openxmlformats.org/wordprocessingml/2006/main`
  consistently. All elements must be represented as Go structs for type-safe parsing.
- **Decision**: Use structs with `xml:"..."` tags qualified by the OOXML namespace.
  Parsing is done per-part (per-file inside the ZIP), not a single large document.xml.
- **Alternatives**:
  - DOM-style streaming with `xml.Decoder`
  - Third-party DOCX parsing library
- **Consequences**:
  - (+) Type-safe, easy to access from Go code
  - (+) Direct unmarshaling via `encoding/xml`
  - (-) Many structs to define
  - (-) Complex OOXML elements (VML, DrawingML) need separate struct hierarchies
- **Owner**: project owner
- **Related**: Spec `docx-preprosessor.md` §3.0 (Noise Matrix)

---

#### `[DECISION] In-memory processing, not streaming`
- **Date**: 2026-07-20
- **Context**: `.docx` files are typically small-to-medium (< 10MB). The ZIP is read
  entirely into memory before processing.
- **Decision**: Read the entire `.docx` into `[]byte`, then open with `zip.NewReader`.
- **Alternatives**:
  - Streaming ZIP extraction — saves memory for large files
  - Temporary file extraction to disk
- **Consequences**:
  - (+) Simple, no temp file management needed
  - (+) All parts accessible simultaneously (styles, rels, document)
  - (-) Very large files (>100MB) may exhaust memory
  - (-) Not suitable for memory-constrained edge runtimes
- **Owner**: project owner
- **Related**: Spec `docx-preprosessor.md` §6

---

#### `[DECISION] Idempotent output — same .docx → same words`
- **Date**: 2026-07-20
- **Context**: For testing, caching, and reproducibility, output must be deterministic.
- **Decision**: Transformations must be idempotent. No randomness, no injected timestamps,
  no output order that depends on map iteration order.
- **Alternatives**:
  - Output with additional metadata (processing timestamp, etc.)
- **Consequences**:
  - (+) Easier testing — expected output can be compared directly
  - (+) Cacheable for identical files
  - (-) Cannot embed processing timestamp in output
- **Owner**: project owner
- **Related**: Spec `docx-preprosessor.md` §6

---

#### `[DECISION] No external dependencies for parsing`
- **Date**: 2026-07-20
- **Context**: Minimize attack surface and avoid dependency management overhead.
- **Decision**: Use only Go stdlib for all parsing: `archive/zip`, `encoding/xml`,
  `strings`, `bytes`, `regexp`, `strconv`, `path/filepath`.
- **Alternatives**:
  - `github.com/nguyenthenguyen/docx` or similar library
  - `github.com/unidoc/unioffice` (commercial)
- **Consequences**:
  - (+) Zero external dependencies for core functionality
  - (+) Easy to build, no internet needed for `go mod download`
  - (+) Full control over parsing behavior
  - (-) Must handle all OOXML edge cases manually
  - (-) More code to write
- **Owner**: project owner
- **Related**: Spec `docx-preprosessor.md`

---

#### `[DECISION] Custom XML serialization, not Marshal`
- **Date**: 2026-07-20
- **Context**: `encoding/xml.Marshal` produces XML that does not match the `words`
  v1.0.1 spec — attribute ordering, self-closing elements, and namespace handling
  all differ.
- **Decision**: Output XML is generated manually with `strings.Builder`. Each element
  is written explicitly with `fmt.Fprintf`.
- **Alternatives**:
  - Custom XML encoder wrapper
  - XML template library
- **Consequences**:
  - (+) Output format matches spec exactly
  - (+) Handles edge cases like `<tab/>`, `<br type="page"/>`, self-closing `<img>`
  - (-) More verbose, but more predictable
- **Owner**: project owner
- **Related**: Reference implementation, spec §2.1

---

#### `[DECISION] Separate packages: ooxml, preprocessor, types`
- **Date**: 2026-07-20
- **Context**: The reference implementation in dcdtunning is a single monolithic file
  (2818 lines). For better maintainability, split into multiple packages.
- **Decision**:
  - `internal/ooxml/` — all OOXML struct definitions
  - `internal/preprocessor/` — processing pipeline
  - `internal/types/` — domain types (ParsedDocument, ParsedParagraph, etc.)
- **Alternatives**:
  - Single monolithic file like the reference implementation
  - Flat package structure
- **Consequences**:
  - (+) Easier to maintain and test
  - (+) Clear separation of concerns — parsing vs processing vs types
  - (+) Can be imported as a library by other projects
  - (-) Slightly longer import paths
- **Owner**: project owner
- **Related**: `PLAN.md` §5

---

#### `[DECISION] CLI tool with stdout output`
- **Date**: 2026-07-20
- **Context**: The compiler should work as a standalone tool and be pipeable to other
  commands (Unix philosophy).
- **Decision**: Default output to stdout. File output via `--output` flag.
- **Alternatives**:
  - Always write to output file
  - Library-only, no CLI
- **Consequences**:
  - (+) Pipeable: `words-xml input.docx | xmllint --format -`
  - (+) Redirect to file: `words-xml input.docx > output.xml`
  - (+) Composable with other tools in pipelines
  - (-) Errors must go to stderr to avoid mixing with output
- **Owner**: project owner
- **Related**: `PLAN.md` §5

---

---

### Template (copy per decision)

```markdown
#### `[DECISION] <short title>`
- **Date**: YYYY-MM-DD
- **Context**: What problem or situation prompted the decision
- **Decision**: What was chosen
- **Alternatives**: Other options evaluated
- **Consequences**: Trade-offs and implications
- **Owner**: Who made/approved it
- **Related**: Link to strategy, issue, or ADR
```
