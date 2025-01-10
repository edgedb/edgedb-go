// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gel_test

import (
	"context"
	"log"

	gel "github.com/geldata/gel-go/internal/client"
)

// Transactions can be executed using the Tx() method. Note that queries are
// executed on the Tx object. Queries executed on the client in a transaction
// callback will not run in the transaction and will be applied immediately. In
// gel-go the callback may be re-run if any of the queries fail in a way
// that might succeed on subsequent attempts. Transaction behavior can be
// configured with TxOptions and the retrying behavior can be configured with
// RetryOptions.
func ExampleTx() {
	ctx := context.Background()
	client, err := gel.CreateClient(ctx, gel.Options{})
	if err != nil {
		log.Println(err)
	}

	err = client.Tx(ctx, func(ctx context.Context, tx *gel.Tx) error {
		return tx.Execute(ctx, "INSERT User { name := 'Don' }")
	})
	if err != nil {
		log.Println(err)
	}
}
