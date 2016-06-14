install:
	@go install github.com/daaku/rell

test:
	@go test $$ARGS $(shell go list github.com/daaku/rell/... | grep -v /vendor/)
