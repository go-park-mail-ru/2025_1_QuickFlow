package user_service

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"quickflow/post_service/internal/delivery/grpc/dto"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

type UserClient struct {
	client pb.UserServiceClient
}

// NewUserClient создает новый gRPC клиент
func NewUserClient(conn *grpc.ClientConn) *UserClient {
	return &UserClient{
		client: pb.NewUserServiceClient(conn),
	}
}

// GetUserByID получает пользователя по ID
func (u *UserClient) GetUserById(ctx context.Context, userId uuid.UUID) (*shared_models.User, error) {
	req := &pb.GetUserByIdRequest{
		Id: userId.String(),
	}

	resp, err := u.client.GetUserById(ctx, req)
	if err != nil {
		return nil, err
	}

	user, err := dto.MapUserDTOToUser(resp.User)
	if err != nil {
		return nil, err
	}

	return user, nil
}
