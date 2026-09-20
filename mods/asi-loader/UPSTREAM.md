# Standalone Ultimate ASI Loader provenance

This package contains the unchanged official Ultimate ASI Loader x64 binary by ThirteenAG. It is renamed from `dinput8.dll` to the officially supported proxy alias `wininet.dll`. It does not contain MGSHDFix or any MGSHDFix plugin, tool, settings, or fixes.

- Release: https://github.com/ThirteenAG/Ultimate-ASI-Loader/releases/tag/v9.7.4
- Source commit: `6b440669144c4a0bef5718ab155df160d231cd42`.
- Asset: `Ultimate-ASI-Loader-NoPDB_x64.zip`, 374,169 bytes.
- Download: https://github.com/ThirteenAG/Ultimate-ASI-Loader/releases/download/v9.7.4/Ultimate-ASI-Loader-NoPDB_x64.zip
- Archive SHA-256: `e5860e7d9a1805267535b65749575b5e406cc6ea3325c7392189c578815045d1`.
- Sole member: `dinput8.dll`, 1,198,304 bytes.
- Member and deployed `wininet.dll` SHA-256: `031a3e5576d91dce1e438d36b9a3d462c7334ab4791990a8ff1e3ddc0e132daf`.
- License: MIT; complete notice in `LICENSE.txt` in source and `LOADER-LICENSE.txt` in the distribution.

The supported game's `Engine.dll` imports `WININET.dll`, so this is the evidenced loader path. The game executable does not import `dinput8.dll`. Renaming is supported by the upstream release; no binary patch or rebuild is performed. No `winhttp.dll` is installed.

The pinned upstream source loads root ASIs by default (`LoadPlugins=1`, `LoadFromScriptsOnly=0`). It also supports plugins in `scripts`, `plugins`, and `update`. The manager's first standalone release supports default loader configuration and does not write INI files. Existing `wininet.ini`, `global.ini`, `scripts/global.ini`, `plugins/global.ini`, or `update/global.ini` must be resolved separately; they can override loading behavior and are not silently ignored or overwritten.

The manager owns only its pinned `wininet.dll`; it refuses an existing unowned file, including a byte-identical one. It retains the loader when QCamo is disabled or removed and blocks loader removal while managed ASIs are enabled. Other manually installed plugins may also depend on this shared loader: keep it installed if you use them. Runtime-generated logs or crash dumps are not removed.

File hashes prove which artifact is installed, not that the game loaded it successfully. Runtime testing must confirm the loader and QCamo initialize in the real game.
