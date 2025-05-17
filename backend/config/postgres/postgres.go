package postgres_config

import (
	getenv "quickflow/utils/get-env"
)

const (
	defaultDataBaseCommunityURL string = "postgresql://app_community:community_password@postgres:5432/quickflow_db"
	defaultDataBaseUserURL      string = "postgresql://app_user:user_password@postgres:5432/quickflow_db"
	defaultDataBaseFeedbackURL  string = "postgresql://app_feedback:feedback_password@postgres:5432/quickflow_db"
	defaultDataBaseMessengerURL string = "postgresql://app_messenger:messenger_password@postgres:5432/quickflow_db"
	defaultDataBasePostURL      string = "postgresql://app_post:post_password@postgres:5432/quickflow_db"
	defaultDataBaseFriendsUrl   string = "postgresql://app_friends:friends_password@postgres:5432/quickflow_db"
)

type PostgresConfig struct {
	DatabaseCommunityUrl string
	DatabaseUserUrl      string
	DatabaseFeedbackUrl  string
	DatabaseMessengerUrl string
	DatabasePostUrl      string
	DatabaseFriendsUrl   string
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		DatabaseCommunityUrl: getenv.GetEnv("DATABASE_COMMUNITY_URL", defaultDataBaseCommunityURL),
		DatabaseUserUrl:      getenv.GetEnv("DATABASE_USER_URL", defaultDataBaseUserURL),
		DatabaseFeedbackUrl:  getenv.GetEnv("DATABASE_FEEDBACK_URL", defaultDataBaseFeedbackURL),
		DatabaseMessengerUrl: getenv.GetEnv("DATABASE_MESSENGER_URL", defaultDataBaseMessengerURL),
		DatabasePostUrl:      getenv.GetEnv("DATABASE_POST_URL", defaultDataBasePostURL),
		DatabaseFriendsUrl:   getenv.GetEnv("DATABASE_FRIENDS_URL", defaultDataBaseFriendsUrl),
	}
}
