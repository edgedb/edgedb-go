quality: lint test bench

lint:
	golangci-lint run

test:
	go test -race -timeout=5m ./...

bench:
	go test -run=^$$ -bench=. -benchmem -timeout=5m ./...

format:
	gofmt -s -w .
