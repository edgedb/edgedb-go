# A Golang driver for EdgeDB
This is an unofficial EdgeDB driver for Golang. It is also the only one I know of at the moment.

# ⚠️ WIP ⚠️
This project is far from production ready. Contributions welcome! 😊

## Installation
$ go get https://github.com/fmoor/edgedb-golang/edgedb

## Basic Usage
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

## License
edgedb-golang is developed and distributed under the Apache 2.0 license.
