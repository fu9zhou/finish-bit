# Third-party notices for the local toolbox

FinishBit's own license remains AGPL-3.0. The following additions retain their upstream licenses. Copies are included under `third_party/licenses/` and in release archives; exact Go dependency versions and checksums are in `go.mod` and `go.sum`.

| Component | Version | License / attribution |
| --- | --- | --- |
| [gozxing](https://github.com/makiuchi-d/gozxing) | v0.1.1 | MIT; derived ZXing code Apache-2.0, both texts in upstream LICENSE |
| [go-pinyin](https://github.com/mozillazg/go-pinyin) | v0.21.0 | MIT, mozillazg |
| [pinyin-data](https://github.com/mozillazg/pinyin-data/tree/v0.15.0) | v0.15.0, embedded by go-pinyin | MIT, mozillazg |
| [OpenCC](https://github.com/BYVoid/OpenCC/tree/ver.1.1.7) | ver.1.1.7, commit e5d6c5f1b78e28a5797e7ad3ede3513314e544b7 | Apache-2.0; ST/TS character and phrase dictionaries only; original files retained. FinishBit's trie algorithm is separate from OpenCC's conversion engine. |
| [YAML for Go](https://github.com/yaml/go-yaml) | v3.0.5 | Apache-2.0, MIT-derived portions noted in source; LICENSE and NOTICE retained |
| [goldmark](https://github.com/yuin/goldmark) | v1.8.6 | MIT |
| [lunar-go calendar data provenance](https://github.com/6tail/lunar-go) | Results evaluated with v1.4.6 | MIT attribution retained; fixed month-length data only, no Go module/runtime dependency. See [data notes](internal/localtools/data/lunar/README.md). |
| [cron](https://github.com/robfig/cron) | v3.0.1 | MIT |
| [go-toml](https://github.com/pelletier/go-toml) | v2.4.3 | MIT |
| [xerrors](https://go.googlesource.com/xerrors) | v0.0.0-20200804184101-5ec99f83aff1 | BSD-3-Clause |

Go's embedded `time/tzdata` uses the toolchain's IANA timezone data. Toolchain licensing remains in the Go distribution. The Go sevenzip decoder and its exclusive transitive modules have been removed; current module dependencies are recorded in `go.mod`.

Optional runtime downloads are kept separate from the CLI binary. Tesseract 5.5.3 and tessdata_fast use Apache-2.0; their distributions retain bundled dependency licenses and a separate model license is downloaded to the private package. The fixed Tesseract installer is unpacked as data, never executed. 7-Zip 26.03 is used only for extraction and retains its LGPL/BSD notices and unRAR restriction notice in its original distribution. These runtimes are not relicensed as FinishBit code.

Cold 7z installation downloads the upstream x86 `7zr.exe` directly from the fixed [7-Zip 26.03 release](https://github.com/ip7z/7zip/releases/tag/26.03), with SHA-256 `ad4c82fadcbdf93c03b4fc440f300509c7d60c5c2f4d183e35d9d70d6957037d`. Upstream [licensing information](https://www.7-zip.org/license.txt) and [corresponding source](https://github.com/ip7z/7zip/tree/26.03) remain available from 7-Zip. FinishBit does not rebuild, repackage, or host this executable.

Windows System.Speech is an optional operating-system API; no proprietary engine or voice files are copied or redistributed. The implementation uses the user's existing installed voices. No website artwork, proprietary dictionaries, AI weights or Office runtime are included.

See [dependency provenance and sizes](docs/research/lightweight-tool-dependencies-2026-09-07.md) for official sources and pinned download hashes. This notice covers this toolbox expansion, not a replacement for the upstream notices shipped with existing managed packages.
