# Build fpvmove.asi, prove the pattern still hits this executable once, and pack schema 2.
param(
    [string]$GameExe = "C:\Games\METAL GEAR SOLID 3 - MCV\METAL GEAR SOLID3.exe",
    [string]$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
)

$ErrorActionPreference = "Stop"
$vswhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
$vs = & $vswhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
if (-not $vs) { throw "Visual Studio C++ tools were not found" }
$vcvars = Join-Path $vs "VC\Auxiliary\Build\vcvars64.bat"
$src = Join-Path $PSScriptRoot "fpvmove.c"
$work = Join-Path $RepoRoot "work\fpv-move-build"
$pkg = Join-Path $RepoRoot "work\fpv-move-pkg"
$zip = Join-Path $RepoRoot "work\fpv-move-0.8.4.zip"
New-Item -ItemType Directory -Force -Path $work, (Join-Path $pkg "payload") | Out-Null
if (Test-Path -LiteralPath $zip) { Remove-Item -LiteralPath $zip -Force }

$build = @"
call "$vcvars" >nul
cl /nologo /O2 /W3 /MT /DFPVMOVE_TEST /Fo"$work\fpvmove-test.obj" /Fe"$work\fpvmove-test.exe" "$src"
if errorlevel 1 exit /b 1
"$work\fpvmove-test.exe" "$GameExe"
if errorlevel 1 exit /b 1
cl /nologo /O2 /W3 /MT /LD /Fo"$work\fpvmove.obj" /Fe"$work\fpvmove.asi" "$src"
if errorlevel 1 exit /b 1
"@
$cmd = Join-Path $work "build.cmd"
Set-Content -LiteralPath $cmd -Value $build -Encoding ASCII
& cmd.exe /c $cmd
if ($LASTEXITCODE -ne 0) { throw "fpvmove build failed: $LASTEXITCODE" }

Copy-Item -LiteralPath (Join-Path $work "fpvmove.asi") -Destination (Join-Path $pkg "payload\fpvmove.asi") -Force
@'
{
  "schemaVersion": 2,
  "id": "fpv-move",
  "version": "0.8.4",
  "name": "Delta aim, move, and fire",
  "profile": "mgs3-mcv-local-0d585dcc6a67",
  "files": [
    {
      "source": "payload/fpvmove.asi",
      "target": "fpvmove.asi",
      "originalSha256": "",
      "payloadSha256": "",
      "payloadBytes": 0,
      "originalAbsent": true
    }
  ]
}
'@ | Set-Content -LiteralPath (Join-Path $pkg "manifest.json") -Encoding ASCII
Write-Output "built $(Join-Path $work 'fpvmove.asi')"
