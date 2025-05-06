package userclient

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	shared_models "quickflow/shared/models"
	db "quickflow/shared/proto/user_service"
)

func MapSchoolEducationToDTO(schoolEducation *shared_models.SchoolEducation) *db.SchoolEducation {
	if schoolEducation == nil {
		return nil
	}

	return &db.SchoolEducation{
		City: schoolEducation.City,
		Name: schoolEducation.School,
	}
}

func MapSchoolEducationDTOToModel(schoolEducationDTO *db.SchoolEducation) *shared_models.SchoolEducation {
	if schoolEducationDTO == nil {
		return nil
	}

	return &shared_models.SchoolEducation{
		City:   schoolEducationDTO.City,
		School: schoolEducationDTO.Name,
	}
}

func MapUniversityEducationToDTO(universityEducation *shared_models.UniversityEducation) *db.UniversityEducation {
	if universityEducation == nil {
		return nil
	}

	return &db.UniversityEducation{
		City:           universityEducation.City,
		University:     universityEducation.University,
		Faculty:        universityEducation.Faculty,
		GraduationYear: int32(universityEducation.GraduationYear),
	}
}

func MapUniversityEducationDTOToModel(universityEducationDTO *db.UniversityEducation) *shared_models.UniversityEducation {
	if universityEducationDTO == nil {
		return nil
	}

	return &shared_models.UniversityEducation{
		City:           universityEducationDTO.City,
		University:     universityEducationDTO.University,
		Faculty:        universityEducationDTO.Faculty,
		GraduationYear: int(universityEducationDTO.GraduationYear),
	}
}

func MapContactInfoToDTO(contactInfo *shared_models.ContactInfo) *db.ContactInfo {
	if contactInfo == nil {
		return nil
	}

	return &db.ContactInfo{
		Email:       contactInfo.Email,
		PhoneNumber: contactInfo.Phone,
		City:        contactInfo.City,
	}
}

func MapContactInfoDTOToModel(contactInfoDTO *db.ContactInfo) *shared_models.ContactInfo {
	if contactInfoDTO == nil {
		return nil
	}

	return &shared_models.ContactInfo{
		Email: contactInfoDTO.Email,
		Phone: contactInfoDTO.PhoneNumber,
		City:  contactInfoDTO.City,
	}
}

func MapBasicInfoToDTO(basicInfo *shared_models.BasicInfo) *db.BasicInfo {
	if basicInfo == nil {
		return nil
	}

	return &db.BasicInfo{
		Firstname: basicInfo.Name,
		Lastname:  basicInfo.Surname,
		Sex:       int32(basicInfo.Sex),
		BirthDate: timestamppb.New(basicInfo.DateOfBirth),
		Bio:       basicInfo.Bio,
		AvatarUrl: basicInfo.AvatarUrl,
		CoverUrl:  basicInfo.BackgroundUrl,
	}
}

func MapBasicInfoDTOToModel(basicInfoDTO *db.BasicInfo) *shared_models.BasicInfo {
	if basicInfoDTO == nil {
		return nil
	}

	return &shared_models.BasicInfo{
		Name:          basicInfoDTO.Firstname,
		Surname:       basicInfoDTO.Lastname,
		Sex:           shared_models.Sex(basicInfoDTO.Sex),
		DateOfBirth:   basicInfoDTO.BirthDate.AsTime(),
		Bio:           basicInfoDTO.Bio,
		AvatarUrl:     basicInfoDTO.AvatarUrl,
		BackgroundUrl: basicInfoDTO.CoverUrl,
	}
}

func MapProfileToProfileDTO(profile *shared_models.Profile) *db.Profile {
	if profile == nil {
		return nil
	}

	return &db.Profile{
		Id:                  profile.UserId.String(),
		Username:            profile.Username,
		BasicInfo:           MapBasicInfoToDTO(profile.BasicInfo),
		Avatar:              file_service.ModelFileToProto(profile.Avatar),
		Cover:               file_service.ModelFileToProto(profile.Background),
		SchoolEducation:     MapSchoolEducationToDTO(profile.SchoolEducation),
		UniversityEducation: MapUniversityEducationToDTO(profile.UniversityEducation),
		ContactInfo:         MapContactInfoToDTO(profile.ContactInfo),
		LastSeen:            timestamppb.New(profile.LastSeen),
	}
}

func MapProfileDTOToProfile(profileDTO *db.Profile) (*shared_models.Profile, error) {
	if profileDTO == nil {
		return nil, nil
	}

	id, err := uuid.Parse(profileDTO.Id)
	if err != nil {
		return nil, err
	}
	return &shared_models.Profile{
		UserId:              id,
		Username:            profileDTO.Username,
		BasicInfo:           MapBasicInfoDTOToModel(profileDTO.BasicInfo),
		SchoolEducation:     MapSchoolEducationDTOToModel(profileDTO.SchoolEducation),
		UniversityEducation: MapUniversityEducationDTOToModel(profileDTO.UniversityEducation),
		ContactInfo:         MapContactInfoDTOToModel(profileDTO.ContactInfo),
		LastSeen:            profileDTO.LastSeen.AsTime(),
		Avatar:              file_service.ProtoFileToModel(profileDTO.Avatar),
		Background:          file_service.ProtoFileToModel(profileDTO.Cover),
	}, nil
}

func MapPublicUserInfoToDTO(publicUserInfo *shared_models.PublicUserInfo) *db.PublicUserInfo {
	if publicUserInfo == nil {
		return nil
	}

	return &db.PublicUserInfo{
		Id:        publicUserInfo.Id.String(),
		Username:  publicUserInfo.Username,
		Firstname: publicUserInfo.Firstname,
		Lastname:  publicUserInfo.Lastname,
		AvatarUrl: publicUserInfo.AvatarURL,
		LastSeen:  timestamppb.New(publicUserInfo.LastSeen),
	}
}

func MapPublicUserInfoDTOToModel(publicUserInfoDTO *db.PublicUserInfo) (*shared_models.PublicUserInfo, error) {
	if publicUserInfoDTO == nil {
		return nil, nil
	}

	id, err := uuid.Parse(publicUserInfoDTO.Id)
	if err != nil {
		return nil, err
	}

	return &shared_models.PublicUserInfo{
		Id:        id,
		Username:  publicUserInfoDTO.Username,
		Firstname: publicUserInfoDTO.Firstname,
		Lastname:  publicUserInfoDTO.Lastname,
		AvatarURL: publicUserInfoDTO.AvatarUrl,
		LastSeen:  publicUserInfoDTO.LastSeen.AsTime(),
	}, nil
}
