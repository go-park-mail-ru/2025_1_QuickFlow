package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/forms"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/gateway/pkg/sanitizer"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type PostHandler struct {
	postUseCase      PostService
	profileUseCase   ProfileUseCase
	communityService CommunityService
	policy           *bluemonday.Policy
}

// NewPostHandler creates new post handler.
func NewPostHandler(postUseCase PostService, profileUseCase ProfileUseCase,
	communityService CommunityService, policy *bluemonday.Policy) *PostHandler {
	return &PostHandler{
		postUseCase:      postUseCase,
		profileUseCase:   profileUseCase,
		communityService: communityService,
		policy:           policy,
	}
}

// AddPost добавляет новый пост
// @Summary Добавить пост
// @Description Добавляет новый пост в ленту
// @Tags Feed
// @Accept multipart/form-data
// @Produce json
// @Param text formData string true "Текст поста"
// @Param pics formData file false "Изображения"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/post [post]
func (p *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while adding post")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested to add post", user.Username))

	err := r.ParseMultipartForm(15 << 20) // 10 MB
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// parsing JSON
	var postForm forms.PostForm
	postForm.Text = r.FormValue("text")
	isRepostString := r.FormValue("is_repost")
	creatorType := r.FormValue("author_type")
	creatorId := r.FormValue("author_id")
	if len(creatorType) == 0 {
		postForm.CreatorType = models.PostUser
		postForm.CreatorId = user.Id
	} else {
		postForm.CreatorType, err = forms.ParseCreationType(creatorType)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to parse creator type: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		if postForm.CreatorType == models.PostCommunity {
			id, err := uuid.Parse(creatorId)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Failed to parse creator id: %s", err.Error()))
				http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
				return
			}
			postForm.CreatorId = id
		} else {
			postForm.CreatorId = user.Id
		}
	}

	if utf8.RuneCountInString(postForm.Text) > 4000 {
		logger.Error(ctx, fmt.Sprintf("Text length validation failed: length=%d", utf8.RuneCountInString(postForm.Text)))
		http2.WriteJSONError(w, "Text must be between 1 and 4096 characters", http.StatusBadRequest)
		return
	}

	sanitizer.SanitizePost(&postForm, p.policy)

	if len(isRepostString) != 0 {
		postForm.IsRepost, err = strconv.ParseBool(r.FormValue("is_repost"))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to parse is_repost: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
	}

	postForm.Images, err = http2.GetFiles(r, "pics")
	if errors.Is(err, http2.TooManyFilesErr) {
		logger.Error(ctx, fmt.Sprintf("Too many pics requested: %s", err.Error()))
		http2.WriteJSONError(w, "Too many pics requested", http.StatusBadRequest)
		return
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get files: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get files", http.StatusBadRequest)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Recieved post: %+v", postForm))

	post := postForm.ToPostModel()

	newPost, err := p.postUseCase.AddPost(ctx, post)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to add post: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to add post", err.HTTPStatus)
		return
	}
	logger.Info(ctx, "Successfully added post")

	var postOut forms.PostOut
	postOut.FromPost(*newPost)
	if newPost.CreatorType == models.PostUser {
		publicUserInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, user.Id)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to get public user info: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get public user info", err.HTTPStatus)
			return
		}
		info := forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
		postOut.Creator = &info
	} else {
		community, err := p.communityService.GetCommunityById(ctx, post.CreatorId)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to get community: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get community", err.HTTPStatus)
			return
		}

		postOut.Creator = forms.ToCommunityForm(*community)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.PostOut]{Payload: postOut})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode post: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode post", http.StatusInternalServerError)
		return
	}
}

// DeletePost удаляет пост
// @Summary Удалить пост
// @Description Удаляет пост из ленты
// @Tags Feed
// @Param post_id path string true "Идентификатор поста"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 403 {object} forms.ErrorForm "Пост не принадлежит пользователю"
// @Failure 404 {object} forms.ErrorForm "Пост не найден"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/posts/{post_id} [delete]
func (p *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while deleting post")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	postIdString := mux.Vars(r)["post_id"]
	if len(postIdString) == 0 {
		logger.Error(ctx, "Failed to get post id from request")
		http2.WriteJSONError(w, "Failed to get post id", http.StatusBadRequest)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested to delete post with id %s", user.Username, postIdString))

	postId, err := uuid.Parse(postIdString)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse post id: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse post id", http.StatusBadRequest)
		return
	}

	err = p.postUseCase.DeletePost(ctx, user.Id, postId)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to delete post: %s", err.Error()))
		http2.WriteJSONError(w, "Failed delete post", err.HTTPStatus)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Successfully deleted post %s", postIdString))
}

// UpdatePost updates a post
// @Summary Update post
// @Description Updates a post
// @Tags Post
// @Accept json
// @Produce json
// @Param post_id path string true "Post ID"
// @Param post body forms.UpdatePostForm true "Post data"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Invalid data"
// @Failure 403 {object} forms.ErrorForm "Post does not belong to user"
// @Failure 404 {object} forms.ErrorForm "Post not found"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/post [put]
func (p *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while updating post")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	postIdString := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdString)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse post id: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse post id", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(15 << 20) // 10 MB
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse form: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	var updatePostForm forms.UpdatePostForm
	updatePostForm.Text = r.FormValue("text")

	if utf8.RuneCountInString(updatePostForm.Text) > 4000 {
		logger.Error(ctx, fmt.Sprintf("Text length validation failed: length=%d", utf8.RuneCountInString(updatePostForm.Text)))
		http2.WriteJSONError(w, "Text must be between 1 and 4096 characters", http.StatusBadRequest)
		return

	}

	sanitizer.SanitizeUpdatePost(&updatePostForm, p.policy)

	updatePostForm.Images, err = http2.GetFiles(r, "pics")
	if errors.Is(err, http2.TooManyFilesErr) {
		logger.Error(ctx, fmt.Sprintf("Too many pics requested: %s", err.Error()))
		http2.WriteJSONError(w, "Too many pics requested", http.StatusBadRequest)
		return
	} else if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get files: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get files", http.StatusBadRequest)
		return
	}

	updatePost, err := updatePostForm.ToPostUpdateModel(postId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse update post: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to parse update post", http.StatusBadRequest)
		return
	}
	post, err := p.postUseCase.UpdatePost(ctx, updatePost, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to update post: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to update post", err.HTTPStatus)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully updated post %s", postIdString))
	var postOut forms.PostOut
	postOut.FromPost(*post)
	publicUserInfo, err := p.profileUseCase.GetPublicUserInfo(ctx, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		logger.Error(ctx, fmt.Sprintf("Failed to get public user info: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to get public user info", err.HTTPStatus)
		return
	}
	info := forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
	postOut.Creator = &info
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(forms.PayloadWrapper[forms.PostOut]{Payload: postOut})
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode post: %s", err.Error()))
		http2.WriteJSONError(w, "Failed to encode post", http.StatusInternalServerError)
		return
	}
}

func (p *PostHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse post ID", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("User %s liked post %s", user.Username, postId.String()))

	err = p.postUseCase.LikePost(ctx, postId, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to like post: %s", err.Error()), err.HTTPStatus)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (p *PostHandler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse post ID", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, fmt.Sprintf("User %s unliked post %s", user.Username, postId.String()))

	err = p.postUseCase.UnlikePost(ctx, postId, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to unlike post: %s", err.Error()), err.HTTPStatus)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
