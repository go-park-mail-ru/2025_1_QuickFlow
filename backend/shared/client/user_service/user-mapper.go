package userclient

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

func MapUserToUserDTO(user *shared_models.User) *pb.User {
	if user == nil {
		return nil
	}

	return &pb.User{
		Id:       user.Id.String(),
		Username: user.Username,
		Password: user.Password,
		Salt:     user.Salt,
		LastSeen: timestamppb.New(user.LastSeen),
	}
}

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

func MapSignInToSignInDTO(signIn *pb.SignIn) *shared_models.LoginData {
	if signIn == nil {
		return nil
	}

	return &shared_models.LoginData{
		Username: signIn.Username,
		Password: signIn.Password,
	}
}

func MapSessionToDTO(session shared_models.Session) *pb.Session {
	return &pb.Session{
		Id:     session.SessionId.String(),
		Expiry: timestamppb.New(session.ExpireDate),
	}
}
