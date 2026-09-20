$ErrorActionPreference = 'Stop'
$projectRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$go = 'C:\Program Files\Go\bin\go.exe'
$env:GOTOOLCHAIN = 'local'
$env:GOWORK = 'off'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
$env:GOCACHE = Join-Path $projectRoot '.cache/go-build'
$env:GOMODCACHE = Join-Path $projectRoot '.cache/go-mod'

function Invoke-CheckedGo {
    & $go @args
    if ($LASTEXITCODE -ne 0) { throw "Go command failed with exit code $LASTEXITCODE." }
}

Push-Location -LiteralPath $projectRoot
try {
    New-Item -ItemType Directory -Path './dist' -Force | Out-Null
    Invoke-CheckedGo mod download
    Invoke-CheckedGo mod verify
    Invoke-CheckedGo test ./...
    Invoke-CheckedGo vet ./...
    Invoke-CheckedGo build -mod=readonly -trimpath -o ./dist/mgs3mod.exe ./cmd/mgs3mod
    & './dist/mgs3mod.exe' doctor --json
    if ($LASTEXITCODE -ne 0) { throw 'Installation check failed. Do not deploy this build.' }
    $stream = [System.IO.File]::OpenRead((Join-Path $projectRoot 'dist/mgs3mod.exe'))
    $hasher = [System.Security.Cryptography.SHA256]::Create()
    try { $digest = [BitConverter]::ToString($hasher.ComputeHash($stream)).Replace('-', '').ToLowerInvariant() }
    finally { $stream.Dispose(); $hasher.Dispose() }
    Write-Output "mgs3mod.exe SHA-256: $digest"
} finally {
    Pop-Location
}
