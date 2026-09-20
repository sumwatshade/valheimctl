# Spec Delta

## Purpose

This delta tightens the backup-management contract around a testable filesystem boundary and the actual service behavior already implemented. It makes the backup service deterministic in tests, keeps the filesystem abstraction local to the service implementation, and validates the backup operations against realistic save-root layouts without depending on a live Valheim installation.

## MODIFIED Requirements

### Requirement: Backup operations use the configured save root and world directory semantics
The system SHALL use the configured save root and the target world directory semantics as the authoritative inputs for backup discovery and restore targeting, rather than relying on hard-coded OS assumptions or implicit host-level paths.

#### Scenario: Backup creation resolves the configured save tree
- **WHEN** the operator runs a backup creation flow
- **THEN** the system SHALL resolve the backup root from the configured save directory and copy the relevant world files beneath the backup storage structure

#### Scenario: Backup restore targets the configured world directory
- **WHEN** the operator applies a backup and the configuration names a world such as `Dedicated` or another explicit world
- **THEN** the system SHALL resolve the matching world directory and overwrite only that target directory, leaving unrelated world directories untouched

### Requirement: The backup service exposes a narrow filesystem dependency
The system SHALL allow the backup service constructor to accept an injected filesystem implementation as an internal detail, while defaulting to the real OS filesystem when no override is supplied.

#### Scenario: Production uses the OS filesystem by default
- **WHEN** a service instance is created without a custom filesystem dependency
- **THEN** the system SHALL use the operating system-backed implementation and preserve the existing runtime behavior

#### Scenario: Tests replace the filesystem implementation
- **WHEN** a test constructs the backup service with a custom filesystem implementation
- **THEN** the system SHALL use that injected implementation for all read, write, and directory operations in the service contract

### Requirement: Read operations use standard-library filesystem interfaces where appropriate
The system SHALL use the standard-library `io/fs` interfaces for read operations such as listing directories, reading files, and checking file metadata when they match the service behavior, while keeping the write and mutation operations explicit in the internal contract.

#### Scenario: Read behavior is expressed through standard fs interfaces
- **WHEN** the backup service reads directory entries or file content
- **THEN** the system SHALL operate through the standard `io/fs` contract rather than directly depending on OS-specific helper functions

#### Scenario: Writable filesystem actions remain explicit
- **WHEN** the backup service creates directories, writes metadata, or removes current data before restoring a backup
- **THEN** the system SHALL use the custom mutation methods required for the service contract rather than a read-only filesystem abstraction

### Requirement: Metadata and backup listing remain resilient to partial filesystem state
The system SHALL list available backups and their metadata by reading the actual backup directories present on disk, using directory timestamps and available file state as fallbacks when metadata is missing or partial.

#### Scenario: Listing backups reads disk-backed metadata
- **WHEN** the operator lists backups from a valid backup root
- **THEN** the system SHALL read each backup record from disk, include the backup name, and populate metadata such as timestamp and world list from the real directory state

#### Scenario: Metadata falls back gracefully when `meta.json` is incomplete
- **WHEN** a backup directory exists but its metadata file is absent or incomplete
- **THEN** the system SHALL fall back to directory timestamps and available file state rather than failing the listing workflow

### Requirement: Applying a backup intentionally replaces the current target world directory
The system SHALL remove or replace the directory that matches the configured world name before copying the backup payload into the target location so that restore semantics are explicit and deterministic.

#### Scenario: Restore writes into the explicitly named world directory
- **WHEN** the operator restores a backup for the configured world
- **THEN** the system SHALL remove the current directory named with the configured world and copy the backup state into that exact world directory

#### Scenario: An unrelated world directory is not overwritten
- **WHEN** the backup applies to one configured world while other world directories exist
- **THEN** the system SHALL limit the overwrite to the configured world directory and not rename, replace, or delete unrelated save directories

### Requirement: Service behavior is validated with filesystem-backed tests
The system SHALL provide service-level tests that exercise realistic save-root layouts and synthetic filesystem trees, ensuring backup create/list/apply flows and missing-backup failures are validated without relying on a live Valheim installation or environment-specific state.

#### Scenario: Tests exercise realistic disk layouts
- **WHEN** the backup service is tested against fabricated but realistic save trees and backup directories
- **THEN** the system SHALL validate the expected file operations without requiring a live Valheim installation or shared external state

#### Scenario: A synthetic filesystem is used for reproducible backup scenarios
- **WHEN** a test harness builds a `fstest.MapFS` tree that includes world folders, backup directories, and metadata files
- **THEN** the system SHALL exercise the same backup logic as production and verify behavior through the same service contract

#### Scenario: Invalid file states produce deterministic failures
- **WHEN** a backup directory is missing or intentionally absent
- **THEN** the system SHALL return explicit errors and preserve the existing save state rather than silently creating an inconsistent result
