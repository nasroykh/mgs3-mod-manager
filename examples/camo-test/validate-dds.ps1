param(
    [Parameter(Mandatory = $true)]
    [string] $Path,
    [string] $ParamPath
)

# Read-only gate. This script never writes DDS, CTXR, param, or game files.
$bytes = [System.IO.File]::ReadAllBytes((Resolve-Path -LiteralPath $Path))
if ($bytes.Length -lt 0x80) { throw 'DDS file is shorter than 128-byte header.' }

function Read-U32([byte[]] $InputBytes, [int] $Offset) {
    return [uint64]$InputBytes[$Offset] + [uint64]$InputBytes[$Offset + 1] * 0x100 +
        [uint64]$InputBytes[$Offset + 2] * 0x10000 + [uint64]$InputBytes[$Offset + 3] * 0x1000000
}

function Require([bool] $Condition, [string] $Message) {
    if (-not $Condition) { throw $Message }
}

Require (($bytes[0] -eq 0x44) -and ($bytes[1] -eq 0x44) -and ($bytes[2] -eq 0x53) -and ($bytes[3] -eq 0x20)) 'DDS magic is invalid.'
Require ((Read-U32 $bytes 4) -eq 0x7C) 'DDS header size is not 0x7C.'
Require (((Read-U32 $bytes 80) -band 0x41) -eq 0x41) 'DDS pixel format lacks RGB plus alpha-pixel flags.'
Require ((Read-U32 $bytes 88) -eq 32) 'DDS pixel format is not 32 bits per pixel.'
Require ((Read-U32 $bytes 92) -eq 0x00FF0000) 'Unexpected red channel mask.'
Require ((Read-U32 $bytes 96) -eq 0x0000FF00) 'Unexpected green channel mask.'
Require ((Read-U32 $bytes 100) -eq 0x000000FF) 'Unexpected blue channel mask.'
Require ((Read-U32 $bytes 104) -eq [uint64]4278190080) 'Unexpected alpha channel mask.'
Require ((Read-U32 $bytes 28) -gt 0) 'DDS mip count is zero.'

if ($ParamPath) {
    $paramLines = @(Get-Content -LiteralPath $ParamPath)
    Require ($paramLines.Count -eq 13) 'CtxrTool param sidecar must contain 13 parameter lines.'
}

[pscustomobject]@{
    path = (Resolve-Path -LiteralPath $Path).Path
    bytes = $bytes.Length
    width = Read-U32 $bytes 16
    height = Read-U32 $bytes 12
    mipCount = Read-U32 $bytes 28
    bitCount = Read-U32 $bytes 88
    masks = 'R=0x00FF0000;G=0x0000FF00;B=0x000000FF;A=0xFF000000'
    alphaPreservedByMetadata = $true
    paramLines = if ($ParamPath) { $paramLines.Count } else { $null }
}
