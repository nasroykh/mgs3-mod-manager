# Build qcamo.asi from pinned QCamo 1.0.4 plus qcamo-face.patch with MSVC, and stage
# the schema-2 package folder and its notices. -Upstream builds the unpatched pinned
# source the same way (toolchain check only; no package).
param(
    [string]$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path,
    [switch]$Upstream
)

$ErrorActionPreference = "Stop"
$version = "1.0.4-face.4"
$pins = [ordered]@{
    qcamo   = @("https://github.com/zexk/mgs3-qcamo.git", "0a8bee4874e6ed9773c96bd48e899056c4c67801")
    imgui   = @("https://github.com/ocornut/imgui.git", "f1cc2ae15e53a861a874c3034aae6798fde194ab")
    minhook = @("https://github.com/TsudaKageyu/minhook.git", "c3fcafdc10146beb5919319d0683e44e3c30d537")
}
$srcRoot = Join-Path $RepoRoot "work\qcamo-face-src"
if ($Upstream) { $work = Join-Path $RepoRoot "work\qcamo-face-build-upstream" }
else { $work = Join-Path $RepoRoot "work\qcamo-face-build" }
$pkg = Join-Path $RepoRoot "work\qcamo-face-pkg"
$notices = Join-Path $RepoRoot "work\qcamo-face-notices"
$zip = Join-Path $RepoRoot "work\qcamo-face-$version.zip"
$patch = Join-Path $PSScriptRoot "qcamo-face.patch"
$utf8 = New-Object Text.UTF8Encoding($false)

# git writes progress and hints to stderr; only its exit code decides failure.
function Invoke-Git {
    $saved = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try { & git -c safe.directory=* @args } finally { $ErrorActionPreference = $saved }
    if ($LASTEXITCODE -ne 0) { throw "git $($args -join ' ') failed: $LASTEXITCODE" }
}

function Get-Sha256([string]$Path) {
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

# Fetch each pinned commit once, then reset to it on every build so the patch always
# lands on a clean upstream tree.
New-Item -ItemType Directory -Force -Path $srcRoot, $work | Out-Null
foreach ($name in $pins.Keys) {
    $url = $pins[$name][0]
    $commit = $pins[$name][1]
    $dir = Join-Path $srcRoot $name
    if (-not (Test-Path -LiteralPath (Join-Path $dir ".git"))) {
        Invoke-Git init -q $dir
        Invoke-Git -C $dir remote add origin $url
    }
    Invoke-Git -C $dir config core.autocrlf false
    $saved = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    & git -c safe.directory=* -C $dir rev-parse -q --verify "$commit^{commit}" | Out-Null
    $have = $LASTEXITCODE -eq 0
    $ErrorActionPreference = $saved
    if (-not $have) { Invoke-Git -C $dir fetch -q --depth 1 origin $commit }
    Invoke-Git -C $dir -c advice.detachedHead=false checkout -q -f --detach $commit
    Invoke-Git -C $dir clean -q -f -d -x
    $head = (& git -c safe.directory=* -C $dir rev-parse HEAD | Out-String).Trim()
    if ($head -ne $commit) { throw "$name is at '$head', not the pinned $commit" }
    Write-Output "$name $head"
}
$q = Join-Path $srcRoot "qcamo"
$imgui = Join-Path $srcRoot "imgui"
$minhook = Join-Path $srcRoot "minhook"

if (-not $Upstream) {
    # The patch is applied with LF endings, matching the upstream blobs, whatever
    # line endings this checkout gave it.
    $lf = Join-Path $work "qcamo-face.patch"
    [IO.File]::WriteAllText($lf, ([IO.File]::ReadAllText($patch) -replace "`r`n", "`n"), $utf8)
    Invoke-Git -C $q apply --check $lf
    Invoke-Git -C $q apply $lf
    Write-Output "applied $patch"
}

$vswhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
$vs = & $vswhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
if (-not $vs) { throw "Visual Studio C++ tools were not found" }
$vcvars = Join-Path $vs "VC\Auxiliary\Build\vcvars64.bat"

$sources = @(
    "$q\src\qcamo.cpp", "$q\src\overlay.cpp", "$q\src\camo_swatch.cpp", "$q\src\camo_index.cpp",
    "$q\src\ctxr.cpp", "$q\src\hud_font.cpp", "$q\src\common\log.cpp",
    "$imgui\imgui.cpp", "$imgui\imgui_draw.cpp", "$imgui\imgui_tables.cpp", "$imgui\imgui_widgets.cpp",
    "$imgui\backends\imgui_impl_dx11.cpp", "$imgui\backends\imgui_impl_win32.cpp",
    "$minhook\src\buffer.c", "$minhook\src\hook.c", "$minhook\src\trampoline.c", "$minhook\src\hde\hde64.c"
) | ForEach-Object { '"' + $_ + '"' }
$asi = Join-Path $work "qcamo.asi"
if (Test-Path -LiteralPath $asi) { Remove-Item -LiteralPath $asi -Force }

# Upstream builds with MinGW through CMake; these are the cl equivalents. NOMINMAX
# keeps windows.h from breaking std::min/std::max, and /utf-8 matches GCC's source
# and execution character set.
$build = @"
@echo off
set "PATH=$(Split-Path -Parent $vswhere);%PATH%"
call "$vcvars" >nul
cd /d "$work"
cl 2>&1 | findstr /C:"Compiler Version" > "$work\cl-version.txt"
echo %VCToolsVersion%> "$work\vc-tools.txt"
echo %WindowsSDKVersion%> "$work\windows-sdk.txt"
cl /nologo /O2 /MT /EHsc /std:c++20 /utf-8 /W3 /Brepro /DNDEBUG /DNOMINMAX /DWIN32_LEAN_AND_MEAN /LD /Fe"$asi" /I"$q\src" /I"$imgui" /I"$imgui\backends" /I"$minhook\include" /I"$minhook\src" /I"$minhook\src\hde" $($sources -join ' ') /link /Brepro d3d11.lib dxgi.lib d3dcompiler.lib dwmapi.lib dbghelp.lib user32.lib gdi32.lib
if errorlevel 1 exit /b 1
"@
$cmd = Join-Path $work "build.cmd"
Set-Content -LiteralPath $cmd -Value $build -Encoding ASCII
& cmd.exe /c $cmd
if ($LASTEXITCODE -ne 0) { throw "qcamo build failed: $LASTEXITCODE" }
$bytes = (Get-Item -LiteralPath $asi).Length
$sha = Get-Sha256 $asi
Write-Output "built $asi ($bytes bytes, sha256 $sha)"
if ($Upstream) { return }

# The package holds only manifest.json and its payload; licenses and provenance
# travel beside it (docs/runtime-plugins.md).
if (Test-Path -LiteralPath $pkg) { Remove-Item -LiteralPath $pkg -Recurse -Force }
if (Test-Path -LiteralPath $zip) { Remove-Item -LiteralPath $zip -Force }
New-Item -ItemType Directory -Force -Path (Join-Path $pkg "payload"), $notices | Out-Null
Copy-Item -LiteralPath $asi -Destination (Join-Path $pkg "payload\qcamo.asi") -Force
@"
{
  "schemaVersion": 2,
  "id": "qcamo-face",
  "version": "$version",
  "name": "QCamo face paint fork (from upstream 1.0.4)",
  "profile": "mgs3-mcv-local-0d585dcc6a67",
  "files": [
    {
      "source": "payload/qcamo.asi",
      "target": "qcamo.asi",
      "originalSha256": "",
      "payloadSha256": "$sha",
      "payloadBytes": $bytes,
      "originalAbsent": true
    }
  ]
}
"@ | Set-Content -LiteralPath (Join-Path $pkg "manifest.json") -Encoding ASCII

$license = [IO.File]::ReadAllText((Join-Path $q "LICENSE")) -replace "`r`n", "`n"
$license += "`n" + ("=" * 80) + "`nqcamo.asi also contains Dear ImGui and MinHook, built from the pinned sources`nlisted in UPSTREAM.md. Their notices follow.`n" + ("=" * 80) + "`n`n"
$license += ([IO.File]::ReadAllText((Join-Path $imgui "LICENSE.txt")) -replace "`r`n", "`n").TrimEnd() + "`n`n" + ("=" * 80) + "`n`n"
$license += ([IO.File]::ReadAllText((Join-Path $minhook "LICENSE.txt")) -replace "`r`n", "`n").TrimStart([char]0xFEFF).TrimEnd() + "`n"
[IO.File]::WriteAllText((Join-Path $notices "LICENSE.txt"), $license, $utf8)

$patchSha = Get-Sha256 (Join-Path $work "qcamo-face.patch")
$clVersion = (Get-Content -LiteralPath (Join-Path $work "cl-version.txt") -Raw).Trim()
$vcTools = (Get-Content -LiteralPath (Join-Path $work "vc-tools.txt") -Raw).Trim()
$sdk = (Get-Content -LiteralPath (Join-Path $work "windows-sdk.txt") -Raw).Trim().TrimEnd('\')
$upstreamMd = @"
# qcamo-face $version provenance

This is a modified build of QCamo, forked from upstream release 1.0.4. It is not the official QCamo binary, is not endorsed by the QCamo authors, and must not be relabelled as the official asset.

- Upstream: https://github.com/zexk/mgs3-qcamo, commit ``0a8bee4874e6ed9773c96bd48e899056c4c67801`` (tag v1.0.4).
- Patch: ``native/qcamo-face/qcamo-face.patch`` in the MGS3 Mod Manager repository, applied with ``git apply``; SHA-256 of the applied (LF) patch ``$patchSha``.
- Dear ImGui ``f1cc2ae15e53a861a874c3034aae6798fde194ab`` (v1.92.9b) and MinHook ``c3fcafdc10146beb5919319d0683e44e3c30d537`` (v1.3.4), the pins in upstream's ``flake.lock``.
- Toolchain: MSVC ``cl`` ($clVersion; VC tools $vcTools), Windows SDK $sdk, built by ``native/qcamo-face/build.ps1``: ``/O2 /MT /EHsc /std:c++20 /utf-8 /W3 /Brepro /DNDEBUG /DNOMINMAX /DWIN32_LEAN_AND_MEAN /LD`` and ``/link /Brepro`` (no timestamps, so a rebuild with the same toolchain is byte-identical), linking d3d11, dxgi, d3dcompiler, dwmapi, dbghelp, user32 and gdi32. Upstream builds with MinGW through CMake/nix instead.
- Output: ``qcamo.asi``, $bytes bytes, SHA-256 ``$sha``. The official 1.0.4 payload is 4,063,819 bytes, SHA-256 ``4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193``; this build is not byte-identical to it and no reproducibility claim is made.
- License: MIT, upstream notice unchanged in ``LICENSE.txt``, followed by the Dear ImGui (MIT) and MinHook (BSD-2-Clause) notices for code linked into ``qcamo.asi``.

## Changes from upstream 1.0.4

- A FACE PAINT list beside the CAMOUFLAGE list. Left/Right (arrow keys, A/D, D-pad, either stick) switch lists; the last list is kept for the session. Owned face paints, best value for the current ground first, cursor on row 0.
- Equipping a face paint changes only the face paint through upstream's face-only change path. Refused when already worn, while Tuxedo is worn (those rows are dimmed, without a gain), or when the change is not accepted.
- Equipping a uniform keeps the face paint Snake has on instead of the automatic best face paint (Tuxedo still forces none).
- While Mask is worn, every equip is refused (rows dimmed): changes made with Mask on crashed at game RVA 0xC8187 (face-only from Mask in 1.0.4-face.1, a uniform keeping Mask in 1.0.4-face.2). Mask comes off in the Survival Viewer; putting it on is allowed.
- Holding D-pad up opens the menu and releasing it closes it, like G; that press does not move the cursor, and L1 alone does not keep such a menu open until L1+Y is pressed in it.
- A D-pad up menu is live: the game keeps running, and the wheel pause is taken only from a queued equip to ``change complete``. G and L1+Y menus pause as upstream. The gate still closes a live menu on any other pause, a game wheel or a cutscene.
- In the open menu the right stick selects and switches lists like the left stick (XInput right stick, and Steam Input's ``ingame_stick_cam_dir`` when it resolves), so the menu works while the left thumb holds D-pad up. Each stick latches one direction with hysteresis (engage past the threshold at 1.5 times the other axis, release below 60% of it).
- Steam D-pad actions fall back from the display names ("Arrow Up" and so on) to the executable's ``ingame_cmn_move_*`` action names; the log says which was used. The close log names the input let go.
- Face paint 4 is labelled DESERT, as the game's own item table names it.
- The camouflage change protocol in ``qcamo.cpp`` is unchanged; its only edits are in the pause bookkeeping for the live menu.

## Package

Package id ``qcamo-face``, version ``$version``, one schema-2 mapping to ``qcamo.asi``, the same target as the ``qcamo`` package. The manager refuses to enable it while another enabled package owns ``qcamo.asi``: disable and remove ``qcamo`` 1.0.4 first. Game profile and executable gate are upstream's: ``mgs3-mcv-local-0d585dcc6a67``, timestamp ``0x6980B92F``.
"@
[IO.File]::WriteAllText((Join-Path $notices "UPSTREAM.md"), ($upstreamMd -replace "`r`n", "`n") + "`n", $utf8)
Write-Output "package folder $pkg"
Write-Output "notices $notices"
