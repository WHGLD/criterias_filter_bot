vendor:
	go mod tidy
	go mod vendor

lint:
	@echo "Linting..."
	go vet ./...

lintFull:
	go vet ./... && golint ./internal/...

unit:
	@echo "Unit testing..."
	go test -race ./...

check: lintFull unit

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -ldflags="-w -s" -o cmd/bot/criterias_filter_bot ./cmd/bot/main.go

run:
	go run ./cmd/bot/main.go