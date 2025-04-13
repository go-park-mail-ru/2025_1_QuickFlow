package forms

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"quickflow/config"
	"quickflow/internal/models"
)

type File struct {
	Name string
	Data []byte
}

type PostForm struct {
	Text     string         `json:"text"`
	Images   []*models.File `json:"pics"`
	IsRepost bool           `json:"is_repost"`
}

func (p *PostForm) ToPostModel(userId uuid.UUID) models.Post {
	var postModel models.Post
	postModel.Desc = p.Text
	postModel.CreatorId = userId
	postModel.CreatedAt = time.Now()
	postModel.Images = p.Images
	postModel.IsRepost = p.IsRepost

	return postModel
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
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url,omitempty"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	IsOnline  *bool  `json:"online,omitempty"`
}

func PublicUserInfoToOut(info models.PublicUserInfo) PublicUserInfoOut {
	return PublicUserInfoOut{
		ID:        info.Id.String(),
		Username:  info.Username,
		FirstName: info.Firstname,
		LastName:  info.Lastname,
		AvatarURL: info.AvatarURL,
	}
}

type PostOut struct {
	Id           string            `json:"id"`
	Creator      PublicUserInfoOut `json:"author"`
	Desc         string            `json:"text"`
	Pics         []string          `json:"pics"`
	CreatedAt    string            `json:"created_at"`
	LikeCount    int               `json:"like_count"`
	RepostCount  int               `json:"repost_count"`
	CommentCount int               `json:"comment_count"`
	IsRepost     bool              `json:"is_repost"`
}

func (p *PostOut) FromPost(post models.Post) {
	var urls []string
	for _, url := range post.ImagesURL {
		urls = append(urls, url)
	}

	p.Id = post.Id.String()
	p.Desc = post.Desc
	p.Pics = urls
	p.CreatedAt = post.CreatedAt.Format(config.TimeStampLayout)
	p.LikeCount = post.LikeCount
	p.RepostCount = post.RepostCount
	p.CommentCount = post.CommentCount
	p.IsRepost = post.IsRepost
}

type UpdatePostForm struct {
	Id     string         `json:"-"`
	Text   string         `json:"text"`
	Images []*models.File `json:"pics"`
}

func (p *UpdatePostForm) ToPostUpdateModel(postId uuid.UUID) (models.PostUpdate, error) {

	return models.PostUpdate{
		Id:    postId,
		Desc:  p.Text,
		Files: p.Images,
	}, nil
}
