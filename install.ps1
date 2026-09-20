#Requires -Version 5.1
[CmdletBinding()]
param(
    [string]$Version,
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA 'Programs\mgs3mod'),
    [switch]$NoPath
)

$ErrorActionPreference = 'Stop'
$ChecksumName = 'checksums.txt'

function Fail([string]$Message) { throw "Installer: $Message" }

function Get-Release([string]$RequestedVersion) {
    if ($RequestedVersion) {
        if ($RequestedVersion -notmatch '^v\d+\.\d+\.\d+$') { Fail 'Version must look like v0.1.0.' }
        return Invoke-RestMethod -Uri ("https://api.github.com/repos/nasroykh/mgs3-mod-manager/releases/tags/{0}" -f $RequestedVersion) -Headers @{ 'User-Agent' = 'mgs3mod-installer' }
    }
    return Invoke-RestMethod -Uri 'https://api.github.com/repos/nasroykh/mgs3-mod-manager/releases/latest' -Headers @{ 'User-Agent' = 'mgs3mod-installer' }
}

function Get-Asset($Release, [string]$Name) {
    $matches = @($Release.assets | Where-Object { $_.name -ceq $Name })
    if ($matches.Count -ne 1) { Fail "Release must contain exactly one asset named $Name." }
    return $matches[0]
}

function Download([string]$Uri, [string]$Path) {
    Invoke-WebRequest -Uri $Uri -OutFile $Path -UseBasicParsing -Headers @{ 'User-Agent' = 'mgs3mod-installer' }
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { Fail "Download did not produce $Path." }
}

function Assert-SafeArchive([string]$ZipPath, [string]$ExtractDir) {
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip = [IO.Compression.ZipFile]::OpenRead($ZipPath)
    try {
        $entries = @($zip.Entries)
        if ($entries.Count -ne 3) { Fail 'Archive must contain exactly three root files.' }
        $names = @($entries | ForEach-Object { $_.FullName })
        foreach ($entry in $entries) {
            if ([string]::IsNullOrWhiteSpace($entry.Name) -or $entry.FullName -match '[\\/]' -or $entry.FullName -match '(^|/|\\)\.\.?($|/|\\)') { Fail "Unsafe archive member: $($entry.FullName)." }
            if ($entry.FullName -notin @('mgs3mod.exe','README.md','THIRD-PARTY-NOTICES.txt')) { Fail "Unexpected archive member: $($entry.FullName)." }
            $unixMode = ($entry.ExternalAttributes -shr 16) -band 0xF000
            if ($unixMode -eq 0xA000) { Fail "Symlink archive member: $($entry.FullName)." }
        }
        foreach ($required in @('mgs3mod.exe','README.md','THIRD-PARTY-NOTICES.txt')) {
            if (@($names | Where-Object { $_ -ceq $required }).Count -ne 1) { Fail "Archive must contain exactly one $required." }
        }
        [IO.Compression.ZipFileExtensions]::ExtractToDirectory($zip, $ExtractDir)
    } finally { $zip.Dispose() }
}

function Normalize-PathEntry([string]$Path) {
    try { return ([IO.Path]::GetFullPath([Environment]::ExpandEnvironmentVariables($Path))).TrimEnd('\','/').ToLowerInvariant() } catch { return $Path.Trim().TrimEnd('\','/').ToLowerInvariant() }
}
function Get-UserPath { [Environment]::GetEnvironmentVariable('Path','User') }
function Set-UserPath([string]$Value) { [Environment]::SetEnvironmentVariable('Path', $Value, 'User') }
function Add-InstallPath([string]$Directory) {
    $entry = Normalize-PathEntry $Directory
    $parts = @((Get-UserPath) -split ';' | Where-Object { $_ -and (Normalize-PathEntry $_) -ne $entry })
    Set-UserPath (($parts + $Directory) -join ';')
}
function Assert-NoReparse([string]$Path) {
    $cursor = [IO.Path]::GetFullPath($Path)
    while ($cursor) {
        if (Test-Path -LiteralPath $cursor) {
            if (((Get-Item -LiteralPath $cursor -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { Fail "Reparse point is not supported: $cursor" }
        }
        $parent = [IO.Path]::GetDirectoryName($cursor)
        if ($parent -eq $cursor) { break }
        $cursor = $parent
    }
}

$architecture = $env:PROCESSOR_ARCHITECTURE
if ($env:PROCESSOR_ARCHITEW6432) { $architecture = $env:PROCESSOR_ARCHITEW6432 }
if (-not [Environment]::Is64BitOperatingSystem -or $architecture -ne 'AMD64') { Fail 'Windows AMD64 is required.' }
if (-not $InstallDir) { Fail 'InstallDir cannot be empty.' }
$installRoot = [IO.Path]::GetFullPath($InstallDir)
$tempRoot = Join-Path ([IO.Path]::GetTempPath()) ('mgs3mod-installer-' + [guid]::NewGuid().ToString('N'))
$backups = @{}
$changed = @()
$completed = $false
$pathAttempted = $false
$previousUserPath = $null
$exePath = Join-Path $installRoot 'mgs3mod.exe'
try {
    New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
    $downloadDir = Join-Path $tempRoot 'downloads'; $extractDir = Join-Path $tempRoot 'extract'
    New-Item -ItemType Directory -Path $downloadDir,$extractDir -Force | Out-Null
    $release = Get-Release $Version
    $releaseTag = [string]$release.tag_name
    if ($releaseTag -notmatch '^v\d+\.\d+\.\d+$') { Fail 'Release tag must be a semantic version with v prefix.' }
    if ($Version -and $releaseTag -cne $Version) { Fail 'Requested version does not match release tag.' }
    $AssetName = 'mgs3mod_{0}_windows_amd64.zip' -f $releaseTag.Substring(1)
    $zipAsset = Get-Asset $release $AssetName
    $sumAsset = Get-Asset $release $ChecksumName
    $zipPath = Join-Path $downloadDir $AssetName; $sumPath = Join-Path $downloadDir $ChecksumName
    Download $zipAsset.browser_download_url $zipPath; Download $sumAsset.browser_download_url $sumPath
    $sumLines = @(Get-Content -LiteralPath $sumPath | Where-Object { $_.Trim() })
    $sumMatches = @($sumLines | Where-Object { $_ -match ('^([0-9A-Fa-f]{64})\s+\*?' + [regex]::Escape($AssetName) + '\s*$') })
    if ($sumMatches.Count -ne 1) { Fail "checksums.txt must contain exactly one checksum for $AssetName." }
    $expectedHash = ([regex]::Match($sumMatches[0], '^([0-9A-Fa-f]{64})')).Groups[1].Value.ToUpperInvariant()
    $actualHash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToUpperInvariant()
    if ($actualHash -cne $expectedHash) { Fail 'SHA256 checksum mismatch.' }
    Assert-SafeArchive $zipPath $extractDir
    if (-not (Test-Path (Join-Path $extractDir 'mgs3mod.exe') -PathType Leaf)) { Fail 'Archive missing expected executable.' }
    Assert-NoReparse $installRoot
    if ((Test-Path -LiteralPath $installRoot) -and !(Test-Path -LiteralPath $installRoot -PathType Container)) { Fail 'InstallDir must be a directory.' }
    New-Item -ItemType Directory -Path $installRoot -Force | Out-Null
    $names = @('mgs3mod.exe','README.md','THIRD-PARTY-NOTICES.txt')
    foreach ($name in $names) {
        $target = Join-Path $installRoot $name
        Assert-NoReparse $target
        if (Test-Path -LiteralPath $target) {
            if (!(Test-Path -LiteralPath $target -PathType Leaf)) { Fail "Destination is not a file: $target" }
            $backups[$name] = Join-Path $tempRoot ('previous-' + $name)
            Copy-Item -LiteralPath $target -Destination $backups[$name]
        }
    }
    foreach ($name in $names) {
        $target = Join-Path $installRoot $name
        $source = Join-Path $extractDir $name
        if ((Test-Path -LiteralPath $target) -and (Get-FileHash -LiteralPath $target).Hash -ceq (Get-FileHash -LiteralPath $source).Hash) { continue }
        $changed += $name
        Copy-Item -LiteralPath $source -Destination $target -Force
    }
    if (-not $NoPath) {
        $previousUserPath = Get-UserPath
        $pathAttempted = $true
        Add-InstallPath $installRoot
    }
    $completed = $true
    Write-Output "Installed mgs3mod $($release.tag_name) to $installRoot"
    if (-not $NoPath) { Write-Output 'Open a new terminal to use mgs3mod from PATH.' }
} catch {
    if ($pathAttempted) {
        try { Set-UserPath $previousUserPath } catch { Write-Warning 'Could not restore the prior user PATH; inspect it in Windows settings.' }
    }
    foreach ($name in $changed) {
        $target = Join-Path $installRoot $name
        try {
            Assert-NoReparse $target
            if ($backups.ContainsKey($name)) { Copy-Item -LiteralPath $backups[$name] -Destination $target -Force }
            elseif (Test-Path -LiteralPath $target -PathType Leaf) { Remove-Item -LiteralPath $target -Force }
        } catch { Write-Warning "Could not restore $target. Backup evidence retained at $tempRoot" }
    }
    Write-Warning "Installer failed; diagnostic files and any backups retained at $tempRoot"
    throw
} finally {
    $resolvedTemp = $null; try { $resolvedTemp = [IO.Path]::GetFullPath($tempRoot) } catch {}
    $tempParent = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\','/')
    if ($completed -and $resolvedTemp -and [IO.Path]::GetDirectoryName($resolvedTemp) -eq $tempParent -and ([IO.Path]::GetFileName($resolvedTemp) -match '^mgs3mod-installer-[0-9a-f]{32}$')) {
        Assert-NoReparse $resolvedTemp
        Remove-Item -LiteralPath $resolvedTemp -Recurse -Force -ErrorAction SilentlyContinue
    }
}
