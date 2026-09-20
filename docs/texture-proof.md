# Offline texture proof

Date: 2026-09-19. Scope: offline copies under `mgs3-mod-manager/work/` only. No game texture was changed, no package script was run, and the game was not launched.

This document records the initial converter investigation. Its original tag `1.3` difference counts were incorrect and are corrected below. The blocker is now resolved by the pinned local patch in [converter-repair.md](converter-repair.md). Current package and deployment evidence is in [mod-validation.md](mod-validation.md).

## Candidate inputs

The two local candidates were copied from the installation before conversion:

| Candidate | Source | Bytes | SHA-256 |
| --- | --- | ---: | --- |
| HQ | `hqtex/flatlist/_win/sna_def_olive.bmp.ctxr` | 22,370,144 | `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d` |
| fallback | `textures/flatlist/_win/sna_def_olive.bmp.ctxr` | 1,398,560 | `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049` |

Copies, exports, and repack outputs are in ignored `mgs3-mod-manager/work/texture-roundtrip/`. The source copies remain unchanged after all commands.

## Converter provenance and build

Official source: [Jayveer/CtxrTool](https://github.com/Jayveer/CtxrTool). Pinned source tag `1.3`, commit `2e64a511872c233f1ea8424bad795c4fe4daeb91`, author Jayveer, commit date 2023-10-27. The repository README declares MIT licensing and calls DDS-to-CTXR conversion experimental; this tag does not contain a `LICENSE.md` file, so license provenance is the upstream README claim.

The source checkout is `mgs3-mod-manager/work/CtxrTool-source/`. A clean Release x64 build was made with the installed Visual Studio 2022 Community C++ toolchain (`MSVC 14.44.35207`) using `vcvars64.bat` and `cl.exe`; no global installation occurred. Build output `mgs3-mod-manager/work/CtxrTool-tag.exe`:

```text
bytes=123392
sha256=9988f4ac3d875a43cad750f331594c452710de745274e601ad806783af0ccfb3
```

The checkout's current upstream HEAD was also pinned for comparison: `a16f1148b68f2cdd1a326e37fa0b36c75fbadeee`. A direct Release x64 build produced `mgs3-mod-manager/work/CtxrTool-head.exe`, 134,144 bytes, SHA-256 `d10362a1a8842eca3a3443aee85cb9e57d253e04fc41413a6b2c15bde652bca1`. Both revisions show the same roundtrip defect below.

## Export validation

Commands, run from each fixture directory, were equivalent to:

```powershell
& CtxrTool-tag.exe sna_def_olive.bmp.ctxr
& CtxrTool-tag.exe sna_def_olive.bmp.dds
```

The first command created `.dds` and `.param`; the second was run on a copy of those outputs in a separate `*-repack` directory. The converter wrote no output beside the input unless its working directory was the fixture directory.

Both candidates exported successfully. Parsed values:

| Candidate | CTXR / DDS dimensions | CTXR / DDS mip count | DDS format flags | Bits | Channel masks |
| --- | --- | --- | --- | ---: | --- |
| HQ | 4096x1024 / 4096x1024 | 13 / 13 | `0x00000041` (RGB + alpha pixels) | 32 | R `0x00FF0000`, G `0x0000FF00`, B `0x000000FF`, A `0xFF000000` |
| fallback | 1024x256 / 1024x256 | 11 / 11 | `0x00000041` (RGB + alpha pixels) | 32 | R `0x00FF0000`, G `0x0000FF00`, B `0x000000FF`, A `0xFF000000` |

Extracted parameter sidecars were retained. HQ sidecar values are `0,0,0,0,0,0,128,206,208,206,128,255,255`; fallback values are `0,0,0,0,0,0,128,170,170,171,157,255,255`.

## Roundtrip gate

Verdict for the unpatched converters: **WRONG** for any claim that their unchanged CTXR roundtrips passed. The later patched converter passes exact equality; see the repair record. Visual correctness was a separate gate, later resolved to the extent recorded in [mod-validation.md](mod-validation.md).

Tag `1.3` preserves file length but corrupts mip data after the first mip: exhaustive comparison finds 3,779,705 differing HQ bytes and 248,506 fallback bytes. Unpatched current HEAD has a different ten-byte tail-mip defect. The original statement that both revisions differed by only ten bytes was wrong. Tag `1.3` hashes:

| Candidate | Original CTXR SHA-256 | Repacked CTXR SHA-256 | Original payload SHA-256 (offset `0x80`) | Repacked payload SHA-256 (offset `0x80`) | Exact |
| --- | --- | --- | --- | --- | --- |
| HQ | `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d` | `cea8c691062792d93c96254db083fdf0eaf348f4d9f4d148957e630233eab154` | `5410ba7b5136d68eb584fb8540bf53a84e790341c2ec1d387d925fb15d79d334` | `b3204131c27735a7118fb9eed27318b53defeadc1d0b994b41aae41fd4674b95` | no |
| fallback | `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049` | `3d01d9031a9936186edb326b61d2b0bc72aac1def72933007c6bc3a83f132c64` | `24017d8536ff58ef27f41bc7258f9cc75535077a880f63ea3e80325df2a73c6e` | `6e212a50e557493638ea10381bee8ff7269db485a723d61a003909b69246a0f8` | no |

Tag `1.3` first differs at `0x10000A7` (HQ) and `0x1000A7` (fallback), because its DDS export carries CTXR chunk metadata into pixel data. Unpatched HEAD first differs at `0x1555723` (HQ) and `0x1556E3` (fallback), due to incorrect non-square tail mip sizes. These unpatched repacks must not be used as mod payloads.

Current HEAD comparison independently produced HQ repacked SHA-256 `40812d652d187cee1df73b594a764504f8d589998ca817ec47a6c0cc679f8125` with payload SHA-256 `e46ba8a237083bb5fe38a1e6cfc818a05e46539f9a450593070113f5987a1b5d`, and fallback repacked SHA-256 `f170c44a67fed428cd65bc263754dd9fcc792e05d3237f16b7e76ef68443d4e4` with payload SHA-256 `cee078452bc167edbe42ea2f2e9eab6499f7289747de1eaf1d9ce1ae5f95d7f9`.

Reproduction uses .NET streaming SHA-256 because `Get-FileHash` is unavailable in this environment. The comparison must hash the complete files and the byte range beginning at offset `0x80`; a header-only check is insufficient.

## Resolution

The locally patched current-source converter produces byte-identical unchanged repacks for both candidates. This stronger equality proof makes a separate unchanged-input rendering test redundant. The actual recolor and its later in-game validation are recorded in [mod-validation.md](mod-validation.md). No authoring edit was made during this initial investigation.
