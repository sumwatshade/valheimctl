# Tasks

## 1. Backup contract and CLI surface

- [ ] 1.1 Define the backup command tree and naming model for `list`, `apply`, and create/snapshot operations, and verify the command help reflects the intended operator flow.
- [ ] 1.2 Define backup metadata fields for name, timestamp, save context, and world identity, and verify the data model covers multi-world environments.

## 2. Catalog and listing behavior

- [ ] 2.1 Implement backup listing that shows each saved snapshot with timestamp and world context, and verify the output includes all available backups.
- [ ] 2.2 Add filtering by world or save context so operators can narrow the backup catalog in multi-world environments, and verify results are restricted to matching entries.
- [ ] 2.3 Add empty-state handling for no backups, and verify the CLI returns a clear message without failing unexpectedly.

## 3. Restore safety and world recovery

- [ ] 3.1 Implement the `backup apply <name>` workflow to resolve the named backup and matching world under `worlds_local`, and verify the target world is identified before changes begin.
- [ ] 3.2 Preserve the configured primary world mapping and guard against invalid or missing restore targets, and verify the current save layout remains unchanged when the restore is rejected.
- [ ] 3.3 Merge relevant `.chunk` files from backup history into the targeted save state, and verify startup remains consistent with the restored world state.

## 4. Backup creation workflow

- [ ] 4.1 Implement the create/snapshot flow for generating a larger backup of all relevant saves outside the standard save directory, and verify the named backup is created with a timestamp and metadata record.
- [ ] 4.2 Include all relevant world data and chunk files in the backup bundle, and verify the archive contains the complete saved world state needed for later restore.
- [ ] 4.3 Support multiple worlds in the same save root during creation, and verify each world’s metadata remains associated with the correct backup entry.

## 5. Verification and acceptance

- [ ] 5.1 Run focused tests for backup listing, restore safety, and multi-world behavior, and verify they pass for valid and invalid backup scenarios.
- [ ] 5.2 Verify the command output and failure messages are clear to operators, and confirm the workflow matches the spec requirements for listing, applying, and creating backups.
