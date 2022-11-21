.PHONY: release

snapshot:
	goreleaser release --snapshot --rm-dist

build:
	goreleaser build --rm-dist

test:
	golangci-lint run
