# Delta controls alpha (v0.4.0-alpha.1)

An early test build of two ASI plugins for Metal Gear Solid 3: Master Collection (PC) that bring the game closer to MGS Delta's New Style controls:

- **fpv-move 0.8.4**: first-person walking, over-the-shoulder aiming with a crosshair and precise shots, strafing while aiming, move-while-aiming, a Delta-style keyboard and pad layout, crouch CQC, wall aim and corner peek, and D-pad shortcuts.
- **qcamo-face 1.0.4-face.4**: a fork of [QCamo](https://github.com/zexk/mgs3-qcamo) 1.0.4 (MIT) with a face paint list beside the camouflage list. It opens by holding D-pad up while the game keeps running, and a uniform change keeps your face paint. It replaces QCamo 1.0.4; do not enable both.

This is an **alpha**. It was built and played on one PC, with one controller and one save. Expect bugs, and expect some things not to work on your setup at all. The list of what has never been tested is in [delta-controls-untested.md](delta-controls-untested.md). Reports on those items are the most useful thing you can send.

## Requirements

- Metal Gear Solid 3 from Master Collection Vol. 1, PC, in the build this manager supports. The manager checks five game files before it changes anything. `doctor` tells you whether your copy matches. A Steam copy of the supported build is accepted since this release; **no Steam copy has been tried yet**, so please report either way.
- Windows 10 or 11, 64-bit, PowerShell 5.1 or later.
- No MGSHDFix. It hooks the same aiming code as fpv-move. Remove it first.
- No other ASI loader (`dinput8.dll`, `winmm.dll` and so on). The manager installs Ultimate ASI Loader 9.7.4 as `wininet.dll` and refuses to overwrite files it does not own.

## Install

1. Close the game and the Master Collection launcher.
2. Extract the kit zip to a folder **outside** the game folder, and open PowerShell in that folder.
3. Check the SHA-256 of each file against `SHA256SUMS.txt` (`Get-FileHash .\fpv-move-0.8.4.mgs3mod.zip`).
4. Set your game folder, then run `doctor`. Continue only if it prints no `error [...]` line (it should say `installation inspected`):

```powershell
$game = 'D:\SteamLibrary\steamapps\common\MGS3'   # your MGS3 folder, the one with "METAL GEAR SOLID3.exe"
.\mgs3mod.exe doctor --game-root $game
.\mgs3mod.exe init --game-root $game
```

5. Install the loader and both plugins:

```powershell
.\mgs3mod.exe add .\asi-loader-9.7.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable asi-loader --game-root $game
.\mgs3mod.exe add .\fpv-move-0.8.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable fpv-move --game-root $game
.\mgs3mod.exe add .\qcamo-face-1.0.4-face.4.mgs3mod.zip --game-root $game
.\mgs3mod.exe enable qcamo-face --game-root $game
.\mgs3mod.exe verify --game-root $game
```

If you already use QCamo through this manager, run `.\mgs3mod.exe disable qcamo --game-root $game` before enabling `qcamo-face`. If `doctor` refuses your copy, do not replace game files to make it pass; send the error line instead (see "Reporting").

6. Start the game as you normally do (from Steam, for a Steam copy). On the first load, `fpvmove.log` and `qcamo.log` appear in the game folder.

To uninstall: close the game, then `disable` and `remove` each package (`fpv-move`, `qcamo-face`, `asi-loader`) with the same `--game-root`. The manager restores what was there before.

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

1. How you own the game (Steam or another store) and whether `doctor` passed. If not, its `error [...]` line (for a different game file it names the file and says `core hash mismatch`).
2. Keyboard or controller, and which controller. On Steam, whether Steam Input is on for the game.
3. What you did, what you expected, what happened, and roughly when (the logs have times).
4. `fpvmove.log` and `qcamo.log` from the game folder. **Copy them before you start the game again**, because both are rewritten at each launch. If the game crashed, say so, and keep `qcamo-crash.dmp`. Only share the dump privately if asked; it can contain personal paths.
5. Other mods installed.

Most wanted: Steam copies, Steam Input on a controller, PlayStation controllers, sniper rifles and other scoped weapons, and any weapon or section of the game listed in [delta-controls-untested.md](delta-controls-untested.md).
