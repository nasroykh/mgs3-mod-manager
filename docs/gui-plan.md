# Plan: a minimal window app for the manager (v0.4.0-alpha.2)

Date: 2026-09-27. Status: done. Steps 1 to 9 plus the UI pass, quick launch and custom launch profiles (sections below), each reviewed and tested by the user on the real install; released as the prerelease v0.4.0-alpha.2 on 2026-09-27 ([record](releases/v0.4.0-alpha.2-validation.md)). Goal: players who never used a terminal can install, check, and remove the Delta controls kit with a few clicks, with every safety check the CLI has. Keep it simple and minimal; improve after tester feedback.

## Decisions (made with the user on 2026-09-27)

- **Wails v2**, stable line (v2.14.0, 2026-08-10, MIT). Not v3: still beta (beta.26 on 2026-09-25, a new beta every few days). Wails v2 on Windows uses the system WebView2 runtime and needs no C toolchain. This machine has Go 1.27.1, Node 24.19.0, npm 11.17.0 and WebView2 153; CI is `windows-latest` with `actions/setup-go`.
- **A separate executable, `mgs3mod-gui.exe`.** The CLI (`mgs3mod.exe`) and its release zip stay unchanged: `install.ps1` requires that zip to hold exactly `mgs3mod.exe`, `README.md` and `THIRD-PARTY-NOTICES.txt` (`install.ps1:39-48`), and the workflow checks the asset set (`.github/workflows/release.yml`, "Publish complete release").
- **The GUI calls the manager's Go code directly**, never the CLI as a child process: `manager.Production().WithRoot(root)` then `Run(command, arg, manager.Options{})` (`internal/manager/manager.go:59`, `root.go:7`, `types.go:51-102`). The fingerprint check, process guard (game or launcher running), locks, journal, backups and recovery all apply unchanged. Errors are `*manager.Error{Code, Message, Paths}`.
- **Frontend: plain HTML, CSS and JavaScript, no npm packages, no framework, no bundler.** Static files embedded with `//go:embed`. In `wails.json` leave `frontend:install` and `frontend:build` empty. No supply chain to audit on the JS side.
- **No UPX** (raises antivirus false positives). **No code signing** for the alpha: users will see SmartScreen ("More info", then "Run anyway"); the guide says so. **WebView2 missing:** use Wails' download strategy (`wails build -webview2 download`), so an old Windows 10 PC is offered the runtime.
- **Not in the first version:** launching the game (direct launch was never tried on a Steam copy; start the game from Steam; the user asked for it on 2026-09-27, see "Quick launch" in the UI pass), texture mods UI, crouch-walk import, recovery with `--restore-missing`, launch profiles. The CLI keeps all of these.
- **Release as v0.4.0-alpha.2 (prerelease).** v0.4.0-alpha.1 is public already but has not been posted to Nexus, Reddit or X; the user posts after alpha.2.

## Layout

- Go: keep one module (`mgs3mod`) unless step 1 shows a problem. `cmd/mgs3mod-gui/` holds `main.go` (Wails setup, window 900x640, title "MGS3 Mod Manager"), the embedded `frontend/` and `wails.json`. The logic lives in a testable package `internal/gui/` (no Wails import): game folder detection, the one-click flows, and turning manager errors into plain text. `cmd/mgs3mod-gui` only binds `internal/gui` methods to the window.
  - Alternative if Wails' dependencies must stay out of the CLI's `go.mod`: a nested module `gui/` with `replace mgs3mod => ../`. The CLI binary never links Wails either way; `THIRD-PARTY-NOTICES.txt` must stay accurate for the CLI zip.
- Packages the buttons install: the GUI looks for `*.mgs3mod.zip` beside its own executable (the kit folder). It does not embed them.

## Screens (one window)

1. **Game folder.** On start, detect it: read `HKCU\Software\Valve\Steam` `SteamPath`, parse `<SteamPath>\steamapps\libraryfolders.vdf` for every library path, and look for `steamapps\common\*\METAL GEAR SOLID3.exe` (the folder name is not assumed). If several match, list them. Always offer **Browse…** (folder picker, `runtime.OpenDirectoryDialog`). Remember the last folder in `%LOCALAPPDATA%\mgs3mod\gui.json` (never inside the game folder). Then run `doctor` and show one line: "Supported game version" (green) or "Not supported: <file> differs" / "Close the game and launcher first" (red), with **Copy details** for reports.
2. **Delta controls.** One status line ("Installed: fpv-move 0.8.4, QCamo face paint 1.0.4-face.4" / "Not installed" / "Older version installed") and two buttons:
   - **Install / Update**: `init` if not initialized; `add` + `enable` `asi-loader` if missing; if `qcamo` is enabled, `disable qcamo` (keep it in the library); for `fpv-move` and `qcamo-face`: if a different version is stored, `disable` + `remove` first, then `add` the package beside the exe, `enable --dry-run`, `enable`; finish with `verify`. Stop at the first error and show which step failed; everything done so far stays consistent because each manager command is its own transaction.
   - **Uninstall**: `disable` + `remove` `qcamo-face` and `fpv-move`; then `asi-loader` only if no other enabled package needs it (the manager refuses otherwise; show that message).
3. **Mods (advanced, collapsed by default).** A list from `status` (id, version, enabled) with an on/off switch per mod (`enable` / `disable`), **Add mod package…** (file picker for `.zip`, then `add`), **Verify**, and **Open game folder** (for `fpvmove.log`, `qcamo.log`).

Every action: buttons disabled while it runs (a goroutine; the manager's own lock also protects against a CLI running at the same time), then the list and status refresh. Confirmation dialog before Uninstall. Wording from the manager's error codes: grep `fail(` and `wrap(` in `internal/manager` for the codes and map the common ones (3 fingerprints/compatibility, process running, conflicts) to one plain sentence each, keeping the original message under "Details".

## Steps

1. **Scaffold.** `go install github.com/wailsapp/wails/v2/cmd/wails@v2.14.0`, `wails doctor`. Create `cmd/mgs3mod-gui` from the `vanilla` template, then strip the template's npm/Vite setup to plain static files. Build once with `wails build -clean -webview2 download -trimpath` and open it. Check that `go build ./...`, `go vet ./...` and the existing tests still pass and that the CLI binary did not change in size because of Wails (`scripts/build.ps1`).
2. **`internal/gui` with tests.** Steam library detection (a small VDF parser for `libraryfolders.vdf` with test fixtures, including paths with escaped backslashes and several libraries), the Install/Update and Uninstall flows against `manager.New(manager.Config{...})` test fixtures as in `internal/manager/manager_test.go` (fake core file, `CheckProcesses` stub), error text mapping, the settings file.
3. **Frontend.** The three sections above, keyboard accessible, readable at 100% and 150% scaling, light and dark following Windows. No external fonts or CDN (offline app).
4. **Build and release plumbing.**
   - `scripts/build.ps1`: also build the GUI (pinned Wails CLI), write `dist/mgs3mod-gui.exe` and print its SHA-256.
   - `scripts/package-release.ps1`: add the asset `mgs3mod-gui_<version>_windows_amd64.zip` with `mgs3mod-gui.exe`, `README.md`, and a GUI notices file listing Wails (MIT) and every Go module it links (from `go version -m dist/mgs3mod-gui.exe`, with each license text).
   - `.github/workflows/release.yml`: install the pinned Wails CLI; expected assets 7, checksums 6.
   - `scripts/package-delta-alpha.ps1`: take `-GuiZip`, verify it against the draft's `checksums.txt`, put `mgs3mod-gui.exe` at the kit root.
   - `scripts/test-installer.ps1` stays valid (CLI zip unchanged).
5. **Docs.** Rewrite `docs/delta-controls-alpha.md` GUI-first (extract, double-click `mgs3mod-gui.exe`, SmartScreen note, Install; CLI steps move to an "Advanced" section). README top section, `docs/releases/v0.4.0-alpha.2.md`, `docs/releasing.md` (GUI asset), `docs/delta-controls-untested.md` (GUI untested items).
6. **Review.** One independent reviewer agent (Opus, high effort) on the whole diff before any commit that changes release plumbing; fix and re-verify.
7. **Local test.** On a scratch copy of the game folder with the Steam-wrapped executable of the supported build (plus `Engine.dll`, `Renderer.dll`, `launcher.exe`, `launcher_Data/Managed/Assembly-CSharp.dll`), run Install, the Mods list and Uninstall through `internal/gui` and through the built exe.
8. **User test (one session, game and launcher closed).** On the real install: open the GUI, check that it finds the folder (this install is not in a Steam library, so Browse), `doctor` green, Uninstall (everything removed), Install (everything back, `verify` ok), then launch and play a minute to confirm fpv-move and the fork load (`fpvmove.log`, `qcamo.log`). Also try the Mods switches once. Note: the installed state before the test is generation 164 with fpv-move 0.8.4 and qcamo-face 1.0.4-face.4 enabled; crouch-walk 0.2.1 enabled, `qcamo` and `camo-test` disabled; the GUI must leave crouch-walk alone.
9. **Release v0.4.0-alpha.2** exactly like alpha.1 (`docs/releasing.md`, "Prereleases"): commit, push `main`, tag, wait for the draft, download the manager zip, loader package, GUI zip and `checksums.txt` with `gh`, build the kit, upload the kit and `delta-controls-checksums.txt`, re-download and compare, publish with `--prerelease --latest=false`, run the installer check, write `docs/releases/v0.4.0-alpha.2-validation.md`.

## Rules that carry over

- Public text never describes the local development setup; the pad path is described as "without Valve Steam Input". The detailed history stays in the local branches `local/delta-full-history` and `delta-controls`, which are never pushed.
- Commit messages and release text read as the user's own: no AI attribution, no co-author lines.
- Close the game and launcher before any `add`, `enable`, `disable`, `remove`. A pushed tag is never moved; a failed release gets a new tag.
- `gh` is installed (`C:\Program Files\GitHub CLI\gh.exe`, logged in as the repository owner). Git in this repo needs `-c safe.directory='*'`; Go builds need `GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=safe.directory GIT_CONFIG_VALUE_0='*'`. No Python on this machine; Node is available.

## UI pass before the release (decided with the user on 2026-09-27)

After the first version passed the user's test, the user asked for a stronger UI and UX before releasing: loading spinners, the steps shown live in the middle of the window over a backdrop that blocks everything else while a change runs, and Tabler icons. The user proposed shadcn/ui (React, Tailwind CSS, Base UI); the decision is to stay on plain HTML, CSS and JavaScript for alpha.2 and to make a React/shadcn version later with a larger UI. The release waits for this pass. The rules above still hold (no npm packages in the build, no CDN, no external fonts).

### Go side

- **Plan first, then run.** Install and Uninstall compute their whole step list from the stored state and the kit before changing anything (pure functions `planInstall` and `planUninstall`, unit tested), so the window shows every step from the start and ticks them off.
- **Progress events.** The service takes a progress callback; each action reports `plan` (the step list), then `running`, `done`, `skipped` (nothing to do) or `failed` per step, then `finished` with the outcome. `cmd/mgs3mod-gui` forwards them with Wails `EventsEmit`. Single mod actions (switch, add, verify) report the same way with one or two steps.
- **Uninstall preview.** `PlanUninstall(root)` returns the steps without running them; the confirmation dialog lists them (what is removed, whether QCamo comes back on, whether the ASI loader stays and why).
- **Close guard.** While an action runs, closing the window is refused (Wails `OnBeforeClose`) and the dialog says why: a manager command is a transaction and should finish.

### Window

- **Layout.** A header with the app name, version and a re-check button; three cards: Game folder, Delta controls, Mods. Tabler outline icons (MIT, pinned version, inlined as an SVG sprite, license in `THIRD-PARTY-NOTICES-GUI.txt`) on buttons and status lines.
- **Status lines** as badges with an icon: checking (spinner), supported (check), not supported / game running (alert), not set up. The folder path is shown in full with a tooltip, never cut in the middle of a word.
- **Activity dialog** for every change: a centred dialog over a dimmed, blurred backdrop; the page behind is `inert` (no clicks, no keyboard focus). It shows the action title, a spinner, the planned steps with a state icon each (waiting, running, done, skipped, failed), the elapsed time, and at the end a success or failure banner with the plain-language summary, **Details** (the manager's message) with **Copy**, and **Close**. Esc and the backdrop do nothing while running.
- **Confirmation dialog** (in the page, not a native message box) for Uninstall, listing the preview steps; **Uninstall** is the danger button, **Cancel** has the focus.
- **Mods card**: a table with real toggle switches, the mod name under its ID, a badge for the Delta controls packages; an empty state with an icon. A switch runs through the activity dialog; a successful short action closes it by itself after a moment and shows a small toast.
- **First run / no folder**: an empty state in the Game folder card pointing at **Browse…**; the Delta controls and Mods cards stay disabled with a note.
- **Accessibility**: keyboard-only use works (focus moves into dialogs and back), visible focus rings, `aria-live` for status and progress, `role="dialog"` with `aria-modal`, reduced motion replaces spinners with a static icon, `forced-colors` keeps borders visible. Light and dark follow Windows.

### Quick launch (asked by the user on 2026-09-27)

A **Play** card starts the game the way `mgs3mod launch` does, skipping the Master Collection selector, through the manager's own `Launch` (fingerprints, initialized state, no unresolved change, game and launcher not running, the lock held until the started process is observed as the game).

- (Superseded by "Custom launch profiles" below.) It shows the launch profiles from `mgs3mod-launch.json` (or the built-in `na-startup` when none is saved), the default one selected, each described in words: region, language, button prompts, destination. Only the combinations the manager accepts can be launched; today that is North America, English, keyboard prompts, game startup (`launch.Selection.Validate`). Other choices are not offered as editable options until they are tested; the card says so.
- **Launch game** runs through the activity dialog: "Check the game folder and launch settings" (the manager's dry run), then "Start the game", then "Game started" with the process ID; the dialog closes by itself.
- Disabled, with the reason, when the manager is not set up in the folder (Install first) or the doctor line is red.
- The executable fingerprint tells a Steam copy apart (the SteamStub-wrapped file). Direct launch was never tried on a Steam copy, so the card says so there and suggests starting from Steam; launching stays possible.
- No save autoload and no main-menu destination (not implemented in the manager).
- Go tests with the manager's launch test seams are not reachable from `internal/gui` (they are private to `internal/manager`); `internal/gui` tests the profile list, the descriptions and the disabled reasons, and the launch path is checked by the dry run on the scratch copy and by the user.

### Custom launch profiles (asked by the user on 2026-09-27)

Players must not be forced into one profile (Europe, other languages, a controller's button prompts). Decided with the user: allow every combination the official launcher itself builds, marked "Not tested yet", in alpha.2.

- **Evidence (launcher `Assembly-CSharp.dll`, read with its bundled Cecil on 2026-09-27).** The language lists in `GameLanguageSelectMGS3::FrameUpdate` offer North America with English, French or Spanish and Europe with English, French, Italian, German or Spanish; Japan (Japanese only) is offered only when a Steam download is installed (`RegionSelectMGS3::FrameUpdate`, `DefManager::SetDLInfo`), so the manager leaves it out until it can check that. (`CheckRegionLanguage` alone lets French and Spanish through for every region; it is not the rule.) `Def::regionLauncher` is set to EU in `Def::.cctor` and never written elsewhere, so `-selfregion EU` is what this launcher build always passes. `-ctrltype` comes from `CommonUIManager.ControllerType` (its controller setting, for example the saved key config) and is one of XBOX, PS4, PS5, NX, KBD.
- **Manager (`internal/launch`).** `Selection.Validate` accepts exactly those region/language pairs with any of the five button prompt types and destination `startup`; `Tested` reports the combinations played (North America, English, keyboard; since 2026-09-27 also Europe, French, Xbox). `BuildArguments` passes the selected button prompt type (before, only keyboard prompts were accepted, so `KBD` was always sent). The argument order stays the launcher's. The CLI selector and `--region/--language/--controller` get the same rule. `internal/manager` gains `PutLaunchProfile` (create, change or rename, and optionally make default, in one atomic update) and `RemoveLaunchProfile` (the last profile cannot be removed; removing the default moves it to the first remaining profile).
- **Window.** The Play card gets **New profile…** and **Edit…**; the profile dialog also deletes a profile (a second click confirms) and makes it the default. The profile dialog has a name (turned into a valid profile ID), region, language (only the region's languages) and button prompts, and says "Not tested yet" for every combination except the tested one. Profiles are saved in `mgs3mod-launch.json` in the game folder through the manager, as the CLI does.
- **Docs.** `docs/direct-launch.md`, the tester guide and the untested list say which combination is tested and ask testers to report the others.

### Checks

- Go unit tests for `planInstall`, `planUninstall` and the event order (plan, running/done per step, failed stops, finished last).
- A browser harness (a scratch HTML page that stubs the Wails bridge with fake data and delays) rendered in headless Microsoft Edge: screenshots of each state (start, checking, not supported, installed, confirm, running, failed, done, toggle toast, no folder) in light and dark at 100% and 150%.
- The built exe driven through UI Automation on the scratch copy of the game folder (Install, switch, Verify, Uninstall with the in-page confirmation), with screenshots.
- One independent reviewer on the diff; then the user tests the new build on the real install before step 9.
