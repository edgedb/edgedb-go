quality: lint test bench

lint:
	golangci-lint run

test:
	go test -race -timeout=5m ./...

bench:
	go test -run=^$$ -bench=. -benchmem -timeout=5m ./...

format:
	gofmt -s -w .

errors:
	type edb || (\
		echo "the edb command must be in your path " && \
		echo "see https://www.edgedb.com/docs/internals/dev/#building-locally" && \
		exit 1 \
		)
	edb gen-errors-json --client | \
		go run internal/cmd/generr/*.go > \
		generatederrors.go
