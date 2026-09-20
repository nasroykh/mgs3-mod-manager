#Requires -Version 5.1
param(
    [Parameter(Mandatory = $true)][ValidatePattern('^v[0-9]+\.[0-9]+\.[0-9]+$')][string]$Version,
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
New-Item -ItemType Directory -Path $OutputDirectory | Out-Null
$archiveName = 'mgs3mod_' + $Version.Substring(1) + '_windows_amd64.zip'
$zip = [IO.Compression.ZipFile]::Open((Join-Path $OutputDirectory $archiveName), [IO.Compression.ZipArchiveMode]::Create)
try {
    $members = [ordered]@{
        'mgs3mod.exe' = $manager
        'README.md' = (Join-Path $projectRoot 'README.md')
        'THIRD-PARTY-NOTICES.txt' = (Join-Path $projectRoot 'THIRD-PARTY-NOTICES.txt')
    }
    foreach ($name in $members.Keys) {
        $entry = $zip.CreateEntry($name, [IO.Compression.CompressionLevel]::Optimal)
        $entry.LastWriteTime = [DateTimeOffset]::new(1980,1,1,0,0,0,[TimeSpan]::Zero)
        $source = [IO.File]::OpenRead($members[$name])
        $target = $entry.Open()
        try { $source.CopyTo($target) } finally { $source.Dispose(); $target.Dispose() }
    }
} finally { $zip.Dispose() }
& (Join-Path $PSScriptRoot 'package-qcamo.ps1') -UpstreamArchive $UpstreamArchive -LoaderArchive $LoaderArchive -ManagerExe $manager -OutputDirectory (Join-Path $OutputDirectory 'qcamo')
foreach ($name in @('qcamo-1.0.4.mgs3mod.zip','asi-loader-9.7.4.mgs3mod.zip','qcamo-1.0.4-standalone-distribution.zip')) {
    Copy-Item -LiteralPath (Join-Path $OutputDirectory ('qcamo/' + $name)) -Destination (Join-Path $OutputDirectory $name)
}
Copy-Item -LiteralPath (Join-Path $projectRoot 'install.ps1') -Destination (Join-Path $OutputDirectory 'install.ps1')
$names = @($archiveName,'asi-loader-9.7.4.mgs3mod.zip','qcamo-1.0.4.mgs3mod.zip','qcamo-1.0.4-standalone-distribution.zip','install.ps1')
$lines = foreach ($name in $names) {
    (Get-FileHash -LiteralPath (Join-Path $OutputDirectory $name) -Algorithm SHA256).Hash.ToLowerInvariant() + '  ' + $name
}
[IO.File]::WriteAllText((Join-Path $OutputDirectory 'checksums.txt'), ($lines -join "`n") + "`n", [Text.UTF8Encoding]::new($false))
Write-Output ($lines -join "`n")
