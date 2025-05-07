GO_FILES=$(shell go list ./... | grep -v 'mocks')
MOCKGEN=go run github.com/golang/mock/mockgen
DELIEVERY_PATH=internal/delivery
USECASE_PATH=internal/usecase
REPOSITORY_PATH=internal/repository


coverage.out:
	go test $(GO_FILES) -coverprofile=coverage.out

.PHONY: summarize-coverage
summarize-coverage: coverage.out
	go tool cover -func=coverage.out

.PHONY: mockgen
mockgen:
	sh ./generate_mocks.sh


gendock:
	swag init --parseDependency --parseInternal --parseDepth=2 --output ../docs/docs.go --dir .