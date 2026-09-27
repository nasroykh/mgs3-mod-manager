# MGS3 Delta New Style controls

Date: 2026-09-27. Status: fpv-move 0.8.4 and the QCamo face-paint fork `qcamo-face` 1.0.4-face.4 are installed (generation 164); both passed their last play on 2026-09-27 (Mask off from the menu, face paints and uniforms with Mask on). fpv-move 0.8.4 with face.3 passed on 2026-09-26 (the right stick in the windows, the live D-pad-up menu holding Snake still). 0.8.0 (the backlog batch) and 0.8.1 were played on the pad on 2026-09-26; 0.8.1 and 0.8.2 carry the fixes and requests from those plays (see "0.8.0", "Still open after 0.8.0", "0.8.1", "0.8.2"). Before that, fpv-move 0.6.9 passed a full regression pass in play on 2026-09-26: an over-the-shoulder aiming camera with precise shots on the crosshair (one aim rig for camera and shots, 0.6.4 to 0.6.9), eased in and out, shoulder swap on middle mouse or pad right stick click with LT. Before that, 0.6.1 passed: shoulder aim on lock-on with the mouse turning the aim (0.5.9), auto-target off while aiming (F8 or a Very Easy/Easy save turns it on), crouched strafe, the pad layout (LT aim, RT fire), a crosshair on the aim point (0.6.0), and manual vertical aim from mouse Y and right stick Y (0.6.1). Earlier: Phase 1 and 2 passed in play (first-person walking, move while aiming, right mouse aims, left mouse fires); Phase 3 is done (crouch, roll, CQC on the fire key, knife, automatic guns); Phase 4 was a skip until 0.8.0. Grenades, lock-on, and an Xbox controller passed in play on 0.5.6 ("Play checks, 0.5.6").

## Resume

Read this section and "2026-09-23 research" before editing `native/fpvmove/fpvmove.c`. Close the game and launcher before any `add`, `enable`, `disable`, or `remove`. Build with `native/fpvmove/build.ps1` (it runs the file-scan test first, then writes `work/fpv-move-build/fpvmove.asi` and the package folder `work/fpv-move-pkg`). Install with `mgs3mod pack work/fpv-move-pkg --out work/fpv-move-<version>.zip`, then `disable`, `remove`, `add`, `enable --dry-run`, `enable`, `verify` against the game root.

Installed: `fpv-move` 0.8.4 (played) and `qcamo-face` 1.0.4-face.4 (played 2026-09-27; `qcamo.asi` SHA-256 `6c40bdaf6ddff2138294c02a1afa1b5e833afc23c4c929db0f99abe9e5d4a44a`, see "Taking the Mask off from the QCamo menu"), manager generation 164. Before face.4, generation 160: `fpvmove.asi` SHA-256 `5438ace896a2cad9657981051664094cda6386bf747a49076e5dbbfbe31a09b7`, matching `work/fpv-move-build/fpvmove.asi`; and `qcamo-face` 1.0.4-face.3 (kept as `work/qcamo-face-1.0.4-face.3.zip` for a rollback; `qcamo.asi` SHA-256 `7826bf46c36ec088f5e1ba0064d43fa80620ba613832c1ea581ccc4a57fd5e5f`, built by `native/qcamo-face/build.ps1`). 0.8.0 to 0.8.4 were played on 2026-09-26 (see "0.8.0" to "0.8.3"); the last play of 0.8.4 with face.3 passed everything the user tried in place of `qcamo` 1.0.4, which stays in the library disabled for a rollback. Last played: `fpv-move` 0.6.9 (shoulder aim finished; full regression pass on 2026-09-26), generation 129, SHA-256 `7d4bacc2d2aac5f64db6da16459278ec330ba0805d4b59a4b962ae448c553d32`. F7 turns the shoulder camera off (0.6.1 behaviour); the user dropped that mode on 2026-09-26. Its step 5 history is in "Step 5a, shoulder aim on lock-on" (0.5.7 lock, 0.5.8 lock fix and context log, 0.5.9 mouse aim, crouched strafe and pad layout, 0.6.0 crosshair, 0.6.1 manual pitch, 0.6.2 shoulder camera prototype, 0.6.3 quick fixes, 0.6.4 to 0.6.9 aim rig, shots on the camera ray, eases, pad swap, stick speed, walls). Before that, 0.5.6 was 0.5.1 plus the step 4a crouch/roll split and the step 4b rules. Also enabled: Ultimate ASI Loader 9.7.4, Crouch Walk 0.2.1 (and the QCamo fork above). `camo-test` 0.1.0 stays disabled. MGSHDFix is not installed and must not be installed beside 0.5.x or later (both own `aimingState`).

What 0.5.0 does: writes the first-person movement flag (Phase 1), patches three walk/look gates so a raised gun no longer blocks first-person walking (step 2a, "Phase 2 result, 0.4.3"), and wraps the player pad component so right mouse held aims and left mouse fires on press for semi-automatic guns (step 2b, "Phase 2 result, 0.5.0"). `fpvmove.log` beside the game executable is rewritten at each launch and records every patch, the walk early-out counts, and the pad rule counts. 0.5.1 adds the component and sub-state trace and F9 marks ("Phase 3 result, 0.5.1 tracer"). 0.5.2 to 0.5.6 add, in the same pad wrapper: C crouches and the Cross key (SPACE) only rolls, with a held roll key ending in prone ("Phase 3 step 4a, 0.5.2", "Phase 3 result, 0.5.5"); the fire key does CQC when right mouse is not held, and a hold is kept while aiming ("Phase 3 step 4b, 0.5.3", "Phase 3 result, 0.5.4"); the knife slashes on right mouse plus left mouse; automatic guns stay up between bursts. The pad log line counts each rule. Play logs are kept locally in `work/fpv-move-trace-20260924/` (`fpvmove-0.5.1-trace.log`, `fpvmove-0.5.3-play.log` to `fpvmove-0.5.6-play.log`, and `fpvmove-0.5.6-checks-AB.log` and `fpvmove-0.5.6-checks-C.log`); `work/` is ignored by git, so a fresh clone does not have them.

What has never been tested, including what this install cannot test (a retail Steam copy, Steam Input, sniper rifles and other weapons not in the save), is listed in [delta-controls-untested.md](delta-controls-untested.md). Released on 2026-09-27 as the GitHub prerelease v0.4.0-alpha.1 (tester kit with fpv-move 0.8.4 and qcamo-face 1.0.4-face.4; record in [v0.4.0-alpha.1-validation.md](releases/v0.4.0-alpha.1-validation.md)). Next: build the window app in [gui-plan.md](gui-plan.md) and release it as v0.4.0-alpha.2 before the user posts the alpha on Nexus, Reddit and X; then collect tester reports, especially from Steam copies and Steam Input. qcamo-face 1.0.4-face.4 passed on 2026-09-27 ("Taking the Mask off from the QCamo menu"); not reached there: the Tuxedo refusal with Mask on. The 0.8.0 list (wall aim, standing CQC, codec, quick toggles, shoulder aim) was played with face.4 on 2026-09-27 and the user reported it all works; the weapons never played are in "Still open after 0.8.0". After any play, copy `fpvmove.log` to `work/fpv-move-trace-20260924/fpvmove-<version>-play.log` (and `qcamo.log` beside it) before any relaunch, read the log, then ask the user short, batched questions. What is still open and why is in "Still open after 0.8.0". Then the branch decision (keep local for now, per the user on 2026-09-26) and release prep (fpv-move and the QCamo fork are not part of a public release yet, per the user). Agents: the user allowed a planned research fan-out and one independent reviewer per risky change for the 0.8.0 batch; otherwise work inline. Play sessions need no F9: the log records context lines.

A full `.text` disassembly of the pinned executable (dumpbin /DISASM) is kept locally at `work/disasm/text.asm`, with `work/disasm/dis.sh <RVA> [COUNT] [BEFORE]` to print from an address (VA = 0x140000000 + RVA; `page.idx` is its index). `work/` is ignored by git.

Do not repeat these misses: `0x38AEFB` / `0x38A750` (0.2.0), `0x387A20` / `0x387B30` (0.3.0), a NOP of `0x38A684`, the `+0x230` word stores, input ids `0xCA`, `0xC4`, `0xCC`, `0xD2` as aim deltas (bit sets through `0x358500` that return 0), clearing status `0x2E` globally (67 readers), and porting MGS2 first-person globals (0 hits in this executable).

## 2026-09-23 research

Three read-only static passes on 2026-09-23, game closed each time. Two worked from this plan and checked it. One was not shown this plan, `fpvmove.c`, or the log, and reached the same three gates from MGSHDFix sources and the executable alone. Spot checks of the key bytes were redone with `dumpbin /DISASM` in the main session. Addresses below are RVAs unless marked "file". `.text` file offset = RVA − `0xC00`; `.rdata` file = RVA − `0xE00`; `.data` file = RVA − `0x1200`.

### Corrections to earlier sections

- `r8` in the `41 83 E8 10 0F 84` chains is not a camera mode. It is a per-frame message. The player update at `0x3304F0` calls the component walker `0x3661F0` with messages `0x10`, `0x20`, `0x40`, `0x80` in that order. The 15 sites are 10 message-`0x10` and 5 message-`0x20` handlers. The traced RVAs in 0.4.0 and 0.4.2 show which components were alive, not which camera mode ran.
- The first-person component is registered at `0x38BD50`: handler `0x38B880`, init `0x38B3F0`, exit `0x38BC90`. `0x38BAC2` (file `0x38AEC2`) is its message `0x10` path and `0x38B8E5` its message `0x20` path. `0x38AEFB` is not an "aim-only camera mode". `0x387A20` belongs to another component.
- The check function at `0x38BB20` (file `0x38AF20`) returns 0 while first person holds (R1 `0x800` in `[actor+0x7E8]`, or a latched byte) and returns 5 on exit. A 5 is ORed into the component flags, the walker skips it, and cleanup `0x366030` runs the exit handler. It reads no aim state and no weapon bit.
- `gBP_1stPersonCamera_EnableMovement` (RVA `0x1E16D60`, BSS) has four readers, not one: walk at file `0x38BB2D` and `0x38BCB1`, and the stick/look function at file `0x389F85` and `0x38A5A6`. All four are `cmp`. No RIP-relative write exists. The walk function reads it before the weapon early-outs, so "weapon aim never reaches the movement flag" was wrong.
- The walk function's weapon early-out mask is `0x8000` (PS2 Square), not `0x800000`. Its full early-out list is in the gate table below.
- `66 89 ?? 30 02 00 00` occurs 7 times, not 4. Two more REX forms (`66 44 89`) exist. The extra stores are file `0x6DFF26`, `0x6E0EA6`, `0x6E12FF`, `0xB9CC6`, `0x34E5D1`. None is the WASD source.
- Word adds into `+0x90..+0x96`: 17 in the scanned range, 6 in `0x389EF0` and 11 in file `0x6CF198`–`0x6D0743`.
- The gun-window function stores `x − x/8` into both `+0xC4` and `+0x600`. It does not halve into `+0x600/+0x602`.
- `0x359020` (file `0x358420`) is not an input bit test. It tests the player status bitset at RVA `0x1E16CF8`. Its siblings are ANY `0x359040`, ALL `0x358FB0`, CLEAR `0x3590A0`, SET `0x359100`. Call counts: 1279 TEST, 1092 ANY, 44 ALL, 332 CLEAR, 781 SET. About 650 of 3528 call sites have an argument the scans could not resolve, so every "no other setter" statement carries that caveat.
- Word `0x1D7A224` is a menu confirm/cancel word (LMB, Enter, key 8 → bit 0; RMB, Backspace, key 9 → bit 1). It is copied as a block and read at RVA `0x111BC5`. It is not the gameplay fire path.
- The LSHIFT dword `0x1D79C60` is read indirectly by the key-config lookup `0x1123E0` as key id 6.
- MGSHDFix upstream now also lives at ShizCalev/MGSHDFix and has `src/features/mgs3_first_person_view_mode.cpp` (a `MGS3_FPS_DEV` skeleton that toggles the same flag). Its `pressure_inputs.cpp` has about 20 MGS3 patterns; all MGS3-labelled MGSHDFix patterns hit once in this executable. MGS2's `gBP_1stPersonCamera_*` shooter globals hit 0 times.

### Input chain

Keyboard and mouse fill the key array at `0x1D799E0` (4 bytes per VK). The packer `0x112780` builds a PS2 pad through key ids from table 0 (RVA `0xAC2230`), table 1 (`0xAC23D0`), or the custom table (`0x1E1FE20`). Table 0 mapping that matters here:

| Key id | Key | PS2 bit |
| --- | --- | --- |
| 0–3 | WASD | D-pad bits and left stick |
| 4 | LCTRL | halves the WASD stick (native slow walk) |
| 6 | LSHIFT | WASD writes D-pad pressure words instead of the stick |
| 8 | H | Circle `0x2000` |
| 9 | SPACE | Cross `0x4000` |
| 10 | LMB | Square `0x8000` |
| 11 | E | Triangle `0x1000` |
| 12 | RMB | L3 `0x0002` |
| 14 | M | L1 `0x400` |
| 16 | F | R1 `0x800` |

The pad becomes the gameplay pad at `0x1E2C560` (held `+0`, press `+0x14`, release `+0x18`, sticks `+4..+7`, pressure `+8..+0x13`, stick-deflected bits `+0x21`). The player pad component at `0x357B30` (registered at `0x3588BA`, hash `0x1810F2B`) copies it to `0x1E2D684` and sets actor `+0x7E0` stick magnitude, `+0x7E8` held, `+0x7EC` released, `+0x7F0` pressed, `+0x7F8` pad pointer. It runs only while `0x1E2D680` is nonzero (setting byte, 1 on disk).

### Ready and fire

That component is Konami's pressure emulator. `aimingState` is `0x1E2D6A8`, its sub-state `0x1E2D6AC`, jump table at `0x35887C`.

- Square press sets `aimingState = 1` (`0x357E7C`). Square release clears it (`0x35800E`, the MGSHDFix keep-aim site) and passes the release to the game.
- L3 (RMB) with no gun raised synthesizes a Square press (`0x357EF6`). While ready and no real Square is held, it holds a fake Square at pressure `0x77` (`0x357ED3`). L3 again synthesizes a slow release at pressure `0x10` with sub-state 3, which lowers without firing.
- The gun attack state machine `0x3741E0` fires a semi-auto weapon on Square release when the previous pressure was above 0. An automatic weapon (weapon flag bit 11) fires while pressure ≥ `0x78`. The fake hold `0x77` sits one below that. The cancel predicate `0x37B3B0` lowers without firing when `aimingState == 0 && sub-state == 3`, or when pressure never passed `0x10`.

So MC's "right mouse ready, left mouse press to raise and release to fire" is RMB = L3 toggle and LMB = Square, filtered by this component.

### Why WASD aims with a gun raised

First-person message `0x20` runs check → `0x38B350` → walk `0x38C6B0` → body yaw → lean → look `0x38AAF0` → clamp, every frame, armed or not.

| Gate | Where | Pattern (count 1 in the file) | Effect with gun raised |
| --- | --- | --- | --- |
| Walk: Square held | `0x38C780` (file `0x38BB80`) | `F7 83 E8 07 00 00 00 80 00 00 0F 85 ?? ?? ?? ?? BA 15 01 00 00` | the fake `0x77` hold or a held LMB exits walk |
| Walk: status list | `0x38C795` (file `0x38BB95`) | `C7 44 24 30 FF FF FF FF C7 44 24 28 27 00 00 00 C7 44 24 20 2E 00 00 00` | ANY(`0x9B`, `0x115`, `0x114`, `0x99`, `0x2E`, `0x27`) exits walk; `0x2E` is set by 8 gun-attack states and means "weapon raised" |
| Look: left stick | `0x38ABA6` (file `0x389FA6`) | `41 8D 4E 2E E8 ?? ?? ?? ?? 85 C0 75 1A 48 8B 86 F8 07 00 00 0F B6 78 04` | with the flag set and `0x2E` set, the view comes from the left stick (WASD) when it is deflected |

Walk's other early-outs, in order: status 7 clear (file `0x38BB20`), flag 0 and status `0xBB` clear (bypassed by the flag), component byte `+0xBB` nonzero (first frame only), then the two weapon gates, then no stick (`[actor+0x7E0] == 0` and statuses `0xBF`/`0xC0` clear). When walk passes, it writes velocity to `[actor+0x620]`/`[actor+0x628]` and sets statuses `0x12E`, `0xC1`, `0xC2`, `0xDE`. Those are read elsewhere by weapon code, so the gun may lower or the ready state may thrash. That is a playtest question.

Rejected: clearing `0x2E` globally (67 readers), clearing it around the look call (it also drives aim speed and body-follow inside `0x389EF0`), calling walk from another site (it already runs every frame).

### Other findings

- Crouch Walk (cipherxof/MGS3CrouchWalk) names ActMovement `0x369800`, ActSquatStill `0x3686B0`, GetButtonHoldingState `0x36A590`, PlayerSetMotion `0x3659E0`, PlayerStatusCheck `0x359020`. (Corrected 2026-09-26: `0x3659E0` does not set a motion; it returns a motion word from the table at `0xB1EA30` by weapon and stance index. The one movement motion send is the call at `0x369972`.) Its README scope and its patch at `0x368828` may already give "crouch stays while moving" and "hold crouch = prone". That overlap is unverified in play and must be checked before any Phase 3 crouch work.
- The local pad path does not go through Valve Steam Input: pad buttons are mapped before the game reads them, from XInput devices. Local pad tests therefore do not prove behavior under Valve Steam Input. A Steam copy's executable is SteamStub-wrapped on disk, so it fails the profile hash (until v0.4.0-alpha.1), and an ASI that scans in `DllMain` on a still-wrapped image would find 0 matches.
- Nine sites pass `4` to the pause set `0x10ED10`: `0x98505` (jmp), `0x32DA44`, `0x32E6C1`, `0x5B7379`, `0x5B76D4`, `0x6CA41F`, `0x6CAB38`, `0x6CBBCC`, `0x6DA520`. None is proved to be radio. (Corrected 2026-09-26: `0x10ED10` has 19 call sites, 18 calls and 1 jmp; 9 pass a literal 4 (these), 3 a register (`0x98051`, `0x2EED09`, `0x6E07AE`), 3 pass 8 and 4 pass 1. The codec opener `0x2EEAF0` sets bit 1 at `0x2EED09`; see "Phase 4. Shortcuts".) The actor loop at `0x10F0FF` skips whole actor lists by mask, not one actor.
- Global `0x1E16D64` sits beside the movement flag, is only read, and gates run statuses at `0x36A96F`. It may be another dormant Bluepoint flag. UNKNOWN.

Launch only when a playtest is needed, with the user's go, and only with the game closed:

```powershell
& 'C:\Games\METAL GEAR SOLID 3 - MCV\mgs3mod.exe' launch --profile na-startup --json --game-root 'C:\Games\METAL GEAR SOLID 3 - MCV'
```

Close the game and launcher before `disable`, `remove`, `add`, or `enable`. Source build is `native/fpvmove/build.ps1`. (2026-09-23 note, since superseded: Phase 3 is done, and Phase 4 was reopened and built in 0.8.0.)

## Goal and decision

Make the fingerprinted Metal Gear Solid 3 Master Collection installation play with the keyboard, mouse, and controller mechanisms of Metal Gear Solid Delta: Snake Eater New Style.

The mod is one new root `.asi` loaded by the Ultimate ASI Loader already installed as `wininet.dll`. It is not an install, fork, or required dependency on MGSHDFix. MGSHDFix stays out of this game folder. Its published pattern strings are a map for Phase 0. A hit means this plugin hooks that site itself. Installing MGSHDFix does not create Delta controls: its shipped MGS3 input work is mouse-sensitivity scaling and keeping a gun raised after a shot.

Do not replace `wininet.dll`. Do not retarget QCamo or Crouch Walk. Do not expand package schemas until a second file is actually required. Key remapping, `-ctrltype`, and Steam Input presets are not this mod.

## Evidence and confidence

Public mod survey, 2026-09-21: no released mod gives Master Collection both keyboard and controller Delta New Style. Closest released pieces are MGSHDFix keep-aiming, the separate Keep Aiming mod (gun stays up after a shot), Crouch Walk, and QCamo. Steam Input presets such as "MGSV/Modern Control Scheme" and "MGS3 - Modernised Controls" only rebind the existing verbs. Nexus's full gameplay catalog was behind Cloudflare, so an unlisted mod remains possible. Confidence: medium.

Delta behavior was checked against the IGN changes page and the 22 Aug 2025 PlayStation Blog hands-on. Master Collection keyboard and attack rules were checked against Konami's PC manual (pages 04 and 05). Confidence on the minimum set below: high. Confidence that a full port is possible on this executable: low until Phase 0 hits.

Local facts, read from this repo and install:

- Supported executable SHA-256 is `0d585dcc6a671be5d64d3d0a856c53f9ee0e58e7e4993f76dff29772c7a4bc80` (`internal/profile/profile.go`).
- `Engine.dll` SHA-256 is `4067774bd2945dfab1a81ee0f657b3b6c9414b1b363d29830652e6d93c516996`. Both files are ordinary PE32+ AMD64. No PDB is in the game root.
- Schema 2 already accepts one new root-level lowercase `.asi` with an absent original. Subdirectories, DLL proxies, executables, arbitrary data files, and replacement of an existing plugin are refused (`docs/runtime-plugins.md`, `internal/packagefmt/packagefmt.go`).
- Schemas in force are 1 through 4 only. Schema 4 is the pinned Crouch Walk set. An INI or extra animation does not fit schema 2.
- The live loader is `wininet.dll` because `Engine.dll` imports that proxy. The game executable's import table does not list `dinput8.dll` or `XInput`.
- This repo has no input-hook source. Crouch Walk's static inspection found 11 embedded signatures, each once, in the supported executable (`docs/crouch-walk.md`). That confirms signature presence only, not what those functions do.
- MGSHDFix master `gamevars.cpp` names `aimingState`, `heldTriggers`, and, inside `#if defined(MGS3_FPS_DEV)` only, `gBP_1stPersonCamera_EnableMovement`. `keep_aiming_after_firing.cpp` has an MGS3 mid-hook and returns immediately when a pressure-sensitive pad is present. Current MGSHDFix README lists Bluepoint's first-person shooter camera and Subsistence freecam under MGS2. MGS3's own feature list is grass distance and mouse sensitivity.
- Phase 0 scanned those patterns against this install. The result is below. MGSHDFix issue 206 printed different offsets from another run on 23 Nov 2025. Those offsets are not this build.

## Target

Ship these behaviors, then stop:

1. Over-the-shoulder aim on right mouse / LT. Fire on left mouse / RT. Holding aim does not fire on release.
2. Walk and strafe while that aim is held.
3. First person that allows movement. IGN says the Delta speed is slow. The PlayStation Blog says it feels like a shooter. This mod takes whatever movement the Master Collection camera flag already implements. It does not invent a speed.
4. A crouched Snake stays crouched while moving. Holding crouch goes prone. Roll is a separate button from crouch. Crouch-walk animation files are already installed and are not part of this plugin.
5. CQC grab on the fire input. Throat slit on the action input, matching Delta's "press the action button while holding" rule rather than a hard Circle press.
6. Stalk on Shift / LB.

D-pad or keyboard shortcuts for camo, weapons, equipment, and radio are Phase 4. They ship only if those windows can open while gameplay is still running. QCamo already changes camouflage. Do not build a second camo wheel by default.

## Out of scope

- Legacy Style. New Style buttons and the New Style camera are one package in Delta. Some setpieces force the old camera even there. This mod does not add a runtime camera toggle that reloads a checkpoint.
- Bullet drop, new directional CQC animations, photo mode, the compass, Ga-Ko ducks, and film reels.
- Aim assist and shoulder swap. Aim assist is a targeting system. No opened source gave a shoulder-swap button.
- A dedicated reload button. The IGN and PlayStation pages opened for this plan do not list one. Prima and other pad charts disagree. Do not build a reload verb from a disputed chart.
- Replacing Steam Input with SDL3. MGSHDFix issue 315 records that as a maintainer "one day maybe," because Steam Input does not pass PS3 analog face buttons. The game still applies pressure if the weapon button is bound to a trigger. That is the opposite of this mod's aim/fire split.
- Open world, an iDroid, or new weapons.
- Redistributing MGSHDFix, Crouch Walk, or QCamo payloads.

## Why remapping cannot do this

Master Collection still uses four verbs: ready/unready, weapon, aim, and first person. Konami Layout A binds left click to the weapon button and right click to ready/unready. The aim button is a different key, and it matters inside first person. First person plants the player. The PS2 light-press "raise" versus hard-press "fire" split is gone; left-stick click is the stand-in that lowers a gun so releasing the weapon button does not shoot. `-ctrltype` selects prompt icons. It does not select a control scheme.

A Steam preset can put ready on LT and fire on RT. Players still cannot aim and move. That limitation is the mod.

## Packaging

Phase 1 and later ship one schema 2 package: a single AMD64 `.asi`, absent origin, managed loader already enabled. No new schema. Settings stay compiled in until a real option exists. The manager does not order ASI load. The new plugin must tolerate `qcamo.asi` and `mgs3crouchwalk.asi`, because that is the installed set. Two plugins writing the same aiming state will fight. This mod and MGSHDFix must not both own that hook.

## Phases

Each phase stops on a miss. Do not invent addresses. Record patterns and match counts. A runtime offset for this executable may be logged as scan evidence and must not be hardcoded.

### Phase 0. Pattern scan

Read-only. Do not launch the game. Do not write a plugin.

Scan `METAL GEAR SOLID3.exe` and `Engine.dll` for the MGSHDFix MGS3 patterns below. `??` is a wildcard byte. These strings were copied from MGSHDFix master `gamevars.cpp` and `keep_aiming_after_firing.cpp`. They are claims until the scan runs.

| Name | Pattern | Role |
| --- | --- | --- |
| `aimingState` | `8B 35 ?? ?? ?? ?? 48 8D 0D ?? ?? ?? ?? 49 89 8D` | Required. A miss stops the project. |
| `heldTriggers` | `48 8D 3D ?? ?? ?? ?? BA ?? ?? ?? ?? 41 B8` | Required. First person, lock-on, weapon menu, and equipment menu are bit tests on this value in MGSHDFix, not separate scans. |
| `gBP_1stPersonCamera_EnableMovement` | `83 3D ?? ?? ?? ?? 00 75 ?? B9 BB 00 00 00 E8 ?? ?? ?? ?? 85 C0 0F 84` | Required for Phase 1. Compiled out of shipping MGSHDFix behind `MGS3_FPS_DEV`. A miss stops Phase 1 and forces a stop/go decision before any aim work. |
| Keep aiming, site 1 | `48 89 1D ?? ?? ?? ?? 4C 8D 15` | Reference only. Shows where a shot clears aim. Not the movement hook. |
| Keep aiming, site 2 | `4C 8D 15 ?? ?? ?? ?? 4C 8B A4 24` | Reference only. Used when always-keep-aiming overrides state. |

Also search both binaries for strings that would identify a crouch/crawl branch or a CQC button dispatch. No MGSHDFix pattern covers those. A string miss stays UNKNOWN. It does not get a guessed address.

Pass: `aimingState` and the first-person movement pattern each match exactly once in the game module. Record the pattern, the module, and the match count.

Fail: zero matches or more than one match on either required pattern. Stop. Do not proceed to a plugin.

The `heldTriggers` bit masks were copied from MGSHDFix `gamevars.hpp` during this phase. They are listed in the result. Do not guess different bits.

#### Phase 0 result

Scanned on 2026-09-21. Read-only. The game was not launched. Both files were hashed before the scan and matched `internal/profile/profile.go`.

| Module | SHA-256 | `aimingState` | `heldTriggers` | first-person movement | keep-aiming site 1 | keep-aiming site 2 |
| --- | --- | --- | --- | --- | --- | --- |
| `METAL GEAR SOLID3.exe` | `0d585dcc6a671be5d64d3d0a856c53f9ee0e58e7e4993f76dff29772c7a4bc80` | 1 | 1 | 1 | 1 | 1 |
| `Engine.dll` | `4067774bd2945dfab1a81ee0f657b3b6c9414b1b363d29830652e6d93c516996` | 0 | 0 | 0 | 0 | 0 |

Pass. The two required patterns each match once, and only in the game executable. Phase 1 is allowed. These RVAs are scan evidence for this hash. A plugin must resolve them from the patterns at load. They must not be hardcoded.

The displacement in each match was read as a signed 32-bit RIP displacement. Target RVA is match RVA plus instruction length plus that displacement. All three globals land in `.data`, whose virtual size (`0x14B0718`) extends past the bytes stored in the file:

| Site | Match RVA | Target RVA | Section |
| --- | --- | --- | --- |
| `aimingState` | `0x357C11` | `0x1E2D6A8` | `.data` |
| `heldTriggers` | `0x111E65` | `0x1D79994` | `.data` |
| `gBP_1stPersonCamera_EnableMovement` | `0x38C72D` | `0x1E16D60` | `.data` |
| Keep-aiming site 1 | `0x35800E` | `0x1E2D6A8` | `.data` |

Keep-aiming site 1 stores to the same RVA as `aimingState`. That cross-check is why the reference pattern is useful. It is still not the movement hook. Keep-aiming site 2 is the next instruction, at match RVA `0x358015`. Its own displacement resolves to RVA 0, so it is a code anchor, not a global.

`gamevars.hpp` on MGSHDFix master defines the MGS3 `heldTriggers` bits as booleans, not analog bytes:

- `MGS3_EquipmentMenu` = `1 << 8` (`0x100`)
- `MGS3_WeaponMenu` = `1 << 9` (`0x200`)
- `MGS3_LockOn` = `1 << 10` (`0x400`)
- `MGS3_FirstPerson` = `1 << 11` (`0x800`)

ASCII and UTF-16 searches of the executable found no `crouch`, `crawl`, `BP_Camera`, `1stPerson`, or `EnableMovement`. `CRAWL` occurs once, inside the UI label `CRAWL BUTTON`. `CQC` occurs seven times: five are localized button labels (`CQC BUTTON` and its translations) and two are the short token `CQC` with no function name beside them. Crouch, crawl, and CQC dispatch addresses stay UNKNOWN.

The immediate `B9 BB 00 00 00` occurs 15 times. Only the full first-person pattern is unique. Issue 206's `aimingState` offset `+1E2B5A8` and `heldTriggers` offset `+1D778C4` do not match this executable.

### Phase 1. First-person walking

Only if Phase 0 found the movement flag once.

One schema 2 `.asi`. Set that flag the way the unshipped MGSHDFix file does. No INI.

Pass, with the game closed afterward: the player can move in first person; firing still works; QCamo and Crouch Walk still work; save payloads are unchanged.

Fail: the flag does nothing, movement plays no animation, or either existing mod breaks. The release is then nothing. Phase 1 is not over-the-shoulder Delta aim. It is the only named mechanism.

#### Phase 1 result

Built and enabled on 2026-09-22. Source is `native/fpvmove/fpvmove.c`. `native/fpvmove/build.ps1` compiles it with the Visual Studio 2022 x64 tools, `/MT`, and no INI.

The file-scan test against `METAL GEAR SOLID3.exe` reported `matches=1 offset=0x38BB2D`, the same site Phase 0 found. The plugin does not store that offset. On load it scans executable sections of the game module for the Phase 0 pattern, requires exactly one match, resolves the RIP displacement, and writes `1`. MGSHDFix's unshipped movement default is `1`, but its feature switch defaults to off. Copying that switch would install a plugin that never writes the flag, so this package always writes `1`. There is no hotkey.

`mgs3mod pack`, `add`, `enable --dry-run`, `enable`, and `verify` all exited 0. `verify` reports generation 17, with `fpv-move` 0.1.0 enabled beside the loader, QCamo, and Crouch Walk. `camo-test` stayed disabled. Installed `fpvmove.asi` is 135,168 bytes, SHA-256 `6c8bccb9bfbc9a8b7ff5a771570f0a41c9a8ddee8c609f269653dfce4a290145`, matching the package. The game process was not running. No gameplay was observed.

The next launch writes an unowned `fpvmove.log` beside the executable. `set movement flag to 1` means the write happened. Any other line means the flag was left alone. The Windows loader zeroes the global before this plugin runs. If the game writes it back to 0 later, the log can still say the flag was set and first person will still be planted. That result is a Phase 1 fail, not a reason to invent a second hook in this phase.

Played on 2026-09-22: first person walks. QCamo, Crouch Walk, and the rest of the installed set still behaved. With a weapon drawn, WASD turns the view instead of moving the body. That is Master Collection weapon aim. Phase 1 does not change it. Correction (2026-09-23): the flag has four readers, in walk and in the look function, and no RIP-relative writer. See "Why WASD aims with a gun raised".

Disable with the game and launcher closed:

```powershell
& 'C:\Games\METAL GEAR SOLID 3 - MCV\mgs3mod.exe' disable fpv-move --game-root 'C:\Games\METAL GEAR SOLID 3 - MCV'
```

### Phase 2. Move while shoulder-aiming

The product. No public mod does this for Master Collection.

Hold aim, move, and fire only on the fire button. Standing strafe while aimed must work. Prone aim must be able to face one way while the body moves another, if the current prone state allows it. Right mouse must not remain "ready/unready" and left mouse must not remain "press to ready, release to fire."

Pass is a played session on this install, not a pattern hit. Fail, and the honest stop is Phase 1 or nothing. Do not hide a miss behind a Steam preset.

#### Phase 2 result, first slice

This slice is 0.2.0. That build is no longer installed. The live plugin is 0.4.2, described under Resume.

File offset `0x38A750` has two callers. File offset `0x38AD18` calls it and then calls the walk function at `0x38BAB0`. File offset `0x38AEFB` calls it and returns. That second site is the aim-only camera mode. The walk function has no other caller. A re-count on 2026-09-22 found the same two `E8` sites and the same one walk caller. Weapon aim never reaches the movement flag.

`0x38A750` is not the instruction that adds the stick into the view. That add is `66 01 BB 90 00 00 00` at file offset `0x38A684`, inside the function that begins at `0x389EF0` and returns at `0x38A747` (`C3`, then padding, then a new prologue at `0x38A750`). The only caller of `0x389EF0` is file offset `0x38AD53`, on the mode that already calls the walk function. The aim-only path does not call it. Do not NOP `0x38A684`. Its inputs are already sums of two words each, and mouse aim was not shown to be one of those words. `GetCursorPos` is still the one call at file offset `0x11583E`. It stores the cursor delta at object offsets `0x50` and `0x54`. No instruction in that function writes the aim words.

`fpv-move` 0.2.0 wrote the flag. It also found those two sites by pattern, not by a stored address, and redirected the aim-only call through a stub. The stub calls the angle function, then the walk function, with the same two arguments. Mouse look goes through the one `GetCursorPos` call at file offset `0x11583E` and is not patched. The angle function still applies the stick, so WASD can still turn the view while the body moves. Fire-on-release and the right-mouse ready verb are unchanged.

The file-scan test reported `matches=1 offset=0x38BB2D aim=1 aim_off=0x38AEF5 angle=0x38A750 move=1 move_off=0x38AD12 walk=0x38BAB0`. `disable`, `remove`, `add`, `enable --dry-run`, `enable`, and `verify` exited 0. `verify` reports generation 21. `fpv-move` 0.2.0 is enabled beside the loader, QCamo, and Crouch Walk. Installed `fpvmove.asi` is 137,728 bytes, SHA-256 `1073e2f0e987ba1abcd9d01522097fe15cee1146f66e520234499ff52ac2fe20`.

Played on 2026-09-22 with 0.2.0: walking and crouching work. First person with the gun holstered still moves on WASD. The mouse still turns the view. With the weapon drawn, WASD turns the view and does not move the body. The mouse does the same. The aim-only walk call did not produce movement. Holstered first person and the other installed mods were left alone.

The 0.2.0 log contained the flag line and `aim-mode call now runs walk after aim`. The installed 0.4.2 log is under Resume. Correction (2026-09-23): the mask is `0x8000`, not `0x800000`, and the list was incomplete. The full list and which gates a raised gun trips are under "Why WASD aims with a gun raised".

### Remaining sequence

Checked read-only on 2026-09-22 against this executable. The game was closed. Five sweeps: stick versus mouse, fire and ready, crouch/roll/CQC/stalk, menus, and a blind look that was not shown this section. The blind look anchored on a different angle add at file offset `0x3BC120` and did not find the pair above. That does not erase the re-counted callers. It does mean a second camera machine exists and is not this hook.

Items 1 and 2 are closed misses. The live plugin and the next edit are under Resume. A miss stops that phase. No stored address in the plugin.

Items 1–6 are the 2026-09-22 record. Their "camera mode" labels, the four-store count, the `0x800000` mask, and the item 4 and 5 input readings are corrected under "2026-09-23 research". Read that section first.

1. Played on 2026-09-22 with 0.2.0. With the gun drawn, the body did not move. WASD and the mouse both turned the view. That hook called the walk function from file offset `0x38AEFB`. The function there, `0x38A750`, copies facing once when input `0x57` is held. It does not turn the view from WASD.
2. The call that does turn the view is file offset `0x387A20`. Camera mode `r8 == 0x10` jumps there and calls file offset `0x387B30`, then returns. `0x387B30` reads either stick bytes at `+6` and `+7` or keyboard direction bits `0x10`, `0x20`, `0x40`, and `0x80`. It has no `+0x50` mouse-cursor operand. `fpv-move` 0.3.0 keeps the movement flag, does not patch `0x38AEFB` or `0x38A684`, and redirects `0x387A20` to the walk function. The file-scan test reported `weapon=1 weapon_off=0x387A20 aim=0x387B30 walk=0x38BAB0`. `verify` reports generation 25. Installed `fpvmove.asi` is 137,728 bytes, SHA-256 `5490f3d48fe45397660850ce582fae83bb0c24c55cd79e72e32031d74561d466`. Gameplay of 0.3.0 was played on 2026-09-22. The log line `weapon-aim stick call now runs walk` was written, walk RVA `0x38C6B0`. With the gun drawn, the body still did not move. WASD still aimed. The mouse still aimed. Same result as 0.2.0. Replacing `0x387A20` did not remove aim, so that call is not the live gun-aim path, or a second function still applies WASD to the view. Another session on 0.3.0 will not change it.

   The four `66 89` stores of object offset `+0x230` are not that second function. File offsets `0x34E4AF`, `0x350272`, `0x38FE87`, and `0x3A57F6` are the only such stores in the scanned image. `0x350272` writes an id beside constants `0x1D` and `0x1E`. `0x38FE87` and `0x3A57F6` store a global scaled with the `0x66666667` division. The dumped bytes do not read a stick. The integrator at `0x3BC120` still adds those words into `+0x228`, but these stores are not a per-frame WASD or mouse source. Word adds into `+0x90`, `+0x92`, `+0x94`, and `+0x96` in file range `0x1000`–`0x900000` sit in the walk-mode function at `0x389EF0`. Its only direct caller is `0x38AD53`. `0x387B30` does not write those words.

   `fpv-move` 0.4.0 keeps the movement flag and removes the 0.3.0 redirect. Each of the 15 `41 83 E8 10 0F 84` branches still reaches its original target. The taken branch sets a byte first. Every two seconds `fpvmove.log` appends the target RVAs that ran. The file-scan test reported `matches=1 offset=0x38BB2D modes=15` and the same 15 targets. `verify` reports generation 29. Installed `fpvmove.asi` is 140,288 bytes, SHA-256 `1e947f836b06dec3763a88a52a74017c5021b66cf8bfcdcff2fadbf1d740c435`. Loader, QCamo, and Crouch Walk stayed enabled. `camo-test` stayed disabled. Played on 2026-09-22. The player entered first person, drew the weapon, aimed with the mouse, then aimed with WASD in every direction, fired once, released the weapon, left first person, and paused. `0x3636A2` ran the whole session. `0x36446F` appeared only in the first sample. `0x38BAC2` (file `0x38AEC2`, the 0.2.0 branch) turned on with first person and was still present after the gun pair dropped. `0x37A36E` and `0x37A641` (file `0x37976E` and `0x379A41`, the start and the tail of one function) were on together for every sample that covered the mouse and the WASD. They did not split by device. `0x388620` (file `0x387A20`, the 0.3.0 branch) never appeared. No second session.

3. The gun-window function at file `0x37976E` does not read the stick. Its calls with ids `0xCA`, `0xC4`, `0xCC`, and `0xD2` go to `0x358500`, set those input bits, and return 0. `0x371B10` fills a stack struct for a float compare. The function halves `[rbx+0xC4]` and `[rbx+0xC6]` into `[rdi+0x600]` and `[rdi+0x602]`. The tail adds a word from another object into `[rbx+0xA8]`. The logged branch `0x38AEC2` is mode `r8 == 0x10` and returns without calling `0x389EF0`. Mode `r8 == 0x20` is the short branch at `0x38ACB1`, target `0x38ACE5`. `0.4.1` missed it because the 15-site patch rewrote the pattern, and its log said `armed (15 sites)`. `0.4.2` is the install under Resume. The 0.4.0 target RVAs were `0x338C55`, `0x338C08`, `0x343B8D`, `0x3438C9`, `0x346507`, `0x346168`, `0x3636A2`, `0x36446F`, `0x37A641`, `0x37A36E`, `0x384772`, `0x387888`, `0x388620`, `0x38855D`, and `0x38BAC2`.
4. Fire and ready stay UNKNOWN. The input bit-test at file offset `0x358420` has 1279 direct calls and 185 immediate ids. None is proved to be release-to-fire. Left and right mouse are VK slots 1 and 2, folded into bits 0 and 1 of word RVA `0x1D7A224`, and that word has no reader outside the packer at file offset `0x111C55`. The next sweep follows that packer's copy, after movement. A Steam preset is not the substitute.
5. Phase 3 starts only after the played movement pass. All five verbs are UNKNOWN. Layout A data has an LSHIFT row and a SPACE row, and the LSHIFT poll at file offset `0x111C50` is unique, but the dword it writes has no direct reader. The first Phase 3 slice is that poll. Crouch, roll, CQC, and throat slit stay UNKNOWN until a unique pattern exists. Crouch Walk's 11 signatures are not those patterns.
6. Phase 4 is SKIP. Weapon open at file offset `0x32CE3F` and equipment open at file offset `0x32DABC` both pass `4` to the pause set at file offset `0x10E110`. The actor loop tests that dword and skips the actor. Camo stays on QCamo. Radio was not proved to pause, and it was not proved to stay live either. No shortcut hook and no second camo wheel. (Corrected 2026-09-26: the labels are swapped; file `0x32CE3F` is the item (L2) window and `0x32DABC` the weapon (R2) window. The radio pauses with bit 1. Phase 4 was reopened and delivered in 0.8.0.)

### Phase 2 steps after the 2026-09-23 research

Step 2a, movement (0.4.3). Keep the movement flag. Patch the three gates by pattern, each only on exactly one match in the executable sections, each with its bytes checked before the write:

- walk Square gate: the `test` immediate `00 80 00 00` becomes `00 00 00 00`, so the `jne` never takes;
- walk status list: the `0x2E` argument becomes `0x99`, which is already in the list; `0x27` stays until the log shows it;
- look gate: `75 1A` becomes `90 90`, so with the flag set the view always comes from the right stick and the mouse path.

All three sit in functions whose only caller is the first-person component, so third person is untouched. Drop the 16-site component trace. Wrap the one walk call (file `0x38AD23`) through a near thunk and log, every 2 s, how many walk calls ran, the first failing early-out per call, how often Square, `0x2E`, and `0x27` were set, and how often walk left a nonzero velocity.

Pass: in first person with a gun raised, WASD moves the body, the mouse aims, and firing still works. Fail branches:

| Log with gun raised and W held | Meaning | Next |
| --- | --- | --- |
| reason `0x27` | status `0x27` still blocks | drop `0x27` from the list too |
| reason status 7 | raising the gun clears status 7 | study the setter at `0x3839D1` before any patch |
| pass with velocity, body still | something after walk zeroes or ignores `+0x620/+0x628` in weapon states | find the integrator; new target |
| pass and moving, gun lowers or fire breaks | walk's statuses `0x12E`/`0xC1`/`0xC2`/`0xDE` disturb weapon states | mask those sets in the wrapper |

#### Phase 2 result, 0.4.3

Built and enabled on 2026-09-23. The file-scan test reported one match each: flag `0x38BB2D`, walk Square gate `0x38BB80`, walk status list `0x38BB95`, look gate `0x389FA6`, walk call `0x38AD12`, status TEST helper `0x358420`, and walk callee `0x38BAB0` with its prologue and `+0x2B` bytes checked. `pack`, `disable`, `remove`, `add`, `enable --dry-run`, `enable`, and `verify` exited 0. Manager generation 41. `fpv-move` 0.4.3 is enabled beside the loader, QCamo, and Crouch Walk; `camo-test` stays disabled. Installed `fpvmove.asi` is 143,360 bytes, SHA-256 `b860732bbc89e7940d553e764931d8dd638091ee206f46562a58bab8cd788502`, matching `work/fpv-move-build/fpvmove.asi`. Not played yet.

The log is rewritten at each launch. It should start with `fpvmove: 0.4.3`, the flag line, three `patched at rva` lines (`0x38C786`, `0x38C7A9`, `0x38ABB1`), and `walk call at rva 0x38B923 traced, walk rva 0x38C6B0`. Every 2 s with first person active it appends `walk calls=N`, the count per first failing early-out (`pass`, `s7`, `flag`, `first`, `square`, `list`, `stick`), how many calls saw Square held, `0x2E`, and `0x27`, and `moved`, the passes that left a nonzero `+0x620`/`+0x628`. `stick` is normal while standing still. Walking with a gun raised should show `pass` and `moved` rising while `b2E` is nonzero.

Played on 2026-09-24: step 2a passes. The user entered first person, walked holstered, then raised a gun and played with WASD moving and the mouse aiming, and reported it working. The load log showed the flag line, all three `patched at rva` lines at the expected RVAs, and the walk trace line. Walk ran 120 times per 2 s sample (60 per second). From 07:13:03 to 07:13:17 every sample had `sq=120 b2E=120` with `pass` 42–115 and `moved` equal to `pass`, so the body moved with Square held and status `0x2E` set. Holstered samples show the same `pass`/`moved` with `sq=0 b2E=0`. `s7` appears only for about 15 calls at each first-person entry. `square` and `list` never fired as the first failing gate except the last sample, `list=40 b27=40`: status `0x27` stopped walk for about 0.7 s. What set `0x27` there (reload or a fire animation) was not recorded. Leave `0x27` in the list unless a player reports a stall. Third-person behavior was not reported separately. The game was closed afterward.

Step 2b, ready and fire (0.5.0, after 2a passes). Wrap the player pad component: retarget the `lea r9` at registration `0x3588BA` (pattern `4C 8D 0D ?? ?? ?? ?? C7 44 24 40 05 00 00 00 45 33 C0 C7 44 24 38 10 00 00 00 BA 2B 0F 81 01`, one match) to a wrapper. The target is checked by `48 89 54 24 10 48 89 4C 24 08 53 55 56 57 41 55 41 56 41 57 48 83 EC 40 0F 10 05` (one match at `0x357B30`). The raw pad address comes from the `0F 10 05` displacement. The wrapper edits the raw pad held/press/release/pressure words, calls the original, then restores them so other readers see the real pad:

- RMB press while not ready: pass the L3 press (native ready).
- RMB release while ready and Square not held: pass an L3 press, which runs the native no-fire lower.
- RMB press while ready: drop it, so holding does not toggle off.
- RMB held after a shot cleared the aim: re-send the L3 press.
- LMB press while ready: send one frame of Square release at pressure `0xFF`, then mask the real Square until LMB is released. Automatic weapons need pressure ≥ `0x78` while LMB is held instead.

#### Phase 2 result, 0.5.0

Built and enabled on 2026-09-24, before play. Rechecked in `dumpbin` first: the component copies the raw pad (`0x1E2C560`, 0x24 bytes) to `0x1E2D684` on every call, before any message test. Actor is `rcx`. Weapon context is `[actor+0x6F8]`; the dword at `+0x20` holds the weapon id in its low byte, the automatic flag in bit 11 (`0x800`, tested by `0x37B3B0` and `0x374723`), and the `0x20000` charge flag. The Square pressure byte is raw `+0x13` (`0xA1C30` returns 30 − bit, and `0x110410` indexes `+8 + bit − 4`), so the research note that put it at `0x1E2D696` was wrong; it is `0x1E2D697` in the copy. The filter's standard path needs the filter on (`0x1E2D680`), a weapon id above 2 and not `0x1B`, no `0x20000`, and statuses `0x78` and `0x85` clear.

The wrapper acts only on that standard path. Rules, per call:

- LMB is hidden from the filter after a shot this wrapper made, until LMB is released.
- LMB press with RMB held, a semi-automatic gun, and the gun up on the fake `0x77` hold (`aimingState` 1, sub-state 2): the raw pad shows a Square release at pressure 0 instead. The filter clears `aimingState` and passes the release; the gun's previous pressure was `0x77`, so it fires.
- LMB press with RMB held and the gun not up: the native press raises; 3 calls later the wrapper forces the release, unless LMB was let go first.
- RMB press while ready is dropped, so a held RMB never toggles ready off.
- RMB release while ready and LMB not held: an L3 press is added, which runs the filter's no-fire lower (pressure `0x10`, sub-state 3).
- RMB held and not ready for 8 calls: an L3 press is added, which re-readies after a shot.
- Automatic guns, weapon ids 0–2 and `0x1B`, charge weapons, and LMB without RMB are passed through unchanged. Hip fire and CQC stay native.

The raw pad is restored after the component returns. The file-scan test reported one match each for the registration (`0x357CBA`), the component (`0x356F30`), and `aimingState` (`0x357011`), with the registration resolving to the component and both raw-pad loads agreeing. `pack`, `disable`, `remove`, `add`, `enable --dry-run`, `enable`, and `verify` exited 0.

The log adds a `pad component rva 0x357B30 wrapped` line at load, one `pad` line with the first call count, then a `pad` line every 2 s in which a rule fired: `calls`, `std` (calls on the standard path), `fire`, `tap`, `drop`, `lower`, `rearm`. `calls` near 120 per 2 s means one call per 60 Hz frame. `calls=0` after entering gameplay means the registration ran before the plugin patched it.

Played on 2026-09-24 with the only weapon available, a pistol: step 2b passes for semi-automatic fire. The user tested RMB hold aim, LMB fire on press, holding RMB after shots, releasing RMB, first and third person, and native LMB without RMB, and reported it working perfectly. The load log showed the pad wrapper line. Pad samples: `calls=120` per 2 s (one call per 60 Hz frame), `std` equal to `calls` while armed, `fire=1` with `rearm=1` for single shots, `fire=8 tap=5 rearm=3` and `fire=8 tap=8` in two 2 s samples of rapid clicking, `lower` 1–3 per sample for RMB releases, `drop=0` throughout. Not yet played: automatic weapons, the knife, charge weapons, a CQC grab or interrogation with RMB held, and a controller.

Do not ship 2b beside MGSHDFix keep-aim (`0x35800E`); both own `aimingState`. Needs play: every weapon class, including full-auto and the `0x20000` weapon-flag class (0.7 s hold path at `0x357CF2`), and lock-on with R1/R2 pressure switching at `0x75E913`.

Controller: through Steam Input, LT would need L3 and RT would need Square. The local pad path does not use Steam Input, so local pad tests prove nothing for retail Steam Input.

### Phase 3. Verb layout

Only after the played movement pass in the remaining sequence. The 2026-09-22 sweep did not find crouch, roll, CQC, throat slit, or stalk.

2026-09-23: the pad bits are known (table above). Player-side edge tests: Cross press 10 sites, Cross release 4, Cross held 4, Circle press 3 (`0x34EBAA`, `0x377DDE`, `0x387165`), Triangle press 8, Square press 7. States are function pointers at `ctx+0x58`, reached from 28 `call [reg+0x58]` sites. `0x34226C` (Cross press, then two probes) leads to state `0x3411B0`, possibly roll, unproved. MGSHDFix `pressure_inputs.cpp` locates the CQC slit test at `0x387279` (Triangle `0x1000`). Order of work:

1. Play-check what Crouch Walk 0.2.1 already gives: crouch held while moving, and hold crouch for prone. Do not rebuild what it ships.
2. Stalk: test native LCTRL (key id 4, half stick) and LSHIFT (key id 6) first. This may need a rebind only, not a hook.
3. Add a state tracer on the `ctx+0x58` dispatch before any verb hook, and name crouch, prone, roll, CQC grab, and throat-slit states from a played log.
4. Then do remaps at the pad wrapper from step 2b. CQC grab on fire and throat slit on action need a grab-range condition that is still UNKNOWN. In Master Collection L3 is also "interrogate".

#### Phase 3 result, steps 1 and 2

Played on 2026-09-24 on 0.5.0 with Crouch Walk 0.2.1, QCamo 1.0.4, and the loader enabled, keyboard only. The user first reported "everything works as expected" and then answered each item directly:

- Tap SPACE (Cross) while standing: Snake crouches and stays crouched after release.
- Crouched, press W: he walks crouched. This is Crouch Walk's feature.
- Hold SPACE about 1 s while standing: he goes prone.
- Prone, tap SPACE: he gets up. Whether he returns to crouch or to standing was not asked.
- Running, press SPACE: he rolls.
- First person with the gun raised (RMB held): crouch and crouch-walk work.
- LCTRL (key id 4) while walking: slower and quieter.
- LSHIFT (key id 6) while walking: slower and quieter than LCTRL, a stalk. The user reads it as the mode for noisy ground (metal, water). Table 0 has LSHIFT write D-pad pressure words, which matches the PS2 D-pad stalk.
- A CQC grab worked. No throat slit, no automatic weapon.

Crouch, crouch-walk, hold for prone, roll, and stalk need no hook. Stalk on LSHIFT already matches the keyboard target. The Delta split is still missing: crouch and roll share SPACE (Cross). The session log has walk samples from 07:58:51 to 07:59:33 and pad samples with `fire=1 rearm=1` and `drop=0`.

#### Phase 3 step 3, static pass

Read-only on 2026-09-24 with the game closed. Executable SHA-256 `0d585dcc6a671be5d64d3d0a856c53f9ee0e58e7e4993f76dff29772c7a4bc80`. Method: `dumpbin /DISASM:NOBYTES /SECTION:.text`, then `grep` on the text. Addresses are RVAs.

- The 28 `call qword ptr [reg+58h]` sites are CORRECT: `grep -cE "call +qword ptr \[[a-z0-9]+\+58h\]"` gives 28, and 20 of them fall in `0x32F000`–`0x389FFF`.
- "States are function pointers at `ctx+0x58`" is PARTLY WRONG. There is no single player state. The player actor (global `0x1E16CC0`) holds a component tree: `[actor+0x58]` is the first node. Node layout from the node init at `0x3667C0`: `+0` next, `+8` previous, `+0x10` first child, `+0x20` flags (the low 3 bits make the walker skip the node), `+0x24` hash (low 24 bits), `+0x28` priority, `+0x30` init, `+0x38` exit, `+0x40` handler, `+0x48` a fourth callback. The walker `0x3661F0` calls `handler(actor, node, message)`. The player update `0x3304F0` walks the tree four times per frame, with messages `0x10`, `0x20`, `0x40`, and `0x80`. Each `call [node+0x58]` is one component calling its own current sub-state. So `node+0x58` is a separate state machine per component.
- The sub-states are set inline. In `0x32F000`–`0x38FFFF` there are 341 `mov qword ptr [reg+58h],reg` stores. 181 of them come straight after a `lea` of a code address, with 113 distinct targets. Stores reached through a shared `jmp` are not in that count. There is no single setter to hook.
- Crouch Walk's names fit this model. `ActMovement` `0x369800` is the handler of component hash `0x27E8110`, node size `0xD0`, registered at `0x36AE3E` on the player actor. `ActSquatStill` `0x3686B0` is one of that component's sub-states: `0x3677A9` loads it and `0x3677B0` stores it to `[rbx+0x58]`, right after a call to `0x3665C0`. There are 6 `lea` references to it.
- `0x3411B0` is loaded at `0x3422D9`, after a call to `0x3665C0`, then jumps to a shared tail at `0x3423C1`. That tail was not read, so the store is UNVERIFIABLE.
- The pad component (`0x357B30`, hash `0x1810F2B`, node size `0x68`) is registered on the same player actor global. So the actor in `rcx` of the 0.5.0 pad wrapper is the player, and it runs once per 60 Hz frame (`calls=120` per 2 s).
- The registration function `0x3663C0` has 70 direct calls. 68 of their node sizes resolved: two are `0x58`, too small to hold a `+0x58` field. Reading `+0x58` on those two nodes reads past the node.

Tracer design for 0.5.1, log only:

1. Hook nothing new. Inside the existing pad wrapper, after the original returns, walk the tree from `[actor+0x58]`: siblings through `+0`, children through `+0x10`, at most 128 nodes and depth 6.
2. For each node, record hash, handler RVA, and `+0x58` as an RVA. Record `+0x58` only when the node is not one of the two `0x58`-size components and the value points into `.text`. Identify those two handlers statically before the build.
3. Log changes only: a component added or removed (hash and handler), and a sub-state change (hash, old RVA, new RVA, frames spent in the old one, and the held pad word). Cap the log at 200 lines per 2 s and report the number dropped.
4. A marker key writes `mark N`. First check that the key is free in key tables 0, 1 and the custom table. The `.rdata` offset in this plan did not give a readable key table on this pass.
5. Limit: polling once per frame misses a sub-state that lasts less than one frame. The motion call `0x3665C0` that comes before most stores is a second channel, if one is needed later.
6. The file-scan test adds a check that the pad registration still names the actor global and that the walker at `0x3661F0` still reads `+0x40` and `+0x10`. The tracer runs only when all checks pass.

Naming rule: a sub-state gets a verb name only if it appears in at least 2 of 3 repeats inside that verb's marker window and not in the standing-idle window.

#### Phase 3 result, 0.5.1 tracer

Static checks before the build: all 70 registrations pass the player actor global, and the node init `0x3667C0` has one caller (`0x3663C0`), so they are the only way a component is added. The two `0x58`-size components have handlers `0x35FBB0` (hash `0xF36F4C`) and `0x382A00` (hash `0x974037`). The tracer skips them by hash. F9 was chosen for marks. It appears in neither default key table (`0xAC2230` and `0xAC23D0`; they sit in `.data`, file = RVA − `0x1200`). The game never reads its key slot `0x1D79BC0`. The packer tests only LWIN, RWIN, and F10 directly (`0x1127A4`–`0x1127CC`). A user rebind could still claim F9.

Build: the file-scan test found one match each for tree walk `0x32F94C`, walker `0x365620`, and hash lookup `0x365800` (file offsets). The player update's call resolves to the walker. The first hash-lookup pattern hit twice, so it was lengthened. `pack`, `disable`, `remove`, `add`, `enable --dry-run`, `enable`, and `verify` exited 0.

Played on 2026-09-24 from 09:08 to 09:14, keyboard. The user reported that it worked as expected, and that guards spotted them at one point. Every 0.5.0 load line was present, plus `tracer armed`. There were 10 marks, no `trace dropped` line, and no fault line. The tree held 25 components, all of them at load, and it never passed the depth cap. An earlier launch at 08:35 wrote only the load lines. That session never reached a pad call in gameplay, so it recorded nothing.

The marks did not line up one-to-one with the play script: mark 1 contains a crouch, mark 3 a prone cycle. So the names below come from the pad input on the transition frame, the frames spent in the old state (`n`), and repetition, not from mark windows. Held bits in the raw pad: Cross `0x4000`, Circle `0x2000`, Triangle `0x1000`, L3 `0x2`, D-pad down `0x40`, and `0x10000`/`0x20000`/`0x40000`/`0x80000` with WASD. `0x100000` and `0x200000` are not identified.

Movement component, hash `0x7E8110`, handler `0x369800` (Crouch Walk's `ActMovement`):

| State | Name | Evidence |
| --- | --- | --- |
| `0x3692C0` | standing still | state at load; every move ends here when input goes to 0 |
| `0x369030` | standing move (walk, run, and LSHIFT stalk) | entered from `0x3692C0` 30 times: 29 with a WASD bit, once with D-pad down `0x40` (LSHIFT); left when input goes to 0 |
| `0x3686B0` | crouched (Crouch Walk's `ActSquatStill`) | entered from `0x3692C0` 9 times, with Cross held or after a tap; static store at `0x3677B0` |
| `0x368AB0` | crouch to prone | entered from crouch with Cross held (`0x204000`) or after crouching; lasts exactly 40 frames all 3 times |
| `0x367370` | prone | follows `0x368AB0` all 3 times; lasts 27, 27, and 137 frames |
| `0x3677D0` | prone to crouch | follows prone all 3 times; lasts exactly 35 frames; ends in `0x3686B0`, so getting up from prone lands crouched |
| `0x367FA0` then `0x367E40` | roll, then roll recovery | 4 times, all from `0x369030`: once at 09:12:47, and 3 of 3 under mark 12 at 09:21:52, later in the same game run, when the user did F9 and then rolled 3 times. `0x367FA0` lasts exactly 39 frames and `0x367E40` exactly 34 every time, then standing still. The transition frame shows only the direction bit (`0x10000`), never Cross |

No separate crouch-walk state was seen. Crouch Walk moves the body inside `0x3686B0`. Entering first person changes no traced state: the first-person component (hash `0xCEB2AE`, handler `0x38B880`) is always in the tree and its `+0x58` stays 0. The walk samples at 09:12:01, 09:13:27, and 09:14:21 show first person was used.

CQC component, hash `0x80DD2A`, handler `0x3877E0`:

| State | Name | Evidence |
| --- | --- | --- |
| `0x3865D0` | CQC grab start | entered 9 times with Circle held (`0x102000`); lasts exactly 95 frames each time |
| `0x386560` | CQC grab start, second form | 2 times (marks 20 and 21), each entered from `0x386CF0` right after a throw with Circle held; exactly 108 frames, then `0x386A10`. Which grab it is (front or re-grab) is UNKNOWN |
| `0x386A10` | holding an enemy | follows every grab start (11 of 11) |
| `0x386910` | throat slit | 4 times, 3 of 3 under marks 13–15, always from `0x386A10` with Triangle added (`0x103000`); MGSHDFix places the slit test at `0x387279` (Triangle) |
| `0x386E60` | throw input (Circle plus a direction while holding) | 7 times, entered with Circle plus one of four WASD direction bits (`0x112000`, `0x122000`, `0x142000`, `0x182000`); 5–62 frames. It led to `0x386CF0` 4 times, `0x386C00` once, and back to `0x386A10` twice (both in the first session) |
| `0x386CF0` | throw | 4 times after `0x386E60` (marks 16, 19, 20, and once in the first session) |
| `0x386C00` | throw, other form | once (mark 21, Circle plus `0x40000`). An earlier `0x40000` throw went to `0x386CF0`, so the direction does not decide it alone. UNKNOWN |
| `0x386990`, `0x386640` | unnamed | once each |

A state stays in `+0x58` after the move ends, until the next CQC move. For example, `0x386910` held for 390–582 frames after each slit. So the frame count of the last state in a chain is not the length of the move.

Second traced session in the same game run, 09:34–09:35: marks 13–15 are one slit each, mark 16 is a throw, mark 17 covers the user's death, and marks 18–21 are the grab and 3 throws. The user reported "3 slits, died, then 3 grabs and throws". The throw at mark 16, before the death, was not in that report. Around the death, component hash `0x927EE8` (handler `0x334F40`) cycled through `0x3345D0`, `0x334A20`, `0x334240`, `0x333C80`, `0x333B90`, `0x334160`, and `0x333F00`. Hash `0xA3F2C8` went to `0x32EE70` and then to `0xFFB0`. At 09:35:10 the same two components as at 09:13:51 were removed and re-added, and all states reset. None of these are named. The log has 329 lines and no dropped, fault, or depth lines.

Weapon component, hash `0xCF5135`, handler `0x37A150`: `0x373560` is entered 0.1–0.2 s before each CQC grab (3 times, with Circle held). `0x3741E0` is the gun attack state machine named in "Ready and fire". It was entered 4 times with RMB held (`0x302002`), and its state returns to 0 when the gun lowers.

Other: hash `0x724F29` (handler `0x388BC0`) went `0x388A60` to `0x388A30` once, with no name. At 09:13:51, components `0x2429F5` and `0xA9DDB1` were removed and added back in the same frame, and the CQC, movement, and weapon states reset. This lines up with the guards spotting the user; the cause is not proved.

For step 4: crouch and roll are both Cross. From standing still (`0x3692C0`), Cross goes to crouch. From moving (`0x369030`), Cross goes to roll (PARTLY WRONG: only above stick magnitude `0x96` with status `0xE5` clear, otherwise crouch; see "Phase 3 step 4a, 0.5.2"). Cross was seen held on some crouch entries (`0x204000`) and on none of the 4 roll entries. So roll may fire on a short tap or on release; that is not proved. A Delta split in the pad wrapper can key off `0x369030` versus `0x3692C0` in the movement node. CQC is on Circle: grab on Circle, slit on Triangle while holding, throw on Circle plus a direction while holding. Any CQC remap must also know when the player is holding (`0x386A10`).

Keyboard: right mouse aims, left mouse fires and CQC-grabs, crouch and roll are different keys, Shift stalks. Controller: LT aims, RT fires and grabs, crouch and roll are different face buttons, LB stalks. Exact face buttons follow Delta New Style once the mechanisms exist: Circle/B crouch, Triangle/Y roll and evade, Cross/A action.

Controller input on this port goes through Steam Input. This executable does not import XInput. The local pad path does not use Steam Input, so behavior on a clean Steam install is unverified. Prove the pad on the install that will ship.

#### Phase 3 step 4, static pass

Read-only on 2026-09-24, game closed, same executable hash, from the `dumpbin /DISASM:NOBYTES` text. Addresses are RVAs.

- The pad component points `[actor+0x7F8]` at the copy `0x1E2D684` (`0x357C17`). It derives actor pressed and released from the held word it copied: pressed = new AND NOT old, released = old AND NOT new (`0x3587DB`–`0x358808`). An injected button only needs its held bit; the edges follow.
- The pad component is the first node the walker visits (`h=810F2B` is the first `+` line of the trace log, before movement `7E8110` and CQC `80DD2A`). A value the wrapper writes to the actor after the original returns reaches movement and CQC in the same frame.
- Movement code (`0x366000`–`0x36BFFF`) reads no raw pad and no copy directly. It reads `+0x7E8` 8 times, `+0x7EC` 4 times, `+0x7E0` 22 times, and never `+0x7F0` (grep of `1E2C5..`, `1E2D6..`, and the actor offsets over that range).
- Roll fires on Cross release, not press: CORRECT. In the move sub-state `0x369030`, `0x3690B3` tests released Cross (tap); otherwise held Cross counts up in `node+0x8E` until `(0x31 + d) / d` with `d = [0x1D79810]` (hold). Either one rolls (`lea 0x367FA0` at `0x369173`) only when status `0xE5` is clear and `[actor+0x7E0] > 0x96` (`0x36911B`–`0x369139`). Otherwise the move sub-state crouches (`0x36917F`–`0x3691D1`). The same tap/hold helper is Crouch Walk's `GetButtonHoldingState` `0x36A590`; the stand sub-state uses it and crouches on either result (`0x3693AB`–`0x369423`).
- The movement init (`0x36977F`) picks prone on status 3, crouch on status 2, else stand. The stand sub-state enters move at `0x3694F9`. Crouch to prone is stored at `0x368A5D`; prone to crouch and prone at `0x3672BA`/`0x3672CA` (a twin function at `0x366F07`/`0x366F17` stores the same pair).
- `0x386C00`: PARTLY resolved. While holding, Circle with a stick direction throws to `0x386C00` when the stick angle is within `0x38E3` (about 80°) of Snake's facing (`+0x162`), else to `0x386CF0` (`0x38717B`–`0x3871B1`). So the throw follows facing, not the key. Which animation is which is UNVERIFIED.
- `0x386560`: PARTLY resolved. The grab start at `0x387513`–`0x387525` picks `0x386560` over `0x3865D0` when the angle test at `0x3874B3`–`0x3874C2` is at most `0x4000` (90°), so it is the grab from the other side. Which side is UNVERIFIED.
- The slit test at `0x387279` needs the Circle pressure float at or above a threshold and a Triangle press in the copy (`0x1E2D698`). Slit is "keep Circle held, press Triangle".
- Keys: the key array is `0x1D799E0 + VK * 4`, nonzero while down (the lookup `0x1123E0` reads the Enter and Backspace slots). The binding getter `0x34BE0(layout, key id, slot)` reads table 0 (`0xAC2230`), table 1 (`0xAC23D0`), or the custom table (`0x1E1FE20`) by the layout global `0x1D799D8` (getter `0x34D20`). Rows are 16 bytes: two VK slots, then two more dwords. Table 0 binds Cross (key id 9) to SPACE and leaves C free. Table 1 binds C as the second key of key id 16 (R1, first person).

Delta PC New Style keys, per Prima's controls page: crouch C, evade Space, CQC and attack left mouse, aim right mouse, action F, stalk left Shift, walk left Ctrl. A second site's summary contradicted itself. Confidence medium. Action stays on E here because F is first person in Master Collection.

#### Phase 3 step 4a, 0.5.2

Approved on 2026-09-24: C crouches, SPACE rolls, SPACE does nothing while standing, crouched, or prone, 4a ships before 4b.

The wrapper resolves six movement sub-states by pattern at load (stand, move, crouch, crouch to prone, prone, prone to crouch). Prone must agree between two sites, and the roll gate must sit between the move and stand sub-state addresses. It also resolves the key array, the layout global, and the three binding tables. Each frame it reads the movement node's `+0x58`, the live keys bound to Cross, and whether C is bound to any key id. If C is bound, the split turns off and logs it.

- C press that starts in one of the six sub-states: Cross is held in the raw pad while C is held (pressure `+0x12` set to `0xFF`). Native then does tap-to-crouch and hold-for-prone. In the move sub-state, while C is held and on its release frame, `[actor+0x7E0]` is capped at `0x96` after the pad component runs, so the move sub-state crouches instead of rolling. Crouch Walk carries the crouched move.
- Roll-key press (Cross binding, SPACE by default) that starts in one of the six sub-states: the physical Cross is hidden until the key is released. In the move sub-state, with a nonzero magnitude and status `0xE5` clear, the press becomes one frame of Cross held, then a release, with the magnitude raised to `0x97` on the release frame if it was at or below `0x96`. So the roll fires one frame after the press, from a walk or a run.
- Presses that start in any other sub-state (ledge, hang, climb, CQC, roll recovery) keep the native Cross until released. A controller Cross is never touched.
- The raw press and release bits of Cross are rebuilt from what the wrapper delivers, and the raw pad is restored after the component returns.

File-scan test: one match each for `moveinit` `0x368B7F`, `movesite` `0x3688DB`, `rollgate` `0x368509`, `s2p` `0x367E58`, `p2s` `0x36669A` (lengthened; the short form hit the twin too), `keylookup` `0x1117E0`, `bindfn` `0x33FE0` (file offsets). The six sub-states resolved to the traced RVAs. Table 0 Cross is `0x20` and C is free; table 1 has C on key id 16.

The log adds `split armed: ...` at load, one `keys layout=N roll=0x20/0x0 crouch=C` line when the layout is first read or changes, and pad counters: `crouch` (C presses taken), `clamp` (frames capped at `0x96`), `roll` (roll presses turned into a tap), `rollfix` (release frames raised to `0x97`), `rollmiss` (the move sub-state was gone or the stick was 0 on the release frame), `hide` (roll-key presses hidden while standing, crouched, or prone).

Play test, keyboard. F9 before each block, 3 repeats each, then close the game:

1. Stand, tap C: crouch (`0x3692C0` to `0x3686B0`). Tap C again and note the result.
2. Stand, hold C about 1 s: prone (`0x368AB0`, then `0x367370`). Tap C to get up.
3. Run (W), tap C: crouch, never `0x367FA0`. Keep W held: crouch-walk.
4. Slow walk (LCTRL + W), tap C: crouch.
5. Run, tap SPACE: roll (`0x367FA0`, then `0x367E40`).
6. Slow walk, tap SPACE: roll; `rollfix` rises.
7. Stand, tap and hold SPACE: nothing happens; `hide` rises.
8. Crouched, then prone, tap SPACE: nothing happens.
9. First person with the gun raised (RMB): C crouches, LMB fires, `fire`/`rearm` behave as in 0.5.0.
10. Anywhere SPACE had another use (drop from a ledge, climb down): it still works.

Pass: every block matches, and `rollmiss` stays near 0. The known cost: C tapped while running crouches on release (the native tap rule), and the run slows to walk speed for the frames C is held.

#### Phase 3 step 4b, static pass

Read-only on 2026-09-24, game closed, by one agent that was given this plan. The key setter was rechecked in the main session. Addresses are RVAs.

- A Circle press does not start a grab directly. It starts the weapon component's strike sub-state `0x373560` (stored only at `0x377EAF` and `0x379D58`; `grep -c "\[0000000140373560h\]"` = 2). At animation event `0x0A` of the first strike, with weapon flag `0x8000` at `[ctx+0x20]` and Circle still held in `+0x7E8`, it sets status `0x6D`, the grab request (`0x373763`–`0x3737B3`, the only setter). Otherwise the strike stays a punch or combo.
- The CQC component's poll callback `0x3872C0` sees `0x6D` (`0x3876F6`), checks a status list, claims the contact probe (`0x360AD0(actor, 1, 2000.0)`), and writes `+0x6E = 1` (`0x38778D`). The next poll takes the first contact tagged `0x13` and picks `0x3865D0` or `0x386560` by the angle test. With no contact it resets, and the punch goes on.
- Reach is only a contact from the probe shape at `actor+0x100` plus a path check `0x360CA0`. There is no stored "enemy in reach" value before the press: contacts are dropped (`0x387B85`) unless CQC owns `+0x696`, and CQC takes it only after `0x6D`. So a pad wrapper cannot know in advance whether an LMB press will grab.
- The weapon switch at `0x3874E7` (table `0x3877B8`) only picks the animation set. The weapon routing in `0x379590` decides whether Circle can reach `0x373560`: ids 0, 3–7, 25/26, and ids of 15 and above that are not special can; ids 8, 9–14, and 18 use other strike states; ids 15/16/17/30 go straight to the gun attack state. Weapon names per id are UNVERIFIABLE statically (the item table is filled from game data).
- Square never starts CQC natively: the only `0x6D` setter needs `node+0xD0` held, and both entries to `0x373560` set it to Circle `0x2000`.
- The research note "a Circle press reaches the grab through the pressure copy" is WRONG: `0x1E2D6B8` holds per-button hold timers (`0x35809F`–`0x3580BD`) and feeds only the slit test at `0x38726F`.

So "grab on the fire input" costs something. Either LMB without RMB always sends Circle (it grabs when the game finds a contact and punches otherwise; LMB-only hip fire and knife slashes go), or the grab stays on H and LMB only takes over once the hold has started. A proximity test would need a hook on `0x387B10` and a trace first.

Decided 2026-09-24: option A. With right mouse not held and a weapon whose `[ctx+0x20]` has `0x8000`, left mouse sends Circle for as long as it is held and Square is hidden. The game then grabs when it finds a contact and punches otherwise. Hip fire from left mouse alone goes; right mouse plus left mouse still aims and fires. The proximity variant is in "Optional backlog".

#### Phase 3 step 4b, 0.5.3

Built on 2026-09-24 on top of 4a. 0.5.2 was never played, so this one session tests both. Three rules were added to the pad wrapper:

- **CQC on the fire key (option A).** The fire keys are read from the live binding of key id 10 (Square), left mouse by default. A fire-key press starts CQC mode when all of these hold: right mouse (L3) is not held, the weapon context exists with `0x8000` set, and either status `0x6C` (CQC hold) is set or status 1 is set with `0xDF` clear. While the key is held, the raw pad carries Circle held and hides Square, and the release is hidden too. The filter sets Circle pressure itself (`0xC7` at `0x357C28`). With a gun on the filter's standard path, a Circle press lowers the gun and reaches the game one frame late (`0x357D76`–`0x357E67`). That is the same as a native H press. Each fire press writes a trace line `fire press id= flags= rmb= hold= cqc=`, so the weapon ids and the `0x8000` flag get recorded in play.
- **Knife (ids 1 and 2).** From a right-mouse press until its release, L3 is hidden, so right mouse plus left mouse is a plain Square (slash). The exception: while status `0x6C` (holding) is set, L3 passes, since L3 is the native interrogate there. Left mouse alone with the knife is CQC under the rule above. This rule assumes ids 1 and 2 are the knife; that is UNVERIFIED (the static routing shows only that their Square goes to `0x376960`).
- **Automatic guns** (flag `0x800`, standard path). With right mouse held and the gun up, a left-mouse release would clear `aimingState` and lower the gun (`0x357E6C` → `0x35800E`). Instead the wrapper hides that release and holds Square at pressure `0x77`, one below the automatic fire threshold `0x78`. A new left-mouse press brings the real pressure back. When right mouse is let go, the held Square is released so the filter lowers the gun. Designed without play.

New pad counters: `cqc` (fire presses turned into Circle), `knife` (right-mouse presses with the knife), `auto` (automatic hold sessions), `autolower` (their end). The `keys` line now also shows `fire=0x1/0x55` (left mouse and U in table 0).

Combined play test, keyboard. F9 before each block, 3 repeats each, then close the game:

1. The 10 blocks of "Phase 3 step 4a, 0.5.2".
2. Pistol out, right mouse not held, walk up behind a guard, hold left mouse: grab (`0x3865D0` or `0x386560`, then `0x386A10`). Keep holding, add E: slit. Grab again, left mouse plus a direction: throw.
3. Left mouse near nobody: punch. Click fast: combo. No shot.
4. Right mouse + left mouse: aims and fires as in 0.5.0. Right mouse held after a shot keeps the gun up.
5. Knife: left mouse alone near a guard grabs, away from guards it punches. Right mouse + left mouse slashes.
6. While holding a guard, right mouse: note what it does (interrogate is expected).
7. H still does CQC as before.
8. First person, right mouse not held, left mouse: note the result.
9. If an automatic gun is available: right mouse, hold left mouse (continuous fire), let go (the gun stays up, no extra shot), fire again, let go of right mouse (the gun lowers). `auto` and `autolower` rise.

Pass: every block matches. Every `fire press` line with `cqc=1` has `rmb=0`. With a weapon that can CQC, left mouse alone never shoots; weapons without `0x8000` (two-handed) keep the native left mouse.

#### Phase 3 result, 0.5.3

Played on 2026-09-25 from about 10:57 to 11:10, keyboard, 25 F9 marks, no dropped, fault, or depth lines. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.5.3-play.log`. The user reported every block passing except item 6 (right mouse while holding a guard). Notes: a second C tap stands Snake up. In first person, left mouse alone does the same CQC as in third person, and with the knife equipped the knife is used.

- Keys line: `layout=0 roll=0x20/0x0 fire=0x1/0x55 crouch=C`.
- Pad totals over the session: `roll=11 rollfix=2 rollmiss=0`, `crouch=49 clamp=63 hide=23`, `cqc=173`, `knife=3`, `auto=4 autolower=4`, `fire=28 tap=3 lower=28 rearm=34 drop=9`.
- Fire presses name the weapons: id 0 flags `0x8000` (empty hands), id 1 flags `0x8001` (knife; so the knife is id 1), id 5 flags `0x8D105` (pistol, semi-automatic), id 11 flags `0x590B` (automatic, bit `0x800`, no `0x8000`, so left mouse stayed native). Every `cqc=1` line has `rmb=0`.
- Item 6 failed. While holding a guard, the user keeps left mouse held (CQC mode, Circle). Right mouse raised the pistol (weapon state `0x3741E0` at 11:07:06.865, held `0x308002`). Releasing right mouse did not lower it: the gun stayed up until the next left-mouse press at 11:07:16. Cause: the fire rules ran before the CQC rule and saw the raw Square of the held left mouse. The lower rule needs Square clear, so it never fired, and right-mouse presses in that state were dropped as "already ready" (`drop=1` samples at 11:07:10 and 11:07:14).

Fix in 0.5.4: the CQC rule runs first, so while left mouse drives Circle the fire rules see no Square. Retest item 6, plus a quick check of CQC, aim and fire, and the knife.

#### Phase 3 result, 0.5.4

Played on 2026-09-25. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.5.4-play.log`. The user reported the retest passing (holding a guard, right mouse raises and lowers the pistol) and three new findings:

1. While holding a guard with right mouse held, a shot with left mouse and then letting go of right mouse (left mouse still held) released the guard. Cause: to shoot, left mouse is pressed again with right mouse held, which is a fire press, not CQC, so Circle had dropped. The game keeps the hold while aiming, then drops it when the aim ends with Circle up. Fix in 0.5.5: while holding (`0x6C`) with right mouse held and Circle held, Circle stays held (counter `keep`). When right mouse is let go, the hold carries on only if left mouse is still held, and that press turns back into CQC.
2. No CQC while crouched. The weapon code enters the strike sub-state only with status 1 set (`0x377E5D`, `0x379CD0`), and status 1 is set by the stand and move sub-states, among others (at least 22 literal SET sites; corrected 2026-09-25), but not by the crouch. So native Circle (H) does not strike from a crouch either. This is not a mod regression; not played with H to confirm. See "Optional backlog".
3. Holding SPACE through a roll no longer ended in prone, which native Cross did. At the end of the roll, the roll sub-state `0x367FA0` goes to `0x367CD0` when Cross is held (`0x368377`–`0x368433`), else to recovery `0x367E40`. The one-frame tap had already released Cross. Fix in 0.5.5: once the move sub-state has turned into the roll sub-state and the roll key is still held, Cross is held again until the key is released (counter `rollprone`). The roll sub-state is resolved from the `lea` at `+0x6A` of the roll-gate pattern (file-scan test: `0x3673A0`).

#### Phase 3 result, 0.5.5

Played on 2026-09-25. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.5.5-play.log`. The user reported that the guard stays held after a shot, that the guard is let go when both buttons are released, and that run plus a held SPACE ends in prone while a tap recovers. Pad totals: `keep=7 rollprone=9 roll=9 hide=15 crouch=18 cqc=78`. Roll into prone traced 4 times (`0x367FA0` to `0x367CD0`, then prone `0x367370`).

- Found: SPACE while prone and crawling stood Snake up. Moving while prone is its own sub-state `0x366FC0` (traced 11:48:59: prone `0x367370` to `0x366FC0` to prone-to-crouch `0x3677D0`), and neither it nor roll-to-prone `0x367CD0` was in the set where SPACE is hidden, so SPACE there was native Cross. Fix in 0.5.6: both are added, plus the twin crawl sub-state `0x366C40` (the same code stores it; not seen in play). Patterns: `crawl` (file `0x36637F`), `crawl2` (file `0x366732`), and `rollprone` (file `0x3677F7`, checked to sit inside the roll sub-state).
- Crouched, with a gun, left mouse aims and fires. CQC is not possible from a crouch (status 1 clear), so the fire key falls through to native Square. With empty hands nothing happens. Both are native.

#### Phase 3 result, 0.5.6

Played on 2026-09-25 from about 12:00 to 12:02. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.5.6-play.log`: 4 marks, no fault or dropped lines. The user reported every item passing: a held SPACE while running rolls into prone, and SPACE then does nothing; SPACE while crawling does nothing and C gets up; a tapped SPACE rolls and recovers; SPACE while standing or crouched does nothing. Pad totals: `hide=75 roll=6 rollprone=6 crouch=21 clamp=28`. Step 4 passes, and Phase 3 is done.

## Play checks, 0.5.6

Played on 2026-09-25 on the installed 0.5.6 (generation 69, SHA-256 `76c1ce7c…d53cd`; `mgs3mod verify` reported `compatible: true` before play). Two runs, both launched with `-ctrltype KBD` (prompt icons only). Logs kept locally: `work/fpv-move-trace-20260924/fpvmove-0.5.6-checks-AB.log` (16:21 to 16:32, 2 marks, 552 lines) and `fpvmove-0.5.6-checks-C.log` (16:38 to 16:43, 11 marks, 414 lines). Neither has a fault, dropped, or depth line (`grep -c -i -E 'fault|dropped|depth'` = 0 on both).

Keyboard, charge weapons and lock-on (first run). The user reported every item passing with the automatic rifle and grenades.

- Two `0x20000` weapons were pressed: id 21 flags `0x20015` (enters weapon sub-state `0x373A20`) and id 19 flags `0x20213`. The user threw grenades; which grenade is which id is not in the log. Neither has `0x8000`, so the CQC rule never takes their fire key (every `fire press` line for them has `cqc=0`), and left mouse stays native. Other `0x20000` weapons (sniper rifle, rocket launcher, if any) were not tested.
- Automatic id 11 `0x590B`: `auto=1` and `autolower=1` repeat; the gun stays up between bursts.
- Lock-on is PS2 L1, held bit `0x400` (key M in table 0; MGSHDFix calls it `MGS3_LockOn`). In MC it fixes Snake's facing and lets him strafe. It works with the mod: moving with L1 and right mouse held shows `held=0x200402` and `0x220402` with the move sub-state `0x369030`, and a shot at 16:31:58 logged `fire=1 rearm=1` with the gun staying up. Static note: `0x75E913` is not lock-on; it reads R2 pressure (R1 with the filter off) from the pad at `0x1E2C500` and blends four float pairs.

Xbox Series X controller, wired, without Valve Steam Input (second run). The user reported every item passing.

- L3 click: raise, then lower on release (`0x3741E0` then `0x0`, `lower=1` per click, marks 1 and 2). This is a change from native, where L3 toggles ready: the fire rules read pad L3 as right mouse, so on a pad the aim is held on L3.
- L3 held plus X (Square): `fire=1 rearm=1` per shot, 5 shots, gun up for 8 s (mark 3).
- X alone: native raise on press (`held=0x8000`), release fires.
- B (Circle): native strike `0x373560` and grab states in component `0x80DD2A` (`0x3865D0`, `0x386A10`, `0x386E60`, `0x386C00`); `cqc=0`, since the CQC rule reads the keyboard fire key only.
- A (Cross): native roll, roll into prone, crawl, and get-up (mark 6); the split reads keyboard keys only.
- Knife, L3 plus X: `knife=1` per press and slash sub-states `0x376960`/`0x376A20` (marks 7 and 8).
- L3 while holding a guard (interrogate): not played; the user skipped it. Expected to match keyboard right mouse while holding, since both are L3 in the raw pad: the knife rule leaves L3 alone while status `0x6C` (hold) is set, and with the pistol right mouse while holding raised and lowered the gun on 0.5.4 ("Phase 3 result, 0.5.4"). With a gun this is the held-aim rule, not the native L3 toggle.
- Automatic, L3 plus held X: `auto=1`, `autolower=2` (mark 9).
- Lock-on on LB with L3 and X while moving: `held=0x80402`, `0x28402`; `auto=1 autolower=1`, `drop=1` (marks 10 and 11). R1 first person with L3 also raised and lowered (`held=0x802`).
- The CQC rule and the roll/crouch split never act on a pad: they read the keyboard bindings of Square and Cross. Pad face buttons stay native. The fire, automatic, and knife rules read pad L3, so they do act on a pad.

Not proved: the controller under Valve Steam Input on a retail install. The local pad path does not use Steam Input.

## Step 5a, shoulder aim on lock-on

### Static pass, 2026-09-25

Read-only, game closed, on `dumpbin /DISASM:NOBYTES /SECTION:.text` of the pinned executable. Direct L1 tests on the actor pad fields (`test dword ptr [..+7E8h|7ECh|7F0h],400h` and the byte forms): 9 sites in 8 functions, 8 on held and 1 on press. (Corrected 2026-09-26: 4 more L1 tests use a register mask set to `0x400` per weapon: `0x359613` (pistol component, id 7) and `0x359F6E`, `0x35A444`, `0x35A612` (rifle component, ids 9, 11, 12). They are the first-person iron-sight switch; see "Research for the 0.8.0 batch".) The 21 `and reg,400h` sites have no load from `+0x7E8`/`+0x7EC`/`+0x7F0`/`+0x7F8` or the pad globals in the 8 instructions before them, and no L1 test sits within 10 instructions of the 327 `+0x7F8` loads. The 188 `bt reg,0Ah` sites were not checked.

| Site | Owner | L1 effect |
| --- | --- | --- |
| `0x369C12` | movement handler `0x369800` (hash `0x7E8110`) | With a gun raised (ANY `0x2E`, `0x27`) and L1 held or status `0x30`: the strafe path (`0x36A180`, velocity into `+0x620`/`+0x628`; corrected 2026-09-26: `0x36A180` only refuses the strafe for weapon id 8 with a gun raised, and the strafe motions are picked by `0x36A690`). It also needs status 1 (standing or moving), so a crouch does not strafe. |
| `0x37A881` press, `0x37A891` held | `0x37A790`, 11 call sites in 9 functions (corrected; this row first said 5 sub-states) | An L1 press sets the facing target `+0x800` once from the stick (`+0x7E2`) or the turn target `+0x16A` and sets status `0x31`. Otherwise it searches for a target (`0x73FBA0`) and, with one at `[node+0x88]`, writes `+0x800` toward it every frame (`0x37AB00`, "the aim follows the enemy"). Corrected: the stick-follow block `0x37A8DD` is dead code, so +0x800 never follows the stick here; movement turns Snake to the stick when it does not strafe. |
| `0x37BFC4` | `0x37BD30` | L1 or status `0x103`, standing, not prone: stance motion `0x42`. |
| `0x36266E` | `0x362600`, 5 weapon callers | For weapon ids 15 to 17 a held L1 counts as a held aim. |
| `0x35BEDB` | component `0x851917`, the item-table component of weapon id 8 (flags `0x108`, table entry at file `0xAC6978`; the flags dword is the 4 bytes before it, file `0xAC6974`: id k's flags sit at file `0xAC67B4 + k * 0x38`) | With the gun raised, L1 changes a rate passed to `0x372FF0`. Effect not identified. |
| `0x35CDB6` | component `0x8516FD`, weapon id 17 (flags `0x10311`, file `0xAC6B70`) | L1 sets `+0x738` to 4.0 instead of 2.0, status `0x12A`, and another motion. |
| `0x34527E`, `0x3458FB` | stage actors `0x3717FC`, `0xAF0733` (stage tables at file `0x988808`, `0x989C88`, and more) | Not identified; never seen in a trace. |

Weapon names by id are not proved; the item table gives flags only. No L1 site is in the CQC grab or strike states, the first-person component (it reads R1 at `0x38BC2E`), or the crouch, prone, and roll states. The wall-press component was not identified, so no wall reader is not proved. (Identified 2026-09-26: hash `0xD9728F`, handler `0x34F950`, registered at `0x351030`. Of the 24 reads of `+0x7E8/+0x7EC/+0x7F0` in `0x34A000`..`0x352000`, 10 test immediates `0x300`, `0x4000`, `0x2000`, `0x50050`, none `0x400`; the other 14 load the word or test a register mask and were not traced.) No literal setter of status `0x30` was found. The raw pad readers `0x212BA5` (`[0x1E2C574]`) and `0x30957F` (`[0x1E2C500]`) see the real input, since the wrapper restores the raw pad after the pad component.

### 0.5.7

In `pad_step`: while RMB (L3) is held with a standard gun up (`aimingState != 0`), weapon id not 15 to 17, and R1 not held, the wrapper holds L1 with a press on the first frame and a release on the last. A real L1 already held gets no added press and keeps its own release. Counters `lock` and `unlock`, trace lines `lock on id=` and `lock off id=`. Installed as manager generation 73, `fpvmove.asi` 161,280 bytes, SHA-256 `a59eec13770a7233d071675bf48b5c4cd50af78414acf7e5525a4026563cee64`; `mgs3mod verify` reported `compatible: true`.

Played on 2026-09-25 from 17:06 to 17:17. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.5.7-play.log`: 5 marks (all before 17:12:24), no fault, dropped, or depth line. Pad totals: `lock=123 unlock=123 fire=60 rearm=54 auto=9 autolower=9 knife=4 cqc=25`. The user reported:

- Holding RMB with the pistol keeps Snake's facing, and A, D, and S strafe or back up with the gun up. The user prefers this to 0.5.6, where WASD turned Snake while aiming.
- The mouse turns only the camera; the aim stays where Snake faces. Wanted next: the mouse turns the aim (5b).
- RMB pressed while running faces the run direction.
- With an enemy in view the aim follows him.
- Found: after each pistol shot the lock goes off and Snake turns to the last WASD direction. The log shows `lock off` and a new `lock on` about 130 ms after every semi-automatic shot (54 re-readies). After a shot `aimingState` is 0 until the re-ready 8 frames later, the rule then drops L1, and the new L1 press takes the facing from the stick again (`0x37A917`).
- Automatic id 11 felt right. Lock went off several times mid-burst (17:12:28 to 17:12:54); not explained, retest.
- M alone resets the camera to Snake's facing and does not strafe. 0.5.7 does not touch L1 unless RMB is held, so this is native; whether M alone strafed on 0.5.6 was not played.
- Crouched aim neither strafes nor locks, the same as before. The strafe path needs status 1 (`0x369C55`), which a crouch clears. Not a 0.5.7 regression.
- Wall aim: no issue noticed.
- First person was skipped as intended: `held=0x210802` (R1 and L3) at 17:11:45 to 17:11:52 with 13 aimed shots and no `lock` line. The knife (id 1) and a grenade (id 19) logged no `lock` line.
- Not played: holding a guard with RMB (`keep=0`), weapon ids 8 and 15 to 17 (not in the log). Pad: the user reported the pad as expected but did not note times; lock flickers at 17:07:43 to 17:07:44 and 17:16:38 to 17:16:45 are not explained. Retest both.
- Wanted on the pad: aim on LT (L2) and fire on RT (R2), as in Delta, instead of L3 and X.

Log reread after the user's answers: every automatic-rifle `lock off` from 17:12:28 to 17:12:54 falls on a frame where held loses L3 (`0x2`) with `lower=1`, so right mouse was released there. The flicker at 17:07:43 to 17:07:44 is right mouse tapped 6 times (`lower=6` in 2 s). The one at 17:16:38 to 17:16:45 is the shot re-ready drop (`fire=3 rearm=3`) plus one weapon swap to id 0 (component `9D5DF5` gone at 17:16:38.716).

### 0.5.8

The lock stays on from the frame the gun is up until RMB is released (a standard gun and no R1 still required), so the re-ready after a semi-automatic shot no longer sends a new L1 press. The log gets context lines instead of F9 marks: `ctx weapon`, `ctx gun up/down` (down after 20 frames), `ctx first person`, `ctx holding`, `ctx real L1 ... src=key|pad`, `ctx aim press src=mouse|pad`, `ctx stance`. Installed as generation 77, `fpvmove.asi` 162,304 bytes, SHA-256 `d2dc87ff87ddc245e74739584a4d2b3328ed962741b94b2c0dd9c03e5e50904d`. Not played; superseded by 0.5.9.

### Research for 5b, the pad layout, and a crouched strafe, 2026-09-25

Six read-only agents (four Opus, two Sonnet), each Opus design checked by an independent Opus refuter. Full reports are kept locally under the session scratchpad, not in git. Verified results:

- Yaws are 16-bit words, `0x10000` per turn, 0 along +Z and `0x4000` along +X. `+0x162` is the current yaw, `+0x16A` the turn target the control update `0xB5DE0` eases toward, and `+0x800` the facing target movement uses while strafing (status `0x11`).
- Frame order: the player update `0x3304F0` sends messages `0x10`, `0x20`, the control update, `0x40`, `0x80`. The pad component runs first (priority 5, message `0x10`), then movement, then the weapon. A write in the pad wrapper therefore comes before movement and before the control update. (Added 2026-09-26: before message `0x10`, an arbiter `0x3665F0` at `0x330547` calls idle nodes with message 4; the weapon router `0x379590` and the CQC poll `0x3872C0` run there, so they see the previous frame's pressed buttons and statuses.)
- The third-person camera object is the global at `0x1D8A068` (set in camera init at `0x21232C`). Its orbit yaw is the word at `+0x338`, turned by the mouse and the right stick in `0x213010`. The stick direction `+0x7E2` is `atan2(stick) + pad heading 0x1E2C582`, with the pad heading = camera heading - `0x8000`, so a camera yaw written into `+0x800` makes Snake face where the camera looks. The camera reads only the raw pad, so the mod's synthetic L1 never changes its mode. Modes 1, 4, and 5 (front and fixed cameras) are excluded.
- The mouse is `GetCursorPos` with recentering (`0x116390`), no Raw Input or DirectInput. Third-person vertical aim is automatic: horizontal without a target, the target's position with one.
- The auto-target in `0x37A790` can be switched off by status `0xE3`: one TEST (`0x37A829`). With it set the helper drops the target and skips its `+0x800` writes. It is a per-frame CQC flag, not free to set (see the correction below).
- Crouched strafe: the only failing strafe gate while crouched is TEST(1) at `0x369C55`; crouch sets status 2. Setting status 1 was rejected (33 or more readers).
- Controller: the game reads it only through Steam Input (action set `NHTCommonSet`) in `0x114F90`, via the one call to `0x115D30` at `0x110F4D`, into a buffer separate from the keyboard's. LT is `equip_window` (L2, item window), RT `weapon_window` (R2, weapon window), LB `weapon_aim` (L1), RB `pov_cam` (R1), L3 `interrogate`, R3 `change_view`. The item and weapon windows (`0x32D950`, `0x32E5D0`) read the gameplay pad directly, so a remap inside the pad wrapper alone cannot move them; the remap goes at the controller read. Corrected: the plan's Phase 4 labels are swapped, `0x32DA44` is the item (L2) window and `0x32E6C1` the weapon (R2) window.
- L1 sweep of the forms skipped before: 4 more L1 reads. `0x35B5E5`/`0x35B617` are in component `0x84FAEC`, registered for weapon ids 15 and 16, which the lock already leaves out. (Corrected 2026-09-26: hash `0x1484FAEC` is registered 3 times, at `0x35A7B9` for ids 9 to 13, `0x35B749` for ids 15 and 16, and `0x35E209` for id 14; the two reads above are in the ids 15/16 instance.) `0x212BA5` and `0x214913` are camera code on the raw pad. No change needed.
- Corrected by review: status `0xE3` is not a free switch. The weapon's message `0x10` handler clears it every frame (`0x37A683`–`0x37A69D`, CLEAR `0xDD`, `0xDF`, `0xE3`, `0x35`), and CQC code sets it at message `0x80` (`0x384702`). Writing it from the pad wrapper therefore never reaches the reader. The fix gates the reader instead.
- Delta New Style (Xbox, Game8 and Prima): LT aim, RT attack and CQC, LB stalk, RB or R3 first person (sources disagree), D-pad for the camo, weapon, radio, and equipment windows. MC's windows open by holding L2/R2 (about 12 frames, `0x32DA25`/`0x32E6A7`) and, per PS2-era control guides, are browsed with left and right while held, so 0.5.9 does not move them to the D-pad.

### 0.5.9 (5b, crouched strafe, pad layout)

Built 2026-09-25 and reviewed before install by two independent Opus reviewers (game semantics, C runtime) and one verifier of the fixes. Played the same day (below).

- Mouse aim (5b). While the lock is on, in the stand or move sub-state with status 1 (or a crouched strafe), not holding an enemy, not in first person, the wrapper writes the camera's orbit yaw (`[0x1D8A068]+0x338`) into `+0x800` and `+0x16A` after the pad component. It skips when movement will not strafe this frame (predicted from the same gates movement tests: TEST 3, ANY 9/`0xBA`/7/`0x4F`/`0xDE`, a gun raised, weapon id not 8), when the camera is another actor's, in a sub-mode, or in modes 1, 4, 5.
- Auto-target off while locked. The TEST(`0xE3`) call at `0x37A829` goes through `notarget_gate`, which also passes while the lock is on, so the helper drops its target and never writes `+0x800` toward an enemy. It stays on for the whole lock (a single frame off would let the press path snap the aim) and comes back in first person, while holding, where the camera cannot be used, and with F8 (the toggle reads F8 only while the game has focus). When the gate drops while locked, the weapon node's press flag `+0xAA` is cleared so the aim does not snap. Consequence: third-person vertical aim has no target, so shots go level along the camera heading (the pose pitches the gun toward a level aim point; see the 2026-09-26 research). Hold-ups that may rely on the auto-target are a play-test item.
- Crouched strafe. The TEST(1) call at `0x369C5A` goes through `cstrafe_gate`, which also passes while locked in the crouch sub-state; the magnitude is capped at `0x96` (the slower strafe speed). Expected risk: Crouch Walk's forward squat-walk animation may add forward drift. The aim probe logs position and velocity to judge it.
- Pad layout. The one call to the Steam Input read (`0x110F4D` to `0x115D30`) goes through `ctrl_wrap`; in gameplay the port-0 buffer is edited: LT to L3 (aim), RT to Square with pressure `0xFF` (fire, and CQC without LT through the keyboard CQC rule), LB to L2 (item window), RB to R2 (weapon window), the right stick click toggles R1 (first person; the toggle survives menus), D-pad up to R3 (camera view), D-pad down to L1 (native lock-on and camera reset). D-pad up and down no longer move Snake; left and right do. Outside gameplay (no pad component call for 2 reads) buttons are native, each button keeps its role until released, and presses made in the 2 reads before gameplay ended fall back to native. The remap is off if the in-game button config is not the default. QCamo reads physical LB+Y for its menu, so that chord now also sends L2 and Triangle.
- Also fixed from review: the L1 edges on the lock-on and lock-off frames, a reset of the lock when the player actor changes, and the gates ignore flags older than 100 ms.
- Known and accepted: the strafe prediction reads the weapon at `+0x6F8`, while movement reads the global weapon `[0x1D32980]`; on a weapon switch the two can differ for a frame.

Installed as generation 81, `fpvmove.asi` 170,496 bytes, SHA-256 `8ad2cbb15a5c4483fad3d0852f239ee7df6dd2e04cee341014eb06ae10fd6b98`; `mgs3mod verify` reported `compatible: true`.

Played on 2026-09-25 from 23:36 to 23:50, keyboard then the wired pad. The user reported everything working as expected. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.5.9-play.log` (2,827 lines, no fault, dropped, or depth line). All four hooks armed at load. Pad totals: `aim=12772 notarget=5311 cstrafe=2197 remap=93 lock=71 unlock=71 fire=89 rearm=83 keep=4 cqc=52`. From the log:

- Weapons: ids 5 and 11 locked; the knife (1), grenades (19, 21), and empty hands (0) never did.
- Mouse aim wrote on 63 stretches. Skips: first person 13 (one frame each), camera mode 4 12 and mode 5 1 (one to three frames each, at lock-on), holding 4, stance 5 (up to 5 s), moving without strafe 42 (mostly under 0.5 s; one of 5.9 s at 23:47:26 with the stick held and Snake not moving, `strafe=0`, position unchanged).
- Auto-target: none of the 1,044 aim probes shows a target while the gate was on. F8 was toggled 19 times; with assist on the probe shows `target=1`, with it off `target=0` (23:47:29 to 23:47:31).
- Crouched strafe: with the facing steady, the position moved at 90 degrees to it at about 1,200 units per second (23:42:29.6 to 23:42:30.2: +90 degrees on every probe). No forward drift from Crouch Walk was measured.
- Holding a guard with RMB: `keep=4`, holding on and off 4 times.
- Pad: 93 remapped presses (LT 14, RT 11, LB 7, RB 14, D-pad up 10, D-pad down 13), first-person toggle 12 on and 12 off, 20 switches between the remapped and native layout (menus and windows). The controller read runs on the game thread.

User answers after play:

- Crouched strafe works. Moving forward and back looks right; moving left and right plays the same forward/back crouch-walk animation. Not critical; fix if possible.
- Vertical aim is needed. With auto-target off, enemies below were missed. Full auto-aim is too easy and stays off by default; it could stay on for easy-difficulty saves.
- Hold-ups work.
- The first-person toggle on the right stick click is fine for now (the button choice may change later).
- The QCamo pad chord was not tried (keyboard G only). Side note: a quick face-paint change would be good later.

### Research for vertical aim, difficulty, crouched side-step, and a crosshair, 2026-09-26

Four read-only Opus agents; the elevation design was checked by an Opus refuter. Full reports are kept locally under the session scratchpad.

- Third-person shots do not follow the fire ray. Every third-person player bullet is spawned with no target point (callbacks `0x359BE0`, `0x35A7F0`, `0x35B780`, `0x35C6E0`, `0x35D190`, `0x35E240`) and flies along the gun; the ray result `+0x6E0` is used in first person only. The gun's pitch comes from the upper-body pose `0x3700A0`, which aims bone `0xE` at the aim point `+0x520` (clamped to -80/+50 degrees). So the height of `+0x520` sets the vertical aim. The pose-to-muzzle link is inferred, not traced.
- The aim point builder `0x378C90` has 7 calls and 1 tail jump; the 6 calls in the gun state `0x3741E0` are the ones that matter. Without a target it puts `+0x520` 10000 ahead along the eased yaw `node+0xA8` at Snake's position height (`+0x134`). The native target search `0x73FBA0` has one caller (`0x37AA32`) and can be called with a matrix centred on the mouse yaw.
- Corrects the 0.5.9 note "the reviewers found no pitch source": the pose is a pitch source; with the auto-target off it was simply fed a level point.
- Difficulty is the word at `[[0xACDE98]+6]` (the pointer is set at init and must be read each time): 10, 20, 30, 40, 50, 60 for Very Easy to European Extreme (mapping from the PS3 cheat table, the menu string order, and the matching health offset `+0x684`; to be confirmed in play). The native aim assist does not read it.
- A sideways crouch animation does not exist in the base game: its player motion archive has 199 motions and none is crouched locomotion; Crouch Walk added the one squat-walk motion (#199). The motion that plays is `node+0x60`, sent every frame at `0x369972`. A crouched wall-press shuffle is the only possible stand-in and is not identified; a log probe is needed. A real fix is two new animations (an asset project).
- Crosshair: the game renders with D3D11 (feature level 11_1). The one Present is at `0x6EF99` inside `0x6EF50`, reached from `0x2887E` on the game thread. There is no third-person reticle; the game's own world-to-screen (`0x123F70`) uses channel 0 (matrix `0x1D56170`, viewport `0x1D563C0`..). QCamo draws with ImGui on a MinHook Present hook, so the plugin draws before that, at its own call site.

### 0.6.0 (elevation assist, crosshair, difficulty default)

Built 2026-09-26, reviewed by two independent Opus reviewers (game semantics, runtime and D3D). Fixed from review: the search now tests bone 2 (the point whose height is used) and retries with a narrower cone when the nearest target is off the aim line; the crosshair maps PS2 pixels with the viewport bound at draw time when the bound target is the back buffer (else the whole buffer), reads the aim point and matrix at present time, clamps its rectangles, and releases its views in `__finally`; the trampoline is set before the detour goes live; the yaw pre-write needs the lock and a fresh gate; an unknown difficulty value keeps the last one. Accepted: the crosshair marks the aim point, not the muzzle ray, so a target beyond the pose pitch clamp (-80/+50 degrees) shows a hit the arms cannot make; the trampoline has no unwind info; a fault caught by the plugin still triggers QCamo's crash dump (check `fpvmove.log` for `elev fault` or `xh fault`). Installed as generation 85, `fpvmove.asi` 195,072 bytes, SHA-256 `d1b8835bfdb21eb00b80aca1f147d34704f7c51352b37e545ecdef738e9c2fa8`.

Played on 2026-09-26 from about 01:21, Hard save. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.0-play.log` (2,309 lines, no fault). All hooks armed. Totals: `aim=8642 notarget=8324 elev=1670 elevmiss=1067 xh=7564 cstrafe=1069 lock=17 fire=37`. The user reported: vertical aim still broken; the crosshair shows but moves only left and right, and up or down only when the auto-aim picks a target; the default camera puts Snake and the crosshair on one vertical line and hides what is ahead, so an over-the-shoulder camera is needed. From the log:

- Elevation found a target on 1,670 frames, nearly all on the same floor (target bone 2 about 320 above Snake's position); 1,067 frames had the nearest target off the aim line. With no target the aim point stays at Snake's position height (`aimpt - pos = 0` on 447 of 509 probes), so the mouse has no vertical control: the camera's vertical mouse input changes its height and zoom, which moves the crosshair on screen (y from 19 to 446 px at 1080p) while the aim stays level.
- Crosshair placement: back buffer 1920x1080 with the bound viewport covering it, channel 512x448; the projection is consistent (level aim point at screen centre, `px=960,539`).
- Difficulty: the read works. The log has `ctx difficulty 40, aim assist off` at 01:21:41 (Hard). (Corrected 2026-09-26: this line first said no such line existed; a faulty grep pattern missed it.)
- Conclusion: elevation assist is the wrong model for the user; the next step is manual pitch from the mouse Y (and right stick Y) plus an over-the-shoulder camera while aiming.
- The user added: shots were not missed; the crosshair looked aligned with where they landed.

### 0.6.1 (manual pitch)

Built 2026-09-26 inline (no subagents, at the user's request). Scope agreed with the user: vertical aim only; the over-the-shoulder camera comes after.

- The elevation assist is removed (its target search call and patterns are gone). The aim point builder detour stays: for gun-state calls during the lock it sets the aim yaw `node+0xA8` to the facing, then sets the aim point height to `pos.y + 320 + tan(pitch) * distance` at the builder's own distance (10000). 320 is the pivot height above Snake's position, taken from the 0.6.0 log (target bone 2 on the same floor); the pose pivots at bone `0xE`, whose height is not read.
- The pitch starts at 0 on each lock and is clamped to +-0.70 rad (about 40 degrees; the pose allows -80/+50). Mouse: `pitch += dy * 0.0009` rad, dy = the game's mouse delta (`0x1D7A364`, the int32 after dx; pattern `BF FF 00 00 00 48 8B 1D` at `0x112FDB`). Pad: the raw right stick Y byte (`0x1E2C565`), dead zone `0x30`, up to 0.035 rad per frame, used only when the mouse did not move.
- While locked (and the aim assist is off), the camera's vertical input no longer changes its height and zoom: `0x212AA0` gets an entry detour (6-byte prologue `40 53 48 83 EC 30`) that returns at once for the player's camera. So the camera stays level; the crosshair moves up and down with the pitch.
- The difficulty read also logs its raw value (`ctx difficulty raw`); added on a wrong premise (see the corrected 0.6.0 note), kept as a cheap check.
- Log: `pitch` counter, and once a second `pitch <deg> aimpt= pos= dy= stick=`.

Installed as generation 89, `fpvmove.asi` 184,320 bytes, SHA-256 `43d5689eca51497b906851f695e438e5e3fb39d50083484fb30bbfe983cd5efa`. Not reviewed by a second agent.

Played on 2026-09-26 from 01:37 to 01:44, Hard save (`ctx difficulty raw 40`). The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.1-play.log` (3,306 lines, no fault). Totals: `pitch=8762 xh=8747 aim=8796 notarget=8670 lock=26 fire=41 cstrafe=742`. Pitch samples (once a second) ranged from -35 to +40 degrees, mostly within +-10. The user reported: working great; the vertical direction is right; shots land on the crosshair; the mouse pitch should be a little slower. Vertical aim passes.

### 0.6.2 (over-the-shoulder camera prototype)

Built 2026-09-26 inline as a quick prototype at the user's request. The one call to the camera's eye and look-at function `0x2137D0(camera, eye = C+0x380, look-at = C+0x390)` at `0x21332F` is redirected (pattern `4C 8B C6 48 8B D7 45 0F 29 4B B8 48 8B D9 E8`, one match). After it returns, while the crosshair shows, the eye is moved to 60% of its distance from the look-at, both are shifted 400 units to Snake's right (right = (cos yaw, 0, -sin yaw) of the camera yaw), and the look-at is raised by tan(pitch) times the horizontal distance. The game's own wall and smoothing code runs after this. F7 toggles it. The mouse pitch rate is lowered to 0.00075 rad per unit. Installed as generation 93, `fpvmove.asi` 191,488 bytes, SHA-256 `5ed2f93ae1044bb5685fdb6913c827f4520a40e7f842cfa90b24777b6e3e7ae0`.

Played on 2026-09-26 from 01:53 to 01:54. Log kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.2-play.log` (843 lines, no fault; `ots on` 13 times, `lock=13 fire=24`). The user reported: it looks decent, but it needs a lot of work:

- Shots did not land where the crosshair aimed.
- The crosshair lines up only if aiming starts from a certain camera position; when the camera looks at the sky and aiming starts, the crosshair is not visible (likely above the screen).
- A key or button to switch the shoulder side is needed.

Analysis for the next session (not yet verified in play):

- The pitch starts at 0 on each lock, but the camera keeps the height and zoom level it had (the zoom is frozen while locked). A camera raised toward the sky therefore looks far above a level aim, so the crosshair leaves the screen. Fix: on lock, reset the camera to a fixed aiming level (zoom index), or seed the pitch from the camera's current pitch.
- With the camera shifted to the side and tilted by the pitch around its own look-at, the camera's centre ray and the gun's aim line (from Snake, pivot 320 above his position) no longer meet, so the crosshair (the projection of the aim point 10000 ahead) and the shots only agree at that distance. Fix, as in shooters: cast the camera's centre ray (the game's line check `0x10B4B0` in the pad wrapper, from the camera eye along yaw and pitch), put the aim point on the hit (or far along the ray), and let the gun aim at it; the crosshair then sits at the screen centre and the bullets converge on it.
- A shoulder-side switch: a key (for example Q or the middle mouse) and a pad button, flipping the sign of the 400-unit offset, with a short ease.

### 0.6.3 (shoulder camera quick fixes)

Built 2026-09-26 inline. The pitch is seeded on the first locked frame from the camera's own direction (eye `C+0x380` to look-at `C+0x390`) instead of 0; the camera's look-at is set to the gun's aim point `+0x520` (the eye keeps its shifted position), so the crosshair should sit near the screen centre; the middle mouse button swaps the shoulder (the offset sign). Installed as generation 97, SHA-256 `68be213677bb962bbaee956137ba288f416d28c3d8592f8e1ab8b14f2e4a13af`.

Played on 2026-09-26 around 02:00. Log kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.3-play.log` (1,658 lines, no fault; `lock=19 fire=37`; pitch samples from -40 to +40 degrees, mostly -10 to 0). The user reported:

- Aim start partly fixed: aiming with the camera at the sky puts the crosshair in the sky, but aiming down from there gets messed up.
- Shots still miss: they seem to land above the crosshair and slightly to one side; the user could not get them onto it.
- The shoulder switch (middle mouse) works well.

Analysis for the next session (not verified):

- The shots miss because third-person bullets do not go to the aim point: they leave the gun along the muzzle, and the pose only turns the arms toward `+0x520` from bone `0xE` (pivot height 320 is a guess) within its clamps and yaw twist limit. With the camera beside Snake, that offset shows. The robust fix is to give third-person shots a target like first person does: the bullet callbacks pick `0x25F4E0` with `&actor+0x6E0` as the target only in first person (status 7), and `0x25F240` with no target otherwise (`0x359BE0` and five siblings). Routing locked third-person shots through the targeted path, with `+0x6E0` set to the hit of the camera's centre ray, would make the bullets go where the crosshair is.
- "Aiming down gets messed up": likely the camera looking at an aim point close below Snake (the look-at jumps near the eye, or the game's smoothing and wall code reacts). Check with a probe of eye, look-at, and pitch, and clamp the look distance.

- Elevation assist: the builder `0x378C90` gets an entry detour (its 5-byte prologue is checked and moved to a trampoline). For calls from the gun state (return address within the window around the fire call) during the lock, the aim yaw `node+0xA8` is set to the facing first, so the shot follows the mouse without the ease; then the native search is asked for a target in a narrow cone (0.14 rad) around that yaw. If the record passes the native checks and the target lies within 600 units of the aim line, only the height is taken: `+0x520` becomes the point at the target's distance along the mouse yaw at its bone-2 height. The yaw stays manual.
- Aim-assist default by difficulty: on for Very Easy and Easy, off above; F8 overrides until the difficulty changes.
- Crosshair: the call to `0x6C2E0` inside the present wrapper is redirected; after it, the aim point snapshot from the pad wrapper is projected with the game's matrix and drawn as a small outlined cross with `ClearView` on the current back buffer. Hidden when not locked, in first person, with unusable cameras, and while paused. Placement on the back buffer uses the renderer's last viewport or the whole buffer; the log records the numbers to check it.
- Motion probe: the movement node's sub-state, motion word, and direction bits are logged on change (at most 10 lines a second).

### Research for precise shots and the shoulder camera, 2026-09-26

Read-only, game closed, on a `dumpbin /DISASM` of `.text`. Addresses are RVAs.

- Bullets. Six player bullet callbacks call an untargeted spawner outside first person and its targeted twin in first person: 0x25F240/0x25F4E0 (callbacks 0x359BE0, 0x35A7F0, 0x35B780, 0x35E240), 0x25F390/0x25F630 (0x35C6E0), 0x262250/0x262350 (0x35D190). Each twin takes the same arguments with a target pointer inserted second. The bullet init 0x25E320 (and 0x261250) then aims from the muzzle matrix position (row 3) at the target xyz; without a target it uses the matrix forward (row 2). The target is a float4 whose w must be 1.0 (the first-person hit point has w = 1.0).
- The gun state's fire frame (0x3749CA) already runs the line check 0x10B4B0(0x1F, actor id `+0x120`, 0x42, from, to, radius 0.0) from `[actor+0x730]+0x20` to the aim point `+0x520`; on a hit, 0x105FD0(0x105E50(), &actor+0x6E0) stores the hit point. First person runs the same check from its eye 10000 ahead (0x373C10..0x373D8D).
- Camera. 0x2137D0 puts the look-at at the target actor's position plus a table height (900, 900, 500, 0 by zoom level) plus the eased stance offset `C+0x358` (0 standing, -200 crouched, -400 prone, -500 holding and a few others), and the eye behind it along the orbit yaw `C+0x338` at a table distance and height: (1000, -1400), (2000, -1000), (3700, +950), (2250, +5500) (table at 0xB029E0, Y up). Zoom levels 0 and 1 put the eye below the look-at, so the native "camera at the sky" eye sits low; 0.6.2 and 0.6.3 built the shoulder eye from it, which is the likely cause of "aiming down after starting high misbehaves". The caller 0x2132D0 then applies a distance extension `C+0xD4`, offsets `C+0x360..0x36C`, and eases eye.y and look.y by 0.25 per update (unless `C+0x64` or `C+0x31C` is set). The wall step 0x2139B0 line-checks from the pivot `C+0x3A0` to the eye, pulls eye and look toward the pivot when blocked, clamps the eye above the floor, and writes the final eye `C+0x3D0` and look `C+0x3E0`; 0x213080 then blends the render camera toward them (factor about 0.58 per frame at 60 fps).
- Reviewed before install by one independent Opus agent: the spawner argument forwarding and the line-check prototypes are correct. Fixed from review: the rig is written at the wall step instead of inside the eye/look step (the distance extension and offsets would have bent it), a shot is converted only when the muzzle points within about 45 degrees of the hit, the aim point for the arm pose is set only while the facing follows the camera and kept within 10000 of Snake, the player id is cached instead of read through a possibly stale actor pointer, the camera mode is rechecked at the wall step, and with F7 off the pitch only moves while the facing follows the camera (as in 0.6.1). Open from review: a pivot inside a wall (Snake pressed against one on the shoulder side) lets the ray pass the wall; another player-id spawner call at 0x3533A0 is not converted (its weapon is unknown); the render camera blend lags the rig a little (the crosshair is projected, so it stays true).

### 0.6.4 (one aim rig for the camera and the shots)

Built 2026-09-26 inline. While locked in third person with the shoulder camera on:

- One rig: a shoulder point 700 above Snake's position (plus 1.5 times the camera's stance offset) and 400 to the side, the camera yaw, and the pitch. The eye is 2000 behind the shoulder point along the aim direction, never lower than 100 above Snake's position. Camera pitch and aim pitch are the same value, and the eye no longer depends on the native zoom.
- Once a frame in the pad wrapper, the game's line check 0x10B4B0 casts the centre ray from the shoulder's depth up to 40000 further. Its hit is the shot target, the crosshair position, and (while the facing follows the camera) the arm pose's aim point.
- The six untargeted spawner calls are redirected to wrappers that call the targeted twin with the hit.
- The camera's wall step 0x2139B0 gets an entry detour (11-byte prologue moved to a trampoline) that writes the rig's eye and look-at before the game's wall pull and floor clamp.
- F7 off gives 0.6.1 behaviour. Middle mouse still swaps the shoulder, with no easing yet.
- Log: `rig`, `shot`, `eyenow`, `eyeold` counters; once a second `rig`, `cam nat`, `cam fin`, and `cam d4` lines; one `shot` line per shot (converted, `native`, or `native: muzzle N deg off`).

Installed as generation 101, `fpvmove.asi` 208,384 bytes, SHA-256 `da0f7f2aeadc236e9dbb6cbe602334778e9570ef3fdd5795995673bd3a4144ec`; `mgs3mod verify` reported `compatible: true`.

Played on 2026-09-26 from 10:50 to 11:37 with the pistol (id 5). The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.4-play.log` (6,071 lines, no fault or dropped line). All hooks armed; 6 of 6 shot calls redirected. The user reported:

- Aiming and shots are better than 0.6.3 but not precise. Sometimes a shot is exact; after moving the aim to another guard, shots pass beside the guards.
- Looking at the sky and then aiming down works. Head shots sometimes land exactly on the crosshair and sometimes do not.
- F7 (shoulder camera off) works, but aiming is worse there; the user drops that mode.
- Wall aiming was not tried.

From the log:

- 85 shots took the rig's hit, and 7 stayed native. All 7 native shots fell in the F7-off stretch (11:33:41 to 11:34:03). No shot was refused for the muzzle angle.
- Of the 85 converted shots, 46 have d between 34,591 and 39,209: the line check hit nothing, so the target was the far end of the ray (40,000 out), including shots the user aimed at guards. The bullet then flies from the muzzle, which sits beside the camera ray, toward a point 40 m out. At a guard 5 to 15 m away it is still most of that side offset away from the crosshair. That matches "sometimes exact, sometimes beside". The line check with flags 0x1F and 0x42 did not report guards here; why is not known.
- The `eyeold` counter stayed 0 (`eyenow` only): the camera update runs after the pad wrapper in the same frame, so the rig the camera shows is the one the shot uses.
- The camera's distance extension `C+0xD4` read 180 on all 272 probes and the offsets `C+0x360..0x36C` read 0, 0, 0, 1; the rig now replaces the eye and look-at after both are applied.

### 0.6.5 (shots along the camera ray)

A converted shot now starts on the camera's centre ray instead of at the muzzle. The spawner gets a copy of the muzzle matrix whose position is moved to the ray point nearest the muzzle. That point is kept at or past the line check's start and 50 short of the hit. The target is still the rig's hit, so the path is the crosshair's ray whether the line check hits or not. The spawners copy the matrix into the bullet (0x25E320, and 0x261250 at 0x2612C6), so a static copy is safe. Each `shot` line also logs how far the start moved (`moved=`).

Installed as generation 105, `fpvmove.asi` 208,896 bytes, SHA-256 `48ea9bf7476232337fce1beb8c94ff9087ac75c258436c5fa4c59a5440052e22`; `mgs3mod verify` reported `compatible: true`.

Played on 2026-09-26 from 11:53 to 11:59. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.5-play.log` (3,412 lines, no fault or dropped line). The user reported: "looking awesome"; shots were precise at all distances with the pistol, the automatic rifle, and the sawed-off shotgun. Task 1 (precise shots) passes. From the log:

- 43 converted shots, no native shot: 11 with the pistol (id 5), 21 with the automatic rifle (id 11), 11 with the shotgun (id 14). All three go through pair A (0x25F240/0x25F4E0); pairs B and C were not used.
- The start moved from the muzzle to the ray by 147 to 493 (median 318). The muzzle pointed 0.2 to 13.3 degrees from the hit.
- 34 shots had a line-check hit and 9 went to the far end of the ray.
- Fire presses: 33 pistol, 15 rifle, 17 shotgun.

### 0.6.6 (eased shoulder camera, pad shoulder swap)

- The camera blends from the native camera to the rig over 150 ms when the aim starts, and back over 150 ms when it ends (smoothstep, timed with the performance counter). While easing out, the rig keeps the last pitch; a new lock during the ease-out continues from that pitch instead of seeding again. A camera change (another mode or actor) ends the ease at once.
- The shoulder swap eases the side offset over 150 ms (the aim ray moves with it).
- Pad shoulder swap: a left stick click while LT aims. That click is the pad's L3, which LT already holds while aiming, so the game sees no change; without LT it keeps its native role. The remap in `ctrl_remap` (LT, RT, LB, RB, right stick click, D-pad up and down) did not use it. Logged as `pad LS press, shoulder swap`; `ctx shoulder` lines now say `mouse` or `pad`.

Installed as generation 109, `fpvmove.asi` 209,408 bytes, SHA-256 `7b1ff404d37e922282c6c0e709f6d1c2e1975d59cea40dcd1417bda256f5856a`; `mgs3mod verify` reported `compatible: true`.

Played on 2026-09-26 from 12:08 to 12:16, keyboard and pad. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.6-play.log` (5,499 lines, no fault or dropped line). The user reported that everything worked and nothing was wrong. Task 3 passes. From the log:

- 80 converted shots and no native shot: 22 with the pistol (id 5) and 58 with the automatic rifle (id 11).
- 132 lock-ons and 105 `ots on` transitions.
- 15 shoulder swaps: 11 with the mouse and 4 with the pad (4 `pad LS press` lines).
- Pad presses remapped: LT 50, RT 92, RB 27, LB 3.
- Pitch samples reached both clamps (-40.1 and +40.1 degrees).

Tuning answers after play (task 4): shoulder offset (400), distance (2000), height, the +-40 degree pitch limit, mouse pitch speed (0.00075), and the match between mouse X and Y are fine for now. The right stick pitches too fast.

### 0.6.7 (stick pitch, wall check)

The right stick's pitch rate is lowered from 0.035 to 0.025 rad per frame at full deflection (about 86 degrees a second at 60 fps, from 120). The stick's yaw is the native camera's and is unchanged. This build is also the wall check of task 4: the rig probe logs the game's wall ratio (`wall=`) and final eye once a second.

Installed as generation 113, `fpvmove.asi` 209,408 bytes, SHA-256 `20eba024c2d932b73f63523b119f635580f6186e3890c28b1712d54309e83b2e`; `mgs3mod verify` reported `compatible: true`.

Played on 2026-09-26 from 12:20 to 12:27, mostly on the pad. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.7-play.log` (3,508 lines, no fault or dropped line). The user reported:

- The right stick is still a bit fast: left and right (the native camera yaw) and up and down.
- Aiming with a wall behind Snake is too hard: his head hides everything. With the wall on the shoulder side it is not too bad.
- The crosshair appears whenever right mouse is held, whatever the position, including against a wall. Which position was meant is to be asked.

From the log:

- 18 converted shots and 3 native shots. The native shots are `muzzle 152`, `157` and `168 deg off` (12:25:52, 12:26:39, 12:26:40): the hit was on a wall about 300 to 600 away to Snake's side (x = 18125 with Snake at x = 18661, left shoulder). Those shots went along the gun while the crosshair showed the wall beside him.
- The game's wall ratio was below 1 on 38 of 138 probes (down to 0.17). The wall step pulls the eye and the look-at toward its pivot at Snake's head, which puts the head in front of the view.

### 0.6.8 (stick speed, wall behind, unreachable crosshair)

- Right stick pitch lowered again, from 0.025 to 0.020 rad per frame at full deflection.
- Right stick yaw while aiming: the camera's yaw step 0x213010 (entry detour, 6-byte prologue) reads the stick X input at `C+0x308` (written from the stick byte at 0x211F77). For the player's camera, while the rig is on and the mouse did not move that frame, the input is scaled to 0.75 for the call and restored after. The mouse turn is unchanged.
- Wall behind: the rig does its own wall check. It runs the line check from the ray point level with the shoulder back to the eye, with the camera's own flags (0xF and 0x840050, or 0x1F and 0x840042 when `C+0x60` is set). On a hit, the eye moves forward along the ray to 100 in front of the wall, and no closer than 300 behind the shoulder point, so the view keeps its line and stays beside the head. After the wall step, the final eye `C+0x3D0` and look-at `C+0x3E0` are set to the rig (by the ease blend), so the game's pull toward the head and its floor clamp give way to the rig.
- Unreachable crosshair: when the hit lies more than about 70 degrees from Snake's body yaw (a wall beside him), the crosshair is hidden (`rig out of reach` in the log), since the shot stays native there.
- Counters `wall` (rig wall pulls) and `turn` (scaled stick yaw calls).

Installed as generation 121, `fpvmove.asi` 211,456 bytes, SHA-256 `8ebb08a4614bd218993a5085b8bd6f5dbed042f5658faea2cf4a8ffa20fee25a`; `mgs3mod verify` reported `compatible: true`. An intermediate build of the same code labelled 0.6.7 was installed as generation 117 and replaced before play.

Played on 2026-09-26 from 12:43 to 12:56. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.8-play.log` (3,260 lines, no fault or dropped line). The user reported everything working well: the stick speeds, the wall behind Snake, and the wall at the shoulder. One issue: in a wall hug, peeking at a corner, the crosshair shows but the aim cannot be moved. From the log:

- 16 converted shots and 2 native ones. There were 13 `out of reach` crosshair hides, 1,462 rig wall pulls, and 1,008 scaled stick-yaw calls.
- The wall hug and corner peek (12:44:35 to 12:44:41 and 12:55:10 to 12:55:16) alternate `aim skip: camera mode 5` and `aim skip: moving without strafe`, with the lock on. Camera mode 5 is the wall camera: the rig and the mouse aim are off there. The `moving without strafe` frames still drew the 0.6.1 crosshair on the native aim point, which the mouse cannot move.

### 0.6.9 (no crosshair in a wall hug)

With the shoulder camera on, the crosshair shows only the rig's hit. So it is hidden wherever the rig is off (camera mode 4 or 5, including the wall hug and corner peek) and when the hit is out of reach. With F7 off, the 0.6.1 crosshair is unchanged. A movable aim in the corner peek would be new work: the peek is the native wall camera.

Pad stick clicks swapped at the user's request (most games swap the shoulder on the right stick click). The right stick click swaps the shoulder while LT aims and keeps its native R3 otherwise. The left stick click toggles first person; its own L3 is hidden in gameplay, and LT still sends L3 while aiming. The 0.6.6 left-click swap is gone.

Installed as generation 129, `fpvmove.asi` 210,944 bytes, SHA-256 `7d4bacc2d2aac5f64db6da16459278ec330ba0805d4b59a4b962ae448c553d32`; `mgs3mod verify` reported `compatible: true`. The first 0.6.9 build (generation 125, without the stick swap) was replaced before play.

Played on 2026-09-26 from 13:31 to 13:43 as the task 5 regression pass: keyboard and pad in one session. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.6.9-play.log` (6,260 lines, no fault or dropped line). The user reported everything working well: the pistol, the automatic rifle, crouched strafe, first person, the wall hug with no crosshair, holding a guard and hold-ups, CQC, the knife, grenades, and the pad. Tasks 1 to 5 pass. From the log:

- 45 converted shots. The 7 native shots all fell on `aim skip: holding` frames or next to them (shots while holding a guard stay native by design).
- Weapons seen: pistol (5), automatic rifle (11), shotgun (14), knife (1), grenades (19, 21), empty hands (0).
- Pad rule totals: `cqc=35 knife=4 keep=7 lock=71 remap=209`, with 12 `holding on` lines. Pad presses remapped: RT 116, LT 46, RB 24, LB 3. There were 10 right stick shoulder swaps and 10 left stick first-person toggles.

Issues the user set aside, to handle later if complex (see "Optional backlog").

### 0.7.0 (pad A crouch from a run, LT after a weapon switch)

Built on 2026-09-26 inline for backlog items 4 and 6. At the user's request it is not installed or played yet; it will be tested together with the next features. Reviewed by one independent Opus agent before commit, and again after the fixes (verdicts below).

- Pad A (item 6). The move sub-state rolls or crouches on the Cross release (0x3690B3; roll at 0x36912E needs magnitude above 0x96, else crouch at 0x36917F), so a crouch from a run needs the magnitude capped at release. In the move sub-state a pad A press is held back. If released within 10 frames, it becomes the C crouch: Cross for 4 frames, then released, with the magnitude capped at 0x96. If still held at 10 frames, it becomes the SPACE roll pulse, and the roll ends in prone only if A is still held 12 frames later. Outside the move sub-state A is native. A second press during the tap is ignored (it would release Cross without the cap and roll). A pending press goes native if Snake stops, status 0xE5 appears, or C is pressed. After a gameplay stop (the controller read's idle flag), the rule starts over and a held A is not a new press. The physical A comes from the controller read (hi byte 0x40 is Cross; checked at 0x34E33) whenever the in-game button config is the default, and counts as up if older than 100 ms. Known and accepted: from a slow walk, a 10-frame hold rolls (the SPACE design; native would crouch); pressing SPACE while A is pending gives a crouch.
- LT after a weapon switch (item 4). Cause: a trigger pressed while a window closes (no pad component call yet) got the native role, so LT sent L2. On the first gameplay frames that either opens the item window (held about 12 frames, counter at 0x32D95x) or triggers the L2 quick-tap toggle of the equipped item (0x32D184 to 0x32B6F0; RT likewise at 0x32DE21 to 0x32B7E0). Fix: a new pending role. An LT or RT press outside gameplay, within 30 reads of an LB or RB release or while LB or RB is still held, with the default config, sends nothing (its L2/R2 bit and pressure are cleared). It takes its gameplay role (aim, fire) on the first gameplay read, or turns native after 120 reads. Log lines `pad LT held into gameplay, remapped (N polls after the press, M after LB/RB)` and `pending timed out` will give the real window-close timing. Accepted: in a menu, an LT or RT press within 30 reads of an LB or RB release is held back and lost if released before gameplay or the 120-read cap.
- Review findings not fixed: pad A shares the `crouch`, `roll` and `rollprone` counters with C and SPACE (the trace lines `pad A tap: crouch`, `pad A hold: roll`, `pad A held on: roll into prone` tell them apart). Whether the controller is read exactly once per gameplay frame is taken from the 0.5.9 log, not re-traced.

Built as `work/fpv-move-build/fpvmove.asi`, 213,504 bytes, SHA-256 `f80d32711177ca53f5bf3c55c4aab85481a859e20c8f5ec5ce7d769104eea29e`; the file-scan test passes. Not installed (generation 129, 0.6.9, stays installed).

To test with the next features (pad): (a) run, tap A: crouch; (b) run, hold A: roll; hold on: roll into prone; (c) tap A twice fast while running: crouch then stand, no roll; (d) RB weapon switch, then LT right away, including LT pressed before RB is released: aim, no item window, and no equipped item toggling; the same with RT; (e) standing, crouched, and prone A unchanged. (Since 0.8.0 the weapon window is on D-pad right, so (d) is played with D-pad right.)

### Research for the 0.8.0 batch, 2026-09-26

The user asked for everything left in this plan in one batch (the set-aside backlog, the older optional backlog, Phase 4, and the open items), one build, one test list. Eight read-only agents ran in one workflow, each seeded with "distrust the source; verify line numbers; sweep, do not sample": hold-up drop, first-person view, wall aim, Phase 4, CQC from a crouch and reach check, crouched sideways animation, face paint (all Opus; face paint at medium effort, the rest high), and an open-items sweep with the weapon table (Sonnet, medium). Their reports are kept locally under the session scratchpad, not in git. The main facts, each re-read in `work/disasm/text.asm` before it was used:

- Hold-up drop. A held-up soldier shakes and drops an item only while status `0xBA` (the first-person view; one SET at `0x38B75B`, one CLEAR at `0x38B903`) is set: TEST(`0xBA`) at `0x2010AA` (hold-up mode `0xE`, sub-state `0x91`) and `0x202D19` (mode `0xF`), and TEST(7) at `0x2015EE` (mode `0xE` variant B, a reaction without a drop). Mode `0xE`'s shake test `0x19D3A0` also wants the player's aim line (a target record of kind `0xAD` at actor `+0xA20`, cast 30000 along the muzzle outside first person by its update `0x382B80`) on body part 1 or 3 of the soldier's damage receiver (`[+0x1CB8]+0x12C`), unless the last ammo used was a gun's or the weapon is id 14 or 17. Mode `0xF`'s test `0x19C5B0` has no part test. After three shake cycles some soldiers leave the hold-up (`0x2010E4`, `0x96` to `0x97`), in first person too. The line checks `0x10B4B0`/`0x108890` are map queries and never report an actor; the mod's own ray cannot name a guard.
- First person. There is no view toggle: the rifle component (`0x359E50`, ids 9, 11, 12) and the pistol component (`0x359280`, id 7) read L1 held every frame (register masks at `0x359E8B`, `0x3592D8`; tests at `0x35A444`, `0x35A612`, `0x359613`), iron sight while held, the gun 80 to the side otherwise (table `0xB1E870`). Status `0x27` forces the side view (`0x35A4DE`). The shot ray always runs from the first-person eye `[actor+0x730]+0x20` along its forward axis (`0x373C10`, `0x3749CA`), so the aim is the screen centre in both views.
- Wall-press component: hash `0xD9728F`, handler `0x34F950`, state function at node `+0x100`, camera record at node `+0x120` (enable bit `0x400` at record `+0xC`; with it on the record owns channel 0 and the third-person camera reports mode 5), wall yaw at `+0x232`. Statuses `0x3C` (corner peek) and `0x3C` plus `0x3D` (the pop-out aim hold) are set by its states every frame. The pop-out event writes `+0x800 = +0x16A + 0x8000` (`0x34D8A3`); the aim hold copies `+0x800` into `+0x16A` every frame (`0x34D4B3`, crouched `0x34C241`); pressed flat, the body is pinned to `+0x232` (`0x34FD27`). The native corner aim is auto-aim only (the aim helper skips the stick while `0x3C` is set, `0x37A865`). The first-person cone from a wall is `+0x232` plus or minus `0x31C7` (`0x34FDF9`, `0x34FE00`).
- Phase 4. Pause bit 4 stops actor levels 4, 5, 6, 8 and 9 (list masks at `0xACE670`); the player is level 5 and both windows are level 10, which bit 4 does not stop, and the windows skip only while bit 2 is set. So the windows could run unpaused if the two pause calls were skipped and Snake's movement hidden; no web source says whether Delta's windows pause. The codec opener `0x2EEAF0` reads a Select press and sets pause bit 1 (`0x2EED09`); the survival viewer `0x302E20` reads Start. Delta New Style's D-pad per the release charts (Game8, MagicGameWorld, Destructoid): up camouflage, right weapons, down radio, left equipment; Prima's chart differs. The controller buffer layout (`0x114F90`): low word Select 1, L3 2, R3 4, Start 8, up `0x10`, right `0x20`, down `0x40`, left `0x80`; high word L2 1, R2 2, L1 4, R1 8, Triangle `0x10`, Circle `0x20`, Cross `0x40`, Square `0x80`; pressure words 0 right, 1 left, 2 up, 3 down. QCamo reads the physical pad itself (XInput here), so the controller buffer cannot open it.
- CQC from a crouch: correct that no native path strikes or grabs from a crouch (the strike state `0x373560` is stored only at `0x377EAF` and `0x379D58`, both after TEST(1); the CQC poll refuses statuses 2 and 3 at `0x387711`). A Cross release in the crouch sub-state stands Snake up in the same frame (`0x36A590`, `0x3687B8`); status 1 follows on the next frame's message `0x80`, so a Circle press sent two frames after the fire press strikes on the third.
- Reach check (option C): the game's own test cannot be run read-only (the CQC probe `0x360AD0` claims the contact slot and tags enemies); an approximation from the class-2 target list would need a calibration trace. The user kept option A (below), so it was not built.
- Crouched sideways animation: the strafe motions are picked by `0x36A690` (`0x21`/`0x22` sideways at a run, `0x25`/`0x26` slow), all standing; no crouched lateral motion was found (Crouch Walk's `SquatCover=98` is never written), and the only movement motion send is the call at `0x369972`. A stand-in would show Snake standing while he counts as crouched.
- Face paint: QCamo 1.0.4 already changes face paint (its face-only path matches the game's own Survival Viewer change `0x300E50`), but it picks the best face paint automatically with every uniform. The equipped face is `[[0xACDE98]]+0x67F`; face f is item `0x4A + f`; its label index 4 is DESERT in the executable, not MOUNTAIN. Upstream QCamo builds with MSVC `cl` on this machine.
- Weapon table: id k's flags dword is at file `0xAC67B4 + k * 0x38`. Only ids 0, 1, 5, 11, 14, 19 and 21 have appeared in the play logs (`grep -o 'ctx weapon id=[0-9]*'` over `work/fpv-move-trace-20260924/*.log`). The flag-`0x20000` (charge) weapons are ids 19 to 24 and 31; `pad_standard` leaves them all native, so no mod rule touches them. Names beyond the six confirmed are not in the executable.

User decisions on 2026-09-26: CQC from a crouch stands Snake up, then CQC; the reach check stays option A (no hip fire without aiming); a crouched sideways stand-in only if a crouched motion exists (none does); the first-person crosshair only for the guns with the iron-sight switch (ids 7, 9, 11, 12); the wall aim keeps the native wall camera with a movable aim; Phase 4 on the D-pad with the native paused windows; the face paint as a QCamo fork, with D-pad up opening its menu and Left/Right switching between camouflage and face paint.

### 0.8.0 (the backlog batch)

Built on 2026-09-26. fpv-move changes (`native/fpvmove/fpvmove.c`):

- Hold-up drop (backlog item 1). The three hold-up status tests (`0x2010AA`, `0x202D19`, `0x2015EE`) go through `holdup_gate`, which also passes while the shoulder aim is live (the lock, the rig on and in reach, not first person). The one call of the aim line's muzzle matrix (`0x382BAE` to `0x3649F0`) goes through `aimline_wrap`, which, while the shoulder aim is live and the drawn laser and first-person bits are clear, puts the line on the rig's ray from the ray point nearest the muzzle, as converted shots start. Everything after the gate is the game's first-person behaviour. Counters `holdup`, `aimline`; the trace line `hold-up: status 0xBA passed for the shoulder aim` (at most every 2 s).
- First-person crosshair (item 2). In first person with weapon id 7, 9, 11 or 12, in the weapon-at-right view (L1 not held, or status `0x27`), a crosshair is drawn at the centre of the game's view. It follows each component's own gates, and for id 7 none while the item is id 6 or 9 (the view flag freezes then, `0x3625D0`). Trace `ctx first-person crosshair on/off`; `xh centre` lines.
- Wall hug and corner peek (item 3). With the shoulder camera on and camera mode 5 owned by the wall-press record, the aim reason is `wall`: the native wall camera stays; the mouse or right stick turns an aim yaw inside `+0x232` plus or minus `0x31C7` when pressed or peeking, and inside the pop-out's own aim plus or minus `0x31C7` in the aim hold; the pitch is the rig's. A ray along that aim, from Snake's head (from the muzzle in the aim hold), gives the rig's hit, so the crosshair, the arm pose and the shot conversion work as in the open. Only the aim hold gets `+0x800` written. Which way a yaw increase moves on screen is read from the camera matrix every frame; with none for 10 frames the native auto-target aim takes that wall camera. Auto-target is off in the wall aim unless F8 or an easy save turns it on. Trace `wall aim: pressed / corner peek / pop-out aim`, a once-a-second `wall aim case=` line (yaw, base, pitch, origin, hit distance), and `wall aim: wall yaw ..., wall check ...` (the check only logs).
- CQC from a crouch (optional backlog). A fire-key press (left mouse or RT) while crouched, not aiming, with a weapon that can CQC, in an active movement crouch sub-state (not in a wall hug), with no other Cross rule in flight, holds Cross for one frame and releases it (the stand-up), then, once the movement sub-state is stand or move, becomes the CQC rule's Circle for as long as the key is held (at least one frame, so a tap punches). Square stays hidden from the press until the key is let go. Crouched hip fire and crouched knife slashes are replaced (accepted by the user). Counters `cqccrouch`, `standmiss`; trace `crouch CQC: standing at frame N, Circle` or `not standing after 6 frames`.
- Phase 4, pad D-pad. In gameplay D-pad left sends L2 (a tap toggles the equipped item, a hold opens the item window; browse with the stick), D-pad right R2 (weapons), D-pad down Select (codec; it keeps its role into the codec), and D-pad up is hidden from the game (the QCamo fork opens its menu on it). LB and RB are native again: L1 (lock-on, camera reset, and the first-person iron sight while held) and R1 (first person while held). The 0.7.0 trigger-pending rule now keys on the D-pad window buttons in their window role. The corner peek on the pad is D-pad left/right (the wall's corner-view masks are L2/R2). D-pad walking and stalking are gone on the pad (the stick moves). The keyboard is unchanged: 1 and 2 hold the item and weapon windows, Esc is the codec, Tab the survival viewer, G the QCamo menu.
- Motion probe (log only). The layer-0 motion that the wall-press component asks the motion component (hash `0x1B9F95`, request at node `+0x58`, `0x28` bytes per layer) for is logged on change (`wall motion N archive=...`), to name a crouched wall shuffle motion for a later crouched sideways strafe.

QCamo fork (`native/qcamo-face/`, package `qcamo-face` 1.0.4-face.1, target `qcamo.asi`), approved by the user: a patch on upstream 0a8bee48 plus a `build.ps1` that fetches the pinned upstream, ImGui and MinHook, applies the patch, and builds with MSVC `cl`. A FACE PAINT list beside CAMOUFLAGE (Left/Right, A/D, the D-pad or either stick switch lists), equipping a uniform keeps the worn face paint (Tuxedo still forces none), holding D-pad up opens the menu (release closes; the right stick selects while the left thumb holds D-pad up), and face label 4 reads DESERT. The manager refuses to enable it while `qcamo` is enabled (same target). Its notices are written beside the package (`work/qcamo-face-notices/`), since a package may hold only the manifest and payload. It must be enabled only together with fpv-move 0.8.0 or later: earlier fpv-move sends D-pad up to the game as R3 (the camera view).

Reviews before install (all Opus, high effort, each told to verify against the disassembly and not to trust comments or research):

- fpv-move, two reviewers. Hold-up gates, aim line, first-person crosshair, D-pad layout and the motion probe: SHIP. Wall aim: FIX FIRST (the screen direction was taken once and could invert or stay 0; the pop-out ray started at the body, which may stand behind the corner; the two-way wall check could flip the cone into cover; line checks outside SEH). Crouch CQC: FIX FIRST (in a crouched wall hug movement is suspended with its crouch sub-state left behind, so a fire press would have sent a Cross tap to the wall state; missing 0xDF/0x3B guards; other Cross rules in flight). Also fixed: the pistol's own view gate and item freeze, the trigger-pending rule counting menu D-pad releases, the stick scaling in the wall aim. One verifier checked the fixes, found one more (the direction test at a fixed 300 units still inverts for far hits under a wall camera beside the aim) plus five minor ones, then checked those: SHIP. Its last minor note (the direction flipping at an edge between a near and a far hit while turning) was then fixed by taking the direction only while the aim is not being turned; that three-line change was not re-reviewed.
- QCamo fork, one reviewer: SHIP on the condition above; its minor findings (L1 kept a D-pad-up menu open, stick chatter at the threshold, Tuxedo rows showing gains, the fallback hint width, the close log, the Steam D-pad action names, the build tree note) were fixed by the builder and re-verified, and the fork rebuilds byte for byte from the patch on a clean pinned tree.
- Accepted: the wall aim's crosshair shows the ray's hit while pressed flat, where only the arms turn; a shot the muzzle cannot bring within about 45 degrees of it stays native (the `shot ... native: muzzle` line). In the pressed and peek cases the aim point builder's yaw is set to the native `+0x800`. Ids 10 and 13 get no first-person crosshair (no iron-sight switch). The hold-up gate opens every held-up soldier to first-person rules, which include leaving the hold-up after three shake cycles.

Installed on 2026-09-26 with the game and launcher closed: `pack`, `disable fpv-move`, `remove fpv-move`, `add work/fpv-move-0.8.0.zip`, `enable fpv-move --dry-run`, `enable fpv-move`; then `disable qcamo` (kept in the library for a rollback), `add work/qcamo-face-1.0.4-face.1.zip` (packed by the builder; the re-pack refused to overwrite it), `enable qcamo-face --dry-run`, `enable qcamo-face`, `verify`: all exit 0, `ok: true`, `compatible: true`, manager generation 136. Installed files: `fpvmove.asi` 224,768 bytes, SHA-256 `d2152f9954742375afc260d9ec4c386a75974ce0ad3e00612d65bedd3756561a`; `qcamo.asi` (qcamo-face) 752,128 bytes, SHA-256 `f3a7de1c9545d998d8c97513e566bb2bdaab8d465b2f9eb8a8cd7b57b79af9d1`. The fork patch is `native/qcamo-face/qcamo-face.patch`, SHA-256 `350b3ce69c69e5a6a3df441024fcd8ef6dc7473e6502508e207b01949c06ab64`. Not played yet.

Test list for the one play session (keyboard first, then the pad; no F9 needed):

1. 0.7.0 pad items: (a) run, tap A: crouch; (b) run, hold A: roll, and holding on: roll into prone; (c) tap A twice fast while running: crouch then stand, no roll; (d) hold D-pad right (weapon window), pick, release and press LT at once, also LT pressed before D-pad right is released: aim, no item window, no item toggle; the same with RT; (e) standing, crouched and prone A unchanged.
2. Phase 4 pad D-pad: tap D-pad left (toggles the equipped item), hold it (item window, browse with the stick), the same with D-pad right for weapons; D-pad down (codec); LB (lock-on and camera reset; in first person, hold for the iron sight); RB (first person while held); at a wall corner, D-pad left/right peek.
3. QCamo fork: keyboard G opens, A/D or Left/Right switch CAMOUFLAGE/FACE PAINT, change the face paint, then change the uniform (the face paint stays); pad: hold D-pad up, right stick to select and switch, A to equip, release to close; Tuxedo; uniform changes while wearing the Mask face paint (twice, then an area change).
4. Hold-up: hold up a guard from behind, aim at his head with the shoulder aim (pistol): he should shake and drop an item as in first person.
5. First person with the automatic rifle: a crosshair at the centre with the gun at the right; hold M (pad LB): iron sight, no crosshair; shots land on the crosshair.
6. Wall aim: pressed flat against a wall with the gun raised (right mouse or LT): the crosshair shows and the mouse or right stick moves it left, right, up and down (about 70 degrees each side); shoot. Then at a corner: peek, pop out with the fire key, aim and shoot.
7. Crouch CQC: crouched, walk up to a guard and hold left mouse (RT): Snake stands and grabs; tap it: stands and punches; also with the knife and empty hands; crouched in a wall hug, left mouse keeps its native use.
8. If convenient: crouch against a wall and shuffle sideways (the log names the motion for a later crouched sideways strafe).
9. Regression of the shoulder aim: pistol and automatic rifle shots on the crosshair, crouched strafe, first person, holding a guard, standing CQC, the knife, grenades, the pad (LT, RT, stick clicks).

Play, 2026-09-26, first run (crashed; to handle later): the game crashed when the user equipped the SWEDEN face paint in the QCamo fork's face list; its thumbnail was blank (its icon file `00433670` is one of the two missing from `textures/flatlist/_win`). `qcamo.log` at 17:12:54: `face 20 SWEDEN equip accepted`, `face-only change 10 -> 20 pending` (from MASK), `face 20 applied; settling`, then `exception code=C0000005 ... offset=C8187 ... phase=face uniform=9 face=20`, write address 0x444. 0xC8187 is the crash upstream QCamo's notes record for a model change run inside actor jobs. Logs kept locally: `work/fpv-move-trace-20260924/fpvmove-0.8.0-play-crash1.log`, `qcamo-0.8.0-play-crash1.log`, `qcamo-crash-0.8.0-play-crash1.dmp`. The game was relaunched to continue the test list.

Second crash, same run of tests: the user put MASK on (17:15:58, `face-only change 1 -> 10`, fine), then chose WOODLAND, which had worked before (17:16:16, `face-only change 10 -> 1 pending`, `face 1 applied; settling`, then the same `C0000005 ... offset=C8187 ... phase=face`, write address 0x444). So the cause is a face-only change away from MASK (face 10), not SWEDEN; SWEDEN's blank thumbnail is only its missing icon file. Upstream QCamo can reach the same path (its automatic face pick can change the face away from MASK with the uniform unchanged), but rarely. To handle later: compare the fork's leave-MASK order with the game's Survival Viewer (state 2 at `0x300E85` writes the new byte before `0x1A000F` when the current face is 10; the refresh `0x1A0014` rebuilds the MASK node from the byte at `0x36B7A6`), and until then refuse a face-only change away from MASK in the fork (or route it through a uniform reload). Logs: `fpvmove-0.8.0-play-crash2.log`, `qcamo-0.8.0-play-crash2.log`, `qcamo-crash-0.8.0-play-crash2.dmp`. Relaunched again.

### Still open after 0.8.0

From an open-items sweep of this plan (Sonnet agent, rechecked in the main session); each is either a play item in the list above or cannot be done here:

- Valve Steam Input on a retail install: not testable here. A Steam copy's executable is SteamStub-wrapped on disk, so the manager refused it until v0.4.0-alpha.1, which also accepts the wrapped executable of the same build (see [delta-controls-untested.md](delta-controls-untested.md)); this install drives the pad without Valve Steam Input. (Ultimate ASI Loader loads plugins from hooked startup calls after the entry point, so a SteamStub image would already be unpacked when fpv-move scans; the hash pin is the blocker.)
- Charge weapons (flag `0x20000`, ids 19 to 24 and 31): no mod rule touches them. Only grenades (ids 19, 21) have been played; the others need nothing new but are unplayed.
- Weapon ids never seen in a play log: 2 to 4, 6 to 10, 12, 13, 15 to 18, 20, 22 to 31 (including the lock-excluded ids 15 to 17 and the no-strafe id 8). Their names are not in the executable; play them when the save has them.
- Pad L3 while holding a guard (interrogate) and the QCamo pad chord were never played on their own; the chord now sends L1 plus Triangle for a frame before QCamo pauses.
- The unconverted player spawner call at `0x3533A0` (weapon unknown) and a shoulder pivot inside a wall (0.6.4 review) are still open; no play report has shown either.
- The global `0x1E16D64` beside the movement flag and the two grab/throw animation variants stay unidentified; nothing depends on them.
- The crouched sideways strafe animation needs new animation assets; the 0.8.0 wall-motion log line is the first step if that is ever wanted.
- Live (unpaused) item and weapon windows: feasible in principle (see "Phase 4"), not built.

Played on 2026-09-26 from 17:20 to 17:30 (third run, after the two crashes above), all on the pad. The log is kept locally at `work/fpv-move-trace-20260924/fpvmove-0.8.0-play.log` (4,636 lines, no fault, dropped or depth line), with `qcamo-0.8.0-play.log`. The user reported: the tests passed except as below, and the QCamo menu controls (right stick) are good. Requests and findings:

- The item and weapon windows are hard to browse with the left stick while the thumb holds the D-pad: add the right stick, keep both.
- First person is hard on the pad: LT to aim, RB held for first person and LB held for the iron sight. The log confirms it: no `pad first person toggle` line, so RB (native R1, a hold) was used, not the left stick click toggle. Make it a toggle, or find easier buttons.
- Hold-up: aiming at the head makes the guard shake; aiming at the crotch does not, while it does in first person (the user also said "default"). The log has `hold-up: status 0xBA passed` and the `holdup`/`aimline` counters running at 17:22:31 to 17:22:42.
- Prone in grass switches the camera to first person; wanted: keep third person in grass, keep the first person where Snake is inside something.
- From the log: crouch CQC stood and grabbed (`fire press id=5 ... crouch=1`, `crouch CQC: standing at frame 3, Circle`, then `ctx holding on`); the wall aim ran through a corner peek and pop-out with converted shots (`wall aim case=2`, `case=3 ... from=muzzle hit=1`); 32 converted shots and 2 native (`muzzle 51 deg off`, `60 deg off`); the wall motion probe named the wall shuffles: standing idle 98, standing shuffles 100 and 101, crouched idle 99, crouched shuffles 102 and 103 (base archive `0x6891CC`, sender the wall component).

### 0.8.1 (from the 0.8.0 play)

- Right stick in the windows: while D-pad left or right, held in from gameplay, keeps its window role and the window is open (gameplay stopped), the right stick words of the controller buffer (`+0x0C` X, `+0x0E` Y, the `ingame_stick_cam_dir` action written at `0x1151DC` and `0x115219`) also send the D-pad directions the windows browse with; hysteresis at `0x40` on, `0x28` off.
- First person on the pad: RB now toggles first person, like the left stick click (natively R1 is a hold). In first person with a gun that has the iron-sight switch (ids 7, 9, 11, 12), the aim (LT, or right mouse) also holds L1, so the iron sight comes up with the aim: first person is RB (or the stick click) once, then LT to aim down the sight and RT to fire. LB still gives the iron sight on its own. The keyboard F stays the native hold.
- Hold-up: the three gates reach `holdup_gate` through a stub that copies `rbx` (the soldier at all three sites) into `rdx`. For a soldier seen in the last 150 ms, the aim line runs from Snake's eye height (position plus 676, as the first-person eye in the logs) to the camera ray's point at the soldier's depth, as the first-person line does, instead of along the camera ray from beside Snake. The likely cause of the crotch miss is that the camera ray comes from the shoulder camera's side and the upper side, so it may meet another body zone of the soldier first; not proved. A trace line logs the body part under the aim line (`hold-up: aim line on part N`; part 3 is the head, part 1 probably the crotch).
- Crouched sideways strafe (the stand-in the user approved on 2026-09-26): during a crouched strafe moving sideways (direction bit 8 or 4 at movement node `+0x68`), the movement handler's one motion send (`0x369972`) gets a copy of its word with layer 0 swapped for the crouched wall shuffle, motion 102 (bit 8) or 103 (bit 4). Which of the two is left is a guess. F6 turns it off and on. Counter `sidestep`.
- Prone in grass (research agent, Opus, verified here): the first-person component is always in the tree and enters on its own while status `0xBB` is set (`0x38B464`, `0x38BB20`); `0x37E5E0` sets `0xBB` while prone from ANY(`0x4F`, `0x50`), where `0x50` comes from the "grassintrude" map volume (hash `0x11C95D`) and `0x4F` from "intrude" (`0x7AFC08`, under a vehicle, bed or vent). The patch at `0x37E6AF` turns the test into ANY(`0x4F`), so grass alone no longer latches first person; `0x50` is still set. Lost in grass: an effect under ALL(`0x50`, `0xE`, `0xBB`) at `0x5639CA` (likely first-person grass particles) and a few `0xBB` readers of unknown meaning. The trace line `ctx intr prone= 4F= 50= BB= BA=` shows which volume each spot is.
- QCamo fork 1.0.4-face.2: a face change away from MASK is refused ("take the Mask off in the Survival Viewer"), as is Tuxedo while MASK is worn. The builder showed that routing it through the uniform reload would not remove the MASK node first (the model rebuild `0xC5560` copies node flags from the old list; only the `0x1A0014` refresh at `0x36B7A6` toggles the MASK node).

Review: one independent Opus reviewer was started on the 0.8.1 diff and stopped at the user's request before it reported. Checked in the main session instead: after each of the three hold-up TEST calls the caller never reads `rdx` before writing it (`0x2010B7` `lea edx,[rdi+1]`, `0x202D26` `mov rcx,rbx` then a one-argument call, `0x2015F7` `mov rcx,[rbx+1CB8h]`), and TEST itself clobbers `edx` (`8B D1`); the motion send's handler reads the word through `r9` during the call only; the grass patch bytes at `0x37E6AF` and the ANY's -1 terminator. Not checked statically: the right stick's Y sign in the controller buffer and which crouched shuffle is left.

Installed on 2026-09-26 with the game closed: fpv-move `pack`, `disable`, `remove`, `add work/fpv-move-0.8.1.zip`, `enable --dry-run`, `enable`; qcamo-face `disable`, `remove`, `add work/qcamo-face-1.0.4-face.2.zip`, `enable --dry-run`, `enable`; `verify`: all exit 0, `ok: true`, `compatible: true`, manager generation 144. `fpvmove.asi` 227,328 bytes, SHA-256 `ddf576c12e93cc17a4630677d7fbe0016a3f15c59038f03f52d365b4cf7b1587`; `qcamo.asi` (qcamo-face 1.0.4-face.2) 752,640 bytes, SHA-256 `4256f11c5190c04ff0163cfd8f0af0e4e256ca06f25777aa335223d2b0e1c4d6`; fork patch SHA-256 `fe299fd469dae2c4fa05b97738413e96269d9e85b5aa089b384d666f2765b217`. Not played yet.

Play of 0.8.1, first run, 2026-09-26 22:18 to 22:24 (crashed; to handle later): with MASK on, a uniform change in the QCamo fork crashed the game the same way. `qcamo-0.8.1-play-crash1.log`: `face 10 MASK equip accepted` (22:23:52, fine), three refused face changes away from MASK (the face.2 refusal works), then `uniform 27 face 10 queued`, `uniform phase applied; face 10 pending`, `face 10 applied; settling`, `C0000005 ... offset=C8187 ... phase=face uniform=27 face=10`, write address 0x444. So the uniform path while MASK is worn (10 to 10) crashes too; the face.2 claim that it is safe was wrong. To handle later: refuse every change while MASK is worn (uniforms too), or find the MASK node teardown that the Survival Viewer gets from its pause bit 1. Logs: `fpvmove-0.8.1-play-crash1.log`, `qcamo-0.8.1-play-crash1.log`, `qcamo-crash-0.8.1-play-crash1.dmp`. Relaunched.

Test list for 0.8.1 (pad): (1) hold D-pad left or right to open a window, browse with the right stick (and the left); (2) RB toggles first person on and off, and the left stick click too; in first person with the automatic rifle, LT brings the iron sight, RT fires, release LT for the weapon-at-right view with the centre crosshair; (3) hold up a guard, aim at the crotch, then the head, with the shoulder aim: both should make him shake (the log line `hold-up: aim line on part N` names what the line touches); (4) crouched with a gun and LT, strafe left and right: the legs play the crouched wall shuffle; say if the feet step the wrong way; F6 on the keyboard turns it off to compare; (5) prone in grass: third person stays; prone under a vehicle or bed: first person still comes; (6) QCamo: with MASK on, other face rows are dimmed and refused; take MASK off in the Survival Viewer; (7) a short regression: shoulder aim shots, CQC, crouch CQC, wall aim.

### 0.8.2 (from the 0.8.1 play)

Played 0.8.1 on 2026-09-26 from 22:25 to 22:33 on the pad (second run after the MASK crash above). Log kept locally: `work/fpv-move-trace-20260924/fpvmove-0.8.1-play.log` (2,643 lines, no fault or dropped line), with `qcamo-0.8.1-play.log`. From the log: RB toggled first person (`pad first person toggle on (RB)`), the iron sight came with LT (`iron sight on the aim id=11`), prone in grass kept third person (`ctx intr prone=1 4F=0 50=1 BB=0 BA=0`), the crouched sidestep ran (218 frames). The hold-up crotch aim was not played. The user reported:

- The right stick does not browse the item and weapon windows. Likely cause: while a window is open the game's Steam Input read reports the camera stick centred (another action set), so the buffer words stay at 0x80.
- In first person with the automatic rifle, releasing LT holsters the gun instead of going back to the weapon-at-right view with the crosshair. The log confirms it: `ctx first-person crosshair on` then `ctx gun down` 0.3 s later (22:32:48.9, 22:32:49.2): the aim release ran the lower rule.
- The sawed-off shotgun (id 14) has no iron sight and no crosshair; it is always at the right.
- Wanted: the QCamo menu on D-pad up without freezing the game. The user found a way to keep it running (holding LT as well) and changed camouflage and face paint that way without trouble.

0.8.2 (fpv-move) and qcamo-face 1.0.4-face.3:

- First person: with a gun shown at the right (ids 7 and 9 to 14), releasing the aim keeps the gun up (`first person: aim released, gun kept up`); the kept gun counts as aimed for the fire and CQC rules; leaving first person lowers it (`first person off: gun lowered`). The iron sight still follows LT (real L3 only).
- Crosshair: ids 10, 13 and 14 (no iron-sight switch, always at the right) show the centre crosshair whenever the gun is up.
- Windows: the right stick is read through XInput (`xinput1_4`, slots 0 to 3, as QCamo does) when the buffer words are centred.
- Live QCamo menu: the fork no longer pauses for a menu opened by holding D-pad up; it pauses only while a change runs (`change pause on/off`); G and L1+Y still pause as upstream; a game window, the codec or a cutscene closes the live menu. While D-pad up is held in gameplay, fpv-move hides A (Cross), the right stick and the other D-pad directions from the game (they browse the menu); the left stick still moves Snake.
- MASK: the fork refuses every change while MASK is worn (uniforms too), after the second MASK crash.

Review: none by a separate agent for 0.8.1 to 0.8.2 (the user asked to wrap up); the builder checked the fork. Installed on 2026-09-26 with the game closed, the same sequence as 0.8.1 (fpv-move and qcamo-face each disabled, removed, added, dry run, enabled; `verify` `ok: true`, `compatible: true`), manager generation 152. `fpvmove.asi` 229,376 bytes, SHA-256 `25dba57a583f67b52ee0873ff9149561831f6ff1cfefda36277d79fefcac23b8`; `qcamo.asi` (qcamo-face 1.0.4-face.3) 752,640 bytes, SHA-256 `7826bf46c36ec088f5e1ba0064d43fa80620ba613832c1ea581ccc4a57fd5e5f`; fork patch SHA-256 `eb8c6a3199559e9cf7342b101f59b827649277da34128b7c777c5b282720506b`. Not played yet.

Test list for 0.8.2 (pad): (1) hold D-pad left or right to open a window, browse with the right stick; (2) first person with the automatic rifle: LT iron sight, release LT: gun stays at the right with the crosshair, RT fires; RB off: gun lowered; (3) the shotgun in first person: crosshair at the centre; (4) hold D-pad up: the game keeps running, right stick selects, A equips (Snake does not crouch), release closes; (5) with MASK on, every other row is refused; (6) the hold-up crotch aim from the shoulder (`hold-up: aim line on part N`); (7) a short regression.

### 0.8.3 (from the 0.8.2 play)

Played 0.8.2 on 2026-09-26 from 22:42 to 22:45 on the pad. Log kept locally: `work/fpv-move-trace-20260924/fpvmove-0.8.2-play.log` (2,410 lines, no fault or dropped line), with `qcamo-0.8.2-play.log`. From the log: the first-person gun stays up after LT (`first person: aim released, gun kept up id=11`, then `first person off: gun lowered`). The user reported: the hold-up crotch aim from the shoulder works well now (0.8.1's eye-height aim line); the live QCamo menu on D-pad up works without freezing the game; but Snake still moves and acts while the menu is used, and the right stick still does not browse the item and weapon windows.

0.8.3: while D-pad up holds the QCamo menu in gameplay, fpv-move sends the game no buttons (the first-person toggle's R1 stays) and centres both sticks, so Snake stands still while the world runs. The window stick gets a probe: `win stick buf= xinput= dir= lo= hi= state=`, on each change and every 64 reads, to show whether the controller read runs while a window is open and what the stick sources give. No review agent. Installed on 2026-09-26 (fpv-move disabled, removed, added, dry run, enabled; `verify` `ok: true`, `compatible: true`), manager generation 156, `fpvmove.asi` 229,376 bytes, SHA-256 `b840ef87badb8ad5f9e4f72af3087c7173be3d5b3e2b9724c006175cdf94adb8`. qcamo-face 1.0.4-face.3 stays. Not played yet.

Still open: the right stick in the windows (the probe decides the fix); the hold-up is closed.

Played 0.8.3 on 2026-09-26 from 22:55 to 22:57. Log: `work/fpv-move-trace-20260924/fpvmove-0.8.3-play.log` (343 lines, no fault). The probe answered: the controller read runs while a window is open (`state=0`, `hi=0001` for the held L2, `0002` for R2), and the right stick is seen and mapped (`win stick buf=127,0 xinput=127,18 dir=1`, up `dir=3`, down `dir=4`, left `dir=2`; both the buffer and XInput report it), yet the window did not browse on the added D-pad bit.

0.8.4: with a direction, the window stick also sets that D-pad direction's pressure word (index 0 right, 1 left, 2 up, 3 down) and, while the physical left stick is centred, writes the direction into the move stick words `+0x10`/`+0x12` (the left stick already browses the windows). Installed on 2026-09-26, manager generation 160, `fpvmove.asi` 229,888 bytes, SHA-256 `5438ace896a2cad9657981051664094cda6386bf747a49076e5dbbfbe31a09b7`. Played 0.8.4 on 2026-09-26 around 23:05. Log: `work/fpv-move-trace-20260924/fpvmove-0.8.4-play.log` (944 lines, no fault; 32 `win stick` lines). The user reported that it works great: the right stick browses the item and weapon windows. The user also confirmed that 0.8.3's D-pad-up menu holds Snake still while the world runs and works great. The window-stick probe line stays in the build (it logs only while a window is open).

### Taking the Mask off from the QCamo menu (done in face.4, 2026-09-27)

Wanted by the user: remove MASK from the D-pad-up menu instead of the Survival Viewer. qcamo-face 1.0.4-face.3 refuses every change while MASK is worn, after three crashes at game `0xC8187` (write to `0x444`). Research by the fork builder, from the stack words QCamo logged in all three crashes (no debugger on this machine):

- The faulting actor is `0x1D55FF0`, registered at `0xBCE34`..`0xBCE50` in actor list 0 (`0x10F920`: list = descriptor >> 56, minus 1); list 0's mask is 0, so no pause value stops it. Its update `0xBCC00` leads to `0xC8177`..`0xC8187`: `rax = 0xC3A10([[rbx+0A0h]+10h])`, a lookup of a loaded resource by id that returns 0 when none matches, then `mov word [rax+444h],r14w`. So a model part still references a resource id that is no longer loaded.
- The Survival Viewer holds pause value 1 (`0x302F28`) and sends `0x1A0014` on close (`0x303190`) before clearing it; value 1 skips lists 2 to 10 only, so the Viewer's pause is not what protects it (not traced further).
- Every crash falls between the face asset reload and the `0x1A0014` refresh, the only place the MASK node is switched off (`0x36B7A6`, `0xC66F0` on node `0x74DDDA`); putting MASK on never crashed (the node stays hidden until the refresh after the load).
- Candidate, not built: hide the MASK node first. Leaving MASK: after the new byte and `0x1A000F`, send an extra `0x1A0014` in the same tick, then load and apply the face on the next tick, then the final refresh. A uniform change while MASK is worn: an extra `0x1A0014` after `0x1A000F` (the uniform phase writes face 0 first), before the face-10 reload. Refuse both in first person (the refresh handler shows the node there, `0x36B7BD`..`0x36B7D7`). Build it as 1.0.4-face.4 and test it once; if it crashes, keep the refusal.

Built and installed on 2026-09-27 as qcamo-face 1.0.4-face.4 (played and passed on 2026-09-27, below): with Mask on, the menu equips again; a change leaving Mask sends an extra `0x1A0014` right after `0x1A000F` in the same tick (log `mask node hidden before the face load`), then the face load runs on the next tick and the final refresh on the one after. Still refused with Mask on: every change in first person (status 7, read from `0x1E16CF8`; the gameplay thread checks again), and the Tuxedo (untried). Not proved: that the face resource the load replaces is the one the Mask node points at; this build is the test. Review: one independent Opus reviewer found no blocking defect; its three low findings were fixed (a stale README bullet, the claim that `0x1A0014` is the only message that hides the node: the component at `0x383670` also drives it from bit `0x10` and keeps it hidden on every update once the refresh clears that bit; and the refusal sound on the gameplay-thread first-person refusal). It noted that the native Viewer never sends `0x1A0014` in the same tick as `0x1A0002` (the uniform path now does; no static conflict found), and that the name "first person" for status 7 is not proved (`0xBA` is the first-person view flag). Installed with the game closed: `disable`, `remove`, `add work/qcamo-face-1.0.4-face.4.zip`, `enable --dry-run`, `enable`, all exit 0; `verify --json` `ok: true`, `compatible: true`, manager generation 164. `qcamo.asi` 753,664 bytes, SHA-256 `6c40bdaf6ddff2138294c02a1afa1b5e833afc23c4c929db0f99abe9e5d4a44a`. If it crashes, reinstall face.3 (`work/qcamo-face-1.0.4-face.3.zip`).

Test list for face.4 (third person, save first): (1) put Mask on from the menu; (2) Mask to another face paint; (3) Mask back on, then a uniform change keeping Mask; (4) with Mask on, the Tuxedo row is dimmed and refused; (5) with Mask on in first person, every row is dimmed and refused. After play, copy `qcamo.log` and any `qcamo-crash.dmp`.

Played face.4 with fpv-move 0.8.4 on 2026-09-27 from 17:50 to 17:58 on the pad. Logs kept locally: `work/fpv-move-trace-20260924/qcamo-face4-play.log` and `fpvmove-0.8.4-face4-play.log` (2,378 lines, no fault, dropped or depth line). The user reported that everything works perfectly. From `qcamo.log` (counted with `grep -c`): 29 changes completed, no exception; 16 `mask node hidden before the face load`: 6 face-only changes from Mask (`face-only change 10 ->`, to faces 5, 0 and 1) and 10 uniform changes keeping Mask (`uniform phase applied; face 10 pending`, uniforms 27, 26, 1, 11, 3, 7, 9, 25); 3 refusals in first person (`leave first person to take the Mask off`). The Tuxedo refusal was not reached (no uniform 16 line). So hiding the Mask node first stops the `0xC8187` crash on every path tried. From `fpvmove.log`, the old 0.8.0 play items: first person toggled on and off 4 times with RB, one standing CQC (`crouch CQC: standing at frame 3`, holding on/off 9/10), 17 `wall aim` lines (1 corner case 2, 16 case 3), D-pad left, right and down presses (windows and codec), and 18 aim presses. The user closed the game with Alt+F4 at 17:58; no exception was logged at that exit.

### Exception at game RVA 0x45D9A (end of three sessions)

`qcamo.log` recorded the same access violation at the end of the 0.8.0, 0.8.1 and 0.8.4 plays (17:30:07, 22:35:39, 23:02:13 on 2026-09-26), with QCamo idle and the same return addresses on the stack (`0x8946F`, `0x8889C`, `0x888FC`, `0x874D0`, `0x8A8DD`, `0x86B82`, `0x29307`). The code at `0x45D8A`..`0x45D9A` takes a handle, indexes the table at `[0x141B9F4A0]` (stride `0xF8`), and for handle type `0x500` calls a virtual method on the object at entry `+0xF0`, which was null (read of address 0). QCamo's handler is a first-chance vectored handler, so the log line alone does not prove a crash. It never showed up in the middle of play, and the 0.8.4 run passed. Most likely an exit-time fault in the game's own code, not caused by the mods; not proved. The user did not remember how those sessions ended. The face.4 session, closed with Alt+F4, logged no exception, so a later session ended from the game's own quit could tell whether the quit path raises it. Dump kept locally: `work/fpv-move-trace-20260924/qcamo-crash-0.8.4-play.dmp`.

## Optional backlog

Set aside by the user after the 0.6.9 regression pass (2026-09-26), to handle later if complex:

1. Hold-up item drop. From the shoulder aim, a hold-up works, but aiming at the head does not make the guard shake and drop an item; it happens only in first person. Built in 0.8.0 (the first-person gates pass for the shoulder aim, and the aim line follows the crosshair), not played yet.
2. First-person view toggle. In first person with the automatic rifle, M (pad LB or L1) switches between the iron-sight view and the weapon-at-right view. Either find a good solution or draw a crosshair in the weapon-at-right view so the screen centre is known. It is an L1 hold, not a toggle; the aim is the screen centre in both views. Built in 0.8.0: a centre crosshair in the weapon-at-right view of ids 7, 9, 11, 12 (the user's choice), not played yet.
3. Wall hug and corner peek. No crosshair there now (0.6.9). Wanted: a crosshair and a movable aim. Built in 0.8.0 under the native wall camera, not played yet.
4. Pad timing. Pressing LT to aim right after a weapon switch sometimes opens the item window. Built in 0.7.0 (keyed on the D-pad window buttons since 0.8.0), not played yet.
5. Face-paint quick change in QCamo (noted before). Built as the QCamo fork `qcamo-face` 1.0.4-face.1 with 0.8.0, not played yet.
6. Pad crouch from a run. Pressing A (Cross) while running rolls; the user wants a tap to crouch and a longer hold to roll. Built in 0.7.0, not played yet.

- Shoulder aim on lock-on (proposed 2026-09-25 by the user; 5a built as 0.5.7, see "Step 5a, shoulder aim on lock-on"). While right mouse or pad L3 is held, hold L1 too, so Snake keeps his facing and strafes while aiming; then let the mouse turn that facing instead of the camera, and later add a crosshair. This is Target item 1, which 0.5.6 does not meet (it aims with the native camera). Proposed order: 5a, send L1 while the aim is held and the gun is up (a pad rule in the existing wrapper; check what else L1 does, such as wall and CQC contexts); 5b, a read-only search for the facing angle, where lock-on stores its direction, and a mouse delta source (input ids `0xCA`, `0xC4`, `0xCC`, `0xD2` are not aim deltas), ending in go or no-go; 5c, a crosshair only if 5b goes, since it needs a render hook the mod does not have, and a crosshair is wrong if third-person vertical aim is automatic (not checked).

- Next after 0.5.9 (requested 2026-09-26): vertical aim without full auto-aim (elevation assist first: the mouse keeps the yaw, a target inside a narrow cone of it gives the shot its height; manual pitch from mouse Y only with a visual indicator), full auto-aim kept for easy-difficulty saves if the difficulty is readable, a third-person crosshair placed on the shot's hit point (render path to be found; QCamo already draws an overlay), and a sideways crouched strafe animation. (Status 2026-09-26: vertical aim, the difficulty default and the crosshair were done in 0.6.1 to 0.6.9. The sideways crouched animation is a no-go without new animation assets: no crouched lateral motion exists, and the only stand-ins are standing strafe steps; 0.8.0 logs the wall-press motions so a crouched wall shuffle can be named.)
- Side note from the user: a quick face-paint change, like QCamo's uniform menu. Built as the QCamo fork with 0.8.0.

- CQC from a crouch. Native Circle does not strike unless status 1 (standing or moving) is set. A remap could send a Cross tap to stand, then Circle, but that is a new behavior with a visible stand-up, not a port. Only on request. (Requested and built in 0.8.0: stand, then CQC; not played yet.)

Not tied to a phase. Do only on request.

- CQC reach check (4b option C). Keep left-mouse hip fire and knife slashes out of reach, and grab only when an enemy is within grab reach. The game has no pre-press reach value ("Phase 3 step 4b, static pass"). This needs the enemy list and positions found by a trace, a distance and angle test tuned in play to the real grab reach, and a hook or poll near the contact handler `0x387B10`. Risk: where the test and the game disagree, a click fires a gun beside a guard or punches instead of shooting. It builds on option A and replaces only its "send Circle" condition. (Closed on 2026-09-26: the user kept option A. The game's own reach test cannot run read-only, and an approximation from the class-2 target list would need a calibration trace; see "Research for the 0.8.0 batch".)

### Phase 4. Shortcuts, optional

Closed as SKIP on 2026-09-22. Weapon and equipment open the pause flag. Camo stays on QCamo. Radio was not proved either way. No shortcut hook. Rechecked 2026-09-23: the item (`0x32DA44`, L2) and weapon (`0x32E6C1`, R2) calls (labels corrected 2026-09-25) pass `4` to `0x10ED10`, which ORs it into `0x1D78F6C`; the actor loop skips whole actor lists by mask. Nine callers pass `4`; none is proved to be radio. SKIP stands.

Up for the short camo list, left equipment, right weapons, down radio, gameplay still running. Skip if the game can only open the full paused screens. Skip a new camo UI while QCamo is the installed camo tool.

Reopened on 2026-09-26 at the user's request and delivered in 0.8.0 on the pad ("Research for the 0.8.0 batch", "0.8.0"): D-pad left the item window and quick item toggle, right the weapon window and quick weapon toggle, down the codec, up the QCamo fork's camouflage and face-paint menu; LB and RB native again. The windows stay paused, the user's choice. The "gameplay still running" condition is not a hard no-go: pause bit 4 does not stop the windows' own actor level, so skipping the two pause calls and hiding Snake's movement while a window shows could keep the game running; that is unbuilt and unplayed, and no source says whether Delta's windows pause. The keyboard already has the native keys (1 and 2 hold the windows, Esc codec, Tab survival viewer, G QCamo).

## Stop rules

- A missed required pattern ends that phase. It is not a reason to hardcode an offset from a GitHub issue. One MGSHDFix issue printed `aimingState` and `heldTriggers` offsets for a different run on 23 Nov 2025. Those offsets are not this build.
- Do not scan, patch, or hook while the game is running.
- Do not claim Delta movement from a flag that only unlocks Bluepoint's unfinished first-person camera.
- Do not treat Crouch Walk's 11 signatures as aim or CQC hooks.
- New Style bullet drop and awareness changes are balance patches in a different game. Leave Master Collection damage and detection alone.
- Every patch resolves by pattern at load, requires exactly one match in executable sections, checks the bytes it replaces, and logs a skip instead of guessing.
- The file-scan test pins file offsets only as a regression check for this hash. The plugin never stores them.
- This plugin and MGSHDFix keep-aim must not both run once step 2b ships.

## Open conflicts

Status 2026-09-26: none of these blocks anything. The first-person button was settled by the user (the left stick click toggles first person since 0.6.9; RB holds it natively since 0.8.0). The Legacy, style-switch and throw-knockout conflicts concern features this mod does not implement. `Renderer.dll` was never needed.

- First-person pad button in Delta: Prima and a player report say click or hold R3. Another chart says R1. Irrelevant until Phase 3.
- Legacy fire button: IGN says Legacy still aims and fires on the triggers and only locks standing first person. The PlayStation Blog describes Legacy as the old PS2 map. This plan follows the IGN split and still does not implement Legacy.
- Style switch reload: IGN says the last checkpoint. The PlayStation Blog says the start of the section. This mod does not implement that switch.
- From-hold CQC throw knockout differs between IGN and Game8. Preserve Master Collection's current throw result unless a tested hook proves a smaller change.
- `Renderer.dll` matches the pinned hash and did not parse as a PE. Hooks target `METAL GEAR SOLID3.exe` and, only if a pattern is absent there, `Engine.dll`.
