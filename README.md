# MGS3 local mod manager

A Windows Go CLI for this specific installation of Metal Gear Solid 3 Master Collection:

`C:\Games\METAL GEAR SOLID 3 - MCV`

The executable checks three compiled binary fingerprints. It manages existing `.ctxr` textures only under `textures/flatlist/_win/` and `hqtex/flatlist/_win/`. It has no root override, network service, runtime dependency downloader, or background process.

**Current validation status:** see [implementation-status.md](docs/implementation-status.md). The design document describes acceptance requirements; it is not proof that the live mod trial has passed.

## Build

Go 1.27.1 is installed at `C:\Program Files\Go\bin\go.exe`. The one dependency is pinned in `go.mod` and `go.sum`. From a fresh source checkout, run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\build.ps1
```

The process-scoped execution-policy flag permits this local build script without changing the machine's persistent policy. The script uses project-local caches, downloads and verifies the pinned dependency, runs tests and vet, builds `dist/mgs3mod.exe`, and performs a read-only installation check. It neither installs the executable into the game root nor enables any mod. Build and authoring files stay in ignored `.cache/`, `work/`, and `dist/` directories. Do not publish game-derived payloads, backups, or original textures with the source.

## Commands

The executable may be invoked from any working directory because the game root is compiled in. The relative command examples below assume the project root.

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

Close the game and launcher before changing textures. `add` stores a disabled package. `remove` disables it if necessary, deletes only the matching stored package files, and keeps captured original backups. Overlapping enabled mods are rejected. Different content cannot replace an existing package ID; remove it first.

`restore --baseline` restores only local files this manager captured. It does not reset game settings, modify saves, certify vanilla provenance, or restore the whole installation.

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

## First mod

The local `camo-test` package is a conspicuous magenta recolor of Olive Drab camo, preserving alpha, dimensions, and mip levels. [The recipe](examples/camo-test/README.md) and [validation record](docs/mod-validation.md) describe converter provenance, exact pixel checks, and successful installation/restoration tests. A hash test alone does not prove that the game uses a particular texture. The live appearance check is a separate acceptance gate; current trial state is in [implementation-status.md](docs/implementation-status.md).
