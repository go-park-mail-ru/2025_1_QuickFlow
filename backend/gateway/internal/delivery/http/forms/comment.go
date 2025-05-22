package forms

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	time2 "quickflow/config/time"
	"quickflow/shared/models"
)

type CommentForm struct {
	Text  string   `json:"text"`
	Media []string `json:"media,omitempty"`
	Audio []string `json:"audio,omitempty"`
	Files []string `json:"files,omitempty"`
}

func (f *CommentForm) ToCommentModel() models.Comment {
	var attachments []*models.File
	for _, file := range f.Files {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeFile,
		})
	}

	for _, file := range f.Media {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeMedia,
		})
	}

	for _, file := range f.Audio {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeAudio,
		})
	}

	return models.Comment{
		Id:        uuid.New(),
		Text:      f.Text,
		Images:    attachments,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LikeCount: 0,
		IsLiked:   false,
	}
}

type CommentUpdateForm struct {
	CommentForm
}

func (f *CommentUpdateForm) ToCommentUpdateModel(commentId uuid.UUID) models.CommentUpdate {
	var attachments []*models.File
	for _, file := range f.Files {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeFile,
		})
	}

	for _, file := range f.Media {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeMedia,
		})
	}

	for _, file := range f.Audio {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeAudio,
		})
	}

	return models.CommentUpdate{
		Id:    commentId,
		Text:  f.Text,
		Files: attachments,
	}
}

type CommentOut struct {
	ID        string            `json:"id"`
	Text      string            `json:"text"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Media     []string          `json:"media,omitempty"`
	Audio     []string          `json:"audio,omitempty"`
	Files     []string          `json:"files,omitempty"`
	Creator   PublicUserInfoOut `json:"author"`
	PostId    uuid.UUID         `json:"post_id"`
	LikeCount int               `json:"like_count"`
	IsLiked   bool              `json:"is_liked"`
}

func (c *CommentOut) FromComment(comment models.Comment, userInfo models.PublicUserInfo) {
	var files, media, audio []string
	for _, file := range comment.Images {
		switch file.DisplayType {
		case models.DisplayTypeFile:
			files = append(files, file.URL)
		case models.DisplayTypeMedia:
			media = append(media, file.URL)
		case models.DisplayTypeAudio:
			audio = append(audio, file.URL)
		}
	}

	c.ID = comment.Id.String()
	c.Text = comment.Text
	c.CreatedAt = comment.CreatedAt.Format(time2.TimeStampLayout)
	c.UpdatedAt = comment.UpdatedAt.Format(time2.TimeStampLayout)
	c.Media = media
	c.Audio = audio
	c.Files = files
	c.PostId = comment.PostId
	c.Creator = PublicUserInfoToOut(userInfo, models.RelationNone)
	c.LikeCount = comment.LikeCount
	c.IsLiked = comment.IsLiked
}

type CommentFetchForm struct {
	Count int    `json:"count"`
	Ts    string `json:"ts"`
}

// GetParams gets parameters from the map
func (f *CommentFetchForm) GetParams(values url.Values) error {
	var (
		err      error
		numPosts int64
	)

	if !values.Has("count") {
		return errors.New("count parameter missing")
	}

	numPosts, err = strconv.ParseInt(values.Get("count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse count")
	}

	f.Count = int(numPosts)
	f.Ts = values.Get("ts")
	return nil
}
