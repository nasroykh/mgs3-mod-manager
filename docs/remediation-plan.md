# MGS3 mod manager remediation plan

Date: 2026-09-20.

Status: completed on 2026-09-20. Final code, artifact, installation, and live read-only evidence is recorded in [implementation-review.md](implementation-review.md). Race-enabled testing, induced physical power/media/full-volume failure, and the omitted detailed visual observations remain explicitly unverified.

## Goal

Resolve every defect and plan/documentation gap identified by the 2026-09-19 final review without mutating a live game target. Preserve compatibility with the existing generation-9 manager history, stored package, and baseline. Final acceptance requires a clean build, a fresh full test run, targeted regression tests for every finding, clean current `doctor`/`status`/`verify` output, and unchanged live target/baseline hashes.

## Constraints and invariants

- The production root, profile ID, core hashes, package format, and supported target namespace remain fixed.
- Existing schema-1 transaction records must continue to decode and replay exactly. New persisted fields require a new plan schema or `omitempty`-compatible decoding so old canonical checksums remain valid.
- No test may write to the real game target or real `.mgs3mod` state. Synthetic roots remain mandatory.
- Only a successfully synchronized and promoted final commit marker may advance logical state.
- Invalid authoritative journal records remain preserved. Exact transaction-owned temporary files may be deleted and recreated only when a valid plan proves ownership.
- Existing target bytes and Windows security metadata must survive enable/disable/restore cycles.

## Phase 1: durable transaction and recovery fixes

1. Make rollback staging resumable. Treat `rollback-%04d` as an exact plan-owned temporary file. If it exists and verifies, reuse or replace it normally. If it is partial or has the wrong hash, validate that it is a safe regular transaction-owned file, remove it, and rebuild it from the verified before-state blob. Add fault tests for partial write, short write, failed sync, and process interruption while creating rollback staging. A second `recover` must finish successfully and leave no partial stage.
2. Replace direct final `COMMITTED` creation with a staged commit protocol. Write and synchronize an exact transaction-owned temporary commit marker first, close it, then promote it to `COMMITTED`. Loader/recovery must recognize an incomplete commit staging file as non-authoritative owned staging, never as a commit. A sync failure must return `Committed:false`, keep recovery required, and roll back safely. A crash after final promotion must finalize the committed state. Add crash and I/O tests on both sides of promotion.
3. Extend new preparation plans with explicit target, temporary name, intended baseline object, and verified-object-existed state. Preserve schema-1 replay for current history and emit a validated newer schema for new transactions. Add compatibility tests that load real-shape schema-1 records and round-trip their canonical plan hashes.

## Phase 2: Windows replacement and locking fixes

1. Add a Windows replacement primitive for existing files that preserves the replaced target's DACL/security metadata. Prefer `ReplaceFileW` or an equivalent handle-based implementation and fail safely if preservation cannot be guaranteed. Continue using root-relative rename only for destinations that are deliberately absent, such as immutable baseline promotion, library import, and explicit missing-file recovery.
2. Add integration tests with a protected explicit DACL. Enable, disable, restore, and rollback must preserve the target security descriptor/protection bit and exact bytes. Sharing denial and read-only behavior must remain safe.
3. Remove `FILE_SHARE_DELETE` from the held manager lock handle. Add a test proving rename/delete of the held lock fails and another manager still receives exit category 6. Existing dead-owner release behavior must continue to pass.

## Phase 3: command semantics and reporting fixes

1. During conditional restore under a core mismatch, omit physical changes whose observed hash already equals the baseline. Commit only the logical disabled state. Add read-only and exclusively-open already-baseline cases; both must succeed without replacing the target.
2. Set `recovery resolved` and dry-run completion messages only after successful recovery. Return only paths actually restored; distinguish planned paths if the API needs preflight reporting. JSON failure output must keep `ok:false`, the correct exit category, and no false success message/path.
3. Run the game/launcher process guard during rollback only when a game target will be replaced or recreated. Interrupted library-only `add` recovery must work while a synthetic process guard reports the game running.

## Phase 4: package, build, and documentation fixes

1. Use the ZIP minimum timestamp `1980-01-01T00:00:00Z` for normalized entries. Add a test that opens a produced package, checks both entry timestamps, and proves byte-for-byte deterministic output across two packs.
2. Add `go mod download` before `go mod verify` in `scripts/build.ps1`. Test the script with project-local caches and confirm its output hash remains recorded.
3. Update stale current-tense claims in `installation-evidence.md`, `mod-validation.md`, and `implementation-plan.md`. Keep historical phase evidence explicit. Update implementation status/review with fixed findings and current verification results.
4. Record the remaining visual evidence limit accurately: active HQ target is user-confirmed, while matched screenshots and post-restoration visual confirmation were not captured. Do not claim those checks occurred.

## Phase 5: verification and release gate

Run, in order:

```powershell
go mod download
go mod verify
go test -count=1 ./...
go vet ./...
gofmt -d <all Go files>
go test ./internal/packagefmt -run '^$' -fuzz '^FuzzParse$' -fuzztime=5s
go test ./internal/packagefmt -run '^$' -fuzz '^FuzzPackagePath$' -fuzztime=5s
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\build.ps1
```

Then compare the audit build, `dist/mgs3mod.exe`, and installed executable; run read-only `doctor --json`, `status --json`, and `verify --json`; verify generation 9 remains disabled with no pending recovery; and recompute full SHA-256 values for the live HQ target and retained baseline. Race-enabled tests remain `UNVERIFIABLE` unless a supported C compiler is installed. Power-loss/media-corruption behavior remains outside the stated guarantee.

## Completion criteria

- Every runtime finding has a focused regression that fails on the old behavior and passes on the fix.
- Existing schema-1 live history loads without migration or rewriting.
- Full fresh suite, vet, formatting, fuzz smoke, build script, and read-only live checks pass.
- No game target, package enablement state, baseline, or live journal is changed by remediation.
- Documentation contains no stale “still pending” claims for completed converter, journal, or active-target work and no unsupported visual claim.
