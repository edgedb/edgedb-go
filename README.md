# The Go driver for EdgeDB

Requires go 1.11 or higher.

# ⚠️ WIP ⚠️
This project is far from production ready. Contributions welcome! 😊

## Installation
$ go get https://github.com/edgedb/edgedb-go/edgedb

## Basic Usage
Follow the [EdgeDB tutorial](https://edgedb.com/docs/tutorial/index)
to get EdgeDB installed and minimally configured.

```go
package main

import (
  "fmt"
  "log"

  "github.com/fmoor/edgedb-golang/edgedb"
)

func main() {
  opts := edgedb.ConnConfig{Database: "edgedb", User: "edgedb"}
  conn, err := edgedb.Connect(opts)
  if err != nil {
    log.Fatal("could not connect: ", err)
  }
  defer conn.Close()

  var result []string
  err = conn.Query("SELECT 'hello EdgeDB!'", &result)
  if err != nil {
    log.Fatal("error running query: ", err)
  }

  fmt.Println(result)
}
```

## Development

A local installation of EdgeDB is required to run tests.
Download EdgeDB from [here](https://edgedb.com/download)
or [build it manually](https://edgedb.com/docs/internals/dev/).

To run the test suite run `make test` from the `edgedb/` directory.
To run lints `make lint`.
To format the code `make format`.

## License
edgedb-go is developed and distributed under the Apache 2.0 license.
