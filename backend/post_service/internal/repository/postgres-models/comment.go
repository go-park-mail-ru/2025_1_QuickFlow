package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type PostgresFile struct {
	URL         pgtype.Text
	DisplayType pgtype.Text
}

func (f *PostgresFile) ToFile() *models.File {
	file := models.File{
		URL: f.URL.String,
	}
	if f.DisplayType.Valid && len(f.DisplayType.String) != 0 {
		file.DisplayType = models.DisplayType(f.DisplayType.String)
	} else {
		file.DisplayType = models.DisplayTypeFile
	}
	return &file
}

func FileToPostgres(file models.File) *PostgresFile {
	resFile := PostgresFile{URL: pgtype.Text{String: file.URL, Valid: true}}
	if len(file.DisplayType) != 0 {
		resFile.DisplayType = pgtype.Text{String: string(file.DisplayType), Valid: true}
	}
	return &resFile
}

type CommentPostgres struct {
	Id        pgtype.UUID
	PostId    pgtype.UUID
	UserId    pgtype.UUID
	Text      pgtype.Text
	Files     []*PostgresFile
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	LikeCount pgtype.Int8
	IsLiked   pgtype.Bool
}

// ConvertCommentToPostgres converts models.Comment to CommentPostgres.
func ConvertCommentToPostgres(comment models.Comment) CommentPostgres {
	var pics []*PostgresFile
	for _, pic := range comment.Images {
		pics = append(pics, FileToPostgres(*pic))
	}

	return CommentPostgres{
		Id:        pgtype.UUID{Bytes: comment.Id, Valid: true},
		PostId:    pgtype.UUID{Bytes: comment.PostId, Valid: true},
		UserId:    pgtype.UUID{Bytes: comment.UserId, Valid: true},
		Text:      convertStringToPostgresText(comment.Text),
		Files:     pics,
		CreatedAt: pgtype.Timestamptz{Time: comment.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: comment.UpdatedAt, Valid: true},
		LikeCount: pgtype.Int8{Int64: int64(comment.LikeCount), Valid: true},
		IsLiked:   pgtype.Bool{Bool: comment.IsLiked, Valid: true},
	}
}

// ToComment converts CommentPostgres to models.Comment.
func (c *CommentPostgres) ToComment() models.Comment {
	var Files []*models.File

	for _, pic := range c.Files {
		Files = append(Files, pic.ToFile())
	}

	return models.Comment{
		Id:        c.Id.Bytes,
		PostId:    c.PostId.Bytes,
		UserId:    c.UserId.Bytes,
		Text:      c.Text.String,
		Images:    Files,
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
		LikeCount: int(c.LikeCount.Int64),
		IsLiked:   c.IsLiked.Bool,
	}
}
