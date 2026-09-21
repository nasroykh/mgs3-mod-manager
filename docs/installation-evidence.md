# Installation evidence

Inspected on 2026-09-19; launcher fingerprints were added and rechecked on 2026-09-21. These observations identify a local installation; they do not certify an official or unmodified release.

## Target

- Directory: `C:\Games\METAL GEAR SOLID 3 - MCV`
- Go: `go version go1.27.1 windows/amd64`
- Go executable: `C:\Program Files\Go\bin\go.exe`, available through `PATH`.
- Go environment: `GOOS=windows`, `GOARCH=amd64`, `GOROOT=C:\Program Files\Go`.
- Filesystem: NTFS on C:.
- The directory is not a Git repository. No repository was initialized around the game files.
- No matching `METAL GEAR SOLID3` or `launcher` process was running at inspection time. This must be checked again before any deployment.
- Visual Studio discovery reports `C:\Program Files\Microsoft Visual Studio\2022\Community` with the C++ tools component. The later converter repair built successfully with that toolchain; its pinned source, patch, binary hash, and byte-exact round-trip evidence are recorded in [converter-repair.md](converter-repair.md).

## Binary fingerprints

| File | Bytes | File version | SHA-256 |
|---|---:|---|---|
| `METAL GEAR SOLID3.exe` | 12948040 | `3.0.0.0` | `0d585dcc6a671be5d64d3d0a856c53f9ee0e58e7e4993f76dff29772c7a4bc80` |
| `Engine.dll` | 1387520 | not populated | `4067774bd2945dfab1a81ee0f657b3b6c9414b1b363d29830652e6d93c516996` |
| `Renderer.dll` | 7320718 | not populated | `663199bce1a252861369710d62219a73d2855043ec955ab7e8a926ea13986ac1` |
| `launcher.exe` | 653824 | `2021.3.16.4200023` | `e061111cef605bdbf0ea7bc9cf686a1d29e317bf1da923e3c92c52f523479784` |
| `launcher_Data/Managed/Assembly-CSharp.dll` | 459776 | `0.0.0.0` | `3f4de01b2de9e9efc31001c0ad376b294eaeec5dd092f7e681204e93684af945` |

The executable's metadata is not a Steam build identifier. The manager build is bound to all five fingerprints and the fixed directory. Each managed asset also has its own expected original hash.

## Reproduction commands

The first mod candidates were also checked by streaming SHA-256:

| Candidate | Bytes | SHA-256 |
|---|---:|---|
| `hqtex/flatlist/_win/sna_def_olive.bmp.ctxr` | 22370144 | `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d` |
| `textures/flatlist/_win/sna_def_olive.bmp.ctxr` | 1398560 | `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049` |

Use the same hash loop below with those relative paths to reproduce. `rg -n -F 'sna_def_olive.bmp.ctxr' 'sp/slot/camoufla-normal/bp_assets.txt' 'sp/stage/v000a_0/bp_assets.txt'` returned the camo mapping at line 11 and stage references at lines 383–384. The manifests reference the logical texture path, not proof that the HQ or standard physical copy is active. No locale or graphics setting has been changed.

```powershell
& 'C:\Program Files\Go\bin\go.exe' version
& 'C:\Program Files\Go\bin\go.exe' env GOOS GOARCH GOROOT GOPATH
Get-Command go,git,rg -ErrorAction SilentlyContinue
git rev-parse --show-toplevel
Get-ChildItem -LiteralPath assets -Directory

$binaryFacts = foreach ($fileName in @('METAL GEAR SOLID3.exe','Engine.dll','Renderer.dll','launcher.exe','launcher_Data\Managed\Assembly-CSharp.dll')) {
    $item = Get-Item -LiteralPath $fileName
    $hashEngine = [System.Security.Cryptography.SHA256]::Create()
    $stream = [System.IO.File]::OpenRead($item.FullName)
    try {
        $digest = [BitConverter]::ToString($hashEngine.ComputeHash($stream)).Replace('-','').ToLowerInvariant()
    } finally {
        $stream.Dispose()
        $hashEngine.Dispose()
    }
    [pscustomobject]@{
        Name = $item.Name
        Bytes = $item.Length
        Version = $item.VersionInfo.FileVersion
        SHA256 = $digest
    }
}
$binaryFacts | Format-List

$gameProcesses = @(Get-Process -Name 'METAL GEAR SOLID3','launcher' -ErrorAction SilentlyContinue)
'MatchingProcessCount=' + $gameProcesses.Count

& 'C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe' `
    -latest -products '*' -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 `
    -property installationPath
```

## Audit results and limitations

- **CORRECT**: Go is installed and runnable. Evidence: the version and environment commands above.
- **CORRECT**: The five binaries have the listed fingerprints. Counting method: one complete streaming SHA-256 computation per file, plus filesystem byte length.
- **CORRECT**: No Git repository contains this directory. Evidence: `git rev-parse --show-toplevel` returned `fatal: not a git repository (or any of the parent directories): .git`.
- **CORRECT**: Matching game/launcher process count was zero at inspection. Evidence: `MatchingProcessCount=0` from the command above. This is a point-in-time observation, not a launch lock.
- **UNVERIFIABLE**: Official vanilla status and precise storefront build identity. Local converter round-trip correctness was later proved byte-for-byte, and the user later confirmed the HQ recolor rendered as a normal magenta uniform in game. Matched screenshots and a separate post-restoration visual launch were not captured.

Initial checks encountered two tooling issues: an invalid PowerShell pipeline after a `foreach` statement, then an unavailable `Get-FileHash` cmdlet. The corrected command above used .NET SHA-256 successfully. A `Get-PSDrive` capacity query did not return usable capacity; `.NET System.IO.DriveInfo` reported NTFS and free space successfully. No failure was interpreted as a game-file problem.

Installed Go source confirms `os.Root.Rename` at `C:\Program Files\Go\src\os\root.go:223`. The warning at `C:\Program Files\Go\src\os\file.go:435` states that rename is not guaranteed atomic on non-Unix platforms. Do not base recovery on an assumed atomic Windows rename.
