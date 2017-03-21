install:
	@go install github.com/fbsamples/fbrell

test:
	@go test $$ARGS $(shell go list github.com/fbsamples/fbrell/... | grep -v /vendor/)
