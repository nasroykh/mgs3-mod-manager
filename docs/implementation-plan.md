# MGS3 local mod manager and first texture mod

Status: acceptance plan. The CLI and an offline test package have now been implemented; see [implementation-status.md](implementation-status.md) for current evidence and remaining live-test gates.

Date: 2026-09-19.

## 1. Outcome and acceptance boundary

Build a Windows/amd64 Go executable named `mgs3mod.exe` that stores local mod packages, enables and disables them, removes them, verifies their files, and restores the original bytes of files it manages. Deliver one conspicuous camouflage texture mod and demonstrate its appearance and exact rollback in this installation.

The only supported installation is `C:\Games\METAL GEAR SOLID 3 - MCV`. Use the three complete binary fingerprints in [installation-evidence.md](installation-evidence.md), not just the executable's `3.0.0.0` version field. The compiled profile ID is `mgs3-mcv-local-0d585dcc6a67`; this is an opaque label, not a substitute for checking all three full hashes.

The first release supports existing regular `.ctxr` file replacements beneath `textures/flatlist/_win/` and `hqtex/flatlist/_win/`. It can manage multiple non-overlapping texture mods, but cannot add new game files. Start with one verified target. Do not silently broaden this scope when importing a package.

Explicitly defer DLL/ASI loaders, executable patches, new files, deletion mods, load order, shared dependencies, configuration merging, profiles, remote downloads, other game versions, and other directories. Saves, game settings, authentication files, redistributables, and launcher files are outside the managed namespace.

`restore --baseline` means restoring the local bytes captured before this manager changed them. It does not reset the whole game, erase saves, manufacture official files, or claim factory settings. Since the first release changes no settings, a separate factory-config command would be misleading and will not exist.

## 2. Technology and layout

Use the installed Go 1.27.1 toolchain with `GOTOOLCHAIN=local`. Module name: `mgs3mod`; `go.mod` directive: `go 1.27.1`. Use the standard library for command parsing (`flag`), JSON, ZIP, hashing, and filesystem operations. Pin `golang.org/x/sys v0.48.0` for Windows locking, process checks, and file-identity checks; its upstream `go.mod` requires Go 1.26.0, within the installed toolchain. Record its verified checksums in `go.sum` during implementation. No web service, database, CGo in the release binary, or resident process.

```text
C:\Games\METAL GEAR SOLID 3 - MCV\
  mgs3mod.exe                        # built CLI
  mgs3-mod-manager\                 # source project, not managed game content
    cmd\mgs3mod\
    internal\cli\                   # parsing, output, exit codes
    internal\profile\               # fixed root, core hashes, target policy
    internal\packagefmt\            # manifest and ZIP/folder validation
    internal\manager\               # library, ownership, operation planning
    internal\transaction\           # journal, commit, recovery
    internal\winfs\                 # Root operations, Windows guards
    docs\
    examples\camo-test\             # recipe and authoring manifest
    work\                           # local original/converted assets; ignored
    dist\                           # local mod ZIP and CLI; ignored
    .cache\                         # build/module caches; ignored
  .mgs3mod\                         # CLI-owned state; never a mod target
    installation.json
    lock
    library\<id>\                   # validated manifest and immutable payload
    baseline\<sha256>               # original bytes; retained on mod removal
    transactions\<sequence-id>\    # operation records and retained recovery data
    state.json                      # disposable cache of committed state
```

Keep the source project independently versionable; do not put the game directory into Git. Ignore game-derived assets, backups, ZIPs, tool binaries, caches, and generated output. A future source repository contains code, docs, manifests without payloads, and recipes only. Do not automatically publish anything.

The executable uses its fixed compiled root regardless of the shell's current directory. Internal constructors may accept temporary fixture roots for tests; the release CLI exposes no override or environment-variable bypass.

## 3. CLI contract

| Command | Result |
|---|---|
| `mgs3mod doctor` | Read-only prerequisite, installation fingerprint, state, and process checks. Works before initialization. |
| `mgs3mod init` | Validate the fixed installation and initialize state. Capture no whole-game backup. Idempotent; never replace an existing baseline. |
| `mgs3mod pack <folder> --out <zip>` | Validate an authoring manifest, compute payload hashes, and create a normalized local package. No game changes. |
| `mgs3mod add <zip-or-folder>` | Import a validated package into the library in disabled state. Never enable implicitly. |
| `mgs3mod list` | Show mod ID, version, enabled/disabled status, and target count. |
| `mgs3mod status` | Show ownership, drift, pending recovery, and compatibility status. |
| `mgs3mod enable <id>` | Preflight every target, capture originals when first needed, and apply the complete package. |
| `mgs3mod disable <id>` | Restore its original targets and retain the package. Already disabled is a successful no-op. |
| `mgs3mod remove <id>` | Disable if active, then remove its library entry and unreferenced payload. Preserve original backups. |
| `mgs3mod restore --baseline` | Restore all enabled managed targets in one operation; keep packages available but disabled. |
| `mgs3mod verify` | Check core fingerprints, every tracked baseline, stored package payloads, and current owned/baseline targets. |
| `mgs3mod recover` | Resolve an interrupted transaction using the rules below; do not guess through drift. |
| `mgs3mod recover --restore-missing <relative-target>` | Explicitly authorize recreating a missing target named by a valid unresolved apply intent. Repeat the flag to name each missing target; all must pass the restricted checks in section 5. |

Mutation commands support `--dry-run`: report actions, conflicts, and required bytes without initializing state, staging files, or changing the game. A live operation repeats validation under the lock. Commands support human-readable output and `--json`; results go to stdout and diagnostics to stderr. `--help` never writes state.

Use stable exit categories: `0` success/no-op, `1` I/O or internal failure, `2` usage/package validation, `3` wrong installation or running game, `4` ownership conflict or external drift, `5` recovery required/corrupt state, `6` manager busy. JSON errors contain an error code, explanation, and affected paths.

There is no blanket `--force` flag. Explicit enable/disable/remove/restore invocations are sufficient authorization for ordinary operations; they do not trigger repeated confirmation prompts. `remove` reports which stored files it will delete. Unexpected drift always needs investigation rather than an automatic overwrite.

The recovery help must show the restricted missing-file syntax, explain that missing files could have been deleted externally, and require the explicit flag for each affected path. Plain `recover` never recreates them. Preflight rejects any unlisted missing path before making changes. Human/JSON results list restored paths and `missingFileRecoveryExplicit: true`; the journal records the same choice. Exit `0` means recovery fully resolved; use `2` for invalid flag/path syntax, `3` for a mismatched core profile, `4` for an existing or unexpected target, and `5` for absent/invalid operation evidence or incomplete recovery. Other I/O and locking failures retain the standard exit categories.

## 4. Package and ownership model

Use `manifest.json` plus `payload/` in a folder or ZIP. A package ID is lowercase ASCII matching `[a-z0-9][a-z0-9-]{0,63}`. Only one version per ID is stored; importing identical content is a no-op, and importing different content under the same ID fails. Updating requires disable/remove/add until a dedicated update transaction is designed.

Example schema (placeholders are not an installable package):

```json
{
  "schemaVersion": 1,
  "id": "camo-test",
  "version": "0.1.0",
  "name": "MGS3 Camo Test",
  "profile": "mgs3-mcv-local-0d585dcc6a67",
  "files": [
    {
      "source": "payload/camo.ctxr",
      "target": "hqtex/flatlist/_win/sna_def_olive.bmp.ctxr",
      "originalSha256": "1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d",
      "payloadSha256": "64-lowercase-hex-digits",
      "payloadBytes": 0
    }
  ]
}
```

An authoring manifest supplies the exact target and original hash. `pack` fills payload size/hash; it must not learn an original hash from an enabled mod. If a target has a baseline record, that original is authoritative. Otherwise the current, unowned target is the proposed local baseline. Original provenance remains local and unverified.

Reject unknown schema fields, duplicate JSON keys, invalid hash/ID/version values, empty packages, unmapped payloads, duplicate ZIP entries, case-colliding paths, and multiple mappings to the same target. Bound parsing and extraction: at most 1 MiB of manifest data, 256 mappings, 64 MiB per payload, and 512 MiB total uncompressed payload. Enforce limits on bytes actually read, not only ZIP headers. These are deliberate small-texture-release limits.

Require canonical `/` separators, printable ASCII path components, and exact approved prefixes. Reject absolute/UNC/device/drive-relative paths, `..`, empty or dot segments, backslashes, colons/ADS, wildcards, reserved Windows names, and trailing spaces or dots. Reject symlinks, junctions/reparse points, directories as payload files, hard-linked targets, and unexpected alternate streams. Canonicalize case for conflict keys while retaining the actual filesystem spelling for display.

All extraction and deployment uses `os.Root` handles plus explicit target policy checks. Inspect the fixed root, state root, and every managed parent for reparse attributes before opening them; reject junction roots as well as linked targets. Resolve the actual long filename and file identity and reject short-name/8.3 aliases or alternate spellings that bypass canonical ownership keys. Merely joining strings beneath the game directory is not sufficient. Never run scripts or programs from packages. Do not follow source-folder links during `pack` or `add`.

Each canonical target has at most one enabled owner. Detect the entire conflict set before any game writes. The baseline record contains path, original size/hash, capture time, and the fixed profile ID. The active record contains owner and expected deployed hash. An absent target is an error in v1, not permission to create a file.

## 5. Transaction and recovery design

Use one exclusive OS-backed byte-range lock (`LockFileEx`, exclusive and fail-immediately) for all state-changing commands and consistent state reads. Explicitly unlock and close the handle; the OS also releases it if the process dies, although release can be delayed briefly. The existence of the lock file is not evidence of a live process. Revalidate all inputs after acquiring it. Read-only and dry-run commands open an existing lock without creating one; before initialization they report absent/incomplete state without writing files. Competing initializers must serialize directory/lock creation and never adopt another process's incomplete state.

Commands that change game targets require the game and its launcher to be closed. Use a Toolhelp process snapshot, filter candidate executable names first, then query their full executable paths. An inaccessible matching candidate is an unknown state and blocks deployment; unrelated protected system processes do not. Also handle sharing violations as ordinary failures. This is a best-effort guard, not a claim that a process scan prevents a concurrent launch. The manager lock serializes manager processes only. Tests must demonstrate safe failure if a target becomes locked after preflight.

The source of truth is the validated sequence of transaction records, not `state.json`. Each transaction records its previous committed generation, desired next state, target before/after hashes, verified backup references, staged files, and operation type. Records use a schema version and content checksums. Validate paths and object IDs loaded from state/journals as strictly as package paths; checksums detect damage, not malicious authorship. Keep this format small and explicit; no general event-processing framework.

1. **Preflight:** validate core identity, state integrity, package bytes, every expected current target, conflicts, process status, permissions, and storage budget. Conservatively budget new baseline copies, staged replacements, before-state snapshots, and 64 MiB of metadata/headroom with checked integer arithmetic; available space is advisory, so later disk-full errors still need recovery. If any check fails, do not touch game files.
2. **Prepare:** persist a preparation inventory before copying, including target, intended baseline hash, transaction-specific temporary name, and whether a complete verified object already existed. Copy a new original to that temporary name, hash/sync/reopen it, then promote it to the content-addressed baseline name. Existence alone never validates a baseline object: a mismatched object blocks work and is never overwritten. A crash before promotion leaves only inventoried staging; a crash after promotion leaves a verifiable immutable object that must be retained. Stage complete replacement files on the same volume. Record before-state recovery data even for disable/restore operations, whose pre-operation bytes are modded files.
3. **Record intent:** persist and sync the immutable transaction plan and all required data. Publish a checksum-validated `READY` record only after preparation is complete. No game-file mutation is permitted before this point.
4. **Apply:** persist a per-target `APPLY_INTENT`, recheck the expected target hash immediately before replacement, perform root-relative replacement, then reopen and verify the resulting hash and record `APPLIED`. Do not edit destination files in place or truncate them before a replacement is ready. Keep all recovery copies until the transaction is resolved.
5. **Commit:** after every target verifies, write and sync a checksum-validated `COMMITTED` record referencing the plan and desired state. Only a complete valid record commits the new logical state. Refresh `state.json` afterward as a cache.
6. **Finish:** perform idempotent cleanup. A committed removal may delete its library payload only now. Cleanup failure is reported without pretending the committed game changes were rolled back.

If the process exits before `READY`, remove only recognized preparation artifacts; no target should have changed. A ready transaction without a valid commit rolls back to its recorded before-state. After every restored target verifies, publish a validated `ROLLED_BACK` terminal record so future commands do not repeatedly recover the same transaction. A valid committed transaction rebuilds the state cache and finishes cleanup, after checking expected results. A missing or truncated cache is rebuildable; ambiguous journal corruption is not silently discarded. Valid committed history is replayed in order and historical target hashes are not compared against files legitimately changed by later committed transactions; only the final expected state and the unresolved tail govern recovery.

Before interpreting any journal record, verify its schema, checksum, referenced plan, generation chain, and allowed state transition. Malformed, torn, duplicate, contradictory, or impossible records produce exit 5 and preserve evidence; never interpret an invalid `READY` as permission to roll back or as harmless abandoned staging. Only a validated `READY` without a validated terminal record enters automatic rollback.

Recovery compares actual bytes with recorded before/after hashes. Known before bytes are left alone; known after bytes can be restored. A missing target is ambiguous: an interrupted replacement and external deletion may look identical. Block by default. A narrowly scoped `recover --restore-missing <relative-target>` may explicitly recreate each listed missing target from its verified before snapshot when a valid unresolved transaction and its `APPLY_INTENT` name the target, the fixed core profile still matches, and all other recovery checks pass. Report this ambiguity and record the explicit choice. It is not a general drift override and cannot overwrite an existing file. Missing files outside an unresolved operation remain external drift. Recovery itself must be resumable after another crash, including when only some targets have rolled back. Never automatically continue unrelated mutations while a transaction is unresolved.

Apply the same metadata protocol to initialization, import, and removal. Initialization has checksum-validated `INIT_PREPARED` and `INIT_COMMITTED` records: the prepared record identifies the profile and exact manager-owned paths before installation state is published, and the committed record follows a synced, verified installation record. Before commit, ordinary commands report incomplete initialization. Recovery may finish only a valid prepared initialization with matching identity; malformed records, unexpected contents, or a directory created without a complete ownership record block and remain untouched. Simultaneous initialization must be covered by the concurrency test.

Imports stage their library payloads before becoming visible in committed state. Removal's plan contains an exact cleanup inventory of the package manifest and payload paths, sizes, hashes, and ID. There is no cross-package payload deduplication in v1: unreferenced means the final committed state and any unresolved recovery plan no longer need those particular library files. Removal commits state without the package before cleanup. For each inventoried file, missing means already cleaned; a matching regular file may be deleted; unexpected contents or link/identity changes are preserved and reported as pending cleanup. Remove directories only when empty. Record completed cleanup, resume it after each possible deletion boundary, and block reuse of the same ID until cleanup resolves. Retain immutable baseline objects and committed transaction records; do not add baseline garbage collection in v1. Orphan staging data may be cleaned only when its ownership and lack of references are proven.

`restore`, `disable`, and `recover` restore only recorded targets, never the three identity binaries. A core fingerprint mismatch blocks deployment and ordinary mutation. Read-only diagnosis still works. To keep recovery available after a game update, allow only an explicit `recover` or `restore --baseline` under mismatch when every affected target still exactly matches recorded before/after bytes and every needed backup verifies. If the update changed a managed target, refuse to overwrite it. Do not automatically adopt the new installation or purge the old baseline.

Recovery target: automatic rollback/finalization after a process crash when surviving files match recorded before/after states; explicit resolution is required for ambiguous missing files or invalid records. Do not claim atomic multi-file visibility, protection against malicious concurrent filesystem mutation, or unconditional recovery after power failure/media corruption. Detect invalid records and preserve evidence instead of fabricating a clean status.

## 6. First mod workflow

Create a conspicuous, cosmetic-only recolor of the Normal/Olive Drab camo. Preserve its pattern, resolution, alpha, and mip chain. The proposed ID is `camo-test`. Local `sp/slot/camoufla-normal/bp_assets.txt:11` associates this camo with logical `textures/flatlist/sna_def_olive.bmp.ctxr`; `sp/stage/v000a_0/bp_assets.txt:383-384` also references it. This establishes a candidate, not the active physical texture under the user's current settings.

Primary candidate: `hqtex/flatlist/_win/sna_def_olive.bmp.ctxr`, 22,370,144 bytes, original SHA-256 `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`.

Standard-resolution fallback: `textures/flatlist/_win/sna_def_olive.bmp.ctxr`, 1,398,560 bytes, original SHA-256 `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049`.

Use one physical target per trial; restore it before changing candidates. Other suffixed or hashed camo variants are separate resources and must not be bulk-replaced. The final package includes only the target demonstrated active in game.

1. Recheck the candidate hashes above, inspect their headers and decoded content, and map the displayed camo name. Distinguish body texture from the menu thumbnail and potential alternate resource copies. Confirm dimensions and mip counts from both the container and decoded DDS rather than assuming them from file size.
2. Copy the source into ignored authoring work files. Pin a CtxrTool source revision or release and record the converter executable's SHA-256, provenance, and license. Do not auto-download or execute dependencies from mod manifests.
3. Export a copy to DDS while retaining the `.param` sidecar. Repack it unchanged. Compare dimensions, channel layout, alpha, mip count, and decoded pixel data. Use the documented 8.8.8.8 ARGB DDS representation. Avoid the BMP/TGA route because extraction does not preserve mip data. Byte-identical output is desirable but not mandatory if harmless container differences are understood; unexplained structural/pixel differences stop this route.
4. Package that unchanged conversion as a temporary local test. Through the tested manager, verify installation and restoration hashes; launch manually to check for rendering regressions. This smoke test alone cannot establish that the target is active; the visible recolor in step 7 must do that. Close the game before disabling. If the round trip fails, stop and resolve tooling before creating artwork.
5. Make an intentionally obvious magenta tint across the existing mip levels, preserving alpha and texture structure. The proposed reproducible byte-channel transform is `R' = floor((R + 255) / 2)`, `G' = floor(G / 3)`, `B' = floor((B + 255) / 2)`, and `A' = A`; use widened arithmetic and the validated DDS channel masks rather than assuming byte order. Apply it only to decoded pixel channels, never headers or padding. Use an established DDS editor or a small, independently validated authoring helper, with fixtures proving dimensions, mip layout, and alpha are unchanged. Do not build a general texture editor into the manager.
6. Repack with the original sidecar, decode and inspect again, then `pack` the final mod. Its manifest must contain the original target hash and final payload hash. Include a concise local recipe describing the changed camo, tool versions, and test conditions.
7. Enable, launch manually, select the intended camo, and capture before/after visual evidence at the same location and lighting. Inspect body, menu preview, close view, and distant/mip behavior. This proves which resource is active; filename inference alone does not.
8. Close the game, disable, verify the original SHA-256, and visually confirm the original appearance. Repeat using `restore --baseline`, then remove the package and verify the baseline backup remains intact.

If the resource mapping is wrong, restore immediately and investigate candidates one at a time. Do not overwrite a whole texture tree or alter graphics settings to hide uncertainty. A fallback to another early camo may be used after recording why the first target failed. A successful hash check without an in-game appearance check is not completion.

## 7. Implementation sequence

| Phase | Work | Exit gate |
|---|---|---|
| A. Skeleton and fixed profile | Create the Go module, command help/output, profile constants, read-only `doctor`, Windows adapter, and synthetic fixture helpers. | Tool builds; real `doctor` reads the correct root and fingerprints; wrong-root/build fixtures are rejected. |
| B. Package/library | Implement strict manifest parsing, folder/ZIP validation, `pack`, `add`, and `list`. | Round-trip synthetic package works; malformed paths, sizes, IDs, duplicate entries, and wrong profile fail without game writes. |
| C. Recovery core | Implement lock, immutable originals, staged writes, transaction records, cache reconstruction, and `recover`. | Crash/fault matrix below passes on synthetic files before any live texture replacement. |
| D. Lifecycle | Wire enable/disable/remove/restore/status/verify, conflicts, drift checks, JSON, dry run, and meaningful errors. | Complete lifecycle and multi-file rollback pass; live reads cannot accidentally become writes. |
| E. Texture proof | Resolve the exact candidate, pin tooling, prove unchanged conversion, author recolor, and create local package. | Package validation and offline texture comparisons pass; exact target identity recorded. |
| F. Live demonstration | Run the manual enable/appearance/disable/restore/remove sequence with game closed during mutations. | Appearance is demonstrated and restoration matches the original hash exactly. |
| G. Delivery | Build the final CLI, preserve recipe/package and original backup, document commands and recovery limitations. | One runnable CLI and one tested local mod; final status clean, mod disabled, backup retained. |

Phases C and D are prerequisites for E's live round-trip test. Offline texture investigation may proceed alongside A/B. Avoid parallel writers on the transaction/store implementation. Independent package validation and authoring research can be delegated with disjoint file ownership.

## 8. Required tests and verification

Use synthetic texture-shaped byte files for manager tests. Tests must never target the real game root. Windows integration tests run against isolated temporary fixture directories on the same filesystem. Production profile constants remain fixed; test injection is internal.

- Full lifecycle: add does not alter targets; enable applies exact bytes; disable and restore return exact originals; remove preserves backups; repeated operations are idempotent.
- Two non-overlapping mods coexist. Two overlapping mods, including case variants, fail before writes. A multi-file package fails as a whole if a later target is invalid.
- Tampered payload, original, active file, or backup is reported. A changed core binary blocks enabling. Conditional restore under core drift cannot overwrite an updated target.
- ZIP traversal, ADS, reserved names, Unicode/normalization ambiguity, duplicate JSON keys, links/reparse points, hard links, oversized extraction, unsupported schema, and package/target mismatches are rejected.
- Kill a child CLI process during baseline staging/promotion, every initialization boundary, every durable transaction boundary, each target replacement, and each library-file cleanup deletion. Restart recovery and assert exact before-state or fully committed after-state where unambiguous, correct ownership, and retained backups; ambiguous cases must block with preserved evidence.
- Inject partial writes, sync errors, disk-full errors, sharing violations, denied access, torn READY/plan/commit/cache records, duplicate records, invalid generations, and interrupted recovery. Test external deletion before apply separately from absence following apply intent, and exercise explicit missing-file recovery. No silent partial success or blind deletion of unknown files.
- Two concurrent CLI processes cannot mutate state together. Game/process-check uncertainty and target locks fail safely. No stale PID-file workaround is needed.
- Every dry-run leaves the complete fixture filesystem unchanged. JSON stdout parses even on failure; diagnostics stay on stderr. Exit codes match documented categories.
- Fuzz the manifest parser and archive path validator. Seed known Windows edge cases and bound allocation.

Run `go test ./...`, `go vet ./...`, and the Windows subprocess/fault integration suite. Run `go test -race ./...` if the installed Windows toolchain supports its C toolchain requirements; otherwise report that check as unverified and retain the explicit concurrency tests. Keep Go build/module caches under the source project's ignored `.cache` folder. Do not treat passing unit tests as proof of visual correctness.

Build from the source-project directory using the following PowerShell recipe after the module and tests exist. Dependency acquisition occurs during development, never at CLI runtime. Commit `go.mod` and `go.sum`, use `go mod verify`, and preserve the resulting binary's SHA-256 in the release notes. The output hash cannot be known before implementation.

```powershell
$env:GOTOOLCHAIN = 'local'
$env:GOWORK = 'off'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
$env:GOCACHE = Join-Path (Get-Location).Path '.cache/go-build'
$env:GOMODCACHE = Join-Path (Get-Location).Path '.cache/go-mod'
function Invoke-Mgs3Go {
    & 'C:\Program Files\Go\bin\go.exe' @args
    if ($LASTEXITCODE -ne 0) { throw 'Go command failed; stopping the build.' }
}
New-Item -ItemType Directory -Path ./dist -Force | Out-Null
Invoke-Mgs3Go mod download
Invoke-Mgs3Go mod verify
Invoke-Mgs3Go test ./...
Invoke-Mgs3Go vet ./...
Invoke-Mgs3Go build -mod=readonly -trimpath -o ./dist/mgs3mod.exe ./cmd/mgs3mod
./dist/mgs3mod.exe doctor --json
if ($LASTEXITCODE -ne 0) { throw 'Installation check failed; do not deploy.' }
```

Run the optional race check separately with the required C toolchain and `CGO_ENABLED=1`. Copy the final verified executable from `dist` to the fixed game root for convenient use only after checks pass. Verify its output hash and all three core fingerprints again through `doctor`; don't automatically initialize or enable a mod as a build side effect.

## 9. Definition of done and completion record

Done means: a buildable source project, documented commands, a runnable Windows CLI, a validated local package, passing recovery tests, observed in-game texture change, and byte-exact restoration with no managed mod active at delivery. The baseline and reusable package remain available.

The implementation later closed the planned active-texture, unchanged-round-trip, converter-build, journal fault-injection, and visible-recolor gates. The user confirmed the HQ texture produced a magenta uniform and rendered normally. Matched screenshots, separately enumerated body/menu/close/distant observations, and a post-restoration visual launch were not captured, so those narrower visual evidence items remain incomplete. Byte-exact post-trial restoration was independently verified.

Planning estimate: roughly 1–2 weeks of focused engineering for a reliable first release, with the texture/converter gate the largest scope uncertainty. This is a planning range, not a delivery promise.

Research rationale and links: [research.md](research.md). Exact local observations: [installation-evidence.md](installation-evidence.md).
