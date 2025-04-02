package postgres_models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"quickflow/internal/models"
)

type PGFileURL struct {
	Id  pgtype.UUID
	URL pgtype.Text
}

type PostPostgres struct {
	Id           pgtype.UUID
	CreatorId    pgtype.UUID
	Desc         pgtype.Text
	ImagesURLs   []PGFileURL
	CreatedAt    pgtype.Timestamp
	LikeCount    pgtype.Int8
	RepostCount  pgtype.Int8
	CommentCount pgtype.Int8
}

// ConvertPostToPostgres converts models.Post to PostPostgres.
func ConvertPostToPostgres(post models.Post) PostPostgres {
	var pics []PGFileURL
	for id, imageURL := range post.ImagesURL {
		if len(imageURL) != 0 {
			pics = append(pics, PGFileURL{
				Id:  pgtype.UUID{Bytes: id, Valid: true},
				URL: convertStringToPostgresText(imageURL),
			})
		}
	}

	return PostPostgres{
		Id:           pgtype.UUID{Bytes: post.Id, Valid: true},
		CreatorId:    pgtype.UUID{Bytes: post.CreatorId, Valid: true},
		Desc:         convertStringToPostgresText(post.Desc),
		ImagesURLs:   pics,
		CreatedAt:    pgtype.Timestamp{Time: post.CreatedAt, Valid: true},
		LikeCount:    pgtype.Int8{Int64: int64(post.LikeCount), Valid: true},
		RepostCount:  pgtype.Int8{Int64: int64(post.RepostCount), Valid: true},
		CommentCount: pgtype.Int8{Int64: int64(post.CommentCount), Valid: true},
	}
}

// ToPost converts PostPostgres to models.Post.
func (p *PostPostgres) ToPost() models.Post {
	var imagesMap = make(map[uuid.UUID]string)

	for _, pgURL := range p.ImagesURLs {
		imagesMap[pgURL.Id.Bytes] = pgURL.URL.String
	}

	return models.Post{
		Id:           p.Id.Bytes,
		CreatorId:    p.CreatorId.Bytes,
		Desc:         p.Desc.String,
		ImagesURL:    imagesMap,
		CreatedAt:    p.CreatedAt.Time,
		LikeCount:    int(p.LikeCount.Int64),
		RepostCount:  int(p.RepostCount.Int64),
		CommentCount: int(p.CommentCount.Int64),
	}
}
