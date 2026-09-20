param(
    [Parameter(Mandatory = $true)][string]$UpstreamArchive,
    [Parameter(Mandatory = $true)][string]$LoaderArchive,
    [string]$ManagerExe = '',
    [string]$OutputDirectory = ''
)
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem
$projectRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
if (-not $ManagerExe) { $ManagerExe = Join-Path $projectRoot 'dist/mgs3mod.exe' }
if (-not $OutputDirectory) { $OutputDirectory = Join-Path $projectRoot 'dist/qcamo-standalone' }
$ManagerExe = [IO.Path]::GetFullPath($ManagerExe)
$OutputDirectory = [IO.Path]::GetFullPath($OutputDirectory)
$source = Join-Path $projectRoot 'mods/qcamo'
$utf8 = New-Object Text.UTF8Encoding($false)

function Get-BytesHash([byte[]]$Bytes) {
    $hasher = [Security.Cryptography.SHA256]::Create()
    try { return [BitConverter]::ToString($hasher.ComputeHash($Bytes)).Replace('-', '').ToLowerInvariant() }
    finally { $hasher.Dispose() }
}
function Get-FileDigest([string]$Path) {
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}
function Add-ZipBytes($Zip, [string]$Name, [byte[]]$Bytes) {
    $entry = $Zip.CreateEntry($Name, [IO.Compression.CompressionLevel]::Optimal)
    $entry.LastWriteTime = [DateTimeOffset]::new(1980, 1, 1, 0, 0, 0, [TimeSpan]::Zero)
    $stream = $entry.Open()
    try { $stream.Write($Bytes, 0, $Bytes.Length) } finally { $stream.Dispose() }
}

if (-not (Test-Path -LiteralPath $ManagerExe -PathType Leaf)) { throw 'Build the manager before packaging.' }
if ((Get-FileDigest $UpstreamArchive) -ne '8418131cecfc364a15e1e8f4bfbfbb7be25b2f0cbf0d58063f6c4397a40b0182') { throw 'Upstream archive SHA-256 does not match official QCamo 1.0.4.' }
$inputZip = [IO.Compression.ZipFile]::OpenRead([IO.Path]::GetFullPath($UpstreamArchive))
try {
    if ($inputZip.Entries.Count -ne 1 -or $inputZip.Entries[0].FullName -cne 'qcamo.asi' -or $inputZip.Entries[0].Length -ne 4063819) { throw 'Unexpected upstream archive layout.' }
    $stream = $inputZip.Entries[0].Open()
    $buffer = New-Object IO.MemoryStream
    try { $stream.CopyTo($buffer); $payload = $buffer.ToArray() } finally { $stream.Dispose(); $buffer.Dispose() }
} finally { $inputZip.Dispose() }
if ((Get-BytesHash $payload) -ne '4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193') { throw 'QCamo payload SHA-256 mismatch.' }

if ((Get-FileDigest $LoaderArchive) -ne 'e5860e7d9a1805267535b65749575b5e406cc6ea3325c7392189c578815045d1') { throw 'Loader archive does not match official Ultimate ASI Loader 9.7.4 x64.' }
$loaderZip = [IO.Compression.ZipFile]::OpenRead([IO.Path]::GetFullPath($LoaderArchive))
try {
    if ($loaderZip.Entries.Count -ne 1 -or $loaderZip.Entries[0].FullName -cne 'dinput8.dll' -or $loaderZip.Entries[0].Length -ne 1198304) { throw 'Unexpected loader archive layout.' }
    $stream = $loaderZip.Entries[0].Open()
    $buffer = New-Object IO.MemoryStream
    try { $stream.CopyTo($buffer); $loaderPayload = $buffer.ToArray() } finally { $stream.Dispose(); $buffer.Dispose() }
} finally { $loaderZip.Dispose() }
if ((Get-BytesHash $loaderPayload) -ne '031a3e5576d91dce1e438d36b9a3d462c7334ab4791990a8ff1e3ddc0e132daf') { throw 'Loader payload SHA-256 mismatch.' }

New-Item -ItemType Directory -Path $OutputDirectory -Force | Out-Null
$innerName = 'qcamo-1.0.4.mgs3mod.zip'
$loaderName = 'asi-loader-9.7.4.mgs3mod.zip'
$outerName = 'qcamo-1.0.4-standalone-distribution.zip'
$innerPath = Join-Path $OutputDirectory $innerName
$outerPath = Join-Path $OutputDirectory $outerName
$loaderPath = Join-Path $OutputDirectory $loaderName
$checksumsPath = Join-Path $OutputDirectory 'qcamo-1.0.4-SHA256SUMS.txt'
foreach ($candidate in @($innerPath, $loaderPath, $outerPath, $checksumsPath)) {
    if (Test-Path -LiteralPath $candidate) { throw "Output exists; select a new output directory: $candidate" }
}
$stage = Join-Path $projectRoot ('work/qcamo-release-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path (Join-Path $stage 'payload') -Force | Out-Null
Copy-Item -LiteralPath (Join-Path $source 'manifest.json') -Destination (Join-Path $stage 'manifest.json')
[IO.File]::WriteAllBytes((Join-Path $stage 'payload/qcamo.asi'), $payload)
& $ManagerExe pack $stage --out $innerPath --json
if ($LASTEXITCODE -ne 0) { throw 'Manager package validation failed.' }

$loaderStage = Join-Path $projectRoot ('work/loader-release-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path (Join-Path $loaderStage 'payload') -Force | Out-Null
Copy-Item -LiteralPath (Join-Path $projectRoot 'mods/asi-loader/manifest.json') -Destination (Join-Path $loaderStage 'manifest.json')
[IO.File]::WriteAllBytes((Join-Path $loaderStage 'payload/wininet.dll'), $loaderPayload)
& $ManagerExe pack $loaderStage --out $loaderPath --json
if ($LASTEXITCODE -ne 0) { throw 'Loader package validation failed.' }

$files = [ordered]@{}
foreach ($name in @('LICENSE.txt', 'README.md', 'UPSTREAM.md')) { $files[$name] = [IO.File]::ReadAllBytes((Join-Path $source $name)) }
$files['LOADER-LICENSE.txt'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'mods/asi-loader/LICENSE.txt'))
$files['LOADER-UPSTREAM.md'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'mods/asi-loader/UPSTREAM.md'))
$files['THIRD-PARTY-NOTICES.txt'] = [IO.File]::ReadAllBytes((Join-Path $projectRoot 'THIRD-PARTY-NOTICES.txt'))
$files['mgs3mod.exe'] = [IO.File]::ReadAllBytes($ManagerExe)
$files[$innerName] = [IO.File]::ReadAllBytes($innerPath)
$files[$loaderName] = [IO.File]::ReadAllBytes($loaderPath)
$sumLines = foreach ($name in $files.Keys) { (Get-BytesHash $files[$name]) + '  ' + $name }
$files['SHA256SUMS.txt'] = $utf8.GetBytes(($sumLines -join "`n") + "`n")
$outputStream = [IO.File]::Open($outerPath, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write)
try {
    $outputZip = New-Object IO.Compression.ZipArchive($outputStream, [IO.Compression.ZipArchiveMode]::Create, $true)
    try { foreach ($name in $files.Keys) { Add-ZipBytes $outputZip $name $files[$name] } }
    finally { $outputZip.Dispose() }
    $outputStream.Flush($true)
} finally { $outputStream.Dispose() }
$releaseSums = (Get-FileDigest $innerPath) + '  ' + $innerName + "`n" + (Get-FileDigest $loaderPath) + '  ' + $loaderName + "`n" + (Get-FileDigest $outerPath) + '  ' + $outerName + "`n"
$sumStream = [IO.File]::Open($checksumsPath, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write)
try { $sumBytes = $utf8.GetBytes($releaseSums); $sumStream.Write($sumBytes, 0, $sumBytes.Length); $sumStream.Flush($true) }
finally { $sumStream.Dispose() }
Write-Output $releaseSums
Write-Output "Distribution: $outerPath"
