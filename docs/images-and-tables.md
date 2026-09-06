# Images and tabular data

Available since v0.1.3. All six Operations run in the Go binary without an external runtime and support ordinary CLI or structured requests.

## Images

```bash
fnsh image info photo.jpg --json
fnsh image resize photo.jpg --width 1200 -o smaller.jpg
fnsh image resize photo.png --width 800 --height 600 -o thumbnail.png
fnsh image convert photo.png --format jpeg --quality 90 -o photo.jpg
```

- Input is an existing PNG or JPEG file. Output supports PNG and JPEG; format defaults to the destination extension (`.png`, `.jpg`, `.jpeg`).
- Resize preserves aspect ratio and fits within the supplied width/height. Supply at least one dimension. Small images are not enlarged unless `--upscale` is set.
- PNG retains transparency. JPEG composites transparency onto white. `--quality` ranges from 1 to 100 and defaults to 90 for JPEG.
- Operations use encoded pixel orientation; they do not apply EXIF orientation. Conversion and resizing do not copy EXIF, ICC profiles, or other metadata. Animated and other image formats are outside this version's scope.
- Input files are limited to 64 MiB. Decoded and output images must have at most 25 million pixels and at most 32768 pixels along either dimension.
- `image.info` reports encoded width, height, format and file byte count. It inspects the header; conversion/resizing also validate the pixel stream.

## CSV and JSON

```bash
fnsh csv info people.csv --json
fnsh csv to-json people.csv -o people.json
fnsh csv to-json people.tsv --delimiter "\t"
fnsh json to-csv people.json --columns id --columns name -o people.csv
```

For TSV, `--delimiter "\t"` accepts the literal escape, and an actual tab character is also accepted. In a structured JSON request use `"delimiter": "\t"`.

CSV input is UTF-8; an initial UTF-8 BOM is accepted. The first row is a header with unique non-empty names. Quoted separators, escaped quotes and multiline cells are supported. Inconsistent row widths and malformed quoting are rejected. Empty physical lines are skipped according to the CSV reader's behavior. `csv.info` reports `columns`, `column_count`, and the number of data rows as `row_count`.

`csv.to-json` emits an array of objects and preserves all values as strings: `001` remains `"001"`. It does not infer dates, numbers or booleans.

`json.to-csv` accepts an array of objects containing strings, numbers, booleans or null. Nested objects and arrays are rejected. Without `--columns`, the union of keys is sorted for stable column order. Repeated `--columns` selects fields and their order, omitting unselected fields. Missing values and null produce empty cells. An empty array requires explicit columns. String content is preserved, including spreadsheet formula text.

Table input is limited to 8 MiB, 100000 data rows, 1000 columns and 1000000 cells (including the header in CSV/output accounting). Serialized output is limited to 64 MiB, including expansion from repeated JSON field names.

## Output and automation

The new file-producing Operations refuse to replace an existing destination unless `--overwrite` is supplied. Output is fully written to a temporary file before publication. This policy applies to the new image/table Operations; older Operations retain their existing behavior. Exclusive publication requires a filesystem that supports hard links; unsupported filesystems report an error without overwriting the destination.

```json
{
  "operation": "image.resize",
  "inputs": ["photo.png"],
  "options": {"width": 800, "height": 600, "output": "thumbnail.png"}
}
```

Run this request with `fnsh run --request request.json --json`. Image results return paths in `outputs` and dimensions/format in `data`. Table conversions return `data.text`, or `outputs` when an output file is requested. Use `fnsh describe <operation> --json` for parameter contracts.
