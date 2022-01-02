# go-perftuner

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]
[![coverage-img]][coverage-url]

Helper tool for manual Go code optimization.

This tool gives you an easy way to get the Go compiler output regarding specific optimisations. Like: function inlining, variable escape and bounds checks.

# Notes

The original implementation was started by [@quasilyte](https://github.com/quasilyte) thanks to him :tada: Than supported by [@cristaloleg](https://github.com/cristaloleg) and now is part of [go-perf](https://github.com/go-perf) organization.

# Installation / Quick Start

```bash
# Install go-perftuner:
$ go get -u github.com/go-perf/go-perftuner

# Check installation (prints help):
$ go-perftuner help

# Run almostInlined sub-command on strings and bytes package:
$ go-perftuner almostInlined strings bytes

# You can use "std" or "..." package name.
# These follow "go build" conventions.
$ go-perftuner almostInlined std
```

# Sub-commands

## almostInlined

Find functions that cross inlining threshold just barely. You may use short command `inl`.

```bash
$ go-perftuner almostInlined -threshold=1 std
almostInlined: std: src/strconv/atof.go:371:6: atof64exact: budget exceeded by 1
almostInlined: std: src/strconv/atof.go:405:6: atof32exact: budget exceeded by 1
almostInlined: std: src/reflect/value.go:1199:6: Value.OverflowComplex: budget exceeded by 1
almostInlined: std: src/vendor/golang_org/x/crypto/cryptobyte/builder.go:77:6: (*Builder).AddUint16: budget exceeded by 1
almostInlined: std: src/crypto/x509/x509.go:1858:58: buildExtensions.func2.1.1: budget exceeded by 1
almostInlined: std: src/crypto/x509/x509.go:1878:58: buildExtensions.func2.3.1: budget exceeded by 1
almostInlined: std: src/crypto/x509/x509.go:1890:58: buildExtensions.func2.4.1: budget exceeded by 1
almostInlined: std: src/crypto/tls/handshake_messages.go:1450:6: (*newSessionTicketMsg).marshal: budget exceeded by 1
almostInlined: std: src/net/http/transfer.go:259:6: (*transferWriter).shouldSendContentLength: budget exceeded by 1
```

## escapedVariables

Find variables that are escaped to the heap. You may use short command `esc`.

```bash
$ go-perftuner escapedVariables fmt
escapedVariables: fmt: src/fmt/format.go:73:13: make(buffer, cap(buf) * 2 + n)
escapedVariables: fmt: src/fmt/format.go:147:14: make([]byte, width)
escapedVariables: fmt: src/fmt/format.go:208:14: make([]byte, width)
```

## boundChecks

Find slice/array that has bound check. You may use short command `bce`.

```bash
$ go-perftuner boundChecks fmt
boundChecks: fmt: src/fmt/format.go:82:16: slice/array has bound checks
boundChecks: fmt: src/fmt/format.go:157:10: slice/array has bound checks
boundChecks: fmt: src/fmt/format.go:159:22: slice/array has bound checks
boundChecks: fmt: src/fmt/format.go:161:10: slice/array has bound checks
```

## License

[MIT License](LICENSE).

[build-img]: https://github.com/go-perf/go-perftuner/workflows/build/badge.svg
[build-url]: https://github.com/go-perf/go-perftuner/actions
[pkg-img]: https://pkg.go.dev/badge/go-perf/go-perftuner
[pkg-url]: https://pkg.go.dev/github.com/go-perf/go-perftuner
[reportcard-img]: https://goreportcard.com/badge/go-perf/go-perftuner
[reportcard-url]: https://goreportcard.com/report/go-perf/go-perftuner
[coverage-img]: https://codecov.io/gh/go-perf/go-perftuner/branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/go-perf/go-perftuner
