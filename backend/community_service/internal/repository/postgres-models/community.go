package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type ContactInfoPostgres struct {
	City  pgtype.Text
	Email pgtype.Text
	Phone pgtype.Text
}

type CommunityPostgres struct {
	Id          pgtype.UUID
	OwnerId     pgtype.UUID
	Name        pgtype.Text
	NickName    pgtype.Text
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	AvatarUrl   pgtype.Text
	CoverUrl    pgtype.Text

	ContactInfo *ContactInfoPostgres
}

func CommunityModelToPostgres(community *models.Community) CommunityPostgres {
	if community == nil {
		return CommunityPostgres{}
	}

	return CommunityPostgres{
		Id:          pgtype.UUID{Bytes: community.ID, Valid: true},
		OwnerId:     pgtype.UUID{Bytes: community.OwnerID, Valid: true},
		Name:        convertStringToPostgresText(community.BasicInfo.Name),
		Description: convertStringToPostgresText(community.BasicInfo.Description),
		CreatedAt:   pgtype.Timestamptz{Time: community.CreatedAt, Valid: true},
		AvatarUrl:   convertStringToPostgresText(community.BasicInfo.AvatarUrl),
		CoverUrl:    convertStringToPostgresText(community.BasicInfo.CoverUrl),
		NickName:    convertStringToPostgresText(community.NickName),
	}
}

func (pgCommunity *CommunityPostgres) PostgresToCommunityModel() models.Community {
	community := models.Community{
		ID:        pgCommunity.Id.Bytes,
		OwnerID:   pgCommunity.OwnerId.Bytes,
		CreatedAt: pgCommunity.CreatedAt.Time,
		NickName:  pgCommunity.NickName.String,
		BasicInfo: &models.BasicCommunityInfo{
			Name:        pgCommunity.Name.String,
			Description: pgCommunity.Description.String,
			AvatarUrl:   pgCommunity.AvatarUrl.String,
			CoverUrl:    pgCommunity.CoverUrl.String,
		},
	}

	if pgCommunity.ContactInfo != nil {
		community.ContactInfo = &models.ContactInfo{
			City:  pgCommunity.ContactInfo.City.String,
			Email: pgCommunity.ContactInfo.Email.String,
			Phone: pgCommunity.ContactInfo.Phone.String,
		}
	}

	return community
}

func convertStringToPostgresText(s string) pgtype.Text {
	if len(s) == 0 {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

type CommunityMemberPostgres struct {
	UserId      pgtype.UUID
	CommunityId pgtype.UUID
	Role        pgtype.Text
	JoinedAt    pgtype.Timestamptz
}

func CommunityMemberModelToPostgres(member *models.CommunityMember) CommunityMemberPostgres {
	if member == nil {
		return CommunityMemberPostgres{}
	}

	return CommunityMemberPostgres{
		UserId:      pgtype.UUID{Bytes: member.UserID, Valid: true},
		CommunityId: pgtype.UUID{Bytes: member.CommunityID, Valid: true},
		Role:        convertStringToPostgresText(string(member.Role)),
		JoinedAt:    pgtype.Timestamptz{Time: member.JoinedAt, Valid: true},
	}
}

func (pgMember *CommunityMemberPostgres) PostgresToCommunityMemberModel() models.CommunityMember {
	member := models.CommunityMember{
		UserID:      pgMember.UserId.Bytes,
		CommunityID: pgMember.CommunityId.Bytes,
		Role:        models.CommunityRole(pgMember.Role.String),
		JoinedAt:    pgMember.JoinedAt.Time,
	}
	return member
}
