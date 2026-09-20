# Managed standalone ASI loader plan

Date: 2026-09-20. This extends the QCamo integration at live state generation 10.

## Goal

Install and manage QCamo and a standalone ASI loader entirely through MGS3 Mod Manager. Do not install MGSHDFix, its plugin, configuration tool, settings, or graphics/gameplay changes.

## Design

1. Pin an official x64 Ultimate ASI Loader release and independently verify its archive, binary, license, supported proxy filenames, and MGS3 import paths.
2. Represent the loader as a separate `asi-loader` package. Reserve schema 3 for an exact supported loader identity and target set; do not allow arbitrary root DLL replacement. Keep schema-1 texture and schema-2 ASI packages unchanged.
3. Reuse the existing absent-baseline transaction path for loader creation, removal, and crash recovery. Existing root DLLs remain unowned and must not be overwritten or silently adopted.
4. Require an enabled, manager-owned loader with correct on-disk hashes before enabling an ASI. Refuse loader disable/removal while any ASI is enabled. Disabling/removing QCamo leaves the shared loader installed; `restore --baseline` may disable all managed packages together.
5. Package the loader and QCamo separately inside one shareable distribution with the manager executable, both licenses, provenance, install/uninstall instructions, and checksums. No runtime downloads or installer scripts inside packages.
6. Run regression, dependency, conflict, interrupted loader and combined loader/ASI restore, actual-artifact lifecycle, and reproducibility checks. Independently review changes before deployment.
7. Deploy the manager and install/enable the standalone loader and QCamo through normal manager commands. Verify installed hashes and ownership. Gameplay remains a separate observed test; do not claim it from file checks.

## Scope and limits

The first supported loader release is pinned rather than accepting arbitrary proxy DLLs. Automatic loader updates, arbitrary dependency graphs, load-order editing, and cleanup of runtime-generated files remain out of scope. Loader configuration that disables root ASIs must not be silently ignored when reporting readiness. Old manager binaries cannot read schema-3 history; downgrading is unsupported.

The earlier external-MGSHDFix-loader prerequisite is superseded. Previous artifacts remain historical; the new standalone-loader distribution will use a distinct output directory and name.
