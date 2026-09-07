# Document workflows (v0.1.4)

The Pandoc group adds 10 Operations through the shared application registry. Install the pinned, SHA-256-verified Pandoc 3.11 package with `fnsh pkg add pandoc`. The registered and locally tested platform is Windows x64. It is downloaded separately under GPL-2.0-or-later; FinishBit does not bundle its binary in releases.

| Operation | Behavior |
| --- | --- |
| `document.formats` | Return the wrapper's supported input/output and bibliography matrix |
| `document.convert` | Convert structured documents, office files, ebooks and presentation output |
| `document.merge` | Combine 2–16 documents using one explicit reader, in input order |
| `document.text` | Extract readable plain text, with configurable wrapping |
| `document.inspect` | Return Pandoc metadata, ordered headings, links, images and AST node counts |
| `document.media` | Extract embedded DOCX/ODT/EPUB resources into numbered files; return original-name mapping |
| `document.split` | Split top-level blocks at headings into numbered text documents |
| `document.template` | Export one of eight built-in text templates |
| `document.reference` | Export default DOCX, ODT or PPTX reference files for customization |
| `document.bibliography` | Convert BibTeX, BibLaTeX, CSL JSON and RIS records |

## Formats and options

Readers: `markdown`, `gfm`, `commonmark`, `html`, `docx`, `odt`, `epub`, `rst`, `latex`, `org`, `textile`, `docbook`, `jats`, `ipynb`, `typst`, `json` (Pandoc AST).

Writers: `markdown`, `gfm`, `commonmark`, `html`, `html5`, `docx`, `odt`, `epub`, `epub3`, `rst`, `latex`, `org`, `textile`, `docbook`, `jats`, `ipynb`, `typst`, `json`, `plain`, `rtf`, `pptx`, `asciidoc`, `man`, `revealjs`, `beamer`.

`--from auto` infers common extensions; use an explicit reader for ambiguous XML, JSON or extensionless files. PPTX is an output format, not an input reader. PDF conversion requires another typesetting engine and is not exposed here. Conversion preserves content structure where possible, not exact pagination or complex Office formatting. Template and bibliography writer subsets are returned in their Operation descriptions.

Conversion/merge expose standalone output, table of contents and depth, section numbering, heading shifts, revision acceptance/rejection, line wrapping/endings, eight syntax highlighting themes, title/author/language/date metadata, matching Office reference files and local bibliography/CSL citation rendering. A format can ignore layout options it does not implement. `document.reference` creates a starting reference file; customize that file with an Office editor and pass it through `--reference`.

```sh
fnsh pkg add pandoc
fnsh document formats --json
fnsh document convert report.md --to docx --toc -o report.docx
fnsh document convert report.docx --to gfm -o report.md
fnsh document merge first.md --files second.md --to epub --title Book -o book.epub
fnsh document text report.docx -o report.txt
fnsh document inspect report.docx --json
fnsh document media report.docx -o embedded
fnsh document split report.md --level 1 --to gfm -o chapters
fnsh document reference --to docx -o reference.docx
fnsh document convert report.md --to docx --reference reference.docx -o styled.docx
fnsh document template --to html -o template.html
fnsh document bibliography references.bib --to csljson -o references.json
fnsh document convert article.md --to html --bibliography references.bib -o article.html
```

## Resource and output rules

- Each input or auxiliary file is at most 32 MiB; combined document inputs are at most 64 MiB. Office/EPUB containers allow at most 10,000 entries and 128 MiB declared expansion, rejecting traversal and links.
- Processing uses a private working directory, Pandoc's IO sandbox, a 512 MiB Haskell heap limit and a three-minute timeout. Arbitrary filters, scripts, custom writers and PDF engines are not exposed.
- Citation rendering accepts explicit local bibliography and independent CSL files. Remove `bibliography`, `csl` and `citation-abbreviations` metadata from source documents before enabling it; dependent CSL styles are rejected. This prevents citeproc from fetching implicit resources outside the reader/writer sandbox.
- External image files/URLs are not loaded implicitly. Use data URI images or existing embedded Office/EPUB media. Missing resources fail the operation instead of silently losing images. Generated HTML is **not** sanitized; sanitize untrusted HTML before serving it.
- File results are staged. Existing files require `--overwrite`; errors and cancellation preserve the previous destination. Output directories must be new. Media names are numbered, with their source names returned as data.
- Splitting returns at most 200 sections. It splits top-level headings at or above the selected level and includes any preamble as the first section. Output is Markdown, HTML, plain text, RST or LaTeX; embedded media are extracted separately with `document.media`.
- Structural inspection returns Pandoc metadata/AST counts, not page counts or visual layout. Empty media extraction reports an error because no files were produced.

## Reproduce acceptance

Run `./scripts/test-documents.ps1`. It installs the managed package and runs real file assertions, repository tests, vet/build, vulnerability scanning, generated Web contract checks, and direct/structured CLI round-trips. Successful runs write an acceptance JSON under the selected runtime home's `evidence` directory. The real suite exercises every Operation, all 25 document writers, Office/EPUB/HTML/JSON text round-trips, all reference/template variants, citation records, embedded PNG extraction, failure preservation, cancellation and sandbox network boundaries.

Upstream references: [Pandoc manual](https://pandoc.org/MANUAL.html), [pinned 3.11 release and artifact checksums](https://github.com/jgm/pandoc/releases/tag/3.11).
