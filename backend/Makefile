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
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/auth-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/auth-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/feed-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/feed-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/auth-usecase.go -destination=$(USECASE_PATH)/mocks/auth-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/post-usecase.go -destination=$(USECASE_PATH)/mocks/post-mock.go -package=mocks
