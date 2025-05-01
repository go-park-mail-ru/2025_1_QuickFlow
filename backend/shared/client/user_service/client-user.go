package userclient

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserServiceClient(conn *grpc.ClientConn) *Client {
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

	resp, err := c.client.SignUp(ctx, req)
	if err != nil {
		return uuid.Nil, shared_models.Session{}, unwrapGRPCError(err)
	}

	sessionID, err := uuid.Parse(resp.Session.Id)
	if err != nil {
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
			Username: authData.Login,
			Password: authData.Password,
		},
	}

	resp, err := c.client.SignIn(ctx, req)
	if err != nil {
		return shared_models.Session{}, unwrapGRPCError(err)
	}

	sessionID, err := uuid.Parse(resp.Session.Id)
	if err != nil {
		return shared_models.Session{}, err
	}

	return shared_models.Session{
		SessionId:  sessionID,
		ExpireDate: resp.Session.Expiry.AsTime(),
	}, nil
}

func (c *Client) DeleteUserSession(ctx context.Context, session string) error {
	req := &pb.SignOutRequest{SessionId: session}
	_, err := c.client.SignOut(ctx, req)
	return unwrapGRPCError(err)
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (shared_models.User, error) {
	resp, err := c.client.GetUserByUsername(ctx, &pb.GetUserByUsernameRequest{Username: username})
	if err != nil {
		return shared_models.User{}, unwrapGRPCError(err)
	}

	user, err := MapUserDTOToUser(resp.User)
	if err != nil {
		return shared_models.User{}, err
	}
	return *user, nil
}

func (c *Client) GetUserById(ctx context.Context, userId uuid.UUID) (shared_models.User, error) {
	resp, err := c.client.GetUserById(ctx, &pb.GetUserByIdRequest{Id: userId.String()})
	if err != nil {
		return shared_models.User{}, unwrapGRPCError(err)
	}

	user, err := MapUserDTOToUser(resp.User)
	if err != nil {
		return shared_models.User{}, err
	}
	return *user, nil
}

func (c *Client) LookupUserSession(ctx context.Context, session shared_models.Session) (shared_models.User, error) {
	resp, err := c.client.LookupUserSession(ctx, &pb.LookupUserSessionRequest{SessionId: session.SessionId.String()})
	if err != nil {
		return shared_models.User{}, unwrapGRPCError(err)
	}

	userId, err := uuid.Parse(resp.UserId)
	if err != nil {
		return shared_models.User{}, err
	}

	return shared_models.User{Id: userId, Username: resp.Username}, nil
}

func unwrapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if ok {
		return fmt.Errorf(s.Message())
	}
	return err
}
