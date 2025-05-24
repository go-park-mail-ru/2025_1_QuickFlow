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

type File struct {
	Name string
	Data []byte
}

type PostForm struct {
	Text        string    `json:"text,omitempty"`
	Media       []string  `form:"media" json:"media,omitempty"`
	Audio       []string  `form:"audio" json:"audio,omitempty"`
	File        []string  `form:"files" json:"files,omitempty"`
	IsRepost    bool      `json:"is_repost,omitempty"`
	CreatorId   uuid.UUID `json:"author_id,omitempty"`
	CreatorType string    `json:"author_type,omitempty"`
}

func ParseCreatorType(creatorType string) (models.PostCreatorType, error) {
	switch creatorType {
	case string(models.PostUser):
		return models.PostUser, nil
	case string(models.PostCommunity):
		return models.PostCommunity, nil
	default:
		return "", errors.New("invalid creator type")
	}
}

func (p *PostForm) ToPostModel(userId uuid.UUID) (models.Post, error) {
	var attachments []*models.File
	for _, file := range p.Media {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeMedia,
		})
	}

	for _, file := range p.Audio {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeAudio,
		})
	}

	for _, file := range p.File {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeFile,
		})
	}
	var postModel models.Post

	if len(p.CreatorType) == 0 {
		postModel.CreatorType = models.PostUser
		postModel.CreatorId = userId
	} else {
		var err error
		postModel.CreatorType, err = ParseCreatorType(p.CreatorType)
		if err != nil {
			return models.Post{}, errors.New("invalid creator type")
		}

		if postModel.CreatorType == models.PostCommunity {
			postModel.CreatorId = p.CreatorId
		} else {
			postModel.CreatorId = userId
		}
	}

	postModel.Desc = p.Text
	postModel.CreatedAt = time.Now()
	postModel.UpdatedAt = time.Now()
	postModel.Files = attachments
	postModel.IsRepost = p.IsRepost

	return postModel, nil
}

type FeedForm struct {
	Posts int    `json:"posts_count"`
	Ts    string `json:"ts"`
}

// GetParams gets parameters from the map
func (f *FeedForm) GetParams(values url.Values) error {
	var (
		err      error
		numPosts int64
	)

	if !values.Has("posts_count") {
		return errors.New("posts_count parameter missing")
	}

	numPosts, err = strconv.ParseInt(values.Get("posts_count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse posts_count")
	}

	f.Posts = int(numPosts)
	f.Ts = values.Get("ts")
	return nil
}

type PublicUserInfoOut struct {
	ID        string              `json:"id"`
	Username  string              `json:"username"`
	AvatarURL string              `json:"avatar_url,omitempty"`
	FirstName string              `json:"firstname"`
	LastName  string              `json:"lastname"`
	IsOnline  *bool               `json:"online,omitempty"`
	Relation  models.UserRelation `json:"relation,omitempty"`
}

func PublicUserInfoToOut(info models.PublicUserInfo, relation models.UserRelation) PublicUserInfoOut {
	return PublicUserInfoOut{
		ID:        info.Id.String(),
		Username:  info.Username,
		FirstName: info.Firstname,
		LastName:  info.Lastname,
		AvatarURL: info.AvatarURL,
		Relation:  relation,
	}
}

type PostOut struct {
	Id           string      `json:"id"`
	Creator      interface{} `json:"author,omitempty"`
	CreatorType  string      `json:"author_type"`
	Desc         string      `json:"text"`
	MediaURLs    []FileOut   `json:"media,omitempty"`
	AudioURLs    []FileOut   `json:"audio,omitempty"`
	FileURLs     []FileOut   `json:"files,omitempty"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
	LikeCount    int         `json:"like_count"`
	RepostCount  int         `json:"repost_count"`
	CommentCount int         `json:"comment_count"`
	IsRepost     bool        `json:"is_repost"`
	IsLiked      bool        `json:"is_liked"`
	LastComment  *CommentOut `json:"last_comment,omitempty"`
}

func (p *PostOut) FromPost(post models.Post) {
	mediaURLs := make([]FileOut, 0)
	audioURLs := make([]FileOut, 0)
	fileURLs := make([]FileOut, 0)

	for _, file := range post.Files {
		if file.DisplayType == models.DisplayTypeMedia {
			mediaURLs = append(mediaURLs, ToFileOut(*file))
		} else if file.DisplayType == models.DisplayTypeAudio {
			audioURLs = append(audioURLs, ToFileOut(*file))
		} else {
			fileURLs = append(fileURLs, ToFileOut(*file))
		}
	}

	p.Id = post.Id.String()
	p.Desc = post.Desc
	p.MediaURLs = mediaURLs
	p.AudioURLs = audioURLs
	p.FileURLs = fileURLs
	p.CreatedAt = post.CreatedAt.Format(time2.TimeStampLayout)
	p.UpdatedAt = post.UpdatedAt.Format(time2.TimeStampLayout)
	p.Creator = &PublicUserInfoOut{
		ID: post.CreatorId.String(),
	}
	p.CreatorType = string(post.CreatorType)
	p.LikeCount = post.LikeCount
	p.RepostCount = post.RepostCount
	p.CommentCount = post.CommentCount
	p.IsRepost = post.IsRepost
	p.IsLiked = post.IsLiked
}

type UpdatePostForm struct {
	Id    string   `json:"-"`
	Text  string   `json:"text"`
	Media []string `form:"media" json:"media,omitempty"`
	Audio []string `form:"audio" json:"audio,omitempty"`
	File  []string `form:"files" json:"files,omitempty"`
}

func (p *UpdatePostForm) ToPostUpdateModel(postId uuid.UUID) (models.PostUpdate, error) {
	var attachments []*models.File
	for _, file := range p.Media {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeMedia,
		})
	}

	for _, file := range p.Audio {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeAudio,
		})
	}

	for _, file := range p.File {
		attachments = append(attachments, &models.File{
			URL:         file,
			DisplayType: models.DisplayTypeFile,
		})
	}

	return models.PostUpdate{
		Id:    postId,
		Desc:  p.Text,
		Files: attachments,
	}, nil
}
