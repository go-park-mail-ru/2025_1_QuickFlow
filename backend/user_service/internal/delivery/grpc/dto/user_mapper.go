package dto

import (
	"bytes"
	"io"
	"path"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	shared_models "quickflow/shared/models"
	proto "quickflow/shared/proto/file_service"
	pb "quickflow/user_service/internal/delivery/grpc/proto"
)

func MapFileToFileDTO(file *proto.File) *shared_models.File {
	if file == nil {
		return nil
	}

	return &shared_models.File{
		Reader:     bytes.NewReader(file.File),
		Size:       file.FileSize,
		Name:       file.FileName,
		Ext:        path.Ext(file.FileName),
		MimeType:   file.FileType,
		AccessMode: shared_models.AccessMode(file.AccessMode),
	}
}

func MapFileDTOToFile(file *shared_models.File) *proto.File {
	if file == nil {
		return nil
	}

	content, err := io.ReadAll(file.Reader)
	if err != nil {
		return nil
	}

	return &proto.File{
		File:       content,
		FileSize:   file.Size,
		FileName:   file.Name,
		FileType:   file.MimeType,
		AccessMode: proto.AccessMode(file.AccessMode),
	}
}

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
		Login:    signIn.Username,
		Password: signIn.Password,
	}
}

func MapSessionToDTO(session shared_models.Session) *pb.Session {
	return &pb.Session{
		Id:     session.SessionId.String(),
		Expiry: timestamppb.New(session.ExpireDate),
	}
}
