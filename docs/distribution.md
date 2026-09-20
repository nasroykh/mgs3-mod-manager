# Distribution and tester guidance

## Repository organization

Keep the QCamo and loader integration recipes in this repository for now: `mods/qcamo/`, `mods/asi-loader/`, and `scripts/package-qcamo.ps1`. They depend on the manager's package contract and should be versioned with it. QCamo itself is unchanged upstream software, not a new mod authored by this project.

When original mods or maintained forks need independent code, assets, issues, and releases, give them their own repository. A future `mgs3-mods` repository can hold a small collection of original mods and packaging recipes; a substantial standalone plugin should have its own repository. Do not create nested Git repositories under this checkout or duplicate upstream source merely to distribute an unchanged binary.

Git stores source, manifests, build recipes, documentation, notices, and assets that may be redistributed. Compiled executables and ZIPs belong in release assets, not ordinary source commits. Never publish `.mgs3mod`, game files, extracted original textures, backups, saves, local logs, or crash dumps. Review redistribution rights for each new mod; the two current upstream license notices are included in the package.

## What testers receive

Send the complete `qcamo-1.0.4-standalone-distribution.zip` plus the adjacent checksum file from the current output directory recorded in [validation](asi-loader-validation.md). The archive contains the manager, both import packages, install/remove instructions, both upstream licenses, provenance, and member checksums. Testers need neither Go nor Git and must not copy the author's manager state.

1. Download the complete distribution, verify its SHA-256 against the separately supplied checksum, and extract it outside the game folder.
2. Follow the bundled README, supplying the tester's actual game path with `--game-root` on each command.
3. Run `doctor` first. The manager accepts only the compiled game fingerprints; a different edition or update is not automatically compatible. Do not bypass this guard or replace game files to make it pass.
4. Close the game and launcher before changes. Initialize state if needed, import and enable `asi-loader`, then import and enable `qcamo`.
5. Launch the game and test. Use the documented disable/remove commands to undo the mods. Do not manually delete manager state or downgrade the manager after schema-3 import.

Native ASI plugins execute code inside the game. A matching checksum establishes file identity, not independent security certification. Use trusted release downloads, retain the notices, and do not advise testers to disable security tools if a download is flagged.

## Delivery channel

For the first small trial, privately send the complete ZIP and checksums to selected testers. For broader testing, publish a clearly labeled prerelease on the existing manager repository, attaching the built distribution and checksums. GitHub supports release notes and binary assets; its automatically generated source ZIP is not the installable distribution. See [GitHub's release documentation](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases).

Use a manager/integration release tag distinct from upstream QCamo's version, for example `qcamo-integration-v0.1.0-preview.1`. Identify QCamo 1.0.4 and Ultimate ASI Loader 9.7.4 separately. Any changed distribution must receive a new integration release revision and fresh checksums; do not silently replace an already published asset. Tag the exact source commit used to prepare it and include the tested game fingerprint requirements and known limitations in release notes.

No repository creation, push, tag, or hosted release is performed by these instructions. Choose public versus private visibility explicitly before publishing.

## Test evidence and issue reports

The local user reported successful in-game operation on 2026-09-20. This supports a tester preview, not a claim of universal stability. Ask testers to report menu opening, camo and face-paint changes, repeated swaps, area transitions, cutscenes, keyboard/controller behavior, scoped controls, and other installed mods separately. Known upstream icon and scoped-control reports remain documented in the bundled README.

For issues, collect game/store version, Windows version, distribution checksum, input device, other mods, reproduction steps, and redacted `status --json` and `verify --json` output. Do not request public uploads of saves, full crash dumps, or personal file paths by default.
