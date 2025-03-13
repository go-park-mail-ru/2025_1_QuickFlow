coverage.out: $(GO_FILES)
	go test -coverprofile=coverage.out $(go test ./... | grep -v "/test/")

.PHONY: summarize-coverage
summarize-coverage: coverage.out
	go tool cover -func=coverage.out