# Proposal

## Why

The project’s current `backup` command is only a placeholder, so operators have no supported workflow to inspect, restore, or create server backups in a way that matches Valheim’s actual save layout. This leaves the project unable to manage multiple world states, backup retention, or disaster recovery in a repeatable way.

## What Changes

- Add a real backup capability with a CRUD-like command surface under `backup`.
- Support `backup list` to show generated backups and their timestamps for easy inspection.
- Support `backup apply <name>` to restore a named backup into the configured world save structure.
- Support a named backup creation command to snapshot all saves outside the normal save directory and keep the operational state explicit.
- Treat the configured `world` as the primary world while preserving the `worlds_local` layout and merging `.chunk` files from any prior backup set without breaking startup.
- Support multi-world scenarios by listing, filtering, and targeting backups by world context as part of the backup catalog.

## Capabilities

### New Capabilities
- `backup-management`: defines the backup catalog, restore process, and snapshot workflow for Valheim save data across multiple worlds and backup generations.

### Modified Capabilities
- None.

## Impact

- Affects the CLI surface for backup behavior and the operational model used by the Valheim server lifecycle.
- Requires new backup metadata, save-path handling, and restore logic that understand the `worlds_local` directory and `.chunk` files.
- Introduces a clearer operator workflow for listing, restoring, and creating backups without changing the existing server startup contract in the planning phase.
- Establishes the foundation for future retention, filtering, and restore safety policies.
