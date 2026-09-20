# MGS3 Mod Manager

A Windows Go CLI for the supported build of Metal Gear Solid 3 Master Collection. The default local installation is:

`C:\Games\METAL GEAR SOLID 3 - MCV`

The executable checks three compiled binary fingerprints. Select another installation of the same build with `--game-root <folder>`. Schema 1 supports existing textures; schema 2 adds absent-origin root `.asi` plugins; schema 3 manages the pinned standalone Ultimate ASI Loader. QCamo and its loader can both be installed through this manager, without MGSHDFix. The manager has no network service, dependency downloader, or background process.

**QCamo and standalone ASI loader:** see the [current plan](docs/managed-asi-loader-plan.md), [package instructions](mods/qcamo/README.md), and [validation record](docs/asi-loader-validation.md). Earlier texture and external-loader checkpoints are historical and are not proof of gameplay testing.

The user reported successful in-game operation on 2026-09-20 with QCamo and the standalone loader enabled, without MGSHDFix. This is a local smoke test, not a completed compatibility matrix. See [distribution and tester guidance](docs/distribution.md) for sharing the package.

## Build

Install Go 1.27.1 or a compatible newer toolchain on PATH. The one dependency is pinned in `go.mod` and `go.sum`. From a fresh source checkout, run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\build.ps1
```

The process-scoped execution-policy flag permits this local build script without changing persistent policy. The script uses project-local caches, downloads and verifies the pinned dependency, runs tests and vet, and builds `dist/mgs3mod.exe` without machine-specific VCS stamping. Pass `-GameRoot '<folder>'` for an optional read-only installation check. It neither installs the executable into the game root nor enables any mod. Build and authoring files stay in ignored `.cache/`, `work/`, and `dist/`. Do not publish game-derived payloads, backups, or original textures.

## Commands

The executable may be invoked from any working directory. Append `--game-root '<folder>'` to every installation command when using a different installation path. The relative examples below assume the project root and default local installation. Existing state remains bound to the absolute path where it was initialized; do not copy state between installations.

After importing schema-2 ASIs or the schema-3 loader, do not downgrade to a manager that lacks that schema: it cannot decode the new journal history, even after the package is removed. The new manager reads the old history without migration.

```powershell
.\dist\mgs3mod.exe doctor
.\dist\mgs3mod.exe init
.\dist\mgs3mod.exe pack .\work\my-package --out .\dist\my-package.zip
.\dist\mgs3mod.exe add .\dist\my-package.zip
.\dist\mgs3mod.exe enable camo-test --dry-run
.\dist\mgs3mod.exe enable camo-test
.\dist\mgs3mod.exe status
.\dist\mgs3mod.exe disable camo-test
.\dist\mgs3mod.exe restore --baseline
.\dist\mgs3mod.exe remove camo-test
.\dist\mgs3mod.exe verify --json
```

Close the game and launcher before changing managed files. `add` stores a disabled package. `remove` disables it if necessary, deletes only matching stored package files, and keeps captured original backups and absence records. Overlapping enabled mods are rejected. Different content cannot replace an existing package ID; remove it first. ASI enable requires the enabled, manager-owned `asi-loader` package. Its standalone `wininet.dll` is pinned to Ultimate ASI Loader 9.7.4. Loader disable/removal is blocked while a managed ASI is enabled. Native plugins execute inside the game, not during manager import or packaging.

`restore --baseline` restores captured bytes or recorded absence, including the managed loader and ASIs. It does not reset game settings, modify saves, certify vanilla provenance, or restore the whole installation. Unowned files and runtime-generated logs/dumps are not removed.

Mutation commands support `--dry-run` without creating state or staging files. Commands support `--json`, including failures. Flags may appear after positional arguments. Exit codes: 0 success/no-op, 1 I/O/internal failure, 2 usage/package validation, 3 installation/process guard, 4 conflict/drift, 5 unresolved recovery/corrupt state, 6 manager busy.

## Recovery

Originals and journal records live in `.mgs3mod/` under the game root. Do not manually edit or delete them. The state cache is rebuildable; the validated journal history is authoritative.

```powershell
.\dist\mgs3mod.exe recover --dry-run --json
.\dist\mgs3mod.exe recover
```

A ready, uncommitted operation rolls back when every target matches its recorded before/after bytes. A committed operation finishes cleanup. Invalid or contradictory records and external file changes block recovery and preserve evidence. New operations never silently continue through unresolved recovery.

A missing target could have been deleted externally. Plain recovery refuses to recreate it. If the reported unresolved operation has a valid apply intent, explicitly name every missing target:

```powershell
.\dist\mgs3mod.exe recover --restore-missing textures/flatlist/_win/example.ctxr
```

Repeat the flag for multiple missing targets. This restricted action requires matching core fingerprints and a verified before snapshot. It cannot overwrite an existing file. Reuse the same target list if that recovery itself was interrupted.

The manager serializes its own processes and checks for the game and launcher, but cannot prevent a simultaneous game launch. Multi-file changes are not atomically visible. Power failure, media corruption, malicious concurrent file changes, and damaged journals may require investigation rather than automatic repair. Recovery cleans only the exact owned temporary files of preparations that never became ready. Promoted original backups, resolved transaction records, and their recovery snapshots are retained; there is no baseline garbage collection in this release.

## First non-test mod: QCamo

[QCamo](mods/qcamo/README.md) is packaged from unchanged upstream 1.0.4 with its MIT license and pinned provenance. The recipe under `mods/qcamo/` and `scripts/package-qcamo.ps1` creates a shareable distribution containing the manager, importable package, instructions, notices, and checksums. This integration does not claim original authorship or a source rebuild of QCamo.

Other package authors can use the [runtime package contract](docs/runtime-plugins.md). The shared loader is a separate managed package with explicit absent-file ownership; see the [standalone-loader plan](docs/managed-asi-loader-plan.md).

## Texture test fixture

The local `camo-test` package is a conspicuous magenta recolor of Olive Drab camo, preserving alpha, dimensions, and mip levels. [The recipe](examples/camo-test/README.md) and [validation record](docs/mod-validation.md) describe converter provenance, exact pixel checks, and successful installation/restoration tests. A hash test alone does not prove that the game uses a particular texture. The live appearance check is a separate acceptance gate; current trial state is in [implementation-status.md](docs/implementation-status.md).
