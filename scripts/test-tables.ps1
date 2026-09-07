param([string]$RuntimeHome = (Join-Path $PSScriptRoot '../.finishbit-test-batch-b'))
$ErrorActionPreference = 'Stop'
$taskRepo = Split-Path $PSScriptRoot -Parent
$taskNames = @('GOTOOLCHAIN','FINISHBIT_HOME','FINISHBIT_TEST_QSV')
$taskPrevious = @{}
foreach ($name in $taskNames) { $taskPrevious[$name] = [Environment]::GetEnvironmentVariable($name,'Process') }
function Invoke-TableCheck {
    param([string]$Program,[string[]]$Arguments)
    & $Program @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Program failed: $LASTEXITCODE" }
}
Push-Location $taskRepo
try {
    $env:GOTOOLCHAIN='go1.26.8'
    $env:FINISHBIT_HOME=[IO.Path]::GetFullPath($RuntimeHome)
    $taskBinary=Join-Path $taskRepo 'fnsh.exe'
    Invoke-TableCheck go @('build','-o',$taskBinary,'./cmd/fnsh')
    Invoke-TableCheck $taskBinary @('pkg','add','qsv')
    $taskPackage=Invoke-TableCheck $taskBinary @('pkg','info','qsv','--json') | ConvertFrom-Json
    $env:FINISHBIT_TEST_QSV=Join-Path $env:FINISHBIT_HOME "packages/qsv/$($taskPackage.installed.version)/$($taskPackage.installed.executables.qsv)"
    Invoke-TableCheck $env:FINISHBIT_TEST_QSV @('--version')
    Invoke-TableCheck go @('test','-count=1','./...')
    Invoke-TableCheck go @('vet','./...')
    Invoke-TableCheck go @('build','./cmd/fnsh')
    Invoke-TableCheck go @('run','golang.org/x/vuln/cmd/govulncheck@v1.7.0','./...')
    Invoke-TableCheck node @('scripts/website-catalog.mjs','--check')
    $taskEvidence=Join-Path $env:FINISHBIT_HOME ('evidence/'+[Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $taskEvidence -Force | Out-Null
    $taskInput=Join-Path $taskEvidence 'orders.csv'
    [IO.File]::WriteAllText($taskInput,"id,amount`n001,10`n002,20`n003,30`n")
    $taskOutput=Join-Path $taskEvidence 'selected.csv'
    $taskRequest=Join-Path $taskEvidence 'request.json'
    @{operation='csv.select';inputs=@($taskInput);options=@{columns=@('id');output=$taskOutput}} | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $taskRequest -Encoding utf8
    Invoke-TableCheck $taskBinary @('run','--request',$taskRequest,'--json')
    $taskRows=@(Import-Csv -LiteralPath $taskOutput)
    if ($taskRows.Count -ne 3 -or $taskRows[0].id -ne '001') { throw 'Structured CLI did not preserve rows/leading zeros' }
    $taskStats=Join-Path $taskEvidence 'stats.csv'
    Invoke-TableCheck $taskBinary @('csv','stats',$taskInput,'--columns','amount','--output',$taskStats,'--json')
    if ((Import-Csv -LiteralPath $taskStats).mean -ne '20') { throw 'CLI statistics mean is incorrect' }
    @{completedAt=[DateTime]::UtcNow.ToString('o');group='qsv';operations=27;package=$taskPackage;checks=@('all operation real-file assertions','go test/vet/build','govulncheck','website contract drift','structured and direct CLI data checks')} | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $taskEvidence 'acceptance.json') -Encoding utf8
    Write-Output "Table group accepted. Evidence: $taskEvidence"
}
finally {
    foreach ($name in $taskNames) { [Environment]::SetEnvironmentVariable($name,$taskPrevious[$name],'Process') }
    Pop-Location
}
