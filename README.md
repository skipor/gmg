# GoMock Generator

`gmg` type-safe, fast and handy alternative [GoMock](https://github.com/golang/mock) generator.

**Work In Progress!**

## Features

* [Up to 4x times faster](#speed-measures) than [`mockgen` in reflect mode](https://github.com/golang/mock#reflect-mode)
  * `mockgen` builds program depending on your interface and analyse it with `reflect` - that is much extra work!
  * `gmg` loads ast and type info using [go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages) which doesn't require executable build.

* Type-safe: `gomock.Call` wrapped so `Do`, `Return` and `DoAndReturn` arguments are concrete types, but just `args ...interface{}`
  * Autocomplete works perfect!
  * After mock regeneration all type inconsistency in tests are visible in IDE as type check errors.

* Easy to use
  * There are sensible defaults for source package (`.`) and destination (`./mocks`).

    That is, usually, you need only to specify the interface name to mock.

## Install

### **Go version >= 1.16**

`go install github.com/skipor/gmg@latest`

### **Go version < 1.16**

`GO111MODULE=on go get github.com/golang/mock/mockgen@latest`

Please fix a specific version, if you use `gmg` in automation.

## Usage

```
$ gmg --help
gmg is a type-safe, fast and handy alternative GoMock generator. See details at: https://github.com/skipor/gmg

Usage: gmg [--src <package path>] [--dst <file path>] [--pkg <package name>] <interface name> [<interface name> ...]

Flags:
  -d, --dst string   Destination directory or file relative path or pattern.
                     '*' in directory path will be replaced with source package name.
                     '*' in file name will be replaced with snake case interface name.
                     If no file name pattern specified, then '*.go' used by default.
                     Examples:
                        ./mocks
                        ./*mocks
                        ./mocks/*_gomock.go
                        ./mocks_test.go # All mocks will be put to single file.
                      (default "./mocks")
  -p, --pkg string   Package name in generated files.
                     '*' will be replaced with the source package name.
                     Examples:
                        mocks_* # mockgen style
                        *mocks # mockery style
                      (default "mocks_*")
  -s, --src string   Source Go package to search for interfaces. Absolute or relative.
                     Maybe a third-party or standard library package.
                     Examples:
                        .
                        ./relative/pkg
                        github.com/third-party/pkg
                        io
                      (default ".")

```

## Speed measures

* Large interface - 4x faster
  `mockgen --destination ./mocks/interface.go k8s.io/client-go/kubernetes`: 3.022 seconds
  `gmg --src k8s.io/client-go/kubernetes Interface`: 0.741 seconds
* Small interface - 2x faster
  `mockgen --destination ./mocks/writer.go io Writer`: 0.676 seconds
  `gmg --src io Writer`: 0.321 seconds

Measured on MacBook Pro (15-inch, 2017) 4 core i7, 16G with warm `go build` cache.
