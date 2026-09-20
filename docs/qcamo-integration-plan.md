# QCamo integration plan

Historical first ASI integration scope. The external loader decision below is superseded by the [managed standalone-loader plan](managed-asi-loader-plan.md).

Date: 2026-09-20. This document records the implementation scope, not evidence that the work has passed.

## Goal

Deliver the first non-test mod package, QCamo 1.0.4, in a documented, repeatable format that other users can manage with MGS3 Mod Manager. Preserve the existing texture package and journal compatibility. The earlier `camo-test` remains a test fixture.

## Decisions

- Start with the official upstream binary, pinned to source commit `0a8bee4874e6ed9773c96bd48e899056c4c67801`. Label the result as repackaged, not rebuilt or independently authored. Preserve the upstream MIT notice.
- Keep authored release recipes under `mods/qcamo/`; generated packages and distribution files belong in ignored `dist/` and `work/`. Do not distribute game files, saves, manager state, or original texture backups.
- Add an explicit runtime-plugin package schema. Existing texture packages retain their old interpretation. Native plugins are executable code and remain opt-in through enable.
- Permit safe root-level ASI targets only for the runtime schema. Record their baseline as absent. Reject adoption or overwriting of pre-existing unowned plugins, even when their bytes happen to match.
- Require a recognized external ASI loader at enable time. Verify known file hashes, not merely a DLL filename. The manager does not download or delete the shared loader.
- Preserve runtime-generated logs and crash dumps. They are not package payloads and automatic deletion could remove diagnostic evidence.
- Add `--game-root` so another user can select a local installation. Keep the supported game's compiled fingerprints; changing location does not bypass compatibility checks. Existing journals remain bound to the root where they were initialized.
- Updates use the existing explicit remove/add/enable sequence. An atomic update command and automatic shared-loader installation are outside this first release.

## Execution sequence

1. Trace package, state, transaction, filesystem, and recovery assumptions; define absent-file semantics and backward compatibility.
2. Implement schema and payload validation, portable root selection, and loader identity checks.
3. Implement managed file creation/removal and recovery. Verify drift, conflict, process guard, and interruption handling.
4. Create a deterministic QCamo packaging recipe, preserve license/provenance, and produce a shareable distribution with instructions and checksums.
5. Run regression tests, vet, package round trips, lifecycle/fault tests, and read-only checks against the real installation. Review changes independently and fix findings.
6. Perform only supported live checks. If the loader is absent, report the exact prerequisite and retain the package disabled. In-game behavior requires a real gameplay trial and must not be reported as tested without one.

## Acceptance

- Original schema-1 packages and existing generation-9 history still load.
- A selected compatible installation can be initialized outside the original hardcoded path.
- Plugin packages reject unsafe targets, invalid binaries, and missing/unknown loader identities.
- Add stores a disabled package; enable creates its previously absent ASI; disable/remove restores absence without deleting changed or unowned files.
- Interrupted creation and removal recover to the recorded state, with external changes blocking recovery.
- Packaging is repeatable, contains no game data, and includes the complete upstream license in the outer distribution.
- Published instructions distinguish automated filesystem tests from unperformed gameplay tests and identify the exact supported game and loader versions.

## Deferred source rebuild

The upstream Nix build and lockfile provide the reference rebuild route. Nix/MinGW/CMake are not currently available in PATH. A later source rebuild must record the toolchain and dependency revisions, compare its binary to the official artifact, investigate differences, and rerun gameplay tests before replacing the packaged binary. Fixing QCamo's open icon/input issues is a separate versioned fork, not an undocumented change to upstream 1.0.4.
