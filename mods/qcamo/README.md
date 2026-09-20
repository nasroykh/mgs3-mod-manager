# QCamo for MGS3 Mod Manager

QCamo adds a quick uniform and face-paint selection menu to Metal Gear Solid 3: Master Collection. Hold G on keyboard or LB+Y on controller; select an owned item and confirm. See the [upstream README](https://github.com/zexk/mgs3-qcamo/blob/v1.0.4/README.md) for complete controls.

This distribution contains unchanged upstream QCamo 1.0.4 and standalone Ultimate ASI Loader 9.7.4, as separate manager packages. It installs no MGSHDFix. QCamo is the project's first non-test integration; `camo-test` is only a texture validation fixture. Share the complete outer distribution, including both upstream licenses and provenance records.

## Requirements

- Windows x64 and the exact game build listed in `UPSTREAM.md`. Other versions fail closed.
- The included MGS3 Mod Manager build, which supports schema-2 ASIs and the pinned schema-3 standalone loader package. No separate loader installation or runtime download is needed.
- No pre-existing unowned `wininet.dll` or `qcamo.asi`. The manager never adopts or overwrites those files, even if their bytes match. Resolve a prior manual installation separately.
- Default loader configuration: no existing `wininet.ini`, `global.ini`, or `global.ini` in `scripts`, `plugins`, or `update`. Custom configurations can disable root ASIs; this release refuses them instead of changing them.
- Close the game and launcher before manager changes.

## Install and manage

Extract the outer distribution to a folder of your choice. Open PowerShell there and set your actual installation path:

```powershell
$game = 'D:\SteamLibrary\steamapps\common\MGS3'
.\mgs3mod.exe doctor --game-root $game
.\mgs3mod.exe init --game-root $game
.\mgs3mod.exe add .\asi-loader-9.7.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable asi-loader --game-root $game
.\mgs3mod.exe add .\qcamo-1.0.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable qcamo --game-root $game --dry-run
.\mgs3mod.exe enable qcamo --game-root $game
.\mgs3mod.exe status --game-root $game
```

`add` stores each package disabled. Enabling `asi-loader` creates the pinned `wininet.dll` proxy imported by MGS3's engine. Enabling QCamo verifies that the loader is manager-owned and enabled, then creates `qcamo.asi`. No other loader proxy, MGSHDFix plugin, config tool, or game setting is installed or changed.

Keep using this schema-3-capable manager after import. Older builds cannot read the resulting journal history, even after packages are removed. Do not downgrade the executable or copy `.mgs3mod` between installations.

```powershell
.\mgs3mod.exe disable qcamo --game-root $game
.\mgs3mod.exe remove qcamo --game-root $game
.\mgs3mod.exe verify --game-root $game
```

Disable restores the original absence of the plugin. Remove also deletes its stored package. Neither removes the shared loader nor QCamo's generated `qcamo.log` or `qcamo-crash.dmp`. Unexpected changes to managed files block mutations. Do not manually delete `.mgs3mod`; it contains recovery records and retained baselines. State is bound to its installation path and is not portable between folders or PCs.

When you no longer use any ASI mods, remove the loader with `remove asi-loader --game-root $game`. The manager blocks this while a managed ASI is enabled. Manually installed ASIs can also depend on the loader; leave it installed if you use them. `restore --baseline --game-root $game` disables all managed packages together, including the loader and QCamo.

After an interrupted operation, run `recover --game-root $game --dry-run` and then `recover --game-root $game`. Recovery preserves unexpected external changes and may require investigation. Updates currently require disable/remove, add the new package, and enable; there is no atomic update command.

## Validation limits

On 2026-09-20, the user reported that this manager-installed QCamo and standalone loader combination works perfectly in game without MGSHDFix. This is a successful user-reported local smoke test; no recording or per-scenario test results were supplied. Keyboard/controller coverage, repeated swaps, area transitions, cutscenes, scoped controls, and coexistence with texture mods are not individually certified by that report. Treat this distribution as a tester preview, not a universal compatibility guarantee.

When reporting a problem, include your game version/store, Windows version, package checksums, other installed mods, input device, exact reproduction steps, and the output of `status --json` and `verify --json`. Redact personal paths before posting. If present, inspect `qcamo.log` for useful errors; do not publish crash dumps or saves without reviewing them for private information. Do not force installation if `doctor` reports an unsupported game build.

Known upstream reports at integration time: [incorrect camouflage icons](https://github.com/zexk/mgs3-qcamo/issues/2) and [LB+Y conflicting with scoped-weapon controls](https://github.com/zexk/mgs3-qcamo/issues/3). This unchanged binary does not claim to fix them.

## Build this distribution

From the manager source repository, build the manager, obtain the official QCamo v1.0.4 ZIP, then run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\build.ps1
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\package-qcamo.ps1 -UpstreamArchive 'C:\Downloads\qcamo-v1.0.4.zip' -LoaderArchive 'C:\Downloads\Ultimate-ASI-Loader-NoPDB_x64.zip'
```

The recipe verifies both upstream archives and payloads, creates two deterministic manager packages and one outer distribution, includes the manager executable and both licenses, and writes checksums. It refuses existing outputs. Generated staging files remain under ignored `work/`; releases go to `dist/qcamo-standalone/`. No game files, saves, backups, MGSHDFix, or manager state are included. Determinism assumes identical input manager bytes and packaging runtime. See `LOADER-UPSTREAM.md` for the pinned standalone loader download and hashes.
