#Requires -Version 5.1
param(
    [Parameter(Mandatory = $true)][ValidatePattern('^v[0-9]+\.[0-9]+\.[0-9]+(-(alpha|beta|rc)\.[0-9]+)?$')][string]$Version,
    [Parameter(Mandatory = $true)][string]$UpstreamArchive,
    [Parameter(Mandatory = $true)][string]$LoaderArchive,
    [string]$OutputDirectory = ''
)
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem
$projectRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
if (!$OutputDirectory) { $OutputDirectory = Join-Path $projectRoot "dist/release-$Version" }
$OutputDirectory = [IO.Path]::GetFullPath($OutputDirectory)
if (Test-Path -LiteralPath $OutputDirectory) { throw 'Release output exists; choose a fresh directory.' }
$manager = Join-Path $projectRoot 'dist/mgs3mod.exe'
if (!(Test-Path -LiteralPath $manager -PathType Leaf)) { throw 'Run scripts/build.ps1 first.' }
$gui = Join-Path $projectRoot 'dist/mgs3mod-gui.exe'
if (!(Test-Path -LiteralPath $gui -PathType Leaf)) { throw 'Run scripts/build.ps1 first.' }
$guiNotices = Join-Path $projectRoot 'THIRD-PARTY-NOTICES-GUI.txt'
# The window app states its release; refuse a stale label.
$versionSource = Get-Content -Raw -LiteralPath (Join-Path $projectRoot 'internal/gui/version.go')
if ($versionSource -cnotmatch ('const Version = "' + [regex]::Escape($Version.Substring(1)) + '"')) { throw "internal/gui/version.go does not say $Version" }
# Every Go module linked into the window app must be named in its notices.
$noticeText = Get-Content -Raw -LiteralPath $guiNotices
$linked = @(& go version -m $gui | ForEach-Object { if ($_ -match '^\s+dep\s+(\S+)\s+(\S+)') { "$($Matches[1]) $($Matches[2])" } })
if ($LASTEXITCODE -ne 0 -or $linked.Count -eq 0) { throw 'Cannot list the modules in mgs3mod-gui.exe' }
$unlisted = @($linked | Where-Object { !$noticeText.Contains("  $_ (") })
if ($unlisted.Count -gt 0) { throw "THIRD-PARTY-NOTICES-GUI.txt misses: $($unlisted -join ', ')" }
New-Item -ItemType Directory -Path $OutputDirectory | Out-Null

function New-ReleaseZip([string]$Path, $Members) {
    $zip = [IO.Compression.ZipFile]::Open($Path, [IO.Compression.ZipArchiveMode]::Create)
    try {
        foreach ($name in $Members.Keys) {
            $entry = $zip.CreateEntry($name, [IO.Compression.CompressionLevel]::Optimal)
            $entry.LastWriteTime = [DateTimeOffset]::new(1980,1,1,0,0,0,[TimeSpan]::Zero)
            $source = [IO.File]::OpenRead($Members[$name])
            $target = $entry.Open()
            try { $source.CopyTo($target) } finally { $source.Dispose(); $target.Dispose() }
        }
    } finally { $zip.Dispose() }
}

$archiveName = 'mgs3mod_' + $Version.Substring(1) + '_windows_amd64.zip'
New-ReleaseZip (Join-Path $OutputDirectory $archiveName) ([ordered]@{
    'mgs3mod.exe' = $manager
    'README.md' = (Join-Path $projectRoot 'README.md')
    'THIRD-PARTY-NOTICES.txt' = (Join-Path $projectRoot 'THIRD-PARTY-NOTICES.txt')
})
$guiArchiveName = 'mgs3mod-gui_' + $Version.Substring(1) + '_windows_amd64.zip'
New-ReleaseZip (Join-Path $OutputDirectory $guiArchiveName) ([ordered]@{
    'mgs3mod-gui.exe' = $gui
    'README.md' = (Join-Path $projectRoot 'README.md')
    'THIRD-PARTY-NOTICES-GUI.txt' = $guiNotices
})
& (Join-Path $PSScriptRoot 'package-qcamo.ps1') -UpstreamArchive $UpstreamArchive -LoaderArchive $LoaderArchive -ManagerExe $manager -OutputDirectory (Join-Path $OutputDirectory 'qcamo')
foreach ($name in @('qcamo-1.0.4.mgs3mod.zip','asi-loader-9.7.4.mgs3mod.zip','qcamo-1.0.4-standalone-distribution.zip')) {
    Copy-Item -LiteralPath (Join-Path $OutputDirectory ('qcamo/' + $name)) -Destination (Join-Path $OutputDirectory $name)
}
Copy-Item -LiteralPath (Join-Path $projectRoot 'install.ps1') -Destination (Join-Path $OutputDirectory 'install.ps1')
$names = @($archiveName,$guiArchiveName,'asi-loader-9.7.4.mgs3mod.zip','qcamo-1.0.4.mgs3mod.zip','qcamo-1.0.4-standalone-distribution.zip','install.ps1')
$lines = foreach ($name in $names) {
    (Get-FileHash -LiteralPath (Join-Path $OutputDirectory $name) -Algorithm SHA256).Hash.ToLowerInvariant() + '  ' + $name
}
[IO.File]::WriteAllText((Join-Path $OutputDirectory 'checksums.txt'), ($lines -join "`n") + "`n", [Text.UTF8Encoding]::new($false))
Write-Output ($lines -join "`n")
