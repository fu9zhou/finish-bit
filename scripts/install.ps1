[CmdletBinding()]
param(
    [string]$Version = $env:FINISHBIT_VERSION,
    [string]$InstallDirectory = $env:FINISHBIT_INSTALL_DIR
)

$ErrorActionPreference = 'Stop'
$repository = 'fu9zhou/finish-bit'
if (-not $InstallDirectory) {
    $InstallDirectory = Join-Path $env:LOCALAPPDATA 'FinishBit\bin'
}

if (-not $Version) {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repository/releases/latest"
    $Version = $release.tag_name
}
if (-not $Version) { throw 'Could not resolve a release version.' }

$architecture = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    'X64' { 'x86_64' }
    'Arm64' { 'arm64' }
    default { throw "Unsupported architecture: $_" }
}
$archive = "finish-bit_$($Version.TrimStart('v'))_windows_$architecture.zip"
$base = "https://github.com/$repository/releases/download/$Version"
$temporary = Join-Path ([System.IO.Path]::GetTempPath()) ("finishbit-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $temporary | Out-Null

try {
    $archivePath = Join-Path $temporary $archive
    $checksumsPath = Join-Path $temporary 'checksums.txt'
    Invoke-WebRequest -Uri "$base/$archive" -OutFile $archivePath
    Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $checksumsPath
    $checksumLine = Get-Content -LiteralPath $checksumsPath | Where-Object { $_ -match "\s$([regex]::Escape($archive))$" } | Select-Object -First 1
    if (-not $checksumLine) { throw 'Archive is missing from checksums.txt.' }
    $expected = ($checksumLine -split '\s+')[0].ToLowerInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $archivePath).Hash.ToLowerInvariant()
    if ($expected -ne $actual) { throw 'Archive checksum verification failed.' }

    Expand-Archive -LiteralPath $archivePath -DestinationPath $temporary -Force
    New-Item -ItemType Directory -Force -Path $InstallDirectory | Out-Null
    Copy-Item -LiteralPath (Join-Path $temporary 'fnsh.exe') -Destination (Join-Path $InstallDirectory 'fnsh.exe') -Force
    Write-Output "Installed fnsh $Version to $InstallDirectory\fnsh.exe"
} finally {
    if (Test-Path -LiteralPath $temporary) { Remove-Item -LiteralPath $temporary -Recurse -Force }
}
