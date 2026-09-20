# Camo test authoring recipe

This directory contains the reproducible recipe, source manifest, converter patch, and authoring helper. Game-derived payloads are kept outside the source directory. The tested HQ mod is packaged in `../../dist/camo-test-hq-0.1.0.zip`. The user confirmed a magenta uniform with normal rendering; the manager then restored the exact original bytes and left the package disabled.

Candidate source files:

- HQ: `hqtex/flatlist/_win/sna_def_olive.bmp.ctxr`, original SHA-256 `1ba4e570dfeb21ee63e29b76f2acbf73b799db93c3335920023037810ead977d`.
- Fallback: `textures/flatlist/_win/sna_def_olive.bmp.ctxr`, original SHA-256 `855dd1256765e23d034b6690365472c01bdf501ac549ff24e86d4371f06d1049`.

Copy each candidate into ignored `mgs3-mod-manager/work/` before tooling. Never run the converter against a game path. Export CTXR to DDS and keep the generated `.param` sidecar beside the DDS. Validate dimensions, mip count, 32-bit pixel format, channel masks, alpha bytes, and all non-pixel bytes before authoring.

The implemented transform applies only to decoded DDS pixels after the unchanged roundtrip gate passes:

```text
R' = floor((R + 255) / 2)
G' = floor(G / 3)
B' = floor((B + 255) / 2)
A' = A
```

The [author helper](author/README.md) uses widened arithmetic and validated DDS masks. It preserves DDS headers, mip ordering, mip dimensions, and alpha. Repacking uses the original `.param` sidecar and the pinned locally patched converter. The [converter repair](../../docs/converter-repair.md) proves byte-identical unchanged round trips for both candidates; this replaces the redundant unchanged-input live smoke test. The [mod validation record](../../docs/mod-validation.md) records the actual recolor, deployment, user-confirmed visible activation, exact restoration, and remaining detailed visual evidence limits.
