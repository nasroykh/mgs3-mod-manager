param([string]$GameRoot = '')
$ErrorActionPreference = 'Stop'
$projectRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$go = (Get-Command go -ErrorAction Stop).Source
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

function Get-Sha256([string]$Path) {
    $stream = [System.IO.File]::OpenRead($Path)
    $hasher = [System.Security.Cryptography.SHA256]::Create()
    try { return [BitConverter]::ToString($hasher.ComputeHash($stream)).Replace('-', '').ToLowerInvariant() }
    finally { $stream.Dispose(); $hasher.Dispose() }
}

# The window app is built with a pinned Wails CLI installed into the project cache.
$wailsVersion = 'v2.14.0'

Push-Location -LiteralPath $projectRoot
try {
    New-Item -ItemType Directory -Path './dist' -Force | Out-Null
    Invoke-CheckedGo mod download
    Invoke-CheckedGo mod verify
    Invoke-CheckedGo test ./...
    Invoke-CheckedGo vet ./...
    Invoke-CheckedGo build -mod=readonly -buildvcs=false -trimpath -o ./dist/mgs3mod.exe ./cmd/mgs3mod
    if ($GameRoot) {
        & './dist/mgs3mod.exe' doctor --game-root $GameRoot --json
        if ($LASTEXITCODE -ne 0) { throw 'Installation check failed. Do not deploy this build.' }
    }
    Write-Output "mgs3mod.exe SHA-256: $(Get-Sha256 (Join-Path $projectRoot 'dist/mgs3mod.exe'))"

    $savedGobin, $savedGoflags = $env:GOBIN, $env:GOFLAGS
    $env:GOBIN = Join-Path $projectRoot '.cache/bin'
    try { Invoke-CheckedGo install "github.com/wailsapp/wails/v2/cmd/wails@$wailsVersion" }
    finally { $wails = Join-Path $env:GOBIN 'wails.exe'; $env:GOBIN = $savedGobin }
    $reported = (& $wails version | Select-Object -First 1).Trim()
    if ($reported -cne $wailsVersion) { throw "Wails CLI reports $reported, expected $wailsVersion." }
    $guiDir = Join-Path $projectRoot 'cmd/mgs3mod-gui'
    $env:GOFLAGS = '-mod=readonly -buildvcs=false'
    Push-Location -LiteralPath $guiDir
    try {
        # -m and -nosyncgomod keep go.mod untouched; the frontend is static files.
        & $wails build -clean -webview2 download -trimpath -m -nosyncgomod -skipbindings
        if ($LASTEXITCODE -ne 0) { throw "Wails build failed with exit code $LASTEXITCODE." }
    } finally {
        Pop-Location
        $env:GOFLAGS = $savedGoflags
    }
    Copy-Item -LiteralPath (Join-Path $guiDir 'build/bin/mgs3mod-gui.exe') -Destination './dist/mgs3mod-gui.exe' -Force
    Write-Output "mgs3mod-gui.exe SHA-256: $(Get-Sha256 (Join-Path $projectRoot 'dist/mgs3mod-gui.exe'))"
} finally {
    Pop-Location
}
