# Tasks

## 1. Backup contract and configuration binding

- [ ] 1.1 Define the save-root and world-target contract from config values and verify the backup service resolves the configured `savedir` and `world` before any file operation. Use filesystem-backed temp directories to confirm the selected target is derived from configuration and not a guessed folder.
- [ ] 1.2 Add validation for missing or invalid save-root states and verify the service returns a deterministic error when the configured save directory does not exist or is unusable.

## 2. Metadata discovery and listing

- [ ] 2.1 Implement backup listing by reading the actual backup directories on disk and verify that each directory becomes a backup record with a name and timestamp even when `meta.json` is absent.
- [ ] 2.2 Add fallback metadata behavior for partial or stale `meta.json` and verify that the list output still includes the backup and uses a valid timestamp fallback.
- [ ] 2.3 Confirm the output includes the world and metadata context expected by the backup catalog and verify the ordering is deterministic when multiple backup directories exist.

## 3. Chunk reconstruction across backups

- [ ] 3.1 Implement chunk collection across the backup set, including cases where some backups contain no chunk files and others do, and verify the result is the union of all relevant `.chunk` data.
- [ ] 3.2 Add partial-history handling so a restore is not allowed to silently drop required chunk files and verify the service reports the missing state explicitly when the dataset is incomplete.
- [ ] 3.3 Validate that the chunk set is rebuilt only for the relevant world context and not across unrelated world directories, using a mock filesystem with multiple worlds present.

## 4. Safe restore behavior

- [ ] 4.1 Implement `backup apply` so it resolves the configured world directory by name, removes the current directory, and then copies the selected backup contents into that exact location. Verify the overwrite target is the configured world name and not an unrelated folder.
- [ ] 4.2 Ensure unrelated world directories remain untouched during the apply flow and verify that only the target world directory is replaced in a multi-world filesystem simulation.
- [ ] 4.3 Add negative-path tests for missing backups and missing world directories and verify the service leaves the current disk state unchanged when a restore cannot be completed.

## 5. Rigorous filesystem test suite

- [ ] 5.1 Add unit tests that use temporary directory trees and mock-style filesystem helpers to simulate valid backups, partial metadata, missing chunks, and multi-world layouts.
- [ ] 5.2 Add regression tests for create/list/apply flows covering nested save directories, file-level chunk operations, and directory replacement semantics, and verify the expected behavior is reproduced in each case.
- [ ] 5.3 Run the focused Go backup test suite and verify all backup file-operation tests pass without relying on shared or live environment state.
