# MGS3 mod manager remediation review

Reviewed on 2026-09-20 against [implementation-plan.md](implementation-plan.md) and [remediation-plan.md](remediation-plan.md). This replaces the stale pre-remediation status report. The independent 2026-09-19 defect report remains preserved at `work/audit-20260919/final-review.md` as historical evidence.

## Scope and method

The declared review scope covered every production Go file under `cmd/` and `internal/`, all related tests, `scripts/build.ps1`, current documentation, both delivered artifacts, and the live manager state through read-only commands. No test wrote to the real game target or live `.mgs3mod` history. Synthetic manager roots were used for every mutation and fault injection. The exact reviewer sequence is a process record and is **UNVERIFIABLE** from retained raw logs; the final source, artifacts, live read-only outputs, and fresh verification results below are independently inspectable.

The independent review record reports that a reviewer with no prior report context found one additional recovery-reporting boundary defect plus stale status/review documentation. The resulting source is directly verifiable: recovery now reloads durable history after an error, and the exact terminal-marker boundary has a regression test.

## Finding disposition

| Claim | Verdict | Current evidence |
|---|---|---|
| Rollback remains resumable after a partial or failed rollback-stage write | **CORRECT** | `internal/manager/apply.go:456-464` validates and removes the exact transaction-owned regular stage before rebuilding it. `internal/manager/io_failure_test.go:112-153` covers seven-byte disk-full output, nil-error short write, failed `Sync`, and interruption after staging; a second recovery completes. |
| A failed `COMMITTED` write or synchronization cannot commit logical state | **CORRECT** | `internal/manager/apply.go:165-193` writes and synchronizes `COMMITTED.stage`, closes it through the durable writer boundary, then promotes it to `COMMITTED`. `internal/manager/journal.go:322-374` treats the staging name as owned but non-authoritative. The I/O matrix and `TestCommitPromotionCrashBoundary` prove rollback before promotion and finalization after promotion. |
| New preparation inventories contain the target, intended object/hash, exact temporary name, and verified prior-existence state without breaking schema 1 | **CORRECT** | `internal/manager/types.go:120-127`, `manager.go:539-549`, and `journal.go:85-190` emit and validate schema 2 while the new fields use `omitempty`. `TestPreparationSchemaTwoAndSchemaOneCompatibility` proves canonical schema-1 bytes omit them. The current binary also loaded all nine live schema-1 transactions read-only without migration. |
| Multiple targets sharing one original content hash prepare safely | **CORRECT** | `internal/manager/apply.go:63-99` reuses a baseline object only when the same executing plan already promoted and reverified it. `TestSchemaTwoReusesSamePlanBaselineObject` exercises the two-target lifecycle. |
| Conditional restore performs no physical replacement when the affected target is already at baseline | **CORRECT** | `internal/manager/manager.go:183-204` removes those changes before process, access, space, and replacement work. Read-only and exclusive-sharing regressions at `review_regression_test.go:65` and `windows_test.go:73` pass as logical-only commits. |
| Recovery results reflect the durable journal state and never report premature success or paths | **CORRECT** | `internal/manager/manager.go:98-129` reloads history on error, reports an unresolved tail only when one remains, and assigns completion paths only after nil error. `TestFailedRecoveryHasNoSuccessMessageOrPrematurePaths` covers failed, dry-run, and completed results. `io_failure_test.go:176-205` forces cache synchronization failure after durable `ROLLED_BACK` and proves exit category 1, resolved recovery state, correct target bytes, and successful later verification. |
| Interrupted library-only import recovery is not blocked by the game process guard | **CORRECT** | `internal/manager/apply.go:400-421` calls the guard only if rollback must mutate a game target. `TestInterruptedLibraryAddRecoveryDoesNotCheckProcesses` passes with a guard that reports the game running. |
| Existing target replacement preserves Windows security metadata | **CORRECT** | `internal/manager/filesystem.go:170-180` routes existing destinations through `winfs.ReplaceExisting`; `internal/winfs/winfs_windows.go:146-187` uses `ReplaceFileW` without a metadata-losing fallback. Direct and full manager lifecycle tests compare the complete owner/group/DACL SDDL through enable, disable, restore, interrupted apply, and rollback. |
| A held manager lock cannot be renamed or deleted | **CORRECT** | The lock handle at `internal/winfs/winfs_windows.go:110-125` omits `FILE_SHARE_DELETE`. `TestHeldLockPathCannotBeRenamedOrDeleted`, two-process contention, release, and dead-owner tests pass. |
| Package ZIP metadata is valid and deterministic | **CORRECT** | `internal/packagefmt/packagefmt.go:274-278` uses the ZIP minimum date. `packagefmt_test.go:88-113` compares two outputs byte-for-byte and checks both timestamps. The delivered package was independently repacked to the same SHA-256, and both entries show `1980-01-01 00:00:00`. |
| The build script obtains pinned modules before verifying them | **CORRECT** | `scripts/build.ps1:20-21` runs checked `go mod download` before `go mod verify`. The complete script exited 0; retained output is `work/final-verification-20260920/build.txt`. |
| Current documentation reflects completed converter, journal, and active-HQ work | **CORRECT** | `installation-evidence.md`, `mod-validation.md`, `implementation-plan.md`, and `implementation-status.md` now distinguish completed evidence from the missing detailed visual observations. Current artifact sizes and hashes are recorded. |
| Detailed body/menu/close/distant screenshots and post-restoration appearance are proven | **UNVERIFIABLE** | The user reported "Magenta uniform; rendering looks normal; game closed," proving visible HQ activation and normal rendering. No screenshot, separately itemized view report, or post-restoration game launch exists. Exact file restoration is proven separately by full-file hashes. |
| Race-enabled tests pass | **UNVERIFIABLE** | The proving command `CGO_ENABLED=1 go test -race ./...` fails before compilation because `gcc` is absent from `PATH`. Native process, sharing, and lock tests pass. |
| Literal power loss, media corruption, or a physically full volume is recoverable | **UNVERIFIABLE** | Actual partial files and injected disk-full, short-write, synchronization, close, and access errors are covered. The real volume was not filled and power/media faults were not induced; the documented guarantee excludes unconditional recovery from these events. |

## Verification evidence

Fresh raw output and separate exit-code files are retained under `work/final-verification-20260920/`. They produced these results:

```text
go mod download                                      exit 0
go mod verify                                        all modules verified
go test -count=1 ./...                               pass; manager 51.672s
go vet ./...                                         pass
gofmt -d <all Go files>                              no differences
FuzzParse, 5 seconds                                 373,757 executions; pass
FuzzPackagePath, 5 seconds                           493,565 executions; pass
scripts/build.ps1                                    exit 0; tests, vet, build, and doctor pass
```

The build script produced `dist/mgs3mod.exe`, 6,294,016 bytes, SHA-256 `f4c9ce1365e5d28defd364c9dfe2801d578bc6e3b31a803998db5d22755e7b1f`. The installed root executable is byte-identical. The deterministically repacked `dist/camo-test-hq-0.1.0.zip` is 22,370,859 bytes with SHA-256 `2edaa130aefcdfba80631325d3dedbe500bceef28501c768a0817ab8f41492fa`; its payload hash remains `384dc7966fecf246162ab8712b065c5159230c151e62af852fc47c11acb3bb3d`.

Installed `doctor --json`, `status --json`, and `verify --json` each exited 0. Their retained JSON reports compatible generation 9, `camo-test` disabled, and no unresolved recovery. The live HQ target and retained baseline are each 22,370,144 bytes with SHA-256 `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`; recomputed records are in `artifact-hashes.json` in the same verification directory.

The final documentation freshness sweep enumerated all 12 authored Markdown files in `README.md`, `docs/`, and `examples/`. All 31 relative links resolve, all 16 external links opened, every code fence is balanced, no trailing whitespace or mojibake marker remains, and current artifact/hash claims were recomputed from disk. `documentation-checks.json` retains the enumerated files and links. All 21 live-test receipts parse and report `ok: true`; `receipts-summary.json` records receipt 21 at generation 9 with the package disabled. Research-phase documents now distinguish historical limitations from later completed evidence. Ignored third-party source checkouts under `work/` are preserved as upstream artifacts and are not treated as project-authored documentation.

## Verdict

The remediated implementation and delivered artifacts are **CORRECT** for the declared release contract, subject to the three explicit **UNVERIFIABLE** limits above. No known concrete code, packaging, installation, or live-state defect remains from the audited scope.
