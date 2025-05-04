package forms

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"quickflow/config/time"
	"quickflow/shared/models"
)

type CreateCommunityForm struct {
	Name        string       `json:"name" validate:"required"`
	Description string       `json:"description" validate:"required"`
	Avatar      *models.File `json:"avatar"`
}

type CommunityForm struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AvatarUrl   string `json:"avatar_url,omitempty"`
	OwnerId     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	Role        string `json:"role,omitempty"`
}

type PaginationForm struct {
	Count int       `json:"count"`
	Ts    time.Time `json:"ts"`
}

type CommunityMemberOut struct {
	UserId      string `json:"user_id"`
	CommunityId string `json:"community_id"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
	PublicUserInfoOut
}

func (f *CreateCommunityForm) CreateFormToModel() models.Community {
	community := models.Community{
		Name:        f.Name,
		Description: f.Description,
		Avatar:      f.Avatar,
	}

	return community
}

func ToCommunityForm(community models.Community) CommunityForm {
	return CommunityForm{
		Id:          community.ID.String(),
		Name:        community.Name,
		Description: community.Description,
		AvatarUrl:   community.AvatarUrl,
		OwnerId:     community.OwnerID.String(),
		CreatedAt:   community.CreatedAt.Format(time_config.TimeStampLayout),
	}
}

// GetParams gets parameters from the map
func (f *PaginationForm) GetParams(values url.Values) error {
	var (
		err   error
		count int64
	)

	if !values.Has("count") {
		return errors.New("count parameter missing")
	}

	count, err = strconv.ParseInt(values.Get("count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse count")
	}

	f.Count = int(count)
	stringTime := values.Get("ts")
	if len(stringTime) != 0 {
		f.Ts, err = time.Parse(time_config.TimeStampLayout, stringTime)
		if err != nil {
			return errors.New("failed to parse ts")
		}
	} else {
		f.Ts = time.Now()
	}

	return nil
}

func ToCommunityMemberOut(member models.CommunityMember, info models.PublicUserInfo) CommunityMemberOut {
	return CommunityMemberOut{
		UserId:      member.UserID.String(),
		CommunityId: member.CommunityID.String(),
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt.Format(time_config.TimeStampLayout),
		PublicUserInfoOut: PublicUserInfoOut{
			ID:        info.Id.String(),
			Username:  info.Username,
			FirstName: info.Firstname,
			LastName:  info.Lastname,
			AvatarURL: info.AvatarURL,
		},
	}
}
