$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$installer = Join-Path $root 'install.ps1'
if (-not [Environment]::Is64BitOperatingSystem) { Write-Warning 'Installer tests require Windows AMD64.'; exit 0 }
$work = Join-Path ([IO.Path]::GetTempPath()) ('mgs3mod-installer-test-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $work -Force | Out-Null
try {
    $fixture = Join-Path $work 'fixture'; New-Item -ItemType Directory -Path $fixture -Force | Out-Null
    Set-Content (Join-Path $fixture 'mgs3mod.exe') 'fixture-exe'
    Set-Content (Join-Path $fixture 'README.md') 'readme'
    Set-Content (Join-Path $fixture 'THIRD-PARTY-NOTICES.txt') 'notices'
    $zip = Join-Path $work 'mgs3mod_0.1.0_windows_amd64.zip'; Compress-Archive (Join-Path $fixture '*') $zip
    $hash = (Get-FileHash $zip -Algorithm SHA256).Hash.ToLowerInvariant()
    Set-Content (Join-Path $work 'checksums.txt') "$hash  mgs3mod_0.1.0_windows_amd64.zip"
    $mock = @"
function Invoke-RestMethod { [pscustomobject]@{ tag_name='v0.1.0'; assets=@([pscustomobject]@{name='mgs3mod_0.1.0_windows_amd64.zip';browser_download_url='$zip'},[pscustomobject]@{name='checksums.txt';browser_download_url='$(Join-Path $work 'checksums.txt')'}) } }
function Invoke-WebRequest([string]`$Uri,[string]`$OutFile) { Copy-Item -LiteralPath `$Uri -Destination `$OutFile -Force }
. '$installer' -Version v0.1.0 -InstallDir '$(Join-Path $work 'install')' -NoPath
"@
    $testScript = Join-Path $work 'run.ps1'; Set-Content $testScript $mock
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $testScript
    if (-not (Test-Path (Join-Path $work 'install\mgs3mod.exe'))) { throw 'success install missing executable' }
    Write-Output 'PASS: success custom dir NoPath'
    $badSum = Join-Path $work 'bad-checksums.txt'; Set-Content $badSum ('0' * 64 + '  mgs3mod_0.1.0_windows_amd64.zip')
    $badRun = $mock.Replace((Join-Path $work 'checksums.txt'), $badSum).Replace((Join-Path $work 'install'), (Join-Path $work 'bad-install'))
    $badScript = Join-Path $work 'bad.ps1'; Set-Content $badScript $badRun
    $savedPreference = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $badScript 2>&1 | Out-Null
    $ErrorActionPreference = $savedPreference
    if ($LASTEXITCODE -eq 0) { throw 'bad checksum unexpectedly succeeded' }
    Write-Output 'PASS: bad checksum rejected'

    function Invoke-FailureCase([string]$Name, [string]$ChecksumText, [string]$ZipPath, [string]$InstallPath) {
        $sumCase = Join-Path $work "$Name-checksums.txt"; Set-Content $sumCase $ChecksumText
        $run = $mock.Replace((Join-Path $work 'checksums.txt'), $sumCase).Replace((Join-Path $work 'install'), $InstallPath).Replace((Join-Path $work 'mgs3mod_0.1.0_windows_amd64.zip'), $ZipPath)
        $caseScript = Join-Path $work "$Name.ps1"; Set-Content $caseScript $run
        $saved = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
        & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $caseScript 2>&1 | Out-Null
        $exit = $LASTEXITCODE; $ErrorActionPreference = $saved
        if ($exit -eq 0) { throw "$Name unexpectedly succeeded" }
    }
    Invoke-FailureCase 'missing-checksum' '' $zip (Join-Path $work 'missing-install')
    Write-Output 'PASS: missing checksum rejected'
    Invoke-FailureCase 'duplicate-checksum' "$hash  mgs3mod_0.1.0_windows_amd64.zip`n$hash  mgs3mod_0.1.0_windows_amd64.zip" $zip (Join-Path $work 'duplicate-install')
    Write-Output 'PASS: duplicate checksum rejected'
    $malformed = Join-Path $work 'malformed.zip'; Set-Content $malformed 'not-a-zip'
    $malformedHash = (Get-FileHash $malformed -Algorithm SHA256).Hash.ToLowerInvariant()
    Invoke-FailureCase 'malformed-archive' "$malformedHash  mgs3mod_0.1.0_windows_amd64.zip" $malformed (Join-Path $work 'malformed-install')
    Write-Output 'PASS: malformed archive rejected'
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $traversal = Join-Path $work 'traversal.zip'; $archive = [IO.Compression.ZipFile]::Open($traversal, [IO.Compression.ZipArchiveMode]::Create)
    foreach ($name in @('mgs3mod.exe','README.md','THIRD-PARTY-NOTICES.txt','..\escape.txt')) { $entry = $archive.CreateEntry($name); $stream = [IO.StreamWriter]::new($entry.Open()); $stream.Write('x'); $stream.Dispose() }
    $archive.Dispose(); $traversalHash = (Get-FileHash $traversal -Algorithm SHA256).Hash.ToLowerInvariant()
    Invoke-FailureCase 'path-traversal' "$traversalHash  mgs3mod_0.1.0_windows_amd64.zip" $traversal (Join-Path $work 'traversal-install')
    Write-Output 'PASS: path traversal archive rejected'
    $preserve = Join-Path $work 'preserve-install'; New-Item $preserve -ItemType Directory | Out-Null; Set-Content (Join-Path $preserve 'mgs3mod.exe') 'old-exe'
    Invoke-FailureCase 'rollback' (('0' * 64) + '  mgs3mod_0.1.0_windows_amd64.zip') $zip $preserve
    if ((Get-Content (Join-Path $preserve 'mgs3mod.exe') -Raw) -ne "old-exe`r`n") { throw 'existing executable was not preserved' }
    Write-Output 'PASS: existing executable preserved on failure'
    $source = Get-Content $installer -Raw
    $tokens = $null; $parseErrors = $null
    $ast = [Management.Automation.Language.Parser]::ParseInput($source, [ref]$tokens, [ref]$parseErrors)
    if ($parseErrors.Count) { throw 'Installer parse failed' }
    foreach ($function in $ast.FindAll({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] }, $false)) {
        . ([scriptblock]::Create($function.Extent.Text))
    }
    $script:mockPath = '%SystemRoot%\system32;' + (Join-Path $work 'install').ToUpperInvariant() + '\;' + (Join-Path $work 'install-other')
    function Get-UserPath { $script:mockPath }
    function Set-UserPath([string]$Value) { $script:mockPath = $Value }
    Add-InstallPath (Join-Path $work 'install')
    $expectedPath = '%SystemRoot%\system32;' + (Join-Path $work 'install-other') + ';' + (Join-Path $work 'install')
    if ($script:mockPath -cne $expectedPath) { throw 'PATH normalization damaged other entries or retained duplicate' }
    Add-InstallPath (Join-Path $work 'install')
    if ($script:mockPath -cne $expectedPath) { throw 'PATH is not idempotent' }
    Write-Output 'PASS: behavioral PATH deduplication, preservation, and idempotence'
    $latestScript = Join-Path $work 'latest.ps1'
    Set-Content $latestScript ($mock.Replace('-Version v0.1.0 ', ''))
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $latestScript | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'latest lookup failed' }
    Write-Output 'PASS: latest release lookup'
    $mismatchScript = Join-Path $work 'mismatch.ps1'
    Set-Content $mismatchScript ($mock.Replace('-Version v0.1.0 ', '-Version v0.2.0 '))
    $saved = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $mismatchScript 2>&1 | Out-Null
    $code = $LASTEXITCODE; $ErrorActionPreference = $saved
    if ($code -eq 0) { throw 'mismatched tag accepted' }
    Write-Output 'PASS: requested tag mismatch rejected'
    $failingSource = $source.Replace("function Set-UserPath([string]`$Value) { [Environment]::SetEnvironmentVariable('Path', `$Value, 'User') }", "function Set-UserPath([string]`$Value) { throw 'injected PATH failure' }")
    if ($failingSource -ceq $source) { throw 'Failure injection did not apply' }
    $failingInstaller = Join-Path $work 'failing-installer.ps1'; Set-Content $failingInstaller $failingSource
    foreach ($mode in @('new','upgrade')) {
        $destination = Join-Path $work ('postcopy-' + $mode)
        if ($mode -eq 'upgrade') {
            New-Item $destination -ItemType Directory | Out-Null
            foreach ($name in @('mgs3mod.exe','README.md','THIRD-PARTY-NOTICES.txt')) { Set-Content (Join-Path $destination $name) ('old-' + $name) }
        }
        $run = $mock.Replace($installer, $failingInstaller).Replace((Join-Path $work 'install'), $destination).Replace(' -NoPath','')
        $case = Join-Path $work ('postcopy-' + $mode + '.ps1'); Set-Content $case $run
        $saved = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
        & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $case 2>&1 | Out-Null
        $code = $LASTEXITCODE; $ErrorActionPreference = $saved
        if ($code -eq 0) { throw 'Injected failure unexpectedly succeeded' }
        foreach ($name in @('mgs3mod.exe','README.md','THIRD-PARTY-NOTICES.txt')) {
            $target = Join-Path $destination $name
            if ($mode -eq 'new' -and (Test-Path $target)) { throw 'New file survived rollback' }
            if ($mode -eq 'upgrade' -and (Get-Content $target -Raw).Trim() -cne ('old-' + $name)) { throw 'Prior file not restored' }
        }
        Write-Output "PASS: post-copy failure rollback ($mode)"
    }
} finally {
    $resolved = [IO.Path]::GetFullPath($work)
    $parent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\','/')
    if ([IO.Path]::GetDirectoryName($resolved) -eq $parent -and [IO.Path]::GetFileName($resolved) -match '^mgs3mod-installer-test-[0-9a-f]{32}$') {
        Remove-Item -LiteralPath $resolved -Recurse -Force -ErrorAction SilentlyContinue
    }
}
