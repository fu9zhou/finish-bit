param(
    [string]$RuntimeHome = (Join-Path $PSScriptRoot '../.finishbit-test-batch-a'),
    [string]$GoToolchain = 'go1.26.8'
)

$ErrorActionPreference = 'Stop'
$taskRepo = Split-Path $PSScriptRoot -Parent
$taskVariables = @('GOTOOLCHAIN', 'FINISHBIT_HOME', 'FINISHBIT_TEST_HOME', 'FINISHBIT_TEST_FFMPEG', 'FINISHBIT_TEST_FFPROBE')
$taskPrevious = @{}
foreach ($name in $taskVariables) { $taskPrevious[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }

function Invoke-Checked {
    param([string]$Program, [string[]]$Arguments)
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed with exit code $LASTEXITCODE" }
}

function Get-RuntimeExecutable {
    param([string]$Package, [string]$Name)
    $info = Invoke-Checked $taskBinary @('pkg', 'info', $Package, '--json') | ConvertFrom-Json
    if (-not $info.installed.executables.$Name) { throw "Missing managed executable: $Package/$Name" }
    return Join-Path $env:FINISHBIT_HOME "packages/$Package/$($info.installed.version)/$($info.installed.executables.$Name)"
}

Push-Location $taskRepo
try {
    if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) { throw 'Batch A currently requires Windows (Poppler and ImageMagick packages).' }
    $env:GOTOOLCHAIN = $GoToolchain
    $env:FINISHBIT_HOME = [IO.Path]::GetFullPath($RuntimeHome)
    $env:FINISHBIT_TEST_HOME = $env:FINISHBIT_HOME
    $taskBinary = Join-Path $taskRepo 'fnsh.exe'
    Invoke-Checked go @('build', '-o', $taskBinary, './cmd/fnsh')
    foreach ($package in @('ffmpeg', 'ffprobe', 'pdfcpu', 'poppler', 'imagemagick')) {
        Invoke-Checked $taskBinary @('pkg', 'add', $package)
    }
    $env:FINISHBIT_TEST_FFMPEG = Get-RuntimeExecutable ffmpeg ffmpeg
    $env:FINISHBIT_TEST_FFPROBE = Get-RuntimeExecutable ffprobe ffprobe
    $taskMagick = Get-RuntimeExecutable imagemagick magick

    # Non-cached tests enable every real-provider acceptance suite before the next batch.
    Invoke-Checked go @('test', '-count=1', './...')
    Invoke-Checked go @('vet', './...')
    Invoke-Checked go @('build', './cmd/fnsh')
    Invoke-Checked go @('run', 'golang.org/x/vuln/cmd/govulncheck@v1.7.0', './...')

    $taskEvidence = Join-Path $env:FINISHBIT_HOME ('evidence/' + [Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $taskEvidence -Force | Out-Null
    $taskInput = Join-Path $taskEvidence 'source.png'
    $taskCrop = Join-Path $taskEvidence 'crop.png'
    $taskPDF = Join-Path $taskEvidence 'images.pdf'
    $taskStamped = Join-Path $taskEvidence 'stamped.pdf'
    Invoke-Checked $taskMagick @('-size', '160x120', 'gradient:red-blue', $taskInput)
    $taskRequest = Join-Path $taskEvidence 'request.json'
    @{ operation = 'image.crop'; inputs = @($taskInput); options = @{ width = 80; height = 60; output = $taskCrop } } |
        ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $taskRequest -Encoding utf8
    Invoke-Checked $taskBinary @('run', '--request', $taskRequest, '--json')
    $taskDimensions = Invoke-Checked $taskMagick @('identify', '-format', '%wx%h', $taskCrop)
    if ($taskDimensions -ne '80x60') { throw "Wrong CLI crop dimensions: $taskDimensions" }
    Invoke-Checked $taskBinary @('pdf', 'from-images', $taskInput, '--output', $taskPDF, '--json')
    Invoke-Checked $taskBinary @('pdf', 'stamp', $taskPDF, '--text', 'ACCEPTED', '--size', '10', '--output', $taskStamped, '--json')
    $taskText = Invoke-Checked $taskBinary @('pdf', 'extract-text', $taskStamped, '--json')
    if (($taskText -join "`n") -notmatch 'ACCEPTED') { throw 'CLI PDF round trip lost stamp text' }
    $taskCatalog = @(Invoke-Checked $taskBinary @('capabilities', '--json') | ConvertFrom-Json)
    if ($taskCatalog.Count -lt 123) { throw "Incomplete catalog: $($taskCatalog.Count)" }
    $taskDoctor = @(Invoke-Checked $taskBinary @('doctor', '--json') | ConvertFrom-Json)
    if (@($taskDoctor | Where-Object { -not $_.ok }).Count) { throw 'Managed dependency doctor failed' }
    $taskReport = Join-Path $taskEvidence 'acceptance.json'
    @{
        completedAt = [DateTime]::UtcNow.ToString('o')
        platform = 'windows'
        toolchain = (Invoke-Checked go @('version'))
        operations = $taskCatalog.Count
        checks = @('go test -count=1 ./... with real providers', 'go vet ./...', 'go build ./cmd/fnsh', 'govulncheck', 'structured CLI crop dimensions', 'CLI PDF text round trip', 'managed dependency doctor')
        packages = $taskDoctor
    } | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $taskReport -Encoding utf8
    Write-Output "Batch A accepted. Evidence: $taskReport"
}
finally {
    foreach ($name in $taskVariables) { [Environment]::SetEnvironmentVariable($name, $taskPrevious[$name], 'Process') }
    Pop-Location
}
