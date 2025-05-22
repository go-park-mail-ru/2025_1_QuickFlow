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
	Text        string                 `json:"text"`
	Images      []*models.File         `json:"pics"`
	IsRepost    bool                   `json:"is_repost"`
	CreatorId   uuid.UUID              `json:"author_id"`
	CreatorType models.PostCreatorType `json:"author_type"`
}

func ParseCreationType(creatorType string) (models.PostCreatorType, error) {
	switch creatorType {
	case string(models.PostUser):
		return models.PostUser, nil
	case string(models.PostCommunity):
		return models.PostCommunity, nil
	default:
		return "", errors.New("invalid creator type")
	}
}

func (p *PostForm) ToPostModel() models.Post {
	var postModel models.Post
	postModel.Desc = p.Text
	postModel.CreatorId = p.CreatorId
	postModel.CreatorType = p.CreatorType
	postModel.CreatedAt = time.Now()
	postModel.UpdatedAt = time.Now()
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
	Pics         []string    `json:"pics"`
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
	var urls []string
	for _, url := range post.ImagesURL {
		urls = append(urls, url)
	}

	p.Id = post.Id.String()
	p.Desc = post.Desc
	p.Pics = urls
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
