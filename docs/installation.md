# Installation

## Supported platforms

FinishBit builds with `CGO_ENABLED=0` for these release targets:

| Operating system | amd64 | arm64 |
| --- | :---: | :---: |
| Windows | Yes | Yes |
| Linux | Yes | Yes |
| macOS | Yes | Yes |

Go 1.24 or later is required only when building or installing from source. Managed Operations may have additional runtime requirements; `fnsh describe <operation>` lists them.

## Install from source

```bash
go install github.com/fu9zhou/finish-bit/cmd/fnsh@latest
```

Ensure the Go binary directory is on `PATH`, then verify the installation:

```bash
fnsh version
fnsh doctor
```

Use an explicit version instead of `@latest` when reproducibility matters.

## Install a release archive

Download the archive and `checksums.txt` for the same version from [GitHub Releases](https://github.com/fu9zhou/finish-bit/releases). Verify its SHA-256 digest before extracting `fnsh` (`fnsh.exe` on Windows) into a directory on `PATH`.

Release archive names follow this pattern:

```text
finish-bit_<version>_<os>_<architecture>.<archive>
```

Windows archives use ZIP; Linux and macOS archives use `tar.gz`. The amd64 archive label is `x86_64`.

## Installer scripts

The repository contains checksum-verifying installers for published releases. Review the script before running it.

PowerShell:

```powershell
.\scripts\install.ps1
```

POSIX shell on Linux or macOS:

```sh
./scripts/install.sh
```

Set `FINISHBIT_VERSION` to install a specific tag and `FINISHBIT_INSTALL_DIR` to choose the binary directory. The default locations are `%LOCALAPPDATA%\FinishBit\bin` on Windows and `$HOME/.local/bin` on Linux/macOS.

## Runtime data

FinishBit stores managed packages, installed extensions, and state below the operating system's user configuration directory under `finishbit`. Set `FINISHBIT_HOME` to use an isolated data directory, which is useful in CI and disposable test environments.

Installing FinishBit does not install FFmpeg automatically. Add it only when an Operation requires it:

```bash
fnsh pkg add ffmpeg
fnsh doctor
```

## Upgrade

- Source installation: rerun `go install` with the desired version.
- Release installation: install the new archive over the existing binary.
- Managed package: `fnsh pkg repair ffmpeg` reinstalls the registry-pinned artifact.

Read the [changelog](../CHANGELOG.md) and [compatibility policy](compatibility.md) before upgrading across pre-1.0 minor versions. Runtime data is kept separately from the binary.

## Uninstall

1. Remove `fnsh` or `fnsh.exe` from its installation directory.
2. If you also want to remove downloaded packages and extensions, locate the data directory with your OS conventions or `FINISHBIT_HOME`, verify that it belongs to FinishBit, and delete that directory explicitly.

Removing only the executable leaves runtime data intact for a later reinstall.
