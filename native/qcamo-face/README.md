# qcamo-face: QCamo with a face paint list

A small fork of [QCamo](https://github.com/zexk/mgs3-qcamo) 1.0.4 (MIT) for Metal Gear Solid 3: Master Collection. It is kept as a patch against the pinned upstream source, built with MSVC, and shipped as the manager package `qcamo-face` 1.0.4-face.4. It is not the official QCamo binary.

## What changes

- **FACE PAINT list.** The menu header names both lists, CAMOUFLAGE and FACE PAINT; the one showing is lit and underlined. Left/Right switch lists (arrow keys, A/D, D-pad left/right, either stick across) with the cursor sound. The list shown last is kept until the game exits. The face paint list shows owned face paints with their icons, best camouflage value for the current ground first, with the cursor on row 0 when the menu opens.
- **Face paint only.** Equipping a face paint row changes the face paint alone, through upstream's face-only change path (the same uniform with a new face). It is refused, with the game's refusal sound, when that face paint is already worn, while Tuxedo is worn (Tuxedo takes no face paint), or when the change is not accepted (a change already running, or the gameplay gate closed). While Tuxedo is worn, the face paint rows other than the worn one are dimmed and show no gain.
- **Taking Mask off from the menu (1.0.4-face.4, played 2026-09-27).** With Mask (face paint 10) on, a face paint row or a uniform row can be equipped again. When the change leaves Mask (the face byte is already off 10), QCamo sends an extra `0x1A0014` right after `0x1A000F`, in the same tick, so the Mask node is hidden before the face asset loads on the next tick (`mask node hidden before the face load`); a uniform keeping Mask goes through face 0 first, so the node is hidden, then shown again by the final refresh. Reason: in face.1 and face.2 every change made with Mask on crashed at game RVA 0xC8187 (write to 0x444) in an actor update after the face load and before the final `0x1A0014`, whose handler sets the Mask node (0x74DDDA, at 0x36B7A6) from the face byte and clears its bit in the component at 0x383670, which then keeps it hidden on every update; the node likely kept pointing at a face resource the load replaced (not proved). face.3 refused every change with Mask on. Still refused with Mask on: every change while status 7 is set (taken here as first person; the name is not proved, and 0xBA is the first-person view flag), because that refresh keeps the node shown there (`leave first person to take the Mask off`; the gameplay thread checks again: `uniform change ignored: Mask worn in first person`), and the Tuxedo (`take the Mask off before the Tuxedo`, never tried). If this crashes, go back to face.3.
- **Live menu from D-pad up (1.0.4-face.3).** A menu opened by holding D-pad up leaves the game running: the wheel pause is taken only while a change is in flight (from the equip being queued to `change complete`) and released after it. The log says `menu live (no pause)` on open and `change pause on` / `change pause off` around each change. Menus opened with G or L1+Y pause the game as upstream (the keyboard selection keys also move Snake); pressing L1+Y in a live menu hands it to L1 and pauses it. The gate still closes a live menu when anything else pauses the game or a cutscene starts; since the game runs, the wheel pause bit counts as ours only while QCamo holds it, so a game item or weapon window opening under a live menu closes it too. fpv-move is expected to hide A, the other D-pad directions and the right stick from the game while D-pad up is held; the left stick still moves Snake.
- **Uniforms keep the face paint.** Equipping a uniform keeps the face paint Snake has on instead of upstream's automatic best face paint. Tuxedo still forces no face paint. Uniform rows read `UNIFORM / FACE PAINT` with the kept face paint and are still ranked by uniform plus that face paint.
- **D-pad up hold.** Holding D-pad up opens the menu and releasing it closes it, like holding G. The press that opens it never moves the cursor: while the menu was opened this way D-pad up does not select. L1 alone does not keep a D-pad-up menu open (L1 is held in play for other things); pressing L1+Y in it hands the menu to L1 as upstream, after which L1 alone keeps it up and D-pad up selects again. L1+Y and G still work as upstream. The gameplay gate is unchanged: gameplay only, backpack recovered, Survival Viewer closed, no other pause.
- **Right stick.** While the menu is open, however it was opened, the right stick works like the left one: up/down moves the cursor, left/right switches the list. With the left thumb on D-pad up, the right thumb selects with the right stick and equips with A; releasing D-pad up closes the menu. The right stick is read through XInput and, when it resolves, Steam Input's `ingame_stick_cam_dir` (the executable's camera stick action, beside `ingame_stick_move`); if it does not, the log says `steam input right stick not found; XInput only`.
- **Stick latch.** Each stick latches one direction at a time. A direction engages when the stronger axis passes the threshold (half travel: 0.5 for Steam, 16384 for XInput) at 1.5 times the other axis, and lets go only once that axis falls below 60% of the threshold (0.3, about 9830). A stick resting near the threshold or near a diagonal therefore cannot step the cursor or switch list repeatedly, and a clean diagonal does nothing.
- **Keys held at open wait.** Movement keys held when the menu opens (W/A/S/D, arrows, Enter) act only once pressed again, as pad buttons already did.
- **DESERT.** Face paint 4 is labelled DESERT instead of MOUNTAIN, as the game's own item table names it (item 0x4E, "FACE/DESERT" / "DESERT").
- **Unchanged.** The camouflage change protocol in `qcamo.cpp` apart from leaving Mask, the crash log and the executable gate. The `qcamo.cpp` edits: in the pause, `set_menu_pause` records the bit as ours and logs `change pause` for live menus, and the worker loop leaves a live menu unpaused; for leaving Mask (face.4), `hide_mask_node` sends the early `0x1A0014`, and `change_camo` refuses a change leaving Mask while status 7 is set, with the refusal sound.

Log lines in `qcamo.log` that are new or changed:

- `menu opened by keyboard|pad chord|D-pad up`, and `menu closed by <the input let go>`.
- `menu live (no pause)` on a D-pad up open, then `change pause on` / `change pause off` around each change it makes (paused menus keep `wheel pause set` / `cleared`).
- `menu switched to face paint|camouflage`, `menu selected face <id>`, `face <id> <NAME> equip accepted`, `face <id> <NAME> equip refused: <reason>`, `uniform <id> equip refused: <reason>`, `mask node hidden before the face load`.
- For each D-pad action, `steam input Arrow Up: handle <n> by display name` (or `by ingame_cmn_move_up` when the fallback found it), or `steam input Arrow Up not found`; likewise Down, Left and Right.
- `steam input right stick not found; XInput only`. The Steam `pad ready` line also prints the right stick handle.

## Controls

| | Keyboard | Pad |
| --- | --- | --- |
| Open (hold) | G | D-pad up, or L1+Y (then L1 alone) |
| Select | W/S, Up/Down | D-pad up/down, left or right stick up/down (D-pad up excluded when it opened the menu) |
| Switch list | A/D, Left/Right | D-pad left/right, left or right stick left/right |
| Equip | Enter | A (cross) |

## Build

Needs git, Visual Studio 2022 with the C++ x64 tools (found through vswhere) and network access for the first fetch. From the repository root:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\native\qcamo-face\build.ps1
..\mgs3mod.exe pack work\qcamo-face-pkg --out work\qcamo-face-1.0.4-face.4.zip
```

`build.ps1` fetches QCamo `0a8bee4874e6ed9773c96bd48e899056c4c67801`, Dear ImGui `f1cc2ae15e53a861a874c3034aae6798fde194ab` and MinHook `c3fcafdc10146beb5919319d0683e44e3c30d537` into `work\qcamo-face-src\` (a shallow fetch of each pinned commit, verified with `git rev-parse`), resets them to those commits, applies `qcamo-face.patch` with `git apply`, and builds with `cl`. It writes:

- `work\qcamo-face-build\qcamo.asi`
- `work\qcamo-face-pkg\`: `manifest.json` (schema 2) and `payload\qcamo.asi`, ready for `pack`
- `work\qcamo-face-notices\`: `LICENSE.txt` (upstream's MIT notice, then the Dear ImGui and MinHook notices for code linked into the ASI) and `UPSTREAM.md` (provenance, toolchain, hashes)

The notices sit beside the package because a package may hold only `manifest.json` and its mapped `payload/` files (`docs/runtime-plugins.md`); `pack` rejects anything else. Ship them with the package zip.

`build.ps1 -Upstream` builds the unpatched pinned source the same way into `work\qcamo-face-build-upstream\`, to separate toolchain problems from patch problems. The build uses `/Brepro`, so a rebuild with the same toolchain gives the same bytes.

To change the fork, work in a separate checkout of the pinned commit, never in `work\qcamo-face-src\qcamo`: every run of `build.ps1` resets that one with `git checkout -f` and `git clean -fdx`, discarding edits there. In the separate checkout, apply the current patch, make the change, and regenerate it with `git diff > qcamo-face.patch`.

## Install (replaces qcamo)

Enable `qcamo-face` only together with fpv-move 0.8.0 or later. Earlier fpv-move sends D-pad up to the game as R3 (the camera view), so every D-pad up hold would also toggle the camera; 0.8.0 hides D-pad up from the game.

`qcamo-face` installs the same file, `qcamo.asi`, as the `qcamo` 1.0.4 package, so only one of them can be enabled: `enable` refuses a target owned by another enabled package. Disable and remove qcamo 1.0.4 first, with the game and launcher closed:

```powershell
.\mgs3mod.exe disable qcamo
.\mgs3mod.exe remove qcamo
.\mgs3mod.exe add .\mgs3-mod-manager\work\qcamo-face-1.0.4-face.4.zip
.\mgs3mod.exe enable qcamo-face --dry-run
.\mgs3mod.exe enable qcamo-face
```

To replace an earlier `qcamo-face` build, disable and remove `qcamo-face` first: `add` refuses an ID that already stores different content.

To go back, disable and remove `qcamo-face`, then add and enable the official `qcamo` package again. The log and crash dump names are upstream's (`qcamo.log`, `qcamo-crash.dmp`).

## Provenance and limits

- Forked from upstream commit `0a8bee4874e6ed9773c96bd48e899056c4c67801` (tag v1.0.4). MIT; the upstream notice is unchanged in `LICENSE.txt`. The build is not byte-identical to the official 4,063,819-byte MinGW release and makes no claim to be it.
- Built with MSVC `cl` instead of upstream's MinGW/CMake. Nothing here was run in the game at build time.
- The Steam Input display names "Arrow Left" and "Arrow Right" are assumed from upstream's "Arrow Up"/"Arrow Down". When the walk misses any arrow, the executable's own action name (`ingame_cmn_move_up`, `_down`, `_left`, `_right`, in its `NHTCommonSet` action set) is looked up with `GetDigitalActionHandle` instead; that those are the D-pad actions is inferred from the names, not checked. The log says which source each arrow came from. Whether `ingame_stick_cam_dir` resolves, and which way its Y axis points under the player's Steam camera settings, is likewise unchecked; the XInput right stick does not depend on it.
- Pressing D-pad up without the backpack plays the refusal sound, as L1+Y and G do.
