# Recovery details

Originals and journal records live in `.mgs3mod/` under the game root. Do not edit or delete them. Close the game and launcher before recovery.

Run `mgs3mod recover --game-root $game --dry-run --json`, inspect the report, then run without `--dry-run` if safe. An uncommitted operation rolls back only when targets match recorded before/after states. A committed operation finishes cleanup. Contradictory records and external changes block recovery.

A missing target might have been deleted externally. Plain recovery refuses to recreate it. If unresolved apply intent and a verified before snapshot justify restoring that exact file, use `mgs3mod recover --game-root $game --restore-missing textures/flatlist/_win/example.ctxr`. Repeat the flag for each justified missing target. This cannot overwrite existing files and requires matching core fingerprints. Reuse the same target list if this recovery is interrupted.

Multi-file changes are not simultaneously visible. The manager cannot prevent concurrent game launches, power loss, media corruption, or malicious external edits. Promoted originals, resolved journal records, and recovery snapshots are retained; no baseline garbage collection is provided. Damaged state may require investigation rather than automatic repair.
