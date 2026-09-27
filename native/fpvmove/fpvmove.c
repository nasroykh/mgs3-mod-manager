/* Phase 1 writes gBP_1stPersonCamera_EnableMovement once.

   Phase 2a (0.4.3). The first-person component calls its walk function
   every frame, armed or not. Three gates stop movement with a gun raised:
   walk exits on held Square 0x8000 (the pressure filter holds a fake
   Square while ready), walk exits on status 0x2E (weapon raised), and the
   look function turns the view from the left stick when 0x2E is set.
   This build clears those three gates and wraps the one walk call to log
   which early-out still stops it. Every site is resolved by pattern at
   load, must match once in executable sections, and has its old bytes
   checked before the write. A miss leaves that site unchanged.

   Phase 2b (0.5.0). The player pad component copies the raw pad and runs
   Konami's pressure filter: RMB is L3 (ready toggle), LMB is Square, and a
   semi-automatic gun fires on Square release. The component is wrapped
   through its registration pointer. While RMB is held with a standard gun,
   the wrapper edits the raw pad before the filter reads it and restores it
   after: an LMB press fires at once, RMB held stays ready, RMB release
   lowers without firing, and the gun re-readies after a shot. Automatic
   guns, other weapons, and LMB without RMB keep the native behavior.

   Phase 3 step 3 (0.5.1) adds a log-only state tracer. The player actor
   holds a component tree at +0x58; each node keeps its own current
   sub-state function at node+0x58. Once per frame the pad wrapper walks
   that tree and logs components that appear or vanish and sub-states that
   change. F9 writes a numbered mark. Nothing in the game is written.

   Phase 3 step 4a (0.5.2) splits crouch from roll. Native Cross crouches
   from standing, crouches from a slow move, and rolls from a run
   (movement sub-state 0x369030 rolls on a Cross release only when the
   stick magnitude is above 0x96 and status 0xE5 is clear). C now crouches:
   the wrapper holds Cross while C is held and, in the move sub-state, caps
   the magnitude at 0x96 so the move becomes a crouch. The key bound to
   Cross (SPACE by default) now only rolls: in the stand, crouch, and prone
   sub-states it is hidden, and in the move sub-state its press becomes a
   one-frame Cross tap with the magnitude raised above 0x96. Presses that
   start in any other sub-state keep the native Cross.

   Phase 3 step 4b (0.5.3). With RMB not held and a weapon that can CQC, a
   fire-key press becomes Circle held for as long as the key is held: the
   game grabs when its strike finds a contact and punches otherwise. With a
   knife, RMB hides L3 so RMB plus LMB slashes. With an automatic gun and
   RMB held, LMB release keeps the gun up on a 0x77 Square hold instead of
   passing the release that lowers it.

   0.5.4 to 0.5.6, from play. The CQC rule runs before the fire rules, so a
   fire key held for CQC never blocks the RMB lower. While holding an enemy
   with RMB held, Circle stays held so shots do not let the enemy go. A roll
   key still held once the roll starts holds Cross, so the roll ends in
   prone. The roll key is also hidden while crawling and in the roll into
   prone.

   0.5.7 (step 5a). While RMB is held with a standard gun up, the wrapper
   also holds L1 (lock-on), so Snake keeps his facing and strafes. Weapon
   ids 15 to 17 and first person (R1 held) are left out.

   0.5.8. The lock stays on through the re-ready after a shot, and the log
   gets context lines (weapon, gun up or down, first person, holding, real
   L1, stance, key or pad source) so play needs no F9 marks.

   0.5.9. While the lock is on, Snake's facing follows the camera yaw, so
   the mouse (or right stick) turns the aim, and the game's auto-target is
   held off (F8 brings it back). A crouch strafes under the lock too. On a
   controller, LT aims and RT fires (CQC without LT), LB and RB open the
   item and weapon windows, the right stick click toggles first person, and
   D-pad up and down take the camera view and native lock-on.

   0.6.0. Elevation assist: when a target lies on the mouse aim line, the
   shot takes its height (the yaw stays manual). A crosshair is drawn where
   the aim point projects. The aim assist default follows the save
   difficulty (on for Very Easy and Easy). A motion probe logs the movement
   motion for a sideways crouch animation.

   0.6.1. Manual pitch replaces the elevation assist: while locked, the
   camera's vertical input (mouse Y, right stick Y) pitches the aim instead
   of the camera height and zoom, and the aim point takes that pitch.

   0.6.2 and 0.6.3. An over-the-shoulder camera while locked (F7 off and
   on, middle mouse swaps the shoulder).

   0.6.4. One aim rig drives the shoulder camera and the shots: a pivot
   above Snake, the camera yaw, and the pitch give the camera's eye and
   centre ray. Once a frame the game's line check casts that ray; the gun
   pose aims at the hit, third-person shots take the first-person
   (targeted) bullet path toward it, and the crosshair marks it.

   0.6.5. Converted shots start on the camera ray, so they follow the
   crosshair even when the line check misses.

   0.6.6. The shoulder camera and the shoulder swap ease over 150 ms, and
   a left stick click while LT aims swaps the shoulder on the pad (0.6.9:
   the right stick click; the left one toggles first person).

   0.7.0. Pad A from a run: a tap crouches, a hold rolls. A trigger held
   through a closing window takes its gameplay role once gameplay is back,
   so LT right after a weapon switch aims instead of opening the item
   window.

   0.8.0. The shoulder aim makes a held-up soldier shake and drop an item
   as first person does. First person shows a centre crosshair in the
   weapon-at-right view of the guns with an iron sight. A wall hug and a
   corner peek get a movable aim and the crosshair under the native wall
   camera. The fire key while crouched stands Snake up, then CQC. On the
   pad, the D-pad takes the item and weapon windows (left, right) and the
   codec (down) as in Delta New Style; LB and RB are native again. A log
   probe names the wall-press motions.

   0.8.1, from play. RB toggles first person like the left stick click, and
   in first person the aim also brings up the iron sight. The right stick
   browses an item or weapon window held open with the D-pad. A held-up
   soldier gets the aim line from Snake's eye, as in first person.

   0.8.2, from play. In first person, releasing the aim keeps the gun up at
   the right with its crosshair (ids 7 and 9 to 14); guns without an iron
   sight (10, 13, 14) get the crosshair too. The window browsing reads the
   right stick through XInput. While D-pad up holds the QCamo menu (now
   without a pause), A, the right stick and the other D-pad directions go
   to the menu only.

   0.8.3, from play. The D-pad-up menu takes over the whole pad (Snake
   stands still, the world runs). A log probe for the window stick.

   0.8.4. The window stick also sends D-pad pressure and the left-stick
   direction. */

#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <process.h>
#include <stdarg.h>
#include <math.h>
#include <intrin.h>
#define COBJMACROS
#include <d3d11_1.h>

#pragma comment(lib, "user32.lib")
#pragma comment(lib, "dxguid.lib")

/* cmp dword ptr [rip+disp32], 0 is 7 bytes. The displacement is the flag. */
static const char kFlag[] = "83 3D ?? ?? ?? ?? 00 75 ?? B9 BB 00 00 00 E8 ?? ?? ?? ?? 85 C0 0F 84";
static const size_t kFlagDisp = 2;
static const size_t kFlagInsn = 7;

/* Walk: test dword ptr [rbx+7E8h], 8000h. The immediate becomes 0. */
static const char kSquare[] = "F7 83 E8 07 00 00 00 80 00 00 0F 85 ?? ?? ?? ?? BA 15 01 00 00";
static const size_t kSquareAt = 6;
static const uint8_t kSquareOld[] = {0x00, 0x80, 0x00, 0x00};
static const uint8_t kSquareNew[] = {0x00, 0x00, 0x00, 0x00};

/* Walk: ANY(0x9B, 0x115, 0x114, 0x99, 0x2E, 0x27). 0x2E becomes 0x99,
   which is already in the list. */
static const char kList[] = "C7 44 24 30 FF FF FF FF C7 44 24 28 27 00 00 00 C7 44 24 20 2E 00 00 00";
static const size_t kListAt = 20;
static const uint8_t kListOld[] = {0x2E};
static const uint8_t kListNew[] = {0x99};

/* Look: if TEST(0x2E) the left stick turns the view. jne becomes NOPs so
   the flag-set path always reads the right stick. */
static const char kLook[] = "41 8D 4E 2E E8 ?? ?? ?? ?? 85 C0 75 1A 48 8B 86 F8 07 00 00 0F B6 78 04";
static const size_t kLookAt = 11;
static const uint8_t kLookOld[] = {0x75, 0x1A};
static const uint8_t kLookNew[] = {0x90, 0x90};

/* The first-person message 0x20 path: call 0x38B350, then call walk. */
static const char kWalkCall[] =
    "48 8B D7 48 8B CB E8 ?? ?? ?? ?? 48 8B D7 48 8B CB E8 ?? ?? ?? ?? 48 8D 93 70 01 00 00";
static const size_t kWalkCallAt = 0x11;
/* Walk prologue, and its CLR(0xC1, 0xC2) call at +0x2B. */
static const uint8_t kWalkHead[] = {0x48, 0x89, 0x5C, 0x24, 0x18, 0x48, 0x89, 0x7C, 0x24, 0x20};
static const uint8_t kWalkBody[] = {0xBA, 0xC2, 0x00, 0x00, 0x00, 0x41, 0xB8, 0xFF, 0xFF, 0xFF, 0xFF, 0x8D, 0x4A, 0xFF, 0xE8};
static const size_t kWalkBodyAt = 0x2B;

/* Status TEST helper: lea rdx, [bitset] at +0x12, displacement at +0x15. */
static const char kBits[] = "8B D1 B8 01 00 00 00 83 E1 1F D3 E0 8B CA 48 C1 E9 05 48 8D 15 ?? ?? ?? ?? 23 04 8A C3";
static const size_t kBitsDisp = 0x15;
static const size_t kBitsNext = 0x19;

/* Player pad component: lea r9, [component] in its registration. */
static const char kPadReg[] =
    "4C 8D 0D ?? ?? ?? ?? C7 44 24 40 05 00 00 00 45 33 C0 C7 44 24 38 10 00 00 00 BA 2B 0F 81 01";
static const size_t kPadRegDisp = 3;
static const size_t kPadRegNext = 7;
/* Component prologue: movups xmm0, [raw pad]; mov eax, [raw pad+0x20];
   xor ebx, ebx; cmp [filter on], ebx. */
static const char kPadFn[] =
    "48 89 54 24 10 48 89 4C 24 08 53 55 56 57 41 55 41 56 41 57 48 83 EC 40 0F 10 05 ?? ?? ?? ?? 8B 05 ?? ?? ?? ?? 33 DB 39 1D";
/* mov esi, [aimingState]; the Phase 0 pattern. Sub-state is the next dword. */
static const char kAim[] = "8B 35 ?? ?? ?? ?? 48 8D 0D ?? ?? ?? ?? 49 89 8D";

/* Player update, message 0x10: mov rdx, [rbx+58h] (tree head); mov r8d, 10h;
   mov rcx, rbx; call walker. */
static const char kTreeWalk[] = "48 8B 53 58 41 B8 10 00 00 00 48 8B CB E8";
static const size_t kTreeWalkCall = 0x0D;
/* Walker loop: flags +0x20, next +0, skip bits 7, child +0x10, handler +0x40.
   It sits 0x30 bytes into the walker. */
static const char kWalker[] = "8B 53 20 48 8B 33 F6 C2 07 0F 85 ?? ?? ?? ?? 48 8B 4B 10 4C 8B 4B 40";
static const size_t kWalkerAt = 0x30;
/* Registration's lookup by hash: mov eax, [rbx+24h]; and eax, 0FFFFFFh;
   cmp eax, edi; je; mov rcx, [rbx+10h]; test rcx, rcx; je. */
static const char kHashLookup[] = "8B 43 24 25 FF FF FF 00 3B C7 74 23 48 8B 4B 10 48 85 C9 74 10";

/* Movement component init: TEST(3) picks prone, TEST(2) crouch, else stand.
   Three lea rax, [state]; mov [rbx+58h], rax. */
static const char kMoveInit[] =
    "B9 03 00 00 00 E8 ?? ?? ?? ?? 85 C0 74 1D 48 8D 05 ?? ?? ?? ?? 48 89 43 58 33 C0 48 8B 5C 24 30 48 8B 74 24 38 "
    "48 83 C4 20 5F C3 B9 02 00 00 00 E8 ?? ?? ?? ?? 85 C0 74 1D 48 8D 05 ?? ?? ?? ?? 48 89 43 58 33 C0 48 8B 5C 24 "
    "30 48 8B 74 24 38 48 83 C4 20 5F C3 48 8D 05 ?? ?? ?? ?? 48 89 43 58";
static const size_t kInitProne = 0x11;
static const size_t kInitSquat = 0x3C;
static const size_t kInitStill = 0x59;
/* Stand sub-state turning into the move sub-state: cmp eax, 31C7h ... lea. */
static const char kMoveSite[] =
    "E8 ?? ?? ?? ?? 3D C7 31 00 00 7D 41 45 33 C0 66 89 6B 6C 48 8B D3 48 8B CF E8 ?? ?? ?? ?? 48 8D 05 ?? ?? ?? ?? 48 89 43 58 EB";
static const size_t kMoveSiteDisp = 0x21;
/* Move sub-state roll gate: TEST(0xE5) clear, magnitude > 0x96, tap or hold. */
static const char kRollGate[] =
    "B9 E0 00 00 00 E8 ?? ?? ?? ?? 85 C0 74 04 83 63 68 DF B9 E5 00 00 00 E8 ?? ?? ?? ?? 85 C0 75 56 B8 96 00 00 00 "
    "66 39 87 E0 07 00 00 7E 48 85 F6 74 44";
/* The move sub-state loads the roll sub-state this far past the gate. */
static const size_t kRollLea = 0x6A;
/* Crouch to prone: lea rax, [state]; mov [rdi+58h], rax. */
static const char kS2P[] = "E8 ?? ?? ?? ?? 48 8D 05 ?? ?? ?? ?? 48 89 47 58 83 67 68 EF 66 83 BE E0 07 00 00 28";
static const size_t kS2PDisp = 8;
/* Prone: ALL(...); lea rbx, [prone to crouch]; ...; lea rbx, [prone]. A
   twin function has the same tail after a different prefix. */
static const char kP2S[] = "B8 FF FF FF FF 8D 4A FF E8 ?? ?? ?? ?? 85 C0 74 55 48 8B D7 48 8B CE E8 ?? ?? ?? ?? 85 C0 74 09 48 8D 1D ?? ?? ?? ?? EB 15 0F B7 86 D4 06 00 00 48 8D 1D ?? ?? ?? ?? 66 89 86 6A 01 00 00";
static const size_t kP2SDisp = 0x23;
static const size_t kP2SProne = 0x33;
/* Prone crawl: two twin functions store each other's sub-state (0x366FC0,
   seen in play, and 0x366C40); only their jmp back differs. */
static const char kCrawl[] = "75 27 48 8B CE E9 2F FF FF FF 33 C0 45 33 C0 48 8B D7 66 89 47 6C 48 8B CE E8 ?? ?? ?? ?? 48 8D 05 ?? ?? ?? ?? 48 89 47 58";
static const size_t kCrawlDisp = 0x21;
static const char kCrawl2[] = "75 27 48 8B CE E9 2C FF FF FF 33 C0 45 33 C0 48 8B D7 66 89 47 6C 48 8B CE E8 ?? ?? ?? ?? 48 8D 05 ?? ?? ?? ?? 48 89 47 58";
/* Roll end with Cross held: lea rax, [roll to prone]; jmp. */
static const char kRollProne[] =
    "E8 ?? ?? ?? ?? 66 44 89 77 6C 66 44 39 77 60 74 12 45 33 C9 48 8B D7 48 8B CB 45 8D 41 22 E8 ?? ?? ?? ?? 66 44 39 "
    "77 62 74 12 45 33 C9 48 8B D7 48 8B CB 45 8D 41 23 E8 ?? ?? ?? ?? 48 8D 05 ?? ?? ?? ?? EB 3E";
static const size_t kRollProneDisp = 0x3F;
/* Key id lookup: ids 0xF0 and 0xF1 read the Enter and Backspace slots of the
   key array (4 bytes per VK). At +0x5E it calls the layout getter, then the
   binding getter. */
static const char kKeyLookup[] =
    "40 53 48 83 EC 20 8B D9 81 F9 F0 00 00 00 75 19 8B 05 ?? ?? ?? ?? 41 B8 00 80 00 00 F7 D8 1B C0 41 23 C0 48 83 C4 "
    "20 5B C3 81 FB F1 00 00 00 75 19 8B 05 ?? ?? ?? ??";
static const size_t kKeyEnter = 0x12;
static const size_t kKeyBack = 0x33;
static const size_t kKeyCalls = 0x5E;
/* Binding getter (layout, key id, slot): layout 0 and 1 are the default
   tables, anything else the custom table. Table operands are image RVAs. */
static const char kBindFn[] =
    "44 8B C9 83 FA 1A 7C 03 33 C0 C3 49 63 C0 48 63 CA 48 8D 0C 88 48 8D 04 8D 00 00 00 00 48 8D 0D ?? ?? ?? ?? 45 85 "
    "C9 74 16 41 83 E9 01 74 08 8B 84 08 ?? ?? ?? ?? C3 8B 84 08 ?? ?? ?? ?? C3 8B 84 08 ?? ?? ?? ?? C3";
static const size_t kBindBase = 0x20;
static const size_t kBindCustom = 0x32;
static const size_t kBindT1 = 0x3A;
static const size_t kBindT0 = 0x42;

/* Step 5b. Third-person camera object: mov rbx, [camera]; test rbx, rbx;
   je; or dword ptr [rbx+320h], 1. Camera init stores the same global. */
static const char kCamObj[] = "48 8B 1D ?? ?? ?? ?? 48 85 DB 74 3C 83 8B 20 03 00 00 01 33 D2 B1 54";
static const size_t kCamObjDisp = 3;
static const size_t kCamObjNext = 7;
/* Yaw integrator tail: add word ptr [rbx+338h], ax (the orbit yaw). */
static const char kCamYaw[] = "F3 0F 2C C0 66 01 83 38 03 00 00 48 83 C4 20 5B C3";
/* Target refresh: mov rax, [player]; mov [rcx+0E8h], rax; ret. */
static const char kCamTarget[] = "48 8B 05 ?? ?? ?? ?? 48 89 81 E8 00 00 00 C3";
/* Mode check: cmp [rdi+318h], esi. */
static const char kCamMode[] = "39 B7 18 03 00 00 0F 85 ?? ?? ?? ?? 48 0F BF 97 F4 00 00 00";
/* Sub-mode by script param T (+0x74): T=3 stores raw held L1 in +0x31C. */
static const char kCamSub[] =
    "8B 49 74 85 C9 74 6A 83 E9 01 74 3D 83 E9 01 74 20 83 F9 01 0F 85 ?? ?? ?? ?? 8B 05 ?? ?? ?? ?? C1 E8 0A 23 C1 "
    "89 83 1C 03 00 00";
/* Weapon aim helper, L1 press path: +0x800 = stick direction (+0x7E2) when
   the magnitude (+0x7E0) is set, else the turn target (+0x16A). */
static const char kFaceWpn[] =
    "66 44 39 A7 E0 07 00 00 74 09 0F B7 87 E2 07 00 00 EB 07 0F B7 87 6A 01 00 00 66 89 87 00 08 00 00 BA FF FF FF FF "
    "8D 4A 32 E8 ?? ?? ?? ?? 66 44 89 A3 AA 00 00 00";
/* Same helper: TEST(0xE3) drops the target and skips every +0x800 write. */
static const char kNoTarget[] = "B9 E3 00 00 00 48 89 AC 24 B0 00 00 00 E8 ?? ?? ?? ?? 85 C0 74 33 4C 89 A3 88 00 00 00";
static const size_t kNoTargetCall = 0x0D;
/* Movement strafe gate: TEST(1); je; cmp [rdi+7E0h], bx; jle; SET(0x11). */
static const char kStrafeStand[] =
    "B9 01 00 00 00 E8 ?? ?? ?? ?? 85 C0 0F 84 ?? ?? ?? ?? 66 39 9F E0 07 00 00 0F 8E ?? ?? ?? ?? 8B D5 B9 11 00 00 00";
static const size_t kStrafeStandCall = 5;
/* Controller read for one port (Steam Input), and its callee. */
static const char kCtrlCall[] = "E8 ?? ?? ?? ?? 85 C0 41 8B FF 40 0F 95 C7 85 F6 75 07 8B CF E8";
static const char kCtrlFn[] = "48 83 EC 28 80 3D ?? ?? ?? ?? 00 74 1F 83 3D ?? ?? ?? ?? 00 7E 16 85 C9 75 12 48 8B CA E8";
/* In-game button config: cmp [flag], 0; lea r9, [image]; ... mov r8d,
   [r9+rax*4+table]; cmovne (flag set means identity). */
static const char kPadCfg[] =
    "83 3D ?? ?? ?? ?? 00 4C 8D 0D ?? ?? ?? ?? 48 63 C1 45 8B 84 81 ?? ?? ?? ?? 44 0F 45 C1 41 83 F8 09 77 62";
static const size_t kPadCfgFlagDisp = 2;
static const size_t kPadCfgFlagNext = 7;
static const size_t kPadCfgBaseDisp = 10;
static const size_t kPadCfgBaseNext = 14;
static const size_t kPadCfgTable = 0x15;

/* 0.6.0. Aim point builder (0x378C90), found through the int3 before it:
   push rsi; push rdi; push r12 (5 bytes, moved to the trampoline); ... mov
   rcx, [rdx+88h] (the weapon node's target record). */
static const char kBuild[] =
    "CC 40 56 57 41 54 41 56 48 83 EC 78 0F 29 7C 24 50 48 8B 05 ?? ?? ?? ?? 48 33 C4 48 89 44 24 40 48 8B F9 45 33 E4 "
    "48 8B 8A 88 00 00 00";
static const uint8_t kBuildHead[] = {0x40, 0x56, 0x57, 0x41, 0x54};
/* The gun state's fire frame: builder call, then TEST(0x133). The builder is
   called from this state function only between these bounds around it. */
static const char kFireCall[] = "48 8B D3 48 8B CF E8 ?? ?? ?? ?? B9 33 01 00 00 E8";
static const size_t kFireCallAt = 6;
#define GUN_SITES_BEFORE 0x400
#define GUN_SITES_AFTER 0x1700
/* Camera height/zoom from its vertical input (0x212AA0): push rbx; sub
   rsp, 30h (6 bytes, moved to the trampoline); cmp [rcx+64h], 0 ... */
static const char kZoom[] = "40 53 48 83 EC 30 83 79 64 00 48 8B D9 74 61 F3 0F 10 05";
static const uint8_t kZoomHead[] = {0x40, 0x53, 0x48, 0x83, 0xEC, 0x30};
/* Mouse delta (dx, dy int32) read by the key packer: mov edi, 0FFh; mov rbx,
   [delta]. */
static const char kMouseDelta[] = "BF FF 00 00 00 48 8B 1D ?? ?? ?? ?? 4C 8B E3";
static const size_t kMouseDeltaDisp = 8;
static const size_t kMouseDeltaNext = 12;
/* Difficulty bucket: mov rax, [status block]; movzx ecx, word ptr [rax+6];
   cmp cx, 0Ah (10 Very Easy ... 60 European Extreme). */
static const char kDifficulty[] = "48 8B 05 ?? ?? ?? ?? 0F B7 48 06 66 83 F9 0A 7F 06 B8 01 00 00 00 C3";
/* Crosshair. Gameplay frames call the renderer's present wrapper (0x6EF50)
   with rcx = [renderer]. */
static const char kPresentCall[] = "48 8B 0D ?? ?? ?? ?? E8 ?? ?? ?? ?? 83 3D ?? ?? ?? ?? 00 74 10 48 8B 0D ?? ?? ?? ?? E8";
/* The wrapper: two calls, then Present(1, 0) on [rdi+2B0h] (the swap
   chain). The crosshair is drawn after the second call. */
static const char kPresentFn[] =
    "48 89 6C 24 20 57 48 83 EC 20 48 8B F9 E8 ?? ?? ?? ?? 48 8B CF E8 ?? ?? ?? ?? 33 ED 40 38 AF A8 06 00 00 0F 85 "
    "?? ?? ?? ?? 48 8D 4C 24 30 48 89 74 24 40 FF 15 ?? ?? ?? ?? 48 8B 8F B0 02 00 00 8D 55 01 45 33 C0 48 8B 01 FF 50 40";
static const size_t kPresentFnCall = 0x15;
/* D3D11CreateDevice: device at +0x298, immediate context at +0x2A0; then
   QueryInterface of ID3D11DeviceContext1 into +0x2A8. */
static const char kDevice[] = "C7 85 30 05 00 00 00 B1 00 00 49 8D 85 98 02 00 00 45 33 C9 49 8D B5 A0 02 00 00";
static const char kContext1[] = "4D 8D 85 A8 02 00 00 48 8D 15 ?? ?? ?? ?? 4C 8B 09 41 FF 11";
/* World to PS2 screen: lea rdx, [channel 0 + 0C0h] (the matrix), and the
   viewport ints W, H, X0, Y0 of the same channel. */
static const char kProjMatrix[] = "48 8D 15 ?? ?? ?? ?? 44 0F 29 88 78 FF FF FF 4C 8B C9";
static const char kProjView[] =
    "F3 0F 2A 1D ?? ?? ?? ?? 49 8B 81 C0 00 00 00 0F 57 C0 0F 57 E4 F3 0F 2A 05 ?? ?? ?? ?? F3 0F 2A 25 ?? ?? ?? ?? "
    "F3 0F 5E EE F3 0F 5E FE F3 0F 59 EA 0F 57 C9 F3 0F 2A 0D ?? ?? ?? ??";
/* GV pause level getter. */
static const char kPauseLevel[] =
    "8B 05 ?? ?? ?? ?? C3 CC CC CC CC CC CC CC CC CC F7 D1 21 0D ?? ?? ?? ?? C3 CC CC CC CC CC CC CC 09 0D ?? ?? ?? ?? C3";
/* 0.6.4. The gun state's shot ray (0x3749CA): line check 0x10B4B0(0x1F,
   actor id, 0x42, from, to, 0.0); on a hit, 0x105FD0(0x105E50(), out)
   writes the hit point. */
static const char kShotRay[] =
    "4C 8B 8F 30 07 00 00 4C 8D B7 20 05 00 00 8B 97 20 01 00 00 41 B8 42 00 00 00 0F 57 C0 49 83 C1 20 F3 0F 11 44 "
    "24 28 4C 89 74 24 20 41 8D 48 DD E8 ?? ?? ?? ?? 85 C0 74 15 E8 ?? ?? ?? ?? 8B C8 48 8D 97 E0 06 00 00 E8";
static const size_t kShotRayLine = 0x30;
static const size_t kShotRayIdx = 0x39;
static const size_t kShotRayPt = 0x47;
static const uint8_t kLineHead[] = {0x48, 0x89, 0x5C, 0x24, 0x08, 0x48, 0x89, 0x74, 0x24, 0x10, 0x57, 0x48, 0x83, 0xEC, 0x20};
/* The player's bullet callbacks call an untargeted spawner outside first
   person and its targeted twin (same arguments with the target point
   inserted second) in first person (0x35B780's first-person call also
   takes another matrix, and 0x35E240 stays untargeted in first person
   while +0x606 is set). Each pattern ends on the call. Pair A:
   0x25F240/0x25F4E0 (callbacks 0x359BE0, 0x35A7F0, 0x35B780, 0x35E240);
   pair B: 0x25F390/0x25F630 (0x35C6E0); pair C: 0x262250/0x262350
   (0x35D190). */
static const char kShotA1[] =
    "44 8B C5 89 74 24 40 8B D0 C7 44 24 38 88 13 00 00 F3 0F 11 44 24 30 C7 44 24 28 32 00 00 00 C7 44 24 20 00 00 "
    "00 00 E8";
static const char kShotA2[] =
    "44 8B C7 89 74 24 40 8B D0 C7 44 24 38 15 34 00 00 F3 0F 11 44 24 30 C7 44 24 28 78 00 00 00 44 89 64 24 20 E8";
static const char kShotA3[] =
    "48 8B CB 44 89 7C 24 48 89 6C 24 40 C7 44 24 38 15 34 00 00 F3 0F 11 44 24 30 89 74 24 28 44 89 7C 24 20 E8";
static const char kShotA4[] =
    "48 8B CF 89 74 24 40 C7 44 24 38 88 13 00 00 F3 0F 11 44 24 30 C7 44 24 28 96 00 00 00 C7 44 24 20 90 01 00 "
    "00 E8";
static const char kShotB[] =
    "48 8B CF 89 44 24 48 89 74 24 40 C7 44 24 38 E8 03 00 00 F3 0F 11 44 24 30 C7 44 24 28 32 00 00 00 89 44 24 20 "
    "E8";
static const char kShotC[] =
    "45 33 C9 8B 93 20 01 00 00 41 B8 81 03 00 00 C7 44 24 30 E8 03 00 00 48 8B CF 89 74 24 28 F3 0F 11 44 24 20 E8";
/* Their targeted calls, which name the twins. */
static const char kShotAT[] =
    "44 8B C0 C7 44 24 40 88 13 00 00 F3 0F 11 44 24 38 C7 44 24 30 32 00 00 00 C7 44 24 28 00 00 00 00 C7 44 24 "
    "20 00 00 00 00 E8";
static const char kShotBT[] =
    "89 44 24 50 89 74 24 48 C7 44 24 40 E8 03 00 00 F3 0F 11 44 24 38 C7 44 24 30 32 00 00 00 89 44 24 28 89 44 24 "
    "20 E8";
static const char kShotCT[] =
    "41 B9 81 03 00 00 C7 44 24 38 E8 03 00 00 48 8B CF 89 74 24 30 F3 0F 11 44 24 28 C7 44 24 20 00 00 00 00 E8";
/* Camera wall step (0x2139B0), found through the int3 before it. Its three
   callers run it right after the eye/look step (0x2132D0, which adds the
   distance extension +0xD4, the offsets +0x360..+0x36C, and the vertical
   ease); it line-checks from the pivot +0x3A0 to the eye +0x380, pulls eye
   and look toward the pivot when blocked, clamps the eye above the floor,
   and writes the final eye +0x3D0 and look +0x3E0. Prologue: mov r11, rsp;
   push rbx; sub rsp, 0E0h (11 bytes, moved to the trampoline). */
static const char kCamWall[] =
    "CC 4C 8B DC 53 48 81 EC E0 00 00 00 48 8B 05 ?? ?? ?? ?? 48 33 C4 48 89 44 24 40 8B 41 60 48 8B D9 49 89 6B 10";
static const uint8_t kCamWallHead[] = {0x4C, 0x8B, 0xDC, 0x53, 0x48, 0x81, 0xEC, 0xE0, 0x00, 0x00, 0x00};
/* Camera yaw step (0x213010): push rbx; sub rsp, 20h (6 bytes, moved to
   the trampoline); ... it turns the orbit yaw +0x338 by the right stick X
   input at +0x308 (written from the stick byte at 0x211F77). */
static const char kCamTurn[] =
    "CC 40 53 48 83 EC 20 F3 0F 10 81 3C 03 00 00 48 8B D9 F3 0F 59 05 ?? ?? ?? ?? F3 0F 11 81 3C 03 00 00 F3 0F 10 "
    "81 80 00 00 00 E8 ?? ?? ?? ?? F3 0F 10 8B 08 03 00 00";
static const uint8_t kCamTurnHead[] = {0x40, 0x53, 0x48, 0x83, 0xEC, 0x20};
/* Camera stance height offset: addss xmm8, [rbx+358h]; movss [rbx+358h],
   xmm8 (0 standing, -200 crouched, -400 prone, -500 while holding an enemy
   and a few other statuses; eased). */
static const char kCamStance[] = "F3 44 0F 58 83 58 03 00 00 F3 44 0F 11 83 58 03 00 00";
/* 0.8.0. Hold-up drop. A held-up soldier shakes and drops an item only
   while status 0xBA (the first-person view) is set: TEST(0xBA) in hold-up
   mode 0xE's sub-state 0x91 (0x2010AA) and mode 0xF's shake (0x202D19),
   and TEST(7) in mode 0xE variant B (0x2015EE, a reaction without a
   drop). Each pattern starts at the mov ecx, id; the call follows. */
static const char kHoldE[] = "B9 BA 00 00 00 33 FF E8 ?? ?? ?? ?? 85 C0 0F 84 ?? ?? ?? ?? 8D 57 01";
static const char kHoldF[] = "B9 BA 00 00 00 33 FF E8 ?? ?? ?? ?? 85 C0 0F 84 ?? ?? ?? ?? 48 8B CB";
static const size_t kHoldCall = 7;
static const char kHoldB[] = "B9 07 00 00 00 E8 ?? ?? ?? ?? 85 C0 74 69 48 8B 8B B8 1C 00 00";
static const size_t kHoldBCall = 5;
/* The player's aim line (a target record of kind 0xAD at actor+0xA20; the
   soldier's damage receiver stores the body part it touches, and mode
   0xE's shake test may want part 1 or 3). Its update 0x382B80 takes the muzzle matrix from
   0x3649F0 into obj+0x90 (row 2 forward, row 3 position) and casts 30000
   along row 2 outside first person. The call is at +12; 0x3649F0 reads the
   player from a global (mov rcx, [player] at +0x12). */
static const char kAimLine[] = "48 81 C1 90 00 00 00 0F 29 74 24 60 E8 ?? ?? ?? ?? F6 43 50 08";
static const size_t kAimLineCall = 12;
static const char kMuzzleFn[] =
    "48 89 5C 24 08 57 48 83 EC 20 48 8B F9 33 DB 48 8B 0D ?? ?? ?? ?? 48 85 C9 74 6B 48 8B 81 F8 06 00 00 BA E3 0E D6 00";
static const size_t kMuzzlePlayer = 0x12;
/* Wall-press component (hash 0xD9728F): registration lea r9, [handler];
   the handler clears statuses 1, 2, 0xE, 0x3C..0x3F at message 0x80 and
   calls its state function at node+0x100. */
static const char kWallReg[] = "4C 8D 0D ?? ?? ?? ?? C7 44 24 40 0D 00 00 00 BA 8F 72 D9 02";
static const char kWallState[] =
    "BA 01 00 00 00 C7 44 24 38 FF FF FF FF C7 44 24 30 3D 00 00 00 C7 44 24 28 3E 00 00 00 C7 44 24 20 3F 00 00 00 "
    "8D 4A 01 44 8D 4A 3B 44 8D 42 0D E8 ?? ?? ?? ?? 48 8B D3 48 8B CF FF 93 00 01 00 00";
/* Its camera record at node+0x120 (init), and its wall yaw word +0x232. */
static const char kWallCam[] = "F3 0F 11 44 24 20 48 89 8F 20 01 00 00 E8";
static const char kWallYaw[] = "0F B7 87 30 02 00 00 66 89 87 32 02 00 00";
/* Motion requests (log only): the motion component (hash 0x1B9F95) passes
   node+0x58 to the decoder 0x365E30 for layers 0 and 1; each layer entry
   is 0x28 bytes: +0x0C archive, +0x10 motion number, +0x14 sender hash. */
static const char kMotionReq[] = "48 83 C2 58 89 44 24 20 48 8B CE E8 ?? ?? ?? ?? 8B 86 E8 00 00 00 48 8D 53 58";
static const char kMotionDec[] = "48 8D 3C 92 89 44 FB 08 0F B7 C5 66 23 C1 66 89 44 FB 10";
/* 0.8.1. The movement handler's one motion send (message 0x18 with
   &node+0x60) to 0x3665C0, whose head forwards to the parent's +0x48. */
static const char kMotionSend[] = "4C 8D 4E 60 41 B8 18 00 00 00 48 8B D6 48 8B CF E8";
static const size_t kMotionSendCall = 0x10;
/* 0.8.1. Prone in grass keeps the third-person camera. The first-person
   component enters on its own while status 0xBB is latched (0x38B464,
   0x38BB20); 0x37E5E0 sets 0xBB from ANY(0x4F, 0x50) while prone, where
   0x50 is the "grassintrude" map volume (hash 0x11C95D, SET at 0x37E68C)
   and 0x4F the "intrude" one (under a vehicle, bed or vent; 0x7AFC08, SET
   at 0x37E6AA). The ANY becomes ANY(0x4F): mov edx, 50h / lea r8d,
   [rdx-51h] / lea ecx, [rdx-1] turns into mov edx, -1 / (the same lea) /
   lea ecx, [rdx+50h]. Status 0x50 itself is still set. */
static const char kGrassFp[] = "BA 50 00 00 00 44 8D 42 AF 8D 4A FF E8 ?? ?? ?? ?? 85 C0 74";
static const uint8_t kGrassOld[] = {0xBA, 0x50, 0x00, 0x00, 0x00, 0x44, 0x8D, 0x42, 0xAF, 0x8D, 0x4A, 0xFF};
static const uint8_t kGrassNew[] = {0xBA, 0xFF, 0xFF, 0xFF, 0xFF, 0x44, 0x8D, 0x42, 0xAF, 0x8D, 0x4A, 0x50};
static const uint8_t kSendHead[] = {0x4C, 0x8B, 0x52, 0x18, 0x49, 0x83, 0x7A, 0x48, 0x00};

#define MAX_PAT 128

static size_t parse_pattern(const char *text, uint8_t *bytes, char *mask) {
    size_t n = 0;
    while (*text != '\0' && n < MAX_PAT) {
        if (*text == ' ') {
            text++;
            continue;
        }
        if (text[0] == '?') {
            bytes[n] = 0;
            mask[n] = '?';
        } else {
            char hex[3] = {text[0], text[1], '\0'};
            bytes[n] = (uint8_t)strtoul(hex, NULL, 16);
            mask[n] = 'x';
        }
        n++;
        text += 2;
    }
    mask[n] = '\0';
    return n;
}

static int match_pattern(const uint8_t *buf, size_t len, const char *text, const uint8_t **first) {
    uint8_t pat[MAX_PAT];
    char mask[MAX_PAT + 1];
    size_t n = parse_pattern(text, pat, mask);
    int count = 0;
    *first = NULL;
    if (len < n) {
        return 0;
    }
    for (size_t i = 0; i + n <= len; i++) {
        int ok = 1;
        for (size_t j = 0; j < n; j++) {
            if (mask[j] == 'x' && buf[i + j] != pat[j]) {
                ok = 0;
                break;
            }
        }
        if (!ok) {
            continue;
        }
        count++;
        if (count == 1) {
            *first = buf + i;
        }
        if (count > 1) {
            break;
        }
    }
    return count;
}

static uint8_t *rip_target(const uint8_t *at, size_t disp_at, size_t next) {
    int32_t disp;
    memcpy(&disp, at + disp_at, sizeof disp);
    return (uint8_t *)(at + next + disp);
}

#ifndef FPVMOVE_TEST

static int write_one(int32_t *flag) {
    MEMORY_BASIC_INFORMATION info;
    DWORD old = 0;
    int changed = 0;
    if (VirtualQuery(flag, &info, sizeof info) != sizeof info || info.State != MEM_COMMIT) {
        return 0;
    }
    if ((info.Protect & (PAGE_READWRITE | PAGE_WRITECOPY | PAGE_EXECUTE_READWRITE | PAGE_EXECUTE_WRITECOPY)) == 0) {
        if (!VirtualProtect(flag, sizeof *flag, PAGE_READWRITE, &old)) {
            return 0;
        }
        changed = 1;
    }
    *flag = 1;
    if (changed) {
        DWORD ignored = 0;
        VirtualProtect(flag, sizeof *flag, old, &ignored);
    }
    return 1;
}

static HMODULE g_self;
static volatile int32_t *g_flag;
static const uint32_t *g_bits;
static int g_square_off;
static int g_list_off;

typedef void(__fastcall *walk_fn)(void *, void *, void *, void *);
static walk_fn g_walk;

enum { R_PASS, R_S7, R_FLAG, R_FIRST, R_SQUARE, R_LIST, R_STICK, R_COUNT };
static const char *const kReason[R_COUNT] = {"pass", "s7", "flag", "first", "square", "list", "stick"};
static volatile LONG g_calls;
static volatile LONG g_reason[R_COUNT];
static volatile LONG g_sq;
static volatile LONG g_b2e;
static volatile LONG g_b27;
static volatile LONG g_moved;

static void append_log(const char *text) {
    wchar_t path[MAX_PATH];
    const wchar_t name[] = L"fpvmove.log";
    DWORD n = GetModuleFileNameW(g_self, path, MAX_PATH);
    DWORD i;
    size_t name_len = (sizeof name / sizeof name[0]) - 1;
    HANDLE file;
    DWORD wrote = 0;
    if (n == 0 || n >= MAX_PATH) {
        return;
    }
    i = n;
    while (i > 0 && path[i - 1] != L'\\' && path[i - 1] != L'/') {
        i--;
    }
    if (i + name_len + 1 >= MAX_PATH) {
        return;
    }
    memcpy(path + i, name, sizeof name);
    file = CreateFileW(path, FILE_APPEND_DATA, FILE_SHARE_READ | FILE_SHARE_WRITE, NULL, OPEN_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
    if (file == INVALID_HANDLE_VALUE) {
        return;
    }
    WriteFile(file, text, (DWORD)strlen(text), &wrote, NULL);
    CloseHandle(file);
}

static void truncate_log(void) {
    wchar_t path[MAX_PATH];
    const wchar_t name[] = L"fpvmove.log";
    DWORD n = GetModuleFileNameW(g_self, path, MAX_PATH);
    DWORD i;
    size_t name_len = (sizeof name / sizeof name[0]) - 1;
    HANDLE file;
    if (n == 0 || n >= MAX_PATH) {
        return;
    }
    i = n;
    while (i > 0 && path[i - 1] != L'\\' && path[i - 1] != L'/') {
        i--;
    }
    if (i + name_len + 1 >= MAX_PATH) {
        return;
    }
    memcpy(path + i, name, sizeof name);
    file = CreateFileW(path, GENERIC_WRITE, FILE_SHARE_READ, NULL, CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
    if (file != INVALID_HANDLE_VALUE) {
        CloseHandle(file);
    }
}

static void log_line(const char *fmt, ...) {
    char line[512];
    va_list args;
    va_start(args, fmt);
    vsnprintf(line, sizeof line - 3, fmt, args);
    va_end(args);
    strcat(line, "\r\n");
    append_log(line);
}

static int scan_exec(uint8_t *base, IMAGE_NT_HEADERS64 *nt, const char *text, const uint8_t **hit) {
    IMAGE_SECTION_HEADER *section = IMAGE_FIRST_SECTION(nt);
    int count = 0;
    unsigned i;
    *hit = NULL;
    for (i = 0; i < nt->FileHeader.NumberOfSections; i++) {
        const uint8_t *found = NULL;
        int n;
        if ((section[i].Characteristics & IMAGE_SCN_MEM_EXECUTE) == 0) {
            continue;
        }
        n = match_pattern(base + section[i].VirtualAddress, section[i].Misc.VirtualSize, text, &found);
        if (n > 0 && count == 0) {
            *hit = found;
        }
        count += n;
        if (count > 1) {
            break;
        }
    }
    return count;
}

static int write_code(uint8_t *at, const uint8_t *bytes, size_t len) {
    DWORD old = 0;
    if (!VirtualProtect(at, len, PAGE_EXECUTE_READWRITE, &old)) {
        return 0;
    }
    memcpy(at, bytes, len);
    VirtualProtect(at, len, old, &old);
    FlushInstructionCache(GetCurrentProcess(), at, len);
    return 1;
}

static int patch_site(uint8_t *base, IMAGE_NT_HEADERS64 *nt, const char *label, const char *text, size_t at,
                      const uint8_t *old_bytes, const uint8_t *new_bytes, size_t len) {
    const uint8_t *hit = NULL;
    int count = scan_exec(base, nt, text, &hit);
    uint8_t *site;
    if (count != 1 || hit == NULL) {
        log_line("fpvmove: %s matches=%d; unchanged", label, count);
        return 0;
    }
    site = (uint8_t *)hit + at;
    if (memcmp(site, old_bytes, len) != 0) {
        log_line("fpvmove: %s bytes differ at rva 0x%X; unchanged", label, (unsigned)(site - base));
        return 0;
    }
    if (!write_code(site, new_bytes, len)) {
        log_line("fpvmove: %s not writable; unchanged", label);
        return 0;
    }
    log_line("fpvmove: %s patched at rva 0x%X", label, (unsigned)(site - base));
    return 1;
}

static int status(unsigned id) {
    return (g_bits[id >> 5] >> (id & 31u)) & 1u;
}

/* Mirrors walk's early-outs in order, using the gates as patched. */
static int walk_reason(const uint8_t *actor, const uint8_t *node) {
    uint32_t held = *(const uint32_t *)(actor + 0x7E8);
    if (!status(0x07)) {
        return R_S7;
    }
    if (*g_flag == 0 && !status(0xBB)) {
        return R_FLAG;
    }
    if (node[0xBB] != 0) {
        return R_FIRST;
    }
    if (!g_square_off && (held & 0x8000u) != 0) {
        return R_SQUARE;
    }
    if (status(0x9B) || status(0x115) || status(0x114) || status(0x99) || status(0x27) || (!g_list_off && status(0x2E))) {
        return R_LIST;
    }
    if (*(const uint16_t *)(actor + 0x7E0) == 0 && !status(0xBF) && !status(0xC0)) {
        return R_STICK;
    }
    return R_PASS;
}

static void __fastcall walk_wrap(void *actor, void *node, void *r8, void *r9) {
    const uint8_t *a = (const uint8_t *)actor;
    int reason = walk_reason(a, (const uint8_t *)node);
    InterlockedIncrement(&g_calls);
    InterlockedIncrement(&g_reason[reason]);
    if ((*(const uint32_t *)(a + 0x7E8) & 0x8000u) != 0) {
        InterlockedIncrement(&g_sq);
    }
    if (status(0x2E)) {
        InterlockedIncrement(&g_b2e);
    }
    if (status(0x27)) {
        InterlockedIncrement(&g_b27);
    }
    g_walk(actor, node, r8, r9);
    if (reason == R_PASS && (*(const float *)(a + 0x620) != 0.0f || *(const float *)(a + 0x628) != 0.0f)) {
        InterlockedIncrement(&g_moved);
    }
}

/* Raw pad fields read by the filter. */
#define PAD_HELD 0x00
#define PAD_SQUARE_PRESSURE 0x13
#define PAD_PRESS 0x14
#define PAD_RELEASE 0x18
#define BTN_L3 0x0002u
#define BTN_SQUARE 0x8000u
/* Frames a native Square press is held before the forced release, and
   frames between a shot and the re-ready while RMB is still held. */
#define TAP_FRAMES 3
#define REARM_FRAMES 8
#define AUTO_HOLD_PRESSURE 0x77

typedef uint64_t(__fastcall *pad_fn)(void *, void *, void *, void *);
static pad_fn g_pad_fn;
static uint8_t *g_pad;
static const volatile int32_t *g_filter_on;
static const volatile int32_t *g_aim;
static int g_latch;
static int g_tap;
static int g_wait;
static int g_auto_hold;

enum {
    P_CALLS,
    P_STD,
    P_FIRE,
    P_TAP,
    P_DROP,
    P_LOWER,
    P_REARM,
    P_CROUCH,
    P_CLAMP,
    P_ROLL,
    P_ROLLFIX,
    P_ROLLMISS,
    P_HIDE,
    P_CQC,
    P_KNIFE,
    P_AUTO,
    P_AUTOLOWER,
    P_KEEP,
    P_ROLLPRONE,
    P_LOCK,
    P_UNLOCK,
    P_AIM,
    P_CSTRAFE,
    P_REMAP,
    P_NOTARGET,
    P_PITCH,
    P_XH,
    P_RIG,
    P_SHOT,
    P_EYENOW,
    P_EYEOLD,
    P_WALL,
    P_TURN,
    P_HOLDUP,
    P_AIMLINE,
    P_CQCCROUCH,
    P_STANDMISS,
    P_WALLAIM,
    P_SIDESTEP,
    P_COUNT
};
static const char *const kPadName[P_COUNT] = {"calls", "std",  "fire",     "tap",     "drop", "lower",
                                              "rearm", "crouch", "clamp",  "roll",    "rollfix", "rollmiss",
                                              "hide",  "cqc",    "knife",  "auto",    "autolower",
                                              "keep",  "rollprone", "lock", "unlock", "aim", "cstrafe", "remap", "notarget", "pitch", "xh",
                                              "rig",   "shot",   "eyenow", "eyeold", "wall", "turn",
                                              "holdup", "aimline", "cqccrouch", "standmiss", "wallaim", "sidestep"};
static volatile LONG g_pad_n[P_COUNT];

/* Lock-on (PS2 L1). With it held and a gun raised, the weapon sub-states stop
   taking Snake's facing from the stick (0x37A790) and the movement handler
   strafes (0x369C12). Weapon ids 15 to 17 also read a held L1 as a held aim
   (0x362600), so they are left out. */
#define BTN_L1 0x0400u
#define BTN_R1 0x0800u
static int g_lock;
/* 0.8.1: the aim holds L1 in first person (iron sight); L1 in the output
   last frame. */
static int g_ads;
/* 0.8.2: in first person with a gun shown at the right, releasing the aim
   keeps the gun up (that view and its crosshair); leaving first person
   lowers it. The kept gun counts as aimed for the fire and CQC rules. */
static int g_fp_keep;

static int fp_keep_gun(uint32_t id) {
    return id == 7 || (id >= 9 && id <= 14);
}
static int g_l1_prev;
/* Shared with the controller-read hook (0.5.9). Pad component calls, bumped
   once per gameplay frame; RT held in its remapped (fire) role; crouched
   strafe allowed this frame. */
static volatile LONG g_pad_seq;
static volatile LONG g_ctrl_rt;
static volatile LONG g_cstrafe;

static void trace_line(const char *fmt, ...);

/* The filter's own standard path: weapon present, not ids 0-2 or 0x1B,
   no 0x20000 charge flag, statuses 0x78 and 0x85 clear, filter on. */
static int pad_standard(const uint8_t *actor, uint32_t *flags) {
    const uint8_t *weapon = *(uint8_t *const *)(actor + 0x6F8);
    uint32_t f;
    uint32_t id;
    if (*g_filter_on == 0 || weapon == NULL) {
        return 0;
    }
    f = *(const uint32_t *)(weapon + 0x20);
    id = f & 0xFFu;
    *flags = f;
    return id > 2 && id != 0x1B && (f & 0x20000u) == 0 && !status(0x78) && !status(0x85);
}

static void pad_step(const uint8_t *actor) {
    uint32_t *held = (uint32_t *)(g_pad + PAD_HELD);
    uint32_t *press = (uint32_t *)(g_pad + PAD_PRESS);
    uint32_t *release = (uint32_t *)(g_pad + PAD_RELEASE);
    uint8_t *pressure = g_pad + PAD_SQUARE_PRESSURE;
    uint32_t h = *held;
    uint32_t p = *press;
    uint32_t r = *release;
    uint8_t sq = *pressure;
    int32_t aim = g_aim[0];
    int32_t sub = g_aim[1];
    uint32_t flags = 0;
    int real_rmb = (h & BTN_L3) != 0;
    int rmb;
    int standard;
    int lock;
    int keep_gun;
    /* After a shot this wrapper made, hide LMB until it is released so the
       filter never sees a second press or a release that fires again. */
    if (g_latch) {
        if ((r & BTN_SQUARE) != 0 || (h & BTN_SQUARE) == 0) {
            g_latch = 0;
        }
        h &= ~BTN_SQUARE;
        p &= ~BTN_SQUARE;
        r &= ~BTN_SQUARE;
        sq = 0;
    }
    standard = pad_standard(actor, &flags);
    keep_gun = standard && (h & BTN_R1) != 0 && fp_keep_gun(flags & 0xFFu);
    if (g_fp_keep && (real_rmb || !keep_gun)) {
        /* The aim is held again, or first person ended (or the weapon
           changed): leaving first person with the gun still up lowers it
           as an aim release would. */
        g_fp_keep = 0;
        if (!real_rmb && standard && aim != 0 && (h & BTN_SQUARE) == 0) {
            p |= BTN_L3;
            InterlockedIncrement(&g_pad_n[P_LOWER]);
            trace_line("first person off: gun lowered");
        }
    }
    rmb = real_rmb || g_fp_keep;
    if (!standard) {
        g_tap = 0;
        g_wait = 0;
        g_auto_hold = 0;
    } else {
        InterlockedIncrement(&g_pad_n[P_STD]);
        if (g_tap > 0 && ((r & BTN_SQUARE) != 0 || (h & BTN_SQUARE) == 0)) {
            /* LMB let go first: the native release already fires. */
            g_tap = 0;
        } else if (g_tap > 0) {
            if (--g_tap == 0) {
                h &= ~BTN_SQUARE;
                p &= ~BTN_SQUARE;
                r |= BTN_SQUARE;
                sq = 0;
                g_latch = 1;
                InterlockedIncrement(&g_pad_n[P_FIRE]);
            }
        } else if ((p & BTN_SQUARE) != 0 && rmb && (flags & 0x800u) == 0) {
            if (aim == 1 && sub == 2) {
                /* Gun up on the fake 0x77 hold: a release now fires. */
                h &= ~BTN_SQUARE;
                p &= ~BTN_SQUARE;
                r |= BTN_SQUARE;
                sq = 0;
                g_latch = 1;
                InterlockedIncrement(&g_pad_n[P_FIRE]);
            } else {
                /* Not raised yet: let the native press raise, release later. */
                g_tap = TAP_FRAMES;
                InterlockedIncrement(&g_pad_n[P_TAP]);
            }
        }
        /* Automatic guns fire while Square pressure is at least 0x78, and a
           Square release clears aimingState, which lowers the gun. With RMB
           still held after LMB is let go, hide that release and hold Square
           at 0x77, one below the threshold, so the gun stays up. A new LMB
           press brings the real pressure back and fires. */
        if ((flags & 0x800u) != 0 && rmb && aim != 0 && (h & BTN_SQUARE) == 0 &&
            (g_auto_hold || (r & BTN_SQUARE) != 0)) {
            if (!g_auto_hold) {
                g_auto_hold = 1;
                InterlockedIncrement(&g_pad_n[P_AUTO]);
            }
            h |= BTN_SQUARE;
            p &= ~BTN_SQUARE;
            r &= ~BTN_SQUARE;
            sq = AUTO_HOLD_PRESSURE;
        } else if (g_auto_hold && (h & BTN_SQUARE) == 0) {
            /* RMB let go or the gun came down: release the held Square so the
               filter clears aimingState. */
            g_auto_hold = 0;
            r |= BTN_SQUARE;
            sq = 0;
            InterlockedIncrement(&g_pad_n[P_AUTOLOWER]);
        }
        if ((p & BTN_L3) != 0 && aim != 0) {
            /* Holding RMB must not toggle ready off. */
            p &= ~BTN_L3;
            InterlockedIncrement(&g_pad_n[P_DROP]);
        } else if ((r & BTN_L3) != 0 && aim != 0 && (h & BTN_SQUARE) == 0 && g_tap == 0) {
            if (keep_gun) {
                /* First person (0.8.2): the gun stays up at the right. */
                g_fp_keep = 1;
                rmb = 1;
                trace_line("first person: aim released, gun kept up id=%u", flags & 0xFFu);
            } else {
                /* RMB released: the filter's L3 path lowers without firing. */
                p |= BTN_L3;
                InterlockedIncrement(&g_pad_n[P_LOWER]);
            }
        }
        if (rmb && aim == 0 && (p & BTN_L3) == 0 && (h & BTN_SQUARE) == 0 && (r & BTN_SQUARE) == 0 && g_tap == 0) {
            if (++g_wait >= REARM_FRAMES) {
                p |= BTN_L3;
                g_wait = 0;
                InterlockedIncrement(&g_pad_n[P_REARM]);
            }
        } else {
            g_wait = 0;
        }
    }
    /* Shoulder aim (step 5a): hold L1 from the moment the gun is up until RMB
       is let go, so Snake keeps his facing and strafes. The lock does not
       need the gun up once it is on: after a semi-automatic shot aimingState
       is 0 until the re-ready, and a new L1 press there would take the
       facing from the stick again. Not in first person (R1 held). A real L1
       already held gets no added press, keeps its own release, and its edges
       are hidden while the lock is on. */
    lock = standard && rmb && (aim != 0 || g_lock) && (h & BTN_R1) == 0 &&
           ((flags & 0xFFu) < 15 || (flags & 0xFFu) > 17);
    if (lock && !g_lock) {
        g_lock = 1;
        if ((h & BTN_L1) == 0) {
            p |= BTN_L1;
        }
        /* A real L1 released on this very frame must not show as a release
           while the lock now holds L1. */
        r &= ~BTN_L1;
        InterlockedIncrement(&g_pad_n[P_LOCK]);
        trace_line("lock on id=%u", flags & 0xFFu);
    } else if (lock) {
        p &= ~BTN_L1;
        r &= ~BTN_L1;
    } else if (g_lock) {
        /* A real L1 pressed on this frame keeps its press. */
        g_lock = 0;
        if ((h & BTN_L1) == 0) {
            r |= BTN_L1;
        }
        InterlockedIncrement(&g_pad_n[P_UNLOCK]);
        trace_line("lock off id=%u", flags & 0xFFu);
    }
    if (g_lock) {
        h |= BTN_L1;
    }
    /* 0.8.1. First person (R1 held: the pad toggle, RB, or F) with a gun
       that has the iron-sight switch (ids 7, 9, 11, 12; the view follows L1
       held, 0x35A444, 0x359613): the aim (RMB, pad LT) also holds L1, so
       the iron sight comes up with the aim and no third button is held. */
    {
        const uint8_t *w = *(const uint8_t *const *)(actor + 0x6F8);
        uint8_t wid = w != NULL ? w[0x20] : 0;
        int ads = real_rmb && (h & BTN_R1) != 0 && !g_lock && (wid == 7 || wid == 9 || wid == 11 || wid == 12);
        if (ads) {
            h |= BTN_L1;
            r &= ~BTN_L1;
            if (g_ads || g_l1_prev) {
                p &= ~BTN_L1;
            } else {
                p |= BTN_L1;
            }
            if (!g_ads) {
                trace_line("iron sight on the aim id=%u", wid);
            }
        } else if (g_ads && (h & BTN_L1) == 0) {
            r |= BTN_L1;
        }
        g_ads = ads;
        g_l1_prev = (h & BTN_L1) != 0;
    }
    *held = h;
    *press = p;
    *release = r;
    *pressure = sq;
}

/* Component tree. Node layout from the node init: +0 next, +0x10 first
   child, +0x24 hash (low 24 bits), +0x40 handler, +0x58 sub-state. */
#define ACTOR_TREE 0x58
#define NODE_NEXT 0x00
#define NODE_CHILD 0x10
#define NODE_HASH 0x24
#define NODE_HANDLER 0x40
#define NODE_STATE 0x58
#define TREE_NODES 128
#define TREE_DEPTH 6
/* Lines per 2 s flush; the rest are counted as dropped. */
#define TRACE_LINES 200
#define RING_LINES 1024
#define RING_WIDTH 160

/* The two components registered with a 0x58-byte node have no +0x58 field. */
static const uint32_t kSmallNode[] = {0xF36F4Cu, 0x974037u};

struct comp {
    const uint8_t *node;
    uint32_t hash;
    uint32_t handler;
    uint32_t state;
    uint32_t frames;
};

static int g_trace;
static uintptr_t g_base;
static uintptr_t g_image_end;
static uintptr_t g_text_begin;
static uintptr_t g_text_end;
static struct comp g_comp[TREE_NODES];
static int g_ncomp;
static int g_marks;
static int g_f9;
static int g_deep_logged;

static char g_ring[RING_LINES][RING_WIDTH];
static volatile LONG g_ring_head;
static volatile LONG g_ring_tail;
static volatile LONG g_ring_window;
static volatile LONG g_ring_dropped;

/* Game thread only. The watch thread drains the ring every 2 s. */
static void trace_line(const char *fmt, ...) {
    LONG head = g_ring_head;
    char *line;
    int used;
    va_list args;
    SYSTEMTIME st;
    if (g_ring_window >= TRACE_LINES || head - g_ring_tail >= RING_LINES) {
        InterlockedIncrement(&g_ring_dropped);
        return;
    }
    line = g_ring[head % RING_LINES];
    GetLocalTime(&st);
    used = snprintf(line, RING_WIDTH, "fpvmove: %02u:%02u:%02u.%03u ", st.wHour, st.wMinute, st.wSecond, st.wMilliseconds);
    if (used < 0 || used >= RING_WIDTH) {
        return;
    }
    va_start(args, fmt);
    vsnprintf(line + used, RING_WIDTH - (size_t)used, fmt, args);
    va_end(args);
    MemoryBarrier();
    InterlockedIncrement(&g_ring_head);
    InterlockedIncrement(&g_ring_window);
}

static void trace_flush(void) {
    static char out[RING_LINES * (RING_WIDTH + 2) + 1];
    LONG tail = g_ring_tail;
    LONG head = g_ring_head;
    LONG dropped;
    size_t used = 0;
    MemoryBarrier();
    while (tail != head) {
        const char *line = g_ring[tail % RING_LINES];
        size_t n = strnlen(line, RING_WIDTH);
        memcpy(out + used, line, n);
        used += n;
        out[used++] = '\r';
        out[used++] = '\n';
        tail++;
    }
    InterlockedExchange(&g_ring_tail, tail);
    if (used > 0) {
        out[used] = '\0';
        append_log(out);
    }
    InterlockedExchange(&g_ring_window, 0);
    dropped = InterlockedExchange(&g_ring_dropped, 0);
    if (dropped > 0) {
        log_line("fpvmove: trace dropped=%ld", dropped);
    }
}

static uint32_t rva_in(uintptr_t value, uintptr_t begin, uintptr_t end) {
    return value >= begin && value < end ? (uint32_t)(value - g_base) : 0;
}

/* Depth-first, in walker order. Returns the node count, or -1 when the tree
   has more than TREE_NODES nodes. */
static int trace_collect(const uint8_t *actor, struct comp *out) {
    const uint8_t *stack[TREE_DEPTH];
    const uint8_t *node = *(const uint8_t *const *)(actor + ACTOR_TREE);
    int depth = 0;
    int n = 0;
    while (node != NULL || depth > 0) {
        const uint8_t *child;
        uint32_t hash;
        size_t i;
        int small_node = 0;
        if (node == NULL) {
            node = stack[--depth];
            continue;
        }
        if (n >= TREE_NODES || ((uintptr_t)node & 7u) != 0) {
            return -1;
        }
        hash = *(const uint32_t *)(node + NODE_HASH) & 0xFFFFFFu;
        for (i = 0; i < sizeof kSmallNode / sizeof kSmallNode[0]; i++) {
            small_node |= hash == kSmallNode[i];
        }
        out[n].node = node;
        out[n].hash = hash;
        out[n].handler = rva_in(*(const uintptr_t *)(node + NODE_HANDLER), g_base, g_image_end);
        out[n].state = small_node ? 0 : rva_in(*(const uintptr_t *)(node + NODE_STATE), g_text_begin, g_text_end);
        out[n].frames = 0;
        n++;
        child = *(const uint8_t *const *)(node + NODE_CHILD);
        if (child != NULL && depth < TREE_DEPTH) {
            stack[depth++] = *(const uint8_t *const *)(node + NODE_NEXT);
            node = child;
        } else {
            if (child != NULL && !g_deep_logged) {
                g_deep_logged = 1;
                trace_line("tree deeper than %d; subtree skipped", TREE_DEPTH);
            }
            node = *(const uint8_t *const *)(node + NODE_NEXT);
        }
    }
    return n;
}

static int trace_walk(const uint8_t *actor, struct comp *now) {
    __try {
        return trace_collect(actor, now);
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return -2;
    }
}

static void trace_step(const uint8_t *actor, uint32_t held) {
    static struct comp now[TREE_NODES];
    int n;
    int i;
    int j;
    int f9 = (GetAsyncKeyState(VK_F9) & 0x8000) != 0;
    if (f9 && !g_f9) {
        trace_line("mark %d", ++g_marks);
    }
    g_f9 = f9;
    n = trace_walk(actor, now);
    if (n < 0) {
        g_trace = 0;
        trace_line(n == -2 ? "tree read faulted; tracer off" : "tree over %d nodes; tracer off", TREE_NODES);
        return;
    }
    for (i = 0; i < g_ncomp; i++) {
        const struct comp *old = &g_comp[i];
        for (j = 0; j < n; j++) {
            if (now[j].node == old->node && now[j].hash == old->hash && now[j].handler == old->handler) {
                break;
            }
        }
        if (j == n) {
            trace_line("- h=%06X fn=0x%X st=0x%X n=%u", old->hash, old->handler, old->state, old->frames);
        }
    }
    for (j = 0; j < n; j++) {
        for (i = 0; i < g_ncomp; i++) {
            if (g_comp[i].node == now[j].node && g_comp[i].hash == now[j].hash && g_comp[i].handler == now[j].handler) {
                break;
            }
        }
        if (i == g_ncomp) {
            trace_line("+ h=%06X fn=0x%X st=0x%X held=0x%X", now[j].hash, now[j].handler, now[j].state, held);
        } else if (g_comp[i].state != now[j].state) {
            trace_line("st h=%06X 0x%X->0x%X n=%u held=0x%X", now[j].hash, g_comp[i].state, now[j].state,
                       g_comp[i].frames, held);
        } else {
            now[j].frames = g_comp[i].frames + 1;
        }
    }
    memcpy(g_comp, now, (size_t)n * sizeof now[0]);
    g_ncomp = n;
}

/* Crouch/roll split. The movement component reads Cross only from the
   actor's held and released words, which the pad component derives from
   the held word it copies, so editing the raw held bit is enough; the raw
   press and release bits are kept consistent for other readers. */
#define BTN_CROSS 0x4000u
#define PAD_CROSS_PRESSURE 0x12
#define ACTOR_MAG 0x7E0
#define ROLL_MAG 0x96
#define MOVE_HASH 0x7E8110u
#define KEY_CROSS 9
#define KEY_SQUARE 10
#define KEY_IDS 26
#define VK_CROUCH 0x43

enum { S_STILL, S_MOVE, S_SQUAT, S_S2P, S_PRONE, S_P2S, S_CRAWL, S_ROLLPRONE, S_CRAWL2, S_COUNT };
static const char *const kStateName[S_COUNT] = {"stand", "move", "crouch", "toprone", "prone", "fromprone", "crawl", "rolltoprone", "crawl2"};
enum { FIX_NONE, FIX_ROLL, FIX_CLAMP };

static int g_split;
static uintptr_t g_state[S_COUNT];
static const volatile uint32_t *g_keys;
static const volatile int32_t *g_layout;
static const uint32_t *g_table[3];
static int g_keys_ok;
static int32_t g_bind_layout = -1;
static uint32_t g_bind_cross[2];
static uint32_t g_bind_square[2];
static int g_bind_owner = -2;
static int g_space_mode; /* 0 up, 1 hidden, 2 native */
static int g_space_prev;
static int g_c_mode;
static int g_c_prev;
static int g_pulse;
static int g_cross_out;
static int g_fix;
static uintptr_t g_move_state;
static uintptr_t g_roll_state;
static int g_roll_hold;
/* Pad A (0.7.0): physical A held in gameplay, from the controller read. */
static volatile LONG g_pad_a;
static volatile ULONGLONG g_pad_a_tick;
/* A read older than this counts as A up. */
#define A_FRESH_MS 100
/* Set by the controller read when gameplay stops (no pad component call
   for CTRL_IDLE_POLLS reads); split_pre then restarts the rule. */
static volatile LONG g_pad_gap;
enum { A_NONE, A_PENDING, A_NATIVE, A_TAP, A_ROLLED };
#define A_HOLD_FRAMES 10
#define A_TAP_FRAMES 4
#define A_PRONE_FRAMES 12
static int g_a_mode;
static int g_a_prev;
static int g_a_frames;
/* CQC from a crouch (0.8.0), set by cqc_step for split_pre: 2 holds Cross,
   1 releases it, 0 leaves Cross to the rules. g_move_active is this
   frame's movement sub-state, or 0 while movement is suspended. */
static int g_stand_req;
static uintptr_t g_move_active;

static const uint8_t *node_find(const uint8_t *actor, uint32_t want) {
    const uint8_t *stack[TREE_DEPTH];
    const uint8_t *node = *(const uint8_t *const *)(actor + ACTOR_TREE);
    int depth = 0;
    int n = 0;
    while (node != NULL || depth > 0) {
        const uint8_t *child;
        if (node == NULL) {
            node = stack[--depth];
            continue;
        }
        if (++n > TREE_NODES || ((uintptr_t)node & 7u) != 0) {
            return NULL;
        }
        if ((*(const uint32_t *)(node + NODE_HASH) & 0xFFFFFFu) == want) {
            return node;
        }
        child = *(const uint8_t *const *)(node + NODE_CHILD);
        if (child != NULL && depth < TREE_DEPTH) {
            stack[depth++] = *(const uint8_t *const *)(node + NODE_NEXT);
            node = child;
        } else {
            node = *(const uint8_t *const *)(node + NODE_NEXT);
        }
    }
    return NULL;
}

/* Current sub-state of the component with this hash, or 0. */
static uintptr_t node_state(const uint8_t *actor, uint32_t want) {
    __try {
        const uint8_t *node = node_find(actor, want);
        return node != NULL ? *(const uintptr_t *)(node + NODE_STATE) : 0;
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return 0;
    }
}

/* The same, or 0 while the walker skips the node (flags +0x20 with any of
   bits 0-2 set, 0x366226): a suspended component keeps a stale sub-state,
   as movement does in a wall hug (0.8.0). */
static uintptr_t node_state_active(const uint8_t *actor, uint32_t want) {
    __try {
        const uint8_t *node = node_find(actor, want);
        return node != NULL && (*(const uint32_t *)(node + 0x20) & 7u) == 0 ? *(const uintptr_t *)(node + NODE_STATE) : 0;
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return 0;
    }
}

static int key_down(uint32_t vk) {
    return vk > 0 && vk < 0x100 && g_keys[vk] != 0;
}

/* Reads the live layout: the keys bound to Cross and to Square (fire), and
   whether C already belongs to a key id. C taken turns the crouch/roll split
   off, since C would then send two buttons; CQC on fire stays. Logs each
   change. */
static void split_bindings(void) {
    int32_t layout = *g_layout;
    const uint32_t *table = g_table[layout == 0 ? 0 : layout == 1 ? 1 : 2];
    uint32_t cross0 = table[KEY_CROSS * 4];
    uint32_t cross1 = table[KEY_CROSS * 4 + 1];
    uint32_t square0 = table[KEY_SQUARE * 4];
    uint32_t square1 = table[KEY_SQUARE * 4 + 1];
    int owner = -1;
    int id;
    for (id = 0; id < KEY_IDS; id++) {
        if (table[id * 4] == VK_CROUCH || table[id * 4 + 1] == VK_CROUCH) {
            owner = id;
        }
    }
    if (layout != g_bind_layout || owner != g_bind_owner || cross0 != g_bind_cross[0] || cross1 != g_bind_cross[1] ||
        square0 != g_bind_square[0] || square1 != g_bind_square[1]) {
        g_bind_layout = layout;
        g_bind_owner = owner;
        g_bind_cross[0] = cross0;
        g_bind_cross[1] = cross1;
        g_bind_square[0] = square0;
        g_bind_square[1] = square1;
        g_keys_ok = owner < 0;
        if (owner < 0) {
            trace_line("keys layout=%d roll=0x%X/0x%X fire=0x%X/0x%X crouch=C", layout, cross0, cross1, square0, square1);
        } else {
            trace_line("keys layout=%d roll=0x%X/0x%X fire=0x%X/0x%X; C is bound to key id %d, split off", layout, cross0,
                       cross1, square0, square1, owner);
        }
    }
}

static void split_pre(const uint8_t *actor) {
    uint32_t *held = (uint32_t *)(g_pad + PAD_HELD);
    uint32_t *press = (uint32_t *)(g_pad + PAD_PRESS);
    uint32_t *release = (uint32_t *)(g_pad + PAD_RELEASE);
    uint16_t mag = *(const uint16_t *)(actor + ACTOR_MAG);
    int native = (*held & BTN_CROSS) != 0;
    int pulse = g_pulse;
    int in_set = 0;
    int space;
    int c;
    int out;
    int i;
    g_fix = FIX_NONE;
    /* g_move_state is read by pad_wrap before cqc_step (0.8.0). */
    for (i = 0; i < S_COUNT; i++) {
        in_set |= g_move_state != 0 && g_move_state == g_state[i];
    }
    if (!g_keys_ok) {
        g_space_mode = 0;
        g_c_mode = 0;
        g_pulse = 0;
        g_a_mode = A_NONE;
        g_cross_out = native;
        return;
    }
    space = key_down(g_bind_cross[0]) || key_down(g_bind_cross[1]);
    c = key_down(VK_CROUCH);
    if (space && !g_space_prev) {
        if (!in_set) {
            g_space_mode = 2;
        } else {
            g_space_mode = 1;
            if (g_move_state == g_state[S_MOVE] && mag != 0 && !status(0xE5) && !g_c_mode && pulse == 0) {
                /* A one-frame tap: held now, released next call. */
                g_pulse = 2;
                pulse = 2;
                InterlockedIncrement(&g_pad_n[P_ROLL]);
            } else {
                InterlockedIncrement(&g_pad_n[P_HIDE]);
            }
        }
    } else if (!space) {
        g_space_mode = 0;
    }
    if (c && !g_c_prev && in_set) {
        g_c_mode = 1;
        InterlockedIncrement(&g_pad_n[P_CROUCH]);
    }
    if (g_c_mode && g_move_state == g_state[S_MOVE]) {
        /* Held, and on the release call (the mode clears below): the move
           sub-state reads the tap or the hold here, and at or below 0x96 it
           crouches instead of rolling. */
        g_fix = FIX_CLAMP;
    }
    if (!c) {
        g_c_mode = 0;
    }
    /* Pad A (Cross) from a run (0.7.0). Native Cross rolls from a run, so a
       crouch needs a stop first. A press in the move sub-state is held back:
       released within A_HOLD_FRAMES it becomes a crouch (Cross for
       A_TAP_FRAMES with the magnitude capped, as C does); still held then,
       it becomes the SPACE roll tap, and the roll ends in prone only if A
       stays held A_PRONE_FRAMES more. Elsewhere A stays native. */
    {
        ULONGLONG now = GetTickCount64();
        /* A read older than A_FRESH_MS (controller reads stopped) is up. */
        int a = g_pad_a != 0 && now - g_pad_a_tick < A_FRESH_MS && !space && !c;
        if (InterlockedExchange(&g_pad_gap, 0) != 0) {
            /* First call after gameplay stopped (pause, codec, window, as
               the controller reads saw it): start over, and an A held
               through it is not a new press. */
            g_a_mode = A_NONE;
            g_a_prev = a;
        }
        /* During the tap no new press is taken (a second tap there would
           release Cross without the cap and roll); an A held when the tap
           ends passes as native Cross. */
        if (a && !g_a_prev && g_a_mode != A_TAP) {
            if (in_set && g_move_state == g_state[S_MOVE] && mag != 0 && !status(0xE5) && !g_c_mode && pulse == 0) {
                g_a_mode = A_PENDING;
                g_a_frames = 0;
            } else {
                g_a_mode = A_NATIVE;
            }
        }
        if (g_a_mode == A_PENDING) {
            if (!a) {
                g_a_mode = A_TAP;
                g_a_frames = A_TAP_FRAMES;
                InterlockedIncrement(&g_pad_n[P_CROUCH]);
                trace_line("pad A tap: crouch");
            } else if (g_move_state != g_state[S_MOVE] || mag == 0 || status(0xE5) || g_c_mode) {
                /* Stopped while held (a crouch from standing), or the roll
                   gate is closed: native Cross from here. */
                g_a_mode = A_NATIVE;
            } else if (++g_a_frames >= A_HOLD_FRAMES) {
                g_a_mode = A_ROLLED;
                g_a_frames = 0;
                g_pulse = 2;
                pulse = 2;
                InterlockedIncrement(&g_pad_n[P_ROLL]);
                trace_line("pad A hold: roll");
            }
        } else if (g_a_mode == A_NATIVE || g_a_mode == A_ROLLED) {
            if (!a) {
                g_a_mode = A_NONE;
            }
        }
        g_a_prev = a;
    }
    out = native;
    if (g_space_mode == 1) {
        out = 0;
    }
    if (g_a_mode == A_PENDING) {
        out = 0;
    } else if (g_a_mode == A_TAP) {
        /* A_TAP_FRAMES held, then one released, all with the cap. */
        if (g_move_state == g_state[S_MOVE]) {
            g_fix = FIX_CLAMP;
        }
        out = g_a_frames > 0;
        if (g_a_frames-- <= 0) {
            g_a_mode = A_NONE;
        }
    } else if (g_a_mode == A_ROLLED) {
        /* The roll tap below; then the held A is hidden, unless it is still
           held A_PRONE_FRAMES later in the roll: then Cross is held until A
           is let go, whatever state follows (the roll into prone). */
        if (g_a_frames < A_PRONE_FRAMES) {
            g_a_frames++;
            out = 0;
        } else if (g_a_frames == A_PRONE_FRAMES) {
            out = g_roll_state != 0 && g_move_state == g_roll_state;
            if (out) {
                g_a_frames = A_PRONE_FRAMES + 1;
                InterlockedIncrement(&g_pad_n[P_ROLLPRONE]);
                trace_line("pad A held on: roll into prone");
            } else if (g_move_state == g_state[S_MOVE]) {
                /* No roll came and Snake still runs: A is native again. */
                g_a_mode = A_NATIVE;
                out = native;
            }
            /* Otherwise (stopped, the pulse crouched him) A stays hidden
               until it is let go. */
        } else {
            out = 1;
        }
    }
    if (pulse == 2) {
        out = 1;
        g_pulse = 1;
    } else if (pulse == 1) {
        g_pulse = 0;
        if (g_fix == FIX_NONE) {
            g_fix = FIX_ROLL;
        }
    }
    /* Roll key still held once the roll has begun: hold Cross, as a native
       long press would, so the roll ends in prone instead of recovering. It
       stays held until the key is released, whatever state follows. */
    if (!space) {
        g_roll_hold = 0;
    } else if (!g_roll_hold && g_space_mode == 1 && pulse == 0 && g_roll_state != 0 && g_move_state == g_roll_state) {
        g_roll_hold = 1;
        InterlockedIncrement(&g_pad_n[P_ROLLPRONE]);
    }
    if (g_roll_hold) {
        out = 1;
    }
    if (g_c_mode) {
        out = 1;
    }
    /* CQC from a crouch (0.8.0): its stand-up tap wins over the rules
       above for its two frames. */
    if (g_stand_req == 2) {
        out = 1;
    } else if (g_stand_req == 1) {
        out = 0;
    }
    if (out != native) {
        *held = out ? *held | BTN_CROSS : *held & ~BTN_CROSS;
        g_pad[PAD_CROSS_PRESSURE] = out ? 0xFF : 0;
    }
    *press = (*press & ~BTN_CROSS) | (out && !g_cross_out ? BTN_CROSS : 0);
    *release = (*release & ~BTN_CROSS) | (!out && g_cross_out ? BTN_CROSS : 0);
    g_cross_out = out;
    g_space_prev = space;
    g_c_prev = c;
}

/* After the pad component has written the magnitude for this frame, and
   before the movement component reads it. */
static void split_post(uint8_t *actor) {
    uint16_t *mag = (uint16_t *)(actor + ACTOR_MAG);
    if (g_fix == FIX_CLAMP) {
        if (*mag > ROLL_MAG) {
            *mag = ROLL_MAG;
            InterlockedIncrement(&g_pad_n[P_CLAMP]);
        }
    } else if (g_fix == FIX_ROLL) {
        if (g_move_state != g_state[S_MOVE] || *mag == 0) {
            InterlockedIncrement(&g_pad_n[P_ROLLMISS]);
        } else if (*mag <= ROLL_MAG) {
            *mag = ROLL_MAG + 1;
            InterlockedIncrement(&g_pad_n[P_ROLLFIX]);
        }
    }
}

/* CQC on the fire key (step 4b, option A). A Circle press starts a strike;
   the game turns it into a grab only if Circle is still held at the strike's
   animation event 0x0A and a contact is found, so Circle is held for as long
   as the fire key is. With a knife (weapon ids 1 and 2, whose Square is the
   slash), RMB held hides L3 so RMB plus LMB is a plain slash. */
#define BTN_CIRCLE 0x2000u
#define WEAPON_CQC 0x8000u

static int g_fire_prev;
static int g_cqc_mode;
static int g_cqc_prev;
static int g_circle_out;
static int g_knife;
static int g_hold_latch;
/* CQC from a crouch (0.8.0). The strike needs status 1, which the crouch
   sub-state never sets, so a fire press while crouched first stands Snake
   up: Cross held for one frame, then released (the crouch sub-state's
   tap/hold helper 0x36A590 sees the release and switches to the stand
   sub-state 0x3692C0 in the same frame). Once the movement sub-state is
   stand or move, the press becomes the CQC rule's Circle, held for at least
   one frame so a tap still punches. Square stays hidden from the press
   until the fire key is let go. g_cc_stage counts frames from the press,
   plus 1; 0 is idle. */
static int g_cc_stage;
static int g_cc_min;
static int g_cc_hide;
#define CC_STAND_FRAMES 6

/* The crouch sub-state of an active movement component (a crouched wall
   hug suspends movement with its crouch sub-state left behind), none of
   the strike entries' blockers (0xDF at 0x379C0E, 0xDF and 0x3B at
   0x377E72), and no other Cross rule in flight. */
static int cqc_crouch_ok(int rmb, const uint8_t *weapon, uint32_t flags, int holding) {
    int pad_a = g_pad_a != 0 && GetTickCount64() - g_pad_a_tick < A_FRESH_MS;
    return !rmb && weapon != NULL && (flags & WEAPON_CQC) != 0 && !holding && g_keys_ok && g_move_active != 0 &&
           g_move_active == g_state[S_SQUAT] && status(2) && !status(3) && !status(0xE0) && !status(0x3C) &&
           !status(0xDF) && !status(0x3B) && !key_down(VK_CROUCH) && !key_down(g_bind_cross[0]) &&
           !key_down(g_bind_cross[1]) && !pad_a && g_a_mode == A_NONE && !g_c_mode && g_pulse == 0 && !g_roll_hold;
}

static void cqc_step(const uint8_t *actor) {
    uint32_t *held = (uint32_t *)(g_pad + PAD_HELD);
    uint32_t *press = (uint32_t *)(g_pad + PAD_PRESS);
    uint32_t *release = (uint32_t *)(g_pad + PAD_RELEASE);
    const uint8_t *weapon = *(const uint8_t *const *)(actor + 0x6F8);
    uint32_t flags = weapon != NULL ? *(const uint32_t *)(weapon + 0x20) : 0;
    uint32_t id = flags & 0xFFu;
    /* A gun kept up in first person counts as aimed (0.8.2). */
    int rmb = (*held & BTN_L3) != 0 || g_fp_keep;
    /* The pad's RT, remapped to Square, counts as the fire key (0.5.9). */
    int fire = key_down(g_bind_square[0]) || key_down(g_bind_square[1]) || g_ctrl_rt != 0;
    int holding = status(0x6C);
    int out = (*held & BTN_CIRCLE) != 0;
    int stand = g_move_active != 0 && (g_move_active == g_state[S_STILL] || g_move_active == g_state[S_MOVE]);
    if (g_pad_gap != 0 && g_cc_stage != 0) {
        /* Gameplay stopped mid-sequence (split_pre consumes the flag): no
           late punch after it. */
        trace_line("crouch CQC: gameplay stopped at frame %d, no Circle", g_cc_stage - 1);
        g_cc_stage = 0;
    }
    if (fire && !g_fire_prev) {
        if (!rmb && weapon != NULL && (flags & WEAPON_CQC) != 0 && (holding || (status(1) && !status(0xDF)))) {
            g_cqc_mode = 1;
            InterlockedIncrement(&g_pad_n[P_CQC]);
        } else if (g_cc_stage == 0 && cqc_crouch_ok(rmb, weapon, flags, holding)) {
            g_cc_stage = 1;
            g_cc_hide = 1;
            InterlockedIncrement(&g_pad_n[P_CQCCROUCH]);
        }
        trace_line("fire press id=%u flags=0x%X rmb=%d hold=%d cqc=%d crouch=%d", id, flags, rmb, holding, g_cqc_mode,
                   g_cc_stage != 0);
    }
    g_stand_req = 0;
    if (g_cc_stage > 0) {
        if (rmb || holding || !g_keys_ok) {
            /* Aim pressed (or a hold began): the aim rules take over; the
               stand-up already sent finishes natively. */
            trace_line("crouch CQC: %s at frame %d, no Circle", rmb ? "aim pressed" : "handed over", g_cc_stage - 1);
            g_cc_stage = 0;
        } else if (g_cc_stage == 1) {
            g_stand_req = 2;
            g_cc_stage++;
        } else if (g_cc_stage == 2) {
            g_stand_req = 1;
            g_cc_stage++;
        } else if (stand) {
            g_cqc_mode = 1;
            g_cc_min = 1;
            trace_line("crouch CQC: standing at frame %d, Circle", g_cc_stage - 1);
            g_cc_stage = 0;
        } else if (++g_cc_stage > CC_STAND_FRAMES + 1) {
            InterlockedIncrement(&g_pad_n[P_STANDMISS]);
            trace_line("crouch CQC: not standing after %d frames, no Circle", CC_STAND_FRAMES);
            g_cc_stage = 0;
        }
    }
    if (g_hold_latch && (!holding || !rmb)) {
        /* RMB let go while holding: the hold goes on only if the fire key
           is still held, which becomes CQC again. Its Square stays hidden
           (the fire rules already hide it after a shot). */
        g_hold_latch = 0;
        if (fire && holding) {
            g_cqc_mode = 1;
        }
    }
    if (!fire) {
        /* A tapped crouch CQC still gets its one frame of Circle. */
        if (!g_cc_min) {
            g_cqc_mode = 0;
        }
        g_cc_hide = 0;
    }
    if (!g_hold_latch && holding && rmb && g_circle_out) {
        /* Aiming while holding an enemy: keep Circle held, so shots taken
           with LMB (Square) do not let the enemy go. */
        g_hold_latch = 1;
        InterlockedIncrement(&g_pad_n[P_KEEP]);
    }
    if (g_cqc_mode || g_cqc_prev || g_cc_stage != 0 || g_cc_hide) {
        /* The game never sees this Square: not while held, not its release. */
        *held &= ~BTN_SQUARE;
        *press &= ~BTN_SQUARE;
        *release &= ~BTN_SQUARE;
        g_pad[PAD_SQUARE_PRESSURE] = 0;
    }
    if (g_cqc_mode || g_hold_latch) {
        out = 1;
        *held |= BTN_CIRCLE;
    }
    *press = (*press & ~BTN_CIRCLE) | (out && !g_circle_out ? BTN_CIRCLE : 0);
    *release = (*release & ~BTN_CIRCLE) | (!out && g_circle_out ? BTN_CIRCLE : 0);
    g_circle_out = out;
    g_cqc_prev = g_cqc_mode;
    g_fire_prev = fire;
    g_cc_min = 0;
    /* Knife: from an RMB press until its release, L3 is hidden, except
       while holding an enemy (L3 there is the native interrogate). */
    if ((*press & BTN_L3) != 0 && weapon != NULL && (id == 1 || id == 2) && !holding) {
        g_knife = 1;
        InterlockedIncrement(&g_pad_n[P_KNIFE]);
    }
    if (g_knife) {
        if (!rmb) {
            g_knife = 0;
        }
        *held &= ~BTN_L3;
        *press &= ~BTN_L3;
        *release &= ~BTN_L3;
    }
}

/* Context lines, so a play session needs no F9 marks: weapon, gun up or
   down, first person (R1), holding an enemy, a real L1, the stance, and for
   each real L3 or L1 press whether a bound key or the pad sent it. Log only;
   reads the real pad, after the pad component has run. */
#define KEY_L3 12
#define KEY_L1 14
/* Frames with aimingState 0 before "gun down" is logged, longer than the
   re-ready after a shot. */
#define GUN_DOWN_FRAMES 20

enum { ST_OTHER, ST_STAND, ST_CROUCH, ST_PRONE, ST_ROLL, ST_NONE };
static const char *const kStance[] = {"other", "stand", "crouch", "prone", "roll", "none"};
static uint32_t g_ctx_flags = 0xFFFFFFFFu;
static int g_ctx_gun = -1;
static int g_ctx_gun_n;
static int g_ctx_fp = -1;
static int g_ctx_hold = -1;
static int g_ctx_l1 = -1;
static int g_ctx_stance = -1;
static int g_ctx_intr = -1;
static uintptr_t g_ctx_other;

static int bound_down(int id) {
    int32_t layout = *g_layout;
    const uint32_t *table = g_table[layout == 0 ? 0 : layout == 1 ? 1 : 2];
    return key_down(table[id * 4]) || key_down(table[id * 4 + 1]);
}

static int stance_of(uintptr_t st) {
    if (st == 0) {
        return ST_NONE;
    }
    if (st == g_state[S_STILL] || st == g_state[S_MOVE]) {
        return ST_STAND;
    }
    if (st == g_state[S_SQUAT]) {
        return ST_CROUCH;
    }
    if (st == g_state[S_S2P] || st == g_state[S_PRONE] || st == g_state[S_P2S] || st == g_state[S_CRAWL] ||
        st == g_state[S_ROLLPRONE] || st == g_state[S_CRAWL2]) {
        return ST_PRONE;
    }
    if (g_roll_state != 0 && st == g_roll_state) {
        return ST_ROLL;
    }
    return ST_OTHER;
}

static void ctx_step(const uint8_t *actor, uint32_t held, uint32_t press) {
    const uint8_t *weapon = *(const uint8_t *const *)(actor + 0x6F8);
    uint32_t flags = weapon != NULL ? *(const uint32_t *)(weapon + 0x20) : 0;
    int gun = g_aim[0] != 0;
    int fp = (held & BTN_R1) != 0;
    int hold = status(0x6C);
    int l1 = (held & BTN_L1) != 0;
    uintptr_t st = g_move_state; /* read by pad_wrap before cqc_step this frame */
    int stance = stance_of(st);
    if (flags != g_ctx_flags) {
        g_ctx_flags = flags;
        if (weapon != NULL) {
            trace_line("ctx weapon id=%u flags=0x%X", flags & 0xFFu, flags);
        } else {
            trace_line("ctx weapon none");
        }
    }
    if (g_ctx_gun < 0 || (gun && !g_ctx_gun)) {
        g_ctx_gun = gun;
        g_ctx_gun_n = 0;
        trace_line("ctx gun %s", gun ? "up" : "down");
    } else if (!gun && g_ctx_gun) {
        if (++g_ctx_gun_n >= GUN_DOWN_FRAMES) {
            g_ctx_gun = 0;
            g_ctx_gun_n = 0;
            trace_line("ctx gun down");
        }
    } else {
        g_ctx_gun_n = 0;
    }
    {
        /* 0.8.1: prone, "intrude" 0x4F, "grassintrude" 0x50, the automatic
           first-person latch 0xBB, the first-person view 0xBA. */
        int intr = status(3) | status(0x4F) << 1 | status(0x50) << 2 | status(0xBB) << 3 | status(0xBA) << 4;
        if (intr != g_ctx_intr) {
            g_ctx_intr = intr;
            trace_line("ctx intr prone=%d 4F=%d 50=%d BB=%d BA=%d", intr & 1, intr >> 1 & 1, intr >> 2 & 1, intr >> 3 & 1,
                       intr >> 4 & 1);
        }
    }
    if (fp != g_ctx_fp) {
        g_ctx_fp = fp;
        trace_line("ctx first person %s", fp ? "on" : "off");
    }
    if (hold != g_ctx_hold) {
        g_ctx_hold = hold;
        trace_line("ctx holding %s", hold ? "on" : "off");
    }
    if (l1 != g_ctx_l1) {
        g_ctx_l1 = l1;
        trace_line("ctx real L1 %s src=%s", l1 ? "down" : "up", !l1 ? "-" : bound_down(KEY_L1) ? "key" : "pad");
    }
    if ((press & BTN_L3) != 0) {
        trace_line("ctx aim press src=%s", bound_down(KEY_L3) ? "mouse" : "pad");
    }
    if (stance != g_ctx_stance || (stance == ST_OTHER && st != g_ctx_other)) {
        g_ctx_stance = stance;
        g_ctx_other = st;
        if (stance == ST_OTHER) {
            trace_line("ctx stance other 0x%X", (unsigned)(st - g_base));
        } else {
            trace_line("ctx stance %s", kStance[stance]);
        }
    }
}

/* Step 5b (0.5.9). While the lock is on, Snake's facing follows the
   third-person camera's orbit yaw, so what turns the camera (mouse or right
   stick) turns the aim. Both are 16-bit yaws, 0x10000 per turn, the same
   convention (stick direction = pad heading + camera heading). The write is
   made at message 0x10, before movement (message 0x20 copies +0x800 into its
   heading while strafing) and before the control update eases the model
   toward +0x16A. The camera reads only the raw pad, so the synthetic L1
   never switches its mode.
   The weapon aim helper 0x37A790 would still pick a target in view and
   write +0x800 toward it after this. Its one TEST(0xE3) call (0xE3 is a
   per-frame CQC flag, cleared by the weapon at message 0x10) goes through
   notarget_gate, which also passes while the lock is on: the helper then
   drops its target and skips its +0x800 writes. F8 turns that off (aim
   assist on). No game status is written. */
#define CAM_TARGET 0xE8
#define CAM_MODE 0x318
#define CAM_SUB 0x31C
#define CAM_YAW 0x338
#define ACT_POS 0x130
#define ACT_ROT_Y 0x162
#define ACT_TURN 0x16A
#define ACT_VEL 0x620
#define ACT_FACE 0x800
#define ST_STRAFE 0x11
#define AIM_TRACE_FRAMES 15
#define WEAPON_HASH 0xCF5135u
/* The gates below act only on flags set by a pad component call at most
   this long ago, so a stopped pad component cannot leave them on. */
#define GATE_FRESH_MS 100

typedef int(__fastcall *test_fn)(unsigned);
static test_fn g_test;
static int g_cstrafe_ok;
static void *g_last_actor;
static volatile LONG g_no_target;
static volatile ULONGLONG g_gate_tick;

static int gate_fresh(void) {
    return GetTickCount64() - g_gate_tick < GATE_FRESH_MS;
}

static int __fastcall notarget_gate(unsigned id) {
    if (g_test(id)) {
        return 1;
    }
    if (g_no_target && gate_fresh()) {
        InterlockedIncrement(&g_pad_n[P_NOTARGET]);
        return 1;
    }
    return 0;
}

/* Crouched strafe (0.5.9). The movement handler's strafe path needs status
   1, which the stand and move sub-states set (and others, not the crouch);
   a crouch sets status 2. The one
   TEST(1) call there (0x369C5A) goes through this gate, which also passes
   while the lock is on and the movement sub-state is the crouch. */
static int __fastcall cstrafe_gate(unsigned id) {
    if (g_test(id)) {
        return 1;
    }
    if (g_cstrafe && gate_fresh()) {
        InterlockedIncrement(&g_pad_n[P_CSTRAFE]);
        return 1;
    }
    return 0;
}

/* Crouched sideways strafe (0.8.1). The crouch sub-state has no sideways
   motion (it sends the squat walk for every direction); the crouched wall
   press has two sideways shuffles, base-archive motions 102 and 103 (named
   by the 0.8.0 wall motion probe; 100 and 101 are the standing ones). While
   the crouched strafe runs sideways (movement direction bit 8 or 4 at
   node+0x68, set with the strafe status at 0x369D61..0x369E12), the one
   movement motion send gets a copy of the word with layer 0 swapped for a
   shuffle; the upper body is not movement's. F6 turns it off and on. */
#define SIDE_MOTION_8 102
#define SIDE_MOTION_4 103
typedef uint64_t(__fastcall *send_fn)(void *, uint8_t *, uint64_t, const uint32_t *);
static send_fn g_send;
static int g_side_on = 1;
static int g_f6;

static uint64_t __fastcall send_wrap(void *actor, uint8_t *node, uint64_t msg, const uint32_t *word) {
    if (g_side_on && g_cstrafe && gate_fresh() && (uint32_t)msg == 0x18 && node != NULL &&
        word == (const uint32_t *)(node + 0x60) && *(const uintptr_t *)(node + NODE_STATE) == g_state[S_SQUAT]) {
        uint32_t bits = *(const uint32_t *)(node + 0x68);
        if ((bits & 0xCu) != 0) {
            uint32_t w = (*word & 0xFFFF0000u) | ((bits & 8u) != 0 ? SIDE_MOTION_8 : SIDE_MOTION_4);
            InterlockedIncrement(&g_pad_n[P_SIDESTEP]);
            return g_send(actor, node, msg, &w);
        }
    }
    return g_send(actor, node, msg, word);
}

enum { A_WRITE, A_OFF, A_FP, A_HOLD, A_STANCE, A_MOVE, A_NOCAM, A_CAMTARGET, A_SUB, A_MODE, A_FAULT, A_WALL, A_COUNT };
static const char *const kAimReason[A_COUNT] = {"write", "off", "first person", "holding", "stance", "moving without strafe",
                                                "no camera", "camera on another actor", "camera sub-mode", "camera mode",
                                                "camera fault", "wall"};
/* Wall hug and corner peek (0.8.0), defined after the aim rig. */
static int wall_candidate(const uint8_t *actor, const uint8_t **wnode);
static void wall_step(uint8_t *actor, const uint8_t *wnode, int wcase);
static int g_wall_ok;
static int g_wall_fail;
static int g_wall_on;
static uint16_t g_wall_yaw;
static uint8_t *const volatile *g_cam_pp;
static int g_aim_ok;
static int g_notarget_ok;
static int g_aim_reason = -1;
static int g_aim_mode = -1;
static int g_assist;
/* Aim assist default by save difficulty (0.6.0): on for Very Easy (10) and
   Easy (20), off above; F8 overrides until the difficulty changes. */
static const uint8_t *const volatile *g_diff_pp;
static int g_diff = -1;
static int g_diff_raw = -1;
static int g_assist_user = -1;
/* The facing was written this frame (read by the aim point builder). */
static int g_aim_writing;
static int g_f8;
static int g_aim_n;

/* The strafe gates movement tests at message 0x20 (0x369AC7..0x369C55)
   that can be read before it runs: TEST(3) clear, ANY(9, 0xBA, 7, 0x4F,
   0xDE) clear, a gun raised ANY(0x2E, 0x27), a weapon with a non-zero id,
   and 0x36A180 (weapon id 8 with a gun raised never strafes). L1 is held
   by the lock; status 1 and the stick are checked by the caller. */
static int strafe_expected(const uint8_t *actor) {
    const uint8_t *weapon = *(const uint8_t *const *)(actor + 0x6F8);
    uint8_t id;
    if (weapon == NULL) {
        return 0;
    }
    id = weapon[0x20];
    return id != 0 && id != 8 && !status(3) && !status(9) && !status(0xBA) && !status(7) && !status(0x4F) &&
           !status(0xDE) && (status(0x2E) || status(0x27));
}

static int game_has_focus(void) {
    DWORD pid = 0;
    HWND fg = GetForegroundWindow();
    if (fg == NULL) {
        return 0;
    }
    GetWindowThreadProcessId(fg, &pid);
    return pid == GetCurrentProcessId();
}

static void aim_step(uint8_t *actor, uint32_t held) {
    int reason = A_WRITE;
    int mode = -1;
    int wcase = 0;
    const uint8_t *wnode = NULL;
    int stand = g_move_state != 0 && (g_move_state == g_state[S_STILL] || g_move_state == g_state[S_MOVE]);
    int moving = *(const uint16_t *)(actor + ACTOR_MAG) != 0;
    uint16_t yaw = 0;
    int f8 = (GetAsyncKeyState(VK_F8) & 0x8000) != 0 && game_has_focus();
    int diff = 0;
    int assist_default;
    if (g_diff_pp != NULL) {
        __try {
            const uint8_t *block = *g_diff_pp;
            diff = block != NULL ? *(const int16_t *)(block + 6) : 0;
        } __except (EXCEPTION_EXECUTE_HANDLER) {
            diff = 0;
        }
    }
    if (diff != g_diff_raw) {
        /* Probe (0.6.1): the Hard save read no 10..60 value in 0.6.0. */
        g_diff_raw = diff;
        trace_line("ctx difficulty raw %d", diff);
    }
    if (diff < 10 || diff > 60 || diff % 10 != 0) {
        /* Not a save difficulty (title screen, menus): keep the last one. */
        diff = g_diff;
    }
    assist_default = diff == 10 || diff == 20;
    if (diff != g_diff) {
        g_diff = diff;
        g_assist_user = -1;
        trace_line("ctx difficulty %d, aim assist %s", diff, assist_default ? "on" : "off");
    }
    if (f8 && !g_f8) {
        g_assist_user = !(g_assist_user >= 0 ? g_assist_user : assist_default);
        trace_line("ctx aim assist %s", g_assist_user ? "on" : "off");
    }
    g_f8 = f8;
    g_assist = g_assist_user >= 0 ? g_assist_user : assist_default;
    if (!g_lock) {
        reason = A_OFF;
    } else if ((held & BTN_R1) != 0 || status(7)) {
        reason = A_FP;
    } else if (status(0x6C)) {
        reason = A_HOLD;
    } else if ((wcase = wall_candidate(actor, &wnode)) != 0) {
        /* The native wall camera (mode 5): the aim moves in a cone, with
           the rig built there but no camera writes; or, with no screen
           direction found, the native aim as before 0.8.0. */
        if (wcase > 0) {
            reason = A_WALL;
        } else {
            reason = A_MODE;
            mode = 5;
        }
    } else if (!(stand && status(1)) && !g_cstrafe) {
        reason = A_STANCE;
    } else if (moving && !strafe_expected(actor)) {
        /* Movement will face the stick this frame; do not fight it. */
        reason = A_MOVE;
    } else {
        __try {
            uint8_t *cam = *g_cam_pp;
            if (cam == NULL) {
                reason = A_NOCAM;
            } else if (*(uint8_t *const *)(cam + CAM_TARGET) != actor) {
                reason = A_CAMTARGET;
            } else if (*(const int32_t *)(cam + CAM_SUB) != 0) {
                reason = A_SUB;
            } else {
                mode = *(const int32_t *)(cam + CAM_MODE);
                /* Modes 4 and 5 copy a fixed camera's heading; mode 1 is a
                   front view that keeps the yaw at facing + 0x8000. */
                if (mode != 2 && mode != 3) {
                    reason = A_MODE;
                } else {
                    yaw = *(const uint16_t *)(cam + CAM_YAW);
                }
            }
        } __except (EXCEPTION_EXECUTE_HANDLER) {
            reason = A_FAULT;
        }
    }
    if (reason == A_WRITE) {
        *(uint16_t *)(actor + ACT_FACE) = yaw;
        *(uint16_t *)(actor + ACT_TURN) = yaw;
        InterlockedIncrement(&g_pad_n[P_AIM]);
    }
    if (reason == A_WALL) {
        __try {
            wall_step(actor, wnode, wcase);
        } __except (EXCEPTION_EXECUTE_HANDLER) {
            /* As rig_step's fault: the wall aim is off for the session. */
            g_wall_ok = 0;
            g_wall_on = 0;
            reason = A_FAULT;
            trace_line("wall aim fault; wall aim off");
        }
    } else {
        g_wall_on = 0;
    }
    if (!g_lock) {
        g_wall_fail = 0;
    }
    g_aim_writing = reason == A_WRITE || reason == A_WALL;
    /* Auto-target stays off for the whole lock, including frames the write
       is skipped for a moment: a single frame with it on would let the
       helper's press path snap the aim to the stick. It comes back where the
       camera cannot be used (fixed or scripted cameras), while holding an
       enemy, in first person, and with F8. */
    {
        LONG was = g_no_target;
        LONG now = g_notarget_ok && g_lock && !g_assist &&
                   (reason == A_WRITE || reason == A_MOVE || reason == A_STANCE || reason == A_WALL);
        InterlockedExchange(&g_no_target, now);
        if (was && !now && g_lock) {
            /* While forced, the helper marks its node for the L1 press path
               (+0xAA = 1). Clear that when the lock goes on without the gate,
               or the aim would snap to the stick for one frame. */
            __try {
                uint8_t *wnode = (uint8_t *)node_find(actor, WEAPON_HASH);
                if (wnode != NULL) {
                    *(uint16_t *)(wnode + 0xAA) = 0;
                }
            } __except (EXCEPTION_EXECUTE_HANDLER) {
            }
        }
    }
    if (g_lock && (reason != g_aim_reason || mode != g_aim_mode)) {
        if (reason == A_MODE) {
            trace_line("aim skip: camera mode %d", mode);
        } else {
            trace_line("aim %s%s", reason == A_WRITE || reason == A_WALL ? "" : "skip: ", kAimReason[reason]);
        }
    }
    g_aim_reason = g_lock ? reason : -1;
    g_aim_mode = mode;
    if (g_lock && ++g_aim_n >= AIM_TRACE_FRAMES) {
        const uint8_t *wnode = NULL;
        uintptr_t target = 0;
        uint16_t aim = 0;
        g_aim_n = 0;
        __try {
            wnode = node_find(actor, WEAPON_HASH);
            if (wnode != NULL) {
                target = *(const uintptr_t *)(wnode + 0x88);
                aim = *(const uint16_t *)(wnode + 0xA8);
            }
        } __except (EXCEPTION_EXECUTE_HANDLER) {
            wnode = NULL;
        }
        {
            float eye = 0.0f;
            __try {
                const uint8_t *fpv = *(const uint8_t *const *)(actor + 0x730);
                eye = fpv != NULL ? *(const float *)(fpv + 0x24) : 0.0f;
            } __except (EXCEPTION_EXECUTE_HANDLER) {
                eye = 0.0f;
            }
            trace_line("aim height pos=%.0f aimpt=%.0f eye=%.0f", *(const float *)(actor + ACT_POS + 4),
                       *(const float *)(actor + 0x524), eye);
        }
        trace_line("aim probe cam=%04X face=%04X turn=%04X rot=%04X aim=%04X target=%d notarget=%d strafe=%d mag=%u "
                   "pos=%.0f,%.0f vel=%.1f,%.1f",
                   yaw, *(const uint16_t *)(actor + ACT_FACE), *(const uint16_t *)(actor + ACT_TURN),
                   *(const uint16_t *)(actor + ACT_ROT_Y), aim, target != 0, (int)g_no_target, status(ST_STRAFE),
                   *(const uint16_t *)(actor + ACTOR_MAG), *(const float *)(actor + ACT_POS),
                   *(const float *)(actor + ACT_POS + 8), *(const float *)(actor + ACT_VEL),
                   *(const float *)(actor + ACT_VEL + 8));
    } else if (!g_lock) {
        g_aim_n = AIM_TRACE_FRAMES - 1;
    }
}

/* Controller layout (0.5.9). The game reads the controller through Steam
   Input into a buffer of its own (port 0), then applies the in-game button
   config and ORs the keyboard in, so editing that buffer after the read
   changes the pad only. In gameplay: LT is L3 (the aim the fire and lock
   rules read), RT is Square. Since 0.6.9 the left stick click toggles R1
   (first person; its own L3 is hidden, LT sends L3 while aiming) and the
   right stick click swaps the shoulder while LT aims (native R3 otherwise).
   0.8.0 (Phase 4, as in Delta New Style): D-pad left is L2 (a tap toggles
   the equipped item, a hold opens the item window), D-pad right is R2 (the
   same for weapons), D-pad down is Select (codec), and D-pad up is hidden
   from the game (the QCamo face-paint fork opens its menu on it). LB is
   native again: L1 (lock-on, camera reset, the first-person iron sight
   while held). RB toggles first person (0.8.1; natively R1 is a hold). Before 0.8.0, LB and RB
   sent L2 and R2, D-pad up R3 and D-pad down L1.
   Outside gameplay (menus, windows opened before the press) buttons stay
   native.
   Each button keeps the role it had when pressed until it is released. */
#define CTRL_LO 0x08
#define CTRL_HI 0x0A
#define CTRL_PRESSURE 0x14
/* Pressure word index per button. */
#define PW_RIGHT 0
#define PW_LEFT 1
#define PW_UP 2
#define PW_DOWN 3
#define PW_SQUARE 7
#define PW_L1 8
#define PW_R1 9
#define PW_L2 10
#define PW_R2 11
/* Trigger pull that counts as held, with hysteresis. */
#define TRIGGER_ON 0x40
#define TRIGGER_OFF 0x20
/* Port-0 reads without a pad component call before gameplay counts as over.
   A press latched as remapped within this many reads before that is turned
   back to native (a window or menu that opened right after). */
#define CTRL_IDLE_POLLS 2

typedef int(__fastcall *ctrl_fn)(int, uint8_t *);
static ctrl_fn g_ctrl_fn;
static const volatile int32_t *g_cfg_flag;
static const volatile uint32_t *g_cfg_table;

enum { SRC_LT, SRC_RT, SRC_LEFT, SRC_RIGHT, SRC_R3, SRC_UP, SRC_DOWN, SRC_L3, SRC_RB, SRC_COUNT };
static const char *const kSrcName[SRC_COUNT] = {"LT", "RT", "Dpad left", "Dpad right", "R3", "Dpad up", "Dpad down", "L3", "RB"};
/* 0.8.1. The right stick in the item and weapon windows: its X and Y words
   in the controller buffer (the ingame_stick_cam_dir action, 0x1151DC and
   0x115219; 0x80 centred, up is low), the deflection that counts, and the
   direction it sends (0 none, 1 right, 2 left, 3 up, 4 down). */
#define CTRL_RSTICK 0x0C
#define WIN_STICK_ON 0x40
#define WIN_STICK_OFF 0x28
static int g_win_stick;
/* 0.8.2. While a window is open the game's Steam Input read reports the
   camera stick centred (the 0.8.1 play: no browsing), so the right stick
   is read through XInput (xinput1_4, as QCamo does; slots 0-3). */
typedef struct {
    DWORD packet;
    WORD buttons;
    BYTE lt;
    BYTE rt;
    SHORT lx;
    SHORT ly;
    SHORT rx;
    SHORT ry;
} xpad_state;
typedef DWORD(WINAPI *xget_fn)(DWORD, xpad_state *);
static xget_fn g_xget;
static int g_xget_tried;
static volatile LONG g_win_probe_calls;

/* The right stick with the largest deflection over the XInput slots, in
   the buffer's scale (-128..127, up negative); 0 when none. */
static void xinput_rstick(int *rx, int *ry) {
    DWORD slot;
    int best = 0;
    *rx = 0;
    *ry = 0;
    if (!g_xget_tried) {
        HMODULE m = LoadLibraryExW(L"xinput1_4.dll", NULL, LOAD_LIBRARY_SEARCH_SYSTEM32);
        g_xget_tried = 1;
        g_xget = m != NULL ? (xget_fn)(void *)GetProcAddress(m, "XInputGetState") : NULL;
    }
    if (g_xget == NULL) {
        return;
    }
    for (slot = 0; slot < 4; slot++) {
        xpad_state st;
        int x;
        int y;
        int mag;
        memset(&st, 0, sizeof st);
        if (g_xget(slot, &st) != ERROR_SUCCESS) {
            continue;
        }
        x = st.rx / 256;
        y = -(st.ry / 256);
        mag = (x < 0 ? -x : x) + (y < 0 ? -y : y);
        if (mag > best) {
            best = mag;
            *rx = x;
            *ry = y;
        }
    }
}
enum { ROLE_UP, ROLE_NATIVE, ROLE_REMAP, ROLE_PENDING, ROLE_HIDE };
static int g_role[SRC_COUNT];
static LONG g_ctrl_seen;
static int g_ctrl_idle = CTRL_IDLE_POLLS;
static int g_ctrl_state = -1;
static unsigned g_ctrl_poll;
static unsigned g_press_poll[SRC_COUNT];
/* The left stick click toggles first person (R1 held) on and off (0.6.9;
   the right stick click before). */
static int g_fp_toggle;
/* 0.7.0: the poll of the last release of a window button (D-pad left or
   right since 0.8.0; LB or RB before), and how long after it a trigger
   press waits for gameplay. */
static unsigned g_win_poll = 0x80000000u;
static int g_wl_prev;
static int g_wr_prev;
#define WIN_GRACE_POLLS 30
#define TRIG_PENDING_POLLS 120
/* Right stick click while LT aims: a shoulder swap for rig_step. */
static volatile LONG g_pad_swap;
/* trace_line is game-thread only. The controller read is logged from it only
   when it runs on the thread that runs the pad component. */
static volatile DWORD g_game_tid;
static volatile DWORD g_ctrl_tid;
static int g_tid_logged;

static int cfg_default(void) {
    int i;
    if (*g_cfg_flag != 0) {
        return 1;
    }
    for (i = 0; i < 10; i++) {
        if (g_cfg_table[i] != (uint32_t)i) {
            return 0;
        }
    }
    return 1;
}

static void ctrl_remap(uint8_t *buf) {
    uint16_t *lo = (uint16_t *)(buf + CTRL_LO);
    uint16_t *hi = (uint16_t *)(buf + CTRL_HI);
    uint16_t *pw = (uint16_t *)(buf + CTRL_PRESSURE);
    int down[SRC_COUNT];
    int gameplay;
    int state;
    int i;
    LONG seq = g_pad_seq;
    int can_log = GetCurrentThreadId() == g_game_tid;
    g_ctrl_tid = GetCurrentThreadId();
    g_ctrl_poll++;
    if (seq != g_ctrl_seen) {
        g_ctrl_seen = seq;
        g_ctrl_idle = 0;
    } else if (g_ctrl_idle < CTRL_IDLE_POLLS) {
        if (++g_ctrl_idle == CTRL_IDLE_POLLS) {
            InterlockedExchange(&g_pad_gap, 1);
        }
    }
    gameplay = g_ctrl_idle < CTRL_IDLE_POLLS;
    state = !gameplay ? 0 : cfg_default() ? 1 : 2;
    if (state != g_ctrl_state && can_log) {
        g_ctrl_state = state;
        trace_line("pad layout %s", state == 1 ? "remap" : state == 2 ? "native: in-game button config changed" : "native");
    }
    down[SRC_LT] = (*hi & 0x0001u) != 0 || pw[PW_L2] >= (g_role[SRC_LT] != ROLE_UP ? TRIGGER_OFF : TRIGGER_ON);
    down[SRC_RT] = (*hi & 0x0002u) != 0 || pw[PW_R2] >= (g_role[SRC_RT] != ROLE_UP ? TRIGGER_OFF : TRIGGER_ON);
    down[SRC_LEFT] = (*lo & 0x0080u) != 0;
    down[SRC_RIGHT] = (*lo & 0x0020u) != 0;
    down[SRC_R3] = (*lo & 0x0004u) != 0;
    down[SRC_UP] = (*lo & 0x0010u) != 0;
    down[SRC_DOWN] = (*lo & 0x0040u) != 0;
    down[SRC_L3] = (*lo & 0x0002u) != 0;
    down[SRC_RB] = (*hi & 0x0008u) != 0;
    /* The item and weapon window buttons (D-pad left, right): their last
       release in that role (menus navigate with the native D-pad). */
    if ((g_wl_prev && !down[SRC_LEFT] && g_role[SRC_LEFT] == ROLE_REMAP) ||
        (g_wr_prev && !down[SRC_RIGHT] && g_role[SRC_RIGHT] == ROLE_REMAP)) {
        g_win_poll = g_ctrl_poll;
    }
    g_wl_prev = down[SRC_LEFT];
    g_wr_prev = down[SRC_RIGHT];
    for (i = 0; i < SRC_COUNT; i++) {
        if (!down[i]) {
            g_role[i] = ROLE_UP;
        } else if (g_role[i] == ROLE_UP) {
            g_role[i] = state == 1 ? ROLE_REMAP : ROLE_NATIVE;
            if (state == 0 && (i == SRC_LT || i == SRC_RT) && cfg_default() &&
                (g_ctrl_poll - g_win_poll <= WIN_GRACE_POLLS || g_role[SRC_LEFT] == ROLE_REMAP ||
                 g_role[SRC_RIGHT] == ROLE_REMAP)) {
                /* 0.7.0. A trigger pressed while the item or weapon window
                   closes (no pad component call yet) sends nothing until
                   gameplay is back (then its gameplay role) or
                   TRIG_PENDING_POLLS pass (then native). Also while a window
                   button is still held (a roll from the window button to
                   the trigger). As native L2/R2 it
                   opened the item window, or its quick tap toggled the
                   equipped item. */
                g_role[i] = ROLE_PENDING;
            }
            if ((i == SRC_LEFT || i == SRC_RIGHT || i == SRC_DOWN) && g_role[SRC_UP] == ROLE_REMAP) {
                /* 0.8.2: D-pad up holds the QCamo fork's menu open with the
                   game running; the other D-pad directions browse it and
                   reach the game neither as windows nor as the codec. */
                g_role[i] = ROLE_HIDE;
            }
            if (i == SRC_R3 && g_role[SRC_LT] != ROLE_REMAP) {
                /* The right stick click swaps the shoulder only while LT aims
                   (0.6.9); otherwise it keeps its native R3. */
                g_role[i] = ROLE_NATIVE;
            }
            g_press_poll[i] = g_ctrl_poll;
            if (g_role[i] == ROLE_REMAP) {
                InterlockedIncrement(&g_pad_n[P_REMAP]);
                if (i == SRC_R3) {
                    InterlockedExchange(&g_pad_swap, 1);
                    if (can_log) {
                        trace_line("pad R3 press, shoulder swap");
                    }
                }
                if (i == SRC_L3 || i == SRC_RB) {
                    /* 0.8.1: RB toggles first person too (native R1 is a
                       hold). */
                    g_fp_toggle = !g_fp_toggle;
                    if (can_log) {
                        trace_line("pad first person toggle %s (%s)", g_fp_toggle ? "on" : "off", kSrcName[i]);
                    }
                }
            }
            if (g_role[i] == ROLE_REMAP && can_log) {
                trace_line("pad %s press, remapped", kSrcName[i]);
            }
        }
    }
    if (state != 1) {
        for (i = 0; i < SRC_COUNT; i++) {
            /* D-pad down (Select) opens the codec at once, and D-pad up the
               QCamo fork's menu (its pause ends gameplay here); both keep
               their role, so neither reaches the game natively. */
            if (g_role[i] == ROLE_REMAP && i != SRC_DOWN && i != SRC_UP &&
                g_ctrl_poll - g_press_poll[i] <= CTRL_IDLE_POLLS) {
                g_role[i] = ROLE_NATIVE;
            }
        }
    }
    for (i = 0; i < SRC_COUNT; i++) {
        if (g_role[i] == ROLE_PENDING) {
            if (state == 1) {
                g_role[i] = ROLE_REMAP;
                InterlockedIncrement(&g_pad_n[P_REMAP]);
                if (can_log) {
                    trace_line("pad %s held into gameplay, remapped (%u polls after the press, %u after the window button)",
                               kSrcName[i], g_ctrl_poll - g_press_poll[i], g_ctrl_poll - g_win_poll);
                }
            } else if (g_ctrl_poll - g_press_poll[i] >= TRIG_PENDING_POLLS) {
                g_role[i] = ROLE_NATIVE;
                if (can_log) {
                    trace_line("pad %s pending timed out, native", kSrcName[i]);
                }
            }
        }
    }
    /* Pad A for the crouch/roll rule in split_pre (0.7.0): the physical A
       whenever the in-game button config is the default (A is Cross), in
       gameplay or not, so the first frame after a window sees it. */
    if (g_role[SRC_UP] == ROLE_REMAP) {
        /* 0.8.2: while D-pad up holds the QCamo menu, A (equip) and the
           right stick (select) belong to the menu, not to the game. */
        uint16_t *rs = (uint16_t *)(buf + CTRL_RSTICK);
        *hi &= (uint16_t)~0x0040u;
        pw[6] = 0;
        rs[0] = 0x80;
        rs[1] = 0x80;
    }
    InterlockedExchange(&g_pad_a, cfg_default() && (*hi & 0x0040u) != 0);
    g_pad_a_tick = GetTickCount64();
    /* Clear every remapped source first, then add the targets, so a target
       bit that is also some other source's native bit is not lost. */
    if (g_role[SRC_LT] == ROLE_REMAP || g_role[SRC_LT] == ROLE_PENDING) {
        *hi &= (uint16_t)~0x0001u;
        pw[PW_L2] = 0;
    }
    if (g_role[SRC_RT] == ROLE_REMAP || g_role[SRC_RT] == ROLE_PENDING) {
        *hi &= (uint16_t)~0x0002u;
        pw[PW_R2] = 0;
    }
    if (g_role[SRC_LEFT] == ROLE_REMAP || g_role[SRC_LEFT] == ROLE_HIDE) {
        *lo &= (uint16_t)~0x0080u;
        pw[PW_LEFT] = 0;
    }
    if (g_role[SRC_RIGHT] == ROLE_REMAP || g_role[SRC_RIGHT] == ROLE_HIDE) {
        *lo &= (uint16_t)~0x0020u;
        pw[PW_RIGHT] = 0;
    }
    if (g_role[SRC_DOWN] == ROLE_HIDE) {
        *lo &= (uint16_t)~0x0040u;
        pw[PW_DOWN] = 0;
    }
    if (g_role[SRC_R3] == ROLE_REMAP) {
        *lo &= (uint16_t)~0x0004u;
    }
    if (g_role[SRC_RB] == ROLE_REMAP) {
        *hi &= (uint16_t)~0x0008u;
        pw[PW_R1] = 0;
    }
    if (g_role[SRC_L3] == ROLE_REMAP) {
        /* LT adds its L3 back below while it aims. */
        *lo &= (uint16_t)~0x0002u;
    }
    if (g_role[SRC_UP] == ROLE_REMAP) {
        *lo &= (uint16_t)~0x0010u;
        pw[PW_UP] = 0;
    }
    if (g_role[SRC_DOWN] == ROLE_REMAP) {
        *lo &= (uint16_t)~0x0040u;
        pw[PW_DOWN] = 0;
    }
    if (g_role[SRC_LT] == ROLE_REMAP) {
        *lo |= 0x0002u;
    }
    if (g_role[SRC_RT] == ROLE_REMAP) {
        *hi |= 0x0080u;
        pw[PW_SQUARE] = 0xFF;
    }
    if (g_role[SRC_LEFT] == ROLE_REMAP) {
        *hi |= 0x0001u;
        pw[PW_L2] = 0xFF;
    }
    if (g_role[SRC_RIGHT] == ROLE_REMAP) {
        *hi |= 0x0002u;
        pw[PW_R2] = 0xFF;
    }
    /* The toggle survives menus and windows and applies again when
       gameplay resumes. */
    if (g_fp_toggle && state == 1) {
        *hi |= 0x0008u;
        pw[PW_R1] = 0xFF;
    }
    /* D-pad up sends nothing (the QCamo fork reads the pad itself). */
    if (g_role[SRC_DOWN] == ROLE_REMAP) {
        *lo |= 0x0001u;
    }
    /* 0.8.1. A window button held in from gameplay keeps its role while
       its window is open (gameplay stopped): the right stick then also
       sends the D-pad directions the windows browse with (0x32DB18,
       0x32DB2F), since the thumb on the D-pad cannot reach the left stick. */
    if (state != 1 && (g_role[SRC_LEFT] == ROLE_REMAP || g_role[SRC_RIGHT] == ROLE_REMAP)) {
        const uint16_t *rs = (const uint16_t *)(buf + CTRL_RSTICK);
        int rx = (int)rs[0] - 0x80;
        int ry = (int)rs[1] - 0x80;
        int ax;
        int ay;
        if (rx > -WIN_STICK_OFF && rx < WIN_STICK_OFF && ry > -WIN_STICK_OFF && ry < WIN_STICK_OFF) {
            xinput_rstick(&rx, &ry);
        }
        ax = rx < 0 ? -rx : rx;
        ay = ry < 0 ? -ry : ry;
        int dir = g_win_stick;
        if (dir != 0) {
            int v = dir == 1 ? rx : dir == 2 ? -rx : dir == 3 ? -ry : ry;
            if (v < WIN_STICK_OFF) {
                dir = 0;
            }
        }
        if (dir == 0) {
            if (ax >= WIN_STICK_ON && ax >= ay) {
                dir = rx > 0 ? 1 : 2;
            } else if (ay >= WIN_STICK_ON) {
                dir = ry < 0 ? 3 : 4;
            }
        }
        if (dir != g_win_stick || (g_ctrl_poll & 63u) == 0) {
            /* 0.8.3 probe: the 0.8.2 play still did not browse. */
            const uint16_t *bs = (const uint16_t *)(buf + CTRL_RSTICK);
            int xx = 0;
            int xy = 0;
            xinput_rstick(&xx, &xy);
            if (can_log) {
                trace_line("win stick buf=%d,%d xinput=%d,%d dir=%d lo=%04X hi=%04X state=%d", (int)bs[0] - 0x80,
                           (int)bs[1] - 0x80, xx, xy, dir, *lo, *hi, state);
            } else {
                g_win_probe_calls++;
            }
        }
        g_win_stick = dir;
        if (dir != 0) {
            /* 0.8.4: the 0.8.3 probe showed the direction right and the
               window still not browsing on the D-pad bit alone; the pad
               also carries D-pad pressure, and the left stick (which
               browses) is the move stick words +0x10/+0x12. Send all three
               while the physical left stick is centred. */
            uint16_t *ls = (uint16_t *)(buf + 0x10);
            int lx = (int)ls[0] - 0x80;
            int ly = (int)ls[1] - 0x80;
            *lo |= dir == 1 ? 0x0020u : dir == 2 ? 0x0080u : dir == 3 ? 0x0010u : 0x0040u;
            pw[dir == 1 ? PW_RIGHT : dir == 2 ? PW_LEFT : dir == 3 ? PW_UP : PW_DOWN] = 0xFF;
            if (lx > -WIN_STICK_OFF && lx < WIN_STICK_OFF && ly > -WIN_STICK_OFF && ly < WIN_STICK_OFF) {
                ls[0] = (uint16_t)(dir == 1 ? 0xFF : dir == 2 ? 0x00 : 0x80);
                ls[1] = (uint16_t)(dir == 3 ? 0x00 : dir == 4 ? 0xFF : 0x80);
            }
        }
    } else {
        g_win_stick = 0;
    }
    if (state == 1 && g_role[SRC_UP] == ROLE_REMAP) {
        /* 0.8.3: while D-pad up holds the QCamo menu with the game running,
           the menu takes over the pad: no button (first person, if on,
           stays) and both sticks centred reach the game. */
        uint16_t *sticks = (uint16_t *)(buf + CTRL_RSTICK);
        uint16_t keep = g_fp_toggle ? (uint16_t)(*hi & 0x0008u) : 0;
        int k;
        *lo = 0;
        *hi = keep;
        for (k = 0; k < 12; k++) {
            pw[k] = 0;
        }
        if (keep) {
            pw[PW_R1] = 0xFF;
        }
        for (k = 0; k < 4; k++) {
            sticks[k] = 0x80;
        }
        InterlockedExchange(&g_pad_a, 0);
    }
    InterlockedExchange(&g_ctrl_rt, g_role[SRC_RT] == ROLE_REMAP && !(state == 1 && g_role[SRC_UP] == ROLE_REMAP));
}

static int __fastcall ctrl_wrap(int port, uint8_t *buf) {
    int ok = g_ctrl_fn(port, buf);
    /* The caller reads ports 0 and 1 each poll; the game reads the
       controller for port 0 only. */
    if (port == 0) {
        if (ok) {
            ctrl_remap(buf);
        } else {
            int i;
            for (i = 0; i < SRC_COUNT; i++) {
                g_role[i] = ROLE_UP;
            }
            g_fp_toggle = 0;
            InterlockedExchange(&g_ctrl_rt, 0);
            InterlockedExchange(&g_pad_a, 0);
        }
    }
    return ok;
}

/* Manual pitch (0.6.1). Third-person bullets leave the gun along the arm
   pose, and the pose (0x3700A0) pitches the arms from bone 0xE toward the
   aim point +0x520 (clamped to -80/+50 degrees). While the lock is on, the
   camera's vertical input (mouse Y, right stick Y) no longer changes its
   height and zoom (0x212AA0 is skipped); it pitches the aim instead. The
   aim point builder (0x378C90) is detoured: for calls from the gun state
   it first sets the aim yaw node+0xA8 to the facing (no ease behind the
   camera), then puts the aim point at the pitch, at the builder's own
   distance along that yaw, from a pivot at the shoulder height. With the
   aim assist on (F8, easy saves) the native aim is left alone. */
typedef uint64_t(__fastcall *build_fn)(uint8_t *, uint8_t *, void *, void *);
typedef uint64_t(__fastcall *zoom_fn)(void *, void *, void *, void *);
/* Pivot height above the actor position: target bone 2 sits about 320
   above Snake's position for enemies on his floor (0.6.0 log). */
#define PITCH_PIVOT 320.0f
#define PITCH_MAX 0.70f
/* Radians per mouse delta unit (about the yaw rate), and per frame at a
   full right stick. */
#define PITCH_MOUSE 0.00075f
#define PITCH_STICK 0.020f
#define STICK_DEAD 0x30
static build_fn g_build_tramp;
static zoom_fn g_zoom_tramp;
static uintptr_t g_gun_lo;
static uintptr_t g_gun_hi;
static int g_pitch_ok;
static int g_zoom_ok;
static float g_pitch;
static int g_pitch_seeded;
static const volatile int32_t *g_mouse_dy;
/* The mouse's dx (the int32 before dy); set with the shoulder camera. */
static const volatile int32_t *g_mouse_dx;
/* The crosshair's world-to-screen matrix (channel 0) and its switch. */
static const float *g_dg_mtx;
static int g_xh_ok;
static ULONGLONG g_pitch_log;
/* 0.6.4 aim rig, built once a frame by rig_step in the pad wrapper: the
   shoulder camera's eye and centre ray, and the hit of that ray. */
struct rig {
    float eye[4];
    float dir[4];
};
static struct rig g_rig;
static float g_hit[4];
/* Distances along the ray from the eye: where the line check starts, and
   the hit. */
static float g_rig_t0;
static float g_rig_th;
static int g_rig_on;
/* 0.6.6. The shoulder camera's blend from the native camera (0) to the rig
   (1), eased over OTS_EASE_MS both ways; g_rig_cam is set while the rig is
   built for the camera (aiming, or easing out). */
static float g_ease_w;
static int g_rig_cam;
static LONG g_rig_seq;
static LONG g_rig_present;
static int g_ray_ok;
static int g_ots_on;
/* The aim point builder's own distance. */
#define AIM_POINT_MAX 10000.0f

/* The lock shows the crosshair and uses the camera for aiming. */
static int aim_live(void) {
    return g_aim_reason == A_WRITE || g_aim_reason == A_MOVE || g_aim_reason == A_STANCE || g_aim_reason == A_WALL;
}

/* Once per frame from the pad wrapper, after aim_step. */
static void pitch_step(void) {
    /* With the shoulder camera off (F7), as in 0.6.1: only while the facing
       follows the camera. */
    int active = g_pitch_ok && g_lock && !g_assist && !status(7) && (g_aim_writing || (g_ots_on && aim_live()));
    int32_t dy;
    int stick;
    if (!g_lock) {
        /* The pitch is kept for the shoulder camera's ease out; the next
           lock seeds it again. */
        g_pitch_seeded = 0;
        return;
    }
    if (!g_pitch_seeded && g_ease_w > 0.0f) {
        /* Locked again while the camera is still easing out: keep going
           from the pitch it shows. */
        g_pitch_seeded = 1;
    }
    if (!g_pitch_seeded && g_cam_pp != NULL) {
        /* Start from where the camera looks, not level. */
        __try {
            const uint8_t *cam = *g_cam_pp;
            if (cam != NULL) {
                const float *eye = (const float *)(cam + 0x380);
                const float *look = (const float *)(cam + 0x390);
                float dx = look[0] - eye[0];
                float dz = look[2] - eye[2];
                g_pitch = atan2f(look[1] - eye[1], sqrtf(dx * dx + dz * dz));
            }
        } __except (EXCEPTION_EXECUTE_HANDLER) {
            g_pitch = 0.0f;
        }
        g_pitch_seeded = 1;
    }
    if (!active) {
        return;
    }
    dy = g_mouse_dy != NULL ? *g_mouse_dy : 0;
    stick = (int)g_pad[5] - 0x80;
    if (dy != 0) {
        g_pitch += (float)dy * PITCH_MOUSE;
    } else if (stick > STICK_DEAD || stick < -STICK_DEAD) {
        /* Right stick up is a low byte. */
        g_pitch -= (float)(stick > 0 ? stick - STICK_DEAD : stick + STICK_DEAD) / (127.0f - STICK_DEAD) * PITCH_STICK;
    }
    if (g_pitch > PITCH_MAX) {
        g_pitch = PITCH_MAX;
    } else if (g_pitch < -PITCH_MAX) {
        g_pitch = -PITCH_MAX;
    }
}

static void pitch_post(uint8_t *a) {
    float *p = (float *)(a + 0x520);
    const float *pos = (const float *)(a + ACT_POS);
    float dx = p[0] - pos[0];
    float dz = p[2] - pos[2];
    float hd = sqrtf(dx * dx + dz * dz);
    if (hd < 1.0f) {
        return;
    }
    p[1] = pos[1] + PITCH_PIVOT + tanf(g_pitch) * hd;
    InterlockedIncrement(&g_pad_n[P_PITCH]);
    if (GetTickCount64() - g_pitch_log >= 1000) {
        g_pitch_log = GetTickCount64();
        trace_line("pitch %.1f deg aimpt=%.0f pos=%.0f dy=%d stick=%d", g_pitch * 57.2958f, p[1], pos[1],
                   g_mouse_dy != NULL ? *g_mouse_dy : 0, (int)g_pad[5] - 0x80);
    }
}

static uint64_t __fastcall builder_wrap(uint8_t *a, uint8_t *node, void *r8, void *r9) {
    uintptr_t ret = (uintptr_t)_ReturnAddress();
    uint64_t result;
    int gun = ret >= g_gun_lo && ret < g_gun_hi && a == (uint8_t *)g_last_actor;
    int own = gun && g_aim_writing && g_lock && gate_fresh();
    if (own) {
        /* The builder eases node+0xA8 toward the facing by a quarter per
           step; start it there so the shot follows the mouse at once. */
        *(uint16_t *)(node + 0xA8) = *(const uint16_t *)(a + ACT_FACE);
    }
    result = g_build_tramp(a, node, r8, r9);
    if (own && g_rig_on && g_rig_seq == g_pad_seq) {
        /* The pose aims the arms at the rig's hit, kept within the
           builder's own distance of Snake (a miss is 40000 out). */
        float *p = (float *)(a + 0x520);
        const float *pos = (const float *)(a + ACT_POS);
        float d[3];
        float len;
        float k = 1.0f;
        int i;
        for (i = 0; i < 3; i++) {
            d[i] = g_hit[i] - pos[i];
        }
        len = sqrtf(d[0] * d[0] + d[1] * d[1] + d[2] * d[2]);
        if (len > AIM_POINT_MAX) {
            k = AIM_POINT_MAX / len;
        }
        for (i = 0; i < 3; i++) {
            p[i] = pos[i] + d[i] * k;
        }
        InterlockedIncrement(&g_pad_n[P_PITCH]);
    } else if (own && g_pitch_ok && !g_assist && g_no_target && !status(7)) {
        pitch_post(a);
    }
    return result;
}

/* Over-the-shoulder aim (0.6.2 prototype, 0.6.4 rig). While the crosshair
   shows (locked, third person, usable camera), one rig gives the camera
   and the shot. Its shoulder point sits above Snake's position (plus the
   camera's eased stance offset, scaled) and OTS_SIDE to the side; the eye
   is OTS_BACK behind it along the aim direction (camera yaw and pitch),
   never lower than OTS_EYE_MIN above the position. Once a frame, in the
   pad wrapper, the game's line check casts the centre ray from the
   shoulder's depth outward. Its hit is the target of third-person shots
   (the targeted spawners of first person), the crosshair, and, while the
   facing follows the camera, the aim point +0x520 for the arm pose. The
   camera's wall step gets the rig's eye and a look-at along the ray in
   place of what the eye/look step made (its distance extension, offsets,
   and vertical ease are dropped); the wall step's wall pull and floor
   clamp still apply. F7 turns it off (0.6.1 behaviour), middle mouse swaps
   the shoulder. */
typedef uint64_t(__fastcall *wall_fn)(void *, void *, void *, void *);
typedef int(__fastcall *line_fn)(int, int, int, const float *, const float *, float);
typedef int64_t(__fastcall *hitidx_fn)(void);
typedef void(__fastcall *hitpt_fn)(int, float *);
#define OTS_SIDE 400.0f
#define OTS_BACK 2000.0f
#define OTS_HEIGHT 700.0f
#define OTS_STANCE 1.5f
#define OTS_EYE_MIN 100.0f
#define OTS_LOOK 1000.0f
#define AIM_RANGE 40000.0f
#define OTS_EASE_MS 150.0f
/* Wall behind: the eye stays this far in front of a wall, and at least
   this far behind the shoulder point. */
#define OTS_WALL_GAP 100.0f
#define OTS_MIN_BACK 300.0f
/* The camera's wall check flags switch (the wall step reads +0x60). */
#define CAM_WALL_FLAGS 0x60
/* A hit within about 70 degrees of the body yaw counts as reachable. */
#define REACH_COS 0.34f
/* Right stick yaw while aiming, as a share of the native rate. */
#define TURN_STICK 0.75f
static int g_rig_reach = 1;
#define CAM_STANCE 0x358
#define CAM_EYE 0x380
#define CAM_LOOK 0x390
static wall_fn g_wall_tramp;
static line_fn g_line;
static hitidx_fn g_hit_idx;
static hitpt_fn g_hit_pt;
static int g_stance_ok;
static int g_ots_ok;
static int g_ots_on = 1;
/* +1 right shoulder, -1 left; the middle mouse button (or the pad's left
   stick click while LT is held) swaps, and the side in use eases there. */
static float g_ots_side = 1.0f;
static float g_side_cur = 1.0f;
static int g_mmb;
static int g_f7;
static int g_ots_logged = -1;
static int g_xh_show;
static volatile LONG g_present_seq;
static float g_nat_eye[3];
static float g_nat_look[3];
static ULONGLONG g_rig_log;

static void rig_make(const float *pos, float adj, uint16_t yaw, float pitch, float side, struct rig *r, float *shoulder) {
    float a = (float)yaw * (3.14159265f / 32768.0f);
    float sa = sinf(a);
    float ca = cosf(a);
    float cp = cosf(pitch);
    float sp = sinf(pitch);
    float floor_y = pos[1] + OTS_EYE_MIN;
    int i;
    r->dir[0] = sa * cp;
    r->dir[1] = sp;
    r->dir[2] = ca * cp;
    r->dir[3] = 0.0f;
    /* Right of the view: (cos a, 0, -sin a). */
    shoulder[0] = pos[0] + ca * OTS_SIDE * side;
    shoulder[1] = pos[1] + OTS_HEIGHT + adj * OTS_STANCE;
    shoulder[2] = pos[2] - sa * OTS_SIDE * side;
    for (i = 0; i < 3; i++) {
        r->eye[i] = shoulder[i] - r->dir[i] * OTS_BACK;
    }
    r->eye[3] = 1.0f;
    if (r->eye[1] < floor_y) {
        r->eye[1] = floor_y;
    }
}

/* The player's own third-person camera in a mode the rig can drive. */
static uint8_t *rig_camera(const uint8_t *actor) {
    uint8_t *cam = *g_cam_pp;
    int32_t mode;
    if (cam == NULL || *(uint8_t *const *)(cam + CAM_TARGET) != actor || *(const int32_t *)(cam + CAM_SUB) != 0) {
        return NULL;
    }
    mode = *(const int32_t *)(cam + CAM_MODE);
    return mode == 2 || mode == 3 ? cam : NULL;
}

static void rig_probe(const uint8_t *actor, const uint8_t *cam, const float *from, int hit) {
    const float *pos = (const float *)(actor + ACT_POS);
    const float *fe = (const float *)(cam + 0x3D0);
    const float *fl = (const float *)(cam + 0x3E0);
    float dx = g_hit[0] - from[0];
    float dy = g_hit[1] - from[1];
    float dz = g_hit[2] - from[2];
    trace_line("rig p=%.1f yaw=%04X side=%+.0f pos=%.0f,%.0f,%.0f adj=%.0f", g_pitch * 57.2958f,
               *(const uint16_t *)(cam + CAM_YAW), g_ots_side, pos[0], pos[1], pos[2],
               g_stance_ok ? *(const float *)(cam + CAM_STANCE) : 0.0f);
    trace_line("rig eye=%.0f,%.0f,%.0f dir=%.2f,%.2f,%.2f hit=%d d=%.0f H=%.0f,%.0f,%.0f", g_rig.eye[0], g_rig.eye[1],
               g_rig.eye[2], g_rig.dir[0], g_rig.dir[1], g_rig.dir[2], hit, sqrtf(dx * dx + dy * dy + dz * dz), g_hit[0],
               g_hit[1], g_hit[2]);
    trace_line("cam nat eye=%.0f,%.0f,%.0f look=%.0f,%.0f,%.0f", g_nat_eye[0], g_nat_eye[1], g_nat_eye[2], g_nat_look[0],
               g_nat_look[1], g_nat_look[2]);
    trace_line("cam fin eye=%.0f,%.0f,%.0f look=%.0f,%.0f,%.0f wall=%.2f", fe[0], fe[1], fe[2], fl[0], fl[1], fl[2],
               *(const float *)(cam + 0x328));
    trace_line("cam d4=%.0f off=%.0f,%.0f,%.0f,%.0f", *(const float *)(cam + 0xD4), *(const float *)(cam + 0x360),
               *(const float *)(cam + 0x364), *(const float *)(cam + 0x368), *(const float *)(cam + 0x36C));
}

/* Milliseconds from the performance counter (GetTickCount64 steps by
   about 16 ms, too coarse for a 150 ms ease). */
static double now_ms(void) {
    static LARGE_INTEGER freq;
    LARGE_INTEGER t;
    if (freq.QuadPart == 0) {
        QueryPerformanceFrequency(&freq);
    }
    QueryPerformanceCounter(&t);
    return (double)t.QuadPart * 1000.0 / (double)freq.QuadPart;
}

static float move_toward(float v, float target, float step) {
    if (v < target) {
        return v + step < target ? v + step : target;
    }
    return v - step > target ? v - step : target;
}

/* Wall hug and corner peek (0.8.0). The wall-press component (hash
   0xD9728F, handler 0x34F950) keeps its state function at node+0x100 and
   its camera record at node+0x120; with the record enabled (+0xC bit
   0x400) the record owns channel 0 and the third-person camera reports
   mode 5, so the shoulder rig and the mouse aim were off there. Its states
   set 0x3C in a corner peek and 0x3C plus 0x3D in the pop-out aim hold,
   whose state copies +0x800 into +0x16A every frame (0x34D4B3, crouched
   0x34C241); pressed flat, the body is pinned to the wall yaw +0x232
   (0x34FD27) and only the arms turn. Natively the corner aim is auto-aim
   only (the helper skips the stick while 0x3C is set).
   Here the native wall camera stays. The mouse (or right stick) turns an
   aim yaw inside the game's own first-person cone from a wall, 0x31C7
   (about 70 degrees, 0x34FE00) each side of a base: the wall yaw +0x232
   when pressed or peeking (the first-person cone's centre, 0x34FDF9), the
   pop-out's own aim (+0x800, written at the pop-out event 0x34D8A3) in the
   aim hold. The pitch is the rig's. A ray along that aim from Snake's head
   (from the muzzle in the aim hold) gives the rig's hit, so the crosshair,
   the arm pose's aim point and the shot conversion work as in the open.
   Only the aim hold gets +0x800 written (its body follows it); pressed or
   peeking, +0x800 is left alone, so the pop-out event's own write stands.
   Which way a yaw increase moves on screen is read from the camera matrix
   each frame, so right is right under any wall camera; with no direction
   for WALL_NOSIGN_FRAMES the native auto-target aim takes over. */
#define WALL_HASH 0xD9728Fu
#define WALL_STATE 0x100
#define WALL_CAM 0x120
#define WALL_YAW 0x232
#define WALL_CONE 0x31C7
#define WALL_PROBE 400.0f
#define WALL_HEAD 700.0f
#define WALL_HEAD_CROUCH 450.0f
/* Radians per mouse delta unit (the pitch rate), and per frame at a full
   right stick. */
#define WALL_MOUSE 0.00075f
#define WALL_STICK 0.025f
/* Frames without a screen direction before the wall aim gives way to the
   native (auto-target) aim for the rest of that wall camera. */
#define WALL_NOSIGN_FRAMES 10
/* The gun's muzzle matrix (0x3649F0: the player global's component
   0xD60EE3, into rcx), also wrapped for the aim line below. */
#define MUZZLE_HASH 0xD60EE3u
typedef uint64_t(__fastcall *muzzle_fn)(float *, uint64_t, uint64_t, uint64_t);
static muzzle_fn g_muzzle;
/* The player global the muzzle function reads (its code at +0x12). */
static uint8_t *const volatile *g_player_pp;
static int g_wall_case;
static int g_wall_seeded;
static uint16_t g_wall_base;
static uint16_t g_wall_raw;
static int g_wall_sign;
static int g_wall_nosign;
static float g_wall_acc;
static ULONGLONG g_wall_log;

static int wall_candidate(const uint8_t *actor, const uint8_t **wnode) {
    const uint8_t *node;
    const uint8_t *rec;
    uintptr_t st;
    /* The crosshair's matrix also gives the turn direction on screen. */
    if (!g_wall_ok || !g_ots_ok || !g_ots_on || !g_pitch_ok || !g_ray_ok || !g_xh_ok || g_dg_mtx == NULL || g_assist) {
        return 0;
    }
    __try {
        const uint8_t *cam = *g_cam_pp;
        if (cam == NULL || *(const int32_t *)(cam + CAM_MODE) != 5) {
            g_wall_fail = 0;
            return 0;
        }
        node = node_find(actor, WALL_HASH);
        if (node == NULL) {
            g_wall_fail = 0;
            return 0;
        }
        st = *(const uintptr_t *)(node + WALL_STATE);
        rec = *(const uint8_t *const *)(node + WALL_CAM);
        if (st < g_text_begin || st >= g_text_end || rec == NULL || (*(const uint32_t *)(rec + 0xC) & 0x400u) == 0) {
            g_wall_fail = 0;
            return 0;
        }
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return 0;
    }
    if (g_wall_fail) {
        /* The native aim for this wall camera (aim_step: camera mode 5). */
        return -1;
    }
    *wnode = node;
    return status(0x3C) ? (status(0x3D) ? 3 : 2) : 1;
}

static void wall_dir(uint16_t yaw, float pitch, float *d) {
    float a = (float)yaw * (3.14159265f / 32768.0f);
    float cp = cosf(pitch);
    d[0] = sinf(a) * cp;
    d[1] = sinf(pitch);
    d[2] = cosf(a) * cp;
    d[3] = 0.0f;
}

static void wall_head(const uint8_t *actor, float *p) {
    const float *pos = (const float *)(actor + ACT_POS);
    p[0] = pos[0];
    p[1] = pos[1] + (status(2) ? WALL_HEAD_CROUCH : WALL_HEAD);
    p[2] = pos[2];
    p[3] = 1.0f;
}

/* Log only: a short wall check along the wall yaw +0x232 and opposite it,
   from Snake's head. +0x232 is taken as the outward direction: the body is
   pinned to it while pressed (0x34FD27) and the first-person cone from a
   wall is centred on it (0x34FDF9); the pressed wall should be the
   opposite hit. */
static void wall_probe(const uint8_t *actor, uint16_t raw) {
    float from[4];
    float dist[2];
    int k;
    int i;
    wall_head(actor, from);
    for (k = 0; k < 2; k++) {
        float d[4];
        float to[4];
        wall_dir((uint16_t)(raw + k * 0x8000), 0.0f, d);
        for (i = 0; i < 3; i++) {
            to[i] = from[i] + d[i] * WALL_PROBE;
        }
        to[3] = 1.0f;
        dist[k] = -1.0f;
        if (g_line(0x0F, *(const int32_t *)(actor + 0x120), 0x840050, from, to, 0.0f)) {
            float q[4];
            float s = 0.0f;
            g_hit_pt((int)g_hit_idx(), q);
            for (i = 0; i < 3; i++) {
                s += (q[i] - from[i]) * (q[i] - from[i]);
            }
            dist[k] = sqrtf(s);
        }
    }
    trace_line("wall aim: wall yaw %04X, wall check %.0f along it, %.0f opposite%s", raw, dist[0], dist[1],
               dist[0] >= 0.0f && (dist[1] < 0.0f || dist[0] < dist[1]) ? " (the nearer wall is along it)" : "");
}

/* The screen direction of a yaw increase under the camera of the last
   frame (channel-0 matrix): +1 right, -1 left, 0 unknown. For a camera
   beside the aim the direction depends on the point's distance, so the
   test uses the crosshair's point: last frame's wall ray and hit distance
   (halved until both points lie in front of the camera, down to 300). */
static int g_wall_rigged;

static int wall_screen_sign(const uint8_t *actor, uint16_t yaw) {
    float from[4];
    float t = g_wall_rigged ? g_rig_th : 1000.0f;
    int i;
    if (g_wall_rigged) {
        for (i = 0; i < 4; i++) {
            from[i] = g_rig.eye[i];
        }
    } else {
        wall_head(actor, from);
    }
    if (!(t >= 300.0f)) {
        t = 300.0f;
    }
    for (; t >= 300.0f; t *= 0.5f) {
        float x[2];
        int ok = 1;
        int k;
        for (k = 0; k < 2 && ok; k++) {
            float d[4];
            float p[3];
            float cx;
            float cw;
            const float *m = g_dg_mtx;
            wall_dir((uint16_t)(yaw + k * 0x400), g_pitch, d);
            for (i = 0; i < 3; i++) {
                p[i] = from[i] + d[i] * t;
            }
            cx = p[0] * m[0] + p[1] * m[4] + p[2] * m[8] + m[12];
            cw = p[0] * m[3] + p[1] * m[7] + p[2] * m[11] + m[15];
            if (cw <= 0.001f) {
                ok = 0;
            } else {
                x[k] = cx / cw;
            }
        }
        if (ok) {
            return x[1] > x[0] ? 1 : x[1] < x[0] ? -1 : 0;
        }
    }
    return 0;
}

static void wall_step(uint8_t *actor, const uint8_t *wnode, int wcase) {
    uint16_t raw = *(const uint16_t *)(wnode + WALL_YAW);
    int32_t dx = g_mouse_dx != NULL ? *g_mouse_dx : 0;
    int stick = (int)g_pad[4] - 0x80;
    float turn = 0.0f;
    int16_t off;
    if (!g_wall_on) {
        g_wall_on = 1;
        g_wall_seeded = 0;
        g_wall_case = 0;
        g_wall_sign = 0;
        g_wall_nosign = 0;
        g_wall_rigged = 0;
        g_wall_acc = 0.0f;
    }
    if (wcase == 3 && g_wall_case != 3) {
        /* The pop-out event set +0x800 to its aim: the cone's base. The
           aim so far carries over when it lies inside that cone. */
        uint16_t pop = *(const uint16_t *)(actor + ACT_FACE);
        int16_t diff = (int16_t)(g_wall_yaw - pop);
        trace_line("wall aim: pop-out aim %04X, wall yaw %04X, aim %04X", pop, g_wall_base, g_wall_yaw);
        g_wall_base = pop;
        if (!g_wall_seeded || diff > WALL_CONE || diff < -WALL_CONE) {
            g_wall_yaw = pop;
        }
        g_wall_seeded = 1;
    } else if (wcase != 3 && (!g_wall_seeded || g_wall_case == 3)) {
        g_wall_raw = raw;
        g_wall_base = raw;
        g_wall_yaw = raw;
        g_wall_seeded = 1;
        wall_probe(actor, raw);
    } else if (wcase != 3 && raw != g_wall_raw) {
        /* The wall yaw steps along a curved wall (0x34F0B6..0x34F0E3):
           the cone and the aim turn with it. */
        int16_t step = (int16_t)(raw - g_wall_raw);
        g_wall_raw = raw;
        g_wall_base = (uint16_t)(g_wall_base + step);
        g_wall_yaw = (uint16_t)(g_wall_yaw + step);
    }
    if (wcase != g_wall_case) {
        trace_line("wall aim: %s, base %04X", wcase == 1 ? "pressed" : wcase == 2 ? "corner peek" : "pop-out aim",
                   g_wall_base);
    }
    g_wall_case = wcase;
    if (dx != 0) {
        turn = (float)dx * WALL_MOUSE;
    } else if (stick > STICK_DEAD || stick < -STICK_DEAD) {
        turn = (float)(stick > 0 ? stick - STICK_DEAD : stick + STICK_DEAD) / (127.0f - STICK_DEAD) * WALL_STICK;
    }
    if (turn == 0.0f || g_wall_sign == 0) {
        /* The turn direction under the current camera (the wall camera
           changes between press, peek and aim hold), taken while the aim is
           not being turned, so an edge between a near and a far hit cannot
           flip it mid-turn; the last one found is kept while none is. */
        int sign = wall_screen_sign(actor, g_wall_yaw);
        if (sign != 0) {
            if (sign != g_wall_sign) {
                trace_line("wall aim: a yaw increase moves %s on screen", sign > 0 ? "right" : "left");
            }
            g_wall_sign = sign;
            g_wall_nosign = 0;
        } else if (g_wall_sign == 0 && ++g_wall_nosign > WALL_NOSIGN_FRAMES) {
            /* No direction at all: the native aim (auto-target) until this
               wall camera ends. */
            g_wall_fail = 1;
            trace_line("wall aim: no screen direction, native aim");
        }
    }
    if (g_wall_sign != 0 && turn != 0.0f) {
        /* Right (mouse or stick) moves the aim right on screen. Whole yaw
           units are applied; the rest carries to the next frame. */
        g_wall_acc += turn * (float)g_wall_sign * (32768.0f / 3.14159265f);
        g_wall_yaw = (uint16_t)(g_wall_yaw + (int)g_wall_acc);
        g_wall_acc -= (float)(int)g_wall_acc;
    }
    off = (int16_t)(g_wall_yaw - g_wall_base);
    if (off > WALL_CONE) {
        off = WALL_CONE;
    } else if (off < -WALL_CONE) {
        off = -WALL_CONE;
    }
    g_wall_yaw = (uint16_t)(g_wall_base + off);
    if (wcase == 3) {
        /* The aim hold turns the body with +0x800. */
        *(uint16_t *)(actor + ACT_FACE) = g_wall_yaw;
    }
    InterlockedIncrement(&g_pad_n[P_WALLAIM]);
}

/* The muzzle position into p, from the game's own muzzle matrix function;
   0 when it is not available or not near Snake. */
static int wall_muzzle(const uint8_t *actor, float *p) {
    float m[16];
    const float *pos = (const float *)(actor + ACT_POS);
    float dx;
    float dz;
    uint64_t ok;
    if (g_muzzle == NULL || g_player_pp == NULL) {
        return 0;
    }
    __try {
        /* 0x3649F0 works on the player global and dereferences its
           component without a check. */
        if (*g_player_pp != (const uint8_t *)actor || node_find(actor, MUZZLE_HASH) == NULL) {
            return 0;
        }
        memset(m, 0, sizeof m);
        ok = g_muzzle(m, 0, 0, 0);
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return 0;
    }
    if ((uint32_t)ok == 0 || (m[12] == 0.0f && m[13] == 0.0f && m[14] == 0.0f)) {
        return 0;
    }
    dx = m[12] - pos[0];
    dz = m[14] - pos[2];
    if (!(dx * dx + dz * dz < 2500.0f * 2500.0f) || !(fabsf(m[13] - pos[1]) < 2500.0f)) {
        return 0;
    }
    p[0] = m[12];
    p[1] = m[13];
    p[2] = m[14];
    p[3] = 1.0f;
    return 1;
}

/* The wall aim's rig: the ray along the aim, no camera writes. From
   Snake's head while pressed or peeking; from the muzzle in the pop-out
   aim hold, where the body may still stand behind the corner. From
   rig_step while the reason is A_WALL. */
static void wall_rig(uint8_t *actor) {
    float to[4];
    int hit = 0;
    int from_muzzle = 0;
    int i;
    if (g_wall_case == 3) {
        from_muzzle = wall_muzzle(actor, g_rig.eye);
    }
    if (!from_muzzle) {
        wall_head(actor, g_rig.eye);
    }
    wall_dir(g_wall_yaw, g_pitch, g_rig.dir);
    for (i = 0; i < 3; i++) {
        to[i] = g_rig.eye[i] + g_rig.dir[i] * AIM_RANGE;
    }
    to[3] = 1.0f;
    hit = g_line(0x1F, *(const int32_t *)(actor + 0x120), 0x42, g_rig.eye, to, 0.0f);
    if (hit) {
        g_hit_pt((int)g_hit_idx(), g_hit);
    } else {
        memcpy(g_hit, to, sizeof g_hit);
    }
    g_hit[3] = 1.0f;
    g_rig_t0 = 0.0f;
    g_rig_th = 0.0f;
    for (i = 0; i < 3; i++) {
        g_rig_th += (g_hit[i] - g_rig.eye[i]) * g_rig.dir[i];
    }
    g_wall_rigged = 1;
    /* The cone keeps the aim within reach of the wall's outward side. */
    if (!g_rig_reach) {
        trace_line("rig reach");
    }
    g_rig_reach = 1;
    g_rig_on = 1;
    g_rig_cam = 0;
    g_rig_seq = g_pad_seq;
    g_rig_present = g_present_seq;
    InterlockedIncrement(&g_pad_n[P_RIG]);
    if (GetTickCount64() - g_wall_log >= 1000) {
        g_wall_log = GetTickCount64();
        trace_line("wall aim case=%d yaw=%04X base=%04X p=%.1f face=%04X turn=%04X rot=%04X 532=%04X from=%s hit=%d "
                   "d=%.0f H=%.0f,%.0f,%.0f",
                   g_wall_case, g_wall_yaw, g_wall_base, g_pitch * 57.2958f, *(const uint16_t *)(actor + ACT_FACE),
                   *(const uint16_t *)(actor + ACT_TURN), *(const uint16_t *)(actor + ACT_ROT_Y),
                   *(const uint16_t *)(actor + 0x532), from_muzzle ? "muzzle" : "head", hit, g_rig_th, g_hit[0],
                   g_hit[1], g_hit[2]);
    }
}

/* Once a frame from the pad wrapper, after pitch_step: F7, the middle
   mouse button and the pad's shoulder swap, the eases, then the rig and
   (while aiming) its line check. */
static void rig_step(uint8_t *actor) {
    static double last;
    int f7 = (GetAsyncKeyState(VK_F7) & 0x8000) != 0 && game_has_focus();
    int mmb = (GetAsyncKeyState(VK_MBUTTON) & 0x8000) != 0 && game_has_focus();
    int pad_swap = InterlockedExchange(&g_pad_swap, 0) != 0;
    double now = now_ms();
    float dt = last > 0.0 && now > last ? (float)(now - last) : 0.0f;
    int aim;
    last = now;
    if (dt > 100.0f) {
        dt = 100.0f;
    }
    if (f7 && !g_f7) {
        g_ots_on = !g_ots_on;
        trace_line("ctx shoulder camera %s (F7)", g_ots_on ? "on" : "off");
    }
    g_f7 = f7;
    if ((mmb && !g_mmb) || pad_swap) {
        g_ots_side = -g_ots_side;
        trace_line("ctx shoulder %s (%s)", g_ots_side > 0.0f ? "right" : "left", pad_swap ? "pad" : "mouse");
    }
    g_mmb = mmb;
    g_side_cur = move_toward(g_side_cur, g_ots_side, 2.0f * dt / OTS_EASE_MS);
    g_rig_on = 0;
    g_rig_cam = 0;
    aim = g_ots_ok && g_ots_on && g_pitch_ok && g_lock && !g_assist && !status(7) && aim_live();
    if (!g_ots_ok || (!aim && g_ease_w <= 0.0f)) {
        g_ease_w = 0.0f;
        return;
    }
    __try {
        const uint8_t *cam = rig_camera(actor);
        const float *pos = (const float *)(actor + ACT_POS);
        float shoulder[3];
        float from[4];
        float to[4];
        float t0 = 0.0f;
        int hit = 0;
        int i;
        if (aim && g_aim_reason == A_WALL) {
            /* The native wall camera stays (0.8.0): no ease, no camera
               writes. */
            g_ease_w = 0.0f;
            wall_rig(actor);
            return;
        }
        if (cam == NULL) {
            /* Another camera took over: no ease back to it. */
            g_ease_w = 0.0f;
            return;
        }
        g_ease_w = move_toward(g_ease_w, aim ? 1.0f : 0.0f, dt / OTS_EASE_MS);
        rig_make(pos, g_stance_ok ? *(const float *)(cam + CAM_STANCE) : 0.0f, *(const uint16_t *)(cam + CAM_YAW), g_pitch,
                 g_side_cur, &g_rig, shoulder);
        if (g_ray_ok) {
            /* Wall behind (0.6.7): the game's wall step pulls the camera
               toward Snake's head. Instead, check from the ray point level
               with the shoulder back to the eye and move the eye forward
               along the ray, so the view keeps its line and stays beside
               the head. */
            float p0[4];
            float back = 0.0f;
            float d2 = 0.0f;
            int flags = *(const int32_t *)(cam + CAM_WALL_FLAGS) != 0;
            for (i = 0; i < 3; i++) {
                back += (shoulder[i] - g_rig.eye[i]) * g_rig.dir[i];
            }
            if (back > OTS_MIN_BACK) {
                for (i = 0; i < 3; i++) {
                    p0[i] = g_rig.eye[i] + g_rig.dir[i] * back;
                }
                p0[3] = 1.0f;
                if (g_line(flags ? 0x1F : 0x0F, *(const int32_t *)(actor + 0x120), flags ? 0x840042 : 0x840050, p0,
                           g_rig.eye, 0.0f)) {
                    float q[4];
                    float d;
                    g_hit_pt((int)g_hit_idx(), q);
                    for (i = 0; i < 3; i++) {
                        d2 += (q[i] - p0[i]) * (q[i] - p0[i]);
                    }
                    d = sqrtf(d2) - OTS_WALL_GAP;
                    if (d < OTS_MIN_BACK) {
                        d = OTS_MIN_BACK;
                    }
                    if (d < back) {
                        for (i = 0; i < 3; i++) {
                            g_rig.eye[i] = p0[i] - g_rig.dir[i] * d;
                        }
                        InterlockedIncrement(&g_pad_n[P_WALL]);
                    }
                }
            }
        }
        g_rig_cam = g_ease_w > 0.0f;
        if (!aim) {
            return;
        }
        /* From the ray point level with the shoulder (the eye may have been
           raised off the floor), so nothing between the camera and Snake,
           such as a wall behind him, is hit. */
        for (i = 0; i < 3; i++) {
            t0 += (shoulder[i] - g_rig.eye[i]) * g_rig.dir[i];
        }
        if (t0 < 0.0f) {
            t0 = 0.0f;
        }
        for (i = 0; i < 3; i++) {
            from[i] = g_rig.eye[i] + g_rig.dir[i] * t0;
            to[i] = from[i] + g_rig.dir[i] * AIM_RANGE;
        }
        from[3] = 1.0f;
        to[3] = 1.0f;
        if (g_ray_ok) {
            hit = g_line(0x1F, *(const int32_t *)(actor + 0x120), 0x42, from, to, 0.0f);
        }
        if (hit) {
            g_hit_pt((int)g_hit_idx(), g_hit);
        } else {
            memcpy(g_hit, to, sizeof g_hit);
        }
        g_hit[3] = 1.0f;
        g_rig_t0 = t0;
        g_rig_th = 0.0f;
        for (i = 0; i < 3; i++) {
            g_rig_th += (g_hit[i] - g_rig.eye[i]) * g_rig.dir[i];
        }
        {
            /* Reach (0.6.7): a hit well to the side of or behind Snake's
               body yaw (a wall beside him) cannot be shot; the shot stays
               native there, so the crosshair is hidden rather than show a
               point the gun cannot reach. */
            float a = *(const uint16_t *)(actor + ACT_ROT_Y) * (3.14159265f / 32768.0f);
            float vx = g_hit[0] - pos[0];
            float vz = g_hit[2] - pos[2];
            float vl = sqrtf(vx * vx + vz * vz);
            int reach = vl < 1.0f || (sinf(a) * vx + cosf(a) * vz) >= REACH_COS * vl;
            if (reach != g_rig_reach) {
                trace_line("rig %s", reach ? "reach" : "out of reach, crosshair hidden");
            }
            g_rig_reach = reach;
        }
        g_rig_on = 1;
        g_rig_seq = g_pad_seq;
        g_rig_present = g_present_seq;
        InterlockedIncrement(&g_pad_n[P_RIG]);
        if (GetTickCount64() - g_rig_log >= 1000) {
            g_rig_log = GetTickCount64();
            rig_probe(actor, cam, from, hit);
        }
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        g_rig_on = 0;
        g_rig_cam = 0;
        g_ease_w = 0.0f;
        g_ots_ok = 0;
        trace_line("rig fault; shoulder aim off");
    }
}

/* The camera's wall step, entered right after the eye/look step. With the
   rig built for this camera, its eye and look-at replace the eye/look
   step's result before the wall step reads them, blended with it while the
   shoulder camera eases in or out (smoothstep of g_ease_w). */
static uint64_t __fastcall wall_wrap(void *cam, void *rdx, void *r8, void *r9) {
    int on = g_rig_cam && gate_fresh() && cam != NULL && g_cam_pp != NULL && cam == (void *)*g_cam_pp &&
             g_last_actor != NULL && rig_camera((const uint8_t *)g_last_actor) == cam;
    if (on) {
        float *eye = (float *)((uint8_t *)cam + CAM_EYE);
        float *look = (float *)((uint8_t *)cam + CAM_LOOK);
        float w = g_ease_w;
        float s = w * w * (3.0f - 2.0f * w);
        int i;
        for (i = 0; i < 3; i++) {
            float re = g_rig.eye[i];
            float rl = g_rig.eye[i] + g_rig.dir[i] * OTS_LOOK;
            g_nat_eye[i] = eye[i];
            g_nat_look[i] = look[i];
            eye[i] += (re - eye[i]) * s;
            look[i] += (rl - look[i]) * s;
        }
        InterlockedIncrement(&g_pad_n[g_present_seq == g_rig_present ? P_EYENOW : P_EYEOLD]);
    }
    if (on != g_ots_logged && GetCurrentThreadId() == g_game_tid) {
        g_ots_logged = on;
        trace_line("ots %s", on ? "on" : "off");
    }
    {
        uint64_t result = g_wall_tramp(cam, rdx, r8, r9);
        if (on) {
            /* 0.6.7: the rig does its own wall pull along the ray, so the
               game's pull toward Snake's head (and its floor clamp) gives
               way to the rig by the same blend. The final eye and look are
               +0x3D0 and +0x3E0. */
            float *fe = (float *)((uint8_t *)cam + 0x3D0);
            float *fl = (float *)((uint8_t *)cam + 0x3E0);
            float w = g_ease_w;
            float s = w * w * (3.0f - 2.0f * w);
            int i;
            for (i = 0; i < 3; i++) {
                fe[i] += (g_rig.eye[i] - fe[i]) * s;
                fl[i] += (g_rig.eye[i] + g_rig.dir[i] * OTS_LOOK - fl[i]) * s;
            }
        }
        return result;
    }
}

/* Right stick yaw while aiming (0.6.7): the camera's yaw step reads the
   stick X input at +0x308; for the player's camera, while the rig is on and
   the mouse did not move this frame, it is scaled by TURN_STICK for the
   call and restored after. */
typedef uint64_t(__fastcall *turn_fn)(void *, void *, void *, void *);
static turn_fn g_turn_tramp;

static uint64_t __fastcall turn_wrap(void *cam, void *rdx, void *r8, void *r9) {
    float *in = (float *)((uint8_t *)cam + 0x308);
    float keep = 0.0f;
    int scaled = 0;
    uint64_t result;
    /* Not in the wall aim (0.8.0), where the stick turns the aim, not the
       camera. */
    if (g_rig_on && g_aim_reason != A_WALL && gate_fresh() && g_cam_pp != NULL && cam == (void *)*g_cam_pp &&
        g_mouse_dx != NULL &&
        *g_mouse_dx == 0 && *in != 0.0f) {
        keep = *in;
        *in = keep * TURN_STICK;
        scaled = 1;
        InterlockedIncrement(&g_pad_n[P_TURN]);
    }
    result = g_turn_tramp(cam, rdx, r8, r9);
    if (scaled) {
        *in = keep;
    }
    return result;
}

/* Third-person shots (0.6.4). The player's bullet callbacks call an
   untargeted spawner outside first person; its twin takes a target point
   as a new second argument and aims the bullet from the muzzle at it (the
   first-person path). While the rig is on, the six untargeted calls go to
   the twin with the rig's hit. Enemies' spawner calls are other sites. */
typedef uint64_t U64;
typedef U64(__fastcall *shot_a_fn)(void *, U64, U64, U64, U64, U64, U64, U64, U64, U64, U64);
typedef U64(__fastcall *shot_at_fn)(void *, const float *, U64, U64, U64, U64, U64, U64, U64, U64, U64, U64);
typedef U64(__fastcall *shot_b_fn)(void *, U64, U64, U64, U64, U64, U64, U64, U64, U64, U64, U64);
typedef U64(__fastcall *shot_bt_fn)(void *, const float *, U64, U64, U64, U64, U64, U64, U64, U64, U64, U64, U64);
typedef U64(__fastcall *shot_c_fn)(void *, U64, U64, U64, U64, U64, U64);
typedef U64(__fastcall *shot_ct_fn)(void *, const float *, U64, U64, U64, U64, U64, U64);
static shot_a_fn g_shot_a;
static shot_at_fn g_shot_at;
static shot_b_fn g_shot_b;
static shot_bt_fn g_shot_bt;
static shot_c_fn g_shot_c;
static shot_ct_fn g_shot_ct;
static float g_shot_pt[4];
/* The muzzle matrix copy the converted shot is spawned from. */
static float g_shot_m[16];
#define SHOT_BACKOFF 50.0f
/* The player's actor id (+0x120), cached by the pad wrapper. */
static volatile uint32_t g_player_id;
/* A shot takes the rig's hit only while the muzzle points within about 45
   degrees of it; otherwise (moving without the strafe, a pose the arms
   cannot reach) it keeps the native muzzle path. */
#define SHOT_COS 0.7f

/* The rig's hit for a shot by the player, or NULL. m is the muzzle matrix
   the spawner is given: row 2 (+0x20) its forward, row 3 (+0x30) its
   position. */
static const float *shot_target(const float *m, U64 id, char kind) {
    int mine = g_last_actor != NULL && GetCurrentThreadId() == g_game_tid && (uint32_t)id == g_player_id;
    float v[3];
    float fl;
    float vl;
    float c;
    int i;
    if (!mine || !g_rig_on || g_rig_seq != g_pad_seq || !gate_fresh()) {
        if (mine && g_lock) {
            trace_line("shot %c native", kind);
        }
        return NULL;
    }
    fl = 0.0f;
    vl = 0.0f;
    c = 0.0f;
    for (i = 0; i < 3; i++) {
        v[i] = g_hit[i] - m[12 + i];
        fl += m[8 + i] * m[8 + i];
        vl += v[i] * v[i];
        c += m[8 + i] * v[i];
    }
    fl = sqrtf(fl);
    vl = sqrtf(vl);
    if (fl < 0.001f || vl < 1.0f || c < SHOT_COS * fl * vl) {
        trace_line("shot %c native: muzzle %.0f deg off, d=%.0f", kind,
                   fl < 0.001f || vl < 1.0f ? 0.0f : acosf(fmaxf(-1.0f, fminf(1.0f, c / (fl * vl)))) * 57.2958f, vl);
        return NULL;
    }
    memcpy(g_shot_pt, g_hit, sizeof g_shot_pt);
    {
        /* Fire along the camera's centre ray (0.6.5): the bullet starts at
           the ray point nearest the muzzle, so its path is the crosshair's
           even when the line check misses (it did not hit guards in
           0.6.4, and a muzzle beside the ray then passed beside them). The
           start stays at or past the line check's start and short of the
           hit. */
        float tm = 0.0f;
        float lat = 0.0f;
        float lo = g_rig_t0;
        float hi = g_rig_th - SHOT_BACKOFF;
        memcpy(g_shot_m, m, sizeof g_shot_m);
        for (i = 0; i < 3; i++) {
            tm += (m[12 + i] - g_rig.eye[i]) * g_rig.dir[i];
        }
        if (tm > hi) {
            tm = hi;
        }
        if (tm < lo) {
            tm = lo;
        }
        for (i = 0; i < 3; i++) {
            g_shot_m[12 + i] = g_rig.eye[i] + g_rig.dir[i] * tm;
            lat += (g_shot_m[12 + i] - m[12 + i]) * (g_shot_m[12 + i] - m[12 + i]);
        }
        InterlockedIncrement(&g_pad_n[P_SHOT]);
        trace_line("shot %c H=%.0f,%.0f,%.0f d=%.0f off=%.1f deg moved=%.0f p=%.1f", kind, g_shot_pt[0], g_shot_pt[1],
                   g_shot_pt[2], vl, acosf(fminf(1.0f, c / (fl * vl))) * 57.2958f, sqrtf(lat), g_pitch * 57.2958f);
    }
    return g_shot_pt;
}

static U64 __fastcall shot_a_wrap(void *m, U64 id, U64 a3, U64 a4, U64 a5, U64 a6, U64 a7, U64 a8, U64 a9, U64 a10,
                                  U64 a11) {
    const float *t = shot_target((const float *)m, id, 'A');
    if (t != NULL) {
        return g_shot_at(g_shot_m, t, id, a3, a4, a5, a6, a7, a8, a9, a10, a11);
    }
    return g_shot_a(m, id, a3, a4, a5, a6, a7, a8, a9, a10, a11);
}

static U64 __fastcall shot_b_wrap(void *m, U64 id, U64 a3, U64 a4, U64 a5, U64 a6, U64 a7, U64 a8, U64 a9, U64 a10,
                                  U64 a11, U64 a12) {
    const float *t = shot_target((const float *)m, id, 'B');
    if (t != NULL) {
        return g_shot_bt(g_shot_m, t, id, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12);
    }
    return g_shot_b(m, id, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12);
}

static U64 __fastcall shot_c_wrap(void *m, U64 id, U64 a3, U64 a4, U64 a5, U64 a6, U64 a7) {
    const float *t = shot_target((const float *)m, id, 'C');
    if (t != NULL) {
        return g_shot_ct(g_shot_m, t, id, a3, a4, a5, a6, a7);
    }
    return g_shot_c(m, id, a3, a4, a5, a6, a7);
}

/* Hold-up drop (0.8.0). A held-up soldier shakes and drops an item only
   in first person: his hold-up states test status 0xBA (the first-person
   view, set only by the first-person component) before the shake
   (0x2010AA, 0x202D19; variant B tests status 7 at 0x2015EE). The shake
   tests then want distance, a gun raised and Snake's body facing him;
   mode 0xF's (0x19C5B0) also that the last ammo used was a gun's, mode
   0xE's (0x19D3A0) also a clear line and the player's aim line (actor+
   0xA20, a target record of kind 0xAD) on body part 1 or 3 of his damage
   receiver, unless the last ammo used was a gun's or the weapon is id 14
   or 17. Outside first person that line
   runs along the muzzle. While the shoulder aim is live, the three tests
   also pass, and the aim line runs along the rig's ray, from the ray point
   nearest the muzzle (as converted shots do). Everything after the gate
   is the game's first-person behaviour, including a soldier leaving the
   hold-up after three shake cycles (0x2010E4, 0x96 to 0x97). */
static ULONGLONG g_holdup_log;
/* 0.8.1. The held-up soldier the gates last passed for (rbx at the three
   sites; the gates reach holdup_gate through a stub that copies rbx to
   rdx), his position (soldier+0x30, the point the game's own line-of-sight
   test aims at, 0x19D43E), and the body part his damage receiver last
   saw under the aim line ([soldier+0x1CB8]+0x12C, 0x1BE410). */
static ULONGLONG g_hu_tick;
static float g_hu_pos[3];
static int g_hu_part = -2;
static ULONGLONG g_hu_part_log;
#define HOLDUP_FRESH_MS 150
/* First-person eye height above Snake's position (the 0.5.x logs: eye
   1476 at position 800). */
#define HOLDUP_EYE 676.0f

static int holdup_live(void) {
    return g_lock && g_rig_on && g_rig_reach && gate_fresh() && !status(7) && GetCurrentThreadId() == g_game_tid;
}

static int __fastcall holdup_gate(unsigned id, const uint8_t *soldier) {
    if (g_test(id)) {
        return 1;
    }
    if (holdup_live()) {
        InterlockedIncrement(&g_pad_n[P_HOLDUP]);
        if (GetTickCount64() - g_holdup_log >= 2000) {
            g_holdup_log = GetTickCount64();
            trace_line("hold-up: status 0x%X passed for the shoulder aim", id);
        }
        if (soldier != NULL) {
            __try {
                const uint8_t *rec = *(const uint8_t *const *)(soldier + 0x1CB8);
                int part = rec != NULL ? *(const int32_t *)(rec + 0x12C) : -3;
                g_hu_pos[0] = *(const float *)(soldier + 0x30);
                g_hu_pos[1] = *(const float *)(soldier + 0x34);
                g_hu_pos[2] = *(const float *)(soldier + 0x38);
                g_hu_tick = GetTickCount64();
                if (part != g_hu_part && GetTickCount64() - g_hu_part_log >= 200) {
                    g_hu_part_log = GetTickCount64();
                    g_hu_part = part;
                    trace_line("hold-up: aim line on part %d (gate 0x%X) pitch %.1f", part, id, g_pitch * 57.2958f);
                }
            } __except (EXCEPTION_EXECUTE_HANDLER) {
                g_hu_tick = 0;
            }
        }
        return 1;
    }
    return 0;
}

static U64 __fastcall aimline_wrap(float *m, U64 rdx, U64 r8, U64 r9) {
    U64 result = g_muzzle(m, rdx, r8, r9);
    const uint8_t *obj = (const uint8_t *)m - 0x90;
    /* obj+0x50: bit 1 the drawn laser (kept on the gun), bit 8 the first-
       person target mode. */
    if (holdup_live() && g_rig_seq == g_pad_seq && g_player_pp != NULL && *g_player_pp == (uint8_t *)g_last_actor &&
        (obj[0x50] & 9) == 0) {
        float t = 0.0f;
        float hi = g_rig_th - SHOT_BACKOFF;
        int done = 0;
        int i;
        if (GetTickCount64() - g_hu_tick < HOLDUP_FRESH_MS) {
            /* A held-up soldier (0.8.1): as in first person, the line runs
               from Snake's eye (centred, not the shoulder camera's side) to
               the crosshair's point at the soldier's depth, so it meets
               his body where the crosshair shows it instead of crossing
               his side first. */
            const float *pos = (const float *)((const uint8_t *)g_last_actor + ACT_POS);
            float h = g_rig.dir[0] * g_rig.dir[0] + g_rig.dir[2] * g_rig.dir[2];
            float tg = h > 1e-4f ? ((g_hu_pos[0] - g_rig.eye[0]) * g_rig.dir[0] +
                                    (g_hu_pos[2] - g_rig.eye[2]) * g_rig.dir[2]) / h
                                 : -1.0f;
            if (tg > g_rig_t0 && tg < 10000.0f) {
                float from[3];
                float d[3];
                float len = 0.0f;
                from[0] = pos[0];
                from[1] = pos[1] + (status(2) ? WALL_HEAD_CROUCH : HOLDUP_EYE);
                from[2] = pos[2];
                for (i = 0; i < 3; i++) {
                    d[i] = g_rig.eye[i] + g_rig.dir[i] * tg - from[i];
                    len += d[i] * d[i];
                }
                len = sqrtf(len);
                if (len > 1.0f) {
                    for (i = 0; i < 3; i++) {
                        m[8 + i] = d[i] / len;
                        m[12 + i] = from[i];
                    }
                    done = 1;
                }
            }
        }
        if (!done) {
            for (i = 0; i < 3; i++) {
                t += (m[12 + i] - g_rig.eye[i]) * g_rig.dir[i];
            }
            if (t > hi) {
                t = hi;
            }
            if (t < g_rig_t0) {
                t = g_rig_t0;
            }
            for (i = 0; i < 3; i++) {
                m[8 + i] = g_rig.dir[i];
                m[12 + i] = g_rig.eye[i] + g_rig.dir[i] * t;
            }
        }
        m[11] = 0.0f;
        m[15] = 1.0f;
        InterlockedIncrement(&g_pad_n[P_AIMLINE]);
    }
    return result;
}

/* The camera's height/zoom from its vertical input, skipped for the
   player's camera while the pitch takes that input. */
static uint64_t __fastcall zoom_wrap(void *cam, void *rdx, void *r8, void *r9) {
    if (g_pitch_ok && g_lock && !g_assist && gate_fresh() && g_cam_pp != NULL && cam == (void *)*g_cam_pp) {
        return 0;
    }
    return g_zoom_tramp(cam, rdx, r8, r9);
}

/* Crosshair (0.6.0). While the lock is on in third person, a small cross is
   drawn on the back buffer where the aim point projects, with the game's
   own world-to-screen: clip = point x the channel-0 matrix, then PS2 pixels
   (x/w*0.5+0.5)*W+X0 (0x123F70). PS2 pixels map to the viewport the game
   last set (renderer +0x88/+0x8C, centered), or the whole back buffer.
   Drawn with ID3D11DeviceContext1::ClearView on a view of the swap chain's
   current buffer, after the renderer's last pre-present call; no pipeline
   state changes. Snapshot in the pad wrapper, drawn on the same thread. */
typedef uint64_t(__fastcall *void_fn)(void *, void *, void *, void *);
static void_fn g_pre_present;
static uint8_t *const volatile *g_renderer_pp;
static const volatile int32_t *g_dg_view[4]; /* W, H, X0, Y0 */
static const volatile int32_t *g_pause;
static LONG g_xh_seq;
static LONG g_xh_seen;
static int g_xh_idle;
static ULONGLONG g_xh_log;
#define RND_DEVICE 0x298
#define RND_CONTEXT1 0x2A8
#define RND_SWAPCHAIN 0x2B0
#define RND_VIEW_W 0x88
#define RND_VIEW_H 0x8C

/* With the rig on, the crosshair marks the rig's hit (where the shots
   go); otherwise the aim point +0x520 read at present time. */
static int g_xh_rig;
static float g_xh_pt[3];
/* 0.8.0. First person with a gun that has the iron-sight switch (weapon
   ids 7, 9, 11, 12): the view follows L1 held every frame (rifle component
   0x35A444/0x35A612, pistol component 0x359613), iron sight while held,
   the weapon at the right otherwise. The first-person shot ray runs from
   the camera eye along its forward axis (0x373C10, 0x3749CA), so the aim
   is the screen centre in both views; the crosshair marks it in the
   weapon-at-right view only. The view logic runs only with status 0xBA,
   ANY(0x2E, 0x27) and 0x2F clear; status 0x27 forces the right view
   (0x35A4DE). */
static int g_xh_centre;
static int g_xh_fp_logged = -1;

static int fp_crosshair(const uint8_t *actor) {
    const uint8_t *weapon = *(const uint8_t *const *)(actor + 0x6F8);
    int l1 = (*(const uint32_t *)(actor + 0x7E8) & BTN_L1) != 0;
    uint8_t id;
    if (weapon == NULL || !status(7) || !status(0xBA)) {
        return 0;
    }
    id = weapon[0x20];
    if (id == 9 || id == 11 || id == 12) {
        /* Rifle component: 0x35A292..0x35A2C9, 0x35A410, 0x35A444. */
        if (status(0x2F) || !(status(0x2E) || status(0x27))) {
            return 0;
        }
        return !l1 || status(0x27);
    }
    if (id == 10 || id == 13 || id == 14) {
        /* 0.8.2: no iron-sight switch; the gun is always shown at the
           right (ids 10 and 13 in the rifle component with a zero L1 mask,
           0x359E79; id 14 in its own component 0x35E1E0), so the centre
           crosshair shows whenever the gun is up. */
        return !status(0x2F) && (status(0x2E) || status(0x27));
    }
    if (id == 7) {
        /* Pistol component (0x35966E..0x3596D6): with status 0x5C it needs
           0x2E and neither 0x2F nor 0x27, else ANY(0x2E, 0x27) and neither
           0x2F nor 6. Its view flag freezes while the item is id 6 or 9
           (0x3625D0), so no crosshair then. */
        const uint8_t *item = *(const uint8_t *const *)(actor + 0x708);
        if (item != NULL && (item[0x20] == 6 || item[0x20] == 9)) {
            return 0;
        }
        if (status(0x5C) ? !status(0x2E) || status(0x2F) || status(0x27)
                         : !(status(0x2E) || status(0x27)) || status(0x2F) || status(6)) {
            return 0;
        }
        return !(l1 && status(0x2E) && !status(0x27));
    }
    return 0;
}

static void xh_snapshot(const uint8_t *actor) {
    int fp = g_xh_ok && fp_crosshair(actor);
    /* With the shoulder camera on, only the rig's hit is shown (0.6.9). A
       wall hug or corner peek switches the camera to mode 5; since 0.8.0 the
       wall aim builds the rig there too (without camera writes). With it off
       (F7), the 0.6.1 crosshair on +0x520. */
    g_xh_show = g_xh_ok && g_lock && !status(7) && aim_live() &&
                (g_ots_ok && g_ots_on ? g_rig_on && g_rig_reach : 1);
    g_xh_centre = !g_xh_show && fp;
    g_xh_show |= g_xh_centre;
    if (fp != g_xh_fp_logged) {
        g_xh_fp_logged = fp;
        trace_line("ctx first-person crosshair %s", fp ? "on (weapon at the right)" : "off");
    }
    g_xh_rig = g_rig_on;
    if (g_rig_on) {
        g_xh_pt[0] = g_hit[0];
        g_xh_pt[1] = g_hit[1];
        g_xh_pt[2] = g_hit[2];
    }
    g_xh_seq = g_pad_seq;
}

static void xh_clamp(D3D11_RECT *rc, LONG w, LONG h) {
    if (rc->left < 0) {
        rc->left = 0;
    }
    if (rc->top < 0) {
        rc->top = 0;
    }
    if (rc->right > w) {
        rc->right = w;
    }
    if (rc->bottom > h) {
        rc->bottom = h;
    }
}

/* The viewport mapping PS2 pixels onto the back buffer: the one bound now
   when the bound target is the back buffer, else the last one seen that
   way, else the whole buffer. */
static float g_xh_vpx;
static float g_xh_vpy;
static float g_xh_vpw;
static float g_xh_vph;
static int g_xh_vpsrc;

static void xh_view(ID3D11DeviceContext1 *ctx, ID3D11Texture2D *tex, const D3D11_TEXTURE2D_DESC *desc) {
    ID3D11RenderTargetView *bound = NULL;
    ID3D11Resource *res = NULL;
    D3D11_VIEWPORT vp;
    UINT n = 1;
    ID3D11DeviceContext1_OMGetRenderTargets(ctx, 1, &bound, NULL);
    if (bound != NULL) {
        ID3D11RenderTargetView_GetResource(bound, &res);
        if (res == (ID3D11Resource *)tex) {
            ID3D11DeviceContext1_RSGetViewports(ctx, &n, &vp);
            if (n == 1 && vp.Width >= 64.0f && vp.Height >= 64.0f && vp.TopLeftX >= 0.0f && vp.TopLeftY >= 0.0f &&
                vp.TopLeftX + vp.Width <= (float)desc->Width + 0.5f && vp.TopLeftY + vp.Height <= (float)desc->Height + 0.5f) {
                g_xh_vpx = vp.TopLeftX;
                g_xh_vpy = vp.TopLeftY;
                g_xh_vpw = vp.Width;
                g_xh_vph = vp.Height;
                g_xh_vpsrc = 1;
            }
        }
        if (res != NULL) {
            ID3D11Resource_Release(res);
        }
        ID3D11RenderTargetView_Release(bound);
    }
    if (g_xh_vpsrc == 0 || g_xh_vpw > (float)desc->Width + 0.5f || g_xh_vph > (float)desc->Height + 0.5f) {
        g_xh_vpx = 0.0f;
        g_xh_vpy = 0.0f;
        g_xh_vpw = (float)desc->Width;
        g_xh_vph = (float)desc->Height;
        g_xh_vpsrc = 0;
    }
}

static void xh_paint(ID3D11Device *dev, ID3D11DeviceContext1 *ctx, ID3D11Texture2D *tex, const float *p, int centre) {
    ID3D11RenderTargetView *rtv = NULL;
    D3D11_TEXTURE2D_DESC desc;
    D3D11_RECT rects[4];
    const float *m = g_dg_mtx;
    float cx;
    float cy;
    float cw;
    float sx;
    float sy;
    int32_t view[4];
    int x;
    int y;
    int len;
    int th;
    int gap;
    int pass;
    int i;
    for (i = 0; i < 4; i++) {
        view[i] = *g_dg_view[i];
    }
    if (view[0] <= 0 || view[1] <= 0) {
        return;
    }
    if (centre) {
        /* The first-person aim: the centre of the game's view. */
        sx = 0.5f * (float)view[0] + (float)view[2];
        sy = 0.5f * (float)view[1] + (float)view[3];
    } else {
        cx = p[0] * m[0] + p[1] * m[4] + p[2] * m[8] + m[12];
        cy = p[0] * m[1] + p[1] * m[5] + p[2] * m[9] + m[13];
        cw = p[0] * m[3] + p[1] * m[7] + p[2] * m[11] + m[15];
        if (cw <= 0.001f || fabsf(cx) > cw || fabsf(cy) > cw) {
            return;
        }
        sx = (cx / cw * 0.5f + 0.5f) * (float)view[0] + (float)view[2];
        sy = (cy / cw * 0.5f + 0.5f) * (float)view[1] + (float)view[3];
    }
    ID3D11Texture2D_GetDesc(tex, &desc);
    xh_view(ctx, tex, &desc);
    x = (int)(g_xh_vpx + sx * g_xh_vpw / 512.0f);
    y = (int)(g_xh_vpy + sy * g_xh_vph / 448.0f);
    len = (int)desc.Height / 90;
    th = (int)desc.Height / 540;
    gap = (int)desc.Height / 270;
    if (len < 6) {
        len = 6;
    }
    if (th < 2) {
        th = 2;
    }
    if (gap < 3) {
        gap = 3;
    }
    if (FAILED(ID3D11Device_CreateRenderTargetView(dev, (ID3D11Resource *)tex, NULL, &rtv)) || rtv == NULL) {
        return;
    }
    __try {
        for (pass = 0; pass < 2; pass++) {
            /* Pass 0: a one-pixel dark outline; pass 1: the white arms. */
            static const float kColor[2][4] = {{0.0f, 0.0f, 0.0f, 1.0f}, {1.0f, 1.0f, 1.0f, 1.0f}};
            int o = pass == 0 ? 1 : 0;
            int h = th / 2;
            rects[0].left = x - gap - len - o, rects[0].right = x - gap + o;
            rects[0].top = y - h - o, rects[0].bottom = y - h + th + o;
            rects[1].left = x + gap - o, rects[1].right = x + gap + len + o;
            rects[1].top = y - h - o, rects[1].bottom = y - h + th + o;
            rects[2].left = x - h - o, rects[2].right = x - h + th + o;
            rects[2].top = y - gap - len - o, rects[2].bottom = y - gap + o;
            rects[3].left = x - h - o, rects[3].right = x - h + th + o;
            rects[3].top = y + gap - o, rects[3].bottom = y + gap + len + o;
            for (i = 0; i < 4; i++) {
                xh_clamp(&rects[i], (LONG)desc.Width, (LONG)desc.Height);
            }
            ID3D11DeviceContext1_ClearView(ctx, (ID3D11View *)rtv, kColor[pass], rects, 4);
        }
    } __finally {
        ID3D11RenderTargetView_Release(rtv);
    }
    InterlockedIncrement(&g_pad_n[P_XH]);
    if (GetCurrentThreadId() == g_game_tid && GetTickCount64() - g_xh_log >= 2000) {
        g_xh_log = GetTickCount64();
        trace_line("xh%s bb=%ux%u view=%.0f,%.0f %.0fx%.0f (%s) ps2=%.0f,%.0f px=%d,%d chan=%d,%d,%d,%d",
                   centre ? " centre" : "", desc.Width, desc.Height, g_xh_vpx, g_xh_vpy, g_xh_vpw, g_xh_vph,
                   g_xh_vpsrc ? "bound" : "full", sx, sy, x, y, view[0], view[1], view[2], view[3]);
    }
}

/* Reads the aim point and the matrix now, at present time, so they belong
   to the frame being shown as far as the frame order allows. */
static void xh_draw(uint8_t *r) {
    ID3D11Device *dev = *(ID3D11Device **)(r + RND_DEVICE);
    ID3D11DeviceContext1 *ctx = *(ID3D11DeviceContext1 **)(r + RND_CONTEXT1);
    IDXGISwapChain *sc = *(IDXGISwapChain **)(r + RND_SWAPCHAIN);
    const uint8_t *actor = (const uint8_t *)g_last_actor;
    ID3D11Texture2D *tex = NULL;
    float p[3];
    LONG seq = g_pad_seq;
    if (seq != g_xh_seen) {
        g_xh_seen = seq;
        g_xh_idle = 0;
    } else if (g_xh_idle < 8) {
        g_xh_idle++;
    }
    if (!g_xh_show || g_xh_seq != seq || g_xh_idle >= 2 || *g_pause != 0 || actor == NULL || dev == NULL ||
        ctx == NULL || sc == NULL) {
        return;
    }
    if (g_xh_rig) {
        p[0] = g_xh_pt[0];
        p[1] = g_xh_pt[1];
        p[2] = g_xh_pt[2];
    } else {
        p[0] = ((const float *)(actor + 0x520))[0];
        p[1] = ((const float *)(actor + 0x520))[1];
        p[2] = ((const float *)(actor + 0x520))[2];
    }
    if (FAILED(IDXGISwapChain_GetBuffer(sc, 0, &IID_ID3D11Texture2D, (void **)&tex)) || tex == NULL) {
        return;
    }
    __try {
        xh_paint(dev, ctx, tex, p, g_xh_centre);
    } __finally {
        ID3D11Texture2D_Release(tex);
    }
}

static uint64_t __fastcall pre_present_wrap(void *r, void *rdx, void *r8, void *r9) {
    uint64_t result;
    /* Frame counter for the camera order probe (eyenow/eyeold). */
    InterlockedIncrement(&g_present_seq);
    result = g_pre_present(r, rdx, r8, r9);
    if (g_xh_ok && r != NULL) {
        __try {
            xh_draw((uint8_t *)r);
        } __except (EXCEPTION_EXECUTE_HANDLER) {
            g_xh_ok = 0;
            if (GetCurrentThreadId() == g_game_tid) {
                trace_line("xh fault; crosshair off");
            }
        }
    }
    return result;
}

/* Motion probe (0.6.0, log only): the movement node's sub-state and its
   motion word (node+0x60: low word layer 0, high word the upper body) and
   direction bits (node+0x68), when they change, at most 10 lines a second.
   It is for finding a sideways crouched motion. */
static uintptr_t g_mv_state;
static uint32_t g_mv_motion;
static uint32_t g_mv_bits;
static ULONGLONG g_mv_tick;
static int g_mv_lines;

static void motion_probe(const uint8_t *actor) {
    const uint8_t *node;
    uintptr_t st;
    uint32_t motion;
    uint32_t bits;
    ULONGLONG now;
    __try {
        node = node_find(actor, MOVE_HASH);
        if (node == NULL) {
            return;
        }
        st = *(const uintptr_t *)(node + NODE_STATE);
        motion = *(const uint32_t *)(node + 0x60);
        bits = *(const uint32_t *)(node + 0x68) & 0xFu;
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return;
    }
    if (st == g_mv_state && motion == g_mv_motion && bits == g_mv_bits) {
        return;
    }
    now = GetTickCount64();
    if (now - g_mv_tick >= 1000) {
        g_mv_tick = now;
        g_mv_lines = 0;
    }
    if (g_mv_lines >= 10) {
        return;
    }
    g_mv_lines++;
    g_mv_state = st;
    g_mv_motion = motion;
    g_mv_bits = bits;
    trace_line("mv st=0x%X motion=0x%04X/0x%04X bits=%X lock=%d", rva_in(st, g_text_begin, g_text_end),
               motion & 0xFFFFu, motion >> 16, bits, g_lock);
}

/* Motion request probe (0.8.0, log only): the layer-0 motion the wall-press
   component asks for, to name a crouched sideways motion (the crouched wall
   shuffle) for the crouched strafe. The motion component (hash 0x1B9F95)
   keeps its request at node+0x58, 0x28 bytes per layer: +0x0C archive,
   +0x10 motion number, +0x14 sender hash. */
#define MOTION_HASH 0x1B9F95u
static int g_mreq_ok;
static uint32_t g_mreq_last;
static ULONGLONG g_mreq_tick;
static int g_mreq_lines;

static void motion_req_probe(const uint8_t *actor) {
    uint32_t arch;
    uint32_t owner;
    uint16_t motion;
    uint16_t motion1;
    uint32_t key;
    ULONGLONG now;
    __try {
        const uint8_t *node = node_find(actor, MOTION_HASH);
        if (node == NULL) {
            return;
        }
        arch = *(const uint32_t *)(node + 0x58 + 0x0C);
        motion = *(const uint16_t *)(node + 0x58 + 0x10);
        owner = *(const uint32_t *)(node + 0x58 + 0x14) & 0xFFFFFFu;
        motion1 = *(const uint16_t *)(node + 0x58 + 0x28 + 0x10);
    } __except (EXCEPTION_EXECUTE_HANDLER) {
        return;
    }
    if (owner != WALL_HASH) {
        g_mreq_last = 0;
        return;
    }
    key = (uint32_t)motion | (uint32_t)status(2) << 16;
    if (key == g_mreq_last) {
        return;
    }
    now = GetTickCount64();
    if (now - g_mreq_tick >= 1000) {
        g_mreq_tick = now;
        g_mreq_lines = 0;
    }
    if (g_mreq_lines >= 10) {
        return;
    }
    g_mreq_lines++;
    g_mreq_last = key;
    trace_line("wall motion %u archive=0x%X layer1=%u crouch=%d stick=%u dir=%04X", motion, arch, motion1, status(2),
               *(const uint16_t *)(actor + ACTOR_MAG), *(const uint16_t *)(actor + 0x7E2));
}

static uint64_t __fastcall pad_wrap(void *actor, void *rdx, void *r8, void *r9) {
    uint32_t held = *(uint32_t *)(g_pad + PAD_HELD);
    uint32_t press = *(uint32_t *)(g_pad + PAD_PRESS);
    uint32_t release = *(uint32_t *)(g_pad + PAD_RELEASE);
    uint8_t sq = g_pad[PAD_SQUARE_PRESSURE];
    uint8_t cross = g_pad[PAD_CROSS_PRESSURE];
    uint64_t result;
    InterlockedIncrement(&g_pad_n[P_CALLS]);
    if (actor != NULL) {
        g_game_tid = GetCurrentThreadId();
        g_player_id = *(const uint32_t *)((const uint8_t *)actor + 0x120);
        InterlockedIncrement(&g_pad_seq);
        g_gate_tick = GetTickCount64();
        if (actor != g_last_actor) {
            /* A new player actor (area load, continue): the lock and the
               gates start over; the old actor's L1 is gone with it. */
            if (g_last_actor != NULL && g_lock) {
                trace_line("lock reset: new player actor");
            }
            g_last_actor = actor;
            g_lock = 0;
            g_aim_reason = -1;
            InterlockedExchange(&g_no_target, 0);
            InterlockedExchange(&g_cstrafe, 0);
            g_rig_on = 0;
            g_rig_cam = 0;
            g_ease_w = 0.0f;
            g_cc_stage = 0;
            g_cc_hide = 0;
            g_wall_on = 0;
            g_wall_fail = 0;
        }
        /* CQC first: while the fire key drives Circle, the fire rules must
           not see it as a held Square, or an RMB release while holding an
           enemy never lowers the gun. */
        if (g_split) {
            split_bindings();
            /* This frame's movement sub-state, for cqc_step and the rules
               after it. */
            g_move_state = node_state((const uint8_t *)actor, MOVE_HASH);
            g_move_active = node_state_active((const uint8_t *)actor, MOVE_HASH);
            cqc_step((const uint8_t *)actor);
        }
        pad_step((const uint8_t *)actor);
        if (g_split) {
            split_pre((const uint8_t *)actor);
        }
        {
            /* F6 turns the crouched sidestep motion off and on (0.8.1). */
            int f6 = (GetAsyncKeyState(VK_F6) & 0x8000) != 0 && game_has_focus();
            if (f6 && !g_f6) {
                g_side_on = !g_side_on;
                trace_line("ctx crouched sidestep motion %s (F6)", g_side_on ? "on" : "off");
            }
            g_f6 = f6;
        }
        /* Read by the strafe gate at movement message 0x20, this frame. */
        InterlockedExchange(&g_cstrafe, g_cstrafe_ok && g_split && g_lock && g_move_state != 0 &&
                                            g_move_state == g_state[S_SQUAT] && (held & BTN_R1) == 0);
    }
    result = g_pad_fn(actor, rdx, r8, r9);
    if (g_split && actor != NULL) {
        split_post((uint8_t *)actor);
        if (g_cstrafe && *(uint16_t *)((uint8_t *)actor + ACTOR_MAG) > ROLL_MAG) {
            /* Crouched strafe at the slower (walk) strafe speed. */
            *(uint16_t *)((uint8_t *)actor + ACTOR_MAG) = ROLL_MAG;
        }
        if (g_aim_ok) {
            aim_step((uint8_t *)actor, held);
            pitch_step();
            rig_step((uint8_t *)actor);
        }
        xh_snapshot((const uint8_t *)actor);
        if (g_trace) {
            motion_probe((const uint8_t *)actor);
            if (g_mreq_ok) {
                motion_req_probe((const uint8_t *)actor);
            }
            ctx_step((const uint8_t *)actor, held, press);
            if (g_ctrl_tid != 0 && !g_tid_logged) {
                g_tid_logged = 1;
                trace_line("pad read thread %s the game thread", g_ctrl_tid == g_game_tid ? "is" : "is not");
            }
        }
    }
    /* Other readers of the raw pad see the real input. */
    *(uint32_t *)(g_pad + PAD_HELD) = held;
    *(uint32_t *)(g_pad + PAD_PRESS) = press;
    *(uint32_t *)(g_pad + PAD_RELEASE) = release;
    g_pad[PAD_SQUARE_PRESSURE] = sq;
    g_pad[PAD_CROSS_PRESSURE] = cross;
    if (g_trace && actor != NULL) {
        trace_step((const uint8_t *)actor, held);
    }
    return result;
}

static unsigned __stdcall watch_walk(void *unused) {
    int first_pad = 1;
    (void)unused;
    for (;;) {
        LONG calls;
        Sleep(2000);
        trace_flush();
        {
            LONG n[P_COUNT];
            LONG events = 0;
            int i;
            for (i = 0; i < P_COUNT; i++) {
                n[i] = InterlockedExchange(&g_pad_n[i], 0);
                if (i >= P_FIRE) {
                    events += n[i];
                }
            }
            if (n[P_CALLS] > 0 && (events > 0 || first_pad)) {
                char line[512];
                int used;
                SYSTEMTIME st;
                GetLocalTime(&st);
                used = snprintf(line, sizeof line, "fpvmove: %02u:%02u:%02u pad", st.wHour, st.wMinute, st.wSecond);
                for (i = 0; i < P_COUNT && used > 0 && (size_t)used < sizeof line; i++) {
                    used += snprintf(line + used, sizeof line - (size_t)used, " %s=%ld", kPadName[i], n[i]);
                }
                if (used > 0 && (size_t)used + 3 < sizeof line) {
                    memcpy(line + used, "\r\n", 3);
                    append_log(line);
                }
                first_pad = 0;
            }
        }
        calls = InterlockedExchange(&g_calls, 0);
        if (calls > 0) {
            char line[400];
            int used;
            int i;
            SYSTEMTIME st;
            GetLocalTime(&st);
            used = snprintf(line, sizeof line, "fpvmove: %02u:%02u:%02u walk calls=%ld", st.wHour, st.wMinute, st.wSecond, calls);
            for (i = 0; i < R_COUNT; i++) {
                LONG n = InterlockedExchange(&g_reason[i], 0);
                if (n > 0 && used > 0 && (size_t)used < sizeof line) {
                    used += snprintf(line + used, sizeof line - (size_t)used, " %s=%ld", kReason[i], n);
                }
            }
            if (used > 0 && (size_t)used < sizeof line) {
                snprintf(line + used, sizeof line - (size_t)used, " sq=%ld b2E=%ld b27=%ld moved=%ld",
                         InterlockedExchange(&g_sq, 0), InterlockedExchange(&g_b2e, 0), InterlockedExchange(&g_b27, 0),
                         InterlockedExchange(&g_moved, 0));
            }
            append_log(line);
            append_log("\r\n");
        }
    }
    return 0;
}

static void *alloc_near(void *site, size_t bytes) {
    SYSTEM_INFO si;
    uintptr_t gran;
    uintptr_t min_a;
    uintptr_t max_a;
    uintptr_t origin;
    uintptr_t step;
    const uintptr_t span = 0x7FFF0000u;
    GetSystemInfo(&si);
    gran = (uintptr_t)si.dwAllocationGranularity;
    if (gran == 0) {
        gran = 0x10000;
    }
    min_a = (uintptr_t)si.lpMinimumApplicationAddress;
    max_a = (uintptr_t)si.lpMaximumApplicationAddress;
    origin = (uintptr_t)site & ~(gran - 1);
    for (step = 0; step <= span; step += gran) {
        uintptr_t candidates[2];
        int n = 0;
        int i;
        if (step == 0) {
            candidates[n++] = origin;
        } else {
            if (origin >= min_a + step) {
                candidates[n++] = origin - step;
            }
            if (origin <= max_a - step) {
                candidates[n++] = origin + step;
            }
        }
        for (i = 0; i < n; i++) {
            void *page;
            if (candidates[i] < min_a || candidates[i] > max_a) {
                continue;
            }
            page = VirtualAlloc((void *)candidates[i], bytes, MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);
            if (page != NULL) {
                return page;
            }
        }
    }
    return NULL;
}

/* Returns 1 when the walk call now goes through walk_wrap. */
static int hook_walk(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    int count;
    uint8_t *site;
    uint8_t *walk;
    uint8_t *page;
    uint64_t wrap = (uint64_t)(uintptr_t)walk_wrap;
    int64_t rel64;
    int32_t rel;
    count = scan_exec(base, nt, kBits, &hit);
    if (count != 1 || hit == NULL) {
        log_line("fpvmove: status bitset matches=%d; walk not traced", count);
        return 0;
    }
    g_bits = (const uint32_t *)rip_target(hit, kBitsDisp, kBitsNext);
    if ((uintptr_t)g_bits < begin || (uintptr_t)g_bits >= end) {
        log_line("fpvmove: status bitset outside the game; walk not traced");
        return 0;
    }
    count = scan_exec(base, nt, kWalkCall, &hit);
    if (count != 1 || hit == NULL) {
        log_line("fpvmove: walk call matches=%d; walk not traced", count);
        return 0;
    }
    site = (uint8_t *)hit + kWalkCallAt;
    walk = rip_target(site, 1, 5);
    if (site[0] != 0xE8 || (uintptr_t)walk < begin || (uintptr_t)walk >= end ||
        memcmp(walk, kWalkHead, sizeof kWalkHead) != 0 || memcmp(walk + kWalkBodyAt, kWalkBody, sizeof kWalkBody) != 0) {
        log_line("fpvmove: walk callee not recognized; walk not traced");
        return 0;
    }
    page = (uint8_t *)alloc_near(site, 4096);
    if (page == NULL) {
        log_line("fpvmove: no page near the walk call; walk not traced");
        return 0;
    }
    rel64 = (int64_t)(page - (site + 5));
    if (rel64 > INT32_MAX || rel64 < INT32_MIN) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: walk thunk out of rel32 range; walk not traced");
        return 0;
    }
    /* jmp qword ptr [rip+0], then the absolute address of walk_wrap. */
    memset(page, 0xCC, 4096);
    page[0] = 0xFF;
    page[1] = 0x25;
    memset(page + 2, 0, 4);
    memcpy(page + 6, &wrap, sizeof wrap);
    g_walk = (walk_fn)walk;
    rel = (int32_t)rel64;
    if (!write_code(site + 1, (const uint8_t *)&rel, sizeof rel)) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: walk call not writable; walk not traced");
        return 0;
    }
    log_line("fpvmove: walk call at rva 0x%X traced, walk rva 0x%X", (unsigned)(site - base), (unsigned)(walk - base));
    return 1;
}

/* Returns 1 when the pad component registration now points at pad_wrap.
   Needs g_bits from hook_walk. Registration runs when the player is built,
   after this plugin loads, so the game stores the wrapper. */
static int hook_pad(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    const uint8_t *fn_hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    uint64_t wrap = (uint64_t)(uintptr_t)pad_wrap;
    uint8_t *site;
    uint8_t *fn;
    uint8_t *pad;
    uint8_t *page;
    int64_t rel64;
    int32_t rel;
    int count;
    if (g_bits == NULL) {
        log_line("fpvmove: pad hook needs the status bitset; pad unchanged");
        return 0;
    }
    count = scan_exec(base, nt, kPadFn, &fn_hit);
    if (count != 1 || fn_hit == NULL) {
        log_line("fpvmove: pad component matches=%d; pad unchanged", count);
        return 0;
    }
    fn = (uint8_t *)fn_hit;
    pad = rip_target(fn, 0x1B, 0x1F);
    if (rip_target(fn, 0x21, 0x25) != pad + 0x20) {
        log_line("fpvmove: pad component loads do not agree; pad unchanged");
        return 0;
    }
    g_filter_on = (const volatile int32_t *)rip_target(fn, 0x29, 0x2D);
    count = scan_exec(base, nt, kAim, &hit);
    if (count != 1 || hit == NULL) {
        log_line("fpvmove: aimingState matches=%d; pad unchanged", count);
        return 0;
    }
    g_aim = (const volatile int32_t *)rip_target(hit, 2, 6);
    if ((uintptr_t)pad < begin || (uintptr_t)pad >= end || (uintptr_t)g_filter_on < begin ||
        (uintptr_t)g_filter_on >= end || (uintptr_t)g_aim < begin || (uintptr_t)g_aim >= end) {
        log_line("fpvmove: pad globals outside the game; pad unchanged");
        return 0;
    }
    count = scan_exec(base, nt, kPadReg, &hit);
    if (count != 1 || hit == NULL) {
        log_line("fpvmove: pad registration matches=%d; pad unchanged", count);
        return 0;
    }
    site = (uint8_t *)hit;
    if (rip_target(site, kPadRegDisp, kPadRegNext) != fn) {
        log_line("fpvmove: pad registration does not point at the component; pad unchanged");
        return 0;
    }
    page = (uint8_t *)alloc_near(site, 4096);
    if (page == NULL) {
        log_line("fpvmove: no page near the pad registration; pad unchanged");
        return 0;
    }
    rel64 = (int64_t)(page - (site + kPadRegNext));
    if (rel64 > INT32_MAX || rel64 < INT32_MIN) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: pad thunk out of rel32 range; pad unchanged");
        return 0;
    }
    memset(page, 0xCC, 4096);
    page[0] = 0xFF;
    page[1] = 0x25;
    memset(page + 2, 0, 4);
    memcpy(page + 6, &wrap, sizeof wrap);
    g_pad = pad;
    g_pad_fn = (pad_fn)fn;
    rel = (int32_t)rel64;
    if (!write_code(site + kPadRegDisp, (const uint8_t *)&rel, sizeof rel)) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: pad registration not writable; pad unchanged");
        return 0;
    }
    log_line("fpvmove: pad component rva 0x%X wrapped at registration rva 0x%X, raw pad rva 0x%X, aim rva 0x%X",
             (unsigned)(fn - base), (unsigned)(site - base), (unsigned)(pad - base), (unsigned)((const uint8_t *)g_aim - base));
    return 1;
}

/* Returns 1 when the tree layout the tracer reads is confirmed: the player
   update passes [actor+0x58] to the walker, the walker reads next, flags,
   child, and handler at the expected offsets, and the hash lookup reads
   +0x24. Needs the pad wrapper, which supplies the actor each frame. */
static int trace_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    IMAGE_SECTION_HEADER *section = IMAGE_FIRST_SECTION(nt);
    const uint8_t *update = NULL;
    const uint8_t *walker = NULL;
    const uint8_t *lookup = NULL;
    unsigned i;
    int count = scan_exec(base, nt, kTreeWalk, &update);
    if (count != 1 || update == NULL) {
        log_line("fpvmove: tree walk matches=%d; tracer off", count);
        return 0;
    }
    count = scan_exec(base, nt, kWalker, &walker);
    if (count != 1 || walker == NULL) {
        log_line("fpvmove: walker matches=%d; tracer off", count);
        return 0;
    }
    if (rip_target(update + kTreeWalkCall, 1, 5) != walker - kWalkerAt) {
        log_line("fpvmove: player update does not call the walker; tracer off");
        return 0;
    }
    count = scan_exec(base, nt, kHashLookup, &lookup);
    if (count != 1 || lookup == NULL) {
        log_line("fpvmove: hash lookup matches=%d; tracer off", count);
        return 0;
    }
    g_base = (uintptr_t)base;
    g_image_end = g_base + nt->OptionalHeader.SizeOfImage;
    for (i = 0; i < nt->FileHeader.NumberOfSections; i++) {
        if (memcmp(section[i].Name, ".text", 6) == 0) {
            g_text_begin = g_base + section[i].VirtualAddress;
            g_text_end = g_text_begin + section[i].Misc.VirtualSize;
        }
    }
    if (g_text_begin == 0) {
        log_line("fpvmove: no .text section; tracer off");
        return 0;
    }
    g_trace = 1;
    log_line("fpvmove: tracer armed, walker rva 0x%X, F9 marks", (unsigned)(walker - kWalkerAt - base));
    return 1;
}

static uintptr_t rip_in(const uint8_t *at, size_t disp_at, size_t next, uintptr_t begin, uintptr_t end) {
    uintptr_t target = (uintptr_t)rip_target(at, disp_at, next);
    return target >= begin && target < end ? target : 0;
}

/* Returns 1 when every movement sub-state, the roll gate, the key array,
   and the binding tables resolve. Needs the tracer's tree checks. */
static int split_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *init = NULL;
    const uint8_t *move = NULL;
    const uint8_t *gate = NULL;
    const uint8_t *s2p = NULL;
    const uint8_t *p2s = NULL;
    const uint8_t *crawl = NULL;
    const uint8_t *rollprone = NULL;
    const uint8_t *crawl2 = NULL;
    const uint8_t *lookup = NULL;
    const uint8_t *layout_fn;
    const uint8_t *bind_fn;
    const uint8_t *hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    uintptr_t enter;
    uintptr_t back;
    uintptr_t layout;
    uintptr_t prone2;
    uint8_t pat[MAX_PAT];
    char mask[MAX_PAT + 1];
    size_t bind_len = parse_pattern(kBindFn, pat, mask);
    int i;
    if (!g_trace) {
        log_line("fpvmove: split needs the tree checks; split off");
        return 0;
    }
    if (scan_exec(base, nt, kMoveInit, &init) != 1 || scan_exec(base, nt, kMoveSite, &move) != 1 ||
        scan_exec(base, nt, kRollGate, &gate) != 1 || scan_exec(base, nt, kS2P, &s2p) != 1 ||
        scan_exec(base, nt, kP2S, &p2s) != 1 || scan_exec(base, nt, kKeyLookup, &lookup) != 1 ||
        scan_exec(base, nt, kCrawl, &crawl) != 1 || scan_exec(base, nt, kRollProne, &rollprone) != 1 ||
        scan_exec(base, nt, kCrawl2, &crawl2) != 1) {
        log_line("fpvmove: split pattern missing or repeated; split off");
        return 0;
    }
    g_state[S_PRONE] = rip_in(init, kInitProne, kInitProne + 4, g_text_begin, g_text_end);
    g_state[S_SQUAT] = rip_in(init, kInitSquat, kInitSquat + 4, g_text_begin, g_text_end);
    g_state[S_STILL] = rip_in(init, kInitStill, kInitStill + 4, g_text_begin, g_text_end);
    g_state[S_MOVE] = rip_in(move, kMoveSiteDisp, kMoveSiteDisp + 4, g_text_begin, g_text_end);
    g_state[S_S2P] = rip_in(s2p, kS2PDisp, kS2PDisp + 4, g_text_begin, g_text_end);
    g_state[S_P2S] = rip_in(p2s, kP2SDisp, kP2SDisp + 4, g_text_begin, g_text_end);
    g_state[S_CRAWL] = rip_in(crawl, kCrawlDisp, kCrawlDisp + 4, g_text_begin, g_text_end);
    g_state[S_ROLLPRONE] = rip_in(rollprone, kRollProneDisp, kRollProneDisp + 4, g_text_begin, g_text_end);
    g_state[S_CRAWL2] = rip_in(crawl2, kCrawlDisp, kCrawlDisp + 4, g_text_begin, g_text_end);
    prone2 = rip_in(p2s, kP2SProne, kP2SProne + 4, g_text_begin, g_text_end);
    for (i = 0; i < S_COUNT; i++) {
        if (g_state[i] == 0) {
            log_line("fpvmove: %s sub-state outside .text; split off", kStateName[i]);
            return 0;
        }
    }
    /* Prone from two sites must agree, and the roll gate must sit inside the
       move sub-state, which ends where the stand sub-state begins. */
    if (prone2 != g_state[S_PRONE] || (uintptr_t)gate <= g_state[S_MOVE] || (uintptr_t)gate >= g_state[S_STILL]) {
        log_line("fpvmove: movement sub-states do not agree; split off");
        return 0;
    }
    /* The roll sub-state: lea rax, [roll] at +0x6A of the roll gate. */
    if (gate[kRollLea] != 0x48 || gate[kRollLea + 1] != 0x8D || gate[kRollLea + 2] != 0x05) {
        log_line("fpvmove: roll sub-state load not found; split off");
        return 0;
    }
    g_roll_state = rip_in(gate, kRollLea + 3, kRollLea + 7, g_text_begin, g_text_end);
    if (g_roll_state == 0) {
        log_line("fpvmove: roll sub-state outside .text; split off");
        return 0;
    }
    /* The roll-to-prone store sits inside the roll sub-state, which ends
       where the crouch sub-state begins. */
    if ((uintptr_t)rollprone <= g_roll_state || (uintptr_t)rollprone >= g_state[S_SQUAT]) {
        log_line("fpvmove: roll-to-prone store outside the roll sub-state; split off");
        return 0;
    }
    log_line("fpvmove: crawl 0x%X and 0x%X, roll to prone 0x%X", (unsigned)(g_state[S_CRAWL] - begin),
             (unsigned)(g_state[S_CRAWL2] - begin),
             (unsigned)(g_state[S_ROLLPRONE] - begin));
    enter = rip_in(lookup, kKeyEnter, kKeyEnter + 4, begin, end);
    back = rip_in(lookup, kKeyBack, kKeyBack + 4, begin, end);
    hit = lookup + kKeyCalls;
    if (enter == 0 || back == 0 || enter - 0x0D * 4 != back - 0x08 * 4 || hit[0] != 0xE8 || hit[0xC] != 0xE8 ||
        memcmp(hit + 5, "\x8B\xC8\x45\x33\xC0\x8B\xD3", 7) != 0) {
        log_line("fpvmove: key lookup not recognized; split off");
        return 0;
    }
    g_keys = (const volatile uint32_t *)(enter - 0x0D * 4);
    layout_fn = rip_target(hit, 1, 5);
    bind_fn = rip_target(hit + 0xC, 1, 5);
    if ((uintptr_t)layout_fn < g_text_begin || (uintptr_t)layout_fn >= g_text_end || layout_fn[0] != 0x8B ||
        layout_fn[1] != 0x05 || layout_fn[6] != 0xC3 || (uintptr_t)bind_fn < g_text_begin ||
        (uintptr_t)bind_fn + bind_len > g_text_end || match_pattern(bind_fn, bind_len, kBindFn, &hit) != 1 ||
        rip_target(bind_fn, kBindBase, kBindBase + 4) != base) {
        log_line("fpvmove: key binding getters not recognized; split off");
        return 0;
    }
    layout = rip_in(layout_fn, 2, 6, begin, end);
    if (layout == 0) {
        log_line("fpvmove: key layout outside the game; split off");
        return 0;
    }
    g_layout = (const volatile int32_t *)layout;
    {
        const size_t at[3] = {kBindT0, kBindT1, kBindCustom};
        for (i = 0; i < 3; i++) {
            uint32_t rva;
            memcpy(&rva, bind_fn + at[i], sizeof rva);
            if (rva == 0 || (uintptr_t)rva + KEY_IDS * 16 > nt->OptionalHeader.SizeOfImage) {
                log_line("fpvmove: key table %d outside the game; split off", i);
                return 0;
            }
            g_table[i] = (const uint32_t *)(base + rva);
        }
    }
    g_split = 1;
    log_line("fpvmove: split armed: roll 0x%X stand 0x%X move 0x%X crouch 0x%X toprone 0x%X prone 0x%X fromprone 0x%X, roll gate "
             "0x%X, keys 0x%X, layout 0x%X, tables 0x%X 0x%X 0x%X",
             (unsigned)(g_roll_state - begin), (unsigned)(g_state[S_STILL] - begin), (unsigned)(g_state[S_MOVE] - begin), (unsigned)(g_state[S_SQUAT] - begin),
             (unsigned)(g_state[S_S2P] - begin), (unsigned)(g_state[S_PRONE] - begin), (unsigned)(g_state[S_P2S] - begin),
             (unsigned)((uintptr_t)gate - begin), (unsigned)((uintptr_t)g_keys - begin), (unsigned)(layout - begin),
             (unsigned)((const uint8_t *)g_table[0] - base), (unsigned)((const uint8_t *)g_table[1] - base),
             (unsigned)((const uint8_t *)g_table[2] - base));
    return 1;
}

/* Writes a jmp [rip+0] stub to a page near the call, then points the call's
   rel32 at it. Returns 1 when the call now goes to target. */
static int redirect_call(uint8_t *site, void *target, const char *label) {
    uint8_t *page;
    uint64_t abs = (uint64_t)(uintptr_t)target;
    int64_t rel64;
    int32_t rel;
    page = (uint8_t *)alloc_near(site, 4096);
    if (page == NULL) {
        log_line("fpvmove: no page near the %s call; unchanged", label);
        return 0;
    }
    rel64 = (int64_t)(page - (site + 5));
    if (rel64 > INT32_MAX || rel64 < INT32_MIN) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: %s thunk out of rel32 range; unchanged", label);
        return 0;
    }
    memset(page, 0xCC, 4096);
    page[0] = 0xFF;
    page[1] = 0x25;
    memset(page + 2, 0, 4);
    memcpy(page + 6, &abs, sizeof abs);
    rel = (int32_t)rel64;
    if (!write_code(site + 1, (const uint8_t *)&rel, sizeof rel)) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: %s call not writable; unchanged", label);
        return 0;
    }
    return 1;
}

/* As redirect_call, but the stub first copies rbx into rdx (mov rdx, rbx),
   so the target gets the caller's rbx as its second argument (0.8.1: the
   hold-up states keep the soldier in rbx; rdx is volatile at a call). */
static int redirect_call_rbx(uint8_t *site, void *target, const char *label) {
    uint8_t *page;
    uint64_t abs = (uint64_t)(uintptr_t)target;
    int64_t rel64;
    int32_t rel;
    page = (uint8_t *)alloc_near(site, 4096);
    if (page == NULL) {
        log_line("fpvmove: no page near the %s call; unchanged", label);
        return 0;
    }
    rel64 = (int64_t)(page - (site + 5));
    if (rel64 > INT32_MAX || rel64 < INT32_MIN) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: %s thunk out of rel32 range; unchanged", label);
        return 0;
    }
    memset(page, 0xCC, 4096);
    page[0] = 0x48;
    page[1] = 0x8B;
    page[2] = 0xD3;
    page[3] = 0xFF;
    page[4] = 0x25;
    memset(page + 5, 0, 4);
    memcpy(page + 9, &abs, sizeof abs);
    rel = (int32_t)rel64;
    if (!write_code(site + 1, (const uint8_t *)&rel, sizeof rel)) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: %s call not writable; unchanged", label);
        return 0;
    }
    return 1;
}

/* Needs g_bits and the tracer's tree layout (node_find). Resolves the camera
   global and checks every hard-coded offset against the code that uses it. */
static int aim_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt, const uint8_t *test_fn_at) {
    const uint8_t *hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    const char *const pats[] = {kCamYaw, kCamTarget, kCamMode, kCamSub, kFaceWpn};
    const char *const names[] = {"camera yaw", "camera target", "camera mode", "camera sub-mode", "facing"};
    uint8_t *const *cam;
    int count;
    size_t i;
    for (i = 0; i < sizeof pats / sizeof pats[0]; i++) {
        count = scan_exec(base, nt, pats[i], &hit);
        if (count != 1) {
            log_line("fpvmove: %s matches=%d; mouse aim off", names[i], count);
            return 0;
        }
    }
    count = scan_exec(base, nt, kNoTarget, &hit);
    if (count != 1 || hit[kNoTargetCall] != 0xE8 || rip_target(hit + kNoTargetCall, 1, 5) != test_fn_at) {
        log_line("fpvmove: auto-target gate matches=%d or does not call TEST; mouse aim off", count);
        return 0;
    }
    g_test = (test_fn)test_fn_at;
    if (redirect_call((uint8_t *)hit + kNoTargetCall, (void *)notarget_gate, "auto-target")) {
        g_notarget_ok = 1;
        log_line("fpvmove: auto-target gate at rva 0x%X", (unsigned)(hit + kNoTargetCall - base));
    }
    count = scan_exec(base, nt, kCamObj, &hit);
    if (count != 1) {
        log_line("fpvmove: camera object matches=%d; mouse aim off", count);
        return 0;
    }
    cam = (uint8_t *const *)rip_target(hit, kCamObjDisp, kCamObjNext);
    if ((uintptr_t)cam < begin || (uintptr_t)cam >= end) {
        log_line("fpvmove: camera global outside the game; mouse aim off");
        return 0;
    }
    g_cam_pp = (uint8_t *const volatile *)cam;
    g_aim_ok = 1;
    log_line("fpvmove: mouse aim armed, camera global rva 0x%X", (unsigned)((uintptr_t)cam - begin));
    return 1;
}

static int cstrafe_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt, const uint8_t *test_fn_at) {
    const uint8_t *hit = NULL;
    uint8_t *site;
    int count = scan_exec(base, nt, kStrafeStand, &hit);
    if (count != 1) {
        log_line("fpvmove: strafe gate matches=%d; crouched strafe off", count);
        return 0;
    }
    site = (uint8_t *)hit + kStrafeStandCall;
    if (site[0] != 0xE8 || rip_target(site, 1, 5) != test_fn_at) {
        log_line("fpvmove: strafe gate does not call TEST; crouched strafe off");
        return 0;
    }
    g_test = (test_fn)test_fn_at;
    if (!redirect_call(site, (void *)cstrafe_gate, "strafe gate")) {
        return 0;
    }
    g_cstrafe_ok = 1;
    log_line("fpvmove: crouched strafe gate at rva 0x%X", (unsigned)(site - base));
    {
        const uint8_t *hit2 = NULL;
        uint8_t *send;
        if (scan_exec(base, nt, kMotionSend, &hit2) != 1 || hit2 == NULL) {
            log_line("fpvmove: motion send not found; crouched sidestep off");
        } else {
            site = (uint8_t *)hit2 + kMotionSendCall;
            send = site[0] == 0xE8 ? rip_target(site, 1, 5) : NULL;
            if (send == NULL || memcmp(send, kSendHead, sizeof kSendHead) != 0) {
                log_line("fpvmove: motion send call not recognized; crouched sidestep off");
            } else {
                g_send = (send_fn)send;
                if (redirect_call(site, (void *)send_wrap, "motion send")) {
                    log_line("fpvmove: crouched sidestep at rva 0x%X, motions %d/%d, F6 toggles", (unsigned)(site - base),
                             SIDE_MOTION_8, SIDE_MOTION_4);
                }
            }
        }
    }
    return 1;
}

static int ctrl_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    const uint8_t *fn_hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    uint8_t *site;
    uint32_t table_rva;
    int count = scan_exec(base, nt, kPadCfg, &hit);
    if (count != 1 || rip_target(hit, kPadCfgBaseDisp, kPadCfgBaseNext) != base) {
        log_line("fpvmove: button config matches=%d; pad layout native", count);
        return 0;
    }
    memcpy(&table_rva, hit + kPadCfgTable, sizeof table_rva);
    g_cfg_flag = (const volatile int32_t *)rip_target(hit, kPadCfgFlagDisp, kPadCfgFlagNext);
    g_cfg_table = (const volatile uint32_t *)(base + table_rva);
    if ((uintptr_t)g_cfg_flag < begin || (uintptr_t)g_cfg_flag >= end || table_rva >= end - begin - 40) {
        log_line("fpvmove: button config outside the game; pad layout native");
        return 0;
    }
    count = scan_exec(base, nt, kCtrlFn, &fn_hit);
    if (count != 1) {
        log_line("fpvmove: controller read matches=%d; pad layout native", count);
        return 0;
    }
    count = scan_exec(base, nt, kCtrlCall, &hit);
    site = (uint8_t *)hit;
    if (count != 1 || rip_target(site, 1, 5) != fn_hit) {
        log_line("fpvmove: controller read call matches=%d or targets another function; pad layout native", count);
        return 0;
    }
    g_ctrl_fn = (ctrl_fn)fn_hit;
    if (!redirect_call(site, (void *)ctrl_wrap, "controller read")) {
        return 0;
    }
    log_line("fpvmove: pad layout hook at rva 0x%X, config table rva 0x%X", (unsigned)(site - base), table_rva);
    return 1;
}

/* Replaces the first 5 bytes of fn, which must equal head (so another hook
   there stops this one), with a jmp to target. Returns a trampoline that
   runs those bytes and continues at fn + 5, or NULL. */
static void *detour_entry(uint8_t *fn, const uint8_t *head, size_t len, void *target, void *volatile *tramp_out,
                          const char *label) {
    uint8_t *page;
    uint64_t back = (uint64_t)(uintptr_t)(fn + len);
    uint64_t abs = (uint64_t)(uintptr_t)target;
    int64_t rel64;
    int32_t rel;
    uint8_t jmp[16];
    /* len must cover whole, position-independent instructions (checked by
       the caller's head bytes), at least a 5-byte jmp and at most 16. */
    if (len < 5 || len > sizeof jmp || memcmp(fn, head, len) != 0) {
        log_line("fpvmove: %s entry bytes differ (hooked by another plugin?); unchanged", label);
        return NULL;
    }
    page = (uint8_t *)alloc_near(fn, 4096);
    if (page == NULL) {
        log_line("fpvmove: no page near the %s entry; unchanged", label);
        return NULL;
    }
    rel64 = (int64_t)((page + 0x40) - (fn + 5));
    if (rel64 > INT32_MAX || rel64 < INT32_MIN) {
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: %s stub out of rel32 range; unchanged", label);
        return NULL;
    }
    memset(page, 0xCC, 4096);
    /* +0x00 trampoline: the moved bytes, then jmp [rip+0] to fn + len. */
    memcpy(page, head, len);
    page[len] = 0xFF;
    page[len + 1] = 0x25;
    memset(page + len + 2, 0, 4);
    memcpy(page + len + 6, &back, sizeof back);
    /* +0x40 stub: jmp [rip+0] to the target. */
    page[0x40] = 0xFF;
    page[0x41] = 0x25;
    memset(page + 0x42, 0, 4);
    memcpy(page + 0x46, &abs, sizeof abs);
    jmp[0] = 0xE9;
    rel = (int32_t)rel64;
    memcpy(jmp + 1, &rel, sizeof rel);
    memset(jmp + 5, 0x90, sizeof jmp - 5);
    /* The wrapper calls the trampoline, so it must be set before the jump
       goes live. */
    *tramp_out = page;
    if (!write_code(fn, jmp, len)) {
        *tramp_out = NULL;
        VirtualFree(page, 0, MEM_RELEASE);
        log_line("fpvmove: %s entry not writable; unchanged", label);
        return NULL;
    }
    return page;
}

/* Needs mouse aim (the camera yaw write) and the auto-target gate. */
static int pitch_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    uint8_t *fn;
    uint8_t *zoom;
    uint8_t *site;
    const uint8_t *mouse;
    int count = scan_exec(base, nt, kBuild, &hit);
    if (count != 1) {
        log_line("fpvmove: aim point builder matches=%d; manual pitch off", count);
        return 0;
    }
    fn = (uint8_t *)hit + 1;
    count = scan_exec(base, nt, kFireCall, &hit);
    site = (uint8_t *)hit + kFireCallAt;
    if (count != 1 || site[0] != 0xE8 || rip_target(site, 1, 5) != fn) {
        log_line("fpvmove: gun fire call matches=%d or targets another function; manual pitch off", count);
        return 0;
    }
    g_gun_lo = (uintptr_t)(site + 5) - GUN_SITES_BEFORE;
    g_gun_hi = (uintptr_t)(site + 5) + GUN_SITES_AFTER;
    count = scan_exec(base, nt, kMouseDelta, &hit);
    if (count != 1) {
        log_line("fpvmove: mouse delta matches=%d; manual pitch off", count);
        return 0;
    }
    mouse = rip_target(hit, kMouseDeltaDisp, kMouseDeltaNext);
    if ((uintptr_t)mouse < begin || (uintptr_t)mouse >= end) {
        log_line("fpvmove: mouse delta outside the game; manual pitch off");
        return 0;
    }
    count = scan_exec(base, nt, kZoom, &hit);
    if (count != 1) {
        log_line("fpvmove: camera zoom matches=%d; manual pitch off", count);
        return 0;
    }
    zoom = (uint8_t *)hit;
    /* dx then dy, two int32 words. */
    g_mouse_dy = (const volatile int32_t *)(mouse + 4);
    if (detour_entry(fn, kBuildHead, sizeof kBuildHead, (void *)builder_wrap, (void *volatile *)&g_build_tramp,
                     "aim point builder") == NULL) {
        return 0;
    }
    if (detour_entry(zoom, kZoomHead, sizeof kZoomHead, (void *)zoom_wrap, (void *volatile *)&g_zoom_tramp,
                     "camera zoom") == NULL) {
        /* The builder hook still gives the yaw without the ease. */
        log_line("fpvmove: builder rva 0x%X hooked; manual pitch off (camera zoom not hooked)", (unsigned)(fn - base));
        return 0;
    }
    g_pitch_ok = 1;
    count = scan_exec(base, nt, kCamWall, &hit);
    if (count != 1) {
        log_line("fpvmove: camera wall step matches=%d; shoulder camera off", count);
    } else {
        const uint8_t *stance = NULL;
        site = (uint8_t *)hit + 1;
        g_stance_ok = scan_exec(base, nt, kCamStance, &stance) == 1;
        if (detour_entry(site, kCamWallHead, sizeof kCamWallHead, (void *)wall_wrap, (void *volatile *)&g_wall_tramp,
                         "camera wall step") != NULL) {
            g_ots_ok = 1;
            log_line("fpvmove: shoulder camera armed at the wall step rva 0x%X, stance offset %s", (unsigned)(site - base),
                     g_stance_ok ? "read" : "not confirmed");
        }
    }
    if (g_ots_ok) {
        const uint8_t *turn = NULL;
        count = scan_exec(base, nt, kCamTurn, &turn);
        g_mouse_dx = (const volatile int32_t *)mouse;
        if (count != 1) {
            log_line("fpvmove: camera yaw step matches=%d; stick yaw native", count);
        } else if (detour_entry((uint8_t *)turn + 1, kCamTurnHead, sizeof kCamTurnHead, (void *)turn_wrap,
                                (void *volatile *)&g_turn_tramp, "camera yaw step") != NULL) {
            log_line("fpvmove: stick yaw while aiming at %.2f, yaw step rva 0x%X", TURN_STICK,
                     (unsigned)(turn + 1 - base));
        }
    }
    log_line("fpvmove: manual pitch armed, builder rva 0x%X, zoom rva 0x%X, gun calls 0x%X-0x%X, mouse rva 0x%X",
             (unsigned)(fn - base), (unsigned)(zoom - base), (unsigned)(g_gun_lo - begin), (unsigned)(g_gun_hi - begin),
             (unsigned)((uintptr_t)mouse - begin));
    return 1;
}

/* 0.6.4. The line check for the rig, then the six untargeted spawner calls
   of the player's bullet callbacks, each to its wrapper. Needs the rig
   (the shoulder camera). */
static uint8_t *call_target(const uint8_t *site, uintptr_t begin, uintptr_t end) {
    uint8_t *t;
    if (site[0] != 0xE8) {
        return NULL;
    }
    t = rip_target(site, 1, 5);
    return (uintptr_t)t >= begin && (uintptr_t)t < end ? t : NULL;
}

/* The call at the end of a pattern (its last byte is the E8), or NULL. */
static uint8_t *pattern_call(uint8_t *base, IMAGE_NT_HEADERS64 *nt, const char *text) {
    const uint8_t *hit = NULL;
    uint8_t pat[MAX_PAT];
    char mask[MAX_PAT + 1];
    size_t len = parse_pattern(text, pat, mask);
    if (scan_exec(base, nt, text, &hit) != 1 || hit == NULL) {
        return NULL;
    }
    return (uint8_t *)hit + len - 1;
}

static void shot_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    const char *const un[6] = {kShotA1, kShotA2, kShotA3, kShotA4, kShotB, kShotC};
    void *const wrap[6] = {(void *)shot_a_wrap, (void *)shot_a_wrap, (void *)shot_a_wrap, (void *)shot_a_wrap,
                           (void *)shot_b_wrap, (void *)shot_c_wrap};
    uint8_t *site[6];
    uint8_t *fn[6];
    uint8_t *twin[3];
    uint8_t *line;
    uint8_t *idx;
    uint8_t *pt;
    int redirected = 0;
    int i;
    if (!g_ots_ok) {
        log_line("fpvmove: shoulder camera off; shot rig off");
        return;
    }
    if (scan_exec(base, nt, kShotRay, &hit) != 1) {
        log_line("fpvmove: shot ray not found; shot rig off");
        return;
    }
    line = call_target(hit + kShotRayLine, begin, end);
    idx = call_target(hit + kShotRayIdx, begin, end);
    pt = call_target(hit + kShotRayPt, begin, end);
    if (line == NULL || idx == NULL || pt == NULL || memcmp(line, kLineHead, sizeof kLineHead) != 0) {
        log_line("fpvmove: line check not recognized; shot rig off");
        return;
    }
    g_line = (line_fn)line;
    g_hit_idx = (hitidx_fn)idx;
    g_hit_pt = (hitpt_fn)pt;
    /* The rig uses the line check even if a shot call below is missing. */
    g_ray_ok = 1;
    for (i = 0; i < 6; i++) {
        site[i] = pattern_call(base, nt, un[i]);
        fn[i] = site[i] != NULL ? call_target(site[i], begin, end) : NULL;
        if (fn[i] == NULL) {
            log_line("fpvmove: shot call %d not found; shot rig off", i);
            return;
        }
    }
    {
        const char *const tw[3] = {kShotAT, kShotBT, kShotCT};
        for (i = 0; i < 3; i++) {
            uint8_t *at = pattern_call(base, nt, tw[i]);
            twin[i] = at != NULL ? call_target(at, begin, end) : NULL;
            if (twin[i] == NULL) {
                log_line("fpvmove: targeted shot call %d not found; shot rig off", i);
                return;
            }
        }
    }
    if (fn[1] != fn[0] || fn[2] != fn[0] || fn[3] != fn[0] || twin[0] == fn[0] || twin[1] == fn[4] || twin[2] == fn[5] ||
        fn[4] == fn[0] || fn[5] == fn[0]) {
        log_line("fpvmove: shot spawners do not pair up; shot rig off");
        return;
    }
    g_shot_a = (shot_a_fn)fn[0];
    g_shot_at = (shot_at_fn)twin[0];
    g_shot_b = (shot_b_fn)fn[4];
    g_shot_bt = (shot_bt_fn)twin[1];
    g_shot_c = (shot_c_fn)fn[5];
    g_shot_ct = (shot_ct_fn)twin[2];
    /* A failed redirect leaves that weapon on its native path. */
    for (i = 0; i < 6; i++) {
        if (redirect_call(site[i], wrap[i], "shot")) {
            redirected++;
        }
    }
    log_line("fpvmove: shot rig armed: line rva 0x%X, %d of 6 shot calls redirected, spawners 0x%X/0x%X 0x%X/0x%X 0x%X/0x%X",
             (unsigned)(line - base), redirected, (unsigned)(fn[0] - base), (unsigned)(twin[0] - base),
             (unsigned)(fn[4] - base), (unsigned)(twin[1] - base), (unsigned)(fn[5] - base), (unsigned)(twin[2] - base));
}

static void diff_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    const uint8_t *g;
    int count = scan_exec(base, nt, kDifficulty, &hit);
    if (count != 1) {
        log_line("fpvmove: difficulty matches=%d; aim assist default off", count);
        return;
    }
    g = rip_target(hit, 3, 7);
    if ((uintptr_t)g < begin || (uintptr_t)g >= end) {
        log_line("fpvmove: difficulty global outside the game; aim assist default off");
        return;
    }
    g_diff_pp = (const uint8_t *const volatile *)g;
    log_line("fpvmove: difficulty read at [rva 0x%X]+6", (unsigned)((uintptr_t)g - begin));
}

/* 0.8.0. The three hold-up status tests go through holdup_gate, and the
   one call of the aim line's muzzle matrix through aimline_wrap. Needs the
   shot rig's line check (the rig itself). */
static void holdup_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt, const uint8_t *test_fn_at) {
    const char *const pats[3] = {kHoldE, kHoldF, kHoldB};
    const size_t at[3] = {kHoldCall, kHoldCall, kHoldBCall};
    const char *const names[3] = {"hold-up 0xE", "hold-up 0xF", "hold-up 0xE variant B"};
    const uint8_t *hit = NULL;
    const uint8_t *fn = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    uint8_t *site;
    int gates = 0;
    int line = 0;
    int count;
    int i;
    if (!g_ray_ok || g_test == NULL || (const uint8_t *)g_test != test_fn_at) {
        log_line("fpvmove: shot rig off; hold-up drop native");
        return;
    }
    for (i = 0; i < 3; i++) {
        count = scan_exec(base, nt, pats[i], &hit);
        if (count != 1 || hit == NULL) {
            log_line("fpvmove: %s matches=%d; unchanged", names[i], count);
            continue;
        }
        site = (uint8_t *)hit + at[i];
        if (site[0] != 0xE8 || rip_target(site, 1, 5) != test_fn_at) {
            log_line("fpvmove: %s does not call TEST; unchanged", names[i]);
            continue;
        }
        if (redirect_call_rbx(site, (void *)holdup_gate, names[i])) {
            gates++;
            log_line("fpvmove: %s gate at rva 0x%X", names[i], (unsigned)(site - base));
        }
    }
    count = scan_exec(base, nt, kAimLine, &hit);
    if (count != 1 || hit == NULL || scan_exec(base, nt, kMuzzleFn, &fn) != 1 || fn == NULL) {
        log_line("fpvmove: aim line matches=%d or muzzle function missing; aim line native", count);
    } else {
        site = (uint8_t *)hit + kAimLineCall;
        g_player_pp = (uint8_t *const volatile *)rip_target(fn, kMuzzlePlayer, kMuzzlePlayer + 4);
        if (site[0] != 0xE8 || rip_target(site, 1, 5) != fn || (uintptr_t)g_player_pp < begin ||
            (uintptr_t)g_player_pp >= end) {
            g_player_pp = NULL;
            log_line("fpvmove: aim line call not recognized; aim line native");
        } else {
            g_muzzle = (muzzle_fn)fn;
            line = redirect_call(site, (void *)aimline_wrap, "aim line");
            if (line) {
                log_line("fpvmove: aim line at rva 0x%X, muzzle rva 0x%X, player rva 0x%X", (unsigned)(site - base),
                         (unsigned)(fn - base), (unsigned)((uintptr_t)g_player_pp - begin));
            }
        }
    }
    log_line("fpvmove: hold-up drop armed: %d of 3 gates, aim line %s", gates, line ? "on the rig" : "native");
}

/* 0.8.0. The wall-press component's layout the wall aim reads: its
   registration (hash 0xD9728F), the state call at node+0x100 inside its
   handler, the camera record store at node+0x120, and the wall yaw store
   at +0x232. */
static void wall_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *reg = NULL;
    const uint8_t *st = NULL;
    const uint8_t *cam = NULL;
    const uint8_t *yaw = NULL;
    const uint8_t *handler;
    if (scan_exec(base, nt, kWallReg, &reg) != 1 || scan_exec(base, nt, kWallState, &st) != 1 ||
        scan_exec(base, nt, kWallCam, &cam) != 1 || scan_exec(base, nt, kWallYaw, &yaw) != 1) {
        log_line("fpvmove: wall-press layout not found; wall aim off");
        return;
    }
    handler = rip_target(reg, 3, 7);
    if (st <= handler || st >= handler + 0x400) {
        log_line("fpvmove: wall state call outside the wall handler; wall aim off");
        return;
    }
    g_wall_ok = 1;
    log_line("fpvmove: wall aim armed, wall handler rva 0x%X", (unsigned)(handler - base));
}

static void mreq_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    if (scan_exec(base, nt, kMotionReq, &hit) != 1 || scan_exec(base, nt, kMotionDec, &hit) != 1) {
        log_line("fpvmove: motion request layout not found; wall motion probe off");
        return;
    }
    g_mreq_ok = 1;
}

static int xh_init(uint8_t *base, IMAGE_NT_HEADERS64 *nt) {
    const uint8_t *hit = NULL;
    const uint8_t *fn_hit = NULL;
    uintptr_t begin = (uintptr_t)base;
    uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
    uint8_t *site;
    const uint8_t *view[4];
    const uint8_t *mtx;
    int i;
    int count = scan_exec(base, nt, kPresentFn, &fn_hit);
    if (count != 1) {
        log_line("fpvmove: present wrapper matches=%d; crosshair off", count);
        return 0;
    }
    count = scan_exec(base, nt, kPresentCall, &hit);
    if (count != 1 || rip_target(hit + 7, 1, 5) != fn_hit) {
        log_line("fpvmove: present call matches=%d or targets another function; crosshair off", count);
        return 0;
    }
    g_renderer_pp = (uint8_t *const volatile *)rip_target(hit, 3, 7);
    if (scan_exec(base, nt, kDevice, &hit) != 1 || scan_exec(base, nt, kContext1, &hit) != 1) {
        log_line("fpvmove: device or context offsets not confirmed; crosshair off");
        return 0;
    }
    if (scan_exec(base, nt, kProjMatrix, &hit) != 1) {
        log_line("fpvmove: projection matrix not found; crosshair off");
        return 0;
    }
    mtx = rip_target(hit, 3, 7);
    if (scan_exec(base, nt, kProjView, &hit) != 1) {
        log_line("fpvmove: projection viewport not found; crosshair off");
        return 0;
    }
    view[0] = rip_target(hit, 4, 8);
    view[1] = rip_target(hit, 0x19, 0x1D);
    view[2] = rip_target(hit, 0x21, 0x25);
    view[3] = rip_target(hit, 0x38, 0x3C);
    if (view[0] - 0x320 != mtx - 0xC0 || view[1] != view[0] + 4 || view[2] != view[0] - 0x10 || view[3] != view[0] - 0xC) {
        log_line("fpvmove: projection fields are not one channel; crosshair off");
        return 0;
    }
    if (scan_exec(base, nt, kPauseLevel, &hit) != 1) {
        log_line("fpvmove: pause level not found; crosshair off");
        return 0;
    }
    g_pause = (const volatile int32_t *)rip_target(hit, 2, 6);
    if ((uintptr_t)g_renderer_pp < begin || (uintptr_t)g_renderer_pp >= end || (uintptr_t)mtx < begin ||
        (uintptr_t)mtx >= end || (uintptr_t)g_pause < begin || (uintptr_t)g_pause >= end) {
        log_line("fpvmove: crosshair globals outside the game; crosshair off");
        return 0;
    }
    g_dg_mtx = (const float *)mtx;
    for (i = 0; i < 4; i++) {
        g_dg_view[i] = (const volatile int32_t *)view[i];
    }
    site = (uint8_t *)fn_hit + kPresentFnCall;
    if (site[0] != 0xE8) {
        log_line("fpvmove: pre-present call not found; crosshair off");
        return 0;
    }
    g_pre_present = (void_fn)rip_target(site, 1, 5);
    if (!redirect_call(site, (void *)pre_present_wrap, "pre-present")) {
        return 0;
    }
    g_xh_ok = 1;
    log_line("fpvmove: crosshair armed at rva 0x%X, matrix rva 0x%X", (unsigned)(site - base), (unsigned)(mtx - base));
    return 1;
}

static void apply(void) {
    uint8_t *base = (uint8_t *)GetModuleHandleW(NULL);
    IMAGE_DOS_HEADER *dos;
    IMAGE_NT_HEADERS64 *nt;
    const uint8_t *match = NULL;
    int count;
    truncate_log();
    log_line("fpvmove: 0.8.4");
    if (base == NULL) {
        log_line("fpvmove: game module missing; nothing changed");
        return;
    }
    dos = (IMAGE_DOS_HEADER *)base;
    if (dos->e_magic != IMAGE_DOS_SIGNATURE) {
        log_line("fpvmove: game module is not a PE; nothing changed");
        return;
    }
    nt = (IMAGE_NT_HEADERS64 *)(base + dos->e_lfanew);
    if (nt->Signature != IMAGE_NT_SIGNATURE || nt->OptionalHeader.Magic != IMAGE_NT_OPTIONAL_HDR64_MAGIC) {
        log_line("fpvmove: game module is not PE32+; nothing changed");
        return;
    }
    count = scan_exec(base, nt, kFlag, &match);
    if (count != 1 || match == NULL) {
        log_line("fpvmove: pattern matches=%d; nothing changed", count);
        return;
    }
    {
        int32_t *flag = (int32_t *)rip_target(match, kFlagDisp, kFlagInsn);
        uintptr_t addr = (uintptr_t)flag;
        uintptr_t begin = (uintptr_t)base;
        uintptr_t end = begin + nt->OptionalHeader.SizeOfImage;
        int32_t previous;
        if (addr < begin || addr > end - sizeof(int32_t) || (addr & 3u) != 0) {
            log_line("fpvmove: resolved flag is outside the game module; nothing changed");
            return;
        }
        previous = *flag;
        if (!write_one(flag)) {
            log_line("fpvmove: flag page is not writable; nothing changed");
            return;
        }
        g_flag = flag;
        log_line("fpvmove: set movement flag to 1 (was %d) at rva 0x%llX", previous, (unsigned long long)(addr - begin));
    }
    /* The movement flag alone is Phase 1. The gates below are Phase 2a. */
    g_square_off = patch_site(base, nt, "walk square gate", kSquare, kSquareAt, kSquareOld, kSquareNew, sizeof kSquareOld);
    g_list_off = patch_site(base, nt, "walk status 0x2E", kList, kListAt, kListOld, kListNew, sizeof kListOld);
    patch_site(base, nt, "look left-stick gate", kLook, kLookAt, kLookOld, kLookNew, sizeof kLookOld);
    patch_site(base, nt, "prone grass first person", kGrassFp, 0, kGrassOld, kGrassNew, sizeof kGrassOld);
    {
        int walk = hook_walk(base, nt);
        int pad = hook_pad(base, nt);
        if (pad && trace_init(base, nt)) {
            split_init(base, nt);
            if (g_split) {
                /* The status TEST helper starts at the bitset pattern. */
                const uint8_t *test_at = NULL;
                if (scan_exec(base, nt, kBits, &test_at) == 1) {
                    if (aim_init(base, nt, test_at) && g_notarget_ok) {
                        if (pitch_init(base, nt)) {
                            shot_init(base, nt);
                            holdup_init(base, nt, test_at);
                        }
                    }
                    cstrafe_init(base, nt, test_at);
                }
                diff_init(base, nt);
                ctrl_init(base, nt);
                xh_init(base, nt);
                wall_init(base, nt);
                mreq_init(base, nt);
            }
        }
        if (walk || pad) {
            _beginthreadex(NULL, 0, watch_walk, NULL, 0, NULL);
        }
    }
}

BOOL WINAPI DllMain(HINSTANCE instance, DWORD reason, LPVOID reserved) {
    (void)reserved;
    if (reason == DLL_PROCESS_ATTACH) {
        DisableThreadLibraryCalls(instance);
        g_self = instance;
        apply();
    }
    return TRUE;
}

#else

/* File offsets for this hash only. The plugin never stores them. */
struct expect {
    const char *label;
    const char *text;
    unsigned long at;
};

static const struct expect kExpect[] = {
    {"flag", kFlag, 0x38BB2Dul},        {"square", kSquare, 0x38BB80ul},  {"list", kList, 0x38BB95ul},
    {"look", kLook, 0x389FA6ul},        {"walkcall", kWalkCall, 0x38AD12ul}, {"bits", kBits, 0x358420ul},
    {"padreg", kPadReg, 0x357CBAul},    {"padfn", kPadFn, 0x356F30ul},    {"aim", kAim, 0x357011ul},
    {"treewalk", kTreeWalk, 0x32F94Cul}, {"walker", kWalker, 0x365620ul}, {"hashlookup", kHashLookup, 0x365800ul},
    {"moveinit", kMoveInit, 0x368B7Ful}, {"movesite", kMoveSite, 0x3688DBul}, {"rollgate", kRollGate, 0x368509ul},
    {"s2p", kS2P, 0x367E58ul},           {"p2s", kP2S, 0x36669Aul},         {"keylookup", kKeyLookup, 0x1117E0ul},
    {"bindfn", kBindFn, 0x33FE0ul},         {"crawl", kCrawl, 0x36637Ful},     {"rollprone", kRollProne, 0x3677F7ul},
    {"crawl2", kCrawl2, 0x366732ul},
    {"camobj", kCamObj, 0x211096ul},     {"camyaw", kCamYaw, 0x21246Ful},   {"camtarget", kCamTarget, 0x21D940ul},
    {"cammode", kCamMode, 0x21143Bul},   {"camsub", kCamSub, 0x213CF9ul},   {"facewpn", kFaceWpn, 0x379D25ul},
    {"notarget", kNoTarget, 0x379C1Cul}, {"strafe", kStrafeStand, 0x369055ul}, {"ctrlcall", kCtrlCall, 0x11034Dul},
    {"ctrlfn", kCtrlFn, 0x115130ul},     {"padcfg", kPadCfg, 0x34200ul},
    {"build", kBuild, 0x37808Ful},       {"firecall", kFireCall, 0x373DA5ul}, {"zoom", kZoom, 0x211EA0ul}, {"camwall", kCamWall, 0x212DAFul}, {"camturn", kCamTurn, 0x21240Ful},
    {"mousedelta", kMouseDelta, 0x1123DBul}, {"difficulty", kDifficulty, 0x5CA5A0ul}, {"presentcall", kPresentCall, 0x27C77ul},
    {"presentfn", kPresentFn, 0x6E350ul}, {"device", kDevice, 0x72830ul},     {"context1", kContext1, 0x728A6ul},
    {"projmatrix", kProjMatrix, 0x123345ul}, {"projview", kProjView, 0x1233E4ul}, {"pause", kPauseLevel, 0x10E0F0ul},
    {"shotray", kShotRay, 0x373DCAul},   {"shota1", kShotA1, 0x3590E9ul},   {"shota2", kShotA2, 0x359CEAul},
    {"shota3", kShotA3, 0x35AC8Aul},     {"shota4", kShotA4, 0x35D6F6ul},   {"shotb", kShotB, 0x35BB90ul},
    {"shotc", kShotC, 0x35C691ul},       {"shotat", kShotAT, 0x3590ADul},   {"shotbt", kShotBT, 0x35BB3Eul},
    {"shotct", kShotCT, 0x35C62Dul},     {"camstance", kCamStance, 0x212BA2ul},
    {"holde", kHoldE, 0x2004A3ul},       {"holdf", kHoldF, 0x202112ul},     {"holdb", kHoldB, 0x2009E9ul},
    {"aimline", kAimLine, 0x381FA2ul},   {"muzzlefn", kMuzzleFn, 0x363DF0ul}, {"wallreg", kWallReg, 0x35046Eul},
    {"wallstate", kWallState, 0x34EF12ul}, {"wallcam", kWallCam, 0x34EC5Cul}, {"wallyaw", kWallYaw, 0x34E4DCul},
    {"motionreq", kMotionReq, 0x380CE0ul}, {"motiondec", kMotionDec, 0x365263ul}, {"motionsend", kMotionSend, 0x368D62ul}, {"grassfp", kGrassFp, 0x37DAAFul},
};

/* Sub-state file offsets for this hash: stand, move, crouch, crouch to
   prone, prone, prone to crouch (RVA - 0xC00). */
static const unsigned long kStateAt[9] = {0x3686C0ul, 0x368430ul, 0x367AB0ul, 0x367EB0ul,
                                          0x366770ul, 0x366BD0ul, 0x3663C0ul, 0x3670D0ul, 0x366040ul};

int main(int argc, char **argv) {
    FILE *file;
    uint8_t *buf;
    long size;
    size_t i;
    int bad = 0;
    if (argc != 2) {
        fprintf(stderr, "usage: fpvmove-test <METAL GEAR SOLID3.exe>\n");
        return 2;
    }
    file = fopen(argv[1], "rb");
    if (file == NULL) {
        perror(argv[1]);
        return 1;
    }
    if (fseek(file, 0, SEEK_END) != 0 || (size = ftell(file)) < 0 || fseek(file, 0, SEEK_SET) != 0) {
        fclose(file);
        fprintf(stderr, "seek failed\n");
        return 1;
    }
    buf = (uint8_t *)malloc((size_t)size);
    if (buf == NULL || fread(buf, 1, (size_t)size, file) != (size_t)size) {
        free(buf);
        fclose(file);
        fprintf(stderr, "read failed\n");
        return 1;
    }
    fclose(file);
    for (i = 0; i < sizeof kExpect / sizeof kExpect[0]; i++) {
        const uint8_t *hit = NULL;
        int count = match_pattern(buf, (size_t)size, kExpect[i].text, &hit);
        unsigned long at = hit == NULL ? 0 : (unsigned long)(hit - buf);
        printf("%s matches=%d offset=0x%lX\n", kExpect[i].label, count, at);
        if (count != 1 || at != kExpect[i].at) {
            fprintf(stderr, "%s: expected one match at 0x%lX\n", kExpect[i].label, kExpect[i].at);
            bad = 1;
        }
    }
    if (!bad) {
        const uint8_t *hit = NULL;
        match_pattern(buf, (size_t)size, kSquare, &hit);
        bad |= memcmp(hit + kSquareAt, kSquareOld, sizeof kSquareOld) != 0;
        match_pattern(buf, (size_t)size, kList, &hit);
        bad |= memcmp(hit + kListAt, kListOld, sizeof kListOld) != 0;
        match_pattern(buf, (size_t)size, kLook, &hit);
        bad |= memcmp(hit + kLookAt, kLookOld, sizeof kLookOld) != 0;
        match_pattern(buf, (size_t)size, kGrassFp, &hit);
        bad |= memcmp(hit, kGrassOld, sizeof kGrassOld) != 0;
        match_pattern(buf, (size_t)size, kWalkCall, &hit);
        {
            /* In the file, .text raw data sits 0xC00 below its RVA, and
               rel32 is the same either way. */
            const uint8_t *site = hit + kWalkCallAt;
            const uint8_t *walk = rip_target(site, 1, 5);
            unsigned long walk_at = (unsigned long)(walk - buf);
            printf("walk offset=0x%lX\n", walk_at);
            bad |= site[0] != 0xE8 || walk_at != 0x38BAB0ul || memcmp(walk, kWalkHead, sizeof kWalkHead) != 0 ||
                   memcmp(walk + kWalkBodyAt, kWalkBody, sizeof kWalkBody) != 0;
        }
        {
            const uint8_t *reg = NULL;
            const uint8_t *fn = NULL;
            match_pattern(buf, (size_t)size, kPadReg, &reg);
            match_pattern(buf, (size_t)size, kPadFn, &fn);
            /* .text is shifted uniformly in the file, so rel32 still lands on
               the component. The raw pad lives in .data, a different shift,
               so only its two loads are compared with each other. */
            printf("padreg target=0x%lX\n", (unsigned long)(rip_target(reg, kPadRegDisp, kPadRegNext) - buf));
            bad |= rip_target(reg, kPadRegDisp, kPadRegNext) != fn;
            bad |= rip_target(fn, 0x21, 0x25) != rip_target(fn, 0x1B, 0x1F) + 0x20;
        }
        {
            const uint8_t *update = NULL;
            const uint8_t *walker = NULL;
            match_pattern(buf, (size_t)size, kTreeWalk, &update);
            match_pattern(buf, (size_t)size, kWalker, &walker);
            printf("treewalk target=0x%lX\n", (unsigned long)(rip_target(update + kTreeWalkCall, 1, 5) - buf));
            bad |= rip_target(update + kTreeWalkCall, 1, 5) != walker - kWalkerAt;
        }
        {
            const uint8_t *init = NULL;
            const uint8_t *move = NULL;
            const uint8_t *gate = NULL;
            const uint8_t *s2p = NULL;
            const uint8_t *p2s = NULL;
            const uint8_t *lookup = NULL;
            const uint8_t *bind = NULL;
            const uint8_t *got[9];
            const uint8_t *crawl2 = NULL;
            const uint8_t *crawl = NULL;
            const uint8_t *rollprone = NULL;
            const uint8_t *call;
            const uint8_t *layout_fn;
            const uint8_t *bind_fn;
            size_t k;
            match_pattern(buf, (size_t)size, kMoveInit, &init);
            match_pattern(buf, (size_t)size, kMoveSite, &move);
            match_pattern(buf, (size_t)size, kRollGate, &gate);
            match_pattern(buf, (size_t)size, kS2P, &s2p);
            match_pattern(buf, (size_t)size, kP2S, &p2s);
            match_pattern(buf, (size_t)size, kKeyLookup, &lookup);
            match_pattern(buf, (size_t)size, kBindFn, &bind);
            got[0] = rip_target(init, kInitStill, kInitStill + 4);
            got[1] = rip_target(move, kMoveSiteDisp, kMoveSiteDisp + 4);
            got[2] = rip_target(init, kInitSquat, kInitSquat + 4);
            got[3] = rip_target(s2p, kS2PDisp, kS2PDisp + 4);
            got[4] = rip_target(init, kInitProne, kInitProne + 4);
            got[5] = rip_target(p2s, kP2SDisp, kP2SDisp + 4);
            match_pattern(buf, (size_t)size, kCrawl, &crawl);
            match_pattern(buf, (size_t)size, kRollProne, &rollprone);
            got[6] = rip_target(crawl, kCrawlDisp, kCrawlDisp + 4);
            got[7] = rip_target(rollprone, kRollProneDisp, kRollProneDisp + 4);
            match_pattern(buf, (size_t)size, kCrawl2, &crawl2);
            got[8] = rip_target(crawl2, kCrawlDisp, kCrawlDisp + 4);
            for (k = 0; k < 9; k++) {
                printf("state %u offset=0x%lX\n", (unsigned)k, (unsigned long)(got[k] - buf));
                bad |= (unsigned long)(got[k] - buf) != kStateAt[k];
            }
            bad |= rip_target(p2s, kP2SProne, kP2SProne + 4) != got[4];
            bad |= gate <= got[1] || gate >= got[0];
            /* Roll sub-state 0x367FA0, loaded at +0x6A of the gate. */
            bad |= gate[kRollLea] != 0x48 || gate[kRollLea + 1] != 0x8D || gate[kRollLea + 2] != 0x05;
            printf("roll offset=0x%lX\n", (unsigned long)(rip_target(gate, kRollLea + 3, kRollLea + 7) - buf));
            bad |= (unsigned long)(rip_target(gate, kRollLea + 3, kRollLea + 7) - buf) != 0x3673A0ul;
            /* Enter and Backspace slots sit 5 VKs apart in the key array. */
            bad |= rip_target(lookup, kKeyEnter, kKeyEnter + 4) - 0x0D * 4 != rip_target(lookup, kKeyBack, kKeyBack + 4) - 0x08 * 4;
            call = lookup + kKeyCalls;
            bad |= call[0] != 0xE8 || call[0xC] != 0xE8 || memcmp(call + 5, "\x8B\xC8\x45\x33\xC0\x8B\xD3", 7) != 0;
            layout_fn = rip_target(call, 1, 5);
            bind_fn = rip_target(call + 0xC, 1, 5);
            printf("layout getter offset=0x%lX binding getter offset=0x%lX\n", (unsigned long)(layout_fn - buf),
                   (unsigned long)(bind_fn - buf));
            bad |= bind_fn != bind || layout_fn[0] != 0x8B || layout_fn[1] != 0x05 || layout_fn[6] != 0xC3;
            /* lea rcx, [image base]: RVA 0 lands 0xC00 below the file start. */
            bad |= rip_target(bind, kBindBase, kBindBase + 4) != buf - 0xC00;
            {
                /* Default tables live in .data (file = RVA - 0x1200). Table 0
                   binds Cross (key id 9) to SPACE and leaves C free. */
                const size_t at[2] = {kBindT0, kBindT1};
                int t;
                for (t = 0; t < 2; t++) {
                    uint32_t rva;
                    const uint32_t *table;
                    int id;
                    int owner = -1;
                    memcpy(&rva, bind + at[t], sizeof rva);
                    table = (const uint32_t *)(buf + rva - 0x1200);
                    for (id = 0; id < 26; id++) {
                        if (table[id * 4] == 0x43 || table[id * 4 + 1] == 0x43) {
                            owner = id;
                        }
                    }
                    printf("table %d rva=0x%X cross=0x%X/0x%X C owner=%d\n", t, rva, table[9 * 4], table[9 * 4 + 1], owner);
                    if (t == 0) {
                        bad |= table[9 * 4] != 0x20 || owner != -1;
                    }
                }
            }
        }
        {
            /* 0.5.9: the redirected and checked calls land on the TEST helper
               and on the controller read, and the config lea is the image
               base. */
            const uint8_t *test_at = NULL;
            const uint8_t *strafe = NULL;
            const uint8_t *notarget = NULL;
            const uint8_t *call = NULL;
            const uint8_t *fn = NULL;
            const uint8_t *cfg = NULL;
            uint32_t table_rva;
            match_pattern(buf, (size_t)size, kBits, &test_at);
            match_pattern(buf, (size_t)size, kStrafeStand, &strafe);
            match_pattern(buf, (size_t)size, kNoTarget, &notarget);
            match_pattern(buf, (size_t)size, kCtrlCall, &call);
            match_pattern(buf, (size_t)size, kCtrlFn, &fn);
            match_pattern(buf, (size_t)size, kPadCfg, &cfg);
            bad |= strafe[kStrafeStandCall] != 0xE8 || rip_target(strafe + kStrafeStandCall, 1, 5) != test_at;
            bad |= notarget[kNoTargetCall] != 0xE8 || rip_target(notarget + kNoTargetCall, 1, 5) != test_at;
            bad |= call[0] != 0xE8 || rip_target(call, 1, 5) != fn;
            bad |= rip_target(cfg, kPadCfgBaseDisp, kPadCfgBaseNext) != buf - 0xC00;
            memcpy(&table_rva, cfg + kPadCfgTable, sizeof table_rva);
            {
                /* 0.6.0: the fire call lands on the builder (after its int3),
                   the search call on the search prologue, the present call
                   on the present wrapper, whose +0x15 is a call, and the
                   viewport ints sit in the matrix's channel. */
                const uint8_t *build = NULL;
                const uint8_t *fire = NULL;
                const uint8_t *pcall = NULL;
                const uint8_t *pfn = NULL;
                const uint8_t *pm = NULL;
                const uint8_t *pv = NULL;
                match_pattern(buf, (size_t)size, kBuild, &build);
                match_pattern(buf, (size_t)size, kFireCall, &fire);
                match_pattern(buf, (size_t)size, kPresentCall, &pcall);
                match_pattern(buf, (size_t)size, kPresentFn, &pfn);
                match_pattern(buf, (size_t)size, kProjMatrix, &pm);
                match_pattern(buf, (size_t)size, kProjView, &pv);
                bad |= rip_target(fire + kFireCallAt, 1, 5) != build + 1 || memcmp(build + 1, kBuildHead, sizeof kBuildHead) != 0;
                {
                    const uint8_t *zoom = NULL;
                    match_pattern(buf, (size_t)size, kZoom, &zoom);
                    bad |= memcmp(zoom, kZoomHead, sizeof kZoomHead) != 0;
                }
                bad |= rip_target(pcall + 7, 1, 5) != pfn || pfn[kPresentFnCall] != 0xE8;
                bad |= rip_target(pv, 4, 8) - 0x320 != rip_target(pm, 3, 7) - 0xC0;
                printf("builder=0x%lX prepresent=0x%lX\n", (unsigned long)(build + 1 - buf),
                       (unsigned long)(rip_target(pfn + kPresentFnCall, 1, 5) - buf));
            }
            {
                /* 0.6.4: the shot ray's three calls, the six untargeted
                   spawner calls (four share one spawner), their three
                   targeted twins, and the camera wall step prologue. */
                const char *const un[6] = {kShotA1, kShotA2, kShotA3, kShotA4, kShotB, kShotC};
                const char *const tw[3] = {kShotAT, kShotBT, kShotCT};
                const uint8_t *site[6];
                const uint8_t *twin[3];
                const uint8_t *ray = NULL;
                const uint8_t *ease = NULL;
                const uint8_t *line;
                int n;
                match_pattern(buf, (size_t)size, kShotRay, &ray);
                line = rip_target(ray + kShotRayLine, 1, 5);
                bad |= memcmp(line, kLineHead, sizeof kLineHead) != 0;
                printf("line=0x%lX hitidx=0x%lX hitpt=0x%lX\n", (unsigned long)(line - buf),
                       (unsigned long)(rip_target(ray + kShotRayIdx, 1, 5) - buf),
                       (unsigned long)(rip_target(ray + kShotRayPt, 1, 5) - buf));
                bad |= (unsigned long)(line - buf) != 0x10A8B0ul;
                for (n = 0; n < 6; n++) {
                    uint8_t pat[MAX_PAT];
                    char mask[MAX_PAT + 1];
                    size_t len = parse_pattern(un[n], pat, mask);
                    match_pattern(buf, (size_t)size, un[n], &site[n]);
                    site[n] += len - 1;
                    printf("shot %d calls 0x%lX\n", n, (unsigned long)(rip_target(site[n], 1, 5) - buf));
                }
                for (n = 0; n < 3; n++) {
                    uint8_t pat[MAX_PAT];
                    char mask[MAX_PAT + 1];
                    size_t len = parse_pattern(tw[n], pat, mask);
                    match_pattern(buf, (size_t)size, tw[n], &twin[n]);
                    twin[n] += len - 1;
                    printf("twin %d calls 0x%lX\n", n, (unsigned long)(rip_target(twin[n], 1, 5) - buf));
                }
                for (n = 1; n < 4; n++) {
                    bad |= rip_target(site[n], 1, 5) != rip_target(site[0], 1, 5);
                }
                bad |= (unsigned long)(rip_target(site[0], 1, 5) - buf) != 0x25E640ul;
                bad |= (unsigned long)(rip_target(twin[0], 1, 5) - buf) != 0x25E8E0ul;
                bad |= (unsigned long)(rip_target(site[4], 1, 5) - buf) != 0x25E790ul;
                bad |= (unsigned long)(rip_target(twin[1], 1, 5) - buf) != 0x25EA30ul;
                bad |= (unsigned long)(rip_target(site[5], 1, 5) - buf) != 0x261650ul;
                bad |= (unsigned long)(rip_target(twin[2], 1, 5) - buf) != 0x261750ul;
                match_pattern(buf, (size_t)size, kCamWall, &ease);
                bad |= memcmp(ease + 1, kCamWallHead, sizeof kCamWallHead) != 0;
                match_pattern(buf, (size_t)size, kCamTurn, &ease);
                bad |= memcmp(ease + 1, kCamTurnHead, sizeof kCamTurnHead) != 0;
            }
            {
                /* 0.8.0: the three hold-up calls land on the TEST helper,
                   the aim line's call on the muzzle function, and the wall
                   state call sits inside the wall handler. */
                const char *const hp[3] = {kHoldE, kHoldF, kHoldB};
                const size_t ha[3] = {kHoldCall, kHoldCall, kHoldBCall};
                const uint8_t *h = NULL;
                const uint8_t *fn = NULL;
                const uint8_t *reg = NULL;
                const uint8_t *st = NULL;
                const uint8_t *handler;
                int n;
                for (n = 0; n < 3; n++) {
                    match_pattern(buf, (size_t)size, hp[n], &h);
                    bad |= h[ha[n]] != 0xE8 || rip_target(h + ha[n], 1, 5) != test_at;
                }
                match_pattern(buf, (size_t)size, kAimLine, &h);
                match_pattern(buf, (size_t)size, kMuzzleFn, &fn);
                bad |= h[kAimLineCall] != 0xE8 || rip_target(h + kAimLineCall, 1, 5) != fn;
                match_pattern(buf, (size_t)size, kWallReg, &reg);
                match_pattern(buf, (size_t)size, kWallState, &st);
                handler = rip_target(reg, 3, 7);
                bad |= st <= handler || st >= handler + 0x400;
                printf("muzzle=0x%lX wall handler=0x%lX\n", (unsigned long)(fn - buf), (unsigned long)(handler - buf));
            }
            printf("strafe call=0x%lX notarget call=0x%lX ctrl callee=0x%lX config table rva=0x%X\n",
                   (unsigned long)(rip_target(strafe + kStrafeStandCall, 1, 5) - buf),
                   (unsigned long)(rip_target(notarget + kNoTargetCall, 1, 5) - buf),
                   (unsigned long)(rip_target(call, 1, 5) - buf), table_rva);
        }
        if (bad) {
            fprintf(stderr, "old bytes, walk callee, or pad component differ\n");
        }
    }
    free(buf);
    return bad;
}

#endif
