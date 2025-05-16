package micro_addr

const (
	DefaultGatewayServiceAddrEnv = "GATEWAY_SERVICE_ADDR"
	DefaultGatewayServiceAddr    = 8080
	DefaultGatewayServiceName    = "gateway servive"

	DefaultFileServiceAddrEnv = "FILE_SERVICE_ADDR"
	DefaultFileServicePort    = 8081
	DefaultFileServiceName    = "file service"

	DefaultPostServiceAddrEnv = "POST_SERVICE_ADDR"
	DefaultPostServicePort    = 8082
	DefaultPostServiceName    = "post service"

	DefaultUserServiceAddrEnv = "USER_SERVICE_ADDR"
	DefaultUserServicePort    = 8083
	DefaultUserServiceName    = "user service"

	DefaultMessengerServiceAddrEnv = "MESSENGER_SERVICE_ADDR"
	DefaultMessengerServicePort    = 8084
	DefaultMessengerServiceName    = "messenger service"

	DefaultFeedbackServiceAddrEnv = "FEEDBACK_SERVICE_ADDR"
	DefaultFeedbackServicePort    = 8085
	DefaultFeedbackServiceName    = "feedback service"

	DefaultFriendsServiceAddrEnv = "FRIENDS_SERVICE_ADDR"
	DefaultFriendsServicePort    = 8086
	DefaultFriendsServiceName    = "friends service"

	DefaultCommunityServiceAddrEnv = "COMMUNITY_SERVICE_ADDR"
	DefaultCommunityServicePort    = 8087
	DefaultCommunityServiceName    = "community service"

	MaxMessageSize = 15 * 1024 * 1024
)
