# Development

[Go](https://golang.org/) version 1.11 or later is needed for developing and testing the application.

If you're setting up a Go environment from scratch, add Go's `bin`
directory to your `PATH`, so that go binaries can be found. In
`.profile` add the following line.
```
export PATH=$PATH:$(go env GOPATH)/bin
```

DSK uses `go mod` to manage and vendor its dependencies. When using
Go 1.11 module support must be explictly enabled, add this line to
`.profile`:
```
export GO111MODULE=on
```

The `make dev` command assumes your test design system definitions are
below the `example` directory.

```
$ go get github.com/atelierdisko/dsk
$ cd $(go env GOPATH)/src/github.com/atelierdisko/dsk
$ make dev
```

To run the unit tests use `make test`, to run the benchmarks use `make bench`,
for performance profiling run `make profile`.

When updating the files of the built-in frontend, the file `frontend_vfsdata.go`
also needs to be remade. Run `make frontend_vfsdata.go` to do so. The file is 
provided to make DSK go gettable.

## Known Limitations

- Symlinks inside the design definitions tree are not supported
- Not tested thoroughly on Windows


