package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/internal/models"
)

type PostPostgres struct {
	Id           pgtype.UUID
	CreatorId    pgtype.UUID
	Desc         pgtype.Text
	ImagesURLs   []pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	LikeCount    pgtype.Int8
	RepostCount  pgtype.Int8
	CommentCount pgtype.Int8
	IsRepost     pgtype.Bool
}

// ConvertPostToPostgres converts models.Post to PostPostgres.
func ConvertPostToPostgres(post models.Post) PostPostgres {
	var pics []pgtype.Text
	for _, pic := range post.ImagesURL {
		if pic != "" {
			pics = append(pics, pgtype.Text{String: pic, Valid: true})
		}
	}

	return PostPostgres{
		Id:           pgtype.UUID{Bytes: post.Id, Valid: true},
		CreatorId:    pgtype.UUID{Bytes: post.CreatorId, Valid: true},
		Desc:         convertStringToPostgresText(post.Desc),
		ImagesURLs:   pics,
		CreatedAt:    pgtype.Timestamptz{Time: post.CreatedAt, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: post.UpdatedAt, Valid: true},
		LikeCount:    pgtype.Int8{Int64: int64(post.LikeCount), Valid: true},
		RepostCount:  pgtype.Int8{Int64: int64(post.RepostCount), Valid: true},
		CommentCount: pgtype.Int8{Int64: int64(post.CommentCount), Valid: true},
		IsRepost:     pgtype.Bool{Bool: post.IsRepost, Valid: true},
	}
}

// ToPost converts PostPostgres to models.Post.
func (p *PostPostgres) ToPost() models.Post {
	var picsSlice []string

	for _, pics := range p.ImagesURLs {
		picsSlice = append(picsSlice, pics.String)
	}

	return models.Post{
		Id:           p.Id.Bytes,
		CreatorId:    p.CreatorId.Bytes,
		Desc:         p.Desc.String,
		ImagesURL:    picsSlice,
		CreatedAt:    p.CreatedAt.Time,
		UpdatedAt:    p.UpdatedAt.Time,
		LikeCount:    int(p.LikeCount.Int64),
		RepostCount:  int(p.RepostCount.Int64),
		CommentCount: int(p.CommentCount.Int64),
		IsRepost:     p.IsRepost.Bool,
	}
}

func convertStringToPostgresText(s string) pgtype.Text {
	if len(s) == 0 {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
