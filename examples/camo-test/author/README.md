# Camo DDS authoring helper

This narrow helper implements only the planned Normal/Olive Drab recolor for the validated uncompressed 32-bit DDS layout. It rejects compressed data, unexpected channel masks, impossible mip counts, truncated mip data, and unexplained trailing bytes. It requires a new output path and atomically refuses every existing output, including same-file aliases. Successful output is flushed and closed before success is reported.

The transform is `R' = floor((R + 255) / 2)`, `G' = floor(G / 3)`, `B' = floor((B + 255) / 2)`, and `A' = A`. The code reads and writes channels through the validated DDS masks. Tests use a non-square `8x2` full mip chain so the last four mip sizes and offsets are checked independently as `64`, `16`, `8`, and `4` bytes at offsets `128`, `192`, `208`, and `216`.

Run tests from `mgs3-mod-manager` with project-local Go caches:

```powershell
$env:GOCACHE = (Resolve-Path '.cache/go-build').Path
$env:GOMODCACHE = (Resolve-Path '.cache/go-mod').Path
go test ./examples/camo-test/author
```

Validate an exported DDS without writing output:

```powershell
go run ./examples/camo-test/author -check -in INPUT.dds
```

After the exact unchanged roundtrip passes, create a separate offline copy with:

```powershell
go run ./examples/camo-test/author -in INPUT.dds -out OUTPUT.dds
```

Independently verify every output pixel, mip boundary, alpha byte, and header byte:

```powershell
go run ./work/converter-repair/audit.go recolor INPUT.dds OUTPUT.dds
```

Success reports `recolor_exact=true`, `header_equal=true`, and `alpha_equal=true`. This verifier has separate parsing and formula code from the authoring helper.

`ctxrtool-a16f1148-roundtrip.patch` applies to upstream CtxrTool commit `a16f1148b68f2cdd1a326e37fa0b36c75fbadeee`. Its DDS exporter already removes CTXR chunk headers and alignment. The patch fixes repacking of non-square mip tails by halving width and height independently, preserving the required `2x1` eight-byte mip before the final `1x1` four-byte mip. It also writes the required 32-byte DDS pixel-format size and scanline pitch, which upstream leaves zero after clearing the header allocation.
