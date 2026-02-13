# geoc

`geoc` parses and formats geographic coordinates and points in Go.

## Installation

```bash
go get github.com/o-kos/geoc
```

## Supported coordinate formats

- DMS (degrees, minutes, seconds)
  - `48°33'27"N`
  - `48-33-27 N`
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
- `Point.String()` uses `48-33.0N 048-33.0E`; if formatting fails, it falls back to
  `p.Lat.String() + " " + p.Lon.String()`.

## Examples

See runnable examples in `example_test.go`:

- `ExampleParseCoord`
- `ExampleParsePoint`
- `ExampleCoord_Format`
- `ExampleCoord_String`
- `ExamplePoint_Format`
- `ExamplePoint_String`
