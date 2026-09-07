# PDF operations

The PDF capability group uses two independently managed runtimes. **pdfcpu 0.15.0** edits PDF structure, pages, forms, attachments and bookmarks. **Poppler 26.07.0-0** reads existing text and renders pages. OCR is a separate capability and is not implemented by these operations.

## Installation

```sh
fnsh pkg add pdfcpu
fnsh pkg add poppler
fnsh describe pdf.merge --json
```

pdfcpu has pinned archives for Windows x64, Linux x64/arm64 and macOS x64/arm64. The current Poppler registry contains a Windows x64 distribution only; other platforms receive `unsupported_platform` when installing it. Windows x64 is the platform exercised by the real acceptance suite. Archives retain runtime DLLs and license files under their original layout. No system PATH fallback is used.

## Catalog

| Group | Operations |
| --- | --- |
| Inspection | `pdf.info`, `pdf.validate` |
| Pages | `pdf.merge`, `pdf.split`, `pdf.select-pages`, `pdf.remove-pages`, `pdf.insert-pages`, `pdf.rotate`, `pdf.crop`, `pdf.nup` |
| Content and storage | `pdf.watermark`, `pdf.stamp`, `pdf.remove-watermark`, `pdf.remove-stamp`, `pdf.from-images`, `pdf.optimize`, `pdf.encrypt`, `pdf.decrypt` |
| Attachments | `pdf.attachments`, `pdf.attach`, `pdf.extract-attachments`, `pdf.remove-attachments` |
| Forms | `pdf.form-fields`, `pdf.form-export`, `pdf.form-fill`, `pdf.form-multifill`, `pdf.form-reset`, `pdf.form-lock`, `pdf.form-unlock` |
| Bookmarks and keywords | `pdf.bookmarks-export`, `pdf.bookmarks-import`, `pdf.bookmarks-remove`, `pdf.keywords`, `pdf.keywords-add`, `pdf.keywords-remove` |
| Text and rendering | `pdf.extract-text`, `pdf.text-boxes`, `pdf.render`, `pdf.extract-images`, `pdf.fonts`, `pdf.to-html`, `pdf.to-ps` |

## Examples

```sh
fnsh pdf merge cover.pdf --files report.pdf -o merged.pdf
fnsh pdf select-pages report.pdf --pages 3,1,2 -o reordered.pdf
fnsh pdf split report.pdf --span 2 -o split-pages
fnsh pdf stamp report.pdf --text APPROVED --opacity 60 -o stamped.pdf
fnsh pdf stamp report.pdf --mode image --asset seal.png -o sealed.pdf
fnsh pdf extract-text report.pdf --first 1 --last 5 -o text.txt
fnsh pdf extract-text article.pdf --unwrap -o article.txt
fnsh pdf render report.pdf --first 1 --last 3 --dpi 120 -o preview
fnsh pdf form-export form.pdf -o form.json
fnsh pdf form-fill form.pdf form.json -o filled.pdf
fnsh pdf form-multifill form.pdf records.json -o filled-forms
```

`pdf.select-pages` is both page extraction and explicit reordering; page expressions accept positive page numbers, closed ranges, `odd` and `even`. Insertion adds blank pages. Watermark/stamp modes accept text, an image or a PDF; a watermark sits behind existing content, while a stamp sits above it. Default text rendering uses Helvetica and is not a promise of arbitrary Unicode font coverage; an image/PDF stamp can supply pre-rendered typography. Removal recognizes pdfcpu watermark/stamp structures and does not promise removal of arbitrary page artwork. Cropping changes page boundaries and is not content redaction.

## Contract boundaries

- Every input is a regular local file. Up to 128 input paths can be supplied. Single-file outputs require the appropriate extension and are staged; existing destinations need `--overwrite`. Directory outputs must be new. Failed generation does not replace an existing output. Input files are never edited in place by attachment commands; a temporary copy is used.
- pdfcpu runs offline with user configuration disabled. The Provider sets force only on its private staging output. No link-checking network validation is enabled. Structural optimization removes redundant resources; it is not a target-size compressor or scan downsampler.
- `pdf.encrypt` uses AES-256 with nonempty owner/user passwords and explicit permissions. `pdf.decrypt` needs the supplied password. For editing encrypted documents, decrypt into a separate file first. Editing signed PDFs can invalidate signatures; this group does not implement signing or a trust-management workflow.
- Form JSON follows the pinned pdfcpu schema and can be obtained with `pdf.form-export`. Multifill accepts that schema with multiple `forms` entries, or the upstream field-ID CSV format. This is AcroForm processing, not arbitrary visual document editing or XFA compatibility. Bookmark import replaces the existing bookmark tree.
- PDF text extraction reads existing text. `--unwrap` joins visual line breaks inside detected paragraphs while retaining headings, lists, blank-line paragraph boundaries and page breaks; it is heuristic and cannot perfectly reconstruct tables or distinguish every discretionary hyphen. `--layout` preserves physical layout and cannot be combined with `--unwrap`. `pdf.text-boxes` returns page dimensions and word rectangles in PDF points measured from the top left. Reading order depends on the document. A scan without text can legitimately yield no words.
- Rendering and directory exports require an explicit range of at most 200 pages. Rendering accepts 36–300 DPI. Image extraction returns embedded image assets, potentially including masks and auxiliary files; it does not recover tables or logical figures. HTML conversion produces a bundle of local files and does not promise source-document layout equivalence.
- Inspection data includes pdfcpu JSON where available. Font, attachment and keyword listings are explicitly returned as human-readable `report` fields from the pinned engine; they are not parsed as stable table schemas. Text results and other process stdout are limited to 16 MiB.
- Windows output paths with non-ASCII characters are handled using a process working directory and a relative export prefix for affected Poppler utilities. Attachment names are preserved through isolated one-file directories to avoid interpreting user basenames as glob patterns.

## Real acceptance tests

Set `FINISHBIT_TEST_HOME` to the runtime root containing both installed packages, then run:

```sh
go test ./internal/pdf -run TestPDFWithManagedTools -count=1 -v
```

The suite creates a two-page text PDF with an editable form, an image and a text attachment. It runs every PDF Operation against the real managed binaries, validates rewritten PDFs, checks page counts, verifies extracted text/coordinates, renders 300×300 pages at 72 DPI, checks attachment bytes and form values, and covers Unicode/special-character paths and failed-output preservation.
