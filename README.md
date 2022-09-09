# The Go driver for EdgeDB

[![Build Status](https://github.com/edgedb/edgedb-go/workflows/Tests/badge.svg?event=push&branch=master)](https://github.com/edgedb/edgedb-go/actions)
[![Join GitHub discussions](https://img.shields.io/badge/join-github%20discussions-green)](https://github.com/edgedb/edgedb/discussions)

## Installation
In your module directory run

```bash
$ go get github.com/edgedb/edgedb-go
```

## Basic Usage

Follow the [EdgeDB tutorial](https://www.edgedb.com/docs/guides/quickstart)
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
	ctx := context.Background()
	client, err := edgedb.CreateClient(ctx, edgedb.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var result string
	err = client.QuerySingle(ctx, "SELECT 'hello EdgeDB!'", &result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```

## Development

A local installation of EdgeDB is required to run tests.
Download EdgeDB from [here](https://www.edgedb.com/download)
or [build it manually](https://www.edgedb.com/docs/reference/dev).

To run the test suite run `make test`.
To run lints `make lint`.

## License

edgedb-go is developed and distributed under the Apache 2.0 license.
