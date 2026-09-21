# Direct launch

Status: complete for the startup-only scope as of 2026-09-21. Save autoload, automatic Main Menu advancement, additional game configurations, and graphical controls are separate future features.

v0.3.0 and current source builds can start the supported MGS3 installation without opening the Master Collection selector. v0.2.0 does not include this command. This feature requires an initialized manager and the exact compiled game fingerprints already enforced by other manager commands.

The first supported preset is `na-startup`:

- game region: North America (`us`)
- language: English (`en`)
- controller prompts: keyboard (`kbd`)
- destination: game startup (`startup`)

The native argument array is fixed:

```text
-region us -lan en -selfregion EU -launcherpath launcher.exe -ctrltype KBD
```

`-selfregion EU` reflects this installation's launcher region and is independent of the selected North American game region. `-launcherpath launcher.exe` is retained deliberately. A trial without it reached gameplay, but shutdown produced a new access-violation dump, so that shorter command is unsupported.

## Use the selector

```powershell
mgs3mod launch --select --game-root $game
```

The numbered selector displays the current profile and resolved choices. It can select a profile, edit the current choices, save a named profile, choose the default, launch, or cancel. Enter accepts the displayed default action. Cancel, `q`, Escape input, or end of input exits without starting the game.

The built-in preset exists only in memory until an explicit save or default action writes preferences. Saved preferences use `mgs3mod-launch.json` in the selected game root. They remain separate from `.mgs3mod` journal state, do not change the mod generation, and can be reset by removing only that preference file while the game and manager are closed.

[`examples/direct-launch/mgs3mod-launch.json`](../examples/direct-launch/mgs3mod-launch.json) shows the canonical schema-1 preset document. Prefer saving through the selector; the example is for inspection and automation that writes a new file while the game and manager are closed.

## Noninteractive commands

```powershell
# Launch the in-memory or saved profile.
mgs3mod launch --profile na-startup --game-root $game

# Supply the complete supported selection for one launch.
mgs3mod launch --region us --language en --controller kbd --target startup --game-root $game

# Inspect the exact executable, working directory, arguments, and selection.
mgs3mod launch --profile na-startup --dry-run --json --game-root $game

# After a default has been saved:
mgs3mod launch --game-root $game
```

Scripts must provide a profile, a complete selection, or a persisted default. They never receive an interactive prompt. `--json` never prompts, and `--select` cannot be combined with JSON, dry-run, a profile, or overrides. CLI overrides affect one launch only and never save implicitly.

Dry-run still takes the manager lock briefly and performs fingerprint, initialization, recovery, process, path, and selection checks. It starts no process and writes no preferences.

## Safety and current limits

Launch preflight rejects an uninitialized manager, incompatible fingerprints, unresolved recovery, unsafe game or launcher paths, a running game or launcher, or another manager holding the lock. The game starts through a fixed argument array with the game directory as its working directory; no shell command is constructed. The lock remains held until the started PID is observed as the expected executable. A failed handoff never starts a second process. If the process remains alive but cannot be verified, the command waits with the lock held and reports failure only after exit; close the game to release that exceptional guard.

A successful receipt means that the expected game process started. It does not by itself prove that a particular screen was reached. Process evidence, a captured frame, and unchanged save payloads established collection-selector bypass for the installed manager. On 2026-09-21 that launch reached the in-game DATA LOAD screen; the frame before that screen was not captured. The user observed manual loading of an existing save, QCamo changes, and Crouch Walk during the earlier native-argument trials. `menu` remains unavailable and `startup` is the only honest destination.

Other regions, languages, controller-prompt modes, game builds, and distribution layouts are rejected until tested. Save autoload is not implemented. See [the experiment log](direct-launch-experiments.md) for local evidence and [the implementation plan](direct-launch-plan.md) for the gated save-autoload research.
