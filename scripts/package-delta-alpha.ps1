#Requires -Version 5.1
# Build the Delta controls tester kit for a prerelease from the release's own
# manager zip, window app zip and loader package plus the locally played plugin
# packages. The plugin payloads are pinned to the played builds; anything else
# is refused.
param(
    [Parameter(Mandatory = $true)][ValidatePattern('^v[0-9]+\.[0-9]+\.[0-9]+-(alpha|beta|rc)\.[0-9]+$')][string]$Version,
    [Parameter(Mandatory = $true)][string]$ManagerZip,
    [Parameter(Mandatory = $true)][string]$LoaderPackage,
    [Parameter(Mandatory = $true)][string]$GuiZip,
    # The release draft's checksums.txt; the three release inputs must match its entries.
    [Parameter(Mandatory = $true)][string]$Checksums,
    [string]$OutputDirectory = ''
)
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem
$projectRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
# .NET resolves relative paths against the process directory, not the
# PowerShell location, so resolve every input here.
$ManagerZip = (Resolve-Path -LiteralPath $ManagerZip).ProviderPath
$LoaderPackage = (Resolve-Path -LiteralPath $LoaderPackage).ProviderPath
$GuiZip = (Resolve-Path -LiteralPath $GuiZip).ProviderPath
$Checksums = (Resolve-Path -LiteralPath $Checksums).ProviderPath
if (!$OutputDirectory) { $OutputDirectory = Join-Path $projectRoot "dist/delta-$Version" }
$OutputDirectory = $PSCmdlet.GetUnresolvedProviderPathFromPSPath($OutputDirectory)
$sumLines = @(Get-Content -LiteralPath $Checksums)
$expectedInputs = [ordered]@{
    ('mgs3mod_' + $Version.Substring(1) + '_windows_amd64.zip')     = $ManagerZip
    ('mgs3mod-gui_' + $Version.Substring(1) + '_windows_amd64.zip') = $GuiZip
    'asi-loader-9.7.4.mgs3mod.zip'                                  = $LoaderPackage
}
foreach ($name in $expectedInputs.Keys) {
    if ([IO.Path]::GetFileName($expectedInputs[$name]) -cne $name) { throw "Input must be the release's $name, got $($expectedInputs[$name])" }
    $entry = @($sumLines | Where-Object { $_ -cmatch ('^[a-f0-9]{64}  ' + [regex]::Escape($name) + '$') })
    $actual = (Get-FileHash -LiteralPath $expectedInputs[$name] -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($entry.Count -ne 1 -or $entry[0].Substring(0, 64) -cne $actual) { throw "$name does not match $Checksums" }
}
if (Test-Path -LiteralPath $OutputDirectory) { throw 'Output exists; choose a fresh directory.' }

# Package file -> [played payload SHA-256, source package]
$plugins = [ordered]@{
    'fpv-move-0.8.4.mgs3mod.zip'          = @('5438ace896a2cad9657981051664094cda6386bf747a49076e5dbbfbe31a09b7', 'work/fpv-move-0.8.4.zip')
    'qcamo-face-1.0.4-face.4.mgs3mod.zip' = @('6c40bdaf6ddff2138294c02a1afa1b5e833afc23c4c929db0f99abe9e5d4a44a', 'work/qcamo-face-1.0.4-face.4.zip')
}
$managerMembers = @('mgs3mod.exe', 'README.md', 'THIRD-PARTY-NOTICES.txt')
$guiMembers = @('mgs3mod-gui.exe', 'README.md', 'THIRD-PARTY-NOTICES-GUI.txt')

function Get-Sha256([byte[]]$Bytes) {
    $sha = [Security.Cryptography.SHA256]::Create()
    try { return ([BitConverter]::ToString($sha.ComputeHash($Bytes)) -replace '-', '').ToLowerInvariant() } finally { $sha.Dispose() }
}

function Read-Member([string]$Zip, [string]$Name) {
    $archive = [IO.Compression.ZipFile]::OpenRead($Zip)
    try {
        $entries = @($archive.Entries | Where-Object { $_.FullName -ceq $Name })
        if ($entries.Count -ne 1) { throw "$Zip must contain exactly one $Name" }
        $stream = $entries[0].Open(); $memory = [IO.MemoryStream]::new()
        try { $stream.CopyTo($memory) } finally { $stream.Dispose() }
        return $memory.ToArray()
    } finally { $archive.Dispose() }
}

function Get-Names([string]$Zip) {
    $archive = [IO.Compression.ZipFile]::OpenRead($Zip)
    try { return @($archive.Entries | ForEach-Object FullName) } finally { $archive.Dispose() }
}

$members = [ordered]@{}
$managerNames = Get-Names $ManagerZip
if (Compare-Object $managerMembers $managerNames) { throw 'Unexpected manager zip members' }
foreach ($name in $managerMembers) { $members[$name] = Read-Member $ManagerZip $name }
$guiNames = Get-Names $GuiZip
if (Compare-Object $guiMembers $guiNames) { throw 'Unexpected window app zip members' }
# The kit carries the window app at its root, next to the packages it installs.
foreach ($name in @('mgs3mod-gui.exe', 'THIRD-PARTY-NOTICES-GUI.txt')) { $members[$name] = Read-Member $GuiZip $name }
$members['asi-loader-9.7.4.mgs3mod.zip'] = [IO.File]::ReadAllBytes([IO.Path]::GetFullPath($LoaderPackage))
foreach ($name in $plugins.Keys) {
    $source = Join-Path $projectRoot $plugins[$name][1]
    $payloads = @(Get-Names $source | Where-Object { $_ -like 'payload/*' })
    if ($payloads.Count -ne 1) { throw "$source must hold one payload" }
    $payloadHash = Get-Sha256 (Read-Member $source $payloads[0])
    if ($payloadHash -cne $plugins[$name][0]) { throw "$source payload is $payloadHash, not the played build" }
    $members[$name] = [IO.File]::ReadAllBytes($source)
}
$members['DELTA-CONTROLS-ALPHA.md'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'docs/delta-controls-alpha.md'))
$members['delta-controls-untested.md'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'docs/delta-controls-untested.md'))
$members['QCAMO-FACE-LICENSE.txt'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'work/qcamo-face-notices/LICENSE.txt'))
$members['QCAMO-FACE-UPSTREAM.md'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'work/qcamo-face-notices/UPSTREAM.md'))
$members['LOADER-LICENSE.txt'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'mods/asi-loader/LICENSE.txt'))
$members['LOADER-UPSTREAM.md'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'mods/asi-loader/UPSTREAM.md'))
$sums = foreach ($name in $members.Keys) { (Get-Sha256 $members[$name]) + '  ' + $name }
$members['SHA256SUMS.txt'] = [Text.UTF8Encoding]::new($false).GetBytes(($sums -join "`n") + "`n")

New-Item -ItemType Directory -Path $OutputDirectory | Out-Null
$kitName = "mgs3-delta-controls-$($Version.Substring(1))-kit.zip"
$zip = [IO.Compression.ZipFile]::Open((Join-Path $OutputDirectory $kitName), [IO.Compression.ZipArchiveMode]::Create)
try {
    foreach ($name in $members.Keys) {
        $entry = $zip.CreateEntry($name, [IO.Compression.CompressionLevel]::Optimal)
        $entry.LastWriteTime = [DateTimeOffset]::new(1980, 1, 1, 0, 0, 0, [TimeSpan]::Zero)
        $target = $entry.Open()
        try { $target.Write($members[$name], 0, $members[$name].Length) } finally { $target.Dispose() }
    }
} finally { $zip.Dispose() }
$line = (Get-FileHash -LiteralPath (Join-Path $OutputDirectory $kitName) -Algorithm SHA256).Hash.ToLowerInvariant() + '  ' + $kitName
[IO.File]::WriteAllText((Join-Path $OutputDirectory 'delta-controls-checksums.txt'), $line + "`n", [Text.UTF8Encoding]::new($false))
Write-Output ($sums -join "`n")
Write-Output $line
