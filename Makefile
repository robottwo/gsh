.PHONY: build
build:
	@go build -o ./bin/gsh ./cmd/gsh/main.go

.PHONY: test
test:
	@go test ./...
