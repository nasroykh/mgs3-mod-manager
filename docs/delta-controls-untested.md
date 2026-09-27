# Delta controls: what has not been tested

Date: 2026-09-27. Scope: `fpv-move` 0.8.4 and `qcamo-face` 1.0.4-face.4. All play so far was on one install (see [delta-controls-plan.md](delta-controls-plan.md)): the pinned executable, the North America/English startup profile, a wired Xbox Series X pad without Valve Steam Input, and the keyboard and mouse. Since 0.6.9 every play session was on the pad only. Evidence for each line below is the saved play logs (`work/fpv-move-trace-20260924/`, local only) and the user's reports in the plan; "never seen" means no line for it in any saved log.

Keep this list current: when a tester or a play session covers an item, move it to the plan with the log evidence and delete it here.

## Not possible on this install

- **A retail Steam copy.** A Steam copy's `METAL GEAR SOLID3.exe` is SteamStub-wrapped on disk. Since v0.4.0-alpha.1 the manager also accepts the wrapped executable of the supported build (SHA-256 `81596a6a…`; same PE timestamp `0x6980B92F` and section sizes). Checked only with `doctor` on a copied folder; never run as a Steam copy. Not known: whether a current Steam download has exactly that file and the same four other fingerprinted files, whether `fpv-move` and the QCamo fork run there (both find their patches by pattern in memory, after the Steam wrapper has unpacked the code, so they should if the code is the same build), and whether the manager's `launch` works on a wrapped executable (start a Steam copy from Steam).
- **Valve Steam Input.** A retail copy reads the pad through the Steam client's Steam Input. Here the pad was played without Valve Steam Input. The whole pad layout (LT aim, RT fire, RB first person, D-pad windows and codec, D-pad-up QCamo menu, the right stick in the windows) has only run without Valve Steam Input.
- **Other game builds.** Any game update after the pinned build, other regions and languages, and the manager's other launch selections (rejected by the manager).
- **Other controllers.** PlayStation pads (DualShock 4, DualSense), wireless Xbox pads, Steam Deck, and other XInput devices. The window stick and the QCamo fork read XInput slots 0 to 3 as a fallback.
- **MGSHDFix.** Must not be installed beside `fpv-move` (both own `aimingState`); nothing checks this at install.
- **Other mods.** Only Ultimate ASI Loader 9.7.4 and Crouch Walk 0.2.1 were enabled beside these two.

## Weapons not reached in the save

- **Sniper rifles and scopes.** Scoped aiming with `fpv-move` (first person through the scope, the aim and fire rules, the shoulder camera around a scope) was never played. Weapon names are not stored in the executable, so which ids are the sniper rifles is not proved.
- **Weapon ids never seen in a play log:** 2 to 4, 6 to 10, 12, 13, 15 to 18, 20, 22 to 31. This includes the ids the lock-on rule excludes (15 to 17), the id with no strafe (8), ids 10 and 13 (centre crosshair in first person since 0.8.2; only id 14, the shotgun, was played), and id 12 (iron sight on the aim).
- **Charge weapons** (flag `0x20000`, ids 19 to 24 and 31). Only the grenades (ids 19 and 21) were played; no rule touches the others, but they are unplayed.

## Built but not played, or not confirmed

- **Keyboard and mouse since 0.7.0.** Every feature added in 0.7.0 to 0.8.4 was played on the pad only: crouch CQC on left mouse, the wall aim with the mouse, right mouse bringing the iron sight in first person (0.8.1), the first-person gun kept up after right mouse is released (0.8.2), the hold-up from the shoulder, and the QCamo fork on G (face paint list, A/D or arrows).
- **Tuxedo in the QCamo fork.** No saved `qcamo.log` has uniform 16: the Tuxedo refusal while Mask is worn (face.4), Tuxedo taking no face paint, and the dimmed face paint rows over Tuxedo are unplayed.
- **Hold-up aimed at the head from the shoulder since 0.8.1.** The crotch aim was confirmed on 0.8.2. The head shake and the item drop were reported on 0.8.0, before 0.8.1 moved the aim line to eye height, and were not rechecked.
- **Crouched sideways strafe direction.** It ran in play (0.8.1, 218 frames), but which of motions 102 and 103 is the left step was a guess, and the user did not say whether the feet step the right way. F6 turns it off.
- **Pad L3 while holding a guard** (interrogate), and **the QCamo pad chord L1+Y**, never played on their own.
- **An Easy or Very Easy save.** Auto-target turns on by default there; every logged save read difficulty 40.
- **Long play.** Area transitions, cutscenes, boss fights, and special sections (ladders, swimming, vehicles, the sniper duel) were not on any test list; sessions were a few minutes each.
- **The first-person grass effect.** Since 0.8.1 grass alone no longer forces first person while prone; an effect gated on the grass volume and first person (likely grass particles, `0x5639CA`) no longer shows there. Not looked for in play.

## Open questions with no play report

- An exception at game RVA `0x45D9A` was logged at the end of three sessions (plan: "Exception at game RVA 0x45D9A"). The Alt+F4 exit on 2026-09-27 logged none; an exit through the game's own quit was not checked.
- The unconverted player shot spawner call at `0x3533A0` (weapon unknown) and a shoulder pivot inside a wall (0.6.4 review) have never shown in play.
