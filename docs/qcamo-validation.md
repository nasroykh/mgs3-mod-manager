# QCamo integration validation

Historical generation-10 checkpoint. The external MGSHDFix-loader gate and artifact hashes below describe the prior release. See [standalone-loader validation](asi-loader-validation.md) for the manager-owned loader extension and current state.

Date: 2026-09-20. Scope: manager schema-2 support, absent-file lifecycle and recovery, portable installation selection, and the unchanged official QCamo 1.0.4 distribution. This record does not certify QCamo gameplay stability or independent source-to-binary reproducibility.

## Completed checks

- Package/profile tests cover schema-1 compatibility, explicit absent origins, root ASI path restrictions, folder/ZIP round trips, and PE32+ AMD64 DLL checks including legitimate uninitialized `.bss` sections.
- Manager regression tests cover creation/deletion, pre-existing unowned files, loader gating, conditional restoration, and interrupted enable/disable at eight boundaries each. They also cover a destination appearing during creation and content changing during the process check before deletion.
- Windows `MoveNew` tests establish that file promotion refuses to overwrite an existing destination. This does not claim immunity to all malicious concurrent filesystem changes.
- CLI tests cover selected-root initialization and preserved fingerprint checks, including malformed root options.
- The actual official QCamo payload completed add, enable, verify, disable, re-enable, remove, and verify in a synthetic installation. `TestQCamoReleaseArtifactLifecycle` uses `MGS3MOD_QCAMO_PACKAGE`; it never executes the DLL.
- The preview manager loaded the existing live generation-9 schema-1 history with `doctor` and `verify`, and successfully preflighted the real QCamo package with `add --dry-run`.

The implementation suite passed with `go test -count=1 ./...` and `go vet ./...`. After the independent review fixes, the final build script passed tests, vet, build, and live `doctor`; the manager suite took 60.973 seconds. `MGS3MOD_QCAMO_PACKAGE` pointed to the generated actual QCamo package during the build, so its opt-in lifecycle test was included. The retained final build output is `work/qcamo-verification/final-build.log`.

## Final artifacts

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `dist/mgs3mod.exe` and installed `../mgs3mod.exe` | 6,438,400 | `ad3db9442c2f183bc261baff97f874f6dbbc639df0c22623dc72dd346522bcdc` |
| `dist/qcamo-1.0.4/qcamo-1.0.4.mgs3mod.zip` | 4,064,453 | `d79d7c91d46c2112f3e45af89b040ce65c40df0b7fa92e9158cc18f1d951a195` |
| `dist/qcamo-1.0.4/qcamo-1.0.4-manager-distribution.zip` | 4,914,347 | `3812b5d6bd86d103822c4e79dbc071c798dff8889eb38c69fb1fb2d805dde927` |

Share the outer distribution. It contains six entries: the manager executable, inner package, README, complete QCamo MIT license, provenance, and checksums. Every one of the five checksummed members was verified, and the bundled executable matches the installed manager. Two independent packaging runs with identical inputs produced byte-identical inner and outer archives. Commands and hashes are retained in `final-package.log` and `final-repeat.log` under `work/qcamo-verification/`. Existing-output rejection was also checked without changing the prior archive; that expected failure is in `no-overwrite.log`.

The pre-review distribution was moved to `work/qcamo-before-review/` and is superseded. Do not share that earlier artifact.

## Live state and verdicts

- Manager deployment and disabled import: CORRECT. `../mgs3mod.exe add dist/qcamo-1.0.4/qcamo-1.0.4.mgs3mod.zip --json` completed at generation 10. `live-add.json` records `qcamo` 1.0.4 stored disabled alongside disabled `camo-test`.
- Loader prerequisite: CORRECT. `../mgs3mod.exe enable qcamo --dry-run --json` returned code 3 because `winhttp.dll` is absent; `live-enable-gate.json` records the exact failure. No live `qcamo.asi` exists. No loader was installed.
- Final managed integrity: CORRECT. `../mgs3mod.exe verify --json` and `status --json` both exited 0 with no recovery pending. Receipts are `live-verify.json` and `live-status.json`. The HQ texture still matches its original SHA-256 `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`.
- Gameplay readiness: UNVERIFIABLE until the external loader is installed and a gameplay trial is performed. The automated manager integration is complete; runtime validation is not.

The previous texture-only manager executable is retained at `work/qcamo-verification/mgs3mod-texture-only.exe` for comparison. It is not a supported downgrade: schema-2 package history requires the new manager, even after package removal. No commit or public release was made; source changes remain in the working tree on `main`, based on `a83ed15f87c7116300cda4f8c75514f0adbdebda`.

## Limits

- Gameplay testing: UNVERIFIABLE. The game was not launched for this integration.
- Source rebuild/reproducibility: UNVERIFIABLE. The artifact is the hash-verified official binary, repackaged unchanged.
- Race-detector testing: UNVERIFIABLE for this release unless a separate successful run is recorded; the existing environment lacks the required C compiler.
- Other game versions and loader releases: unsupported. The manager keeps its full game fingerprint checks and pins both MGSHDFix 4.1.2 loader DLLs.
- Physical power loss, media corruption, and hostile concurrent changes remain outside the tested process-interruption model.

## Known failures resolved during integration

Independent review found an equal-byte ownership race: when a foreign file appeared before creation, recovery could remove it if its hash matched the intended payload. Recovery now treats a surviving apply stage as proof that promotion did not occur and preserves that destination. The journal also rejects an APPLIED marker with a surviving apply stage. `TestPluginExactPayloadRaceRecoveryPreservesUnownedFile` and the existing crash matrix pass. The reviewer rechecked the fix and found no remaining substantive code defect in the changed scope.

Review also found the PE entry check accepted a non-executable section. The validator now requires executable section characteristics and actual backing bytes for the entry point, with non-executable and virtual-tail regression cases.

The first real package check failed with `plugin payload section exceeds file bounds`: the validator incorrectly rejected a zero-length `.bss` raw section. The corrected check accepts that legitimate uninitialized section and still rejects initialized data beyond file bounds. A synthetic regression and real-package dry run passed afterward.

The initial preview build without explicit VCS settings failed with `error obtaining VCS status: exit status 128` because sandbox-owned Git metadata has a different owner. Release builds now use `-buildvcs=false`; the project does not change global Git trust settings.

Default sandbox execution and direct patch tools failed with `helper_unknown_error: setup refresh had errors`. Authorized escalated execution and the direct patch utility were used instead.
