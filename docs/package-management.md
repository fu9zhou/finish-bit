# Managed packages

FinishBit owns the versions of external tools used by its providers. Users install a logical package without selecting a version:

```bash
fnsh pkg add ffmpeg
```

The embedded registry pins FFmpeg 6.1.1 artifacts for Windows amd64, Linux amd64/arm64, and macOS amd64/arm64. Assets come from the trusted `eugeneware/ffmpeg-static` GitHub release, with an npmmirror transport fallback for restricted networks. Every source must match the same fixed SHA-256 digest before extraction.

## Installation guarantees

1. Downloads require HTTPS.
2. Content is written to an isolated staging directory.
3. SHA-256 is checked before decompression.
4. Decompressed output is bounded to 1 GiB.
5. Package metadata and executables activate only after all checks pass.
6. A failed installation removes its staging directory.

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
