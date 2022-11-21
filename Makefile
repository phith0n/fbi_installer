.PHONY: release

release:
	goreleaser release --snapshot --rm-dist

test:
	golangci-lint run
