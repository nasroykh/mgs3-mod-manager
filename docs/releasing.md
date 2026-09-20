# Release process

Windows AMD64 tags use `vMAJOR.MINOR.PATCH`. v0.1.0 is the first narrowly scoped release, not support for every game build. The installer resolves the latest published non-prerelease or an explicit version and verifies its required SHA-256 entry. Public installation requires public repository/release access.

## Assets

The manager ZIP is `mgs3mod_<version-without-v>_windows_amd64.zip` with exactly `mgs3mod.exe`, `README.md`, and `THIRD-PARTY-NOTICES.txt` at root. Other assets: QCamo and standalone loader packages, the complete QCamo distribution, `install.ps1`, and `checksums.txt`. Never upload Crouch Walk's original archive or locally converted package, game files, state, saves, logs, caches, or the packaging staging subdirectory. Crouch Walk support ships as adapter code and pinned metadata only; users provide their own download.

## Release gates

1. Review source, game compatibility limits, and `docs/releases/<tag>.md`.
2. Run `scripts/build.ps1` and `scripts/test-installer.ps1` on Windows.
3. Obtain the official pinned QCamo and UAL archives. Run `scripts/package-release.ps1 -Version v0.2.0 -UpstreamArchive <zip> -LoaderArchive <zip>`. The recipe rejects mismatched upstream hashes and existing outputs.
4. Repeat packaging with a fresh output directory; compare assets. Run real-artifact tests with `MGS3MOD_QCAMO_PACKAGE` and `MGS3MOD_ASI_LOADER_PACKAGE` set to the generated packages.
5. Commit, push, and tag the reviewed source. Do not move a published tag or replace published bytes.

The tag-triggered workflow builds on Windows using `go.mod`, tests the installer, downloads pinned upstream releases, validates their hashes, packages, and runs actual-artifact tests. It creates a draft with all assets, then publishes as latest after successful upload. No personal access token is needed; GitHub provides the workflow token. A failed draft is not a published release.

After publication, check anonymous asset URLs and uploaded hashes. Run the downloaded installer into an isolated directory using `-NoPath`, then check the executable's help and read-only verification. This does not prove a persisted PATH update; that branch has separate tests. Gameplay cannot be tested in CI without the game and a human trial.

To repair a failed publishing workflow without moving an existing version tag, push the reviewed workflow fix to `release/<version-tag>` (for example `release/v0.1.0`). This uses the workflow from the repair branch but explicitly checks out the original immutable tag for build, tests, assets, and notes. It cannot silently overwrite an existing release: draft creation fails if the release already exists. Installer tests run in a child PowerShell process so expected negative-test exit codes do not leak into the runner's final status.

Hosted Windows runners may use an 8.3 alias in TEMP and grant parent-directory DELETE_CHILD rights. The workflow uses a canonical scratch path under `work/test-temp` and denies that alternate delete-child route on both this scratch tree and the separate `.cache/winfs-tests` fixture tree. File DELETE rights remain available until a fixture explicitly denies them. This keeps strict production path validation and the negative ACL test meaningful without skipping tests or changing game-directory permissions.

If a gate fails, stop publication. For a published defect, issue a fixed new release; never silently downgrade state or rewrite the existing version. Older binaries may not read schema-3 or schema-4 history, even after the associated mod is removed. Binaries are not Authenticode-signed; do not advise disabling security tools.
