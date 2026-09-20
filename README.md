# MGS3 Mod Manager

A Windows command-line mod manager for the supported build of Metal Gear Solid 3: Master Collection. Install the compiled release without Go or Git, then import and manage texture mods and ASI plugins.

[Download v0.1.0](https://github.com/nasroykh/mgs3-mod-manager/releases/tag/v0.1.0) · [Public installation verification](docs/releases/v0.1.0-validation.md)

The executable checks three compiled binary fingerprints. Select another installation of the same build with `--game-root <folder>`. Schema 1 supports existing textures; schema 2 adds absent-origin root `.asi` plugins; schema 3 manages the pinned standalone Ultimate ASI Loader. QCamo and its loader can both be installed through this manager, without MGSHDFix. The manager has no network service, dependency downloader, or background process.

**QCamo and standalone ASI loader:** see the [current plan](docs/managed-asi-loader-plan.md), [package instructions](mods/qcamo/README.md), and [validation record](docs/asi-loader-validation.md). Earlier texture and external-loader checkpoints are historical and are not proof of gameplay testing.

The user reported successful in-game operation on 2026-09-20 with QCamo and the standalone loader enabled, without MGSHDFix. This is a local smoke test, not a completed compatibility matrix. See [distribution and tester guidance](docs/distribution.md) for sharing the package.

## Install without Go or Git

Windows AMD64 and PowerShell 5.1+ are required. Download the installer, review it, then run it:

```powershell
$installer = Join-Path $env:TEMP ('mgs3mod-install-' + [guid]::NewGuid().ToString('N') + '.ps1')
Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/nasroykh/mgs3-mod-manager/v0.1.0/install.ps1' -OutFile $installer
notepad $installer
# After reviewing:
powershell -NoProfile -ExecutionPolicy Bypass -File $installer -Version v0.1.0
```

The installer downloads the compiled release, requires a matching SHA-256 checksum, installs to `%LOCALAPPDATA%\Programs\mgs3mod`, and adds a user PATH entry. No admin rights, Git, or Go are needed. It does not install mods or change game files. Open a new terminal and run `mgs3mod --help`. The execution-policy override applies only to that process.

Omit `-Version` for the latest release. For a custom directory without PATH changes, add `-InstallDir 'D:\Tools\mgs3mod' -NoPath`. If lookup fails, use `& "$env:LOCALAPPDATA\Programs\mgs3mod\mgs3mod.exe" --help`. Use `Get-Command mgs3mod -All` to detect conflicting copies.

For manual installation, download `mgs3mod_0.1.0_windows_amd64.zip` and `checksums.txt` from [v0.1.0](https://github.com/nasroykh/mgs3-mod-manager/releases/tag/v0.1.0), compare `Get-FileHash` with its SHA-256 entry, extract it, and run `.\mgs3mod.exe --help`. GitHub's Source code ZIP is not the binary distribution. The executable is not Authenticode-signed; do not disable security software to bypass warnings.

## Quick start: game setup and QCamo

Set your actual game path. Only the [compiled fingerprints](mods/qcamo/UPSTREAM.md#supported-game-profile) are supported, not every game build. Close the game and launcher before mod changes.

```powershell
$game = 'D:\SteamLibrary\steamapps\common\MGS3'
mgs3mod doctor --game-root $game
# Continue only if doctor succeeds:
mgs3mod init --game-root $game
```

Extract `qcamo-1.0.4-standalone-distribution.zip` from the release and open PowerShell there. It includes both packages, manager, licenses, and instructions. Use `.\mgs3mod.exe` instead of `mgs3mod` if you skipped the installer.

```powershell
$game = 'D:\SteamLibrary\steamapps\common\MGS3'
mgs3mod add .\asi-loader-9.7.4.mgs3mod.zip --game-root $game
mgs3mod enable asi-loader --game-root $game
mgs3mod add .\qcamo-1.0.4.mgs3mod.zip --game-root $game
mgs3mod enable qcamo --game-root $game --dry-run
mgs3mod enable qcamo --game-root $game
mgs3mod verify --game-root $game
```

Launch the game and hold **G** or **LB+Y**. No MGSHDFix is installed. Existing manual plugin/loader files and loader INI overrides are refused rather than overwritten. See [controls and known issues](mods/qcamo/README.md).

## Import and manage other mods

Use manager-format packages, not arbitrary Nexus ZIPs or raw DLLs. Use the ID shown by `list`, which can differ from the archive filename:

```powershell
mgs3mod add 'C:\Downloads\my-mod.mgs3mod.zip' --game-root $game
mgs3mod list --game-root $game
mgs3mod enable my-mod --game-root $game --dry-run
mgs3mod enable my-mod --game-root $game
mgs3mod status --game-root $game
mgs3mod verify --game-root $game
mgs3mod disable my-mod --game-root $game
mgs3mod remove my-mod --game-root $game
```

`add` stores disabled. `enable` installs; `--dry-run` validates without writes. `disable` restores originals or absence. `remove` also removes the stored package, retaining baselines/recovery records. Update by disable/remove, then add/enable; no atomic updater exists. Native ASIs execute in the game: install only trusted packages.

## Command behavior and safety

The executable may be invoked from any working directory. Always supply `--game-root '<folder>'` for your installation; omitting it uses the development path `C:\Games\METAL GEAR SOLID 3 - MCV`. State remains bound to its original absolute path; do not copy it between folders or PCs.

After importing schema-2 ASIs or the schema-3 loader, do not downgrade to a manager that lacks that schema: it cannot decode the new journal history, even after the package is removed. The new manager reads the old history without migration.

Close the game and launcher before changing managed files. `add` stores a disabled package. `remove` disables it if necessary, deletes only matching stored package files, and keeps captured original backups and absence records. Overlapping enabled mods are rejected. Different content cannot replace an existing package ID; remove it first. ASI enable requires the enabled, manager-owned `asi-loader` package. Its standalone `wininet.dll` is pinned to Ultimate ASI Loader 9.7.4. Loader disable/removal is blocked while a managed ASI is enabled. Native plugins execute inside the game, not during manager import or packaging.

`restore --baseline` restores captured bytes or recorded absence, including the managed loader and ASIs. It does not reset game settings, modify saves, certify vanilla provenance, or restore the whole installation. Unowned files and runtime-generated logs/dumps are not removed.

```powershell
mgs3mod restore --baseline --game-root $game --dry-run
mgs3mod restore --baseline --game-root $game
```

To uninstall the manager, disable/restore mods first, then remove only its installed program files and exact user PATH entry. Deleting the executable does not undo mods. Keep `.mgs3mod` baselines/recovery records and do not downgrade the manager after importing newer schemas.

Mutation commands support `--dry-run` without creating state or staging files. Commands support `--json`, including failures. Flags may appear after positional arguments. Exit codes: 0 success/no-op, 1 I/O/internal failure, 2 usage/package validation, 3 installation/process guard, 4 conflict/drift, 5 unresolved recovery/corrupt state, 6 manager busy.

## Recovery

Originals and journal records live in `.mgs3mod/` under the game root. Do not manually edit or delete them. The state cache is rebuildable; the validated journal history is authoritative.

```powershell
mgs3mod recover --game-root $game --dry-run --json
mgs3mod recover --game-root $game
```

A ready, uncommitted operation rolls back when every target matches its recorded before/after bytes. A committed operation finishes cleanup. Invalid or contradictory records and external file changes block recovery and preserve evidence. New operations never silently continue through unresolved recovery.

A missing target could have been deleted externally. Plain recovery refuses to recreate it. If the reported unresolved operation has a valid apply intent, explicitly name every missing target:

```powershell
mgs3mod recover --game-root $game --restore-missing textures/flatlist/_win/example.ctxr
```

Repeat the flag for multiple missing targets. This restricted action requires matching core fingerprints and a verified before snapshot. It cannot overwrite an existing file. Reuse the same target list if that recovery itself was interrupted.

The manager serializes its own processes and checks for the game and launcher, but cannot prevent a simultaneous game launch. Multi-file changes are not atomically visible. Power failure, media corruption, malicious concurrent file changes, and damaged journals may require investigation rather than automatic repair. Recovery cleans only the exact owned temporary files of preparations that never became ready. Promoted original backups, resolved transaction records, and their recovery snapshots are retained; there is no baseline garbage collection in this release.

## First non-test mod: QCamo

[QCamo](mods/qcamo/README.md) is packaged from unchanged upstream 1.0.4 with its MIT license and pinned provenance. The recipe under `mods/qcamo/` and `scripts/package-qcamo.ps1` creates a shareable distribution containing the manager, importable package, instructions, notices, and checksums. This integration does not claim original authorship or a source rebuild of QCamo.

Other package authors can use the [runtime package contract](docs/runtime-plugins.md). The shared loader is a separate managed package with explicit absent-file ownership; see the [standalone-loader plan](docs/managed-asi-loader-plan.md).

## Texture test fixture

The local `camo-test` package is a conspicuous magenta recolor of Olive Drab camo, preserving alpha, dimensions, and mip levels. [The recipe](examples/camo-test/README.md) and [validation record](docs/mod-validation.md) describe converter provenance, exact pixel checks, and successful installation/restoration tests. A hash test alone does not prove that the game uses a particular texture. The live appearance check is a separate acceptance gate; current trial state is in [implementation-status.md](docs/implementation-status.md).

## Developer build and releases

End users do not need a compiler. For development, install Go 1.27.1 or a compatible newer toolchain, clone this repository, and run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\build.ps1
```

The script verifies dependencies, runs tests/vet, and builds `dist/mgs3mod.exe`. Optional `-GameRoot '<folder>'` runs a read-only check. It does not deploy or enable mods. Use `mgs3mod pack <folder> --out <archive>` and the [package contract](docs/runtime-plugins.md) to author packages. See [release process](docs/releasing.md) for tagging, automated builds, checksums, and publication. Never publish game-derived payloads, originals, backups, saves, or manager state.
