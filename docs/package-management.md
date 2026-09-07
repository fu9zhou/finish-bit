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
| Pandoc | 3.11 | Windows x64 |
| qsv | 22.0.1 | Windows x64 (MSVC build) |
| ImageMagick | 7.1.2-31 | Windows x64/arm64 |

Media assets come from `eugeneware/ffmpeg-static`, with an npmmirror transport fallback. pdfcpu and ImageMagick use upstream GitHub releases; Poppler uses the `oschwartz10612/poppler-windows` distribution. Platform registration is distinct from native acceptance testing; the complete new-provider suite has been run on Windows x64.

## Installation guarantees

1. Downloads require HTTPS.
2. Content is written to an isolated staging directory.
3. SHA-256 is checked before decompression.
4. Decompressed output is bounded to 1 GiB.
5. Package metadata and executables activate only after all checks pass.
6. A failed installation removes its staging directory.

The installer supports raw executables, gzip, ZIP, tar.gz, tar.xz and 7z. Multi-file archives retain their runtime layout under `payload`; executable mappings reference exact relative paths. Extraction rejects traversal, links, special files, Windows reserved paths, duplicate file entries, more than 50000 entries and expanded sizes above 1 GiB. A bounded Windows retry handles temporary activation locks after executable extraction.

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
