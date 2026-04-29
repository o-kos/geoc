# geoc

`geoc` parses and formats geographic coordinates and points in Go.

[![CI](https://github.com/o-kos/geoc/actions/workflows/ci.yml/badge.svg)](https://github.com/o-kos/geoc/actions/workflows/ci.yml)
[![CodeQL](https://github.com/o-kos/geoc/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/o-kos/geoc/actions/workflows/github-code-scanning/codeql)

## Installation

```bash
go get github.com/o-kos/geoc
```

## Supported coordinate formats

- DMS (degrees, minutes, seconds)
  - `48°33'27"N`
  - `48-33-27 N`
  - `48.33.27N`
  - `120-5749E` (compact `MMSS`, equivalent to `120°57'49"E`)
  - `48-33-26.9604N`
- MinDec (degrees and decimal minutes)
  - `48-33N`
  - `48°33.4493'N`
  - `48-33.49128N`
- DegDec (decimal degrees)
  - `48.557489`
  - `+48.557489`
  - `-39.298358`

## Public API

- Parse coordinate:
  - `ParseCoord(s string) (Coord, error)`
- Parse point:
  - `ParsePoint(s string) (Point, error)`
- Find coordinates inside arbitrary text:
  - `FindCoords(s string, opts ...FindOption) []Match`
  - Options: `RequireDirection`, `OnlyLat`, `OnlyLon`
- Embed the coordinate regex into your own patterns:
  - `CoordRegexSource() string`
- Format coordinate by example:
  - `Coord.Format(example string) (string, error)`
- Format point by examples:
  - `Point.Format(latFmt, lonFmt, separator string) (string, error)`
- Default string views:
  - `Coord.String() string`
  - `Point.String() string`
- Location enum:
  - `LocNone`, `LocLat`, `LocLon`

Deprecated wrappers are still available:

- `StringToCoord`
- `StringToPoint`

## Notes

- `ParsePoint` expects latitude and longitude in compatible format classes:
  - DMS with DMS
  - MinDec with MinDec
  - DegDec with DegDec
- Exact textual representation inside one class may differ.
  - Example: `48-33-27N` and `120-5749E` are both DMS and can be parsed together.
- `Coord.String()` uses defaults:
  - latitude: `48-33.0N`
  - longitude: `048-33.0E`
  - unspecified location: decimal degrees
- `Point.String()` uses `48-33.0N 048-33.0E`.

## Finding coordinates in text

`FindCoords` scans an arbitrary string and returns every coordinate
substring it can parse, in order of appearance. Useful for free-form
inputs such as NAVTEX message bodies:

```go
text := "WARNING 175/22 - mines reported near 39 27 25.55N - 009 39 25E, " +
    "stay clear within 3 NM."

for _, m := range geoc.FindCoords(text, geoc.RequireDirection()) {
    fmt.Printf("%d-%d %s %q\n", m.Start, m.End, m.Coord.Loc, m.Text)
}
// Output:
// 37-49 Lat "39 27 25.55N"
// 52-62 Lon "009 39 25E"
```

The `175/22` message number, surrounding noise, and `3 NM` (nautical
miles) are correctly skipped.

`RequireDirection` skips matches without an `N/S/E/W` letter (dates,
message numbers, distances). `OnlyLat` / `OnlyLon` further restrict to
one axis. To embed the same coordinate regex into your own pattern, use
`CoordRegexSource()` — it returns a non-capturing source string with no
surrounding whitespace.

## Benchmarks

```bash
go test -bench=. -benchmem ./...
```

## Examples

See runnable examples in `example_test.go`:

- `ExampleParseCoord`
- `ExampleParsePoint`
- `ExampleCoord_Format`
- `ExampleCoord_String`
- `ExamplePoint_Format`
- `ExamplePoint_String`
