// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
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

package bench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/edgedb/edgedb-go/edgedb"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

var (
	client *edgedb.Client
	ids    randIDs
	ctx    context.Context = context.TODO()
)

type randIDs struct {
	Users  []types.UUID `edgedb:"users"`
	Movies []types.UUID `edgedb:"movies"`
	People []types.UUID `edgedb:"people"`
}

func server() (opts edgedb.Options, err error) {
	credFileName, ok := os.LookupEnv("EDGEDB_CREDENTIALS_FILE")
	if !ok {
		return opts, errors.New("EDGEDB_CREDENTIALS_FILE env variable not set")
	}

	data, err := ioutil.ReadFile(credFileName)
	if err != nil {
		return opts, fmt.Errorf("reading credentials failed: %q", credFileName)
	}

	err = json.Unmarshal(data, &opts)
	if err != nil {
		return opts, fmt.Errorf("parsing credentials failed: %q", credFileName)
	}

	opts.Database = "edgedb_bench"
	return opts, nil
}

func getIDs() (rids randIDs, err error) {
	err = client.QueryOne(
		ctx,
		`
		WITH
			U := User {id, r := random()},
			M := Movie {id, r := random()},
			P := Person {id, r := random()}
		SELECT (
			users := array_agg((SELECT U ORDER BY U.r LIMIT 250).id),
			movies := array_agg((SELECT M ORDER BY M.r LIMIT 250).id),
			people := array_agg((SELECT P ORDER BY P.r LIMIT 250).id),
		);`,
		&rids,
	)

	return rids, err
}

func TestMain(m *testing.M) {
	code := 1
	defer func() { os.Exit(code) }()

	opts, err := server()
	if err != nil {
		log.Println("error:", err)
		log.Println("skipping benchmarks")
		code = 0
		return
	}

	fmt.Println("these bench marks expect the initialized database")
	fmt.Println("from github.com/edgedb/webapp-bench")
	fmt.Println()

	client, err = edgedb.Connect(ctx, opts)
	if err != nil {
		log.Println(err)
		return
	}

	defer client.Close() // nolint errcheck

	rand.Seed(time.Now().Unix())
	ids, err = getIDs()
	if err != nil {
		log.Println(err)
		return
	}

	code = m.Run()
}

func runFor(d time.Duration, fn func()) {
	timeout := time.After(d)
	for {
		select {
		case <-timeout:
			goto done
		default:
			fn()
		}
	}

done:
	return
}

func getUser(id types.UUID) ([]byte, error) {
	return client.QueryOneJSON(
		ctx,
		`
		SELECT User {
			id,
			name,
			image,
			latest_reviews := (
				WITH UserReviews := User.<author[IS Review]
				SELECT UserReviews {
					id,
					body,
					rating,
					movie: {
						id,
						image,
						title,
						avg_rating
					}
				}
				ORDER BY .creation_time DESC
				LIMIT 10
			)
		}
		FILTER .id = <uuid>$0
		`,
		id,
	)
}

func userID() types.UUID {
	return ids.Users[rand.Intn(len(ids.Users))]
}

func BenchmarkRandIDFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		userID()
	}
}

func BenchmarkUsers(b *testing.B) {
	// warmup
	runFor(5*time.Second, func() {
		_, err := getUser(userID())
		if err != nil {
			log.Println(err)
			b.FailNow()
		}
	})

	b.Run("run", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := userID()
			getUser(id) // nolint errcheck
		}
	})
}

func personID() types.UUID {
	return ids.People[rand.Intn(len(ids.People))]
}

func getPerson(id types.UUID) ([]byte, error) {
	return client.QueryOneJSON(
		ctx,
		`
        SELECT Person {
            id,
            full_name,
            image,
            bio,

            acted_in := (
                WITH M := Person.<cast[IS Movie]
                SELECT M {
                    id,
                    image,
                    title,
                    year,
                    avg_rating
                }
                ORDER BY .year ASC THEN .title ASC
            ),

            directed := (
                WITH M := Person.<directors[IS Movie]
                SELECT M {
                    id,
                    image,
                    title,
                    year,
                    avg_rating
                }
                ORDER BY .year ASC THEN .title ASC
            ),
        }
        FILTER .id = <uuid>$0
		`,
		id,
	)
}

func BenchmarkPerson(b *testing.B) {
	// warmup
	runFor(5*time.Second, func() {
		_, err := getPerson(personID())
		if err != nil {
			log.Println(err)
			b.FailNow()
		}
	})

	b.Run("run", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getPerson(personID()) // nolint errcheck
		}
	})
}

func movieID() types.UUID {
	return ids.Movies[rand.Intn(len(ids.Movies))]
}

func getMovie(id types.UUID) ([]byte, error) {
	return client.QueryOneJSON(
		ctx,
		`
        SELECT Movie {
            id,
            image,
            title,
            year,
            description,
            avg_rating,

            directors: {
                id,
                full_name,
                image,
            }
            ORDER BY Movie.directors@list_order EMPTY LAST
                THEN Movie.directors.last_name,

            cast: {
                id,
                full_name,
                image,
            }
            ORDER BY Movie.cast@list_order EMPTY LAST
                THEN Movie.cast.last_name,

            reviews := (
                SELECT Movie.<movie[IS Review] {
                    id,
                    body,
                    rating,
                    author: {
                        id,
                        name,
                        image,
                    }
                }
                ORDER BY .creation_time DESC
            ),
        }
        FILTER .id = <uuid>$0
		`,
		id,
	)
}

func BenchmarkMovie(b *testing.B) {
	// warmup
	runFor(5*time.Second, func() {
		_, err := getMovie(movieID())
		if err != nil {
			log.Println(err)
			b.FailNow()
		}
	})

	b.Run("run", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getMovie(movieID()) // nolint errcheck
		}
	})
}
