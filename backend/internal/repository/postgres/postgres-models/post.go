package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/internal/models"
)

type PostPostgres struct {
	Id           pgtype.UUID
	CreatorId    pgtype.UUID
	Desc         pgtype.Text
	Pics         []pgtype.Text
	CreatedAt    pgtype.Timestamp
	LikeCount    pgtype.Int8
	RepostCount  pgtype.Int8
	CommentCount pgtype.Int8
}

// ConvertPostToPostgres converts models.Post to PostPostgres.
func ConvertPostToPostgres(post models.Post) PostPostgres {
	var pics []pgtype.Text
	for _, pic := range post.Pics {
		if pic != "" {
			pics = append(pics, pgtype.Text{String: pic, Valid: true})
		}
	}

	return PostPostgres{
		Id:           pgtype.UUID{Bytes: post.Id, Valid: true},
		CreatorId:    pgtype.UUID{Bytes: post.CreatorId, Valid: true},
		Desc:         convertStringToPostgresText(post.Desc),
		Pics:         pics,
		CreatedAt:    pgtype.Timestamp{Time: post.CreatedAt, Valid: true},
		LikeCount:    pgtype.Int8{Int64: int64(post.LikeCount), Valid: true},
		RepostCount:  pgtype.Int8{Int64: int64(post.RepostCount), Valid: true},
		CommentCount: pgtype.Int8{Int64: int64(post.CommentCount), Valid: true},
	}
}

// ToPost converts PostPostgres to models.Post.
func (p *PostPostgres) ToPost() models.Post {
	var picsSlice []string

	for _, pics := range p.Pics {
		picsSlice = append(picsSlice, pics.String)
	}

	return models.Post{
		Id:           p.Id.Bytes,
		CreatorId:    p.CreatorId.Bytes,
		Desc:         p.Desc.String,
		Pics:         picsSlice,
		CreatedAt:    p.CreatedAt.Time,
		LikeCount:    int(p.LikeCount.Int64),
		RepostCount:  int(p.RepostCount.Int64),
		CommentCount: int(p.CommentCount.Int64),
	}
}
