$ErrorActionPreference = 'Stop'
$yamlExperimentDir = Join-Path $env:TEMP ('finishbit-yaml-compat-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $yamlExperimentDir | Out-Null
foreach ($name in @('main.go', 'bridge.go', 'reference.go', 'go.mod')) {
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot ($name + '.txt')) -Destination (Join-Path $yamlExperimentDir $name)
}
Push-Location $yamlExperimentDir
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw 'go mod tidy failed' }
    go run . | Set-Content -LiteralPath results.jsonl
    if ($LASTEXITCODE -ne 0) { throw 'go run failed' }
    Get-Content results.jsonl | ConvertFrom-Json |
        Where-Object case | Select-Object case, candidate_matches, old_error, candidate_error |
        Format-Table -AutoSize
    Write-Output "Full results: $yamlExperimentDir/results.jsonl"
} finally {
    Pop-Location
}
