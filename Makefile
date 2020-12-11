quality: lint test bench

lint:
	golangci-lint run

test:
	go test -race -timeout=2m ./...

bench:
	go test -run=^$$ -bench=. -benchmem -timeout=2m ./...

format:
	gofmt -s -w .
