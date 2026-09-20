# Runtime plugin packages

Schema 2 extends the existing package format with an explicit absent original state. Schema 1 remains the existing-texture format. Schema 3 is reserved for the pinned standalone ASI loader. A package still contains only `manifest.json` and its mapped `payload/` files; distribute license notices and user documentation in an outer release archive, as the [QCamo recipe](../mods/qcamo/README.md) does.

## Manifest contract

See the complete [QCamo manifest](../mods/qcamo/manifest.json). A plugin mapping uses:

```json
{
  "source": "payload/qcamo.asi",
  "target": "qcamo.asi",
  "originalSha256": "",
  "originalAbsent": true,
  "payloadSha256": "4f518f5d1f53db6041533f22af14f6e95648fd20878d6cd240eed309d00b5193",
  "payloadBytes": 4063819
}
```

The manifest must set `schemaVersion` to `2` and use the compiled supported profile ID. Plugin targets must be canonical, lowercase, root-level `.asi` filenames. Subdirectories, DLL proxies, executables, arbitrary data files, and adoption/replacement of pre-existing plugins are not supported. `originalAbsent` must be true and `originalSha256` empty. Texture entries may also use schema 2 but still require an existing approved `.ctxr` target and its original hash; they cannot claim absence.

Plugin payload checks require an AMD64 PE32+ DLL with bounded section data and a file-backed executable entry point. They establish structural constraints, not authenticity, absence of malicious code, or successful game loading. Texture payloads retain their original hash/size contract; this extension does not add a CTXR decoder. Package authors remain responsible for payload provenance and runtime correctness.

The existing limits still apply: 256 mappings, 64 MiB per payload, 512 MiB total payload, and strict unique canonical paths. `pack` computes or verifies payload hashes and sizes and refuses to overwrite its output. It never executes package content.

Schema 3 is deliberately narrow: ID `asi-loader`, version `9.7.4`, and exactly the single `wininet.dll` mapping in the [loader manifest](../mods/asi-loader/manifest.json). The absent origin, payload length, and SHA-256 must match the supported official release, including during authoring. No arbitrary DLL target, loader version, omitted file, extra payload, or pre-existing file adoption is accepted. Root ASIs must remain schema-2 packages.

## Lifecycle and dependencies

`add` stores a disabled package and rejects an unowned existing target. `enable` requires full core fingerprints, no enabled owner conflict, and the target's original state. ASIs additionally require the enabled manager-owned loader. It records an absent baseline and promotes each staged file only if its destination remains absent. `disable`, `remove`, and `restore --baseline` restore absence instead of inventing empty backup files. Captured absence records are retained like texture baselines.

The standalone loader is official Ultimate ASI Loader 9.7.4 x64, installed as `wininet.dll`; the supported game's engine imports that proxy. See [loader provenance](../mods/asi-loader/UPSTREAM.md). MGSHDFix is not required or installed. Enable the loader package before QCamo; disabling or removing the loader is blocked while any managed ASI is enabled. Disabling/removing QCamo leaves the loader alone and remains possible if loader files drift. Restore refuses unknown changed bytes; it can disable all healthy managed packages in one transaction. The manager does not establish plugin load order or certify coexistence between unrelated native hooks.

The loader is owned by its separate package. Generated logs/crash dumps remain unowned and are preserved. Only default loader configuration is supported: existing proxy/global INI files at the upstream lookup paths block enable and are never edited. There is no runtime download, configuration merging, or dependency updater. Manually installed ASIs can also use the shared loader; leave it installed for those consumers.

## Recovery and compatibility

An absent baseline has no backup blob. Transactions record both file content and absence and copy all required recovery data before applying changes. A surviving apply stage proves that promotion did not occur; recovery must not adopt or delete an identical external file at that destination. Contradictory markers, unexpected bytes, unsafe paths, and ambiguous ownership preserve evidence and block recovery.

Use `recover` for interrupted operations. Do not manually edit journals or delete `.mgs3mod`. The extension reads existing schema-1 and schema-2 history. Older manager binaries cannot decode history containing newer package schemas; executable downgrade is not supported. Installation state is bound to its original root, while packages and the executable can be shared.

Updates currently use disable/remove, add the new package, and enable. Keep the previous distribution available if you need to reinstall its package; do not restore an older manager executable over new journal history.
