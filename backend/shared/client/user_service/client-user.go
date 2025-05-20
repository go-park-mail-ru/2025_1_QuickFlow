package userclient

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"quickflow/shared/logger"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:   conn,
		client: pb.NewUserServiceClient(conn),
	}
}

// Реализация UserUseCase
func (c *Client) CreateUser(ctx context.Context, user shared_models.User, profile shared_models.Profile) (uuid.UUID, shared_models.Session, error) {
	req := &pb.SignUpRequest{
		User:    MapUserToUserDTO(&user),
		Profile: MapProfileToProfileDTO(&profile),
	}

	logger.Info(ctx, "Sending request to create user: %v", req)
	resp, err := c.client.SignUp(ctx, req)
	if err != nil {
		logger.Error(ctx, "Failed to create user: %v", err)
		return uuid.Nil, shared_models.Session{}, err
	}

	sessionID, err := uuid.Parse(resp.Session.Id)
	if err != nil {
		logger.Error(ctx, "Failed to parse session ID: %v", err)
		return uuid.Nil, shared_models.Session{}, err
	}

	return user.Id, shared_models.Session{
		SessionId:  sessionID,
		ExpireDate: resp.Session.Expiry.AsTime(),
	}, nil
}

func (c *Client) AuthUser(ctx context.Context, authData shared_models.LoginData) (shared_models.Session, error) {
	req := &pb.SignInRequest{
		SignIn: &pb.SignIn{
			Username: authData.Username,
			Password: authData.Password,
		},
	}

	logger.Info(ctx, "Sending request to authenticate user: %v", req)
	resp, err := c.client.SignIn(ctx, req)
	if err != nil {
		logger.Error(ctx, "Failed to authenticate user: %v", err)
		return shared_models.Session{}, err
	}

	sessionID, err := uuid.Parse(resp.Session.Id)
	if err != nil {
		logger.Error(ctx, "Failed to parse session ID: %v", err)
		return shared_models.Session{}, err
	}

	return shared_models.Session{
		SessionId:  sessionID,
		ExpireDate: resp.Session.Expiry.AsTime(),
	}, nil
}

func (c *Client) DeleteUserSession(ctx context.Context, session string) error {
	req := &pb.SignOutRequest{SessionId: session}
	logger.Info(ctx, "Sending request to delete session: %v", req)
	_, err := c.client.SignOut(ctx, req)
	return err
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (shared_models.User, error) {
	logger.Info(ctx, "Sending request to get user by username: %v", username)
	resp, err := c.client.GetUserByUsername(ctx, &pb.GetUserByUsernameRequest{Username: username})
	if err != nil {
		logger.Error(ctx, "Failed to get user by username: %v", err)
		return shared_models.User{}, err
	}

	user, err := MapUserDTOToUser(resp.User)
	if err != nil {
		logger.Error(ctx, "Failed to convert to User: %v", err)
		return shared_models.User{}, err
	}
	return *user, nil
}

func (c *Client) GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error) {
	logger.Info(ctx, "Sending request to get user by id: %v", userId)
	resp, err := c.client.GetUserById(ctx, &pb.GetUserByIdRequest{Id: userId.String()})
	if err != nil {
		logger.Error(ctx, "Failed to get user by id: %v", err)
		return shared_models.User{}, err
	}

	user, err := MapUserDTOToUser(resp.User)
	if err != nil {
		logger.Error(ctx, "Failed to convert to User: %v", err)
		return shared_models.User{}, err
	}
	return *user, nil
}

func (c *Client) LookupUserSession(ctx context.Context, session shared_models.Session) (shared_models.User, error) {
	logger.Info(ctx, "Sending request to lookup user session: %v", session)
	resp, err := c.client.LookupUserSession(ctx, &pb.LookupUserSessionRequest{SessionId: session.SessionId.String()})
	if err != nil {
		logger.Error(ctx, "Failed to lookup user session: %v", err)
		return shared_models.User{}, err
	}

	userId, err := uuid.Parse(resp.UserId)
	if err != nil {
		logger.Error(ctx, "Failed to parse user ID: %v", err)
		return shared_models.User{}, err
	}

	return shared_models.User{Id: userId, Username: resp.Username}, nil
}

func (c *Client) SearchSimilarUser(ctx context.Context, toSearch string, usersCount uint) ([]shared_models.PublicUserInfo, error) {
	logger.Info(ctx, "Sending request to search similar users: %v", toSearch)
	resp, err := c.client.SearchSimilarUser(ctx, &pb.SearchSimilarUserRequest{
		ToSearch: toSearch,
		NumUsers: int32(usersCount),
	})
	if err != nil {
		logger.Error(ctx, "Failed to search similar users: %v", err)
		return nil, err
	}

	users := make([]shared_models.PublicUserInfo, len(resp.UsersInfo))
	for i, userDTO := range resp.UsersInfo {
		user, err := MapPublicUserInfoDTOToModel(userDTO)
		if err != nil {
			logger.Error(ctx, "Failed to convert to PublicUserInfo: %v", err)
			return nil, err
		}
		users[i] = *user
	}

	return users, nil
}
