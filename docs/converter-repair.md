# Offline CtxrTool roundtrip repair

Date: 2026-09-19. All conversion work used copies under `mgs3-mod-manager/work/converter-repair/`. No game texture was written, no package was installed, no recolor was applied to either real candidate, and the game was not launched.

These observations describe the converter-repair phase, before authoring and deployment. See [mod-validation.md](mod-validation.md) for the later recolor and current live-trial state.

## Findings

The initial proof underestimated the tag `1.3` defect. Tag `1.3` exports CTXR chunk size words and 32-byte alignment padding as though they were DDS pixels. Its repacker skips only the first size word (`dataPtr = 0x4`) and then reads every later mip from the wrong payload offset. The HQ unchanged repack differs in 3,779,705 bytes beginning at `0x10000A7`; the fallback differs in 248,506 bytes beginning at `0x1000A7`. In both cases mip 0 is exact and every later mip differs. These counts come from comparing every byte, not a sample.

Upstream commit `a16f1148b68f2cdd1a326e37fa0b36c75fbadeee` correctly strips CTXR size words and padding during DDS export. Its remaining ten-byte difference has a separate cause: `saveExtend` divides the previous byte count by four for every mip. These textures are non-square, so their final dimensions are `4x1`, `2x1`, and `1x1`, requiring 16, 8, and 4 bytes. The upstream repacker instead emits 16, 4, and 1 bytes. Halving width and height independently fixes this defect.

The upstream DDS header also has a zero pixel-format size and zero pitch. `DDS::create` clears a raw byte allocation, so the `DDS_PIXELFORMAT::pxlFmtSize = 0x20` default member initializer never runs. The local patch explicitly writes pixel-format size `0x20` and pitch `width * 4`.

The provenance patch is [`ctxrtool-a16f1148-roundtrip.patch`](../examples/camo-test/author/ctxrtool-a16f1148-roundtrip.patch). It applies cleanly to commit `a16f1148b68f2cdd1a326e37fa0b36c75fbadeee` and has SHA-256 `ab1040da93127bef31d59c1df6b5750a07111f94edd05adb40d5066290ecc9b4` (1,451 bytes). The patched checkout used for the proof is `work/converter-repair/CtxrTool-fixed/`.

## Build

Visual Studio 2022 Community Developer Command Prompt 17.14.37 and its x64 C++ toolchain built the local executable directly:

```powershell
& cmd.exe /d /s /c 'call "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat" && cd /d "C:\Games\METAL GEAR SOLID 3 - MCV\mgs3-mod-manager\work\converter-repair\CtxrTool-fixed" && cl.exe /nologo /std:c++17 /EHsc /O2 /DNDEBUG /Fe:..\CtxrTool-fixed.exe main.cpp interface\cli.cpp image\ctxr\ctxr.cpp image\dds\dds.cpp image\bmp\bmp.cpp image\tga\tga.cpp /link /INCREMENTAL:NO'
```

`work/converter-repair/CtxrTool-fixed.exe` is 371,712 bytes with SHA-256 `5ae0fb7881b093a31f9828a9bf6dbd458ab838629a88a8167a90373fb2cf8cf4`.

The patch check was:

```powershell
git -c safe.directory='C:/Games/METAL GEAR SOLID 3 - MCV/mgs3-mod-manager/work/CtxrTool-source-head' -C 'work/CtxrTool-source-head' apply --check '..\..\examples\camo-test\author\ctxrtool-a16f1148-roundtrip.patch'
```

The repair-phase notes record exit 0 with no output. The exact historical command transcript was not retained, so that execution self-report is **UNVERIFIABLE**. The current patch bytes and hash, patched executable, and later successful converter outputs remain independently inspectable.

## End-to-end unchanged proof

For each candidate, the original CTXR was copied to `work/converter-repair/roundtrip-e2e/<candidate>/export/`. The patched executable exported DDS plus the parameter sidecar. Those two files were copied to the sibling `repack/` directory and converted back to CTXR. The original copies were never used as converter output paths.

The exhaustive verifier parses all CTXR chunk headers, data ranges, and alignment ranges; compares every original and repacked byte; and verifies that every exported DDS mip equals the corresponding original CTXR mip data. It also validates DDS dimensions, mip count, 32-bit RGB-plus-alpha flags, and all four masks.

```powershell
$env:GOCACHE = (Resolve-Path '.cache/go-build').Path
go run ./work/converter-repair/audit.go `
  'work/texture-roundtrip-head/hq/sna_def_olive.bmp.ctxr' `
  'work/converter-repair/roundtrip-e2e/hq/repack/sna_def_olive.bmp.ctxr' `
  'work/converter-repair/roundtrip-e2e/hq/export/sna_def_olive.bmp.dds'

go run ./work/converter-repair/audit.go `
  'work/texture-roundtrip-head/fallback/sna_def_olive.bmp.ctxr' `
  'work/converter-repair/roundtrip-e2e/fallback/repack/sna_def_olive.bmp.ctxr' `
  'work/converter-repair/roundtrip-e2e/fallback/export/sna_def_olive.bmp.dds'
```

| Candidate | Dimensions | Mips | Original and repacked bytes | Original and repacked CTXR SHA-256 | Exported DDS SHA-256 | Result |
| --- | --- | ---: | ---: | --- | --- | --- |
| HQ | 4096x1024 | 13 | 22,370,144 | `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d` | `fc2e50ba3b6d41c2633b972bce8b31249127806f67515b64907f050cfe03f55c` | zero differing bytes; every mip and padding range exact |
| fallback | 1024x256 | 11 | 1,398,560 | `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049` | `91e3ba33a457ceaf5bf52715a95b098a13ff0749d899da227cbc5537034cda27` | zero differing bytes; every mip and padding range exact |

The HQ DDS is 22,369,756 bytes; the fallback DDS is 1,398,236 bytes. Both use pixel-format size `0x20`, flags `0x41`, 32 bits per pixel, and masks R `0x00FF0000`, G `0x0000FF00`, B `0x000000FF`, A `0xFF000000`. Their pitches are 16,384 and 4,096 bytes respectively. Full mip-byte equality also proves every alpha byte is preserved.

An unchanged live rendering smoke test would execute the original bytes under their original hash, so it cannot add evidence about converter fidelity or rendering differences. The exact unchanged result therefore closes the offline roundtrip gate without that redundant launch. The later recolor trial proved the active HQ target and normal rendering through the user's report; close and distant views were not separately documented.

## Recolor helper

[`examples/camo-test/author`](../examples/camo-test/author/) contains a narrow Go helper for the planned transform. It rejects unexpected DDS format fields, masks, mip counts, truncation, and trailing data; uses case-insensitive explicit path comparison; opens output with `O_CREATE|O_EXCL` so every existing path or same-file alias is refused; checks write, flush, and close results; preserves the 128-byte header; and changes only RGB bits. Synthetic tests independently assert all four masks and the `8x2`, `4x1`, `2x1`, `1x1` mip offsets and sizes that expose the upstream defect.

`go run ./work/converter-repair/audit.go recolor INPUT.dds OUTPUT.dds` independently parses both files and checks the exact RGB formula for every pixel, unchanged alpha, an unchanged 128-byte header, all mip offsets, and the complete file length. Its synthetic tests prove it rejects header, alpha, and RGB corruption. At the end of this repair phase it was ready for offline authoring; no real candidate had been recolored during the phase.

```powershell
$env:GOCACHE = (Resolve-Path '.cache/go-build').Path
$env:GOMODCACHE = (Resolve-Path '.cache/go-mod').Path
go test -count=1 ./examples/camo-test/author
go vet ./examples/camo-test/author
go run ./examples/camo-test/author -check -in 'work/converter-repair/roundtrip-e2e/hq/export/sna_def_olive.bmp.dds'
go run ./examples/camo-test/author -check -in 'work/converter-repair/roundtrip-e2e/fallback/export/sna_def_olive.bmp.dds'
Push-Location work
go test -count=1 ./converter-repair
go vet ./converter-repair
Pop-Location
go test -count=1 ./...
```

The repair-phase notes record that all commands exited 0, the targeted tests reported `ok mgs3mod/examples/camo-test/author` and `ok mgs3mod-local-authoring-work/converter-repair`, the read-only checks accepted the real exports as 4096x1024 with 13 mips and 1024x256 with 11 mips, and the full suite passed all packages. Those exact historical command outputs were not retained and are therefore **UNVERIFIABLE** as process evidence. Current artifact hashes, byte-equality outputs, later authoring results, and the final fresh project suite are separately retained or reproducible. No recolor invocation was made against either real DDS during this repair phase because authoring was still gated.

## Audit verdicts

- The corrected counts at `docs/texture-proof.md:55` -> **CORRECT**. `go run ./work/converter-repair/audit.go ORIGINAL TAG_REPACK` reports 3,779,705 HQ differences and 248,506 fallback differences, with every mip after mip 0 unequal. The earlier ten-byte claim was wrong and has been replaced.
- The distinct tag/current-source defects at `docs/texture-proof.md:62` -> **CORRECT**. Tag `1.3` mishandles embedded chunk metadata; current HEAD mishandles non-square tail dimensions.
- `docs/implementation-plan.md:154` statement that byte identity is desirable rather than mandatory when all differences are understood -> **CORRECT**. The local repair exceeds that gate: both complete CTXR outputs are byte-identical.
- Patched converter preserves dimensions, mip count, pixel bytes, alpha, headers, and CTXR padding -> **CORRECT**. Both exhaustive verifier runs report `byte_equal=true`, `all_mip_data_equal=true`, and `dds_exact=true` across all 24 mip levels.
- Unchanged live smoke adds converter-fidelity evidence beyond byte identity -> **WRONG**. The candidate and repack are the same complete byte sequence and hash; running either supplies the game with identical input.
- Visual correctness or active in-game target for the recolor -> **UNVERIFIABLE** in this converter proof. No recolor or game launch occurred during this phase; later trial results belong in [mod-validation.md](mod-validation.md).
