# Table cleaning and analysis

This v0.1.4 capability group adds 27 Operations through managed **qsv 22.0.1**. The existing 0.1.3 CSV inspection and CSV/JSON conversion remain available without qsv. These new Operations use the shared application service and require local input files.

```sh
fnsh pkg add qsv
fnsh csv select orders.csv --columns order_id --columns amount -o selected.csv
fnsh csv filter orders.csv --columns status --pattern paid --mode exact -o paid.csv
fnsh csv deduplicate orders.csv --columns order_id --keep first -o unique.csv
fnsh csv join orders.csv customers.csv --left-keys customer_id --right-keys id --kind left -o joined.csv
fnsh csv stats orders.csv --columns amount -o statistics.csv
fnsh csv split orders.csv --rows 500 -o chunks
fnsh csv validate orders.csv --schema rules.json --json
```

## Catalog

| Group | Operations |
| --- | --- |
| Columns and rows | `csv.select`, `csv.rename`, `csv.filter`, `csv.sort`, `csv.deduplicate`, `csv.slice`, `csv.reverse` |
| Multiple tables | `csv.concat`, `csv.join`, `csv.diff` |
| Statistics and sampling | `csv.frequency`, `csv.stats`, `csv.sample`, `csv.shuffle` |
| Reshaping | `csv.transpose`, `csv.melt`, `csv.explode`, `csv.split`, `csv.partition` |
| Cleaning | `csv.fill`, `csv.replace`, `csv.transform`, `csv.format` |
| Rules and interchange | `csv.schema`, `csv.validate`, `csv.to-jsonl`, `jsonl.to-csv` |

## Contracts

- Column options take exact names, including punctuation and Unicode; repeat a string-list option for multiple names. Names are resolved to column indices before qsv is called. Arbitrary qsv selectors, scripts, shell commands and remote inputs are not accepted.
- Inputs are UTF-8, with an optional BOM, a header row and consistent widths. Delimiters are single ASCII characters (including tab). Headers must be unique and nonempty. Each file is limited to 16 MiB, 100,000 data rows and 1,000,000 cells; combined input bytes are limited to 32 MiB and input files to 32. Result capture is limited to 16 MiB and each operation has a three-minute deadline. Validation reports malformed CSV rather than silently repairing it.
- Input files are copied into a private working directory. Runtime caches and intermediates stay there and are removed afterward. Caller `QSV_*` configuration is isolated. The provider never modifies source files and does not expose qsv network, AI or script execution commands.
- File outputs require `--output`; existing destinations require `--overwrite`. The completed result is staged before publication. Split and partition require a new directory, limited to 500 output files. Partition keys must be 1–80 letters/numbers/underscores/hyphens; names that collide by case are rejected.
- Selection and ordinary editing preserve cell strings, including leading zeros. Numeric sorting is explicit. JSONL export infers scalar types; this differs from the existing core `csv.to-json` string-preserving contract. JSONL import requires scalar values and identical keys on each nonempty line; nested structures and changing keys are rejected.
- Deduplication produces sorted key order, with explicit `keep=first|last`. Default first retention uses stable sorting followed by streaming deduplication. Join follows qsv's trimming of leading/trailing whitespace in keys; null matching and case-insensitive matching are explicit options. Diff requires unique keys and identical ordered headers, and emits qsv's change-marker column.
- Concatenation offers identical-header rows, union-by-header rows (`rowskey`), and positional columns. Padding applies only to column concatenation. Join and column concatenation disambiguate duplicate output headers with `_2`, `_3`, etc., skipping existing names so results remain usable by subsequent Operations. Sample/shuffle seeds make repeated runs reproducible for the pinned runtime.
- Transpose includes the original header. Melt follows qsv's `field,attribute,value` form: the unselected identifier cells are joined with `|`, and empty value cells are omitted. It is not a lossless round trip when identifiers themselves contain `|`; retain the source file.
- Frequency preserves surrounding whitespace by default. Statistics include types, counts and numeric aggregates; optional extended statistics include median, quartiles and cardinality. All are emitted as CSV for further reuse.
- Fill supports previous value, first value or an explicit replacement, with optional leading backfill. Replacement is literal by default; regex mode uses Rust regex syntax and capture substitutions. Transform exposes trim/ltrim/rtrim, case conversion, whitespace squeeze, titlecase, rounding and currency-number conversion, with finite option bounds.
- Schema inference describes the observed data, not a complete business specification. Validation accepts a local JSON Schema; remote/file references and dynamic-enum lookups are rejected. Failed validation returns a structured operation error and does not publish partial valid/invalid record files.

## Acceptance

Accepted on Windows x64 on 2026-09-07 with the checksum-verified managed MSVC runtime. The final gate also enabled the previous media/PDF/image real-runtime suites; all tests, vet, build, vulnerability scan, generated website check and direct/structured CLI assertions passed. The local report is under `.finishbit-test-batch-b/evidence/37d9871f4bc9462fa6a42e43dd316f58/acceptance.json`. CI jobs are configured but this local run does not claim a remote workflow or deployment has run.

`TestTablesWithManagedQSV` exercises every registered Operation against real CSV/JSONL files and inspects results with independent Go CSV/JSON readers. Set `FINISHBIT_TEST_QSV` to the managed executable and run:

```sh
go test ./internal/tabular -count=1 -v
```

On Windows x64, `./scripts/test-tables.ps1` reproduces installation, real-provider tests, repository checks, vulnerability scanning, website catalog verification, and direct/structured CLI checks. A successful run writes an acceptance report under `.finishbit-test-batch-b/evidence/`.

Tests check exact values, leading zeros, row/column counts, join results, numeric mean/frequencies, duplicate retention, seeded sampling, schema failures, Unicode/special-character paths, cancellation and existing-output preservation. Unit tests cover malformed input and invalid requests without installing qsv.

The Windows x64 distribution is the first deployment target for this group. Registering other platforms requires verifying each build's enabled commands and real output behavior; upstream platform availability alone does not establish FinishBit support.

Sources: [qsv 22.0.1 release](https://github.com/dathere/qsv/releases/tag/22.0.1), [command source and help](https://github.com/dathere/qsv/tree/22.0.1/src/cmd). Excel editing remains deferred: [Excelize's upstream advisory](https://github.com/qax-os/excelize/security/advisories/GHSA-g27h-8qhm-6pff) reports a parsing panic affecting the latest stable 2.11.0 without a patched release as checked on 2026-09-07. Excelize has not been added to this repository.
