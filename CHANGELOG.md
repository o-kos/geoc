# Changelog

## [v0.2.1] by 2026-02-16

### Changed

- Enhance coordinate formatting to handle out-of-range values and apply sign consistently


## [v0.2.0] by 2026-02-16

### Added

- New functions:
  - `ParseCoord`
  - `ParsePoint` (closes #2)
  - `Coord.Format` (closes #5)
  - `Coord.String`
  - `Point.Format`
  - `Point.String`
- New table-driven tests for:
  - `Coord.Format` (split into positive and negative scenarios)
  - `Point.Format`
  - `Coord.String`
  - `Point.String`
- Additional branch/error coverage tests for internal parsing helpers.
- Go examples (`example_test.go`) for all exported non-deprecated APIs.

### Changed

- Improved package and API documentation comments in `geoc.go` to better match actual behavior and `godoc` expectations.
- Deprecate functions:
  - `StringToCoord`
  - `StringToPoint`
