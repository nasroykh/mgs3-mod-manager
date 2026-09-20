# Research and design decisions

Research date: 2026-09-19. Scope: a Windows Go CLI for this one MGS3 Master Collection installation, plus a small texture mod. This is a design study, not a certification of third-party tools or of the installed game's provenance.

Current-status note, 2026-09-20: implementation, fault-injection testing, converter repair, packaging, and the HQ visual activation trial were completed after this research phase. The evidence audit below records both the original research limits and their later disposition.

## Sources and findings

| Source | Evidence | Decision for this project |
|---|---|---|
| [Go: traversal-resistant filesystem APIs](https://go.dev/blog/osroot) | `os.Root` contains relative operations within a directory; lexical sanitization alone does not address symlink races. | Use root-relative filesystem operations and strict package path validation together. |
| [Go os documentation](https://pkg.go.dev/os#Rename), also installed `src/os/file.go:435` | Windows rename does not have the atomic guarantee sometimes assumed by Unix applications. | Keep backups and a durable operation journal; do not equate a renamed file with an atomic multi-file transaction. |
| [Microsoft ReplaceFileW](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-replacefilew) | Replacement can fail after changing file names or attributes; replacement and backup are subject to volume constraints. | Use `ReplaceFileW` for existing targets so their security metadata survives replacement. Keep journal verification and recovery because this does not create an atomic multi-file transaction. |
| [Microsoft LockFileEx](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-lockfileex) | An exclusive fail-immediately lock supports single-writer access. The OS releases locks on handle close/process termination, but release may be delayed. | Use a dedicated handle-backed lock and explicit release; never infer ownership from a stale lock file or PID alone. This does not prevent the game from launching. |
| [Microsoft process snapshots](https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-createtoolhelp32snapshot) and [QueryFullProcessImageNameW](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-queryfullprocessimagenamew) | Process enumeration is a snapshot; executable paths require an accessible query handle. | Filter matching executable names, resolve their paths, and block if a relevant candidate cannot be inspected. Handle later file-sharing failures separately. |
| [x/sys Windows API](https://pkg.go.dev/golang.org/x/sys/windows) and [tagged module file](https://github.com/golang/sys/blob/v0.48.0/go.mod) | The tagged module requires Go 1.26.0. | Pin v0.48.0 with the installed Go 1.27.1 and record module checksums when building. |
| [Command Line Interface Guidelines](https://clig.dev/) | The guide recommends useful help, meaningful exit statuses, separating result output from diagnostics, and optional JSON output. | Provide concise human output, `--json`, `--dry-run`, stable error categories, and no interactive prompt loops. |
| [MGS Mod Manager README](https://github.com/ANTIBigBoss/MGS-MC-Mod-Manager-and-Tool#readme) | The author documents file installation/uninstallation and initial backups, explicitly requiring vanilla files for a vanilla backup. | Copy the concept of a separate library and baseline, but call our snapshot a local baseline. Never infer official defaults from the current directory. |
| [MGS Mod Manager ConfigManager.cs](https://github.com/ANTIBigBoss/MGS-MC-Mod-Manager-and-Tool/blob/main/ConfigManager.cs) | `ModMapping` has `ModFile` and `TargetPath`; tracking separates active mods, mappings, and replaced files. | Use explicit source-to-target mappings and an ownership ledger. No filename guessing or silent replacement of another mod. |
| [MGSHDFix](https://github.com/ShizCalev/MGSHDFix#readme) | The runtime fix uses loader/plugin files and exposes game-specific controls and settings. | Runtime mods need dependency and shared-file ownership rules. Keep those out of the first texture-only release. |
| [CtxrTool](https://github.com/Jayveer/CtxrTool#readme) | CTXR can be exported and repacked. The author describes reverse conversion as experimental, recommends DDS in 8.8.8.8 ARGB, and preserves additional data through a `.param` sidecar. | Pin the converter, retain sidecars, keep resolution and alpha, and validate an unchanged round trip before creating a mod. |
| [Master Collection Noesis plugin](https://github.com/Jayveer/MGS-Master-Collection-Noesis#readme) | The project documents loading Master Collection models and animations. | Viewing support is not evidence of a complete native model-writing workflow. Do not make model authoring a dependency. |
| [GCX tool](https://github.com/Jayveer/Gcx#readme) | The documented MGS3 mode decompiles; compilation is not provided by this tool. | Do not make scripting or new missions part of this deliverable. |

Source links above point to mutable upstream branches or documentation. The completed implementation records the exact revision and binary hash of every tool actually used. The source review alone does not prove compatibility with the installed game.

## Chosen patterns

1. **A fixed installation profile:** one root path and the three observed binary hashes. No root discovery, game selection, storefront integration, or public `--game-dir` override.
2. **Explicit packages:** a small JSON manifest maps each payload file to one existing texture target and records both original and replacement hashes.
3. **Separate storage:** development sources, imported mods, original backups, and the deployed game files have different directories and responsibilities.
4. **Immutable originals:** capture only affected targets, once, before changing them. Re-verify backup bytes before deployment and restoration. Never replace an original backup with an already-modded file.
5. **One owner per target:** reject conflicts before deployment. Load order and automatic merging add complexity that the first mod does not need.
6. **Journaled changes:** durable intent precedes writes; verified results precede commit. Recovery uses recorded hashes and does not blindly overwrite unexpected files.
7. **A reproducible mod:** preserve the original texture's structural properties, record the transformation, and test both visible appearance and exact restoration.

These are project design choices, not claims that every existing manager uses the same rules.

## Evidence audit

- **CORRECT**: Local Go and binary identity are established by the commands and full-file hashes in [installation-evidence.md](installation-evidence.md).
- **CORRECT**: Go documents non-atomic rename behavior on non-Unix platforms; the installed source carries the same warning.
- **CORRECT**: The cited CtxrTool README describes conversion in both directions and explicitly qualifies reverse conversion as experimental.
- **CORRECT**: The cited manager source represents explicit file mappings and tracks active mods. No broad correctness verdict about that manager is made here.
- **CORRECT**: The later patched CtxrTool roundtrip is byte-exact, and the user confirmed its authored HQ recolor produced a normally rendered magenta uniform. Detailed screenshots and separately itemized close/distant observations were not captured; see [converter-repair.md](converter-repair.md) and [mod-validation.md](mod-validation.md).
- **UNVERIFIABLE**: The currently installed assets are official factory defaults; local hashing provides identity, not provenance.
- **CORRECT**: The later CLI implementation and synthetic fault matrix cover the declared crash-recovery boundaries; see [implementation-review.md](implementation-review.md).
- **UNVERIFIABLE**: Performance outside the tested local workflows was not benchmarked and no general performance guarantee is made.

## Research scope exclusions

No files were installed into the game, no converters were executed, and the game was not launched during this planning task. The work does not establish complete support for arbitrary Nexus packages, shared DLL loaders, model import, animation authoring, audio authoring, or script compilation. They are not needed to validate a first texture mod.
