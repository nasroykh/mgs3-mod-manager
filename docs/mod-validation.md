# Camo test validation

Date: 2026-09-19. The selected first trial replaces only `hqtex/flatlist/_win/sna_def_olive.bmp.ctxr`. The user confirmed this HQ resource was active by observing the conspicuous magenta uniform render normally in game.

## Offline package

- Package: `dist/camo-test-hq-0.1.0.zip`.
- Package SHA-256: `2edaa130aefcdfba80631325d3dedbe500bceef28501c768a0817ab8f41492fa` (22,370,859 bytes). The package was deterministically repacked on 2026-09-20 after correcting normalized ZIP timestamps to the 1980 minimum; its manifest and payload bytes are unchanged.
- ID/version: `camo-test`, `0.1.0`.
- Target original: 22,370,144 bytes; SHA-256 `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`.
- Payload: 22,370,144 bytes; SHA-256 `384dc7966fecf246162ab8712b065c5159230c151e62af852fc47c11acb3bb3d`.
- Patched converter: SHA-256 `5ae0fb7881b093a31f9828a9bf6dbd458ab838629a88a8167a90373fb2cf8cf4`; provenance and source patch in [converter-repair.md](converter-repair.md).
- Working package folder: `work/camo-test-package/`. Authoring and re-export files: `work/camo-test-author/`.

The author helper transformed all 5,592,407 pixels across 13 mip levels. The independent verifier returned `recolor_exact=true header_equal=true alpha_equal=true mips=13 pixels=5592407` both before and after CTXR repacking/re-export. The authored and re-exported DDS files have identical SHA-256 `93b3d3910b5fa63419bd60672580161f7b96616a5e99b49c66a29b00ad4a32c8`.

The CTXR header, all 13 mip sizes and offsets, and every padding byte remain identical to the original. The pixel payload intentionally differs. The equality-only CTXR audit therefore exited 1 with `byte_equal=false` and `all_mip_data_equal=false`; this expected result is not claimed as a passing unchanged test. The independent recolor audit is the acceptance check for changed pixels. A PNG preview was inspected and shows a magenta cloth/equipment texture with the original pattern retained.

Both `pack --dry-run --json` and `pack --json` accepted the local package. They did not initialize manager state or modify game files.

## Reproduction

Run from `mgs3-mod-manager` with project-local Go caches and new output paths:

```powershell
go run ./examples/camo-test/author -in ./work/converter-repair/roundtrip-e2e/hq/export/sna_def_olive.bmp.dds -out ./work/camo-test-author/sna_def_olive.bmp.dds
go run ./work/converter-repair/audit.go recolor ./work/converter-repair/roundtrip-e2e/hq/export/sna_def_olive.bmp.dds ./work/camo-test-author/sna_def_olive.bmp.dds
```

Copy the retained `sna_def_olive.bmp.param` beside that DDS, run the patched converter with the authoring directory as its working directory, and re-export the resulting CTXR from a separate verification directory. Repeat the independent recolor check against the re-exported DDS. The helper refuses existing outputs; never run the converter against a game path.

## Live acceptance

The real-installation lifecycle passed with the game closed: initialize, import disabled, enable, verify, disable, verify, enable again, restore baseline, verify, remove, verify, and re-import disabled. Read-only checks and dry runs also passed. The first 18 receipts under `work/live-test-20260919/` all report `ok: true`, counted by enumerating and parsing every receipt before the visual trial.

Post-lifecycle `status` reported generation 7, one stored mod (`camo-test`) with `enabled: false`, one retained baseline, compatible core fingerprints, and no unresolved recovery. Independent .NET SHA-256 checks confirmed the live HQ target and retained original backup both equaled `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`. The standard-resolution candidate remained `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049`.

The user then requested preparation of the visual trial. Receipt `19-enable-visual-trial.json` records successful activation at generation 8; a subsequent `verify --json` passed. The user reported: "Magenta uniform; rendering looks normal; game closed." This confirms the HQ candidate through the user's observation; no independent screenshot was captured.

After game closure, `disable camo-test --json` and `verify --json` both exited 0. Receipts `20-disable-after-visual.json` and `21-verify-after-visual.json` record final generation 9, `enabled: false`, compatible installation, and no unresolved transaction. Independent .NET SHA-256 checks again confirm that both the live target and retained backup equal original hash `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`. The package remains stored for reuse. A separate post-restoration visual launch was not performed; restoration is established by complete byte identity.

The requested visual procedure was to select Olive Drab/Normal camouflage, inspect Snake's body and menu preview, and check close and distant views. The user's report established visible activation and normal rendering, but did not separately record each requested view and no screenshot was captured. The standard-resolution candidate was therefore not tried. A separate post-restoration visual launch also remains unperformed; byte identity proves file restoration, not the omitted visual observation.
