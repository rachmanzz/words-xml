# words-xml

Go library + CLI to convert `.docx` (OOXML) files into **words-XML** — a compact, deterministic, LLM-friendly XML representation.

Spec: [words-XML v1.0.1](https://github.com/rachmanzz/docx-preprocessor/blob/main/docx-preprosessor.md)

## Install

```bash
go get github.com/rachmanzz/words-xml
```

## Usage

### CLI

```bash
# semantic mode (default)
go run ./cmd/words-xml -i document.docx -o output.xml

# lossless mode
go run ./cmd/words-xml -i document.docx -o output.xml --mode lossless

# stdin/stdout
cat document.docx | go run ./cmd/words-xml > output.xml
```

### Library

```go
import "github.com/rachmanzz/words-xml/words"

// From file
result, err := words.ProcessDOCXFile("document.docx")

// From bytes
result, err := words.ProcessDOCXBytes(data)

// Lossless mode
result, err := words.ProcessDOCXFileMode("document.docx", "lossless")

// Result
fmt.Println(result.WordsXML)    // words-XML output
fmt.Println(result.Document)    // *ParsedDocument (structured access)
```

### Verification

```go
result := words.Verify(wordsXML)
if !result.Valid {
    for _, e := range result.Errors {
        fmt.Println(e)
    }
}
```

## Output Example

```xml
<words xmlns="urn:words:v1" version="1.0.1" mode="semantic">
  <style unit="in">
    <s:page size="Letter" mt="1.00" mb="1.00" ml="1.00" mr="1.00"/>
    <s:gap el="h" c="Heading1" before="18.00" after="4.00"/>
  </style>
  <write>
    <h1><span font="Arial" size="20" color="0f4761">Title</span></h1>
    <p>Body text with <b>bold</b> and <i>italic</i>.</p>
  </write>
</words>
```

## Features

- Semantic mode: whitespace normalization, monospace detection, code blocks
- Lossless mode: preserves original formatting exactly
- 100% spec coverage (96 elements, 140+ attributes)
- Zero external dependencies (Go stdlib only)
- Table support: colspec, borders, colspan/rowspan, vertical merge
- List support: nested lists, numFmt from numbering definitions
- RTL support: `<span dir="rtl">`, `<p dir="rtl">`
- Hyperlinks with relationship resolution
- Footnotes, endnotes, bookmarks
- Metadata: title, author, dates, keywords

## Testing

```bash
go test ./words/ -v
```

## Scripts

```bash
# Generate XML from all example .docx files
go run ./examples/scripts/generate/

# Verify all generated XML files
go run ./examples/scripts/verify/
```

## License

MIT
