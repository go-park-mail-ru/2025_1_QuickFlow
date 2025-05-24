package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type PostPostgres struct {
	Id           pgtype.UUID
	CreatorId    pgtype.UUID
	CreatorType  pgtype.Text
	Desc         pgtype.Text
	Files        []PostgresFile
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	LikeCount    pgtype.Int8
	RepostCount  pgtype.Int8
	CommentCount pgtype.Int8
	IsRepost     pgtype.Bool
	IsLiked      pgtype.Bool
}

// ConvertPostToPostgres converts models.Post to PostPostgres.
func ConvertPostToPostgres(post models.Post) PostPostgres {
	return PostPostgres{
		Id:           pgtype.UUID{Bytes: post.Id, Valid: true},
		CreatorId:    pgtype.UUID{Bytes: post.CreatorId, Valid: true},
		CreatorType:  convertStringToPostgresText(string(post.CreatorType)),
		Desc:         convertStringToPostgresText(post.Desc),
		Files:        FilesToPostgres(post.Files),
		CreatedAt:    pgtype.Timestamptz{Time: post.CreatedAt, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: post.UpdatedAt, Valid: true},
		LikeCount:    pgtype.Int8{Int64: int64(post.LikeCount), Valid: true},
		RepostCount:  pgtype.Int8{Int64: int64(post.RepostCount), Valid: true},
		CommentCount: pgtype.Int8{Int64: int64(post.CommentCount), Valid: true},
		IsRepost:     pgtype.Bool{Bool: post.IsRepost, Valid: true},
		IsLiked:      pgtype.Bool{Bool: post.IsLiked, Valid: true},
	}
}

// ToPost converts PostPostgres to models.Post.
func (p *PostPostgres) ToPost() models.Post {

	return models.Post{
		Id:           p.Id.Bytes,
		CreatorId:    p.CreatorId.Bytes,
		CreatorType:  models.PostCreatorType(p.CreatorType.String),
		Desc:         p.Desc.String,
		Files:        PostgresFilesToModels(p.Files),
		CreatedAt:    p.CreatedAt.Time,
		UpdatedAt:    p.UpdatedAt.Time,
		LikeCount:    int(p.LikeCount.Int64),
		RepostCount:  int(p.RepostCount.Int64),
		CommentCount: int(p.CommentCount.Int64),
		IsRepost:     p.IsRepost.Bool,
		IsLiked:      p.IsLiked.Bool,
	}
}

func convertStringToPostgresText(s string) pgtype.Text {
	if len(s) == 0 {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
