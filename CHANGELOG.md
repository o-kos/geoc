# Changelog

## [0.2.1] by 2026-02-13

### Added

- New functions:
  - `ParseCoord`
  - `ParsePoint` 
  - `Coord.Format`
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
