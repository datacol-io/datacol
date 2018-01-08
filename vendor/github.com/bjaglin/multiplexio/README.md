multiplexio
===========

Go library providing I/O wrappers for aggregating streams.

[![Build Status](https://travis-ci.org/bjaglin/multiplexio.svg?branch=master)](https://travis-ci.org/bjaglin/multiplexio)
[![GoDoc](https://godoc.org/github.com/bjaglin/multiplexio?status.svg)](https://godoc.org/github.com/bjaglin/multiplexio)

## Usage

See [godoc](http://godoc.org/github.com/bjaglin/multiplexio). For usage examples, the best is currently to check the project's [tests](https://github.com/bjaglin/multiplexio/blob/master/multiplexio_test.go).

## Versioning policy

The API defined by this package is still considered unstable. However, until a fully backward-compatible v1 branch is created, the only changes in the existing API will be new configuration or customization fields. Therefore, as long as you initialize all structs declared in this project using explicit names, you will will not be affected by them.

For example, instead of:

    Source{reader1}

Prefer:

    Source{Reader: reader1}
