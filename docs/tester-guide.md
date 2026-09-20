# Small-group tester guide

Start with a few volunteers who already own MGS3 Master Collection. Windows AMD64 and the supported game build are required; compatibility with other builds is not established. One local user has reported successful combined QCamo/Crouch Walk gameplay. We are collecting independent results, not advertising universal compatibility.

## Get ready

1. Use your own game installation and keep a private backup of important saves using your normal backup method. Do not upload saves or game files.
2. Install [MGS3 Mod Manager v0.2.0](https://github.com/nasroykh/mgs3-mod-manager/releases/tag/v0.2.0) using the [README](../README.md#install-without-go-or-git). Verify checksums. Go and Git are not required.
3. Use your actual game path with `--game-root`. Close the game and launcher before manager changes. Run `doctor`; stop if the build is unsupported. Do not bypass fingerprint checks or change game files to force compatibility.
4. Follow the [QCamo setup](../README.md#quick-start-game-setup-and-qcamo) and, optionally, [Crouch Walk local-import instructions](crouch-walk.md). Each tester must obtain their own original Nexus Crouch Walk download. Do not share its original archive or generated package.
5. Record the manager release and executable used, mod versions, game/store version and language, Windows version, input device, and other installed mods. Redact personal path components. Use the updated executable explicitly if several copies exist.

Keep the first run simple: use the managed standalone loader and selected supported mods. Do not install MGSHDFix for this test, add a second proxy DLL, edit the managed INI, or rearrange an existing manual mod setup just to qualify. Report an incompatible setup instead. Never downgrade to v0.1.0 after schema-4 import, even after removing Crouch Walk.

## Manual smoke checklist

Use an existing safe save/checkpoint. Record Pass, Fail, or Not tested for each row; do not infer a pass for something you did not exercise. Repeat only cases relevant to the input device and mods you have. Do not deliberately interrupt writes or delete files to test recovery; those cases belong in isolated developer fixtures.

| Area | What to exercise | Evidence to report |
| --- | --- | --- |
| Preflight and install | Doctor, documented add/enable, then verify | Commands, exit result, or exact error |
| QCamo | Open menu, select owned camo and face paint, repeat swaps | Input device, menu/selection result |
| Crouch Walk | Slow/fast movement and stand/prone transitions | Movement and animation result; mark unsupported input cases not tested |
| Controls and weapons | Switch weapons and enter/leave first-person view | Any stuck input, incorrect motion, or view problem |
| Combined mods | Use QCamo while crouched, then resume movement | Whether both functions remain usable |
| Gameplay state | Observe camo index, movement sound and nearby NPC response; damage/knockdown only when convenient | Observations, location and reproducible steps, not guessed internal behavior |
| Transitions | Change area, reload a safe save, and observe a normal cutscene when available | Crashes, lost controls, or persistent problems |
| Safe removal check | Close game, disable Crouch Walk if installed, then verify | Success/failure; whether QCamo remains enabled |

Disabling only Crouch Walk should leave QCamo and the shared loader managed separately. Use the executable you installed; for a manual ZIP installation, replace `mgs3mod` with its full path:

```powershell
$game = 'D:\SteamLibrary\steamapps\common\MGS3' # Replace with your actual folder.
mgs3mod disable crouch-walk --game-root $game
mgs3mod verify --game-root $game
```

Run the disable command only if you installed Crouch Walk. If you want it back, follow the documented enable dry-run and enable steps. Do not disable the shared loader while managed ASIs remain enabled.

## Report results safely

Use [Compatibility report](https://github.com/nasroykh/mgs3-mod-manager/issues/new?template=compatibility_report.md) for a completed session, including successes or unsupported-build stops. Use [Bug report](https://github.com/nasroykh/mgs3-mod-manager/issues/new?template=bug_report.md) for a specific reproducible defect. Search existing issues first and add matching evidence instead of duplicate reports. GitHub issues are public.

For diagnostics, close the game and run these read-only commands with your actual game root:

```powershell
mgs3mod status --game-root $game --json
mgs3mod verify --game-root $game --json
```

Inspect output before posting. Remove usernames, personal directories, account identifiers, and unrelated information. Quote only relevant errors or a redacted summary if full output is unnecessary. Never attach game binaries/assets, Crouch Walk packages, `.mgs3mod`, saves, full crash dumps, credentials, or unreviewed logs. There is no automatic diagnostic upload or telemetry.

If gameplay crashes or controls break, close the game and try the documented disable operation. If a manager command fails, reports drift, or reports pending recovery, stop repeated mutations and preserve local evidence. Consult [recovery guidance](recovery.md); do not manually delete state or force replacement. Do not disable security software in response to a warning.

## Maintainer follow-up

Track individual outcomes in issues, not a public spreadsheet containing personal paths. Separate unsupported game/build reports from reproducible defects. Prioritize data-integrity or recovery failures, then crashes, then gameplay regressions. A successful report broadens evidence only for its stated environment and tested cases. Missing cases remain unverified. Address concrete reports before expanding configuration options or starting another mod integration.
