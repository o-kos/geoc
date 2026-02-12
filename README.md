# geoc - Golang geographic coordinates converter

Provide conversion from popular string representation of
geographic coordinates to golang-native float64 format.

## Supported formats

`geoc` support coordinate in the three basic forms:

- Degrees (integer), minutes (integer), and seconds (integer, or real number) (DMS).
  - `48°33'27"N`,
  - `48-33-27 N`,
  - `120-5749E` (compact `MMSS`, i.e. `120°57'49"E`),
  - `48-33-26.9604N`, etc.
- Degrees (integer) and minutes (real number) (MinDec).
  - `48-33N`,
  - `48°33.4493'N`,
  - `48-33.49128N`, etc.
- Degrees (real number) (DegDec).
  - `48.557489`,
  - `+48.557489`,
  - `-39.298358`, etc.

## StringToPoint format matching

`StringToPoint(lat, lon)` accepts latitude and longitude in the same format class:

- DMS with DMS,
- MinDec with MinDec,
- DegDec with DegDec.

Exact textual representation may differ inside one class (for example,
`48-33-27N` and `120-5749E` are both treated as DMS and are accepted together).

### Installation

Once you have [installed Go][golang-install], run this command
to install the `geoc` package:

    go get github.com/o-kos/geoc

[golang-install]: http://golang.org/doc/install.html
