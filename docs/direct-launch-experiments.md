# Direct-launch experiment log

Date: 2026-09-21. Game executable SHA-256: `0D585DCC6A671BE5D64D3D0A856C53F9EE0E58E7E4993F76DFF29772C7A4BC80`.

## Trial 1: normal launcher baseline

- Route: `launcher.exe`, started from the game directory with no arguments.
- Result: no launcher or game process was present when queried after `Start-Process` returned. The launch command produced no process receipt, error, or exit code.
- Landing screen, controller prompts, save availability, return behavior, and mod behavior: UNVERIFIABLE.
- Interpretation: this noninteractive launch attempt does not establish a normal-launch baseline. Repeat from the interactive desktop and record the selected region, language, controller, process tree, and visible result.

## Trial 2: native launch with launcher path

- Working directory: `C:\Games\METAL GEAR SOLID 3 - MCV`.
- Executable: `C:\Games\METAL GEAR SOLID 3 - MCV\METAL GEAR SOLID3.exe`.
- Argument array: `-region`, `us`, `-lan`, `en`, `-selfregion`, `EU`, `-launcherpath`, `launcher.exe`, `-ctrltype`, `KBD`.
- Process receipt: PID 40064 started at 2026-09-21T12:13:28.8051356+01:00. It remained alive and responsive beyond five minutes with main-window title `METAL GEAR SOLID 3 SNAKE EATER`. WMI reported the expected executable and exact command line. No launcher process was observed alongside it.
- Pre-launch save inventory: four files under `mgs3_savedata_win`, 23,449 total bytes. The account directory was `76561197960271872`; one game-save directory, `STE1316501G000001`, was present. Full relative paths, sizes, timestamps, and SHA-256 values were captured in the session evidence.
- Visible result: the user reported that it went directly to the game and appeared to work well. This is user-reported evidence that the collection selector was bypassed. The exact first game screen remains UNVERIFIABLE because it was not named.
- Exit behavior: the user closed the game manually. The recorded PID exited and no game or launcher process remained afterward; no launcher reopening was observed.
- Post-launch save inventory: file count and total bytes were unchanged. Both files in `STE1316501G000001` retained their pre-launch hashes and timestamps. `launcher/launcher_sv` was unchanged. `launcher/usersv` retained its 4,096-byte size but changed hash from `898535DFEED138CE06E2CAB2065D2B8A4535D959FA301B2CF436713B22585D82` to `8D56EFF4256937823DC1C7D7E0BC302977E27A9C5104B679F4B5EF0BE2AA2749` at 12:18:05 local. Treat this as a launcher-settings change, not a game-save rewrite.
- QCamo evidence: `qcamo.log` was recreated for this process at 12:13:28. It reported readiness, found the inventory table, recorded keyboard and pad menu input, applied multiple uniform selections, and completed changes. The prior `qcamo-crash.dmp` timestamp did not change. This proves QCamo loaded and operated during this direct-launch trial.
- Crouch Walk behavior, exact controller prompts, manual save-list visibility/loading, and the precise landing screen: UNVERIFIABLE.
- Post-launch manager verification: `mgs3mod.exe verify --json` returned `ok: true`, generation 15, initialized and compatible, with no issues. PowerShell serialization displayed a singleton `null` for the absent issue list; this was not a manager-reported issue.
- Screenshot attempt: failed with `Exception calling "CopyFromScreen" with "5" argument(s): "The handle is invalid."`; do not treat the generated image as visual evidence.

## Trial 3: native launch without launcher path

- Working directory and executable matched Trial 2.
- Argument array: `-region`, `us`, `-lan`, `en`, `-selfregion`, `EU`, `-ctrltype`, `KBD`. This was the one additional launch explicitly approved by the user.
- Process receipt: PID 46912 started, remained alive after the five-second observation deadline, and had the expected executable and exact command line. No launcher process was observed.
- User report: the game opened directly, an existing save loaded manually, Crouch Walk worked, and QCamo worked. The user then closed the game manually. This establishes the save-list/load path and both mods as user-observed behavior, but it still does not name the exact first game screen.
- QCamo evidence: the new log began at 12:23:38, reported readiness and the inventory table, and recorded successful uniform changes. At 12:24:55 it then recorded access violation `C0000005`, a null read, and a new `qcamo-crash.dmp`. The process exited immediately afterward. The user experienced this as a manual close, but the new dump means shutdown was not clean at artifact level.
- Post-launch save inventory: all four files retained the exact pre-launch sizes, timestamps, and SHA-256 values. No game-save or launcher-settings file changed.
- Post-launch manager verification: `mgs3mod.exe verify --json` returned `ok: true` at generation 15 with initialized and compatible state.
- Decision: omission of `-launcherpath launcher.exe` is rejected. Gameplay worked, but its shutdown access violation fails the plan's acceptable-exit requirement. The supported preset retains `-launcherpath launcher.exe`.

## Gate status

Phase 0 supports a startup-only implementation using the full Trial 2 argument array, including `-launcherpath launcher.exe`. Across both native trials, the collection selector was bypassed, an existing save loaded manually, QCamo and Crouch Walk worked, game-save payloads remained intact, and manager verification passed. Trial 2 also had a clean manual close without observed launcher reopening. Trial 3 proves that omitting the launcher path is unsafe on this installation because shutdown produced a new access-violation dump.

The exact first game screen remains UNVERIFIABLE, so expose destination `startup` and preset `na-startup`; keep literal `menu` unavailable. Only the tested North America, English, and KBD argument combination is initially supported. Phase 1 and Phase 2 may proceed within that capability. Save autoload remains gated.
