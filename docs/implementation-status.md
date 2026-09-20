# Implementation status

Historical texture-release checkpoint. The statements and artifact hashes below describe the pre-QCamo release. For the current manager-owned loader, user-reported in-game success, and generation-13 state, see [standalone-loader validation](asi-loader-validation.md). The intermediate external-loader checkpoint is in [QCamo validation](qcamo-validation.md).

Updated 2026-09-20. The CLI and local HQ test package are implemented. The 2026-09-19 review findings have been remediated and the complete release gate passes. The user confirmed the magenta uniform rendered normally in game and then closed the game. The manager restored the exact original texture; live state remains generation 9, with `camo-test` stored but disabled and no recovery pending.

## Delivered artifacts

- Installed executable: `C:\Games\METAL GEAR SOLID 3 - MCV\mgs3mod.exe`.
- Matching build output: `dist/mgs3mod.exe`, 6,294,016 bytes, SHA-256 `f4c9ce1365e5d28defd364c9dfe2801d578bc6e3b31a803998db5d22755e7b1f`.
- Local package: `dist/camo-test-hq-0.1.0.zip`, 22,370,859 bytes, SHA-256 `2edaa130aefcdfba80631325d3dedbe500bceef28501c768a0817ab8f41492fa`.
- Source, tests, build script, strict manifest, reproducible recolor helper, and pinned converter repair patch.
- Manager state and immutable original backup: `.mgs3mod/` under the game root.

The automated lifecycle and visible-activation check are complete. The HQ target and retained baseline both have original SHA-256 `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`; the fallback target was not modded. The user's short report did not separately document the menu, close, and distant views, no screenshot was captured, and no post-restoration visual launch was performed. The manager does not edit game configuration, saves, identity binaries, or launcher files. Its additional installed executable is `mgs3mod.exe`.

## Validation

| Claim | Verdict | Evidence |
|---|---|---|
| Pinned dependencies, complete test suite, vet, build, and real installation diagnosis pass | CORRECT | `powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\build.ps1` exited 0. It ran `go mod download`, `go mod verify`, `go test ./...`, `go vet ./...`, the release build, and `doctor --json`. A separately retained `go test -count=1 ./...` run passed with the manager package at 51.672s. Raw output and exit codes are under `work/final-verification-20260920/`. |
| Process crashes and durable I/O errors follow the documented recovery rules | CORRECT | The commit marker is synchronized under `COMMITTED.stage` before promotion; failed writes, syncs, and closes remain uncommitted. Rollback staging is recreated safely after partial writes, short writes, sync failures, and process interruption. Subprocess crash matrices still cover init, import, enable, disable, restore, remove, recovery, baseline promotion, and cleanup. |
| Windows path, metadata, access, locking, and concurrency checks work on this host | CORRECT | Actual junction, hard-link, ADS, 8.3 alias, sharing denial, read-only attribute, DACL denial, two-process lock, dead-owner release, held-lock rename/delete denial, and full target security-descriptor preservation through enable, disable, restore, interrupted apply, and rollback passed. |
| Existing generation-9 schema-1 history remains compatible | CORRECT | The remediation build's read-only `doctor --json`, `status --json`, and `verify --json` all loaded the live history successfully without migration or writes. New operations emit schema 2 preparation inventories; regressions preserve canonical schema-1 encoding. |
| Recovery failure output reflects the durable terminal state | CORRECT | Recovery reloads journal history after every failure. A regression forces cache `Sync` failure after durable `ROLLED_BACK` and proves exit 1, `recoveryRequired: false`, the rolled-back logical state, exact original target bytes, and a successful later `verify`. |
| Package output is deterministic with valid normalized timestamps | CORRECT | Unit tests compare two independent pack outputs byte-for-byte and inspect every entry. The final package was independently repacked to the same SHA-256; both entries show the ZIP minimum timestamp `1980-01-01 00:00:00`. |
| Real installation enable/disable/restore/remove preserves exact originals | CORRECT | All first 18 lifecycle receipts in `work/live-test-20260919/` report success. Independent full-file .NET SHA-256 matched the original for the live target and retained backup. The completed lifecycle ended at generation 7 with one disabled package and no recovery pending. Receipt 19 subsequently enabled the visual trial at the user's request. |
| Converter and recolor preserve required texture structure | CORRECT | Both unchanged candidates repack byte-identically. The independent recolor verifier checks all 5,592,407 HQ pixels across 13 mips, exact RGB math, unchanged alpha/header, and lossless repack/re-export. [Texture evidence](mod-validation.md). |
| Race-enabled tests pass | UNVERIFIABLE | `CGO_ENABLED=1 go test -race ./...` fails before tests because `gcc` is not on PATH. Native process/concurrency tests passed separately. No compiler was installed. |
| Literal full-volume exhaustion or power failure is proven safe | UNVERIFIABLE | Synthetic disk-full/write/sync errors are covered; the real volume was not filled and power was not interrupted. The design does not promise unconditional recovery from media loss or damaged journals. |
| The user reported successful HQ recolor rendering | CORRECT | The user's response was: "Magenta uniform; rendering looks normal; game closed." This is user-reported visual evidence, not an independently captured screenshot. |
| The texture was restored after the visual trial | CORRECT | `disable camo-test --json` and `verify --json` exited 0; receipts 20 and 21 record generation 9 with `enabled: false`. Independent full-file .NET SHA-256 matched the original for both live target and retained backup. A separate post-restoration game launch was not performed. |

The retained final fuzz runs completed 373,757 manifest executions and 493,565 path executions without failure. Fuzzing is bounded testing, not proof that every input is safe.

## Trial commands

From the game directory, with the game and launcher closed:

```powershell
.\mgs3mod.exe enable camo-test
```

Launch MGS3 normally, select Olive Drab/Normal camouflage, and inspect the body and menu preview at close and distant views. After closing the game:

```powershell
.\mgs3mod.exe disable camo-test
.\mgs3mod.exe verify
```

If the appearance does not change, leave this candidate disabled and investigate the standard-resolution texture next. Do not enable both candidates together. A factory-settings reset is intentionally outside this texture-only release; `restore --baseline` restores only captured local managed files.

## Scope and implementation decisions

The manager is fixed to the recorded root and three complete fingerprints. It uses Go 1.27.1, the standard library, and pinned `golang.org/x/sys v0.48.0`. Production exposes no root override, environment bypass, test fault hook, mod scripts, or automatic downloads. Source lives entirely under `mgs3-mod-manager`; game assets and generated tools remain ignored in `work`, `dist`, and `.cache`. No Git repository was created or published.

An unchanged-input rendering smoke test was superseded by the stronger proof that the converter outputs every original byte exactly. This does not waive the actual recolor visual trial. Pre-ready cleanup removes only explicitly owned temporary names; promoted baselines and resolved journal recovery data remain retained. Records are bounded to 8 MiB, checked before journal creation.

The first direct build-script invocation was blocked by Windows' script execution policy. The successful invocation used a process-scoped flag and changed no persistent execution policy. Earlier test failures and the complete 2026-09-19 review findings were fixed with targeted regressions. See [remediation-plan.md](remediation-plan.md) and [implementation-review.md](implementation-review.md) for the final evidence.
