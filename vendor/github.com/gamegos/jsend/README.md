# jsend

[![Build Status](https://travis-ci.org/gamegos/jsend.svg?branch=master)](https://travis-ci.org/gamegos/jsend)
[![GoDoc](https://godoc.org/github.com/gamegos/jsend?status.svg)](http://godoc.org/github.com/gamegos/jsend)


Golang **JSend** library


## Installation
```
$ go get github.com/gamegos/jsend
```

## Usage

```go
import "github.com/gamegos/jsend"
```

See [API documentation](http://godoc.org/github.com/gamegos/jsend)

## Format

Jsend is a very simple json format to wrap your json responses.

```json
{
  "status": "success|fail|error",
  "data": {
    "your data": "here..."
  },
  "message": "error message when status is error"
}
```

See [JSend specification](http://labs.omniti.com/labs/jsend) for details.


## License

MIT. See [LICENSE](./LICENSE).
