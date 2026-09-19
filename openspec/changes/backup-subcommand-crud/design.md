# Design

## Context

The project already has a CLI surface for lifecycle management, and the backup command is still a deferral placeholder. This change introduces a concrete design for backup operations that align with the Valheim save model: world data lives under a configured save directory, a primary world is determined by the configured `world` name, and backup operations must remain safe in multi-world environments.

## Goals / Non-Goals

**Goals:**
- Provide a consistent backup command tree with list, apply, and creation behavior.
- Allow operators to inspect all generated backups and filter by world context.
- Restore a named backup into the current `worlds_local` layout without losing the active primary world definition.
- Preserve chunk data and world metadata from backup history so startup remains consistent.

**Non-Goals:**
- Replacing the game’s own save format or introducing a new storage backend.
- Automatically deleting or pruning older backups during this planning phase.
- Requiring a single-world assumption in the CLI design.

## Decisions

1. Model backups as named snapshots with explicit metadata.
   - Each backup should capture the snapshot name, creation timestamp, and world/save context.
   - This gives the CLI a predictable catalog and makes restore operations deterministic.
   - Alternative considered: storing only a raw directory copy without metadata; rejected because operators need to inspect and filter by world and timestamp.

2. Treat `worlds_local` as the authoritative restore location.
   - Restoring a backup should operate inside the configured save directory under a `worlds_local` subfolder so the game sees the same world layout it expects.
   - The primary world is the folder that matches the configured `world` name, and the restore workflow must preserve that mapping.
   - Alternative considered: restoring backups outside the game save tree; rejected because it would not match Valheim’s startup assumptions and could create an inconsistent world layout.

3. Merge chunk data from backup history into the target world state.
   - The restore flow should include the set of `.chunk` files discovered across prior backup generations, because those files are part of the save continuity needed for startup and world integrity.
   - This design keeps the backup system resilient even if old backups contain data not present in the active save directory.
   - Alternative considered: restoring only the latest folder snapshot; rejected because it discards potentially needed chunk data and makes world recovery less reliable.

4. Design for multi-world operators from the start.
   - Backup listing, filtering, and restore operations should expose enough world metadata to prevent cross-world confusion.
   - This keeps the system usable when several worlds are managed under the same save root.
   - Alternative considered: making the CLI world-agnostic without metadata; rejected because it increases the risk of applying the wrong backup to the wrong world.

5. Keep the backup command surface CRUD-like without over-specifying internal implementation.
   - The command tree should support creating snapshots, listing existing ones, and restoring them explicitly.
   - This preserves a simple operator mental model and leaves room for future retention, cleanup, and archive policies.
   - Alternative considered: creating a highly abstract “backup sync” model; rejected because it hides the real operator actions in a way that is harder to audit or verify.

## Risks / Trade-offs

- [Partial restore state] → Mitigation: restore only after confirming the target world exists and the backup name resolves cleanly; fail without mutating state if the match is not valid.
- [Chunk drift across backup generations] → Mitigation: include relevant `.chunk` files from the backup set while keeping the restore target anchored to the configured primary world.
- [Multi-world confusion] → Mitigation: show backup metadata with world identity and support filtering by world before apply.
- [Accidental overwrite of live saves] → Mitigation: preserve the current save layout and fail fast when the target world or backup is not found.

## Migration Plan

1. Define the backup metadata contract and backup naming model.
2. Add the CLI command tree for `list`, `apply`, and backup creation flows.
3. Implement world-aware catalog rendering and filtering for multiple save contexts.
4. Add restore logic that targets the configured world under `worlds_local` and merges chunk data from relevant backups.
5. Add validation and failure messages for missing backups, missing world directories, and invalid restore targets.
6. Verify the behavior with targeted tests for list output, apply safety, and multi-world handling.

## Open Questions

- Should the backup creation command explicitly use a dedicated subcommand name such as `backup create` or `backup snapshot` for operator clarity?
- Should backup retention and pruning be handled as a separate follow-up capability, or should they be included in the initial backup feature set?
