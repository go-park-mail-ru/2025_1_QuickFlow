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
	Nickname    string       `json:"nickname"`
	Name        string       `json:"name" validate:"required"`
	Description string       `json:"description,omitempty"`
	Avatar      *models.File `json:"avatar"`
	Cover       *models.File `json:"cover"`
}

type CommunityInfo struct {
	NickName    string `json:"nickname"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarUrl   string `json:"avatar_url"`
	CoverUrl    string `json:"cover_url"`
}

func CommunityInfoFromModel(communityInfo models.BasicCommunityInfo, nickName string) *CommunityInfo {
	return &CommunityInfo{
		NickName:    nickName,
		Name:        communityInfo.Name,
		Description: communityInfo.Description,
		AvatarUrl:   communityInfo.AvatarUrl,
		CoverUrl:    communityInfo.CoverUrl,
	}
}

type CommunityForm struct {
	Id        string             `json:"id"`
	Owner     *PublicUserInfoOut `json:"owner"`
	OwnerId   string             `json:"owner_id"`
	CreatedAt string             `json:"created_at"`
	Avatar    *models.File       `json:"-"`
	Cover     *models.File       `json:"-"`
	Role      string             `json:"role,omitempty"`

	CommunityInfo *CommunityInfo `json:"community"`
	ContactInfo   *ContactInfo   `json:"contact_info,omitempty"`
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
	return models.Community{
		NickName: f.Nickname,
		BasicInfo: &models.BasicCommunityInfo{
			Name:        f.Name,
			Description: f.Description,
		},
		Avatar: f.Avatar,
		Cover:  f.Cover,
	}
}

func ToCommunityForm(community models.Community, ownerInfo models.PublicUserInfo) CommunityForm {
	return CommunityForm{
		Id:      community.ID.String(),
		OwnerId: community.OwnerID.String(),
		Owner: &PublicUserInfoOut{
			ID:        ownerInfo.Id.String(),
			Username:  ownerInfo.Username,
			FirstName: ownerInfo.Firstname,
			LastName:  ownerInfo.Lastname,
			AvatarURL: ownerInfo.AvatarURL,
		},
		CreatedAt:     community.CreatedAt.Format(time_config.TimeStampLayout),
		CommunityInfo: CommunityInfoFromModel(*community.BasicInfo, community.NickName),
		ContactInfo:   ContactInfoToForm(community.ContactInfo),
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
