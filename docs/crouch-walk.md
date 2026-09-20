# MGS3 Crouch Walk: local import

## Status and scope

This extension requires MGS3 Mod Manager v0.2.0 or later. Follow the [binary installation instructions](../README.md#install-without-go-or-git); Go and Git are not required. The v0.1.0 executable cannot read schema 4 or run `import-crouch`. Do not downgrade after importing a schema-4 package, even after removing that package, because transaction history retains its manifest.

The user confirmed successful gameplay on 2026-09-20, explicitly testing Crouch Walk and QCamo together. The local combined smoke-test gate is complete. This is a user-reported result, not independently recorded proof or a completed scenario-by-scenario compatibility matrix.

The adapter supports the original Nexus MGS3 Crouch Walk 0.2.1, file 32, archive SHA-256 `81c364f1253f33cda9244c89b7d8f9bc499fe780e4c00149d2ec505eca27a757` (22,061,918 bytes). Download it yourself from [Nexus Mods](https://www.nexusmods.com/metalgearsolid3mc/mods/27?tab=files&file_id=32). The manager does not download or embed this mod. The original archive and generated package must not be uploaded or included in releases without appropriate permission. The [author's permissions](https://www.nexusmods.com/metalgearsolid3mc/mods/27?tab=description) restrict redistribution and modification. This adapter preserves all included payload bytes; it does not rebuild the plugin.

## Design and implementation plan

1. Inspect every archive member and all embedded hook signatures without executing the plugin.
2. Pin schema 4 to `crouch-walk` version `0.2.1` and the complete 119-file set. Preserve existing schema policies; do not allow arbitrary animations, configurations, or DLLs.
3. Provide an offline conversion command that validates the original archive and payload checksums, then reuses normal package validation and deterministic packaging.
4. Reuse existing backup, ownership, exclusive creation, rollback, and recovery transactions. Test mixed existing/absent targets in isolated fixtures before any live installation.
5. Verify the live manager state, install through normal manager commands only, and require a separate user gameplay test.

The source archive has 117 animation files, an ASI plugin, an INI configuration, and a legacy `d3d11.dll`. The adapter excludes that DLL and requires the existing managed standalone ASI loader. MGSHDFix is not required by the adapter. The plugin and INI target names are normalized to lowercase, as required by manager policy; their bytes remain unchanged. Windows resolves the original mixed-case INI filename to the same file.

The animation paths contain 29 files each for `fr`, `it`, `jp`, and `sp`, plus one for `us`. There are 82 present-origin animation replacements and 35 absent-origin animation additions; ASI and INI add two more absent-origin targets. All 119 target states and payload hashes are pinned in `internal/profile/crouch-files.json`. Baselines describe this supported local installation, not a claim that every file is an untouched factory original. Different asset sets are rejected rather than silently captured or overwritten.

All target parent directories must already exist, as they do in the inspected installation. The adapter does not create missing language directories in the game. A failed preflight is a compatibility stop, not a reason to bypass path checks.

## Local conversion and installation

Close the game and launcher. Open PowerShell in a writable folder. Use the new executable's full path if another manager version is on PATH. The output directory must already exist; existing output files are never replaced.

```powershell
$manager = 'mgs3mod'
$game = 'C:\Games\METAL GEAR SOLID 3 - MCV'
$archive = 'C:\Games\MGS3 Mods\MGS3CrouchWalk-27-0-2-1-1701782548.zip'
$package = '.\crouch-walk-0.2.1.mgs3mod.zip'
& $manager import-crouch $archive --out $package --dry-run
& $manager import-crouch $archive --out $package
& $manager verify --game-root $game
& $manager add $package --game-root $game
& $manager enable crouch-walk --game-root $game --dry-run
& $manager enable crouch-walk --game-root $game
& $manager verify --game-root $game
```

Conversion does not initialize the manager, import the package into its library, enable it, or touch the game. Normal `add` and `enable` perform those distinct steps. The managed `asi-loader` must already be enabled; use the README's loader instructions if needed. QCamo is optional and remains independently managed.

## Disable, remove, and configuration

```powershell
& $manager disable crouch-walk --game-root $game
& $manager remove crouch-walk --game-root $game
```

Disable restores the captured 82 existing animation files and removes the 37 files whose original state was absence. Remove also removes the package from the manager library. Neither operation removes QCamo or the shared loader. Empty asset directories may remain. Interrupted operations use the existing `recover` workflow; do not manually copy files over an unresolved transaction.

The INI is deliberately immutable in this first adaptation. Editing it causes drift verification to fail, and the manager refuses to overwrite or delete the edited file. Preserve your edit outside the game and restore the exact managed INI before disabling. Configurable settings require a separately designed managed configuration feature; they are not silently adopted.

## Validation limits

Local validation on 2026-09-20 passed: strict schema/profile tests, original-archive checksum rejection, no-write dry-run, output preservation, deterministic conversion, complete package loading, all-target enable/disable/remove, configuration drift refusal, QCamo preservation in fixtures, six mixed-file crash-recovery cases, and explicit missing-animation recovery. The remaining Go regression suite and `go vet ./...` passed. Real-artifact tests are opt-in because upstream and game payloads are not redistributed with the repository.

Live installation then passed normal `add`, enable dry-run, `enable`, and `verify`: generation 15, Crouch Walk 0.2.1 enabled alongside unchanged QCamo 1.0.4 and Ultimate ASI Loader 9.7.4. No `d3d11.dll` or MGSHDFix was installed. This is installation verification, not a gameplay result. Local package SHA-256: `37cb9ca04ecd536cd4bee7a04444698856a8014860fb8cb4fd9d59a3d2d722e5`; tested/deployed manager SHA-256: `70f475e36d4962e3983621c9fb7af1fc5ee1de76ac193eb62b67157b326339bc`.

An initial artifact test failed because its fixture omitted the existing Japanese asset parent directory. The live import dry-run succeeded; correcting the fixture to mirror verified source directories resolved all eight failing cases without weakening production path checks. Logs are retained locally under `work/crouch-manager-tests.log`, `work/crouch-manager-tests-fixed.log`, and `work/crouch-suite.log`.

To reproduce artifact checks before enabling Crouch Walk, set `MGS3MOD_CROUCH_ARCHIVE` to the original download, `MGS3MOD_CROUCH_PACKAGE` to the locally converted ZIP, and `MGS3MOD_CROUCH_BASELINE_ROOT` to the supported installation with Crouch Walk disabled. Then run `go test -count=1 ./...` and `go vet ./...`. The baseline-root tests read that installation and write only isolated fixture directories. They deliberately reject an enabled/modified baseline. Tests without these variables do not exercise the real upstream payloads.

Static inspection of the supplied plugin found 11 embedded hook signatures. A scan of the PE image reconstructed from the supported executable found each exactly once. This confirms signature presence only, not function semantics, structure layouts, hook coexistence, or gameplay correctness. Latest upstream source was inspected at commit `422c01757cef2722fbd959e7725b1b0e27e2037a`; it is not asserted to be the exact source of the downloaded binary. No rebuild was performed.

The user reported that the combined gameplay test worked well; individual scenario outcomes were not supplied. Broader tester coverage should still include slow and fast crouch movement, stand/prone transitions, weapons, first-person view, damage/knockdown, animation/audio, camo index, NPC detection, save/load and area transitions, and QCamo swaps while crouched. Stop on crashes, frozen movement, broken animations, or unexpected camo values; close the game and disable Crouch Walk using the new manager. Passing package tests alone is not a gameplay pass.
