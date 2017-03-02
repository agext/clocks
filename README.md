# A Go package enabling time travel (sort of)

[![Release](https://img.shields.io/github/release/agext/clocks.svg?style=flat&colorB=eebb00)](https://github.com/agext/clocks/releases/latest)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/agext/clocks) 
[![Build Status](https://travis-ci.org/agext/clocks.svg?branch=master&style=flat)](https://travis-ci.org/agext/clocks)
[![Coverage Status](https://coveralls.io/repos/github/agext/clocks/badge.svg?style=flat)](https://coveralls.io/github/agext/clocks)
[![Go Report Card](https://goreportcard.com/badge/github.com/agext/clocks?style=flat)](https://goreportcard.com/report/github.com/agext/clocks)

This package provides an abstraction layer for all the time-passage-dependent features from the standard [Go](http://golang.org) time package. This allows code to be written once and run on different clocks, e.g. for testing or for "replay" applications.

## Project Status

v0.1 Edge: Breaking changes to the API unlikely but possible until the v1.0 release. May be robust enough to use in production, though provided on "AS IS" basis. Vendoring recommended.

This package is under active development. If you encounter any problems or have any suggestions for improvement, please [open an issue](https://github.com/agext/clocks/issues). Pull requests are welcome.

## Overview

The `Clock` interface groups all the time-passage-dependent features from the standard time package. A `Timer` and a `Ticker` interface are included to allow abstraction of these concepts.

A "live" Clock is provided in this package giving pass-through access to the standard functionality.

A "manual" Clock is included as a separate package, because it is mostly useful for testing and it is rarely if ever needed in the actual program.


## Installation

```
go get github.com/agext/clocks
```

## License

Package clocks is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
