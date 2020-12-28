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
	edb gen-errors-json --client | \
		go run internal/cmd/generr/main.go -file=codes > \
		generatederrors.go
