# Changelog

## [v0.3.6] by 2026-05-20

### Fixed

- `ParseCoord`/`FindCoords` now recognize compact MinDec inputs where degrees
  and decimal-minutes are written as a single concatenated number with no
  separator — common in NAVTEX traffic, e.g. `3630.055N` (36° 30.055′ N) or
  `01202.598E` (12° 02.598′ E). Without this the shape was parsed as a single
  degDec value (`3630.055`) and rejected as out-of-range. The rewrite is
  triggered narrowly: requires a direction letter, no deg-min separator,
  exactly 4 (lat) or 5 (lon) integer digits, a decimal fraction, and the
  implied deg/min parts must satisfy axis bounds and `min < 60`.

### Changed

- Internal: `normalizeCompact`, `normalizeItalianMinFrac`, and
  `normalizeDotDMS` now use precompiled `regexp.Regexp` patterns for their
  trigger logic, matching the style of `normalizeCompactMinDec`. No behavior
  change.

## [v0.3.5] by 2026-05-11

### Fixed

- `FindCoords` no longer reports date-list fragments before month
  abbreviations as coordinates, e.g. `24 25 28 NOV` is not returned as
  `24 25 28 N`.
- `FindCoords` now preserves the full longitude when a latitude ending in
  `N`/`S` is followed by a glued hyphen separator, e.g.
  `41 55.00 N-032 08.00 E` returns `032 08.00 E` instead of the truncated
  `08.00 E`.

## [v0.3.4] by 2026-05-05

### Fixed

- `ParseCoord`/`FindCoords` now accept italian-NAVTEX OCR output where the
  fractional part of minutes is written as 3 digits without a decimal point
  ("44 33 367N" really meaning 44° 33.367′ N). The shape — space-separated
  deg/min/sec where sec is exactly 3 digits with no decimal — is now
  reinterpreted as `min.frac`, so previously rejected inputs (sec ≥ 60) parse
  correctly and `FindCoords` no longer falls back to noise sub-candidates
  ("44 33 367N" used to surface as just `67N` via the v0.3.3 inline backtrack).

### Changed

- **Behavior change** for `<deg> <mm> <ddd>[NSEW]` shape with sec value < 60.
  Previously interpreted as classical DMS (e.g. `44 33 048N` → 44° 33′ 48″ ≈
  44.5633°); now interpreted as italian min.frac (44° 33.048′ ≈ 44.5508°).
  The 3-digit no-decimal sec field with pure-space separators is a reliable
  fingerprint of the italian format; classical DMS keeps working with
  hyphen/`°`/`:` separators or a decimal in the sec field.

## [v0.3.3] by 2026-05-01

### Fixed

- `FindCoords` now recovers the real coordinate when a numeric prefix is
  glued to it by a space, e.g. `235 43-18.0N` returns `43-18.0N` instead of
  the truncated `35 43-18.0N`. Backtracking on a failed candidate now skips
  past the first whitespace inside the consumed span (same shape as the
  v0.3.2 cross-line `\n` recovery, generalized).
- `FindCoords` accepts a coordinate followed by a single non-direction
  letter that is itself followed (optionally after whitespace) by a digit —
  the fingerprint of an OCR glitch wedged between two adjacent coords. So
  `46-20.0NR142-2.0E` now returns both `46-20.0N` and `142-2.0E`. Existing
  unit / adjacent-direction / multi-letter-word rejections are preserved.

## [v0.3.2] by 2026-04-30

### Fixed

- `FindCoords` no longer mismatches the `N`/`S` in distance abbreviations
  like `1 N.M.`, `5 S.M.`, or `10 N MILE` as a coordinate.
- `FindCoords` now recovers a real coordinate that follows a noise number
  separated by a line break, e.g. `BR-117\n55-54.0N` returns `55-54.0N`
  instead of being lost inside the oversized 117° candidate.
- `FindCoords` accepts a fully-specified DMS coord glued to a non-coord
  word (NAVTEX terminator `NNN`, message word `END`, etc.), so
  `124-50-00ENNN` returns `124-50-00E`.

### Changed

- Internal coordinate regex: `\n` between deg and min components is now
  allowed only when bridged by an explicit `-`/`°`/`.` separator (the
  intra-coord linebreak shape `38-\n34.2N`). This affects `CoordRegexSource`
  too and is what enables the cross-line fix above.

## [v0.3.1] by 2026-04-29

### Changed

- No code changes; v0.3.0 re-released under a signed tag.

## [v0.3.0] by 2026-04-29

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
