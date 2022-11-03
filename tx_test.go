package edgedb_test

import (
	"context"
	"log"

	edgedb "github.com/edgedb/edgedb-go/internal/client"
)

// Transactions can be executed using the Tx() method. Note that queries are
// executed on the Tx object. Queries executed on the client in a transaction
// callback will not run in the transaction and will be applied immediately. In
// edgedb-go the callback may be re-run if any of the queries fail in a way
// that might succeed on subsequent attempts. Transaction behavior can be
// configured with TxOptions and the retrying behavior can be configured with
// RetryOptions.
func ExampleTx() {
	ctx := context.Background()
	client, err := edgedb.CreateClient(ctx, edgedb.Options{})

	err = client.Tx(ctx, func(ctx context.Context, tx *edgedb.Tx) error {
		return tx.Execute(ctx, "INSERT User { name := 'Don' }")
	})

	if err != nil {
		log.Println(err)
	}
}
