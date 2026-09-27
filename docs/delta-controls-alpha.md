# Delta controls alpha (v0.4.0-alpha.2)

An early test build of two ASI plugins for Metal Gear Solid 3: Master Collection (PC) that bring the game closer to MGS Delta's New Style controls:

- **fpv-move 0.8.4**: first-person walking, over-the-shoulder aiming with a crosshair and precise shots, strafing while aiming, move-while-aiming, a Delta-style keyboard and pad layout, crouch CQC, wall aim and corner peek, and D-pad shortcuts.
- **qcamo-face 1.0.4-face.4**: a fork of [QCamo](https://github.com/zexk/mgs3-qcamo) 1.0.4 (MIT) with a face paint list beside the camouflage list. It opens by holding D-pad up while the game keeps running, and a uniform change keeps your face paint. It replaces QCamo 1.0.4; do not enable both.

This is an **alpha**. It was built and played on one PC, with one controller and one save. Expect bugs, and expect some things not to work on your setup at all. The list of what has never been tested is in [delta-controls-untested.md](delta-controls-untested.md). Reports on those items are the most useful thing you can send.

## Requirements

- Metal Gear Solid 3 from Master Collection Vol. 1, PC, in the build this manager supports. The manager checks five game files before it changes anything. A Steam copy of the supported build is accepted; **no Steam copy has been tried yet**, so please report either way.
- Windows 10 or 11, 64-bit, with the Microsoft Edge WebView2 runtime (Windows 11 has it; on Windows 10 the app offers to download it the first time).
- No MGSHDFix. It hooks the same aiming code as fpv-move. Remove it first.
- No other ASI loader (`dinput8.dll`, `winmm.dll` and so on). The manager installs Ultimate ASI Loader 9.7.4 as `wininet.dll` and refuses to overwrite files it does not own.

## Install

1. Close the game and the Master Collection launcher.
2. Download `mgs3-delta-controls-0.4.0-alpha.2-kit.zip` and `delta-controls-checksums.txt` from the release. Optional: check the kit's SHA-256 against `delta-controls-checksums.txt` (`Get-FileHash .\mgs3-delta-controls-0.4.0-alpha.2-kit.zip` in PowerShell).
3. Extract the whole kit to a folder **outside** the game folder, for example `Downloads\mgs3-delta-controls`. Keep the files together: the app installs the packages that sit beside it.
4. Double-click **`mgs3mod-gui.exe`**. The app is not code-signed, so Windows may show "Windows protected your PC": click **More info**, then **Run anyway**. Do not turn off your antivirus.
5. **Game folder.** The app looks for the game in your Steam libraries. If it shows nothing, or the wrong folder, click **Browse…** and pick the folder that holds `METAL GEAR SOLID3.exe`. The app remembers it.
6. Wait for the line under the folder. **"Supported game version"** (green) means you can continue. A red line says why not, for example that a game file differs or that the game is still running. If a game file differs, do not replace game files to make it pass: click **Copy details** and send that text (see "Reporting").
7. Click **Install / Update**. A window in the middle lists every step and ticks them off as they run; keep the app open until it says "Delta controls installed" (the app refuses to close while a step runs). If a step fails, it stops there and says why; the steps before it stay done and consistent, and you can click **Install / Update** again once the cause is fixed.
8. Start the game as you normally do (from Steam, for a Steam copy), or with **Launch game** in the app (see "Play"). On the first load, `fpvmove.log` and `qcamo.log` appear in the game folder.

If you use QCamo 1.0.4 through this manager, Install turns it off (it stays stored) because the face paint fork replaces it, and Uninstall turns it back on. Other mods you installed with the manager are left alone.

**Update** to a later kit: close the game, extract the new kit to its own folder, run its `mgs3mod-gui.exe` and click **Install / Update**. It replaces the older plugins.

**Uninstall:** close the game, open the app and click **Uninstall**. The app first lists what it will do; click **Uninstall** again to go ahead. It removes fpv-move and the face paint fork and puts the original files back, and turns QCamo 1.0.4 back on if Install turned it off. It also removes the ASI loader, unless another mod that is turned on still needs it (the app says so).

### Play (quick launch)

**Launch game** starts the game directly, skipping the Master Collection menus, with the settings of a launch profile (you then load your save as usual). The built-in profile is North America, English, keyboard button prompts. **New profile…** makes your own: game region (North America or Europe), language (the ones the launcher offers for that region) and button prompts (keyboard, Xbox, PlayStation 4, PlayStation 5, Nintendo Switch); **Edit…** changes, renames or deletes one, and one profile is the default. Japan is not offered yet (the launcher shows it only with a Steam download installed). Two combinations have been played: the built-in one and Europe, French, Xbox button prompts; every other one uses the launcher's own settings but is marked "Not tested yet": please report whether it works. It needs Install done first and a green game folder line. **No Steam copy has been tried with quick launch**: on a Steam copy the app says so; if the game does not start that way, start it from Steam and please report it.

### Mods

The collapsed **Mods** section (advanced) lists every mod stored for this game folder, with a switch to turn each one on or off, **Add mod package…** (stores a `.mgs3mod.zip`, turned off), **Verify** (checks that every managed file still matches) and **Open game folder** (for the logs). Turning on two mods that change the same file is refused with the name of the mod to turn off first.

### Advanced: the command line

The kit also holds `mgs3mod.exe`, the command-line manager the app is built on. It does the same thing, and more (texture mods, recovery after an interrupted change). Open PowerShell in the kit folder:

```powershell
$game = 'D:\SteamLibrary\steamapps\common\MGS3'   # your MGS3 folder, the one with "METAL GEAR SOLID3.exe"
.\mgs3mod.exe doctor --game-root $game
.\mgs3mod.exe init --game-root $game
.\mgs3mod.exe add .\asi-loader-9.7.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable asi-loader --game-root $game
.\mgs3mod.exe add .\fpv-move-0.8.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable fpv-move --game-root $game
.\mgs3mod.exe add .\qcamo-face-1.0.4-face.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable qcamo-face --game-root $game
.\mgs3mod.exe verify --game-root $game
```

Continue after `doctor` only if it prints no `error [...]` line. If you already use QCamo through this manager, run `.\mgs3mod.exe disable qcamo --game-root $game` before enabling `qcamo-face`. To uninstall, `disable` and `remove` each package (`qcamo-face`, `fpv-move`, `asi-loader`) with the same `--game-root`. If the app reports an interrupted change, run `.\mgs3mod.exe recover --game-root $game`. The individual files can be checked against `SHA256SUMS.txt` in the kit.

## Controls

Keyboard and mouse:

| Action | Key |
| --- | --- |
| Aim over the shoulder (Snake keeps his facing and strafes) | Right mouse, hold |
| Fire (while aiming) | Left mouse |
| CQC (not aiming, with a CQC weapon) | Left mouse; hold to grab |
| Knife slash | Right mouse + left mouse |
| Crouch | C |
| Roll (hold for prone) | Space |
| Swap shoulder while aiming | Middle mouse |
| First person (native hold) | F |
| Iron sight in first person (native hold) | M |
| Item / weapon window (native hold) | 1 / 2 |
| Codec, Survival Viewer | Esc, Tab |
| QCamo menu (hold) | G |
| Auto-target on/off while aiming | F8 |
| Shoulder camera on/off | F7 |
| Crouched sideways step animation on/off | F6 |

Controller (tested with an Xbox pad):

| Action | Button |
| --- | --- |
| Aim over the shoulder | LT, hold |
| Fire (while aiming) | RT |
| CQC (not aiming) | RT; hold to grab |
| Aim the crosshair | Right stick while aiming |
| Swap shoulder | Right stick click while holding LT |
| First person on/off | RB or left stick click (in first person, LT brings the iron sight) |
| Lock-on, camera reset; iron sight in first person | LB, hold |
| Crouch / roll | A: tap to crouch, hold while running to roll |
| Item window / quick item toggle | D-pad left: hold / tap |
| Weapon window / quick weapon toggle | D-pad right: hold / tap (right stick also browses the windows) |
| Codec | D-pad down |
| QCamo menu, game keeps running | D-pad up, hold; right stick selects, left/right switches CAMOUFLAGE and FACE PAINT, A equips, release to close |
| Corner peek at a wall | D-pad left / right |

Notes:

- Auto-target is off while you aim, so shots go where the crosshair is. It stays on for Very Easy and Easy saves, and F8 toggles it.
- Crouching and pressing fire near a guard makes Snake stand and grab or punch (crouched hip fire is replaced by this).
- Mask face paint: it can be put on and taken off from the menu in third person. In first person, or for the Tuxedo, take it off first; the menu dims those rows.

## Reporting

Please post on the [issue tracker](https://github.com/nasroykh/mgs3-mod-manager/issues) (or wherever you found this build) with:

1. How you own the game (Steam or another store) and whether the app said "Supported game version". If not, the text from **Copy details** (for a different game file it names the file and says `core hash mismatch`).
2. Keyboard or controller, and which controller. On Steam, whether Steam Input is on for the game.
3. What you did, what you expected, what happened, and roughly when (the logs have times).
4. `fpvmove.log` and `qcamo.log` from the game folder (**Mods**, **Open game folder**). **Copy them before you start the game again**, because both are rewritten at each launch. If the game crashed, say so, and keep `qcamo-crash.dmp`. Only share the dump privately if asked; it can contain personal paths.
5. Other mods installed.

Most wanted: Steam copies, Steam Input on a controller, PlayStation controllers, sniper rifles and other scoped weapons, and any weapon or section of the game listed in [delta-controls-untested.md](delta-controls-untested.md).
