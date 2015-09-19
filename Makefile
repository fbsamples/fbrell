export GOVENDOREXPERIMENT=1

docker:
	@docker build -t rell-builder .
	@docker run rell-builder | docker build -t daaku/rell -

install:
	@go install github.com/daaku/rell

test:
	@go test $(filter-out $@,$(MAKECMDGOALS)) $(go list github.com/daaku/rell/... | grep -v /vendor/)
