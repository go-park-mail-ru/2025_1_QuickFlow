coverage.out: $(GO_FILES)
	go test -coverprofile=coverage.out ./config/... ./internal/... ./pkg/... ./utils/...

.PHONY: summarize-coverage
summarize-coverage: coverage.out
	go tool cover -func=coverage.out