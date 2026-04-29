# Changelog

## [v0.3.0] — unreleased

### Added

- `FindCoords(s string, opts ...FindOption) []Match` for locating coordinate
  substrings in arbitrary text (free-form descriptions, message bodies).
- `Match` struct with `Start`, `End`, `Text`, `Coord` fields.
- `FindOption` constructors: `RequireDirection`, `OnlyLat`, `OnlyLon`.
- `CoordRegexSource() string` returning the non-capturing regex source for
  one coordinate, suitable for embedding into user-defined patterns.

### Changed

- Internal coordinate regex refactored into shared building blocks
  (`rxNum`, `rxDegSep`, `rxMinSep`, `rxSecSep`) plus a non-capturing core
  reused by `CoordRegexSource`. `ParseCoord`/`ParsePoint` behavior is
  unchanged.

## [v0.2.2] by 2026-02-25

### Changed

- Support new formats for coordinate parsing:
  - `48.33.27N` (DMS with dots as separators)
  - `120.57.49E` (DMS with dots as separators, equivalent to `120°57'49"E`)

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
