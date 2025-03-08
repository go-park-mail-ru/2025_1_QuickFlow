package forms

import (
	"quickflow/config"
	"quickflow/internal/models"
)

type PostForm struct {
	Desc string   `json:"text"`
	Pics []string `json:"pics"`
}

type FeedForm struct {
	Posts int    `json:"posts_count"`
	Ts    string `json:"ts"`
}

type PostOut struct {
	Id           string   `json:"id"`
	CreatorId    string   `json:"creator_id"`
	Desc         string   `json:"text"`
	Pics         []string `json:"pics"`
	CreatedAt    string   `json:"created_at"`
	LikeCount    int      `json:"like_count"`
	RepostCount  int      `json:"repost_count"`
	CommentCount int      `json:"comment_count"`
}

func (p *PostOut) FromPost(post models.Post) {
	p.Id = post.Id.String()
	p.CreatorId = post.CreatorId.String()
	p.Desc = post.Desc
	p.Pics = post.Pics
	p.CreatedAt = post.CreatedAt.Format(config.TimeStampLayout)
	p.LikeCount = post.LikeCount
	p.RepostCount = post.RepostCount
	p.CommentCount = post.CommentCount
}
