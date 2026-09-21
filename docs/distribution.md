# Distribution and tester guidance

## Repository organization

Keep the QCamo and loader integration recipes in this repository for now: `mods/qcamo/`, `mods/asi-loader/`, and `scripts/package-qcamo.ps1`. They depend on the manager's package contract and should be versioned with it. QCamo itself is unchanged upstream software, not a new mod authored by this project.

When original mods or maintained forks need independent code, assets, issues, and releases, give them their own repository. A future `mgs3-mods` repository can hold a small collection of original mods and packaging recipes; a substantial standalone plugin should have its own repository. Do not create nested Git repositories under this checkout or duplicate upstream source merely to distribute an unchanged binary.

Git stores source, manifests, build recipes, documentation, notices, and assets that may be redistributed. Compiled executables and ZIPs belong in release assets, not ordinary source commits. Never publish `.mgs3mod`, game files, extracted original textures, backups, saves, local logs, or crash dumps. Review redistribution rights for each new mod; the two current upstream license notices are included in the package.

## What testers receive

Send the complete `qcamo-1.0.4-standalone-distribution.zip` plus `checksums.txt` from [v0.3.0 release assets](https://github.com/nasroykh/mgs3-mod-manager/releases/tag/v0.3.0). The archive contains the manager, both import packages, install/remove instructions, upstream licenses, manager dependency notices, provenance, and member checksums. Testers need neither Go nor Git and must not copy the author's manager state. Local pre-release artifacts in the historical validation record are superseded by the versioned release assets.

For Crouch Walk, share the manager release and [local-import instructions](crouch-walk.md), not the mod archive or generated package. Each tester obtains the original Nexus download themselves. The adapter is included in v0.2.0 and later; Crouch Walk payloads are excluded from all release assets. Direct launch is included in v0.3.0. Never send the v0.1.0 manager to someone with schema-4 history.

1. Download the complete distribution, verify its SHA-256 against the separately supplied checksum, and extract it outside the game folder.
2. Follow the bundled README, supplying the tester's actual game path with `--game-root` on each command.
3. Run `doctor` first. The manager accepts only the compiled game fingerprints; a different edition or update is not automatically compatible. Do not bypass this guard or replace game files to make it pass.
4. Close the game and launcher before changes. Initialize state if needed, import and enable `asi-loader`, then import and enable `qcamo`.
5. Launch the game and test. Use the documented disable/remove commands to undo the mods. Do not manually delete manager state or downgrade the manager after schema-3 import.

Native ASI plugins execute code inside the game. A matching checksum establishes file identity, not independent security certification. Use trusted release downloads, retain the notices, and do not advise testers to disable security tools if a download is flagged.

## Delivery channel

The first local trial succeeded and the user authorized v0.1.0 publication. Share the release asset link rather than a development folder. GitHub's automatic source ZIP is not the installable distribution. See [release process](releasing.md) and [GitHub release documentation](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases). Future experimental versions can be marked prerelease; the default latest installer selects non-prereleases.

Manager release v0.1.0 is distinct from upstream QCamo 1.0.4 and Ultimate ASI Loader 9.7.4. Changed distributions require a new release and fresh checksums; never silently replace published assets. Tag the reviewed source commit and include fingerprint requirements and known limitations.

No repository creation, push, tag, or hosted release is performed by these instructions. Choose public versus private visibility explicitly before publishing.

## Test evidence and issue reports

Use the [tester guide](tester-guide.md) for the manual checklist and reporting safeguards. A [shareable invitation draft](tester-invitation.md) is available for an explicitly chosen audience; creating the draft does not send it. GitHub provides separate compatibility and bug-report templates. Keep untested cases distinct from passes.

The local user reported successful in-game operation on 2026-09-20. This supports a tester preview, not a claim of universal stability. Ask testers to report menu opening, camo and face-paint changes, repeated swaps, area transitions, cutscenes, keyboard/controller behavior, scoped controls, and other installed mods separately. Known upstream icon and scoped-control reports remain documented in the bundled README.

For issues, collect game/store version, Windows version, distribution checksum, input device, other mods, reproduction steps, and redacted `status --json` and `verify --json` output. Do not request public uploads of saves, full crash dumps, or personal file paths by default.
