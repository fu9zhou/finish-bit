param([string]$RuntimeHome = (Join-Path $PSScriptRoot '../.finishbit-test-batch-b'))
$ErrorActionPreference = 'Stop'
$taskRepo = Split-Path $PSScriptRoot -Parent
$taskNames = @('GOTOOLCHAIN','FINISHBIT_HOME','FINISHBIT_TEST_7ZIP')
$taskPrevious = @{}
foreach ($name in $taskNames) { $taskPrevious[$name] = [Environment]::GetEnvironmentVariable($name,'Process') }
function Invoke-ArchiveCheck {
    param([string]$Program,[string[]]$Arguments)
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed: $LASTEXITCODE" }
}
Push-Location $taskRepo
try {
    $env:GOTOOLCHAIN='go1.26.8'
    $env:FINISHBIT_HOME=[IO.Path]::GetFullPath($RuntimeHome)
    $taskBinary=Join-Path $taskRepo 'fnsh.exe'
    Invoke-ArchiveCheck go @('build','-o',$taskBinary,'./cmd/fnsh')
    Invoke-ArchiveCheck $taskBinary @('pkg','add','7zip')
    $taskPackage=Invoke-ArchiveCheck $taskBinary @('pkg','info','7zip','--json') | ConvertFrom-Json
    $env:FINISHBIT_TEST_7ZIP=Join-Path $env:FINISHBIT_HOME "packages/7zip/$($taskPackage.installed.version)/$($taskPackage.installed.executables.'7zip')"
    Invoke-ArchiveCheck $env:FINISHBIT_TEST_7ZIP @('i')
    Invoke-ArchiveCheck go @('test','-count=1','./...')
    Invoke-ArchiveCheck go @('vet','./...')
    Invoke-ArchiveCheck go @('build','./cmd/fnsh')
    Invoke-ArchiveCheck go @('run','golang.org/x/vuln/cmd/govulncheck@v1.7.0','./...')
    Invoke-ArchiveCheck node @('--check','website/app.js')
    Invoke-ArchiveCheck node @('scripts/website-catalog.mjs','--check')
    $taskEvidence=Join-Path $env:FINISHBIT_HOME ('evidence/'+[Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $taskEvidence -Force | Out-Null
    $taskInput=Join-Path $taskEvidence 'report.txt'
    [IO.File]::WriteAllText($taskInput,"Archive acceptance 42.`n")
    $taskOutput=Join-Path $taskEvidence 'report.zip'
    $taskRequest=Join-Path $taskEvidence 'request.json'
    @{operation='archive.create';inputs=@($taskInput);options=@{format='zip';output=$taskOutput}} | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $taskRequest -Encoding utf8
    Invoke-ArchiveCheck $taskBinary @('run','--request',$taskRequest,'--json')
    $taskExtracted=Join-Path $taskEvidence 'extracted'
    Invoke-ArchiveCheck $taskBinary @('archive','extract',$taskOutput,'--output',$taskExtracted,'--json')
    if ((Get-FileHash -LiteralPath $taskInput).Hash -ne (Get-FileHash -LiteralPath (Join-Path $taskExtracted 'report.txt')).Hash) { throw 'CLI archive round-trip changed file bytes' }
    Invoke-ArchiveCheck $taskBinary @('archive','test',$taskOutput,'--json')
    @{completedAt=[DateTime]::UtcNow.ToString('o');group='7zip';operations=10;package=$taskPackage;checks=@('all operation real-file assertions','8 write formats and byte-identical archive round-trips','go test/vet/build','govulncheck','website contract drift','structured and direct CLI round-trip')} | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $taskEvidence 'acceptance.json') -Encoding utf8
    Write-Output "Archive group accepted. Evidence: $taskEvidence"
}
finally {
    foreach ($name in $taskNames) { [Environment]::SetEnvironmentVariable($name,$taskPrevious[$name],'Process') }
    Pop-Location
}
