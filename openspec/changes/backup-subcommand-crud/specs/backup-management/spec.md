# Spec Delta

## Purpose

This capability adds a backup-management workflow for Valheim saves so operators can inspect existing backup sets, restore a named world snapshot safely, and create a broader archive of all world saves without losing the active save layout or world metadata.

## ADDED Requirements

### Requirement: Backup catalog exposes saved snapshots and their metadata
The system SHALL provide a backup listing workflow that shows each generated backup name, the timestamp it was created, and the associated world context so operators can inspect and filter available snapshots.

#### Scenario: Listing backups shows generated snapshots
- **WHEN** the operator runs `backup list`
- **THEN** the system SHALL return each backup name, creation timestamp, and the world identity or save context associated with it

#### Scenario: Listing backups supports filtering by world context
- **WHEN** multiple worlds are present and an operator filters the backup catalog by a world name or identifier
- **THEN** the system SHALL show only the matching backup entries and omit unrelated saves from the result

#### Scenario: No backups exist
- **WHEN** the operator runs `backup list` before any backup has been created
- **THEN** the system SHALL report a clear empty-state message rather than failing silently

### Requirement: Named backup restore applies the correct world state
The system SHALL provide a restore workflow that applies a named backup to the configured world save location while preserving the current world directory structure and the primary world definition.

#### Scenario: Restoring a backup targets the configured primary world
- **WHEN** the operator runs `backup apply <name>` for an existing backup
- **THEN** the system SHALL locate the matching world under `worlds_local`, restore the named snapshot into that world context, and preserve the configured primary world name as the active save target

#### Scenario: Restore is rejected when a backup or world is missing
- **WHEN** the operator requests a backup restore for a name that does not exist or a world folder cannot be matched
- **THEN** the system SHALL stop the restore, report the missing backup or world, and leave the current save layout unchanged

#### Scenario: Restore copies world chunk data from backup history
- **WHEN** a backup contains `.chunk` files that are relevant to the selected world
- **THEN** the system SHALL copy those chunk files into the restored save structure so startup can continue without losing the world state represented by earlier backup sets

### Requirement: Backup creation captures all relevant save data outside the standard save directory
The system SHALL provide a workflow to create a larger backup of all tracked saves, storing it outside the standard save directory while preserving the metadata needed to restore it later.

#### Scenario: Creating a named backup stores all saves and their world context
- **WHEN** the operator creates a backup with a defined name
- **THEN** the system SHALL capture the full set of relevant save data, preserve the timestamp and world metadata, and store the snapshot in a clearly named backup location

#### Scenario: Backup creation includes all world chunk files from the active save set
- **WHEN** the backup includes multiple world folders or a set of `.chunk` files across the active save state
- **THEN** the system SHALL include all relevant chunk files in the backup so a restore can rehydrate the world without leaving the save incomplete

#### Scenario: Multi-world environments are supported during backup creation
- **WHEN** the system is managing more than one world at a time
- **THEN** the backup creation flow SHALL include the relevant world metadata needed to list, filter, and restore the correct world snapshots without mixing unrelated worlds together
