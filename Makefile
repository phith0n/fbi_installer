.PHONY: release

release:
	goreleaser release --snapshot --rm-dist
