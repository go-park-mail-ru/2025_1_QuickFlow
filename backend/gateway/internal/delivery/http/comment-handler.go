package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"

	time2 "quickflow/config/time"
	"quickflow/gateway/internal/delivery/http/forms"
	errors2 "quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type CommentService interface {
	FetchCommentsForPost(ctx context.Context, postId uuid.UUID, numComments int, timestamp time.Time) ([]models.Comment, error)
	AddComment(ctx context.Context, comment models.Comment) (*models.Comment, error)
	DeleteComment(ctx context.Context, userId uuid.UUID, commentId uuid.UUID) error
	UpdateComment(ctx context.Context, update models.CommentUpdate, userId uuid.UUID) (*models.Comment, error)
	LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	GetComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) (*models.Comment, error)
	GetLastPostComment(ctx context.Context, postId uuid.UUID) (*models.Comment, error)
}

type CommentHandler struct {
	commentUseCase CommentService
	profileService ProfileUseCase
	policy         *bluemonday.Policy
}

// NewCommentHandler creates a new comment handler.
func NewCommentHandler(commentUseCase CommentService, profileService ProfileUseCase, policy *bluemonday.Policy) *CommentHandler {
	return &CommentHandler{
		commentUseCase: commentUseCase,
		profileService: profileService,
		policy:         policy,
	}
}

// AddComment добавляет новый комментарий
// @Summary Добавить комментарий
// @Description Добавляет новый комментарий к посту
// @Tags Comments
// @Accept json
// @Produce json
// @Param comment body forms.CommentForm true "Комментарий"
// @Success 200 {object} forms.CommentOut "Комментарий успешно добавлен"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/comments [post]
func (c *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse comment ID: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid comment ID", http.StatusBadRequest))
		return
	}

	var commentForm forms.CommentForm
	if err := json.NewDecoder(r.Body).Decode(&commentForm); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse comment form: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid data", http.StatusBadRequest))
		return
	}

	// Проверка длины текста комментария
	if len(commentForm.Text) > 4096 {
		logger.Error(ctx, fmt.Sprintf("Text length validation failed: %d", len(commentForm.Text)))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Text length validation failed", http.StatusBadRequest))
		return
	}

	// sanitize the text
	commentForm.Text = c.policy.Sanitize(commentForm.Text)

	// Добавление комментария
	commentModel := commentForm.ToCommentModel()
	commentModel.UserId = user.Id
	commentModel.PostId = postId
	newComment, err := c.commentUseCase.AddComment(ctx, commentModel)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to add comment: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	// Подготовка ответа
	// get public user info
	publicUserInfo, err := c.profileService.GetPublicUserInfo(ctx, newComment.UserId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get public user info: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	var commentOut forms.CommentOut
	commentOut.FromComment(*newComment, publicUserInfo)

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commentOut); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode comment: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode comment", http.StatusInternalServerError))
	}
}

// DeleteComment удаляет комментарий
// @Summary Удалить комментарий
// @Description Удаляет комментарий по ID
// @Tags Comments
// @Param comment_id path string true "Идентификатор комментария"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 403 {object} forms.ErrorForm "Комментарий не принадлежит пользователю"
// @Failure 404 {object} forms.ErrorForm "Комментарий не найден"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/comments/{comment_id} [delete]
func (c *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	// Извлекаем ID комментария из URL
	commentIdStr := mux.Vars(r)["comment_id"]
	commentId, err := uuid.Parse(commentIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse comment ID: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid comment ID", http.StatusBadRequest))
		return
	}

	// Удаление комментария
	err = c.commentUseCase.DeleteComment(ctx, user.Id, commentId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to delete comment: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateComment обновляет комментарий
// @Summary Обновить комментарий
// @Description Обновляет комментарий по ID
// @Tags Comments
// @Accept json
// @Produce json
// @Param comment_id path string true "Идентификатор комментария"
// @Param comment body forms.CommentForm true "Обновление комментария"
// @Success 200 {object} forms.CommentOut "Комментарий успешно обновлен"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 404 {object} forms.ErrorForm "Комментарий не найден"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/comments/{comment_id} [put]
func (c *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	// Извлекаем ID комментария из URL
	commentIdStr := mux.Vars(r)["comment_id"]
	commentId, err := uuid.Parse(commentIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse comment ID: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid comment ID", http.StatusBadRequest))
		return
	}

	// sanitize the text
	commentIdStr = c.policy.Sanitize(commentIdStr)

	// Декодируем данные комментария из JSON
	var commentForm forms.CommentUpdateForm
	if err := json.NewDecoder(r.Body).Decode(&commentForm); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to decode comment data: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid comment data", http.StatusBadRequest))
		return
	}

	// Обновление комментария
	commentUpdate := commentForm.ToCommentUpdateModel(commentId)
	commentUpdate.Id = commentId
	updatedComment, err := c.commentUseCase.UpdateComment(ctx, commentUpdate, user.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to update comment: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	// Подготовка ответа
	// get public user info
	publicUserInfo, err := c.profileService.GetPublicUserInfo(ctx, updatedComment.UserId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to get public user info: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	var commentOut forms.CommentOut
	commentOut.FromComment(*updatedComment, publicUserInfo)

	// Отправка обновленного комментария
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commentOut); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode updated comment: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode updated comment", http.StatusInternalServerError))
	}
}

// FetchCommentsForPost возвращает комментарии для поста
// @Summary Получить комментарии для поста
// @Description Возвращает список комментариев для указанного поста
// @Tags Comments
// @Produce json
// @Param post_id path string true "Идентификатор поста"
// @Param posts_count query int true "Количество комментариев"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.CommentOut "Список комментариев"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/posts/{post_id}/comments [get]
func (c *CommentHandler) FetchCommentsForPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем ID поста из URL
	postIdStr := mux.Vars(r)["post_id"]
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse post ID: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid post ID", http.StatusBadRequest))
		return
	}

	// Извлекаем параметры запроса
	var feedForm forms.CommentFetchForm
	err = feedForm.GetParams(r.URL.Query())
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse query params: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid query parameters", http.StatusBadRequest))
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	// Получаем комментарии для поста
	comments, err := c.commentUseCase.FetchCommentsForPost(ctx, postId, feedForm.Count, ts)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to fetch comments for post %s: %s", postId.String(), err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	// Получаем информацию о пользователе для каждого комментария
	publicUserInfos := make(map[uuid.UUID]models.PublicUserInfo)
	for _, comment := range comments {
		if _, exists := publicUserInfos[comment.UserId]; !exists {
			publicUserInfo, err := c.profileService.GetPublicUserInfo(ctx, comment.UserId)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Failed to get public user info for comment %s: %s", comment.Id.String(), err.Error()))
				http2.WriteJSONError(w, err)
				return
			}
			publicUserInfos[comment.UserId] = publicUserInfo
		}
	}
	// Подготовка списка комментариев для ответа
	var commentsOut []forms.CommentOut
	for _, comment := range comments {
		var commentOut forms.CommentOut
		commentOut.FromComment(comment, publicUserInfos[comment.UserId])
		commentsOut = append(commentsOut, commentOut)
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(commentsOut); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to encode comments: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to encode comments", http.StatusInternalServerError))
	}
}

// LikeComment ставит лайк на комментарий
// @Summary Поставить лайк на комментарий
// @Description Позволяет пользователю поставить лайк на комментарий
// @Tags Comments
// @Param comment_id path string true "ID комментария"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/comments/{comment_id}/like [post]
func (c *CommentHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	// Извлекаем пользователя из контекста
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while liking comment")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	// Извлекаем ID комментария из параметров URL
	commentIdStr := mux.Vars(r)["comment_id"]
	commentId, err := uuid.Parse(commentIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse comment ID: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid comment ID", http.StatusBadRequest))
		return
	}

	// Пытаемся поставить лайк
	err = c.commentUseCase.LikeComment(ctx, commentId, user.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to like comment: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	// Возвращаем успех
	w.WriteHeader(http.StatusOK)
}

// UnlikeComment убирает лайк с комментария
// @Summary Убрать лайк с комментария
// @Description Позволяет пользователю убрать лайк с комментария
// @Tags Comments
// @Param comment_id path string true "ID комментария"
// @Success 200 {string} string "OK"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/comments/{comment_id}/unlike [post]
func (c *CommentHandler) UnlikeComment(w http.ResponseWriter, r *http.Request) {
	// Извлекаем пользователя из контекста
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while unliking comment")
		http2.WriteJSONError(w, errors2.New(errors2.InternalErrorCode, "Failed to get user from context", http.StatusInternalServerError))
		return
	}

	// Извлекаем ID комментария из параметров URL
	commentIdStr := mux.Vars(r)["comment_id"]
	commentId, err := uuid.Parse(commentIdStr)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to parse comment ID: %s", err.Error()))
		http2.WriteJSONError(w, errors2.New(errors2.BadRequestErrorCode, "Invalid comment ID", http.StatusBadRequest))
		return
	}

	// Пытаемся убрать лайк
	err = c.commentUseCase.UnlikeComment(ctx, commentId, user.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to unlike comment: %s", err.Error()))
		http2.WriteJSONError(w, err)
		return
	}

	// Возвращаем успех
	w.WriteHeader(http.StatusOK)
}
