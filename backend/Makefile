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
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/feed-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/feed-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/chat-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/chat-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/friends-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/friends-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/message-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/message-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/message-handlerWS.go -destination=$(DELIEVERY_PATH)/http/mocks/messageWS-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/profile-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/profile-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/http/search-handler.go -destination=$(DELIEVERY_PATH)/http/mocks/search-mock.go -package=mocks
	${MOCKGEN} -source=$(DELIEVERY_PATH)/ws/manager.go -destination=$(DELIEVERY_PATH)/ws/mocks/manager-mock.go -package=mocks

	${MOCKGEN} -source=$(USECASE_PATH)/auth-usecase.go -destination=$(USECASE_PATH)/mocks/auth-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/post-usecase.go -destination=$(USECASE_PATH)/mocks/post-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/chat-usecase.go -destination=$(USECASE_PATH)/mocks/chat-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/friends-usecase.go -destination=$(USECASE_PATH)/mocks/friends-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/message-usecase.go -destination=$(USECASE_PATH)/mocks/message-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/profile-usecase.go -destination=$(USECASE_PATH)/mocks/profile-mock.go -package=mocks
	${MOCKGEN} -source=$(USECASE_PATH)/search-usecase.go -destination=$(USECASE_PATH)/mocks/search-mock.go -package=mocks

	${MOCKGEN} -source=$(REPOSITORY_PATH)/postgres/user.go -destination=$(REPOSITORY_PATH)/postgres/mocks/user-mock.go -package=mocks
	${MOCKGEN} -source=$(REPOSITORY_PATH)/postgres/post.go -destination=$(REPOSITORY_PATH)/postgres/mocks/post-mock.go -package=mocks
	${MOCKGEN} -source=$(REPOSITORY_PATH)/postgres/friends.go -destination=$(REPOSITORY_PATH)/postgres/mocks/friends-mock.go -package=mocks
	${MOCKGEN} -source=$(REPOSITORY_PATH)/postgres/message.go -destination=$(REPOSITORY_PATH)/postgres/mocks/message-mock.go -package=mocks
	${MOCKGEN} -source=$(REPOSITORY_PATH)/postgres/profile.go -destination=$(REPOSITORY_PATH)/postgres/mocks/profile-mock.go -package=mocks



gendock:
	swag init --parseDependency --parseInternal --parseDepth=2 --output ../docs/docs.go --dir .