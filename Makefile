quality: lint test bench

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.1 run --sort-results

test:
	go test -v -count=1 -race -bench=$$^ -timeout=20m ./...

bench:
	go test -run=^$$ -bench=. -benchmem -timeout=10m ./...

format:
	gofmt -s -w .

errors:
	type edb || (\
		echo "the edb command must be in your path " && \
		echo "see https://www.edgedb.com/docs/guides/contributing/code#building-locally" && \
		exit 1 \
		)
	edb gen-errors-json --client | \
		go run internal/cmd/generrdefinition/main.go > internal/client/errors_gen.go
	edb gen-errors-json --client | \
		go run internal/cmd/generrexport/main.go > errors_gen.go
	make format

gendocs:
	go run internal/cmd/gendocs/*.go

gendocs-lint:
	go run internal/cmd/gendocs/*.go --lint
