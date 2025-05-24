package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type PostgresFile struct {
	URL         pgtype.Text
	DisplayType pgtype.Text
	Name        pgtype.Text
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

	if f.Name.Valid {
		file.Name = f.Name.String
	} else {
		file.Name = ""
	}
	return &file
}

func PostgresFilesToModels(files []PostgresFile) []*models.File {
	var resFiles []*models.File
	for _, file := range files {
		resFiles = append(resFiles, file.ToFile())
	}
	return resFiles
}

func FileToPostgres(file models.File) *PostgresFile {
	resFile := PostgresFile{URL: pgtype.Text{String: file.URL, Valid: true}, Name: pgtype.Text{String: file.Name, Valid: true}}
	if len(file.DisplayType) != 0 {
		resFile.DisplayType = pgtype.Text{String: string(file.DisplayType), Valid: true}
	}
	return &resFile
}

func FilesToPostgres(files []*models.File) []PostgresFile {
	var postgresFiles []PostgresFile
	for _, file := range files {
		postgresFiles = append(postgresFiles, *FileToPostgres(*file))
	}
	return postgresFiles
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
