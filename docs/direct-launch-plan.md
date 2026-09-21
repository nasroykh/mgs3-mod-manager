# MGS3 direct launch and optional save autoload

Date: 2026-09-21. Status: Phases 0–2 are implemented in the working tree for startup-only direct launch. The native argument route was runtime-tested, and the built manager passed real-root dry-run preflight. Literal Main Menu arrival and save autoload remain unavailable.

## Goal and decision

Extend the existing `mgs3mod` executable with a small selector that remembers how to start MGS3. The first supported preset will launch the North American game in English without opening the Master Collection selection menu. A later, separately gated extension will load a chosen save automatically if research establishes a reliable native loading path.

Use the game's existing command-line arguments. Do not install or fork all of MGSHDFix. Do not replace the existing ASI loader, patch game executables, or expand mod package schemas for basic launching. Keep existing QCamo and Crouch Walk behavior intact.

The initial selector will be a numbered console interface inside the existing Go CLI. This is a deliberate minimal interface: saved defaults, explicit choices, and one launch action. A graphical front end is outside this plan unless the console interface proves unsuitable.

## Evidence and confidence

The installed managed launcher was inspected using its bundled Mono.Cecil library. `Def::.cctor` defines `-region jp/us/eu`, `-lan jp/en/fr/it/gr/sp`, `-selfregion JP/NA/EU`, and `-ctrltype XBOX/PS4/PS5/NX/KBD`. `DefManager::GetArguments` combines region, language, and launcher region. `BootGameSteam::CreateProcess` appends launcher path and controller arguments. Its constructor builds `-launcherpath <product-name>.exe`.

The local launcher's `regionLauncher` default is EU. This is separate from the selected game region: selecting North America must not automatically change `-selfregion` to NA. The supported local startup route is:

```powershell
& '.\METAL GEAR SOLID3.exe' -region us -lan en -selfregion EU -launcherpath launcher.exe -ctrltype KBD
```

This command was first derived from source, then exercised in the Phase 0 trials from the game directory. The trial with the full argument array established process lifetime and collection-selector bypass; artifact checks preserved save identity and manager state. The user observed manual save loading, keyboard control, QCamo, and Crouch Walk behavior.

Local evidence fingerprints:

- Game executable SHA-256: `0D585DCC6A671BE5D64D3D0A856C53F9EE0E58E7E4993F76DFF29772C7A4BC80`.
- `launcher.exe` SHA-256: `E061111CEF605BDBF0EA7BC9CF686A1D29E317BF1DA923E3C92C52F523479784`.
- `launcher_Data/Managed/Assembly-CSharp.dll` SHA-256: `3F4DE01B2DE9E9EFC31001C0AD376B294EAEEC5DD092F7E681204E93684AF945`.
- Repository baseline: branch `main`, HEAD `804a4e0ccb398d7cd3eb0c391c1a3cdaee31d2c0`; clean before this document was added.
- Installed manager status: generation 15; ASI loader, QCamo, and Crouch Walk enabled; camo test disabled. This was read from `mgs3mod.exe status --json`, not inferred from the historical handoff.

End-to-end startup compatibility is established for this fingerprinted installation and exact argument array. The exact first game screen remains UNVERIFIABLE because it was not recorded, and arbitrary save autoload is also UNVERIFIABLE because no native load entry point or supported save argument has been established.

## User-facing behavior

The selector shows the current preset, game region, language, controller prompts, and destination. Enter accepts an explicitly displayed default; Escape or a cancel choice exits without launching. There is no countdown or automatic launch while editing settings.

Implemented startup-only commands:

```text
mgs3mod launch --select
mgs3mod launch --profile na-startup
mgs3mod launch --region us --language en --controller kbd --target startup
mgs3mod launch --profile na-startup --dry-run --json
```

`launch --select` is the interactive entry point. `launch` alone uses the saved default if one exists; otherwise it opens the selector only on an interactive terminal. Scripts must supply a complete selection or saved profile and must never hang on a prompt. `--json` must never prompt. Conflicting `--select` and noninteractive modes are rejected clearly.

The first selector provides: select an existing profile, edit current choices, save as a named profile, choose the default, launch, or cancel. Seed `na-startup` in memory with North America, English, keyboard prompts, and no save autoload; do not write configuration until the user saves it. Existing controller preference was not reliably established before the trials, so the selector shows the default prominently.

Profile values use enums, not arbitrary command strings. CLI overrides apply to the selected profile for that launch only. Saving is explicit. Region and language combinations must be verified against installed content; do not present every theoretical combination as supported merely because its argument exists.

The requested `menu` destination must mean the actual Main Menu without loading a save. Collection-menu bypass does not prove that intro logos or the game's own title confirmation are skipped. Phase 0 did not record the exact landing screen, so the implementation exposes `startup` with preset `na-startup` and keeps `menu` unavailable until state-aware title advancement is proven. Reuse the bounded state research in Phase 3 for that advancement; do not hide timed keypresses behind the menu option. The startup-only milestone does not complete the literal Main Menu requirement.

Save selection is added only after the autoload gate passes. Until then, do not display a working-looking “load save” option or silently interpret it as normal startup.

## Phase 0: prove the smallest launch path

Perform this phase before building the selector. Keep a short experiment log with the executable hash, exact argument array, working directory, launch route, resulting process tree, visible landing screen, exit behavior, and result.

1. Establish a normal-launch baseline for North America/English using the existing launcher. Record visible controller prompts, available saves, and return-to-launcher behavior.
2. Run the native candidate above with the same preferences. Confirm the initial collection selector does not appear and the expected game content starts.
3. Test the otherwise identical command without `-launcherpath`. Choose omission only if startup, saves, normal quit, and any return-to-collection action behave acceptably. Do not assume omission suppresses launcher reopening.
4. Confirm an existing save is listed and can be loaded manually, then verify QCamo and Crouch Walk in gameplay. Distinguish user-reported gameplay results from artifact-proven checks.
5. Test the controller mode actually used by the user. Test other exposed choices before documenting them as runtime-supported.
6. For a Steam launch route, check Steam initialization, overlay/input behavior, and save identity separately. The presence of local Steam-related files does not establish ownership, Steam client behavior, or cloud compatibility. Start with the current installation's working route.
7. Capture before/after save inventories and managed-file verification. Explain legitimate game-written settings separately from unexpected save changes. Do not replace, rename, decrypt, or rewrite saves during these trials.

Do not create `steam_appid.txt` speculatively. If a verified Steam initialization failure requires it, document the exact reason, ownership, cleanup, and limitations before adding that behavior. Do not treat a direct executable shortcut as proven equivalent to launching through Steam.

Gate: proceed when the direct native launch reliably bypasses the collection selector, reaches the documented game startup screen, preserves save identity, and retains existing mod functionality. If it fails, isolate the missing behavior; do not respond by installing all of MGSHDFix.

Outcome on 2026-09-21: the startup-only gate passed. The full argument route with `-launcherpath launcher.exe` bypassed the collection selector, preserved game-save payloads, allowed a manual save load, and retained QCamo and Crouch Walk. The user closed that trial normally. A second approved trial without `-launcherpath` reached gameplay but produced a shutdown access-violation dump, so that variant is rejected. The exact first game screen was not recorded; the implementation therefore exposes only `startup` and does not claim literal Main Menu arrival. See [the experiment log](direct-launch-experiments.md).

## Phase 1: launch service and saved profiles

Keep responsibility split into three small layers:

1. `internal/launch`: profile types, enum validation, argument construction, and versioned profile persistence. The argument builder is pure and independently testable.
2. `internal/manager/launch.go`: installation-bound preflight and launch coordination, reusing the existing root, fingerprint, process, and lock protections.
3. `internal/cli/launch.go`: command handling and the numbered selector. Existing CLI parsing/help and the executable entry point receive only the wiring they need.

Use a new `mgs3mod-launch.json` in the selected game root, outside `.mgs3mod`. The initialization recovery inventory (`internal/manager/init.go`, `initSpec` and `finishInit`) allows only its declared state entries; preferences outside it avoid a recovery-schema change. The initial feature requires an already initialized manager. Missing or incomplete initialization returns an actionable error rather than silently initializing state. Keep it separate from transactional mod state: saving a launch preference must not increment a mod generation or alter installed packages. Give it its own schema version, default profile ID, and a map of named profiles containing region, language, controller, and `destination` (`startup` or `menu`). The CLI option remains `--target`; it maps to `destination` in configuration. Do not store executable paths, shell commands, user account IDs, or save payloads in a profile.

Load configuration strictly; reject malformed or unsupported versions with an actionable error. Do not overwrite an invalid file automatically. Serialize explicit profile writes with existing manager locking. The existing transaction JSON writer exclusively creates files and must not be mistaken for a replacement operation. Write a uniquely named temporary file in the same directory, flush and close it, then use a verified Windows atomic replacement operation; clean up only the temporary file owned by this write. Test failure before and during replacement and interruption after replacement: readers must see either the old complete document or the new complete document, never partial JSON. If no suitable existing primitive exists, add one focused helper with fault tests. Preserve the previous valid configuration when replacement fails. Apply the manager's path/reparse protections to the new file location.

Launch flow:

1. Resolve and validate the complete selection. Resolve the game root through existing manager behavior.
2. Acquire the manager lock. Reject unresolved recovery, incompatible fingerprints, missing selected content, or an already running game/launcher. Do not auto-recover or change mods.
3. Build a fixed argument array and use `os/exec` directly with the absolute game executable and game root as working directory. Never run a user-constructed shell command.
4. In dry-run mode, emit the resolved executable, working directory, arguments, and profile without starting a process or saving preferences. State any transient locking behavior explicitly.
5. Start the game while retaining the lock through the handoff. Verify the started PID and its expected executable are observable through the process guard, using bounded polling with a five-second initial deadline. Test and tune that deadline during Phase 0. An immediate exit or observation timeout produces a failure receipt; never automatically start a second process. Release the lock after successful observation or confirmed child exit. If observation fails while the child remains alive, the command stays attached and retains the lock until exit, then reports that launch protection remained active. This exceptional path deliberately reports after exit because returning earlier would release the process-owned lock and permit concurrent mod mutation. Successful handoff releases the lock promptly, without supervising the whole gaming session.
6. Return a start receipt containing PID and resolved settings. “Process started” must not be reported as “Main Menu reached” or “save loaded.” Handle immediate start failures without corrupting saved preferences.

Do not run a full managed-asset hash sweep on every launch. Reuse required compatibility and recovery checks; leave full integrity verification as an explicit command. Trace existing preflight internals before splitting anything, so performance changes do not weaken existing mutation safeguards.

Initially support this existing fingerprinted installation. New executable builds or distribution layouts need explicit compatibility work, not a force flag.

## Phase 2: selector verification and delivery

Automated checks should cover behavior rather than implementation shape:

- Exact native arguments for supported selections, especially `region us` with local `selfregion EU`.
- Reject invalid enums, unsupported combinations, missing profiles, contradictory options, and save targets before autoload exists.
- Paths with spaces; correct working directory; no shell execution.
- Cancel and noninteractive input; JSON contains no prompts; CLI overrides do not persist implicitly.
- Dry-run does not start a process or persist a profile.
- Corrupt config, interrupted write, stale default ID, and competing launch attempts.
- Recovery pending, wrong fingerprint, running game/launcher, launch-start failure, delayed process visibility, immediate child exit, observation timeout, and no duplicate start after failure.
- A fake process runner and injected stdin exercise the flow without starting the real game in tests.

Run relevant new tests, then the existing Go suite and `go vet ./...`; use the current build script for the deliverable. Run the existing manager verification once after live trials. Record any unavailable check as UNVERIFIABLE rather than claiming it passed.

Deliver the updated single executable, documentation, example presets, and an optional Windows shortcut targeting `launch --select` with the correct game root. No background service, web server, second launcher runtime, or extra loader is needed. Do not replace the installed executable until its build and runtime checks pass. Publication is a separate release action.

Gate: the user can choose and save a preset, launch it again directly, and reproduce the Phase 0 result. Removing the new preferences or using the original launcher must remain straightforward.

Outcome on 2026-09-21: profile persistence, selector behavior, process handoff, failure paths, and dry-run behavior have automated coverage. The full Go suite, `go vet ./...`, the repository build script, and a real-root JSON dry-run passed. The dry-run did not create `mgs3mod-launch.json`. A non-dry launch through the newly built manager remains UNVERIFIABLE because no additional game opening was authorized after the build; Phase 0 independently validates the exact native argument route. After verification, the root-level manager was atomically replaced and the installed binary passed the same dry-run. Installed `verify --json` returned compatible and initialized state at generation 15 with no issues. The previous executable is backed up as `dist/mgs3mod.previous.exe`.

## Phase 3: bounded save-autoload feasibility research

This is a research milestone, not a promise that the implementation is already known. Complete Phase 2 independently.

1. Locate the active save account and slot mapping used by this exact game build. Inventory saves read-only. Start with a stable slot identifier and file metadata; decode richer labels only if needed and understood. If multiple accounts exist, require an explicit selection rather than guessing from timestamps.
2. Trace normal title, load-list, selection, and load-confirmation behavior. Identify the native state transitions, the operation that resolves a chosen slot, and the thread on which it is safe to call it. Inspect loaded native modules as well as the executable; do not assume the entry point belongs to one binary.
3. Establish whether a small state-aware input sequence can reliably select the exact slot, or whether invoking the native load routine is necessary. Prefer the smallest reliable mechanism; fixed delays and blind repeated confirmation keys are not acceptable.
4. If a hook is required, document the supported module fingerprints, unique validated signatures, calling convention, arguments, lifetime requirements, thread constraints, and observable success/failure states. Never ship guessed addresses or call a loader-thread function into gameplay code.
5. Prove a one-shot load of one chosen save on the current build, then test another slot, missing slot, changed save contents, unsupported build, and timeout. A load request must not select “latest” by accident or depend on list order.

A save resumes according to MGS3's own saved checkpoint semantics. Do not promise restoration to an arbitrary exact position.

Required research output: evidence of the actual load path, a minimal proof of concept, compatibility limitations, and a concrete test result for safe cancellation. If no reliable path is established, stop this extension and retain the working selector. Do not expand into a general trainer, save editor, or full MGSHDFix fork.

## Phase 4: optional one-shot autoload extension

Only implement after Phase 3 establishes the mechanism. If native integration is needed, package a small `mgs3autoload.asi` through the manager's existing root-level ASI support and use the installed `wininet.dll` loader. Verify that the packaging rules accept it; mutable launch requests remain manager data, not immutable mod payloads.

Proposed contract:

- The selector chooses a stable save ID under the selected account. Resolve and hash its current bytes for that launch; do not permanently pin a profile to a save hash because normal saving changes it.
- Pass a per-launch request through a child-only environment variable, containing or identifying a versioned bounded request and a random nonce. An ordinary launch without that variable leaves the ASI inert. Keep communication minimal; add a manager-owned request file only if necessary.
- Bind the request to the supported build, selected account/slot, and current save identity. Reject path traversal and out-of-root/reparse targets. Revalidate before loading if files can change during startup.
- Wait for a verified load-ready game state, then invoke the proven operation once on the correct game thread. Mark the request consumed before any retry could trigger another load.
- Missing/changed saves, unsupported builds, timeout, or ambiguous state must cancel before invoking the load operation and leave normal game control available. If the native operation has already started, handle failure only through states proven in Phase 3; do not promise arbitrary recovery to Main Menu.
- Report distinct states: request accepted, waiting, loading, loaded, or canceled with reason. Report “loaded” only after a verified gameplay transition associated with the chosen slot.
- The extension does not write save payloads. It must coexist with QCamo and Crouch Walk without replacing their loader or relying on undocumented plugin load order.

Extend the selector only when this works. Use profile schema version 2 with `destination: save` and a separate stable `saveID`; do not overload the destination with a path or slot value. The version-2 reader accepts version-1 startup/menu profiles in memory and writes version 2 only on an explicit save. Older readers reject version 2 without overwriting it. Test migration, unsupported future versions, missing save IDs, and ordinary version-1 launch behavior. Test cold launch, repeated launch, ordinary launch without a request, wrong build, missing/moved/modified save, timeout, and coexistence with existing mods. Verify that no stale request is replayed and that disabling this ASI restores ordinary launching.

## Scope and stopping points

The recommended first implementation is Phases 0–2: native launch proof, saved profiles, and the small selector. Save autoload remains part of the intended outcome, with explicit research and implementation gates in Phases 3–4. It must not delay a usable launcher bypass or be described as complete before native behavior is proven.

Intro-logo removal, graphics fixes, resolution changes, executable patching, save editing, automatic cloud conflict resolution, and a graphical launcher are outside the initial scope.

Phases 0–2 now have source, tests, documentation, an example schema-1 preset, a built executable in `dist`, and the same verified executable installed at the game root. Two explicitly approved native trials were performed before final build verification. No save payload was rewritten by those trials. The previous root-level manager is retained as `dist/mgs3mod.previous.exe`.

## Sources and integration anchors

- The fingerprinted local launcher is the primary source for this installation. Inspect `Def::.cctor`, `DefManager::GetArguments`, and `BootGameSteam::CreateProcess` with the bundled Mono.Cecil assembly. These establish argument construction, not successful runtime behavior.
- MGSHDFix's `Init_LauncherConfigOverride` constructs the same argument family, including fixed `-selfregion EU` and `-launcherpath launcher.exe`. It also handles Steam app-ID creation and launcher reentry separately. Copying only the arguments does not prove identical lifecycle behavior. [Upstream source, inspected 2026-09-21](https://github.com/ShizCalev/MGSHDFix/blob/master/src/dllmain.cpp). This is a moving reference, not the compatibility fingerprint.
- Konami documents saves at the most recent continue point and resumption through Main Menu's LOAD GAME. [Official PC manual](https://metalgear.konami.net/manual/mc1/mgs3/pc/en/page07.html).
- `internal/cli/cli.go`: extend `parse`, `Help`, and dispatch. Add an input-aware entry point while preserving the existing `Run` wrapper for callers and tests; wire `os.Stdin` from `cmd/mgs3mod/main.go`.
- `internal/manager/init.go`: reuse `compatible` and `processes`; preserve `initSpec`, `checkInit`, and `finishInit` behavior.
- `internal/manager/filesystem.go`: reuse session locking and guarded root access. Verify replacement primitives before using them for preferences; do not assume journal writes and replaceable preferences have identical semantics.
- `internal/manager/manager.go`: preserve unresolved-transaction rejection. Do not route launching through the full asset inspection branches for `status`/`verify`.

## Verification record

Scope: this plan's current-state claims, native argument evidence, integration constraints, implementation artifacts, automated checks, and observed runtime behavior.

- Fingerprints in Evidence and confidence: CORRECT. `Get-FileHash -Algorithm SHA256 -LiteralPath 'METAL GEAR SOLID3.exe','launcher.exe','launcher_Data\Managed\Assembly-CSharp.dll'` returned the exact three hashes listed above.
- Repository baseline: CORRECT. `git -c safe.directory='C:/Games/METAL GEAR SOLID 3 - MCV/mgs3-mod-manager' -C mgs3-mod-manager status --short --branch` returned `## main...origin/main` before writing this file; the corresponding `rev-parse HEAD` returned the recorded commit. This does not establish current remote-server state.
- Installed state: CORRECT. `mgs3mod.exe status --json`, parsed through `ConvertFrom-Json`, returned `ok: true`, generation 15, enabled `asi-loader`, `qcamo`, and `crouch-walk`, and disabled `camo-test`.
- Native argument construction: CORRECT. The bundled Mono.Cecil inspection of the three methods identified above returned the listed argument strings and the independent EU launcher-region default. Upstream `Init_LauncherConfigOverride` corroborates the command family.
- Initialization inventory concern: CORRECT within initialization recovery. `Get-Content mgs3-mod-manager\internal\manager\init.go -TotalCount 75` shows the explicit inventory and rejection of unexpected entries in `finishInit`. This does not claim every normal command rejects arbitrary files in an already committed state directory.
- Direct native launch and existing-mod coexistence: CORRECT for the tested local route. The experiment log records the exact process arguments and artifact checks; the user reported collection-menu bypass, manual save loading, QCamo, and Crouch Walk working. The report is user evidence, while unchanged save hashes and manager verification are artifact evidence.
- Literal Main Menu arrival and chosen-save autoload: UNVERIFIABLE. The first visible game screen was not named, and Phases 3–4 have not been performed. The implementation rejects `menu` and save destinations.
- Built-manager non-dry launch: UNVERIFIABLE. `dist\mgs3mod.exe launch --profile na-startup --dry-run --json --game-root 'C:\Games\METAL GEAR SOLID 3 - MCV'` returned exit code 0 with the exact supported argument array and left the preference file absent, but dry-run deliberately does not start the game.
- Planning-only activity and no game launch: WRONG as a current claim. Source, tests, documentation, and a build artifact now exist, and two approved native trials were run. The stale claim was replaced above.
- Automated and build verification: CORRECT. `go test -count=1 ./...`, `go vet ./...`, formatting and diff checks, and `scripts\build.ps1 -GameRoot 'C:\Games\METAL GEAR SOLID 3 - MCV'` passed. The rebuilt `dist\mgs3mod.exe` SHA-256 is `4EA2850BD65942AF90511F8688616D1D56E5B42F7FBBE32556FA8ED75C91141E`.

Environment limitation: the default shell and a subsequent patch update failed with `helper_unknown_error: setup refresh had errors`. Read checks and the document correction succeeded through reviewed escalated PowerShell. Git required a command-scoped `safe.directory` setting because the sandbox and interactive accounts differ; no global Git setting was changed by those commands.
