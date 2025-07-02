.PHONY: default

default: init

init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@v0.15.0

clean:
	rm -rf ./build

linter:
	fieldalignment -fix ./...
	golangci-lint run -c .golangci.yml --timeout=5m -v --fix

lint:
	golangci-lint run -c .golangci.yml --timeout=5m -v

test:
	go test ./... -bench . -benchmem

integration-test:
	docker-compose up -d
	sleep 20
	cd test/integration && go test -v
	docker-compose down

tidy:
	go mod tidy
	cd example/simple && go mod tidy && cd ../..
	cd example/default-mapper && go mod tidy && cd ../..
	cd example/simple-logger && go mod tidy && cd ../..
	cd example/struct-config && go mod tidy && cd ../..
	cd example/grafana && go mod tidy && cd ../..
	cd test/integration && go mod tidy && cd ../..