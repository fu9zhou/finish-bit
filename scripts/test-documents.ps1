param([string]$RuntimeHome = (Join-Path $PSScriptRoot '../.finishbit-test-batch-b'))
$ErrorActionPreference = 'Stop'
$taskRepo = Split-Path $PSScriptRoot -Parent
$taskNames = @('GOTOOLCHAIN','FINISHBIT_HOME','FINISHBIT_TEST_PANDOC')
$taskPrevious = @{}
foreach ($name in $taskNames) { $taskPrevious[$name] = [Environment]::GetEnvironmentVariable($name,'Process') }
function Invoke-DocumentCheck {
    param([string]$Program,[string[]]$Arguments)
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed: $LASTEXITCODE" }
}
Push-Location $taskRepo
try {
    $env:GOTOOLCHAIN='go1.26.8'
    $env:FINISHBIT_HOME=[IO.Path]::GetFullPath($RuntimeHome)
    $taskBinary=Join-Path $taskRepo 'fnsh.exe'
    Invoke-DocumentCheck go @('build','-o',$taskBinary,'./cmd/fnsh')
    Invoke-DocumentCheck $taskBinary @('pkg','add','pandoc')
    $taskPackage=Invoke-DocumentCheck $taskBinary @('pkg','info','pandoc','--json') | ConvertFrom-Json
    $env:FINISHBIT_TEST_PANDOC=Join-Path $env:FINISHBIT_HOME "packages/pandoc/$($taskPackage.installed.version)/$($taskPackage.installed.executables.pandoc)"
    Invoke-DocumentCheck $env:FINISHBIT_TEST_PANDOC @('--version')
    Invoke-DocumentCheck go @('test','-count=1','./...')
    Invoke-DocumentCheck go @('vet','./...')
    Invoke-DocumentCheck go @('build','./cmd/fnsh')
    Invoke-DocumentCheck go @('run','golang.org/x/vuln/cmd/govulncheck@v1.7.0','./...')
    Invoke-DocumentCheck node @('--check','website/app.js')
    Invoke-DocumentCheck node @('scripts/website-catalog.mjs','--check')
    $taskEvidence=Join-Path $env:FINISHBIT_HOME ('evidence/'+[Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $taskEvidence -Force | Out-Null
    $taskInput=Join-Path $taskEvidence 'report.md'
    [IO.File]::WriteAllText($taskInput,"# Document acceptance`n`nVerified content 42.`n")
    $taskOutput=Join-Path $taskEvidence 'report.docx'
    $taskRequest=Join-Path $taskEvidence 'request.json'
    @{operation='document.convert';inputs=@($taskInput);options=@{to='docx';output=$taskOutput}} | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $taskRequest -Encoding utf8
    Invoke-DocumentCheck $taskBinary @('run','--request',$taskRequest,'--json')
    $taskText=Join-Path $taskEvidence 'report.txt'
    Invoke-DocumentCheck $taskBinary @('document','text',$taskOutput,'--output',$taskText,'--json')
    if (-not ([IO.File]::ReadAllText($taskText).Contains('Verified content 42.'))) { throw 'CLI round-trip lost document content' }
    @{completedAt=[DateTime]::UtcNow.ToString('o');group='pandoc';operations=10;package=$taskPackage;checks=@('all operation real-file assertions','25 writer formats','go test/vet/build','govulncheck','website contract drift','structured and direct CLI round-trip')} | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $taskEvidence 'acceptance.json') -Encoding utf8
    Write-Output "Document group accepted. Evidence: $taskEvidence"
}
finally {
    foreach ($name in $taskNames) { [Environment]::SetEnvironmentVariable($name,$taskPrevious[$name],'Process') }
    Pop-Location
}
