# Changelog

## 0.2.0 - 2020-12-20

### Changed

- Containerized and deployed to Cloud Run.
- Update dependencies.

## 0.1.1 - 2020-09-06

### Fixed

- Bug: `goki` now writes back JSON files with all changes to prevent data loss on crash.

### Changed

- Update dependencies.
- Refactoring: Get-like methods returns copies instead of references to prevent unintended data changes.

## 0.1.0 - 2020-09-06

### Added

- `goki` server application with HTML views.
- Datastore using JSON files.
