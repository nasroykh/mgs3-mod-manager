# QCamo provenance

This distribution adapts the official QCamo binary to MGS3 Mod Manager's package format. QCamo itself is authored by its upstream contributors; it has not been rebuilt or modified here. The package integration is the first non-test mod in this manager project. It is not a claim of original authorship of QCamo.

- Source: https://github.com/zexk/mgs3-qcamo
- Release: https://github.com/zexk/mgs3-qcamo/releases/tag/v1.0.4
- Source commit: `0a8bee4874e6ed9773c96bd48e899056c4c67801`
- License: MIT; complete notice in `LICENSE.txt`.
- Official asset: `qcamo-v1.0.4.zip`, 1,206,725 bytes.
- Archive SHA-256: `8418131cecfc364a15e1e8f4bfbfbb7be25b2f0cbf0d58063f6c4397a40b0182`.
- Sole payload: `qcamo.asi`, 4,063,819 bytes.
- Payload SHA-256: `4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193`.
- Upstream executable timestamp gate: `0x6980B92F`.

The binary and source tag are unsigned. Matching the official release digest establishes release identity, not independent source-to-binary reproducibility. This package contains executable native code.

## Supported game profile

The manager additionally checks complete SHA-256 identities. A different game build is rejected even if QCamo's weaker timestamp gate matches.

| File | SHA-256 |
| --- | --- |
| METAL GEAR SOLID3.exe | `0d585dcc6a671be5d64d3d0a856c53f9ee0e58e7e4993f76dff29772c7a4bc80` |
| Engine.dll | `4067774bd2945dfab1a81ee0f657b3b6c9414b1b363d29830652e6d93c516996` |
| Renderer.dll | `663199bce1a252861369710d62219a73d2855043ec955ab7e8a926ea13986ac1` |

## Managed standalone loader

Supported release: https://github.com/ThirteenAG/Ultimate-ASI-Loader/releases/tag/v9.7.4

- Official archive: `Ultimate-ASI-Loader-NoPDB_x64.zip`.
- Archive SHA-256: `e5860e7d9a1805267535b65749575b5e406cc6ea3325c7392189c578815045d1`.
- Managed root file: `wininet.dll`, 1,198,304 bytes, renamed from the official `dinput8.dll` using an upstream-supported alias.
- Payload SHA-256: `031a3e5576d91dce1e438d36b9a3d462c7334ab4791990a8ff1e3ddc0e132daf`.

The loader is included as its own `asi-loader` manager package, with complete license and provenance in `LOADER-LICENSE.txt` and `LOADER-UPSTREAM.md`. MGSHDFix is neither required nor installed. The earlier MGSHDFix-specific gate was an overly restrictive manager choice and is superseded. The manager handles loader installation, dependency checks, disable/removal, and recovery; it does not certify other loader versions or custom INI configuration. File identity checks do not prove successful in-game loading.

## Future rebuilds

Use the upstream pinned commit and `flake.lock` with `nix build`; record the toolchain and output SHA-256. ImGui is pinned to `f1cc2ae15e53a861a874c3034aae6798fde194ab`; MinHook to `c3fcafdc10146beb5919319d0683e44e3c30d537`. Investigate binary differences instead of relabeling a rebuilt binary as the official asset. A modified fork must use a distinct version and updated provenance. No such rebuild is claimed for this release.
