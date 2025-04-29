package dto

import (
	"github.com/google/uuid"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

func MapUserDTOToUser(userDTO *pb.User) (*shared_models.User, error) {
	if userDTO == nil {
		return nil, nil
	}

	id, err := uuid.Parse(userDTO.Id)
	if err != nil {
		return nil, err
	}

	return &shared_models.User{
		Id:       id,
		Username: userDTO.Username,
		Password: userDTO.Password,
		Salt:     userDTO.Salt,
		LastSeen: userDTO.LastSeen.AsTime(),
	}, nil
}
