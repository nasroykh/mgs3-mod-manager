# Standalone ASI loader validation

Date: 2026-09-20. Scope: pinned standalone Ultimate ASI Loader package, manager ownership/dependency handling, combined QCamo distribution, and local installation. Automated and filesystem validation is complete. The user subsequently reported successful in-game operation; detailed scenario coverage is not claimed.

## Provenance

The official Ultimate ASI Loader v9.7.4 NoPDB x64 ZIP hashes to `e5860e7d9a1805267535b65749575b5e406cc6ea3325c7392189c578815045d1`. Its only member, `dinput8.dll`, is 1,198,304 bytes with SHA-256 `031a3e5576d91dce1e438d36b9a3d462c7334ab4791990a8ff1e3ddc0e132daf`. It is unchanged and renamed to the supported `wininet.dll` alias because the local game's Engine.dll imports WININET. Official license/provenance are retained under `mods/asi-loader/`.

## Acceptance checks

- Schema 3 accepts only the pinned loader identity, exact target set, absent origin, size, and hash. Schema-1 and schema-2 data keep their prior encoding.
- Loader enable does not depend on itself; ASI enable requires enabled managed loader ownership and correct bytes.
- Loader disable/removal is blocked by enabled managed ASIs. Removing QCamo leaves the loader installed. Restore handles both in one transaction.
- Existing unowned loader DLLs and configuration overrides are refused, not overwritten or adopted.
- The real loader and QCamo payloads are tested in a synthetic installation without executing either DLL.
- Final distribution must reproduce byte-for-byte and preserve both upstream licenses. No MGSHDFix content is included.

## Results and evidence

Evidence paths below are relative to the repository. Generated logs and releases are intentionally ignored by Git.

- Regression and real-artifact lifecycle checks: CORRECT. `go test -count=1 ./...` with `MGS3MOD_ASI_LOADER_PACKAGE` and `MGS3MOD_QCAMO_PACKAGE` passed; manager tests took 108.477 seconds. `go vet ./...` passed. Evidence: `work/asi-loader-verification-manager-tests.log`.
- Post-gameplay precommit checks: CORRECT. `go test ./...` with both final inner-package environment variables passed (manager tests 83.514 seconds), followed by successful `go vet ./...`. Evidence: `work/asi-loader-verification/precommit-tests.log`. A read-only live `verify --json` again returned code 0 at generation 13; no game files were changed during documentation finalization.
- Final build gates: CORRECT. `scripts/build.ps1 -GameRoot 'C:\Games\METAL GEAR SOLID 3 - MCV'` ran module verification, tests, vet, build, and live doctor successfully. Manager tests took 82.310 seconds. Evidence: `work/asi-loader-verification/final-build.log`. Real-artifact environment variables were set for this build.
- Dependency spoof prevention: CORRECT. `internal/manager/loader_test.go` covers a schema-2 ASI named `asi-loader`; ASIs are classified by their targets and the loader requires schema 3 plus the pinned identity. Independent review found no remaining frozen-source blocker. The pre-freeze preview executable was stale and is not the released executable.
- Reproducibility and distribution contents: CORRECT. Two runs of `scripts/package-qcamo.ps1` produced identical SHA-256 values for both inner packages and the outer ZIP. An exact member-name comparison found 9 outer entries; all 8 checksummed members passed, including the final manager executable. Evidence: `final-package.log`, `final-repeat.log`, and `distribution-check.log` under `work/asi-loader-verification/`. The inner packages are byte-identical to the real artifacts used by the tests.
- Live installation: CORRECT. Normal manager commands imported the loader at generation 11, enabled it at generation 12, reused the unchanged QCamo package, and enabled QCamo at generation 13. `verify`, `status`, and `doctor` all returned code 0. Both packages are enabled; `camo-test` remains disabled; no recovery is pending. Evidence: `live-*.json` and `live-check.log` in the same directory.
- Live dependency guard: CORRECT. `disable asi-loader --dry-run --json` and `remove asi-loader --dry-run --json` both returned code 4 with `enabled ASI packages depend on the managed loader`. Neither changes generation 13.
- Installed bytes and unchanged texture: CORRECT. `Get-FileHash` confirmed the loader and QCamo hashes below. The original HQ olive texture remains `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`. `Test-Path` found no root `winhttp.dll`, `dinput8.dll`, `MGSHDFix.asi`, or `MGSHDFix.ini`. Evidence: `live-check.log`.
- User-reported local gameplay smoke test: the user stated "It works perfectly!" after being asked to launch the installed combination and test QCamo. This supersedes the pending local smoke-test gate. Independent observation and individual control/transition/cutscene scenarios remain UNVERIFIABLE; the report supplied no recording or scenario checklist.

## Final artifacts

- Installed and bundled `mgs3mod.exe`: 6,448,128 bytes; SHA-256 `2879a8281b8dd26a0edae654a7e8e527f22f82a4f3ab7d35aa9d19919cb2206a`.
- Current tester distribution, `dist/qcamo-standalone-tested/qcamo-1.0.4-standalone-distribution.zip`: 5,404,788 bytes; SHA-256 `c02c4c48e8ff3638a469d14e5659cf4a956d79f4d981f5fd0db5a22313a29709`. Its adjacent `qcamo-1.0.4-SHA256SUMS.txt` contains the distribution and inner-package checksums.
- `qcamo-1.0.4.mgs3mod.zip`: 4,064,453 bytes; SHA-256 `d79d7c91d46c2112f3e45af89b040ce65c40df0b7fa92e9158cc18f1d951a195`.
- `asi-loader-9.7.4.mgs3mod.zip`: 1,198,965 bytes; SHA-256 `9b0e26fc7ede3099a1112c0270568d5aa375bb121ec55d3f384d026af27ba9f1`.
- Installed `qcamo.asi`: 4,063,819 bytes; SHA-256 `4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193`.
- Installed `wininet.dll`: 1,198,304 bytes; SHA-256 `031a3e5576d91dce1e438d36b9a3d462c7334ab4791990a8ff1e3ddc0e132daf`.

After the user test, the bundled README was updated and the distribution rebuilt twice into `dist/qcamo-standalone-tested/` and `work/standalone-tested-repeat/`. All three ZIPs reproduced byte-for-byte; the exact 9-member layout and 8 member checksums passed. The bundled README matches the updated source and the executable remains unchanged. Evidence: `tested-package.log`, `tested-repeat.log`, and `tested-distribution-check.log` under `work/asi-loader-verification/`. The original `dist/qcamo-standalone/` ZIP (SHA-256 `a2d1cad7867283493e87f15d1a5176ccfa6032d8ac9d5a147a862814cc17f050`) is superseded only in documentation, not payload bytes.

The prior executable is retained as `work/asi-loader-verification/mgs3mod-before-standalone.exe` for comparison only. It is not a supported downgrade after schema-3 import. Historical previews and the old MGSHDFix-gated distribution are not the current release. Local commits were requested after the successful user test; no push or public release is authorized by that request.

## Limits

The user-reported smoke test does not establish long-session stability, all input modes, other game builds, or compatibility with other native mods. Module loading was not independently captured. No source rebuild is claimed. Loader custom configuration, arbitrary versions/proxy names, automatic dependency updates, and manually installed plugin dependency discovery are not supported.
