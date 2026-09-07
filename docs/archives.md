# Archive workflows (unreleased)

Install the pinned, SHA-256-verified 7-Zip Extra 26.03 runtime with `fnsh pkg add 7zip`. The registered and locally tested target is Windows x64. The standalone `7za.exe` does not require a system installation or modify PATH. It is separately downloaded with its LGPL and bundled component notices.

| Operation | Behavior |
| --- | --- |
| `archive.formats` | Return supported read, write, encryption, mutation and volume formats |
| `archive.list` | Return validated entry paths, sizes, directory and encryption flags |
| `archive.test` | Verify decompression/CRC and entry boundaries |
| `archive.create` | Pack 1–64 local files/directories with compression level 0–9 |
| `archive.extract` | Extract everything or exact selected files/directories into a new directory |
| `archive.update` | Add/replace local files and directories in a new ZIP/7z output |
| `archive.remove` | Remove exact file/directory selections in a new ZIP/7z output |
| `archive.rename` | Rename one file entry in a new ZIP/7z output |
| `archive.repack` | Convert/recompress archives; retain, change or explicitly remove encryption |
| `archive.volumes` | Pack ZIP/7z into numbered volumes in a new directory |

## Format matrix

| Format | Read/test/extract | Create | Update/remove/rename | Encrypt | Volumes |
| --- | --- | --- | --- | --- | --- |
| 7z | Yes | Yes | Yes | AES, including headers | Yes |
| ZIP | Yes | Yes | Yes | AES-256 | Yes |
| TAR | Yes | Yes | No | No | No |
| GZIP, BZIP2, XZ | Yes | One file | No | No | No |
| tar.gz, tar.xz | Yes, including inner TAR | Yes | No | No | No |
| CAB, LZMA, Zstandard | Yes, available in pinned runtime | No | No | No | No |
| RAR | No in the selected standalone build | No | No | No | No |

The eight write formats have real create/list/test/extract round-trip coverage. The additional read-only formats are advertised by the pinned runtime; this acceptance corpus does not include CAB/LZMA/Zstandard fixtures. The `auto` input format recognizes standard extensions and `.001` volumes; use `--format` for ambiguous names. Generic gzip/bzip2/xz/lzma/zstd streams are materialized as a single file named after the input with its final extension removed. tar.gz/tar.xz are expanded as a directory tree.

## Examples

```sh
fnsh pkg add 7zip
fnsh archive create reports --files notes.txt --format zip -o reports.zip
fnsh archive list reports.zip --json
fnsh archive test reports.zip --json
fnsh archive extract reports.zip --entries reports/summary.txt -o extracted
fnsh archive update reports.zip --files reports -o updated.zip
fnsh archive rename updated.zip --entry notes.txt --name notes/archive.txt -o renamed.zip
fnsh archive remove renamed.zip --entries reports -o remaining.zip
fnsh archive repack reports.zip --to tar.xz -o reports.tar.xz
fnsh archive volumes reports --format 7z --volume-mib 10 -o parts
fnsh archive extract parts/archive.7z.001 -o restored
```

Selection uses exact, case-sensitive archive entry paths. Directory selections include descendants; wildcard patterns and `@listfiles` are not exposed. Updating uses input basenames as archive roots, so passing an updated `reports` directory matches entries below `reports/`. Original archives remain unchanged unless their path is explicitly selected as the output with `--overwrite`. Missing selections and rename collisions fail.

For encryption, put the password in an environment variable and use its name:

```powershell
$env:ARCHIVE_PASSWORD='your-password'
fnsh archive create reports --format 7z --password-env ARCHIVE_PASSWORD -o private.7z
fnsh archive extract private.7z --password-env ARCHIVE_PASSWORD -o restored
fnsh archive repack private.7z --password-env ARCHIVE_PASSWORD --drop-password --to tar -o plain.tar
```

`archive.repack --new-password-env OTHER_VARIABLE` changes the output password. Without it, encryption is retained; `--drop-password` explicitly removes it. ZIP encrypts content but exposes entry names. 7z also encrypts headers. Password values are not in the FinishBit request contract, but 7-Zip receives them through its child-process arguments, which privileged local process inspection may observe.

## Limits and publication

- Maximum 64 MiB per expanded file, 256 MiB total archive/input tree, and 10,000 entries. Input collections allow up to 64 roots. Declared sizes and actual streamed bytes are both checked.
- Operations have a five-minute timeout. Bounded streaming handles files larger than the ordinary 16 MiB result capture limit without loading them into memory.
- Extraction validates traversal, absolute paths, Windows reserved names, alternate streams, case collisions, file/directory conflicts, links and special files. It writes each approved file through bounded stdout into a private tree, rather than letting archive paths choose filesystem destinations.
- New output directories are required. Failed generation and cancellation do not publish results. File outputs are staged and require `--overwrite` to replace existing files.
- Windows `Zone.Identifier` origin metadata is retained when extracting or deriving another archive from a marked source. Archive entries cannot supply arbitrary alternate streams.
- Volumes range from 1–256 MiB, must stay adjacent, and are opened at `.001`. The combined volume input is bounded to 256 MiB and fewer than 1,000 parts. Repack volumes to one archive before mutation. Missing/corrupt parts fail integrity processing.
- File modification times, permissions, links and arbitrary archive metadata are not preserved by extraction/repacking; file contents and the supported directory tree are the contract. Empty directories are supported.

## Reproduce acceptance

Run `./scripts/test-archives.ps1`. It installs the managed runtime, exercises all ten Operations with real archives, checks all eight writable formats, Unicode paths, selected extraction, AES encryption, wrong/missing passwords, explicit decryption, ZIP/7z mutation, volume reassembly, origin marks, cancellation, existing-output preservation, traversal and expanded-size limits. A 17 MiB file verifies streaming beyond the normal capture limit; SHA-256 compares round-trip bytes. The script also runs repository tests, vet/build, vulnerability scanning, Web catalog drift checks and direct/structured CLI assertions, then records an acceptance JSON under the runtime home's `evidence` directory.

Upstream references: [official downloads](https://www.7-zip.org/download.html), [pinned release and checksums](https://github.com/ip7z/7zip/releases/tag/26.03). The installed `readme.txt` and `7-zip.chm` describe the standalone build and command semantics.
