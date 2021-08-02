# The Go driver for EdgeDB
[![Build Status](https://github.com/edgedb/edgedb-go/workflows/Tests/badge.svg?event=push&branch=master)](https://github.com/edgedb/edgedb-go/actions)
[![Join GitHub discussions](https://img.shields.io/badge/join-github%20discussions-green)](https://github.com/edgedb/edgedb/discussions)

## Installation
$ go get https://github.com/edgedb/edgedb-go

## Basic Usage
Follow the [EdgeDB tutorial](https://edgedb.com/docs/tutorial/index)
to get EdgeDB installed and minimally configured.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/edgedb/edgedb-go"
)

func main() {
	opts := edgedb.Options{
		Database: "edgedb",
		User: "edgedb",
		MinConns: 1,
		MaxConns: 4,
	}

	ctx := context.Background()
	pool, err := edgedb.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var result string
	err = pool.QuerySingle(ctx, "SELECT 'hello EdgeDB!'", &result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```

## Development

A local installation of EdgeDB is required to run tests.
Download EdgeDB from [here](https://edgedb.com/download)
or [build it manually](https://edgedb.com/docs/internals/dev/).

To run the test suite run `make test`.
To run lints `make lint`.

## License
edgedb-go is developed and distributed under the Apache 2.0 license.
