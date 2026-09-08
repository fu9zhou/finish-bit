# Managed packages

FinishBit owns the versions of external tools used by its providers. Users install a logical package without selecting a version:

```bash
fnsh pkg add ffmpeg
```

The embedded registry pins the following runtime groups. Every artifact must match its fixed SHA-256 digest before extraction.

| Package | Version | Registered platforms |
| --- | --- | --- |
| FFmpeg, ffprobe | 6.1.1 | Windows x64, Linux x64/arm64, macOS x64/arm64 |
| pdfcpu | 0.15.0 | Windows x64, Linux x64/arm64, macOS x64/arm64 |
| Poppler | 26.07.0-0 | Windows x64 |
| 7-Zip Extra | 26.03 | Windows x64 |
| 7-Zip Full (Tesseract installer extraction) | 26.03 | Windows x64 |
| 7-Zip Bootstrap (official x86 7zr.exe) | 26.03 | Windows x64; Windows arm64 via OS x86 emulation |
| Tesseract + pinned English/Chinese fast models | 5.5.3-fast-8741641 | Windows x64 |
| Pandoc | 3.11 | Windows x64 |
| qsv | 22.0.1 | Windows x64 (MSVC build) |
| ImageMagick | 7.1.2-31 | Windows x64/arm64 |

Media assets come from `eugeneware/ffmpeg-static`, with an npmmirror transport fallback. pdfcpu and ImageMagick use upstream GitHub releases; Poppler uses the `oschwartz10612/poppler-windows` distribution. Platform registration is distinct from native acceptance testing; the complete new-provider suite has been run on Windows x64.

Tesseract uses the upstream 5.5.3 Windows installer, extracted as an archive by managed 7-Zip Full without running the installer. The English and simplified Chinese models remain pinned to tessdata_fast commit `87416418657359cb625c412a48b6e1d6d41c29bd`. Run `fnsh pkg add tesseract` with the updated CLI to install the registered version, or `fnsh pkg repair tesseract` to download and reinstall it.

## Installation guarantees

1. Downloads require HTTPS.
2. Content is written to an isolated staging directory.
3. SHA-256 is checked before decompression.
4. Decompressed output is bounded to 1 GiB.
5. Package metadata and executables activate only after all checks pass.
6. A failed installation removes its staging directory.

The installer supports raw executables, gzip, ZIP, tar.gz, tar.xz and 7z. Multi-file archives retain their runtime layout under `payload`; executable mappings reference exact relative paths. Extraction rejects traversal, links, special files, Windows reserved paths, duplicate file entries, more than 50000 entries and expanded sizes above 1 GiB. A bounded Windows retry handles temporary activation locks after executable extraction.

For 7z packages, installation reuses an already installed, current-version managed `7zip-full`, `7zip`, or `7zip-bootstrap`, excluding the package being installed. If none is available, it automatically installs `7zip-bootstrap` from the official fixed-version raw executable (589 KB); this breaks the cold-install extraction cycle without embedding a Go 7z decoder or distributing our own binaries. The bootstrap is also available through `fnsh pkg add 7zip-bootstrap`. Windows arm64 uses the OS x86 emulation documented by [Microsoft](https://learn.microsoft.com/en-us/windows/arm/apps-on-arm-x86-emulation); it has not been tested on physical arm64 hardware.

7z extraction first validates a bounded technical listing, then requests each exact filename through stdout with wildcard matching disabled. FinishBit creates the destination exclusively and caps the stream at its declared size; it checks the final size and the reader's exit status. This avoids giving the archive reader control over destination paths. Solid archives may be decoded repeatedly, trading installation speed for bounded writes. The pinned Tesseract NSIS path continues to use full 7-Zip and its existing listing/output validation.

Verified downloads are cached at `FINISHBIT_HOME/cache/<sha256>` and revalidated before reuse. Repair always obtains fresh bytes from a registered source. Removing a package leaves the download cache available for reinstallation.

An archive artifact may declare an exact `resources` list. In this mode only its mapped executables and listed resources are extracted, and every listed file must exist. The full downloaded archive still passes its pinned SHA-256 check. qsv uses this to retain its standalone MSVC executable and license/notices while omitting alternate Python/MCP builds and debugging symbols; selected output remains within the 1 GiB extraction bound. Other packages retain their full runtime layouts by default.

Data is stored below the operating system's user configuration directory under `finishbit`. Set `FINISHBIT_HOME` to isolate the runtime for CI or testing.

Downloaded FFmpeg binaries retain their own `GPL-3.0-or-later` licensing terms and are not embedded in FinishBit release archives.

## Lifecycle commands

| Command | Effect |
| --- | --- |
| `fnsh pkg info ffmpeg` | Show the pinned version, platform, license, and installation state |
| `fnsh pkg add ffmpeg` | Install the verified artifact when absent |
| `fnsh pkg repair ffmpeg` | Replace the installed artifact with a newly verified copy |
| `fnsh pkg remove ffmpeg` | Remove FinishBit's private installation |
| `fnsh pkg ls` | List installed packages |

FinishBit never mutates a system FFmpeg installation and does not use an unverified executable found on `PATH` as its managed runtime.

## Failure behavior

Network, unsupported-platform, checksum, decompression, and activation failures leave the previous active installation unchanged. A digest mismatch is an integrity failure and must not be bypassed. Use `fnsh pkg repair ffmpeg` for a corrupted or incomplete local installation.

Mirrors are transport alternatives for the same pinned bytes, not independent package sources. Adding or updating a registry artifact requires review of its upstream provenance, fixed digest, license, platform coverage, and download bounds.

See [installation](installation.md) for the user-facing data directory and [security policy](../SECURITY.md) for the managed-runtime trust boundary.
